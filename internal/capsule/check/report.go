// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"fmt"
	"sort"
	"strings"
)

type Level int

const (
	LevelError Level = iota
	LevelWarning
)

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

type Issue struct {
	Level   Level
	Path    string
	Field   string
	Message string
}

type Report struct {
	Issues []Issue
	Fixed  []string
}

func (r *Report) HasErrors() bool {
	return r.Count(LevelError) > 0
}

func (r *Report) Count(l Level) int {
	n := 0
	for i := range r.Issues {
		if r.Issues[i].Level == l {
			n++
		}
	}
	return n
}

func (r *Report) ErrorSummary() string {
	const maxListed = 3

	parts := make([]string, 0, maxListed)
	total := 0
	for i := range r.Issues {
		if r.Issues[i].Level != LevelError {
			continue
		}
		total++
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
