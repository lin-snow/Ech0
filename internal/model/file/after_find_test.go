// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type afterFindParent struct {
	ID        string     `gorm:"type:char(36);primaryKey"`
	EchoFiles []EchoFile `gorm:"foreignKey:EchoID"`
}

func openAfterFindDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&afterFindParent{}, &EchoFile{}, &File{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAfterFindRecomputesManagedURL(t *testing.T) {
	db := openAfterFindDB(t)

	RegisterURLResolver(func(_, key string) string { return "https://cdn.new/" + key })
	t.Cleanup(func() { RegisterURLResolver(nil) })

	local := &File{Key: "a.png", StorageType: "local", URL: "https://cdn.OLD/a.png", Name: "a.png", Category: "image", UserID: "u1"}
	external := &File{Key: "", StorageType: storageTypeExternal, URL: "https://other.site/x.png", Name: "x.png", Category: "image", UserID: "u1"}
	if err := db.Create(local).Error; err != nil {
		t.Fatalf("create local: %v", err)
	}
	if err := db.Create(external).Error; err != nil {
		t.Fatalf("create external: %v", err)
	}

	var direct File
	if err := db.First(&direct, "id = ?", local.ID).Error; err != nil {
		t.Fatalf("load local: %v", err)
	}
	if direct.URL != "https://cdn.new/a.png" {
		t.Fatalf("direct load: want recomputed url, got %q", direct.URL)
	}

	var ext File
	if err := db.First(&ext, "id = ?", external.ID).Error; err != nil {
		t.Fatalf("load external: %v", err)
	}
	if ext.URL != "https://other.site/x.png" {
		t.Fatalf("external load: want stored url kept, got %q", ext.URL)
	}

	parent := &afterFindParent{ID: "p1"}
	if err := db.Create(parent).Error; err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if err := db.Create(&EchoFile{EchoID: parent.ID, FileID: local.ID}).Error; err != nil {
		t.Fatalf("create echo file: %v", err)
	}
	var loaded afterFindParent
	if err := db.Preload("EchoFiles.File").First(&loaded, "id = ?", parent.ID).Error; err != nil {
		t.Fatalf("preload parent: %v", err)
	}
	if len(loaded.EchoFiles) != 1 {
		t.Fatalf("want 1 echo file, got %d", len(loaded.EchoFiles))
	}
	if got := loaded.EchoFiles[0].File.URL; got != "https://cdn.new/a.png" {
		t.Fatalf("nested preload: AfterFind did not refresh url, got %q", got)
	}
}

func TestAfterFindNoResolverKeepsSnapshot(t *testing.T) {
	db := openAfterFindDB(t)
	RegisterURLResolver(nil)

	f := &File{Key: "b.png", StorageType: "object", URL: "https://snapshot/b.png", Name: "b.png", Category: "image", UserID: "u1"}
	if err := db.Create(f).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	var got File
	if err := db.First(&got, "id = ?", f.ID).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.URL != "https://snapshot/b.png" {
		t.Fatalf("want snapshot kept, got %q", got.URL)
	}
}
