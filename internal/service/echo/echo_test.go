// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"

	model "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
)

func TestNormalizeEchoExtension(t *testing.T) {
	tests := []struct {
		name    string
		input   *model.EchoExtension
		wantNil bool
		wantErr bool
	}{
		{
			name:    "nil extension",
			input:   nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name: "music extension",
			input: &model.EchoExtension{
				Type: model.Extension_MUSIC,
				Payload: map[string]any{
					"url": "https://music.163.com/#/song?id=123",
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "website missing title",
			input: &model.EchoExtension{
				Type: model.Extension_WEBSITE,
				Payload: map[string]any{
					"site": "https://example.com",
				},
			},
			wantNil: false,
			wantErr: true,
		},
		{
			name: "unsupported type",
			input: &model.EchoExtension{
				Type: "UNKNOWN",
				Payload: map[string]any{
					"foo": "bar",
				},
			},
			wantNil: false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeEchoExtension(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantNil && got != nil {
				t.Fatalf("expected nil extension, got %#v", got)
			}
			if !tc.wantNil && !tc.wantErr && got == nil {
				t.Fatalf("expected non-nil extension")
			}
		})
	}
}

func TestIsEchoEmpty(t *testing.T) {
	tests := []struct {
		name  string
		echo  *model.Echo
		empty bool
	}{
		{
			name:  "nil echo is empty",
			echo:  nil,
			empty: true,
		},
		{
			name: "whitespace content with no file and no extension is empty",
			echo: &model.Echo{
				Content: "   \n\t   ",
			},
			empty: true,
		},
		{
			name: "whitespace content with files is not empty",
			echo: &model.Echo{
				Content: "   ",
				EchoFiles: []fileModel.EchoFile{
					{FileID: "test-file-id"},
				},
			},
			empty: false,
		},
		{
			name: "whitespace content with extension is not empty",
			echo: &model.Echo{
				Content: "  ",
				Extension: &model.EchoExtension{
					Type: model.Extension_MUSIC,
					Payload: map[string]any{
						"url": "https://example.com/song",
					},
				},
			},
			empty: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isEchoEmpty(tc.echo)
			if got != tc.empty {
				t.Fatalf("expected %v, got %v", tc.empty, got)
			}
		})
	}
}
