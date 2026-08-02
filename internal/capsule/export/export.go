// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package export 实现 `ech0 export capsule`：把当前实例的库内容转储成一个符合
// docs/dev/capsule/spec.md 的胶囊（目录或 .zip）。
//
// 本包只做「转储」——字段名与字段值一律原样落进胶囊，唯一的表示层差异是时间
// （Unix 秒 → RFC3339 UTC）、形态（行 → frontmatter-markdown）与关系实体的内容化
// （Tags → 名称数组、EchoFile.SortOrder → 数组顺序）。任何面向消费者的转换
// （URL 计算、dataset 烘焙、统计冻结）都属于 build，不在这里（spec §11）。
//
// 读库一律直连 GORM 而不过 service 层：service 会发事件、会按当前登录者裁剪可见性，
// 而导出要的是全量的、无副作用的快照。
package export

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/virefs"
	vizip "github.com/lin-snow/ech0/pkg/virefs/plugin/zip"
	"gorm.io/gorm"
)

// Deps 是导出所需的外部能力：库、字节、设置。三者都不可为 nil——导出没有
// 「降级产出」这一档，缺任何一个都只会得到不自包含的胶囊。
type Deps struct {
	DB       *gorm.DB
	Selector *storage.StorageSelector
	KV       kvstore.Store
}

// Options 对应 CLI flag（spec §9）。
type Options struct {
	Output         string // 输出目录；Zip 时为输出文件（缺 .zip 后缀自动补）
	IncludePrivate bool
	Zip            bool
	Generator      string // 写入 manifest.generator 的生产者标识
}

// Result 是导出报告，供 CLI 打印。Files 为写入胶囊的 files 表记录总数（含外链），
// 其中 ExternalFiles 条只带 URL、字节不随胶囊走（spec §11 保真度表）。
type Result struct {
	Path                              string
	Echoes, Files, Comments, Connects int
	SkippedPrivate                    int
	ExternalFiles                     int
}

func (d Deps) validate() error {
	switch {
	case d.DB == nil:
		return errors.New("capsule export: database is required")
	case d.Selector == nil:
		return errors.New("capsule export: storage selector is required")
	case d.KV == nil:
		return errors.New("capsule export: kv store is required")
	}
	return nil
}

// Run 执行一次导出。
func Run(ctx context.Context, deps Deps, opts Options) (*Result, error) {
	if err := deps.validate(); err != nil {
		return nil, err
	}
	output := strings.TrimSpace(opts.Output)
	if output == "" {
		return nil, errors.New("capsule export: output path is empty")
	}

	data, err := collect(ctx, deps, opts)
	if err != nil {
		return nil, err
	}

	// 两种形态都先落到暂存目录，成功才落位：媒体字节取不回时（S3 抖动、对象丢失）
	// 导出必须失败，而失败不该在输出路径上留下半棵树——否则修好 S3 重跑会撞上
	// 「目录已存在且非空」，逼用户手动 rm。暂存目录放在输出的同级而非系统临时目录，
	// 保证落位是同文件系统内的 rename（原子且不复制字节）。
	if !opts.Zip {
		if err := ensureEmptyDir(output); err != nil {
			return nil, err
		}
	}

	parent := filepath.Dir(output)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("capsule export: create output parent %q: %w", parent, err)
	}
	stageDir, err := os.MkdirTemp(parent, ".ech0-capsule-*")
	if err != nil {
		return nil, fmt.Errorf("capsule export: create staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(stageDir) }()

	stage, err := virefs.NewLocalFS(stageDir, virefs.WithCreateRoot(), virefs.WithAtomicWrite())
	if err != nil {
		return nil, fmt.Errorf("capsule export: open staging dir %q: %w", stageDir, err)
	}

	keys, err := writeCapsule(ctx, deps, stage, data, opts)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Path:           output,
		Echoes:         len(data.echoes),
		Files:          len(data.files),
		Comments:       len(data.comments),
		Connects:       len(data.connects),
		SkippedPrivate: data.skippedPrivate,
		ExternalFiles:  data.externalFiles,
	}

	if opts.Zip {
		path, err := packZip(ctx, stage, keys, output)
		if err != nil {
			return nil, err
		}
		result.Path = path
		return result, nil
	}

	// 落位：ensureEmptyDir 允许 output 是个已存在的空目录，而 rename 到已存在目录
	// 并非所有平台都可靠，故先撤掉那个空壳再搬。
	if err := os.Remove(output); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("capsule export: clear output %q: %w", output, err)
	}
	if err := os.Rename(stageDir, output); err != nil {
		return nil, fmt.Errorf("capsule export: move capsule into place: %w", err)
	}
	return result, nil
}

// ensureEmptyDir 拒绝往已有内容的目录里写：胶囊是一整棵树，覆盖式写入会把上一次
// 导出的残留（已删除的 Echo、改名的媒体）混进新胶囊，且可能悄悄吃掉用户的数据。
func ensureEmptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("capsule export: inspect output %q: %w", dir, err)
	case len(entries) > 0:
		return fmt.Errorf("capsule export: output %q already exists and is not empty", dir)
	}
	return nil
}

// packZip 把暂存树按写入顺序打成 zip。keys 即 zip 内条目名，故 zip 解开后与目录
// 形态逐字节同形。
func packZip(ctx context.Context, stage virefs.FS, keys []string, output string) (string, error) {
	out := output
	if !strings.EqualFold(filepath.Ext(out), ".zip") {
		out += ".zip"
	}
	if dir := filepath.Dir(out); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("capsule export: create output dir %q: %w", dir, err)
		}
	}

	f, err := os.Create(out)
	if err != nil {
		return "", fmt.Errorf("capsule export: create %q: %w", out, err)
	}
	if err := vizip.Pack(ctx, stage, keys, f); err != nil {
		_ = f.Close()
		_ = os.Remove(out) // 半个 zip 比没有 zip 更危险
		return "", fmt.Errorf("capsule export: pack %q: %w", out, err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("capsule export: close %q: %w", out, err)
	}
	return out, nil
}
