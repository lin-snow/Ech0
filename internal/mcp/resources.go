// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

type ResourceDefinition struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourcesListResult struct {
	ResultEnvelope
	CacheInfo
	Resources []ResourceDefinition `json:"resources"`
}

type ResourceReadParams struct {
	URI string `json:"uri"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

type ResourceReadResult struct {
	ResultEnvelope
	CacheInfo
	Contents []ResourceContent `json:"contents"`
}
