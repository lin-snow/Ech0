// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"fmt"
	"sort"
	"strings"
)

const apiFilesMarker = "/api/files/"

func instanceMarker(text, serverURL string) string {
	if text == "" {
		return ""
	}
	if base := strings.TrimRight(serverURL, "/"); base != "" && strings.Contains(text, base) {
		return base
	}
	if strings.Contains(text, apiFilesMarker) {
		return apiFilesMarker
	}
	return ""
}

type payloadHit struct {
	field  string
	marker string
}

func scanPayload(payload map[string]any, serverURL string) []payloadHit {
	var hits []payloadHit
	scanValue("extension.payload", payload, serverURL, &hits)
	return hits
}

func scanValue(field string, v any, serverURL string, hits *[]payloadHit) {
	switch t := v.(type) {
	case string:
		if marker := instanceMarker(t, serverURL); marker != "" {
			*hits = append(*hits, payloadHit{field: field, marker: marker})
		}
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			scanValue(field+"."+k, t[k], serverURL, hits)
		}
	case []any:
		for i := range t {
			scanValue(fmt.Sprintf("%s[%d]", field, i), t[i], serverURL, hits)
		}
	}
}
