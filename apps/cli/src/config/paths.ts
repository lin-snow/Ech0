// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { homedir } from "node:os";
import { dirname, join } from "node:path";

export interface PathEnv {
  ECH0_CONFIG?: string;
  XDG_CONFIG_HOME?: string;
  HOME?: string;
  [key: string]: string | undefined;
}

/** `$ECH0_CONFIG` → `$XDG_CONFIG_HOME/ech0/config.json` → `~/.config/ech0/config.json`. */
export function configFilePath(env: PathEnv = process.env): string {
  if (env.ECH0_CONFIG) return env.ECH0_CONFIG;
  const base = env.XDG_CONFIG_HOME || join(env.HOME || homedir(), ".config");
  return join(base, "ech0", "config.json");
}

/** Tokens live next to the config: `<config dir>/credentials.json`. */
export function credentialsFilePath(env: PathEnv = process.env): string {
  return join(dirname(configFilePath(env)), "credentials.json");
}
