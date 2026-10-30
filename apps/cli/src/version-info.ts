// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import pkg from "../package.json" with { type: "json" };

/**
 * Version reported by `ech0 version` / `--version`.
 *
 * Release builds pass `--define process.env.ECH0_CLI_VERSION="<tag>"` to
 * `bun build --compile`, which inlines the value here. Dev builds and `bun run`
 * fall back to package.json so the command still reports something meaningful.
 */
export const CLI_VERSION: string = process.env.ECH0_CLI_VERSION ?? pkg.version;
