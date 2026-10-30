// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type EchoExtensionDto struct {
	Type    string         `json:"type" doc:"扩展卡片类型，如 weather / music"`
	Payload map[string]any `json:"payload" doc:"扩展渲染所需的任意字段"`
}

type CreateTagDto struct {
	Name string `json:"name" binding:"required" doc:"标签名"`
}

type EchoUpsertDto struct {
	ID        string            `json:"id,omitempty" doc:"更新时的 Echo ID；创建时忽略"`
	Content   string            `json:"content" doc:"正文（Markdown）；与附件、扩展至少有一项"`
	EchoFiles []EchoFile        `json:"echo_files,omitempty" swaggertype:"array,object" doc:"附件引用列表：{file_id, sort_order}；文件先通过上传接口获取 file_id"`
	Layout    string            `json:"layout,omitempty" doc:"布局：waterfall / grid / horizontal / carousel / stack / none"`
	Private   bool              `json:"private" doc:"仅自己可见"`
	Extension *EchoExtensionDto `json:"extension,omitempty" doc:"扩展卡片 {type, payload}"`
	Tags      []Tag             `json:"tags,omitempty" doc:"标签列表，元素为 {name}；同名标签复用"`
	CreatedAt *int64            `json:"created_at,omitempty" doc:"发布时间（Unix 秒），缺省为当前时间"`
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
