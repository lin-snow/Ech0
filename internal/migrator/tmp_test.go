// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTmpDir(t *testing.T) {
	t.Run("valid tmp dir", func(t *testing.T) {
		path, ok := resolveTmpDir(map[string]any{
			"tmp_dir": "files/tmp/ech0_v4_test",
		})
		if !ok {
			t.Fatalf("expected valid tmp dir")
		}
		expected := filepath.Clean(filepath.Join("data", "files/tmp/ech0_v4_test"))
		if path != expected {
			t.Fatalf("unexpected path: got=%s want=%s", path, expected)
		}
	})

	t.Run("reject path traversal", func(t *testing.T) {
		if _, ok := resolveTmpDir(map[string]any{
			"tmp_dir": "../outside",
		}); ok {
			t.Fatalf("expected traversal path to be rejected")
		}
	})

	t.Run("reject non tmp subtree", func(t *testing.T) {
		if _, ok := resolveTmpDir(map[string]any{
			"tmp_dir": "files/static",
		}); ok {
			t.Fatalf("expected non tmp dir to be rejected")
		}
	})
}

func TestCleanupTmpDirFromPayload(t *testing.T) {
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	tmpName := "cleanup_test_dir"
	targetDir := filepath.Join("data", "files/tmp", tmpName)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("create target dir failed: %v", err)
	}
	testFile := filepath.Join(targetDir, "sample.txt")
	if err := os.WriteFile(testFile, []byte("sample"), 0o644); err != nil {
		t.Fatalf("write sample file failed: %v", err)
	}

	if err := CleanupTmpDirFromPayload(map[string]any{
		"tmp_dir": filepath.ToSlash(filepath.Join("files/tmp", tmpName)),
	}); err != nil {
		t.Fatalf("CleanupTmpDirFromPayload failed: %v", err)
	}

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Fatalf("expected target dir removed, stat err=%v", err)
	}
}
