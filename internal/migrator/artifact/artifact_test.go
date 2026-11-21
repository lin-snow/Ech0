// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyArtifactsMovesSlots(t *testing.T) {
	t.Chdir(t.TempDir())

	legacySnapshot := filepath.Join(DataDir, legacySnapshotDir, "ech0_snapshot_2026-01-01_00-00-00.zip")
	legacyCapsule := filepath.Join(DataDir, legacyCapsuleDir, "ech0_capsule_2026-01-01_00-00-00.zip")
	legacyTmpFile := filepath.Join(DataDir, legacyTmpDir, "ech0_v4", "nested", "ech0.db")
	media := filepath.Join(DataDir, "files", "images", "keep.png")

	for _, path := range []string{legacySnapshot, legacyCapsule, legacyTmpFile, media} {
		mkdirAll(t, filepath.Dir(path))
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	// Pre-existing artifact must survive the merge.
	existing := filepath.Join(DataDir, SnapshotDir, "ech0_snapshot_2025-12-31_00-00-00.zip")
	mkdirAll(t, filepath.Dir(existing))
	if err := os.WriteFile(existing, []byte("y"), 0o644); err != nil {
		t.Fatalf("write %s: %v", existing, err)
	}

	if err := MigrateLegacyArtifacts(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	moved := []string{
		filepath.Join(DataDir, SnapshotDir, "ech0_snapshot_2026-01-01_00-00-00.zip"),
		filepath.Join(DataDir, CapsuleDir, "ech0_capsule_2026-01-01_00-00-00.zip"),
		filepath.Join(DataDir, TmpDir, "ech0_v4", "nested", "ech0.db"),
	}
	for _, path := range moved {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s after migration: %v", path, err)
		}
	}
	for _, path := range []string{legacySnapshot, legacyCapsule, legacyTmpFile} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected legacy %s removed, stat err=%v", path, err)
		}
	}
	for _, dir := range []string{
		filepath.Join(DataDir, legacySnapshotDir),
		filepath.Join(DataDir, legacyCapsuleDir),
		filepath.Join(DataDir, legacyTmpDir),
	} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatalf("expected legacy dir %s removed, stat err=%v", dir, err)
		}
	}
	if _, err := os.Stat(existing); err != nil {
		t.Fatalf("existing artifact must not be touched: %v", err)
	}
	if _, err := os.Stat(media); err != nil {
		t.Fatalf("media outside artifact slots must stay: %v", err)
	}

	// Idempotent: a second run has nothing to move.
	if err := MigrateLegacyArtifacts(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

func TestExcludedCoversArtifactAndLegacySlots(t *testing.T) {
	excluded := map[string]bool{}
	for _, key := range Excluded() {
		excluded[key] = true
	}
	for _, key := range []string{ArtifactDir, legacySnapshotDir, legacyCapsuleDir, legacyTmpDir} {
		if !excluded[key] {
			t.Fatalf("expected %q to be excluded from snapshots", key)
		}
	}
	if excluded["files/images"] {
		t.Fatalf("media must not be excluded from snapshots")
	}
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}
