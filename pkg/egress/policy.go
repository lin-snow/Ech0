// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"log/slog"
	"sync/atomic"

	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// allowPrivate disables the loopback/private/reserved-address blocks. It is a
// process-wide switch because Validate/Fetch are package-level and called from
// every domain. Set once at startup from ECH0_EGRESS_ALLOW_PRIVATE.
var allowPrivate atomic.Bool

// SetAllowPrivate toggles the private-address guard. Only local development
// should enable it: every egress caller (connect, webhook, OAuth CIMD, site
// scraping) then accepts localhost and private/reserved targets, which is the
// SSRF surface the guard exists to close.
func SetAllowPrivate(allow bool) {
	allowPrivate.Store(allow)
	if allow {
		logUtil.Warn(
			"egress private-address guard disabled; localhost/private targets are reachable",
			slog.String("module", "egress"),
			slog.String("env", "ECH0_EGRESS_ALLOW_PRIVATE"),
		)
	}
}
