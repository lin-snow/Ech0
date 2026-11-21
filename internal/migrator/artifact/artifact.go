// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package artifact

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrNone = errors.New("artifact: no artifact available")

const timeLayout = "2006-01-02_15-04-05"

// Artifacts live outside the local storage root (DataRoot, served
// unauthenticated at /api/files) on purpose: a snapshot contains the whole
// database, so it must never be reachable without the admin-checked download
// route.
const (
	DataDir     = "data"
	ArtifactDir = "artifacts"
	SnapshotDir = ArtifactDir + "/snapshots"
	CapsuleDir  = ArtifactDir + "/capsules"
	TmpDir      = ArtifactDir + "/tmp"

	// Legacy slots used by versions <= 5.9.10, when artifacts were written
	// inside the public storage root. MigrateLegacyArtifacts moves them out.
	legacySnapshotDir = "files/snapshots"
	legacyCapsuleDir  = "files/capsules"
	legacyTmpDir      = "files/tmp"
)

func Snapshots() Slot {
	return NewSlot(filepath.Join(DataDir, SnapshotDir), "ech0_snapshot")
}

func Capsules() Slot {
	return NewSlot(filepath.Join(DataDir, CapsuleDir), "ech0_capsule")
}

// TmpRoot is the working directory for unpacked migration sources.
func TmpRoot() string {
	return filepath.Join(DataDir, TmpDir)
}

// Excluded lists data-relative prefixes that must never enter a snapshot: the
// artifact slots plus the legacy slots, in case a stale copy survived.
func Excluded() []string {
	return []string{ArtifactDir, legacySnapshotDir, legacyCapsuleDir, legacyTmpDir}
}

// MigrateLegacyArtifacts moves artifacts written by versions <= 5.9.10 out of
// the public storage root. It is idempotent and safe to call on every startup.
func MigrateLegacyArtifacts() error {
	moves := []struct{ from, to string }{
		{filepath.Join(DataDir, legacySnapshotDir), Snapshots().Dir()},
		{filepath.Join(DataDir, legacyCapsuleDir), Capsules().Dir()},
		{filepath.Join(DataDir, legacyTmpDir), TmpRoot()},
	}
	for _, move := range moves {
		if err := moveEntries(move.from, move.to); err != nil {
			return err
		}
	}
	return nil
}

func moveEntries(from, to string) error {
	entries, err := os.ReadDir(from)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read legacy artifact dir %s: %w", from, err)
	}
	for _, entry := range entries {
		if err := moveEntry(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
			return err
		}
	}
	if err := os.Remove(from); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy artifact dir %s: %w", from, err)
	}
	return nil
}

func moveEntry(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create artifact dir for %s: %w", dst, err)
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat legacy artifact %s: %w", src, err)
	}
	if !info.IsDir() {
		if err := copyFile(src, dst); err != nil {
			return err
		}
		return os.Remove(src)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read legacy artifact dir %s: %w", src, err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create artifact dir %s: %w", dst, err)
	}
	for _, entry := range entries {
		if err := moveEntry(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
			return err
		}
	}
	return os.Remove(src)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open legacy artifact %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create artifact %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("copy legacy artifact %s: %w", src, err)
	}
	return out.Close()
}

type Slot struct {
	dir    string
	prefix string
}

func NewSlot(dir, prefix string) Slot {
	return Slot{dir: dir, prefix: prefix}
}

func (s Slot) Dir() string {
	return s.dir
}

func (s Slot) Name(at time.Time) string {
	return fmt.Sprintf("%s_%s.zip", s.prefix, at.Format(timeLayout))
}

func (s Slot) Path(name string) string {
	return filepath.Join(s.dir, name)
}

func (s Slot) Latest() (string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNone
		}
		return "", fmt.Errorf("read artifact dir: %w", err)
	}

	latest := ""
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if name := entry.Name(); strings.HasSuffix(name, ".zip") && name > latest {
			latest = name
		}
	}
	if latest == "" {
		return "", ErrNone
	}
	return filepath.Join(s.dir, latest), nil
}

func (s Slot) KeepOnly(keep string) error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("read artifact dir: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == keep {
			continue
		}
		if err := os.RemoveAll(filepath.Join(s.dir, name)); err != nil {
			return fmt.Errorf("cleanup stale artifact %s: %w", name, err)
		}
	}
	return nil
}
