// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/internal/storage"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

// heatmapDays 是热力图窗口长度，与活实例 GET /heatmap 一致。
const heatmapDays = 30

// idNamespace 是派生 id 用的 UUID 命名空间。胶囊不携带关系实体的主键
// （Tag / EchoFile 在 spec §11 里被降解成「内容字段」），但前端要按 tagIds
// 过滤、按 id 做列表 key，所以必须现造。用 SHA-1 命名 UUID 而不是随机 UUID：
// 同一个胶囊重复 build 必须得到同一份 dataset，否则前端缓存与外部引用全部失效。
var idNamespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://github.com/lin-snow/Ech0/capsule/build"))

// derivedID 由「种类 + 定位要素」派生稳定 id。用 \x00 分隔避免拼接歧义
// （("ab","c") 与 ("a","bc") 必须派生出不同的 id）。
func derivedID(kind string, parts ...string) string {
	name := kind + "\x00" + strings.Join(parts, "\x00")
	return uuid.NewSHA1(idNamespace, []byte(name)).String()
}

// mediaSchema 与胶囊 / 实例本地存储用的是同一份路由表，
// 所以 <dist>/api/files/ 下的位置与 serve 模式逐字一致，URL 零改写。
var mediaSchema = storage.NewFileSchema()

// renderCategory 决定前端拿到的 category。胶囊里的原值一律不动，这里只为渲染
// 归一：认得的取值原样，不认得的走 NormalizeCategory 的兜底。
//
// category 省略时（手写胶囊常见）按扩展名派生——判据直接取自 mediaSchema 的
// 路由前缀，而不是另抄一份扩展名表，免得两处约定各自漂移。documents/ 与兜底
// files/ 无法进一步区分（.pdf 与 .docx 同前缀），统一落到 file。
func renderCategory(ref capsule.FileRef) string {
	if ref.Category != "" {
		return string(storage.NormalizeCategory(ref.Category))
	}

	name := ref.Key
	if name == "" {
		name = ref.URL
	}
	switch {
	case strings.HasPrefix(mediaSchema.Resolve(path.Base(name)), "images/"):
		return string(storage.CategoryImage)
	case strings.HasPrefix(mediaSchema.Resolve(path.Base(name)), "audios/"):
		return string(storage.CategoryAudio)
	case strings.HasPrefix(mediaSchema.Resolve(path.Base(name)), "videos/"):
		return string(storage.CategoryVideo)
	default:
		return string(storage.CategoryFile)
	}
}

// bakeInput 是烘焙的全部输入：已加载的胶囊 + 归一化后的基址 + 构建时刻。
type bakeInput struct {
	loaded      *capsule.Loaded
	baseURL     string
	generatedAt time.Time
}

// bake 把胶囊转换成 dataset。这是「构建即转换」的落点：所有面向消费者的
// 派生（排序、聚合、URL 计算、默认值填充）都集中在这里，export / import 侧
// 一律原样转储。
func bake(in bakeInput) (*dataset, error) {
	loaded := in.loaded
	if loaded.Manifest == nil {
		return nil, fmt.Errorf("capsule manifest is missing or unreadable")
	}
	site := loaded.Manifest.Site
	now := in.generatedAt.UTC()

	echos, err := bakeEchos(loaded, in.baseURL)
	if err != nil {
		return nil, err
	}

	tags := bakeTags(echos)
	attachTags(echos, tags)

	info := versionPkg.Get()
	title := site.SiteTitle
	if title == "" {
		title = "Ech0"
	}

	ds := &dataset{
		SchemaVersion: datasetSchemaVersion,
		GeneratedAt:   now.Unix(),
		BaseURL:       in.baseURL,
		InitStatus:    initStatus{Initialized: true, OwnerExists: true},
		Settings: settings{
			SiteTitle:  site.SiteTitle,
			ServerLogo: rebaseLogo(site.ServerLogo, in.baseURL),
			ServerName: site.ServerName,
			ServerURL:  site.ServerURL,
			// 静态站没有后端，注册入口必须关死，不能让 UI 渲染出无效表单。
			AllowRegister: false,
			DefaultLocale: site.DefaultLocale,
			ICPNumber:     site.ICPNumber,
			FooterContent: site.FooterContent,
			FooterLink:    site.FooterLink,
			MetingAPI:     site.MetingAPI,
			CustomCSS:     site.CustomCSS,
			CustomJS:      site.CustomJS,
		},
		Hello: hello{
			Hello:     title,
			Copyright: versionPkg.Copyright(),
			Version:   info.Version,
			Commit:    info.Commit,
			BuildTime: info.BuildTime,
			License:   info.License,
			Author:    info.Author,
			RepoURL:   info.RepoURL,
		},
		Agent:       agent{},
		Echos:       echos,
		Tags:        tags,
		Heatmap:     bakeHeatmap(echos, now),
		Comments:    bakeComments(loaded, echos),
		CommentForm: commentForm{},
		Connects:    bakeConnects(loaded.Manifest.Connects),
		Connect: connectInfo{
			ServerName:  site.ServerName,
			ServerURL:   site.ServerURL,
			Logo:        connectLogo(site.ServerLogo, site.ServerURL, in.baseURL),
			TotalEchos:  len(echos),
			TodayEchos:  countOnDay(echos, now),
			SysUsername: loaded.Manifest.Owner.Username,
			Version:     versionPkg.Version,
		},
	}
	return ds, nil
}

// bakeEchos 把胶囊里的 Echo 文档转成 dataset 元素并按 created_at 降序排好。
// private 条目直接剔除：静态站是公开站，把它烘进 dataset 等于泄露。
func bakeEchos(loaded *capsule.Loaded, baseURL string) ([]echo, error) {
	owner := loaded.Manifest.Owner.Username
	out := make([]echo, 0, len(loaded.Echoes))

	for _, le := range loaded.Echoes {
		if le.Err != nil {
			return nil, fmt.Errorf("%s: %w", le.Path, le.Err)
		}
		doc := le.Doc
		if doc == nil || doc.Private {
			continue
		}

		createdAt, err := capsule.ParseTime(doc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: created_at: %w", le.Path, err)
		}

		// 未知/缺失 layout 一律退回默认值：这是「构建即转换」该管的事——胶囊里
		// 逐字保留原值（可能来自旧版本或第三方工具），静态站只负责渲染得出来。
		layout := doc.Layout
		if _, ok := capsule.ValidLayouts[layout]; !ok {
			layout = capsule.DefaultLayout
		}
		username := doc.Username
		if username == "" {
			username = owner
		}

		e := echo{
			ID:        doc.ID,
			Content:   doc.Content,
			Username:  username,
			Layout:    layout,
			Private:   false,
			UserID:    "",
			FavCount:  doc.FavCount,
			CreatedAt: createdAt,
			EchoFiles: bakeFiles(doc, createdAt, baseURL),
			Tags:      nil, // 由 attachTags 在全局聚合出 usage_count 后回填
			Extension: bakeExtension(doc, createdAt),
			tagNames:  dedupeNames(doc.Tags),
		}
		out = append(out, e)
	}

	// 降序是前端查询引擎的默认序；id 做次级键，保证同秒条目的顺序可复现。
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// bakeFiles 展开 frontmatter 的媒体引用。托管文件的 URL 在这里现算——
// 胶囊里不存 URL（存了就会随部署基址过期），它是 baseURL + 路由表的纯函数。
func bakeFiles(doc *capsule.EchoDoc, createdAt int64, baseURL string) []echoFile {
	out := make([]echoFile, 0, len(doc.Files))
	for i, ref := range doc.Files {
		fileID := ref.ID
		if fileID == "" {
			// 胶囊允许省略 file id（它是内部主键）。按定位要素派生，
			// 同一个 key / url 在任何一次 build 里都得到同一个 id。
			if ref.Managed() {
				fileID = derivedID("file", "key", ref.Key)
			} else {
				fileID = derivedID("file", "url", ref.URL)
			}
		}

		f := file{
			ID:          fileID,
			Name:        ref.Name,
			ContentType: ref.ContentType,
			Size:        ref.Size,
			// 与 layout 同理：前端按 category 分支渲染（图/视频/音频/附件），
			// 不认得的取值归一成兜底类别，胶囊里的原值不动。
			Category:  renderCategory(ref),
			Width:     ref.Width,
			Height:    ref.Height,
			CreatedAt: createdAt,
		}
		if ref.Managed() {
			f.Key = ref.Key
			f.StorageType = string(storage.StorageTypeLocal)
			f.URL = baseURL + "api/files/" + mediaSchema.Resolve(ref.Key)
		} else {
			// 外链的 URL 是权威值，没有 key 可以重算，原样透传。
			f.StorageType = string(storage.StorageTypeExternal)
			f.URL = ref.URL
		}

		out = append(out, echoFile{
			ID:        derivedID("echo_file", doc.ID, strconv.Itoa(i)),
			EchoID:    doc.ID,
			FileID:    fileID,
			SortOrder: i,
			File:      f,
		})
	}
	return out
}

// bakeExtension 转换扩展卡片。胶囊不携带扩展的主键与时间戳，
// 前者派生、后者跟随宿主 Echo。
func bakeExtension(doc *capsule.EchoDoc, createdAt int64) *extension {
	if doc.Extension == nil {
		return nil
	}
	payload := doc.Extension.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	return &extension{
		ID:        derivedID("extension", doc.ID),
		EchoID:    doc.ID,
		Type:      doc.Extension.Type,
		Payload:   payload,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

// dedupeNames 去掉空白与重复标签名，保留首次出现的顺序。
// 手写胶囊里重复列同一个标签是常见笔误，不去重会把 usage_count 算多。
func dedupeNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// bakeTags 从所有 Echo 的标签名聚合出全站标签表。id 由名称派生——
// 前端按 tagIds 过滤，同一胶囊重复 build 必须得到同一批 id。
func bakeTags(echos []echo) []tag {
	type acc struct {
		name    string
		count   int
		created int64
	}
	byName := make(map[string]*acc)
	for _, e := range echos {
		for _, name := range e.tagNames {
			a, ok := byName[name]
			if !ok {
				a = &acc{name: name, created: e.CreatedAt}
				byName[name] = a
			}
			a.count++
			// created_at 取最早一次被引用的时刻：比「构建时刻」稳定，
			// 也比 0 有意义（前端标签页会按它展示）。
			if e.CreatedAt < a.created {
				a.created = e.CreatedAt
			}
		}
	}

	out := make([]tag, 0, len(byName))
	for _, a := range byName {
		out = append(out, tag{
			ID:         derivedID("tag", a.name),
			Name:       a.name,
			UsageCount: a.count,
			CreatedAt:  a.created,
		})
	}
	// GET /tags 的契约是按 usage_count 降序；名称做次级键保证可复现。
	sort.Slice(out, func(i, j int) bool {
		if out[i].UsageCount != out[j].UsageCount {
			return out[i].UsageCount > out[j].UsageCount
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// attachTags 把全局标签对象回填进每条 Echo：id 与 usage_count 必须与
// dataset.tags 里的同一份，否则前端「点标签过滤」会匹配不上。
func attachTags(echos []echo, tags []tag) {
	byName := make(map[string]tag, len(tags))
	for _, t := range tags {
		byName[t.Name] = t
	}
	for i := range echos {
		list := make([]tag, 0, len(echos[i].tagNames))
		for _, name := range echos[i].tagNames {
			if t, ok := byName[name]; ok {
				list = append(list, t)
			}
		}
		echos[i].Tags = list
	}
}

// bakeHeatmap 产出以构建当天（UTC）为末端的最近 30 天计数。
func bakeHeatmap(echos []echo, now time.Time) []heatmapEntry {
	counts := make(map[string]int, len(echos))
	for _, e := range echos {
		counts[utcDay(e.CreatedAt)]++
	}
	end := now.UTC().Truncate(24 * time.Hour)
	out := make([]heatmapEntry, 0, heatmapDays)
	for i := heatmapDays - 1; i >= 0; i-- {
		day := end.AddDate(0, 0, -i).Format(time.DateOnly)
		out = append(out, heatmapEntry{Date: day, Count: counts[day]})
	}
	return out
}

// bakeComments 转换评论。只保留宿主 Echo 仍在 dataset 里的条目：
// private Echo 被剔除后，它的评论若留下会变成前端永远取不到宿主的孤儿。
func bakeComments(loaded *capsule.Loaded, echos []echo) []comment {
	if loaded.Comments == nil {
		return []comment{}
	}
	known := make(map[string]struct{}, len(echos))
	for _, e := range echos {
		known[e.ID] = struct{}{}
	}

	out := make([]comment, 0, len(loaded.Comments.Comments))
	for _, c := range loaded.Comments.Comments {
		if _, ok := known[c.EchoID]; !ok {
			continue
		}
		// 时间解析失败不该炸掉整次构建：单条评论坏了就退化成 0（check 会先报错）。
		createdAt, _ := capsule.ParseTime(c.CreatedAt)

		status := c.Status
		if status == "" {
			status = capsule.DefaultCommentStatus
		}
		source := c.Source
		if source == "" {
			source = "guest"
		}

		out = append(out, comment{
			ID:       c.ID,
			EchoID:   c.EchoID,
			ParentID: c.ParentID,
			// 胶囊里禁止出现 user_id / email（spec §5）：静态站的评论一律匿名投影。
			UserID:    nil,
			Nickname:  c.Nickname,
			Email:     "",
			Website:   c.Website,
			Content:   c.Content,
			Status:    status,
			Hot:       false,
			Source:    source,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		})
	}
	// GET /comments 的契约是按 created_at 升序。
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt < out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// bakeConnects 转换互联实例列表；胶囊只存 url，id 由 url 派生。
func bakeConnects(list []capsule.Connect) []connectItem {
	out := make([]connectItem, 0, len(list))
	for _, c := range list {
		out = append(out, connectItem{
			ID:         derivedID("connect", c.URL),
			ConnectURL: c.URL,
		})
	}
	return out
}

// countOnDay 统计落在 now 所在 UTC 日的 Echo 数（connect.today_echos 的冻结值）。
func countOnDay(echos []echo, now time.Time) int {
	day := now.UTC().Format(time.DateOnly)
	n := 0
	for _, e := range echos {
		if utcDay(e.CreatedAt) == day {
			n++
		}
	}
	return n
}

func utcDay(sec int64) string {
	return time.Unix(sec, 0).UTC().Format(time.DateOnly)
}

// rebaseLogo 把托管 logo 的绝对根路径迁到部署基址下。其余取值（外链、
// 相对路径、空）原样透传——胶囊里是什么就是什么。
func rebaseLogo(logo, baseURL string) string {
	const managedPrefix = "/api/files/"
	if baseURL == "/" || !strings.HasPrefix(logo, managedPrefix) {
		return logo
	}
	return baseURL + strings.TrimPrefix(logo, "/")
}

// connectLogo 归一化 Connect 名片里的 logo。这份载荷是给**远端实例**消费的
// （api/connect 与 GET /connect 同形），相对路径在对方页面上必然解析错，故必须
// 绝对化。分支逐字对齐活实例的 ConnectService.GetConnect，保证两边同形。
// server_url 缺失时无从绝对化，退回本站可用的相对路径。
func connectLogo(logo, serverURL, baseURL string) string {
	rebased := rebaseLogo(logo, baseURL)
	origin := strings.TrimRight(strings.TrimSpace(serverURL), "/")
	if origin == "" {
		return rebased
	}

	path := strings.TrimSpace(rebased)
	switch {
	case path == "" || path == "Ech0.svg" || path == "/Ech0.svg":
		return origin + "/Ech0.svg"
	case strings.HasPrefix(path, "http://"), strings.HasPrefix(path, "https://"):
		return path
	case strings.HasPrefix(path, "/"):
		return origin + path
	default:
		return origin + "/" + path
	}
}
