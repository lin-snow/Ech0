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

type dataset struct {
	echoes   []echoModel.Echo
	files    []fileModel.File
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
		HomeLayout:    system.HomeLayout,
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
