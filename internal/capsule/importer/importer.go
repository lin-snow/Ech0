// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package importer 把一个已通过 check 的胶囊落进目标库（spec §11.1 / §11.3）。
//
// 它是「导出即转储」的反向端点：胶囊字段 1:1 写回 DB 列，禁止任何数值转换。
// 唯一的补全是 Echo.UserID —— 胶囊不携带这个内部必填外键，按 username 挂接同名
// 用户、否则挂 owner；展示归属始终以原样保留的 username 为准。
//
// 三条硬纪律，破坏其一即破坏整个边界：
//
//   - 不 import internal/capsule/check。校验由 CLI 统一前置（check.Run 有 error
//     即拒绝），本包的前提就是「胶囊已合法」，反向依赖只会让两个包成环。
//   - 不发布事件、不调 service 层。导入是数据搬运，不该唤醒 webhook / embedding /
//     agent 订阅者；只用 GORM 与领域模型。
//   - 无 --overwrite。id 已存在即跳过并计数，不覆盖不合并，重复执行安全。
package importer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/internal/kvstore"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// logModule 让胶囊各阶段的日志能被一条 module 过滤出来。
const logModule = "capsule"

// errDryRun 是 --dry-run 的回滚哨兵：统计照常累加在 Result 上（Result 由调用方
// 持有，不随事务回滚而失效），事务本身则靠这个 error 被丢弃。真正需要单独拦截的
// 只有两处非事务性副作用——写字节与写 KV 缓存，见 putBytes / applySite。
var errDryRun = errors.New("capsule import: dry-run rollback")

type Deps struct {
	DB       *gorm.DB
	Tx       transaction.Transactor
	Selector *storage.StorageSelector
	KV       kvstore.Store
}

type Options struct {
	IncludePrivate bool
	DryRun         bool
}

type Result struct {
	EchoesCreated, EchoesSkipped, SkippedPrivate int
	FilesCreated, FilesReused, FilesRenamed      int
	CommentsCreated, CommentsSkipped             int
	TagsCreated                                  int
	SiteFieldsFilled                             []string
	Renames                                      []string // "oldkey -> newkey"
	OrphanComments                               int
}

// Run 在单个事务内完成整个导入。调用方必须先跑 check 且确认无 error。
func Run(ctx context.Context, deps Deps, loaded *capsule.Loaded, opts Options) (*Result, error) {
	switch {
	case loaded == nil:
		return nil, errors.New("capsule import: loaded capsule is nil")
	case loaded.Manifest == nil:
		return nil, errors.New("capsule import: manifest is missing (run check first)")
	case loaded.Source == nil:
		return nil, errors.New("capsule import: capsule source is not open")
	case deps.Tx == nil:
		return nil, errors.New("capsule import: transactor is required")
	case deps.DB == nil:
		return nil, errors.New("capsule import: database handle is required")
	case deps.Selector == nil:
		return nil, errors.New("capsule import: storage selector is required")
	case deps.KV == nil:
		return nil, errors.New("capsule import: kv store is required")
	}

	result := &Result{}
	err := deps.Tx.Run(ctx, func(txCtx context.Context) error {
		s := newSession(txCtx, deps, loaded, opts, result)
		if err := s.run(txCtx); err != nil {
			return err
		}
		if opts.DryRun {
			return errDryRun
		}
		return nil
	})
	if err != nil && !errors.Is(err, errDryRun) {
		return nil, err
	}

	logUtil.GetLogger().Info("capsule import finished",
		slog.String("module", logModule),
		slog.Bool("dry_run", opts.DryRun),
		slog.Int("echoes_created", result.EchoesCreated),
		slog.Int("echoes_skipped", result.EchoesSkipped),
		slog.Int("skipped_private", result.SkippedPrivate),
		slog.Int("files_created", result.FilesCreated),
		slog.Int("files_reused", result.FilesReused),
		slog.Int("files_renamed", result.FilesRenamed),
		slog.Int("tags_created", result.TagsCreated),
		slog.Int("comments_created", result.CommentsCreated),
		slog.Int("comments_skipped", result.CommentsSkipped),
		slog.Int("orphan_comments", result.OrphanComments),
	)
	return result, nil
}

// session 承载一次导入的全部可变状态。所有跨阶段的查表缓存都挂在这里，
// 免得同一个 username / tag / file 在几百条 Echo 上重复打库。
type session struct {
	db       *gorm.DB
	selector *storage.StorageSelector
	kv       kvstore.Store
	loaded   *capsule.Loaded
	opts     Options
	res      *Result

	// ownerID 是归属兜底：胶囊里没有对应同名用户时，Echo 挂到执行导入的 owner。
	ownerID string
	// userIDByName 缓存 username -> 目标库 user id（未命中同名用户时缓存 ownerID）。
	userIDByName map[string]string
	// tagIDByName 缓存本次导入接触过的标签，同时充当 dry-run 下「已 find-or-create」的账本。
	tagIDByName map[string]string
	// landed 是本次运行新建的 Echo id。它只是 dry-run 的补丁：dry-run 下这些行
	// 不会真写进库，评论归属判定得靠它才不至于把新内容的评论全记成孤儿。
	landed map[string]struct{}

	// 当前存储后端的路由三元组，files 表的唯一索引 idx_file_route 就是按它 + key 建的。
	storageType storage.StorageType
	provider    string
	bucket      string
	keygen      storage.KeyGenerator
	// routeCache 记录「(路由, key) -> 已存在或本次新建的文件行」。dry-run 下新建的行
	// 只活在这里，好让同一胶囊内的重复 key 依然能被识别。
	routeCache map[string]*fileEntry
}

func newSession(ctx context.Context, deps Deps, loaded *capsule.Loaded, opts Options, res *Result) *session {
	db := deps.DB
	if tx, ok := transaction.TxFromContext(ctx); ok {
		db = tx
	}

	// 托管文件一律落到当前默认后端；external 由 url 条目单独表达，绝不进 selector。
	storageType := storage.StorageTypeLocal
	provider, bucket := "", ""
	if deps.Selector.ObjectEnabled() {
		storageType = storage.StorageTypeObject
		provider, bucket = deps.Selector.ObjectRoute()
	}

	return &session{
		db:           db.WithContext(ctx),
		selector:     deps.Selector,
		kv:           deps.KV,
		loaded:       loaded,
		opts:         opts,
		res:          res,
		userIDByName: make(map[string]string),
		tagIDByName:  make(map[string]string),
		landed:       make(map[string]struct{}),
		storageType:  storageType,
		provider:     provider,
		bucket:       bucket,
		keygen:       storage.NewRandomKeyGenerator(),
		routeCache:   make(map[string]*fileEntry),
	}
}

func (s *session) run(ctx context.Context) error {
	if err := s.resolveOwner(); err != nil {
		return err
	}
	if err := s.importEchoes(ctx); err != nil {
		return err
	}
	if err := s.importUnattachedFiles(ctx); err != nil {
		return err
	}
	if err := s.recountTagUsage(); err != nil {
		return err
	}
	if err := s.importComments(); err != nil {
		return err
	}
	if err := s.applySite(ctx); err != nil {
		return err
	}
	return s.applyConnects()
}

// resolveOwner 定位 owner。Echo.UserID 是 not null 外键，没有 owner 就无从兜底，
// 这属于目标库不可用而非胶囊问题，直接失败。
func (s *session) resolveOwner() error {
	var owner userModel.User
	if err := s.db.Where("is_owner = ?", true).First(&owner).Error; err != nil {
		return fmt.Errorf("capsule import: locate owner user: %w", err)
	}
	s.ownerID = owner.ID
	return nil
}

// resolveUserID 是本包唯一允许的字段补全：username 逐字入库不改写，UserID 另按
// 同名用户挂接、缺失则挂 owner（spec §11.3）。
func (s *session) resolveUserID(username string) (string, error) {
	if username == "" {
		return s.ownerID, nil
	}
	if id, ok := s.userIDByName[username]; ok {
		return id, nil
	}

	var user userModel.User
	err := s.db.Where("username = ?", username).First(&user).Error
	id := s.ownerID
	switch {
	case err == nil:
		id = user.ID
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return "", fmt.Errorf("capsule import: lookup user %q: %w", username, err)
	}
	s.userIDByName[username] = id
	return id, nil
}

func (s *session) importEchoes(ctx context.Context) error {
	for i := range s.loaded.Echoes {
		le := &s.loaded.Echoes[i]
		// check 已经拒绝过解析失败的胶囊；此处再拒一次，避免调用方漏跑 check 时静默丢内容。
		if le.Err != nil {
			return fmt.Errorf("capsule import: %s: %w", le.Path, le.Err)
		}
		if le.Doc == nil {
			return fmt.Errorf("capsule import: %s: empty echo document", le.Path)
		}
		if err := s.importEcho(ctx, le.Path, le.Doc); err != nil {
			return err
		}
	}
	return nil
}

func (s *session) importEcho(ctx context.Context, path string, doc *capsule.EchoDoc) error {
	if doc.Private && !s.opts.IncludePrivate {
		s.res.SkippedPrivate++
		return nil
	}
	if doc.ID == "" {
		return fmt.Errorf("capsule import: %s: missing id", path)
	}

	var existing int64
	if err := s.db.Model(&echoModel.Echo{}).Where("id = ?", doc.ID).Count(&existing).Error; err != nil {
		return fmt.Errorf("capsule import: %s: probe echo: %w", path, err)
	}
	if existing > 0 {
		// 幂等：不覆盖不合并，其 files / tags / comments 一并跳过。
		s.res.EchoesSkipped++
		return nil
	}

	createdAt, err := capsule.ParseTime(doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("capsule import: %s: created_at: %w", path, err)
	}
	userID, err := s.resolveUserID(doc.Username)
	if err != nil {
		return err
	}
	layout := doc.Layout
	if layout == "" {
		layout = capsule.DefaultLayout
	}

	echo := echoModel.Echo{
		ID:       doc.ID,
		Content:  doc.Content,
		Username: doc.Username,
		Layout:   layout,
		Private:  doc.Private,
		UserID:   userID,
		FavCount: doc.FavCount,
		// CreatedAt 带 autoCreateTime：GORM 只在字段为零值时才代填，显式赋非零值即被原样保留。
		CreatedAt: createdAt,
	}
	// 关联全部手工落地（tags 需 find-or-create、files 需去重与改名），交给 GORM
	// 级联只会绕过这些语义。
	if err := s.db.Omit(clause.Associations).Create(&echo).Error; err != nil {
		return fmt.Errorf("capsule import: %s: create echo: %w", path, err)
	}
	s.res.EchoesCreated++
	s.landed[doc.ID] = struct{}{}

	if err := s.importTags(path, doc); err != nil {
		return err
	}
	if err := s.importFiles(ctx, path, doc, userID); err != nil {
		return err
	}
	return s.importExtension(path, doc)
}

func (s *session) importTags(path string, doc *capsule.EchoDoc) error {
	for _, name := range doc.Tags {
		// 空标签名建不出有意义的 tags 行（name 唯一索引），直接忽略而非造一条 "" 标签。
		if name == "" {
			continue
		}
		tagID, err := s.ensureTag(name)
		if err != nil {
			return fmt.Errorf("capsule import: %s: tag %q: %w", path, name, err)
		}
		// 同一篇里重复列同名标签时 DoNothing 兜底，联合主键不会炸。
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&echoModel.EchoTag{EchoID: doc.ID, TagID: tagID}).Error; err != nil {
			return fmt.Errorf("capsule import: %s: link tag %q: %w", path, name, err)
		}
	}
	return nil
}

func (s *session) ensureTag(name string) (string, error) {
	if id, ok := s.tagIDByName[name]; ok {
		return id, nil
	}

	var tag echoModel.Tag
	err := s.db.Where("name = ?", name).First(&tag).Error
	switch {
	case err == nil:
		s.tagIDByName[name] = tag.ID
		return tag.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return "", err
	}

	// UsageCount 不从胶囊取，导入完统一按 echo_tags 行数重算，这里留零值。
	tag = echoModel.Tag{ID: uuidUtil.MustNewV7(), Name: name}
	if err := s.db.Create(&tag).Error; err != nil {
		return "", err
	}
	s.res.TagsCreated++
	s.tagIDByName[name] = tag.ID
	return tag.ID, nil
}

// recountTagUsage 只重算本次接触过的标签：导入不该顺手改写与本次无关的行。
func (s *session) recountTagUsage() error {
	if len(s.tagIDByName) == 0 {
		return nil
	}
	ids := make([]string, 0, len(s.tagIDByName))
	for _, id := range s.tagIDByName {
		ids = append(ids, id)
	}
	if err := s.db.Exec(
		"UPDATE tags SET usage_count = (SELECT COUNT(*) FROM echo_tags WHERE echo_tags.tag_id = tags.id) WHERE id IN ?",
		ids,
	).Error; err != nil {
		return fmt.Errorf("capsule import: recount tag usage: %w", err)
	}
	return nil
}

func (s *session) importExtension(path string, doc *capsule.EchoDoc) error {
	if doc.Extension == nil {
		return nil
	}
	ext := echoModel.EchoExtension{
		EchoID:  doc.ID,
		Type:    doc.Extension.Type,
		Payload: doc.Extension.Payload,
	}
	if err := s.db.Create(&ext).Error; err != nil {
		return fmt.Errorf("capsule import: %s: create extension: %w", path, err)
	}
	return nil
}

func (s *session) importComments() error {
	if s.loaded.Comments == nil {
		return nil
	}
	hosts, err := s.resolveCommentHosts()
	if err != nil {
		return err
	}

	for i := range s.loaded.Comments.Comments {
		c := &s.loaded.Comments.Comments[i]
		if c.ID == "" {
			return fmt.Errorf("capsule import: %s: comment %d missing id", capsule.CommentsPath, i)
		}

		var existing int64
		if err := s.db.Model(&commentModel.Comment{}).Where("id = ?", c.ID).Count(&existing).Error; err != nil {
			return fmt.Errorf("capsule import: probe comment %s: %w", c.ID, err)
		}
		if existing > 0 {
			// 幂等由评论自己的 id 保证，重复导入不会叠加。
			s.res.CommentsSkipped++
			continue
		}
		if _, ok := hosts[c.EchoID]; !ok {
			// 孤儿的定义是「库里根本没有这个 Echo」，不是「这轮没建这个 Echo」：
			// 往既有 Echo 追加评论不构成对该 Echo 的修改，正是增量导入的主用例。
			// 被 private 排除的 Echo 自然落在同一条规则下——库里没有，其评论即孤儿。
			s.res.OrphanComments++
			continue
		}

		createdAt, err := capsule.ParseTime(c.CreatedAt)
		if err != nil {
			return fmt.Errorf("capsule import: comment %s: created_at: %w", c.ID, err)
		}
		status := c.Status
		if status == "" {
			status = capsule.DefaultCommentStatus
		}
		source := c.Source
		if source == "" {
			source = string(commentModel.SourceGuest)
		}

		// Email / IPHash / UserAgent / UserID 是隐私投影剔除的字段，胶囊里没有也
		// 不该凭空造，一律留零值。
		row := commentModel.Comment{
			ID:        c.ID,
			EchoID:    c.EchoID,
			ParentID:  c.ParentID,
			Nickname:  c.Nickname,
			Website:   c.Website,
			Content:   c.Content,
			Status:    commentModel.Status(status),
			Source:    commentModel.SourceType(source),
			CreatedAt: createdAt,
		}
		if err := s.db.Create(&row).Error; err != nil {
			return fmt.Errorf("capsule import: create comment %s: %w", c.ID, err)
		}
		s.res.CommentsCreated++
	}
	return nil
}

// resolveCommentHosts 一次性查出 comments.yaml 引用到的 echo_id 里哪些真在库中，
// 避免每条评论打一次库。事务内查询，本轮新建的 Echo 也算数；dry-run 下那些行不会
// 真写进去，故并上 landed 兜底。
func (s *session) resolveCommentHosts() (map[string]struct{}, error) {
	referenced := make([]string, 0, len(s.loaded.Comments.Comments))
	seen := make(map[string]struct{}, len(s.loaded.Comments.Comments))
	for i := range s.loaded.Comments.Comments {
		id := s.loaded.Comments.Comments[i].EchoID
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		referenced = append(referenced, id)
	}

	hosts := make(map[string]struct{}, len(referenced))
	for id := range s.landed {
		hosts[id] = struct{}{}
	}
	if len(referenced) == 0 {
		return hosts, nil
	}

	var found []string
	if err := s.db.Model(&echoModel.Echo{}).
		Where("id IN ?", referenced).
		Pluck("id", &found).Error; err != nil {
		return nil, fmt.Errorf("capsule import: probe comment hosts: %w", err)
	}
	for _, id := range found {
		hosts[id] = struct{}{}
	}
	return hosts, nil
}

// applySite 只填空位：目标实例已配置的项一律不动（spec §11.3）。
//
// 「空位」不能只按空字符串判：setting.Get 在 KV 缺键时会返回一份 config 派生的
// 默认值（SiteTitle="Ech0"、ServerLogo="/Ech0.svg"、ServerURL 等都非空），若只认
// 空串，站点身份在全新实例上永远导不进去——而「把站点搬到新实例」正是胶囊的头号
// 用例。故把「当前值 == 未经修改的默认值」也视为空位：没人配过的项让胶囊填，
// 配过的项一个都不动。
func (s *session) applySite(ctx context.Context) error {
	site := s.loaded.Manifest.Site
	current, err := coreSetting.Get(ctx, s.kv, coreSetting.System)
	if err != nil {
		// Get 在后端故障时也返回一份可用默认值；此时回写等于用默认值覆盖真实配置，
		// 宁可整单失败也不能做这件事。
		return fmt.Errorf("capsule import: read system setting: %w", err)
	}

	// 记录的字段名用胶囊/DB 的 json tag，报告与 ech0.yaml 对得上。
	fields := []struct {
		name    string
		target  *string
		capsule string
	}{
		{"site_title", &current.SiteTitle, site.SiteTitle},
		{"server_logo", &current.ServerLogo, site.ServerLogo},
		{"server_name", &current.ServerName, site.ServerName},
		{"server_url", &current.ServerURL, site.ServerURL},
		{"default_locale", &current.DefaultLocale, site.DefaultLocale},
		{"ICP_number", &current.ICPNumber, site.ICPNumber},
		{"footer_content", &current.FooterContent, site.FooterContent},
		{"footer_link", &current.FooterLink, site.FooterLink},
		{"meting_api", &current.MetingAPI, site.MetingAPI},
		{"custom_css", &current.CustomCSS, site.CustomCSS},
		{"custom_js", &current.CustomJS, site.CustomJS},
	}
	pristine := coreSetting.System.Default()
	coreSetting.System.Normalize(&pristine)
	defaults := map[string]string{
		"site_title":     pristine.SiteTitle,
		"server_logo":    pristine.ServerLogo,
		"server_name":    pristine.ServerName,
		"server_url":     pristine.ServerURL,
		"default_locale": pristine.DefaultLocale,
		"ICP_number":     pristine.ICPNumber,
		"footer_content": pristine.FooterContent,
		"footer_link":    pristine.FooterLink,
		"meting_api":     pristine.MetingAPI,
		"custom_css":     pristine.CustomCSS,
		"custom_js":      pristine.CustomJS,
	}

	filled := make([]string, 0, len(fields))
	for _, f := range fields {
		configured := *f.target != "" && *f.target != defaults[f.name]
		if configured || f.capsule == "" || *f.target == f.capsule {
			continue
		}
		*f.target = f.capsule
		filled = append(filled, f.name)
	}
	if len(filled) == 0 {
		return nil
	}
	s.res.SiteFieldsFilled = append(s.res.SiteFieldsFilled, filled...)

	// KV 仓储写完会同步刷进程内缓存，那一步不随事务回滚——dry-run 必须绕开。
	if s.opts.DryRun {
		return nil
	}
	if err := coreSetting.Set(ctx, s.kv, coreSetting.System, current); err != nil {
		return fmt.Errorf("capsule import: write system setting: %w", err)
	}
	return nil
}

// applyConnects 按 connect_url 去重追加：互联列表是运维自主项，导入只补不删不改。
func (s *session) applyConnects() error {
	connects := s.loaded.Manifest.Connects
	if len(connects) == 0 {
		return nil
	}

	var rows []connectModel.Connected
	if err := s.db.Find(&rows).Error; err != nil {
		return fmt.Errorf("capsule import: load connects: %w", err)
	}
	known := make(map[string]struct{}, len(rows))
	for i := range rows {
		known[rows[i].ConnectURL] = struct{}{}
	}

	for _, c := range connects {
		if c.URL == "" {
			continue
		}
		if _, ok := known[c.URL]; ok {
			continue
		}
		if err := s.db.Create(&connectModel.Connected{ConnectURL: c.URL}).Error; err != nil {
			return fmt.Errorf("capsule import: create connect %q: %w", c.URL, err)
		}
		known[c.URL] = struct{}{}
	}
	return nil
}
