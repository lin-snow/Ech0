// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { createInterface } from "node:readline/promises";

import { UsageError } from "../output/console.ts";

/** Reads one line from stdin; `-` arguments use it as the payload. */
export async function readStdin(): Promise<string> {
  if (process.stdin.isTTY) {
    throw new UsageError("需要从 stdin 读取内容，但 stdin 是终端");
  }
  return Bun.stdin.text();
}

/** Interactive confirmation; non-interactive runs must pass `--yes`. */
export async function confirm(question: string, assumeYes: boolean): Promise<boolean> {
  if (assumeYes) return true;
  if (!process.stdin.isTTY) {
    throw new UsageError("非交互环境需要 --yes 才能继续");
  }
  const rl = createInterface({ input: process.stdin, output: process.stdout });
  try {
    const answer = (await rl.question(`${question} [y/N] `)).trim().toLowerCase();
    return answer === "y" || answer === "yes";
  } finally {
    rl.close();
  }
}
