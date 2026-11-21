// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bootstrap

import (
	"github.com/lin-snow/ech0/internal/config"
	artifact "github.com/lin-snow/ech0/internal/migrator/artifact"
	"github.com/lin-snow/ech0/pkg/egress"
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

func initEgress() {
	egress.SetAllowPrivate(config.Config().Egress.AllowPrivate)
}

// initArtifacts moves snapshots, capsules and migration working copies written
// by older versions out of the unauthenticated storage root. Failure is not
// fatal: the public file route refuses those directories anyway, so an
// unmoved artifact stays unreachable.
func initArtifacts() {
	if err := artifact.MigrateLegacyArtifacts(); err != nil {
		logUtil.Warn("Failed to migrate legacy artifacts out of the public storage root", logUtil.Err(err))
	}
}

func Bootstrap() {
	initConfig()
	initLogger()
	initEgress()
	initArtifacts()
}
