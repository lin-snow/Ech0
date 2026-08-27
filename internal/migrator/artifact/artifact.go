// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package artifact

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrNone = errors.New("artifact: no artifact available")

const timeLayout = "2006-01-02_15-04-05"

const (
	DataDir     = "data"
	SnapshotDir = "files/snapshots"
	CapsuleDir  = "files/capsules"
	TmpDir      = "files/tmp"
)

func Snapshots() Slot {
	return NewSlot(filepath.Join(DataDir, SnapshotDir), "ech0_snapshot")
}

func Capsules() Slot {
	return NewSlot(filepath.Join(DataDir, CapsuleDir), "ech0_capsule")
}

func Excluded() []string {
	return []string{SnapshotDir, CapsuleDir, TmpDir}
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
