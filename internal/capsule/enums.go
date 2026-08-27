// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/storage"
)

var ValidLayouts = map[string]struct{}{
	echoModel.LayoutWaterfall:  {},
	echoModel.LayoutGrid:       {},
	echoModel.LayoutHorizontal: {},
	echoModel.LayoutCarousel:   {},
	echoModel.LayoutStack:      {},
	echoModel.LayoutNone:       {},
}

const DefaultLayout = echoModel.LayoutWaterfall

var ValidExtensionTypes = map[string]struct{}{
	echoModel.Extension_MUSIC:      {},
	echoModel.Extension_VIDEO:      {},
	echoModel.Extension_GITHUBPROJ: {},
	echoModel.Extension_WEBSITE:    {},
	echoModel.Extension_LOCATION:   {},
	echoModel.Extension_TWEET:      {},
}

var ValidCategories = map[string]struct{}{
	string(storage.CategoryImage):    {},
	string(storage.CategoryVideo):    {},
	string(storage.CategoryAudio):    {},
	string(storage.CategoryPDF):      {},
	string(storage.CategoryMarkdown): {},
	string(storage.CategoryFile):     {},
}

const DefaultCommentStatus = string(commentModel.StatusApproved)
