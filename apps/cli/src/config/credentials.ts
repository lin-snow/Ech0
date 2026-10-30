// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

import { Ech0Error } from "../output/console.ts";
import { credentialsFilePath } from "./paths.ts";

const SERVICE = "ech0-cli";

export type TokenSource = "keychain" | "file";

export interface StoredToken {
  token: string;
  source: TokenSource;
}

function keychain(): typeof Bun.secrets | undefined {
  // `ECH0_NO_KEYCHAIN=1` forces the file backend (headless CI, tests).
  if (process.env.ECH0_NO_KEYCHAIN) return undefined;
  const secrets: typeof Bun.secrets | undefined = Bun.secrets;
  return typeof secrets?.get === "function" ? secrets : undefined;
}

function readFileMap(file: string): Record<string, string> {
  let raw: string;
  try {
    raw = readFileSync(file, "utf8");
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") return {};
    throw new Ech0Error(`无法读取凭据文件 ${file}：${(error as Error).message}`);
  }
  try {
    const parsed = JSON.parse(raw) as unknown;
    if (typeof parsed !== "object" || parsed === null) return {};
    const map: Record<string, string> = {};
    for (const [name, value] of Object.entries(parsed)) {
      if (typeof value === "string") map[name] = value;
    }
    return map;
  } catch {
    return {};
  }
}

function writeFileMap(file: string, map: Record<string, string>): void {
  mkdirSync(dirname(file), { recursive: true, mode: 0o700 });
  writeFileSync(file, `${JSON.stringify(map, null, 2)}\n`, { mode: 0o600 });
}

/** Keychain first (`Bun.secrets`), then a 0600 file next to the config. */
export async function readToken(
  profile: string,
  file: string = credentialsFilePath(),
): Promise<StoredToken | undefined> {
  const secrets = keychain();
  if (secrets) {
    try {
      const value = await secrets.get({ service: SERVICE, name: `token:${profile}` });
      if (value) return { token: value, source: "keychain" };
    } catch {
      // Keychain unavailable (headless Linux, locked keyring): fall through.
    }
  }
  const value = readFileMap(file)[profile];
  return value ? { token: value, source: "file" } : undefined;
}

/** Returns the backend that ended up holding the token. */
export async function writeToken(
  profile: string,
  token: string,
  file: string = credentialsFilePath(),
): Promise<TokenSource> {
  const secrets = keychain();
  if (secrets) {
    try {
      await secrets.set({ service: SERVICE, name: `token:${profile}`, value: token });
      return "keychain";
    } catch {
      // Fall back to the file, see readToken.
    }
  }
  const map = readFileMap(file);
  map[profile] = token;
  writeFileMap(file, map);
  return "file";
}

/** Clears the token from both backends; missing entries are not an error. */
export async function deleteToken(profile: string, file: string = credentialsFilePath()): Promise<void> {
  const secrets = keychain();
  if (secrets) {
    try {
      await secrets.delete({ service: SERVICE, name: `token:${profile}` });
    } catch {
      // Nothing stored in the keychain.
    }
  }
  const map = readFileMap(file);
  if (profile in map) {
    delete map[profile];
    writeFileMap(file, map);
  }
}
