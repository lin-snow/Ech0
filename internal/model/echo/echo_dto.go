// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type EchoExtensionDto struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type CreateTagDto struct {
	Name string `json:"name" binding:"required"`
}

type EchoUpsertDto struct {
	ID        string            `json:"id,omitempty"`
	Content   string            `json:"content"`
	EchoFiles []EchoFile        `json:"echo_files,omitempty" swaggertype:"array,object"`
	Layout    string            `json:"layout,omitempty"`
	Private   bool              `json:"private"`
	Extension *EchoExtensionDto `json:"extension,omitempty"`
	Tags      []Tag             `json:"tags,omitempty"`
	CreatedAt *int64            `json:"created_at,omitempty"`
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
