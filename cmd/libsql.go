// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cmd

import (
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "prepare-libsql", Short: "Prepare a new remote libSQL database for Ech0", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return database.PrepareRemoteSchema(cmd.Context(), config.Config().Database)
		},
	})
}
