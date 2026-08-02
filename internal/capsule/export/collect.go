// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package export

import (
	"context"
	"fmt"

	"github.com/lin-snow/ech0/internal/capsule"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"gorm.io/gorm"
)

// dataset 是一次导出的全部库内容快照。先整体读出再整体写盘，避免边读边写时
// 「Echo 已写、其引用的 file 行刚被删」这类撕裂。
type dataset struct {
	echoes   []echoModel.Echo
	files    []fileModel.File // 需要写进胶囊的 files 表记录（含 external 行）
	comments []capsule.Comment
	site     capsule.Site
	owner    capsule.Owner
	connects []capsule.Connect

	skippedPrivate int
	externalFiles  int
}

func collect(ctx context.Context, deps Deps, opts Options) (*dataset, error) {
	db := deps.DB.WithContext(ctx)
	data := &dataset{}

	if err := collectEchoes(db, opts, data); err != nil {
		return nil, err
	}
	if err := collectFiles(db, opts, data); err != nil {
		return nil, err
	}
	if err := collectComments(db, data); err != nil {
		return nil, err
	}
	if err := collectSite(ctx, deps, data); err != nil {
		return nil, err
	}
	if err := collectOwner(db, data); err != nil {
		return nil, err
	}
	return data, collectConnects(db, data)
}

// collectEchoes 按创建时间升序读全量 Echo。EchoFiles 必须带 sort_order 排序读出——
// 展示顺序在胶囊里由 files 数组顺序表达（spec §4.2），预加载的顺序就是胶囊的顺序。
func collectEchoes(db *gorm.DB, opts Options, data *dataset) error {
	query := db.
		Preload("EchoFiles", func(d *gorm.DB) *gorm.DB {
			return d.Order("echo_files.sort_order ASC")
		}).
		Preload("EchoFiles.File").
		Preload("Extension").
		Preload("Tags").
		Order("created_at ASC")
	if !opts.IncludePrivate {
		query = query.Where("private = ?", false)
	}
	if err := query.Find(&data.echoes).Error; err != nil {
		return fmt.Errorf("capsule export: load echoes: %w", err)
	}

	if !opts.IncludePrivate {
		var private int64
		if err := db.Model(&echoModel.Echo{}).Where("private = ?", true).Count(&private).Error; err != nil {
			return fmt.Errorf("capsule export: count private echoes: %w", err)
		}
		data.skippedPrivate = int(private)
	}
	return nil
}

// collectFiles 以 files 表记录为准挑出待导出的媒体（spec §6：记录驱动，禁止盲拷
// DataRoot）。未被任何 Echo 引用的悬空文件照常导出——它合法，check 侧只告警。
func collectFiles(db *gorm.DB, opts Options, data *dataset) error {
	var files []fileModel.File
	if err := db.Find(&files).Error; err != nil {
		return fmt.Errorf("capsule export: load files: %w", err)
	}

	if !opts.IncludePrivate {
		hidden, err := privateOnlyFiles(db)
		if err != nil {
			return err
		}
		kept := files[:0]
		for _, f := range files {
			if _, skip := hidden[f.ID]; skip {
				continue
			}
			kept = append(kept, f)
		}
		files = kept
	}

	for i := range files {
		if storage.NormalizeStorageType(files[i].StorageType) == storage.StorageTypeExternal {
			data.externalFiles++
		}
	}
	data.files = files
	return nil
}

// privateOnlyFiles 返回「仅被 private Echo 引用」的文件 id 集合。这些字节不能随
// 公开胶囊出门；只要还有任一公开 Echo 引用它，它就是公开内容的一部分，必须导出。
func privateOnlyFiles(db *gorm.DB) (map[string]struct{}, error) {
	var refs []struct {
		FileID  string
		Private bool
	}
	if err := db.Model(&fileModel.EchoFile{}).
		Select("echo_files.file_id AS file_id, echos.private AS private").
		Joins("JOIN echos ON echos.id = echo_files.echo_id").
		Scan(&refs).Error; err != nil {
		return nil, fmt.Errorf("capsule export: resolve file visibility: %w", err)
	}

	privateOnly := make(map[string]struct{})
	public := make(map[string]struct{})
	for _, ref := range refs {
		if ref.Private {
			privateOnly[ref.FileID] = struct{}{}
			continue
		}
		public[ref.FileID] = struct{}{}
	}
	for id := range public {
		delete(privateOnly, id)
	}
	return privateOnly, nil
}

// collectComments 只导出已通过审核的评论，且只保留指向本次导出 Echo 集合的那些：
// private Echo 被排除后，它名下的评论就是孤儿，带出去只会让消费者报警告。
// 出胶囊前必过 Public 投影，隐私字段（email/ip_hash/user_agent/user_id）在此剥离。
func collectComments(db *gorm.DB, data *dataset) error {
	var comments []commentModel.Comment
	if err := db.
		Where("status = ?", commentModel.StatusApproved).
		Order("created_at asc").
		Find(&comments).Error; err != nil {
		return fmt.Errorf("capsule export: load comments: %w", err)
	}

	exported := make(map[string]struct{}, len(data.echoes))
	for i := range data.echoes {
		exported[data.echoes[i].ID] = struct{}{}
	}

	for i := range comments {
		if _, ok := exported[comments[i].EchoID]; !ok {
			continue
		}
		public := commentModel.ToPublicComment(comments[i])
		data.comments = append(data.comments, capsule.Comment{
			ID:        public.ID,
			EchoID:    public.EchoID,
			ParentID:  public.ParentID,
			Nickname:  public.Nickname,
			Website:   public.Website,
			Content:   public.Content,
			Status:    string(public.Status),
			Source:    string(public.Source),
			CreatedAt: capsule.FormatUnix(public.CreatedAt),
		})
	}
	return nil
}

// collectSite 逐字段拷贝站点设置的公开子集。这里不用整体序列化：AllowRegister 是
// 运维行为开关，必须留在库里（spec §3）；逐字段列出让「哪些进了胶囊」一眼可查。
func collectSite(ctx context.Context, deps Deps, data *dataset) error {
	system, err := coreSetting.Get(ctx, deps.KV, coreSetting.System)
	if err != nil {
		return fmt.Errorf("capsule export: load system setting: %w", err)
	}
	data.site = capsule.Site{
		SiteTitle:     system.SiteTitle,
		ServerLogo:    system.ServerLogo,
		ServerName:    system.ServerName,
		ServerURL:     system.ServerURL,
		DefaultLocale: system.DefaultLocale,
		ICPNumber:     system.ICPNumber,
		FooterContent: system.FooterContent,
		FooterLink:    system.FooterLink,
		MetingAPI:     system.MetingAPI,
		CustomCSS:     system.CustomCSS,
		CustomJS:      system.CustomJS,
	}
	return nil
}

// collectOwner 取站长。owner.username 是清单的必须字段（Echo 未标 username 时的
// 归属兜底），取不到就没有合法胶囊可产，直接失败而非留空。
func collectOwner(db *gorm.DB, data *dataset) error {
	var owner userModel.User
	if err := db.Where("is_owner = ?", true).First(&owner).Error; err != nil {
		return fmt.Errorf("capsule export: load owner user: %w", err)
	}
	data.owner = capsule.Owner{Username: owner.Username}
	return nil
}

func collectConnects(db *gorm.DB, data *dataset) error {
	var connects []connectModel.Connected
	if err := db.Find(&connects).Error; err != nil {
		return fmt.Errorf("capsule export: load connects: %w", err)
	}
	data.connects = make([]capsule.Connect, 0, len(connects))
	for i := range connects {
		data.connects = append(data.connects, capsule.Connect{URL: connects[i].ConnectURL})
	}
	return nil
}
