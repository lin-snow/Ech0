// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** What every command action returns: one human rendering, one machine payload. */
export interface CommandOutput {
  human: string;
  json: string;
}
