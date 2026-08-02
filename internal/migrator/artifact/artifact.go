// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package artifact 管理「只保留最新一份」的导出产物槽位。
//
// 快照与胶囊各占一个槽位，这不是整洁强迫症而是正确性要求：两者都遵循「保留最新一份」，
// 挤在同一目录里就会互删——定时快照走 gocron 直连 ExportEngine、不经过 job.Manager，
// 作业互斥拦不住它，用户导完胶囊还没点下载，定时快照一响就把胶囊清掉了。
package artifact

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrNone 表示槽位里尚无可用产物（需先跑一次导出）。
var ErrNone = errors.New("artifact: no artifact available")

// timeLayout 的字段按从大到小排列，因此文件名的字典序与时间序一致——Latest 取字典序最大者
// 即最新一份，无需 stat 每个文件。
const timeLayout = "2006-01-02_15-04-05"

// 产物目录布局。三者都在 data/ 下，且都**必须**被排除在快照之外——快照是「data/ 的 zip」，
// 把派生产物打进去会让快照套娃式膨胀。新增产物目录时改这里，Excluded 会自动带上，
// 不会漏掉排除这一步。
const (
	DataDir     = "data"
	SnapshotDir = "files/snapshots"
	CapsuleDir  = "files/capsules"
	TmpDir      = "files/tmp"
)

// Snapshots 是快照产物槽位（整个 data/ 的 zip，含账号与凭据）。
func Snapshots() Slot {
	return NewSlot(filepath.Join(DataDir, SnapshotDir), "ech0_snapshot")
}

// Capsules 是胶囊产物槽位（可分享的内容子集）。与快照分居两个目录，否则「只保留最新一份」
// 会让两者互删。
func Capsules() Slot {
	return NewSlot(filepath.Join(DataDir, CapsuleDir), "ech0_capsule")
}

// Excluded 返回不进快照的子树（相对 data/）。
func Excluded() []string {
	return []string{SnapshotDir, CapsuleDir, TmpDir}
}

// Slot 是一个产物目录加文件名前缀。零值不可用，用 NewSlot 构造。
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

// Name 生成该时刻的产物文件名。
func (s Slot) Name(at time.Time) string {
	return fmt.Sprintf("%s_%s.zip", s.prefix, at.Format(timeLayout))
}

func (s Slot) Path(name string) string {
	return filepath.Join(s.dir, name)
}

// Latest 返回槽位里最新一份产物的路径，空槽位返回 ErrNone。
// 只认已完成的 .zip：打包中的临时文件是 "." + name + ".tmp"，不会被选中。
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

// KeepOnly 删除槽位里除 keep 之外的一切，实现「只保留最新一份」的本地留存策略。
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
