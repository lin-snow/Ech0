// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package journal

import (
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRedactionPrecedesStorageAndSearch(t *testing.T) {
	raw, err := encodeEntry(logUtil.LogEntry{Raw: `{"msg":"GET /stream?token=dummy-private","level":"info","headers":{"Authorization":"Bearer dummy-private"},"nested":[{"api_key":"dummy-private"}],"request_id":"trace"}`})
	require.NoError(t, err)
	require.NotContains(t, raw, "dummy-private")
	require.Contains(t, raw, "trace")
}
