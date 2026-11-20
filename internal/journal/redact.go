// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package journal

import (
	"encoding/json"
	"regexp"
	"strings"

	logUtil "github.com/lin-snow/ech0/pkg/log"
)

var credentialsInText = regexp.MustCompile(`(?i)(bearer\s+|[?&](?:token|code|state|access_token|refresh_token|authToken|x-amz-signature|x-amz-credential|x-amz-security-token)=)[^&\s"'<>]+`)

func safeValue(key string, value any) any {
	name := strings.NewReplacer("-", "", "_", "", " ", "").Replace(strings.ToLower(key))
	switch name {
	case "authorization", "cookie", "setcookie", "password", "smtppassword", "secret", "clientsecret", "secretkey", "apikey", "accesskey", "accesskeyid", "secretaccesskey", "token", "accesstoken", "refreshtoken", "authtoken", "tokenvalue", "jwtsecret", "cronsecret":
		return "[redacted]"
	}
	switch v := value.(type) {
	case string:
		return credentialsInText.ReplaceAllString(v, "${1}[redacted]")
	case map[string]any:
		for k, field := range v {
			v[k] = safeValue(k, field)
		}
	case []any:
		for n, field := range v {
			v[n] = safeValue("", field)
		}
	}
	return value
}

func encodeEntry(entry logUtil.LogEntry) (string, error) {
	// Work on decoded JSON instead of mutating the upstream ring or publishing
	// a second copy of an unredacted search string.
	var fields map[string]any
	if err := json.Unmarshal([]byte(entry.Raw), &fields); err != nil || fields == nil {
		fields = map[string]any{"time": entry.Time, "level": entry.Level, "msg": entry.Msg, "module": entry.Module, "caller": entry.Caller, "error": entry.Error}
	}
	raw, err := json.Marshal(safeValue("", fields))
	return string(raw), err
}
