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

/** What the caller can do next, keyed by the failure the server reported. */
function errorHints(error: Ech0Error): string[] {
  const status = error.details.status;
  if (status === 401) {
    return ["token 无效或已过期：运行 `ech0 auth status --verify` 检查", "重新登录：`ech0 auth login`"];
  }
  if (status === 403) {
    return ["token 缺少所需 scope：在面板重新铸造，或 `ech0 auth login --scopes <list>`"];
  }
  if (status === undefined) {
    return ["无法连接实例：检查 --server / profile（`ech0 config list`）与网络"];
  }
  if (status >= 500) {
    return ["服务端错误：稍后重试；必要时 `ech0 catalog refresh`"];
  }
  return [];
}

function report(error: unknown): number {
  // A namespace invoked bare prints its help and commander calls that an error;
  // help is the answer, not a usage failure.
  if (error instanceof CommanderError) {
    return error.exitCode === 0 || error.code === "commander.help" ? 0 : EXIT_USAGE;
  }

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
    for (const hint of errorHints(error)) printErr(palette.dim(`  提示：${hint}`));
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
