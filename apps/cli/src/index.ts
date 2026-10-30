#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command, CommanderError } from "commander";

import { createRootCommand } from "./commands/root.ts";
import {
  Ech0Error,
  EXIT_ERROR,
  EXIT_USAGE,
  UsageError,
  colors,
  printErr,
} from "./output/console.ts";

function report(error: unknown): number {
  if (error instanceof CommanderError) return error.exitCode === 0 ? 0 : EXIT_USAGE;

  const palette = colors();
  if (error instanceof UsageError) {
    printErr(`${palette.red("错误")} ${error.message}`);
    return EXIT_USAGE;
  }
  if (error instanceof Ech0Error) {
    const status = error.details.status;
    const detail = status ? palette.dim(` (HTTP ${status}${error.details.errorCode ? ` · ${error.details.errorCode}` : ""})`) : "";
    printErr(`${palette.red("错误")} ${error.message}${detail}`);
    if (error.details.messageKey) printErr(palette.dim(`  key: ${error.details.messageKey}`));
    if (status === 401) printErr(palette.dim("  提示：token 无效或已过期，运行 `ech0 auth status --verify`"));
    if (status === 403) printErr(palette.dim("  提示：token 缺少所需 scope，请在面板重新铸造"));
    return EXIT_ERROR;
  }
  printErr(`${palette.red("错误")} ${error instanceof Error ? error.message : String(error)}`);
  return EXIT_ERROR;
}

function overrideExit(command: Command): void {
  command.exitOverride();
  for (const child of command.commands) overrideExit(child);
}

const program = await createRootCommand();
overrideExit(program);

try {
  await program.parseAsync(process.argv);
} catch (error) {
  process.exitCode = report(error);
}
