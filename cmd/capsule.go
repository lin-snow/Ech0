// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cmd

import (
	"github.com/lin-snow/ech0/internal/cli"
	"github.com/spf13/cobra"
)

// 命令语法见 docs/dev/capsule/spec.md §9：动词在前、格式为子命令。
// 格式不做成 `--type=capsule|snapshot` 是因为两种产物的 flag 集合本就分叉，
// 且 snapshot 导入是破坏性整库替换——这种语义差别不该只隔着一个 flag 值。
var (
	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export instance data (choose a format sub-command)",
		Long:  "Export instance data. `capsule` produces a human-readable, portable content capsule; `snapshot` produces a full instance backup archive.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	importCmd = &cobra.Command{
		Use:   "import",
		Short: "Import data into this instance (choose a format sub-command)",
		Long:  "Import data into this instance. `capsule` merges portable content idempotently; `snapshot` restores a full instance backup.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
)

var (
	exportCapsuleOpts cli.ExportCapsuleOptions
	exportSnapshotOut string

	importCapsuleOpts cli.ImportCapsuleOptions
	importSnapshotYes bool

	checkFix bool

	buildOutput  string
	buildBaseURL string
)

var exportCapsuleCmd = &cobra.Command{
	Use:          "capsule",
	Short:        "Export this instance as a portable content capsule",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return cli.DoExportCapsule(exportCapsuleOpts)
	},
}

var exportSnapshotCmd = &cobra.Command{
	Use:          "snapshot",
	Short:        "Export a full instance backup archive",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return cli.DoExportSnapshot(exportSnapshotOut)
	},
}

var importCapsuleCmd = &cobra.Command{
	Use:          "capsule [path]",
	Short:        "Import a content capsule into this instance",
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return cli.DoImportCapsule(pathArg(args, cli.DefaultCapsuleDir), importCapsuleOpts)
	},
}

var importSnapshotCmd = &cobra.Command{
	Use:          "snapshot <archive.zip>",
	Short:        "Restore a full instance backup archive (destructive)",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return cli.DoImportSnapshot(args[0], importSnapshotYes)
	},
}

var checkCmd = &cobra.Command{
	Use:          "check [path]",
	Short:        "Validate a content capsule",
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return cli.DoCheck(pathArg(args, cli.DefaultCapsuleDir), checkFix)
	},
}

var buildCmd = &cobra.Command{
	Use:          "build [path]",
	Short:        "Build a static, read-only site from a content capsule",
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return cli.DoBuild(pathArg(args, cli.DefaultCapsuleDir), buildOutput, buildBaseURL)
	},
}

// pathArg 取可选的位置参数，缺省回落到规格约定的默认路径。
func pathArg(args []string, fallback string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return fallback
}

func init() {
	exportCapsuleCmd.Flags().
		StringVarP(&exportCapsuleOpts.Output, "output", "o", cli.DefaultCapsuleDir, "output directory (or .zip path with --zip)")
	exportCapsuleCmd.Flags().
		BoolVar(&exportCapsuleOpts.IncludePrivate, "include-private", false, "include private echoes")
	exportCapsuleCmd.Flags().BoolVar(&exportCapsuleOpts.Zip, "zip", false, "pack the capsule into a single .zip")

	exportSnapshotCmd.Flags().
		StringVarP(&exportSnapshotOut, "output", "o", "", "copy the archive to this path (default: keep it under data/files/snapshots)")

	importCapsuleCmd.Flags().
		BoolVar(&importCapsuleOpts.IncludePrivate, "include-private", false, "include private echoes")
	importCapsuleCmd.Flags().
		BoolVar(&importCapsuleOpts.DryRun, "dry-run", false, "report what would change without writing")

	importSnapshotCmd.Flags().
		BoolVar(&importSnapshotYes, "yes", false, "confirm this destructive whole-instance restore")

	checkCmd.Flags().BoolVar(&checkFix, "fix", false, "write back auto-fixable problems (missing ids)")

	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", cli.DefaultDistDir, "output directory")
	buildCmd.Flags().
		StringVar(&buildBaseURL, "base-url", cli.DefaultBaseURL, "site root path when deploying under a sub-path")

	exportCmd.AddCommand(exportCapsuleCmd, exportSnapshotCmd)
	importCmd.AddCommand(importCapsuleCmd, importSnapshotCmd)
	rootCmd.AddCommand(exportCmd, importCmd, checkCmd, buildCmd)
}
