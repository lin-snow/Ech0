// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterStaticFilesServesOnlyMedia(t *testing.T) {
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	for rel, body := range map[string]string{
		"images/a.png": "png",
		"snapshots/ech0_snapshot_2026-01-01_00-00-00.zip": "secret",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	engine := gin.New()
	registerStaticFiles(engine, root)

	cases := []struct {
		path string
		want int
	}{
		{"/api/files/images/a.png", http.StatusOK},
		{"/api/files/snapshots/ech0_snapshot_2026-01-01_00-00-00.zip", http.StatusNotFound},
		{"/api/files/snapshots/", http.StatusNotFound},
		{"/api/files/", http.StatusNotFound},
		{"/api/files/ech0.db", http.StatusNotFound},
	}
	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if recorder.Code != tc.want {
			t.Fatalf("GET %s = %d, want %d", tc.path, recorder.Code, tc.want)
		}
	}
}
