// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// FileDTO is the wire shape of a file resource: what a client may see, never where
// the bytes live. Storage ids, object keys and ref URLs stay internal; clients
// address bytes through URL and nothing else.
type FileDTO struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Access      string    `json:"access"`
	Status      string    `json:"status"`
	Name        string    `json:"name,omitempty"`
	ContentType string    `json:"content_type"`
	Category    string    `json:"category"`
	Size        int64     `json:"size"`
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	URL         string    `json:"url"`
	UploaderID  string    `json:"uploader_id,omitempty"`
	CreatedAt   int64     `json:"created_at"`
	Links       []LinkDTO `json:"links,omitempty"`
}

// LinkDTO reports where a file is attached, which is what a delete conflict
// must tell its caller.
type LinkDTO struct {
	OwnerKind string `json:"owner_kind"`
	OwnerID   string `json:"owner_id"`
	Role      string `json:"role"`
	SortOrder int    `json:"sort_order"`
}

// Page is one page of the file browser.
type Page struct {
	Total int64     `json:"total"`
	Items []FileDTO `json:"items"`
}

// DirectUpload is the answer to a presigned upload request: where to PUT, with
// which headers, and until when.
type DirectUpload struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt int64             `json:"expires_at"`
}

// Reference is one attachment of a file that blocked its deletion.
type Reference struct {
	OwnerKind string `json:"owner_kind"`
	OwnerID   string `json:"owner_id"`
	Role      string `json:"role"`
}
