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
	"github.com/lin-snow/ech0/internal/serverless"
	"github.com/lin-snow/ech0/internal/serverless/platform"
	"gorm.io/gorm"
)

var bootstrapOnce sync.Once
var runtimeMu sync.Mutex
var engine *serverless.Runtime

func functionHandler(ctx context.Context) (*serverless.Runtime, error) {
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
	engine = state
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
	handler.Server.GinEngine.ServeHTTP(w, r.WithContext(platform.WithRequest(r.Context(), r)))
}
