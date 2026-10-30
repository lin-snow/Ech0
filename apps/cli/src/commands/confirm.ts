// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { CatalogOp } from "../catalog/types.ts";
import { UsageError } from "../output/console.ts";
import { confirm } from "../ui/prompt.ts";
import type { FieldValues } from "./operation.ts";

/** The request a destructive operation is about to send, with path values filled in. */
export function destructiveTarget(op: CatalogOp, values: FieldValues): string {
  let path = op.path;
  for (const [field, location] of Object.entries(op.bind)) {
    if (location !== "path") continue;
    const value = values[field];
    if (value !== undefined) path = path.replace(`{${field}}`, encodeURIComponent(String(value)));
  }
  return `${op.method.toUpperCase()} ${path}`;
}

export interface ConfirmationOptions {
  /** `--yes`: the caller already accepted the consequence. */
  yes: boolean;
  /** Machine output never prompts: the flag has to be explicit. */
  json: boolean;
  yaml: boolean;
  /** Defaults to `process.stdin.isTTY`; tests inject it. */
  interactive?: boolean;
  /** Defaults to the terminal prompt; tests inject it. */
  ask?: (question: string) => Promise<boolean>;
}

/**
 * Guards operations whose catalog `effect` is `destructive`.
 *
 * Interactive terminals get a `[y/N]` prompt. Pipes, CI, and `--json`/`--yaml`
 * never prompt — they fail with exit code 2 unless `--yes` was passed.
 */
export async function confirmDestructive(
  op: CatalogOp,
  values: FieldValues,
  options: ConfirmationOptions,
): Promise<void> {
  if (options.yes) return;

  const target = destructiveTarget(op, values);
  if (options.json || options.yaml) {
    throw new UsageError(
      `危险操作需要显式 --yes：${target}`,
      "脚本输出（--json / --yaml）不进行交互确认",
    );
  }
  if ((options.interactive ?? process.stdin.isTTY === true) !== true) {
    throw new UsageError(
      `非交互环境执行危险操作需要显式 --yes：${target}`,
      "stdin 不是终端，无法弹出确认；管道 / CI 请传入 --yes",
    );
  }

  const ask = options.ask ?? confirm;
  if (!(await ask(`确认执行危险操作 ${target}？[y/N] `))) {
    throw new UsageError("已取消");
  }
}
