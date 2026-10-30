// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { CommandOutput } from "../commands/types.ts";
import { print } from "./console.ts";
import { yamlFromJson } from "./yaml.ts";

/** The output modes every command shares. */
export interface OutputMode {
  json: boolean;
  yaml: boolean;
}

/**
 * Writes one command result: raw JSON as served, the same document as YAML, or
 * the human rendering. `--json` keeps the bytes, so int64 values stay exact.
 */
export function printResult(result: CommandOutput, mode: OutputMode): void {
  if (mode.json) {
    print(result.json);
    return;
  }
  if (mode.yaml) {
    print(yamlFromJson(result.json).trimEnd());
    return;
  }
  print(result.human);
}
