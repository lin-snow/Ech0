// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package url

import (
	"net"
	neturl "net/url"
	"strings"
)

func TrimURL(url string) string {
	if url == "" {
		return ""
	}

	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	return url
}

// Host returns the bare lower-case host of a URL, or of a host[:port] string:
// no scheme, no port, no IPv6 brackets.
func Host(raw string) string {
	host := strings.TrimSpace(raw)
	if parsed, err := neturl.Parse(host); err == nil && parsed.Host != "" {
		host = parsed.Host
	}
	if stripped, _, err := net.SplitHostPort(host); err == nil {
		host = stripped
	}

	return strings.ToLower(strings.Trim(host, "[]"))
}

// SafeHost renders Host using only DNS/URN characters, for derived identifiers;
// address literals fall back to `localhost`.
func SafeHost(raw string) string {
	host := Host(raw)
	if strings.Contains(host, ":") {
		return "localhost"
	}

	var builder strings.Builder
	for _, char := range host {
		switch {
		case char >= 'a' && char <= 'z', char >= '0' && char <= '9', char == '.', char == '-':
			builder.WriteRune(char)
		}
	}

	if trimmed := strings.Trim(builder.String(), ".-"); trimmed != "" {
		return trimmed
	}

	return "localhost"
}
