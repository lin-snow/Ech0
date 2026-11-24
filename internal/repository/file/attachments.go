// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"gorm.io/gorm"
)

// EchoOwnerKind is the owner_kind an echo stores its attachments under.
const EchoOwnerKind = fileModel.OwnerKindEcho

// EchoAttachmentRole is the link role of an echo attachment. Other roles
// (cover, …) are single-slot slots and never lists.
const EchoAttachmentRole = fileModel.RoleAttachment

// HydrateEchoAttachments fills the attachment projection of a batch of echoes
// from file_links + files in two queries total. The echo table itself holds no
// attachment rows, so every read path — cache fill, search, page, single — calls
// this once after loading its echoes.
func HydrateEchoAttachments(ctx context.Context, db *gorm.DB, echoes []echoModel.Echo) error {
	if len(echoes) == 0 {
		return nil
	}
	ids := make([]string, 0, len(echoes))
	for i := range echoes {
		ids = append(ids, echoes[i].ID)
	}
	var links []fileModel.Link
	if err := db.WithContext(ctx).
		Where("owner_kind = ? AND role = ? AND owner_id IN ?",
			EchoOwnerKind, EchoAttachmentRole, ids).
		Order("owner_id, sort_order").
		Find(&links).Error; err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}
	fileIDs := make([]string, 0, len(links))
	for _, link := range links {
		fileIDs = append(fileIDs, link.FileID)
	}
	var files []fileModel.File
	if err := db.WithContext(ctx).Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return err
	}
	byID := make(map[string]fileModel.File, len(files))
	for i := range files {
		byID[files[i].ID] = files[i]
	}
	byOwner := make(map[string][]echoModel.EchoFile, len(echoes))
	for _, link := range links {
		file, ok := byID[link.FileID]
		if !ok {
			continue
		}
		byOwner[link.OwnerID] = append(byOwner[link.OwnerID], echoModel.EchoFile{
			FileID:    link.FileID,
			SortOrder: link.SortOrder,
			File:      file,
		})
	}
	for i := range echoes {
		if attachments, ok := byOwner[echoes[i].ID]; ok {
			echoes[i].EchoFiles = attachments
		}
	}
	return nil
}
