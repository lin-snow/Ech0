// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"

	"github.com/lin-snow/ech0/internal/bootstrap"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/di"
	"gorm.io/gorm"
)

var bootstrapOnce sync.Once
var runtimeMu sync.Mutex
var engine http.Handler

func functionHandler(ctx context.Context) (http.Handler, error) {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	if engine != nil {
		return engine, nil
	}
	cfg := config.Config()
	if cfg.Database.Type != "libsql" || !cfg.Upload.VerifiedDirect || cfg.Log.FileEnable || (os.Getenv("JWT_SECRET") == "" && os.Getenv("ECH0_JWT_SECRET") == "") {
		return nil, errors.New("incomplete function configuration")
	}
	bootstrapOnce.Do(bootstrap.Bootstrap)
	db, err := database.OpenRemoteDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}
	state, err := di.BuildFunction(ctx, func() *gorm.DB { return db })
	if err != nil {
		pool, _ := db.DB()
		_ = pool.Close()
		return nil, err
	}
	database.SetDB(db)
	engine = state.Server.GinEngine
	return engine, nil
}

func Ech0(w http.ResponseWriter, r *http.Request) {
	handler, err := functionHandler(r.Context())
	if err != nil {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Retry-After", "5")
		http.Error(w, "Ech0 initialization unavailable", http.StatusServiceUnavailable)
		return
	}
	handler.ServeHTTP(w, r)
}
