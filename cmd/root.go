// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cmd

import (
	"os"

	"github.com/lin-snow/ech0/internal/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ech0",
	Short: "A self-hosted, lightweight microblog platform for personal thoughts",
	Long:  `Ech0 is a new-generation open-source, self-hosted, lightweight publishing platform focused on the flow of personal thoughts.`,

	Run: func(cmd *cobra.Command, args []string) {
		cli.DoTui()
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Ech0 TUI",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoTui()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoVersion()
	},
}

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Print the Ech0 logo",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoHello()
	},
}

func init() {
	cobra.MousetrapHelpText = ""
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(helloCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
