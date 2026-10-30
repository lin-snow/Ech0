// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { UsageError } from "../output/console.ts";

/** Starting point for `--edit`: the fields an Echo can carry, nothing filled in. */
export function payloadTemplate(): string {
  return `${JSON.stringify(
    {
      content: "",
      tags: [],
      echo_files: [],
      private: false,
      layout: "waterfall",
      extension: null,
    },
    null,
    2,
  )}\n`;
}

export function parsePayload(raw: string, source: string): Record<string, unknown> {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (error) {
    throw new UsageError(`${source} 不是合法 JSON：${(error as Error).message}`);
  }
  if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
    throw new UsageError(`${source} 需要是一个 JSON 对象`);
  }
  return parsed as Record<string, unknown>;
}

/** Reads `--from`: a file path, or `-` for stdin (agents and pipelines). */
export async function readPayload(source: string): Promise<Record<string, unknown>> {
  if (source === "-") {
    return parsePayload(await Bun.stdin.text(), "stdin");
  }
  try {
    return parsePayload(readFileSync(source, "utf8"), source);
  } catch (error) {
    if (error instanceof UsageError) throw error;
    throw new UsageError(`无法读取 ${source}：${(error as Error).message}`);
  }
}

/**
 * `--edit`: opens the template in `$EDITOR`, then parses what came back. The
 * file is scratch space, not a draft: nothing survives the command.
 */
export async function editPayload(): Promise<Record<string, unknown>> {
  const editor = process.env.VISUAL || process.env.EDITOR || "vi";
  const dir = mkdtempSync(join(tmpdir(), "ech0-echo-"));
  const file = join(dir, "echo.json");
  writeFileSync(file, payloadTemplate(), { mode: 0o600 });

  try {
    const [command, ...args] = editor.split(/\s+/);
    if (command === undefined || command === "") {
      throw new UsageError("$EDITOR 为空：请设置 EDITOR 或改用 --from");
    }
    const result = Bun.spawnSync([command, ...args, file], {
      stdio: ["inherit", "inherit", "inherit"],
    });
    if (result.exitCode !== 0) {
      throw new UsageError(`编辑器退出码 ${result.exitCode}，已取消`);
    }
    const payload = parsePayload(readFileSync(file, "utf8"), file);
    if (!hasContent(payload)) {
      throw new UsageError("没有可提交的内容：content / echo_files / extension 都是空的");
    }
    return payload;
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

export function hasContent(payload: Record<string, unknown>): boolean {
  const content = typeof payload.content === "string" ? payload.content.trim() : "";
  const files = Array.isArray(payload.echo_files) ? payload.echo_files.length : 0;
  const extension = payload.extension;
  return content !== "" || files > 0 || (typeof extension === "object" && extension !== null);
}
