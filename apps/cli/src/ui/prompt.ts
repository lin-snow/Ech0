// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { UsageError } from "../output/console.ts";

/** Reads one line from stdin; `-` arguments use it as the payload. */
export async function readStdin(): Promise<string> {
  if (process.stdin.isTTY) {
    throw new UsageError("需要从 stdin 读取内容，但 stdin 是终端");
  }
  return Bun.stdin.text();
}
