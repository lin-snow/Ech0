// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type EchoExtensionDto struct {
	Type    string         `json:"type" doc:"Card type, e.g. weather or music"`
	Payload map[string]any `json:"payload" doc:"Arbitrary fields the card needs to render"`
}

type CreateTagDto struct {
	Name string `json:"name" binding:"required" doc:"标签名"`
}

type EchoUpsertDto struct {
	ID        string            `json:"id,omitempty" doc:"Echo ID to update; ignored on create"`
	Content   string            `json:"content" doc:"Markdown body; supply at least one of content, files, or extension"`
	EchoFiles []EchoFile        `json:"echo_files,omitempty" swaggertype:"array,object" doc:"Attachments as {file_id, sort_order}; upload files first to get file_id"`
	Layout    string            `json:"layout,omitempty" doc:"Layout: waterfall / grid / horizontal / carousel / stack / none"`
	Private   bool              `json:"private" doc:"Visible only to the owner"`
	Extension *EchoExtensionDto `json:"extension,omitempty" doc:"Card payload: {type, payload}"`
	Tags      []Tag             `json:"tags,omitempty" doc:"Tags as {name}; an existing tag with the same name is reused"`
	CreatedAt *int64            `json:"created_at,omitempty" doc:"Publish time in Unix seconds; defaults to now"`
}

func (dto *EchoUpsertDto) ToModel() *Echo {
	echo := &Echo{
		ID:        dto.ID,
		Content:   dto.Content,
		EchoFiles: dto.EchoFiles,
		Layout:    dto.Layout,
		Private:   dto.Private,
		Tags:      dto.Tags,
	}
	if dto.CreatedAt != nil {
		echo.CreatedAt = *dto.CreatedAt
	}
	if dto.Extension != nil {
		echo.Extension = &EchoExtension{
			Type:    dto.Extension.Type,
			Payload: dto.Extension.Payload,
		}
	}
	return echo
}
