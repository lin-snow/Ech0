// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bootstrap

import (
	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

func initLogger() {
	cfg := config.Config()

	dev := cfg.Server.Mode == "debug"
	logFormat := cfg.Log.Format
	if dev && (logFormat == "" || logFormat == "json") {
		logFormat = "console"
	}

	logUtil.InitLoggerWithConfig(logUtil.LogConfig{
		Level:   cfg.Log.Level,
		Format:  logFormat,
		Console: cfg.Log.Console,
		Color:   dev,
		File: logUtil.FileConfig{
			Enable:     cfg.Log.FileEnable,
			Filename:   cfg.Log.FilePath,
			MaxSize:    cfg.Log.FileMaxSize,
			MaxBackups: cfg.Log.FileMaxBackups,
			MaxAge:     cfg.Log.FileMaxAge,
			Compress:   cfg.Log.FileCompress,
		},
		Stream: logUtil.StreamConfig{
			BufferSize:      cfg.Log.BufferSize,
			RecentSize:      cfg.Log.RecentSize,
			DropPolicy:      cfg.Log.DropPolicy,
			FlushBatch:      cfg.Log.FlushBatch,
			FlushIntervalMs: cfg.Log.FlushIntervalMs,
		},
	})
}

func initConfig() {
	config.Config()
}

func Bootstrap() {
	initConfig()
	initLogger()
}
