// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

import { Ech0Error, UsageError } from "../output/console.ts";
import { configFilePath } from "./paths.ts";

export interface Profile {
  /** Instance root, e.g. `https://memo.sn0w.fyi`. */
  server: string;
  /** Last verified username, for `auth status` without a round trip. */
  user?: string;
  isAdmin?: boolean;
}

export interface Config {
  default?: string;
  profiles: Record<string, Profile>;
}

const EMPTY: Config = { profiles: {} };

export function loadConfig(path: string = configFilePath()): Config {
  let raw: string;
  try {
    raw = readFileSync(path, "utf8");
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") return structuredClone(EMPTY);
    throw new Ech0Error(`无法读取配置 ${path}：${(error as Error).message}`);
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    throw new Ech0Error(`配置文件不是合法 JSON：${path}`);
  }
  if (typeof parsed !== "object" || parsed === null) {
    throw new Ech0Error(`配置文件结构无效：${path}`);
  }

  const candidate = parsed as Partial<Config>;
  const profiles: Record<string, Profile> = {};
  for (const [name, value] of Object.entries(candidate.profiles ?? {})) {
    if (typeof value?.server !== "string") continue;
    profiles[name] = {
      server: value.server,
      user: typeof value.user === "string" ? value.user : undefined,
      isAdmin: typeof value.isAdmin === "boolean" ? value.isAdmin : undefined,
    };
  }
  return { default: typeof candidate.default === "string" ? candidate.default : undefined, profiles };
}

export function saveConfig(config: Config, path: string = configFilePath()): void {
  mkdirSync(dirname(path), { recursive: true, mode: 0o700 });
  writeFileSync(path, `${JSON.stringify(config, null, 2)}\n`, { mode: 0o600 });
}

/** Inserts or replaces a profile, keeping the config immutable for callers. */
export function withProfile(config: Config, name: string, profile: Profile): Config {
  if (name.trim() === "") throw new UsageError("profile 名称不能为空");
  const profiles = { ...config.profiles, [name]: profile };
  return { default: config.default ?? name, profiles };
}

export function withoutProfile(config: Config, name: string): Config {
  const { [name]: _removed, ...rest } = config.profiles;
  const defaultName = config.default === name ? Object.keys(rest)[0] : config.default;
  return { default: defaultName, profiles: rest };
}

/** Resolves the active profile name: flag → `ECH0_PROFILE` → config default. */
export function activeProfileName(
  config: Config,
  flag: string | undefined,
  env: NodeJS.ProcessEnv = process.env,
): string | undefined {
  return flag || env.ECH0_PROFILE || config.default;
}
