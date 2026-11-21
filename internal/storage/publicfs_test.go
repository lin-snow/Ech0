// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicFSServesMediaOnly(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"images/a.png":                           "png",
		"images/nested/b.jpg":                    "jpg",
		"audios/a.mp3":                           "mp3",
		"videos/a.mp4":                           "mp4",
		"documents/a.pdf":                        "pdf",
		"files/anything.bin":                     "bin",
		"snapshots/ech0_snapshot_2026-01-01.zip": "secret",
		"tmp/ech0_v4/ech0.db":                    "db",
		"capsules/c.zip":                         "capsule",
		"ech0.db":                                "db",
		"app.log":                                "log",
	}
	for rel, body := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	public := NewPublicFS(root)

	openable := []string{"/images/a.png", "/images/nested/b.jpg", "/audios/a.mp3", "/videos/a.mp4", "/documents/a.pdf", "/files/anything.bin"}
	for _, name := range openable {
		file, err := public.Open(name)
		if err != nil {
			t.Fatalf("Open(%q): %v", name, err)
		}
		_ = file.Close()
	}

	refused := []string{
		"/",
		"/snapshots/ech0_snapshot_2026-01-01.zip",
		"/tmp/ech0_v4/ech0.db",
		"/capsules/c.zip",
		"/ech0.db",
		"/app.log",
		"/images",        // directory
		"/images/",       // directory with trailing slash
		"/images/nested", // nested directory
		"/images/../snapshots/ech0_snapshot_2026-01-01.zip",
		"/images/../../app.log",
		"/imagesx/a.png", // missing separator after prefix
		"images/a.png",   // relative name
	}
	for _, name := range refused {
		if file, err := public.Open(name); !errors.Is(err, fs.ErrNotExist) {
			if file != nil {
				_ = file.Close()
			}
			t.Fatalf("Open(%q) = %v, want ErrNotExist", name, err)
		}
	}
}

func TestPublicFSHTTPResponses(t *testing.T) {
	root := t.TempDir()
	for rel, body := range map[string]string{
		"images/a.png":                     "png",
		"snapshots/ech0_snapshot_2026.zip": "secret",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	server := httptest.NewServer(http.FileServer(NewPublicFS(root)))
	t.Cleanup(server.Close)

	cases := []struct {
		path string
		want int
	}{
		{"/images/a.png", http.StatusOK},
		{"/snapshots/ech0_snapshot_2026.zip", http.StatusNotFound},
		{"/snapshots/", http.StatusNotFound},
		{"/", http.StatusNotFound},
	}
	for _, tc := range cases {
		resp, err := server.Client().Get(server.URL + tc.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Fatalf("GET %s = %d, want %d", tc.path, resp.StatusCode, tc.want)
		}
	}
}
