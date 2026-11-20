// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	"gorm.io/gorm"
)

var poolMu sync.Mutex
var pool *gorm.DB

func maintenanceDB() (*gorm.DB, error) {
	poolMu.Lock()
	defer poolMu.Unlock()
	if pool != nil {
		return pool, nil
	}
	db, err := database.OpenRemoteDatabase(config.Config().Database)
	if err == nil {
		pool = db
	}
	return db, err
}

// Handler executes the hourly visitor-history retention task. The cutoff is
// pinned to the scheduled occurrence, so redelivery repeats the same DELETE.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	secret := os.Getenv("CRON_SECRET")
	if secret == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+secret)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var message struct {
		Due int64 `json:"due"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&message); err != nil || message.Due <= 0 || message.Due%3600 != 0 || message.Due > time.Now().Unix() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	db, err := maintenanceDB()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	repo := visitorRepository.NewVisitorRepository(func() *gorm.DB { return db.WithContext(r.Context()) })
	cutoff := time.Unix(message.Due, 0).UTC().AddDate(0, 0, -6).Format(time.DateOnly)
	if err := repo.DeleteSeenBefore(r.Context(), time.Unix(message.Due, 0).UTC().AddDate(0, 0, -1).Format(time.DateOnly)); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if err := repo.DeleteOlderThan(r.Context(), cutoff); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
