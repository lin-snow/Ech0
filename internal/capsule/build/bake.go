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

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/internal/storage"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

const heatmapDays = 30

var idNamespace = uuidUtil.NewV5(uuidUtil.NameSpaceURL, []byte("https://github.com/lin-snow/Ech0/capsule/build"))

func derivedID(kind string, parts ...string) string {
	name := kind + "\x00" + strings.Join(parts, "\x00")
	return uuidUtil.NewV5(idNamespace, []byte(name)).String()
}

var mediaSchema = storage.NewFileSchema()

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

type bakeInput struct {
	loaded      *capsule.Loaded
	baseURL     string
	generatedAt time.Time
}

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
			SiteTitle:     site.SiteTitle,
			ServerLogo:    rebaseLogo(site.ServerLogo, in.baseURL),
			ServerName:    site.ServerName,
			ServerURL:     site.ServerURL,
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
			Tags:      nil,
			Extension: bakeExtension(doc, createdAt),
			tagNames:  dedupeNames(doc.Tags),
		}
		out = append(out, e)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func bakeFiles(doc *capsule.EchoDoc, createdAt int64, baseURL string) []echoFile {
	out := make([]echoFile, 0, len(doc.Files))
	for i, ref := range doc.Files {
		fileID := ref.ID
		if fileID == "" {
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
			Category:    renderCategory(ref),
			Width:       ref.Width,
			Height:      ref.Height,
			CreatedAt:   createdAt,
		}
		if ref.Managed() {
			f.Key = ref.Key
			f.StorageType = string(storage.StorageTypeLocal)
			f.URL = baseURL + "api/files/" + mediaSchema.Resolve(ref.Key)
		} else {
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
	sort.Slice(out, func(i, j int) bool {
		if out[i].UsageCount != out[j].UsageCount {
			return out[i].UsageCount > out[j].UsageCount
		}
		return out[i].Name < out[j].Name
	})
	return out
}

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
			ID:        c.ID,
			EchoID:    c.EchoID,
			ParentID:  c.ParentID,
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
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt < out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}

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

func rebaseLogo(logo, baseURL string) string {
	const managedPrefix = "/api/files/"
	if baseURL == "/" || !strings.HasPrefix(logo, managedPrefix) {
		return logo
	}
	return baseURL + strings.TrimPrefix(logo, "/")
}

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
