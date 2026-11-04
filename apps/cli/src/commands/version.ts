// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { CommandOutput } from "./types.ts";

/** Local CLI version. Server version and instance identity are `ech0 info`. */
export function version(localVersion: string): CommandOutput {
  return {
    human: `ech0 cli ${localVersion}`,
    json: JSON.stringify({ cli: localVersion }, null, 2),
  };
}
