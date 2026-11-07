// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

// Cursor is the keyset position of an Echo in the timeline: ordering is
// (created_at DESC, id DESC), so both values are required to resume without
// duplicates or gaps.
type Cursor struct {
	CreatedAt int64
	ID        string
}

const (
	cursorVersionPrefix = "v1:"
	maxCursorIDLength   = 64
)

var ErrInvalidCursor = errors.New("invalid echo cursor")

// EncodeCursor renders an opaque, URL-safe position. The payload is not a
// capability: authorization and every filter are re-applied on each request,
// so tampering can only move the starting point.
func EncodeCursor(cursor Cursor) string {
	payload := cursorVersionPrefix + strconv.FormatInt(cursor.CreatedAt, 10) + ":" + cursor.ID
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func DecodeCursor(raw string) (Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}

	payload := string(decoded)
	if !strings.HasPrefix(payload, cursorVersionPrefix) {
		return Cursor{}, ErrInvalidCursor
	}

	parts := strings.SplitN(strings.TrimPrefix(payload, cursorVersionPrefix), ":", 2)
	if len(parts) != 2 {
		return Cursor{}, ErrInvalidCursor
	}

	createdAt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || createdAt <= 0 {
		return Cursor{}, ErrInvalidCursor
	}

	id := parts[1]
	if id == "" || len(id) > maxCursorIDLength {
		return Cursor{}, ErrInvalidCursor
	}

	return Cursor{CreatedAt: createdAt, ID: id}, nil
}
