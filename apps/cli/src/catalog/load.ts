// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Client } from "../api/http.ts";
import { Ech0Error } from "../output/console.ts";
import {
  CATALOG_PATH,
  catalogFilePath,
  fingerprintOf,
  readCatalogFile,
  writeCatalogFile,
} from "./store.ts";
import { parseCatalog, type Catalog } from "./types.ts";

export interface LoadedCatalog {
  doc: Catalog;
  bytes: Uint8Array;
  /** sha256 of `bytes`; every request carries it, the server rejects stale ones. */
  fingerprint: string;
  file: string;
  source: "cache" | "network";
}

export interface LoadCatalogOptions {
  server: string;
  token?: string;
  lang?: string;
  /** Config file path; the cache lives beside it. */
  configPath: string;
  fetchImpl?: typeof fetch;
  /** Skip the cache and refetch (412 recovery, `ech0 catalog refresh`). */
  force?: boolean;
}

export async function loadCatalog(options: LoadCatalogOptions): Promise<LoadedCatalog> {
  const file = catalogFilePath(options.configPath, options.server);

  if (!options.force) {
    const cached = readCatalogFile(file);
    if (cached) {
      try {
        return {
          doc: parseCatalog(cached),
          bytes: cached,
          fingerprint: fingerprintOf(cached),
          file,
          source: "cache",
        };
      } catch {
        // Unreadable cache is a miss, not a dead end: refetch and replace it.
      }
    }
  }

  const client = new Client({
    server: options.server,
    token: options.token,
    lang: options.lang,
    fetch: options.fetchImpl,
  });
  const response = await client.raw({ path: CATALOG_PATH });
  if (response.status === 404) {
    throw new Ech0Error("服务端没有 /api/_catalog：实例版本过旧，请升级服务端");
  }
  if (!response.ok) {
    throw new Ech0Error(`无法获取 catalog（HTTP ${response.status}）：${response.text.slice(0, 200)}`);
  }

  const bytes = new TextEncoder().encode(response.text);
  let doc: Catalog;
  try {
    doc = parseCatalog(bytes);
  } catch (error) {
    throw new Ech0Error(error instanceof Error ? error.message : String(error));
  }
  writeCatalogFile(file, bytes);
  return { doc, bytes, fingerprint: fingerprintOf(bytes), file, source: "network" };
}
