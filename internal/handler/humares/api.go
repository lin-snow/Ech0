// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

const SecuritySchemeBearer = "bearerAuth"

const schemaRefPrefix = "#/components/schemas/"

func newSchemaNamer() func(reflect.Type, string) string {
	used := map[string]reflect.Type{}
	return func(t reflect.Type, hint string) string {
		name := huma.DefaultSchemaNamer(t, hint)
		key := derefType(t)
		if prev, ok := used[name]; ok && prev != key {
			qualified := packageQualifiedName(t, hint)
			name = qualified
			for i := 2; ; i++ {
				if p, ok := used[name]; !ok || p == key {
					break
				}
				name = qualified + strconv.Itoa(i)
			}
		}
		used[name] = key
		return name
	}
}

func packageQualifiedName(t reflect.Type, hint string) string {
	name := derefType(t).String()
	if name == "" || name == "interface {}" {
		name = hint
	}
	name = strings.ReplaceAll(name, "[]", "List[")
	var b strings.Builder
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		return r == '[' || r == ']' || r == '*' || r == ','
	}) {
		typeName, pkgSeg := part, ""
		if pkgPath, tn, ok := strings.CutLast(part, "."); ok {
			typeName, pkgSeg = tn, pkgPath
			if _, seg, ok := strings.CutLast(pkgPath, "/"); ok {
				pkgSeg = seg
			}
		}
		seg, tn := titleCase(pkgSeg), titleCase(typeName)
		if seg != "" && !strings.EqualFold(seg, tn) {
			b.WriteString(seg)
		}
		b.WriteString(tn)
	}
	return b.String()
}

func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func titleCase(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func NewAPI(engine *gin.Engine, group *gin.RouterGroup, title, version, basePath string, docs DocsRenderer) huma.API {
	installErrorModel()

	cfg := huma.DefaultConfig(title, version)
	cfg.FieldsOptionalByDefault = true
	cfg.AllowAdditionalPropertiesByDefault = true
	cfg.Components.Schemas = huma.NewMapRegistry(schemaRefPrefix, newSchemaNamer())
	cfg.Servers = []*huma.Server{{URL: basePath}}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		SecuritySchemeBearer: {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}

	cfg.CreateHooks = nil
	cfg.Transformers = nil
	cfg.SchemasPath = ""

	useScalar := docs == DocsRendererScalar
	if useScalar {
		cfg.DocsPath = ""
	}

	api := humagin.NewWithGroup(engine, group, cfg)
	api.UseMiddleware(injectLocalizer)
	if useScalar {
		registerScalarDocs(group, basePath)
	}
	return api
}

func Secured(scopes ...string) []map[string][]string {
	if scopes == nil {
		scopes = []string{}
	}
	return []map[string][]string{{SecuritySchemeBearer: scopes}}
}
