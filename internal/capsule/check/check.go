// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package check 实现 `ech0 check` 的校验矩阵（spec §7）。
//
// 本包只读胶囊、不碰数据库：import 与 build 由 CLI 统一前置调用 Run，
// 有 LevelError 即拒绝。因此这里的分级就是「能不能落地」的唯一判据——
// 凡是消费者无法自行补全的缺陷才算 error，前向兼容类（未知字段、未知路径）
// 一律降级为 warning（spec §8：禁止因未知内容拒绝处理）。
package check

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/lin-snow/ech0/internal/capsule"
)

// Options 是校验开关。Fix 目前只覆盖 spec §7 唯一列入的自动修复项
// （缺失 id → 生成 UUIDv7 回写）；扩展修复项必须先进规格。
type Options struct {
	Fix bool
}

// Validate 对已加载的胶囊执行全部校验规则。
//
// 除 --fix 的写回失败外不返回 error：一切内容缺陷都进 Report，好让用户
// 一次看全问题清单，而不是修一个跑一次。
func Validate(ctx context.Context, loaded *capsule.Loaded, opts Options) (*Report, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if loaded == nil {
		return nil, errors.New("capsule check: nil capsule")
	}
	if opts.Fix {
		if err := ensureWritable(loaded); err != nil {
			return nil, err
		}
	}

	r := &Report{}

	site := capsule.Site{}
	if loaded.Manifest != nil {
		site = loaded.Manifest.Site
	}

	validateManifest(r, loaded, site)
	echoIDs, referenced, err := validateEchoes(r, loaded, opts, site.ServerURL)
	if err != nil {
		return nil, err
	}
	validateComments(r, loaded, echoIDs)
	validateMedia(r, loaded, referenced, site)
	validatePaths(r, loaded)

	sortIssues(r.Issues)
	return r, nil
}

// Run 是 CLI 入口：加载 + 校验一步到位，并把 Loaded 交回给调用方，
// 免得 import / build 为了拿同一份解析结果再遍历一次胶囊。
func Run(ctx context.Context, src *capsule.Source, opts Options) (*capsule.Loaded, *Report, error) {
	loaded, err := capsule.Load(ctx, src)
	if err != nil {
		return nil, nil, err
	}
	report, err := Validate(ctx, loaded, opts)
	if err != nil {
		return loaded, nil, err
	}
	return loaded, report, nil
}

// ensureWritable 把「zip 胶囊不可写」挡在任何修改之前：--fix 需要就地重写
// Echo 文件，压缩包形态做不到部分更新，与其改一半不如直接拒绝。
func ensureWritable(loaded *capsule.Loaded) error {
	if loaded.Source == nil || loaded.Source.Path == "" {
		return errors.New("capsule check: --fix requires a capsule directory on disk")
	}
	info, err := os.Stat(loaded.Source.Path)
	if err != nil {
		return fmt.Errorf("capsule check: --fix cannot access %q: %w", loaded.Source.Path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("capsule check: --fix is not supported for archive capsules (%s)", loaded.Source.Path)
	}
	return nil
}
