// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package securitystate

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Entry represents short-lived replay protection or throttling state.
// Keys are domain-separated HMACs. Expiry is expressed in milliseconds.
type Entry struct {
	ID      string `gorm:"primaryKey"`
	Expires int64  `gorm:"not null;index"`
	Value   int64  `gorm:"not null"`
}

func (Entry) TableName() string { return "security_windows" }

func Digest(secret []byte, fields ...string) []byte {
	mac := hmac.New(sha256.New, secret)
	for _, field := range fields {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(field)))
		mac.Write(size[:])
		mac.Write([]byte(field))
	}
	return mac.Sum(nil)
}

func Key(secret []byte, fields ...string) string {
	return hex.EncodeToString(Digest(secret, fields...))
}

func Acquire(db *gorm.DB, id string, until, now time.Time) (bool, error) {
	result := db.Exec("INSERT INTO security_windows(id,expires,value) VALUES(?,?,1) ON CONFLICT(id) DO UPDATE SET expires=excluded.expires WHERE security_windows.expires<=?", id, until.UnixMilli(), now.UnixMilli())
	return result.RowsAffected == 1, result.Error
}

func Consume(db *gorm.DB, id string) (time.Time, bool, error) {
	var expiry int64
	err := db.Raw("DELETE FROM security_windows WHERE id=? RETURNING expires", id).Row().Scan(&expiry)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false, nil
	}
	return time.UnixMilli(expiry), err == nil, err
}

func Increment(db *gorm.DB, id string, until time.Time) (int64, error) {
	var count int64
	err := db.Raw("INSERT INTO security_windows(id,expires,value) VALUES(?,?,1) ON CONFLICT(id) DO UPDATE SET value=value+1 RETURNING value", id, until.UnixMilli()).Row().Scan(&count)
	return count, err
}

// Allow advances the theoretical arrival time by one token only if it stays
// within the burst allowance. Concurrent callers use one conditional write.
func Allow(ctx context.Context, db *gorm.DB, id string, now time.Time, rps, burst int) (bool, error) {
	step := int64(time.Second) / int64(rps)
	debt := step * int64(burst)
	at := now.UnixNano()
	result := db.WithContext(ctx).Exec(`INSERT INTO security_windows(id,expires,value) VALUES(?,?,?)
        ON CONFLICT(id) DO UPDATE SET expires=excluded.expires,value=max(security_windows.value,?)+?
        WHERE security_windows.value<=?`, id, now.Add(time.Duration(debt)+time.Second).UnixMilli(), at+step, at, step, at+debt-step)
	return result.RowsAffected == 1, result.Error
}
