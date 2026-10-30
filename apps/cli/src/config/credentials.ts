// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

import { Ech0Error } from "../output/console.ts";
import { credentialsFilePath } from "./paths.ts";

const SERVICE = "ech0-cli";

/** Server-side refresh cookie (httpOnly, path /api/auth). */
export const REFRESH_COOKIE_NAME = "ech0_refresh_token";

export type TokenSource = "keychain" | "file";
export type SecretKind = "token" | "cookie";

export interface StoredSecret {
  value: string;
  source: TokenSource;
}

function keychain(): typeof Bun.secrets | undefined {
  // `ECH0_NO_KEYCHAIN=1` forces the file backend (headless CI, tests).
  if (process.env.ECH0_NO_KEYCHAIN) return undefined;
  const secrets: typeof Bun.secrets | undefined = Bun.secrets;
  return typeof secrets?.get === "function" ? secrets : undefined;
}

function secretName(profile: string, kind: SecretKind): string {
  return kind === "token" ? `token:${profile}` : `cookie:${profile}`;
}

function fileKey(profile: string, kind: SecretKind): string {
  return kind === "token" ? profile : `${profile}:cookie`;
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
async function readSecret(
  kind: SecretKind,
  profile: string,
  file: string,
): Promise<StoredSecret | undefined> {
  const secrets = keychain();
  if (secrets) {
    try {
      const value = await secrets.get({ service: SERVICE, name: secretName(profile, kind) });
      if (value) return { value, source: "keychain" };
    } catch {
      // Keychain unavailable (headless Linux, locked keyring): fall through.
    }
  }
  const value = readFileMap(file)[fileKey(profile, kind)];
  return value ? { value, source: "file" } : undefined;
}

/** Returns the backend that ended up holding the secret. */
async function writeSecret(
  kind: SecretKind,
  profile: string,
  value: string,
  file: string,
): Promise<TokenSource> {
  const secrets = keychain();
  if (secrets) {
    try {
      await secrets.set({ service: SERVICE, name: secretName(profile, kind), value });
      return "keychain";
    } catch {
      // Fall back to the file, see readSecret.
    }
  }
  const map = readFileMap(file);
  map[fileKey(profile, kind)] = value;
  writeFileMap(file, map);
  return "file";
}

async function removeSecret(kind: SecretKind, profile: string, file: string): Promise<void> {
  const secrets = keychain();
  if (secrets) {
    try {
      await secrets.delete({ service: SERVICE, name: secretName(profile, kind) });
    } catch {
      // Nothing stored in the keychain.
    }
  }
  const map = readFileMap(file);
  const key = fileKey(profile, kind);
  if (key in map) {
    delete map[key];
    writeFileMap(file, map);
  }
}

export async function readToken(
  profile: string,
  file: string = credentialsFilePath(),
): Promise<StoredSecret | undefined> {
  return readSecret("token", profile, file);
}

export async function writeToken(
  profile: string,
  token: string,
  file: string = credentialsFilePath(),
): Promise<TokenSource> {
  return writeSecret("token", profile, token, file);
}

export async function readRefreshCookie(
  profile: string,
  file: string = credentialsFilePath(),
): Promise<StoredSecret | undefined> {
  return readSecret("cookie", profile, file);
}

export async function writeRefreshCookie(
  profile: string,
  cookie: string,
  file: string = credentialsFilePath(),
): Promise<TokenSource> {
  return writeSecret("cookie", profile, cookie, file);
}

export async function deleteRefreshCookie(
  profile: string,
  file: string = credentialsFilePath(),
): Promise<void> {
  await removeSecret("cookie", profile, file);
}

/** Clears the token and the refresh cookie; missing entries are not an error. */
export async function deleteToken(profile: string, file: string = credentialsFilePath()): Promise<void> {
  await removeSecret("token", profile, file);
  await removeSecret("cookie", profile, file);
}
