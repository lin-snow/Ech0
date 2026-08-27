// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lin-snow/ech0/internal/router"
)

const outPath = "internal/openapi/openapi.yaml"

func main() {
	data, err := router.GenerateOpenAPIYAML()
	if err != nil {
		fmt.Fprintf(os.Stderr, "openapi-gen: 生成 spec 失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "openapi-gen: 创建目录失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "openapi-gen: 写入 %s 失败: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("openapi-gen: 已写入 %s (%d bytes)\n", outPath, len(data))
}
