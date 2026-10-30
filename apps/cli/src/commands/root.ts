// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import pkg from "../../package.json" with { type: "json" };

/** Builds the `ech0` command tree. Subcommands are registered here as they land. */
export function createRootCommand(): Command {
  const program = new Command("ech0")
    .description(pkg.description)
    .version(pkg.version, "-v, --version")
    .showHelpAfterError()
    .showSuggestionAfterError()
    .configureHelp({ sortSubcommands: true });

  program.action(() => program.outputHelp());

  return program;
}
