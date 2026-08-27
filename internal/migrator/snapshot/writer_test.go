// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreate_KeepOnlyLatestAndExcludeSnapshotDir(t *testing.T) {
	workspace := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dataDir, "files/images"), 0o755); err != nil {
		t.Fatalf("mkdir images failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, snapshotRelativeDir), 0o755); err != nil {
		t.Fatalf("mkdir snapshot dir failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "ech0.db"), []byte("db"), 0o644); err != nil {
		t.Fatalf("write ech0.db failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "files/images/demo.txt"), []byte("demo"), 0o644); err != nil {
		t.Fatalf("write image file failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, tmpRelativeDir), 0o755); err != nil {
		t.Fatalf("mkdir tmp dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, tmpRelativeDir, "payload.txt"), []byte("tmp payload"), 0o644); err != nil {
		t.Fatalf("write tmp file failed: %v", err)
	}
	legacySnapshotPath := filepath.Join(dataDir, snapshotRelativeDir, "legacy.zip")
	if err := os.WriteFile(legacySnapshotPath, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("write legacy snapshot failed: %v", err)
	}

	snapshotPath, fileName, err := Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !strings.HasPrefix(snapshotPath, filepath.Join(dataDir, snapshotRelativeDir)) {
		t.Fatalf("snapshot path should be under %s, got %s", filepath.Join(dataDir, snapshotRelativeDir), snapshotPath)
	}
	if fileName == "" {
		t.Fatal("snapshot filename should not be empty")
	}

	entries, err := os.ReadDir(filepath.Join(dataDir, snapshotRelativeDir))
	if err != nil {
		t.Fatalf("readdir snapshot dir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("snapshot dir should keep only one file, got %d", len(entries))
	}
	if entries[0].Name() != fileName {
		t.Fatalf("kept file should be latest snapshot %s, got %s", fileName, entries[0].Name())
	}

	zr, err := zip.OpenReader(snapshotPath)
	if err != nil {
		t.Fatalf("open snapshot zip failed: %v", err)
	}
	defer func() { _ = zr.Close() }()

	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, snapshotRelativeDir+"/") || f.Name == snapshotRelativeDir {
			t.Fatalf("snapshot zip should exclude %s, got entry: %s", snapshotRelativeDir, f.Name)
		}
		if strings.HasPrefix(f.Name, tmpRelativeDir+"/") || f.Name == tmpRelativeDir {
			t.Fatalf("snapshot zip should exclude %s, got entry: %s", tmpRelativeDir, f.Name)
		}
	}
}

func TestCreate_WithConsistentDBReplacesLiveDBAndExcludesSidecars(t *testing.T) {
	workspace := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dataDir, "files/images"), 0o755); err != nil {
		t.Fatalf("mkdir images failed: %v", err)
	}
	for name, content := range map[string]string{
		"ech0.db":               "live-db",
		"ech0.db-wal":           "wal",
		"ech0.db-shm":           "shm",
		"files/images/demo.txt": "demo",
	} {
		if err := os.WriteFile(filepath.Join(dataDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s failed: %v", name, err)
		}
	}

	snapshotPath, _, err := Create(WithConsistentDB(func(dstPath string) error {
		return os.WriteFile(dstPath, []byte("consistent-db"), 0o644)
	}))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	zr, err := zip.OpenReader(snapshotPath)
	if err != nil {
		t.Fatalf("open snapshot zip failed: %v", err)
	}
	defer func() { _ = zr.Close() }()

	entries := make(map[string]string, len(zr.File))
	for _, f := range zr.File {
		rc, openErr := f.Open()
		if openErr != nil {
			t.Fatalf("open zip entry %s failed: %v", f.Name, openErr)
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			t.Fatalf("read zip entry %s failed: %v", f.Name, readErr)
		}
		entries[f.Name] = string(content)
	}

	if got := entries[dbFileName]; got != "consistent-db" {
		t.Fatalf("zip should pack the consistent copy as %s, got content %q", dbFileName, got)
	}
	for _, sidecar := range []string{dbFileName + "-wal", dbFileName + "-shm"} {
		if _, ok := entries[sidecar]; ok {
			t.Fatalf("zip should exclude live sidecar %s", sidecar)
		}
	}
	if got := entries["files/images/demo.txt"]; got != "demo" {
		t.Fatalf("zip should keep regular files, got %q", got)
	}

	if _, err := os.Stat(filepath.Join(dataDir, dbStagingRelativeDir)); !os.IsNotExist(err) {
		t.Fatalf("staging dir should be cleaned up after Create, stat err: %v", err)
	}
}

func TestCreate_WithConsistentDBCopyErrorFailsExport(t *testing.T) {
	workspace := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir data failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "ech0.db"), []byte("live-db"), 0o644); err != nil {
		t.Fatalf("write ech0.db failed: %v", err)
	}

	_, _, err = Create(WithConsistentDB(func(string) error {
		return errors.New("boom")
	}))
	if err == nil {
		t.Fatal("Create should fail when the consistent copy cannot be written")
	}
	if !strings.Contains(err.Error(), "consistent db copy") {
		t.Fatalf("error should mention the consistent copy step, got: %v", err)
	}
}

func TestIsDBArtifact(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{key: "ech0.db", expected: true},
		{key: "ech0.db-wal", expected: true},
		{key: "ech0.db-shm", expected: true},
		{key: "ech0.db-journal", expected: true},
		{key: "files/images/a.png", expected: false},
		{key: "other.db", expected: false},
	}

	for _, tc := range tests {
		if got := isDBArtifact(tc.key); got != tc.expected {
			t.Fatalf("isDBArtifact(%q)=%v, expected %v", tc.key, got, tc.expected)
		}
	}
}

func TestShouldExcludeFromSnapshot(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{key: "files/snapshots", expected: true},
		{key: "files/snapshots/a.zip", expected: true},
		{key: "files/tmp", expected: true},
		{key: "files/tmp/a.zip", expected: true},
		{key: "files/capsules", expected: true},
		{key: "files/capsules/ech0_capsule_2026-08-02_10-00-00.zip", expected: true},
		{key: "files/images/a.png", expected: false},
		{key: "ech0.db", expected: false},
	}

	for _, tc := range tests {
		got := shouldExcludeFromSnapshot(tc.key)
		if got != tc.expected {
			t.Fatalf("shouldExcludeFromSnapshot(%q)=%v, expected %v", tc.key, got, tc.expected)
		}
	}
}
