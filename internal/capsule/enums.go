// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/storage"
)

// 枚举词表一律引用领域常量而非字面量重抄一遍：胶囊取值就是库中取值，
// 领域侧新增成员时这里自动跟随，不会漂移。

// ValidLayouts 是 Echo.Layout 的合法取值。
var ValidLayouts = map[string]struct{}{
	echoModel.LayoutWaterfall:  {},
	echoModel.LayoutGrid:       {},
	echoModel.LayoutHorizontal: {},
	echoModel.LayoutCarousel:   {},
	echoModel.LayoutStack:      {},
	echoModel.LayoutNone:       {},
}

// DefaultLayout 是 frontmatter 缺省 layout 时的取值（与 GORM 列默认值一致）。
const DefaultLayout = echoModel.LayoutWaterfall

// ValidExtensionTypes 是 EchoExtension.Type 的合法取值。
var ValidExtensionTypes = map[string]struct{}{
	echoModel.Extension_MUSIC:      {},
	echoModel.Extension_VIDEO:      {},
	echoModel.Extension_GITHUBPROJ: {},
	echoModel.Extension_WEBSITE:    {},
	echoModel.Extension_LOCATION:   {},
	echoModel.Extension_TWEET:      {},
}

// ValidCategories 是 File.Category 的合法取值（storage.Category 全集）。
var ValidCategories = map[string]struct{}{
	string(storage.CategoryImage):    {},
	string(storage.CategoryVideo):    {},
	string(storage.CategoryAudio):    {},
	string(storage.CategoryPDF):      {},
	string(storage.CategoryMarkdown): {},
	string(storage.CategoryFile):     {},
}

// DefaultCommentStatus 是胶囊内评论应当具备的状态；其余状态仅告警。
const DefaultCommentStatus = string(commentModel.StatusApproved)
