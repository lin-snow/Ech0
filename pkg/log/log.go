// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const initLoggerPanic = "初始化 Logger 失败"

var (
	Logger        *slog.Logger
	loggerMu      sync.Mutex
	levelVar      = new(slog.LevelVar)
	fileWriter    *lumberjack.Logger
	currentConfig LogConfig
	streamHub     *LogStreamHub
	fileSinkStop  chan struct{}
	fileSinkDone  chan struct{}
	fileSinkID    int64
)

type LogConfig struct {
	Level   string       `yaml:"level"   json:"level"`
	Format  string       `yaml:"format"  json:"format"`
	Console bool         `yaml:"console" json:"console"`
	Color   bool         `yaml:"-"       json:"-"`
	File    FileConfig   `yaml:"file"    json:"file"`
	Stream  StreamConfig `yaml:"stream"  json:"stream"`
}

type FileConfig struct {
	Enable     bool   `yaml:"enable"     json:"enable"`
	Filename   string `yaml:"filename"   json:"filename"`
	MaxSize    int    `yaml:"maxsize"    json:"maxsize"`
	MaxBackups int    `yaml:"maxbackups" json:"maxbackups"`
	MaxAge     int    `yaml:"maxage"     json:"maxage"`
	Compress   bool   `yaml:"compress"   json:"compress"`
}

type StreamConfig struct {
	BufferSize      int    `yaml:"buffer_size" json:"buffer_size"`
	RecentSize      int    `yaml:"recent_size" json:"recent_size"`
	DropPolicy      string `yaml:"drop_policy" json:"drop_policy"`
	FlushBatch      int    `yaml:"flush_batch" json:"flush_batch"`
	FlushIntervalMs int    `yaml:"flush_interval_ms" json:"flush_interval_ms"`
}

type LogEntry struct {
	Time   string         `json:"time"`
	Level  string         `json:"level"`
	Msg    string         `json:"msg"`
	Module string         `json:"module,omitempty"`
	Caller string         `json:"caller,omitempty"`
	Error  string         `json:"error,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
	Raw    string         `json:"raw,omitempty"`
}

func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:   "info",
		Format:  "json",
		Console: false,
		File: FileConfig{
			Enable:     true,
			Filename:   "data/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
		Stream: StreamConfig{
			BufferSize:      2048,
			RecentSize:      2000,
			DropPolicy:      "drop_oldest",
			FlushBatch:      128,
			FlushIntervalMs: 500,
		},
	}
}

func InitLogger() {
	InitLoggerWithConfig(DefaultLogConfig())
}

func InitLoggerWithConfig(config LogConfig) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	initializeLogger(config)
}

func initializeLogger(config LogConfig) {
	currentConfig = config

	Logger = nil
	stopFileSink()
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}
	if streamHub != nil {
		streamHub.Close()
		streamHub = nil
	}

	levelVar.Set(parseLevel(config.Level))

	streamHub = newLogStreamHub(
		safePositive(config.Stream.BufferSize, 2048),
		safePositive(config.Stream.RecentSize, 2000),
		normalizeDropPolicy(config.Stream.DropPolicy),
	)

	if config.File.Enable {
		logDir := filepath.Dir(config.File.Filename)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			panic(initLoggerPanic + ": 创建日志目录失败: " + err.Error())
		}
		fileWriter = &lumberjack.Logger{
			Filename:   config.File.Filename,
			MaxSize:    config.File.MaxSize,
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAge,
			Compress:   config.File.Compress,
			LocalTime:  true,
		}
		startFileSink(config)
	}

	leaves := []slog.Handler{
		newConsoleLeaf(config, levelVar),
		&ringHandler{hub: streamHub, level: levelVar},
	}

	Logger = slog.New(&fanoutHandler{leaves: leaves})
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "panic":
		return LevelPanic
	case "fatal":
		return LevelFatal
	default:
		return slog.LevelInfo
	}
}

func GetLogger() *slog.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if Logger == nil {
		cfg := currentConfig
		if cfg == (LogConfig{}) {
			cfg = DefaultLogConfig()
		}
		initializeLogger(cfg)
	}
	return Logger
}

func CloseLogger() {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	Logger = nil
	stopFileSink()
	if fileWriter != nil {
		_ = fileWriter.Close()
		fileWriter = nil
	}
	if streamHub != nil {
		streamHub.Close()
		streamHub = nil
	}
}

func ReopenLogger() {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if Logger != nil {
		return
	}

	cfg := currentConfig
	if cfg == (LogConfig{}) {
		cfg = DefaultLogConfig()
	}
	initializeLogger(cfg)
}

//go:noinline
func Debug(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelDebug, msg, attrs)
}

//go:noinline
func Info(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelInfo, msg, attrs)
}

//go:noinline
func Warn(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelWarn, msg, attrs)
}

//go:noinline
func Error(msg string, attrs ...slog.Attr) {
	logWithAttrs(slog.LevelError, msg, attrs)
}

//go:noinline
func Panic(msg string, attrs ...slog.Attr) {
	logWithStack(LevelPanic, msg, attrs)
	panic(msg)
}

//go:noinline
func Fatal(msg string, attrs ...slog.Attr) {
	logWithStack(LevelFatal, msg, attrs)
	loggerMu.Lock()
	stopFileSink()
	loggerMu.Unlock()
	os.Exit(1)
}

//go:noinline
func logWithAttrs(level slog.Level, msg string, attrs []slog.Attr) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = os.Stderr.WriteString("Recovered panic in logger\n")
		}
	}()

	logger := GetLogger()
	ctx := context.Background()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = logger.Handler().Handle(ctx, r)
}

//go:noinline
func logWithStack(level slog.Level, msg string, attrs []slog.Attr) {
	logger := GetLogger()
	ctx := context.Background()
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	r.AddAttrs(slog.String("stacktrace", string(debug.Stack())))
	_ = logger.Handler().Handle(ctx, r)
}

func SubscribeLogs(bufferSize int) (int64, <-chan LogEntry, func()) {
	loggerMu.Lock()
	if streamHub == nil {
		cfg := currentConfig
		if cfg == (LogConfig{}) {
			cfg = DefaultLogConfig()
		}
		initializeLogger(cfg)
	}
	hub := streamHub
	loggerMu.Unlock()
	return hub.Subscribe(bufferSize)
}

func RecentLogs(limit int) []LogEntry {
	loggerMu.Lock()
	hub := streamHub
	loggerMu.Unlock()
	if hub == nil {
		return nil
	}
	return hub.Recent(limit)
}

func CurrentLogFilePath() string {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if currentConfig.File.Filename == "" {
		return DefaultLogConfig().File.Filename
	}
	return currentConfig.File.Filename
}

func QueryLogFileTail(path string, limit int, level, keyword string) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 5000 {
		limit = 5000
	}
	level = strings.ToLower(strings.TrimSpace(level))
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []LogEntry{}, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	buffer := make([]LogEntry, 0, limit)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entry := parseLogLine(line)
		if !matchLogFilters(entry, level, keyword) {
			continue
		}
		if len(buffer) < limit {
			buffer = append(buffer, entry)
			continue
		}
		copy(buffer, buffer[1:])
		buffer[len(buffer)-1] = entry
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return buffer, nil
}

type LogStreamHub struct {
	mu               sync.RWMutex
	subs             map[int64]chan LogEntry
	nextID           int64
	recent           []LogEntry
	recentCap        int
	recentPos        int
	recentLen        int
	dropPolicy       string
	dropped          atomic.Uint64
	closed           bool
	defaultSubBuffer int
}

func newLogStreamHub(subBufferSize, recentSize int, dropPolicy string) *LogStreamHub {
	if subBufferSize <= 0 {
		subBufferSize = 2048
	}
	if recentSize <= 0 {
		recentSize = 2000
	}
	return &LogStreamHub{
		subs:             make(map[int64]chan LogEntry),
		recent:           make([]LogEntry, recentSize),
		recentCap:        recentSize,
		dropPolicy:       dropPolicy,
		defaultSubBuffer: subBufferSize,
	}
}

func (h *LogStreamHub) Subscribe(bufferSize int) (int64, <-chan LogEntry, func()) {
	if bufferSize <= 0 {
		bufferSize = h.defaultSubBuffer
		if bufferSize <= 0 {
			bufferSize = 256
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		ch := make(chan LogEntry)
		close(ch)
		return 0, ch, func() {}
	}
	h.nextID++
	id := h.nextID
	ch := make(chan LogEntry, bufferSize)
	h.subs[id] = ch
	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		c, ok := h.subs[id]
		if !ok {
			return
		}
		delete(h.subs, id)
		close(c)
	}
	return id, ch, cancel
}

func (h *LogStreamHub) Publish(entry LogEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.recent[h.recentPos] = entry
	h.recentPos = (h.recentPos + 1) % h.recentCap
	if h.recentLen < h.recentCap {
		h.recentLen++
	}
	for _, ch := range h.subs {
		select {
		case ch <- entry:
		default:
			if h.dropPolicy == "drop_oldest" {
				select {
				case <-ch:
				default:
				}
				select {
				case ch <- entry:
				default:
					h.dropped.Add(1)
				}
				continue
			}
			h.dropped.Add(1)
		}
	}
}

func (h *LogStreamHub) Recent(limit int) []LogEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.recentLen == 0 {
		return nil
	}
	if limit <= 0 || limit > h.recentLen {
		limit = h.recentLen
	}
	out := make([]LogEntry, 0, limit)
	start := h.recentPos - limit
	if start < 0 {
		start += h.recentCap
	}
	for i := 0; i < limit; i++ {
		idx := (start + i) % h.recentCap
		out = append(out, h.recent[idx])
	}
	return out
}

func (h *LogStreamHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, ch := range h.subs {
		close(ch)
		delete(h.subs, id)
	}
}

func startFileSink(config LogConfig) {
	if streamHub == nil || fileWriter == nil {
		return
	}
	fileSinkStop = make(chan struct{})
	fileSinkDone = make(chan struct{})
	bufferSize := safePositive(config.Stream.BufferSize, 2048)
	var stream <-chan LogEntry
	fileSinkID, stream, _ = streamHub.Subscribe(bufferSize)
	flushBatch := safePositive(config.Stream.FlushBatch, 128)
	flushInterval := time.Duration(safePositive(config.Stream.FlushIntervalMs, 500)) * time.Millisecond
	go func() {
		defer close(fileSinkDone)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()
		lines := make([]string, 0, flushBatch)
		flush := func() {
			if len(lines) == 0 || fileWriter == nil {
				return
			}
			payload := strings.Join(lines, "\n") + "\n"
			_, _ = fileWriter.Write([]byte(payload))
			lines = lines[:0]
		}
		appendLine := func(entry LogEntry) {
			if line := strings.TrimSpace(entry.Raw); line != "" {
				lines = append(lines, line)
			}
		}

		for {
			select {
			case <-fileSinkStop:
				for {
					select {
					case entry, ok := <-stream:
						if ok {
							appendLine(entry)
							continue
						}
						flush()
						return
					default:
						flush()
						return
					}
				}
			case <-ticker.C:
				flush()
			case entry, ok := <-stream:
				if !ok {
					flush()
					return
				}
				appendLine(entry)
				if len(lines) >= flushBatch {
					flush()
				}
			}
		}
	}()
}

func stopFileSink() {
	if fileSinkStop != nil {
		close(fileSinkStop)
		fileSinkStop = nil
	}
	if streamHub != nil && fileSinkID > 0 {
		streamHub.mu.Lock()
		if ch, ok := streamHub.subs[fileSinkID]; ok {
			delete(streamHub.subs, fileSinkID)
			close(ch)
		}
		streamHub.mu.Unlock()
		fileSinkID = 0
	}
	if fileSinkDone != nil {
		<-fileSinkDone
		fileSinkDone = nil
	}
}

func safePositive(v int, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func normalizeDropPolicy(policy string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "drop_newest":
		return "drop_newest"
	default:
		return "drop_oldest"
	}
}

func parseLogLine(line string) LogEntry {
	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return LogEntry{
			Level: "info",
			Msg:   line,
			Raw:   line,
		}
	}
	return parseMapAsEntry(payload, line)
}

func parseMapAsEntry(payload map[string]any, raw string) LogEntry {
	entry := LogEntry{
		Time:   toString(payload["time"]),
		Level:  strings.ToLower(toString(payload["level"])),
		Msg:    toString(payload["msg"]),
		Module: toString(payload["module"]),
		Caller: toString(payload["caller"]),
		Error:  toString(payload["error"]),
		Raw:    raw,
	}
	fields := make(map[string]any)
	for k, v := range payload {
		switch k {
		case "time", "level", "msg", "module", "caller", "error":
			continue
		case "source":
			if entry.Caller == "" {
				entry.Caller = sourceMapToCaller(v)
			}
			continue
		default:
			fields[k] = v
			if k == "err" && entry.Error == "" {
				entry.Error = toString(v)
			}
		}
	}
	if len(fields) > 0 {
		entry.Fields = fields
	}
	if entry.Msg == "" {
		entry.Msg = raw
	}
	return entry
}

func sourceMapToCaller(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	file := toString(m["file"])
	if file == "" {
		return ""
	}
	caller := trimSourcePath(file)
	if line := toString(m["line"]); line != "" {
		caller += ":" + line
	}
	return caller
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch tv := v.(type) {
	case string:
		return tv
	default:
		b, err := json.Marshal(tv)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func matchLogFilters(entry LogEntry, level, keyword string) bool {
	if level != "" && level != "all" && strings.ToLower(entry.Level) != level {
		return false
	}
	if keyword != "" {
		raw := strings.ToLower(entry.Raw)
		msg := strings.ToLower(entry.Msg)
		if !strings.Contains(raw, keyword) && !strings.Contains(msg, keyword) {
			return false
		}
	}
	return true
}
