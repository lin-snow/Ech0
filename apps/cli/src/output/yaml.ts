// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { UsageError } from "./console.ts";

function encoder(): { stringify(value: unknown, replacer: null, space: number): string } {
  const api = (Bun as unknown as { YAML?: { stringify?: (value: unknown, replacer: null, space: number) => string } }).YAML;
  if (typeof api?.stringify !== "function") {
    throw new UsageError("当前 Bun 没有 Bun.YAML.stringify：升级到 bun >= 1.4 后重试 --yaml");
  }
  return api as { stringify(value: unknown, replacer: null, space: number): string };
}

/**
 * YAML rendering for machine output. `--yaml` and `--json` describe the same
 * document; only the encoding differs, so everything funnels through here.
 */
export function yamlOf(value: unknown): string {
  const text = encoder().stringify(value, null, 2);
  return text.endsWith("\n") ? text : `${text}\n`;
}

/**
 * Re-encodes a JSON payload (usually the raw response body) as YAML. A body
 * that is not JSON is passed through untouched: a server error page should
 * reach the terminal as it arrived, not as a YAML parse failure.
 */
export function yamlFromJson(text: string): string {
  try {
    return yamlOf(JSON.parse(text) as unknown);
  } catch {
    return text.endsWith("\n") ? text : `${text}\n`;
  }
}
