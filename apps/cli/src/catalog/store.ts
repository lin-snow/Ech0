// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";

/** Fingerprint header: sent on every request, stamped on every `/api` response. */
export const CATALOG_HEADER = "X-Ech0-Catalog";
/** Catalog route relative to the API base. */
export const CATALOG_PATH = "/_catalog";

export function fingerprintOf(bytes: Uint8Array): string {
  return createHash("sha256").update(bytes).digest("hex");
}

function hostKey(server: string): string {
  try {
    return new URL(server).host.replaceAll(":", "_");
  } catch {
    return server.replace(/[^a-zA-Z0-9.-]+/g, "_");
  }
}

/** One cache file per instance: `<config dir>/catalog/<host>.json`. */
export function catalogFilePath(configPath: string, server: string): string {
  return join(dirname(configPath), "catalog", `${hostKey(server)}.json`);
}

export function readCatalogFile(file: string): Uint8Array | undefined {
  try {
    return readFileSync(file);
  } catch {
    return undefined;
  }
}

export function writeCatalogFile(file: string, bytes: Uint8Array): void {
  mkdirSync(dirname(file), { recursive: true, mode: 0o700 });
  writeFileSync(file, bytes, { mode: 0o600 });
}
