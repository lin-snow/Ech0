// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"path"
	"strings"
)

type Route struct {
	prefix string
	match  func(key string) bool
}

func RouteByExt(prefix string, exts ...string) Route {
	lower := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		lower[strings.ToLower(ext)] = struct{}{}
	}
	return Route{
		prefix: prefix,
		match: func(key string) bool {
			ext := strings.ToLower(path.Ext(key))
			_, ok := lower[ext]
			return ok
		},
	}
}

func RouteByFunc(prefix string, fn func(key string) bool) Route {
	return Route{prefix: prefix, match: fn}
}

func DefaultRoute(prefix string) Route {
	return Route{prefix: prefix, match: func(string) bool { return true }}
}

type Schema struct {
	routes []Route
}

func NewSchema(routes ...Route) *Schema {
	return &Schema{routes: routes}
}

func (s *Schema) Resolve(key string) string {
	for _, r := range s.routes {
		if r.match(key) {
			return r.prefix + key
		}
	}
	return key
}
