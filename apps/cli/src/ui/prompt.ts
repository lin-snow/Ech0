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

/** Reads a password without echo; pipes must use `--password-stdin`. */
export async function readPassword(question: string): Promise<string> {
  if (!process.stdin.isTTY) {
    throw new UsageError("非交互环境请用 --password-stdin 提供密码");
  }

  process.stderr.write(question);
  const stdin = process.stdin;
  stdin.setRawMode?.(true);
  const reader = Bun.stdin.stream().getReader();
  let value = "";

  try {
    for (;;) {
      const { value: chunk, done } = await reader.read();
      if (done) break;
      for (const byte of chunk ?? new Uint8Array()) {
        if (byte === 13 || byte === 10) return value;
        if (byte === 3) throw new UsageError("已取消");
        if (byte === 127 || byte === 8) {
          value = value.slice(0, -1);
          continue;
        }
        value += String.fromCharCode(byte);
      }
    }
    return value;
  } finally {
    stdin.setRawMode?.(false);
    reader.releaseLock();
    process.stderr.write("\n");
  }
}
