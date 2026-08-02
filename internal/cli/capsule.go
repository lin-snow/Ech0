// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
	capsuleBuild "github.com/lin-snow/ech0/internal/capsule/build"
	capsuleCheck "github.com/lin-snow/ech0/internal/capsule/check"
	capsuleExport "github.com/lin-snow/ech0/internal/capsule/export"
	capsuleImporter "github.com/lin-snow/ech0/internal/capsule/importer"
	tuiUtil "github.com/lin-snow/ech0/internal/util/tui"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

// 默认路径与 spec §9 的命令语法一致。
const (
	DefaultCapsuleDir = "./capsule"
	DefaultDistDir    = "./dist"
	DefaultBaseURL    = "/"
)

// ExportCapsuleOptions 对应 `ech0 export capsule` 的 flag 集合。
type ExportCapsuleOptions struct {
	Output         string
	IncludePrivate bool
	Zip            bool
}

// DoExportCapsule 把当前实例导出为一个胶囊。
func DoExportCapsule(opts ExportCapsuleOptions) error {
	rt, err := newCapsuleRuntime()
	if err != nil {
		return err
	}

	ctx := context.Background()
	result, err := capsuleExport.Run(ctx, capsuleExport.Deps{
		DB:       rt.db,
		Selector: rt.selector(),
		KV:       rt.kv,
	}, capsuleExport.Options{
		Output:         opts.Output,
		IncludePrivate: opts.IncludePrivate,
		Zip:            opts.Zip,
		Generator:      "ech0 v" + versionPkg.Version,
	})
	if err != nil {
		return err
	}

	items := []tuiUtil.CLIInfoItem{
		{Title: "📦 Capsule", Msg: result.Path},
		{Title: "Echoes", Msg: strconv.Itoa(result.Echoes)},
		{Title: "Files", Msg: fmt.Sprintf("%d (external: %d)", result.Files, result.ExternalFiles)},
		{Title: "Comments", Msg: strconv.Itoa(result.Comments)},
		{Title: "Connects", Msg: strconv.Itoa(result.Connects)},
	}
	if result.SkippedPrivate > 0 {
		items = append(items, tuiUtil.CLIInfoItem{
			Title: "Skipped (private)",
			Msg:   strconv.Itoa(result.SkippedPrivate) + "  — use --include-private to carry them",
		})
	}
	tuiUtil.PrintCLIWithBox(items...)
	return nil
}

// DoCheck 校验一个胶囊并按 spec §7 分级报告。存在错误级问题时返回 error（退出码 1）。
func DoCheck(path string, fix bool) error {
	src, err := capsule.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	_, report, err := capsuleCheck.Run(context.Background(), src, capsuleCheck.Options{Fix: fix})
	if err != nil {
		return err
	}
	printCheckReport(path, report)
	if report.HasErrors() {
		return fmt.Errorf("capsule check failed: %d error(s)", report.Count(capsuleCheck.LevelError))
	}
	return nil
}

// ImportCapsuleOptions 对应 `ech0 import capsule` 的 flag 集合。
type ImportCapsuleOptions struct {
	IncludePrivate bool
	DryRun         bool
}

// DoImportCapsule 把一个胶囊导入当前实例。校验在此前置（spec §7：import 隐式执行同一套校验），
// 有错误级问题就拒绝落库。
func DoImportCapsule(path string, opts ImportCapsuleOptions) error {
	src, err := capsule.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	loaded, report, err := capsuleCheck.Run(context.Background(), src, capsuleCheck.Options{})
	if err != nil {
		return err
	}
	printCheckReport(path, report)
	if report.HasErrors() {
		return fmt.Errorf("refusing to import an invalid capsule: %d error(s)", report.Count(capsuleCheck.LevelError))
	}

	rt, err := newCapsuleRuntime()
	if err != nil {
		return err
	}

	result, err := capsuleImporter.Run(context.Background(), capsuleImporter.Deps{
		DB:       rt.db,
		Tx:       rt.tx,
		Selector: rt.selector(),
		KV:       rt.kv,
	}, loaded, capsuleImporter.Options{
		IncludePrivate: opts.IncludePrivate,
		DryRun:         opts.DryRun,
	})
	if err != nil {
		return err
	}

	title := "📥 Imported"
	if opts.DryRun {
		title = "🔍 Dry run (nothing written)"
	}
	items := []tuiUtil.CLIInfoItem{
		{Title: title, Msg: path},
		{Title: "Echoes", Msg: fmt.Sprintf("created %d, skipped %d", result.EchoesCreated, result.EchoesSkipped)},
		{
			Title: "Files",
			Msg: fmt.Sprintf("created %d, reused %d, renamed %d",
				result.FilesCreated, result.FilesReused, result.FilesRenamed),
		},
		{Title: "Tags", Msg: fmt.Sprintf("created %d", result.TagsCreated)},
		{
			Title: "Comments",
			Msg:   fmt.Sprintf("created %d, skipped %d, orphan %d", result.CommentsCreated, result.CommentsSkipped, result.OrphanComments),
		},
	}
	if result.SkippedPrivate > 0 {
		items = append(items, tuiUtil.CLIInfoItem{Title: "Skipped (private)", Msg: strconv.Itoa(result.SkippedPrivate)})
	}
	if len(result.SiteFieldsFilled) > 0 {
		items = append(items, tuiUtil.CLIInfoItem{Title: "Site fields filled", Msg: strings.Join(result.SiteFieldsFilled, ", ")})
	}
	if len(result.Renames) > 0 {
		items = append(items, tuiUtil.CLIInfoItem{Title: "Renamed keys", Msg: strings.Join(result.Renames, "\n")})
	}
	tuiUtil.PrintCLIWithBox(items...)

	// 导入不发布事件（spec §11.3）：向量索引不会自动跟进，明说一句免得用户以为坏了。
	if !opts.DryRun && result.EchoesCreated > 0 {
		tuiUtil.PrintCLIInfo("ℹ️  Note", "no events were emitted; rebuild the embedding index from the dashboard to cover imported content")
	}
	return nil
}

// DoBuild 从胶囊编译出可静态部署的只读站点。
func DoBuild(path, output, baseURL string) error {
	src, err := capsule.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	loaded, report, err := capsuleCheck.Run(context.Background(), src, capsuleCheck.Options{})
	if err != nil {
		return err
	}
	printCheckReport(path, report)
	if report.HasErrors() {
		return fmt.Errorf("refusing to build from an invalid capsule: %d error(s)", report.Count(capsuleCheck.LevelError))
	}

	result, err := capsuleBuild.Run(context.Background(), loaded, capsuleBuild.Options{
		Output:  output,
		BaseURL: baseURL,
	})
	if err != nil {
		return err
	}

	tuiUtil.PrintCLIWithBox(
		tuiUtil.CLIInfoItem{Title: "🌐 Static site", Msg: result.Path},
		tuiUtil.CLIInfoItem{Title: "Echoes", Msg: strconv.Itoa(result.Echoes)},
		tuiUtil.CLIInfoItem{Title: "Files", Msg: strconv.Itoa(result.Files)},
		tuiUtil.CLIInfoItem{Title: "Comments", Msg: strconv.Itoa(result.Comments)},
	)
	return nil
}

// printCheckReport 把校验结果写到 stderr（stdout 留给产物摘要，便于管道消费）。
func printCheckReport(path string, report *capsuleCheck.Report) {
	for _, fixed := range report.Fixed {
		fmt.Fprintf(os.Stderr, "fixed   %s\n", fixed)
	}
	for _, issue := range report.Issues {
		location := issue.Path
		if issue.Field != "" {
			location += " [" + issue.Field + "]"
		}
		fmt.Fprintf(os.Stderr, "%-7s %s: %s\n", issue.Level.String(), location, issue.Message)
	}

	errs := report.Count(capsuleCheck.LevelError)
	warns := report.Count(capsuleCheck.LevelWarning)
	if errs == 0 && warns == 0 && len(report.Fixed) == 0 {
		tuiUtil.PrintCLIInfo("✅ Capsule OK", path)
		return
	}
	fmt.Fprintf(os.Stderr, "%s: %d error(s), %d warning(s)\n", path, errs, warns)
}
