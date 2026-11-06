// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorRoundTrip(t *testing.T) {
	want := Cursor{CreatedAt: 1735689600, ID: "2b9f2c1e-6f0a-4a0e-9a4b-8f3f2f1c9d10"}

	raw := EncodeCursor(want)
	got, err := DecodeCursor(raw)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCursorRejectsInvalidPayloads(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "not base64", raw: "not a cursor!"},
		{name: "wrong version", raw: encodeRaw("v2:100:echo-id")},
		{name: "missing id", raw: encodeRaw("v1:100:")},
		{name: "missing separator", raw: encodeRaw("v1:100")},
		{name: "zero timestamp", raw: encodeRaw("v1:0:echo-id")},
		{name: "negative timestamp", raw: encodeRaw("v1:-1:echo-id")},
		{name: "non numeric timestamp", raw: encodeRaw("v1:soon:echo-id")},
		{name: "id too long", raw: encodeRaw("v1:100:" + strings.Repeat("a", maxCursorIDLength+1))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeCursor(tc.raw)
			require.ErrorIs(t, err, ErrInvalidCursor)
		})
	}
}

func encodeRaw(payload string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}
