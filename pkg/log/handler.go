// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lin-snow/ech0/pkg/log/tint"
)

const (
	LevelPanic = slog.LevelError + 4
	LevelFatal = slog.LevelError + 8
)

type fanoutHandler struct {
	leaves []slog.Handler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, leaf := range h.leaves {
		if leaf.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, leaf := range h.leaves {
		if leaf.Enabled(ctx, r.Level) {
			_ = leaf.Handle(ctx, r.Clone())
		}
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	leaves := make([]slog.Handler, len(h.leaves))
	for i, leaf := range h.leaves {
		leaves[i] = leaf.WithAttrs(attrs)
	}
	return &fanoutHandler{leaves: leaves}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	leaves := make([]slog.Handler, len(h.leaves))
	for i, leaf := range h.leaves {
		leaves[i] = leaf.WithGroup(name)
	}
	return &fanoutHandler{leaves: leaves}
}

func newConsoleLeaf(config LogConfig, level slog.Leveler) slog.Handler {
	if config.Format == "json" {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: fileReplace,
		})
	}
	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		NoColor:    !config.Color,
		TimeFormat: "15:04:05",
	})
}

type ringHandler struct {
	hub    *LogStreamHub
	level  slog.Leveler
	attrs  []slog.Attr
	groups []string
}

func (h *ringHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *ringHandler) Handle(_ context.Context, r slog.Record) error {
	if h.hub != nil {
		h.hub.Publish(recordToEntry(r, h.groups, h.attrs))
	}
	return nil
}

func (h *ringHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &nh
}

func (h *ringHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	nh := *h
	nh.groups = append(append([]string{}, h.groups...), name)
	return &nh
}

func recordToEntry(r slog.Record, groups []string, base []slog.Attr) LogEntry {
	e := LogEntry{
		Time:  r.Time.Format(time.RFC3339),
		Level: levelString(r.Level),
		Msg:   r.Message,
	}
	fields := make(map[string]any)
	prefix := ""
	if len(groups) > 0 {
		prefix = strings.Join(groups, ".") + "."
	}
	put := func(a slog.Attr) {
		if a.Equal(slog.Attr{}) {
			return
		}
		if prefix == "" {
			switch a.Key {
			case "module":
				e.Module = a.Value.String()
				return
			case "error", "err":
				if e.Error == "" {
					e.Error = a.Value.String()
				}
				return
			}
		}
		fields[prefix+a.Key] = a.Value.Any()
	}
	for _, a := range base {
		put(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		put(a)
		return true
	})
	if r.PC != 0 {
		e.Caller = shortCaller(r.PC)
	}
	if len(fields) > 0 {
		e.Fields = fields
	}
	e.Raw = compactEntryJSON(e)
	return e
}

func compactEntryJSON(e LogEntry) string {
	m := make(map[string]any, len(e.Fields)+6)
	m["time"] = e.Time
	m["level"] = e.Level
	m["msg"] = e.Msg
	if e.Module != "" {
		m["module"] = e.Module
	}
	if e.Caller != "" {
		m["caller"] = e.Caller
	}
	if e.Error != "" {
		m["error"] = e.Error
	}
	maps.Copy(m, e.Fields)
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

func levelString(l slog.Level) string {
	switch {
	case l >= LevelFatal:
		return "fatal"
	case l >= LevelPanic:
		return "panic"
	case l >= slog.LevelError:
		return "error"
	case l >= slog.LevelWarn:
		return "warn"
	case l >= slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}

func shortCaller(pc uintptr) string {
	f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	if f.File == "" {
		return ""
	}
	return trimSourcePath(f.File) + ":" + strconv.Itoa(f.Line)
}

func fileReplace(groups []string, a slog.Attr) slog.Attr {
	if len(groups) != 0 {
		return a
	}
	switch a.Key {
	case slog.LevelKey:
		if lvl, ok := a.Value.Any().(slog.Level); ok {
			return slog.String(slog.LevelKey, levelString(lvl))
		}
	case slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
			return slog.String("caller", trimSourcePath(src.File)+":"+strconv.Itoa(src.Line))
		}
	}
	return a
}

func trimSourcePath(file string) string {
	if file == "" {
		return ""
	}
	if slash := strings.LastIndexByte(file, '/'); slash >= 0 {
		if prev := strings.LastIndexByte(file[:slash], '/'); prev >= 0 {
			return file[prev+1:]
		}
		return file[slash+1:]
	}
	return file
}

func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String("error", err.Error())
}
