// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"fmt"
	"sort"
	"strings"
)

// Level 是校验发现的严重级别。零值即 LevelError——漏填 Level 的 Issue
// 宁可被当成拦截项，也不要静默放过。
type Level int

const (
	// LevelError 表示胶囊不可被 import / build 消费（spec §7 错误行）。
	LevelError Level = iota
	// LevelWarning 表示可疑但不阻断（spec §7 警告行）。
	LevelWarning
)

// String 返回 CLI 展示用的级别名。
func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarning:
		return "warning"
	default:
		return fmt.Sprintf("level(%d)", int(l))
	}
}

// Issue 是一条校验发现。Path 是胶囊内相对路径（`/` 分隔），Field 是该文件内
// 的字段定位（如 `files[0].key`）；两者共同构成可直接跳转的坐标。
type Issue struct {
	Level   Level
	Path    string
	Field   string
	Message string
}

// Report 是一次校验的完整结果。Fixed 记录 --fix 实际改动了什么，
// 与 Issues 分开——修好的项不再作为 Issue 上报，但用户必须知道文件被改过。
type Report struct {
	Issues []Issue
	Fixed  []string
}

// HasErrors 报告是否存在拦截级发现；import / build 以此为准入门槛。
func (r *Report) HasErrors() bool {
	return r.Count(LevelError) > 0
}

// Count 统计指定级别的发现条数。
func (r *Report) Count(l Level) int {
	n := 0
	for i := range r.Issues {
		if r.Issues[i].Level == l {
			n++
		}
	}
	return n
}

// ErrorSummary 把错误级发现压成一行，供只能显示单个字符串的消费者交代拒绝理由
// （如 Web 导入作业的 error_message，用户看不到 CLI 那张报告表）。
// Issues 已由 sortIssues 排成错误在前，故直接顺序取即可。
func (r *Report) ErrorSummary() string {
	const maxListed = 3

	parts := make([]string, 0, maxListed)
	total := 0
	for i := range r.Issues {
		if r.Issues[i].Level != LevelError {
			continue
		}
		total++
		// 校验错误常成片出现（一个字段错，同类文件全中），全列会淹没 UI。
		if len(parts) < maxListed {
			parts = append(parts, r.Issues[i].String())
		}
	}
	if total == 0 {
		return ""
	}

	summary := strings.Join(parts, "; ")
	if total > len(parts) {
		summary = fmt.Sprintf("%s (+%d more)", summary, total-len(parts))
	}
	return fmt.Sprintf("%d error(s): %s", total, summary)
}

// String 返回 "path[field]: message" 形式的坐标化描述。
func (i Issue) String() string {
	switch {
	case i.Path != "" && i.Field != "":
		return fmt.Sprintf("%s [%s]: %s", i.Path, i.Field, i.Message)
	case i.Path != "":
		return fmt.Sprintf("%s: %s", i.Path, i.Message)
	default:
		return i.Message
	}
}

func (r *Report) errorf(path, field, format string, args ...any) {
	r.Issues = append(r.Issues, Issue{
		Level:   LevelError,
		Path:    path,
		Field:   field,
		Message: fmt.Sprintf(format, args...),
	})
}

func (r *Report) warnf(path, field, format string, args ...any) {
	r.Issues = append(r.Issues, Issue{
		Level:   LevelWarning,
		Path:    path,
		Field:   field,
		Message: fmt.Sprintf(format, args...),
	})
}

// sortIssues 让输出稳定：错误先于警告，其余按坐标排。校验过程本身是
// 遍历顺序驱动的，不排序会让报告随文件系统枚举顺序抖动。
func sortIssues(issues []Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		if a.Level != b.Level {
			return a.Level < b.Level
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Field < b.Field
	})
}
