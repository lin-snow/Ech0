// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"crypto/subtle"
	"encoding/json"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	"github.com/lin-snow/ech0/internal/securitystate"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"net/http"
	"os"

	"gorm.io/gorm"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/snapshot"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
)

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
		snapshot.Wake
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&message); err != nil || (message.Run == "" && message.Cleanup == "" && message.SnapshotDue == 0 && (message.Due <= 0 || message.Due%3600 != 0 || message.Due > time.Now().Unix())) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	runtime, err := functionHandler(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	db := database.GetDB()
	if message.Cleanup != "" {
		next, err := snapshot.CleanupExport(r.Context(), db, runtime.Objects.GetStorageConfig(r.Context()), message.Wake)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeWake(w, next)
		return
	}
	if message.Run != "" {
		next, err := snapshot.AdvanceExport(r.Context(), db, runtime.Objects.GetStorageConfig(r.Context()), message.Wake)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeWake(w, next)
		return
	}
	schedule, err := coreSetting.Snapshot.Get(r.Context(), runtime.Settings)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	planned, err := snapshot.NextOccurrence(schedule, time.Now())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if message.SnapshotDue != 0 {
		if err := snapshot.RunOccurrence(r.Context(), db, schedule, message.Wake, time.Now()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	} else {
		if err := database.ExpireSnapshotHistory(r.Context(), db, time.Unix(message.Due, 0)); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if err := db.WithContext(r.Context()).Where("expires <= ?", time.Unix(message.Due, 0).UnixMilli()).Delete(&securitystate.Entry{}).Error; err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if err := db.WithContext(r.Context()).Where("expires_at <= ?", time.Unix(message.Due, 0).UnixMilli()).Delete(&authModel.DeviceGrant{}).Error; err != nil {
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
	}
	next, err := snapshot.PendingExport(r.Context(), db)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	gc, err := snapshot.PendingCleanup(r.Context(), db)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	writeWake(w, next, planned, gc)
}

func writeWake(w http.ResponseWriter, messages ...*snapshot.Wake) {
	pending := make([]*snapshot.Wake, 0, len(messages))
	for _, m := range messages {
		if m != nil {
			pending = append(pending, m)
		}
	}
	if len(pending) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if len(pending) == 1 {
		_ = json.NewEncoder(w).Encode(pending[0])
		return
	}
	_ = json.NewEncoder(w).Encode(pending)
}
