// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	openapiCatalog "github.com/lin-snow/ech0/internal/openapi/catalog"
)

// The CLI allowlist is a client-side file; a server rename must not silently
// drop a command from the tree. This is the cross-language contract test.
func TestCLIScopeIsCatalogSubset(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "apps", "cli", "src", "catalog", "scope.ts"))
	if err != nil {
		t.Fatalf("read CLI scope: %v", err)
	}

	const marker = "export const CLI_OPERATIONS"
	start := strings.Index(string(source), marker)
	if start < 0 {
		t.Fatalf("CLI_OPERATIONS not found in scope.ts")
	}
	block := string(source)[start:]
	end := strings.Index(block, "];")
	if end < 0 {
		t.Fatalf("CLI_OPERATIONS array is not terminated")
	}
	matches := regexp.MustCompile(`"([a-z0-9_.]+)"`).FindAllStringSubmatch(block[:end], -1)
	if len(matches) == 0 {
		t.Fatalf("CLI_OPERATIONS is empty")
	}

	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	api := setupHumaAPI(gin.New())
	registerOperations(api, buildTestHandlers(), nil)
	doc, _, _, err := openapiCatalog.Build(api.OpenAPI())
	if err != nil {
		t.Fatalf("catalog.Build: %v", err)
	}

	for _, match := range matches {
		id := match[1]
		if _, ok := doc.Ops[id]; !ok {
			t.Errorf("CLI scope id %q is not in the operation catalog", id)
		}
	}
}
