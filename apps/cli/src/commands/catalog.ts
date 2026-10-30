// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { loadCatalog, type LoadedCatalog } from "../catalog/load.ts";
import { catalogFilePath } from "../catalog/store.ts";
import { readToken } from "../config/credentials.ts";
import { colors } from "../output/console.ts";
import { loadContext, resolveTarget, type CliContext, type GlobalFlags } from "../session.ts";
import type { CommandOutput } from "./types.ts";

/** Refetches the catalog for the profile addressed by `flags`. */
export async function refreshCatalog(flags: GlobalFlags): Promise<LoadedCatalog> {
  const ctx = loadContext(flags);
  const target = resolveTarget(ctx);
  const stored = await readToken(target.profile, ctx.credentialsPath);
  return loadCatalog({
    server: target.server,
    token: flags.token ?? ctx.env.ECH0_TOKEN ?? stored?.value,
    lang: flags.lang ?? ctx.env.ECH0_LANG,
    configPath: ctx.configPath,
    force: true,
  });
}

export function summarize(loaded: LoadedCatalog): CommandOutput {
  const { green, dim } = colors();
  return {
    human: [
      `${green("✓")} catalog 已更新：${Object.keys(loaded.doc.ops).length} 个操作`,
      dim(`${loaded.file} · ${loaded.fingerprint}`),
    ].join("\n"),
    json: JSON.stringify(
      {
        file: loaded.file,
        fingerprint: loaded.fingerprint,
        ops: Object.keys(loaded.doc.ops).length,
        source: loaded.source,
      },
      null,
      2,
    ),
  };
}

/** Prints the cache path the CLI uses for this profile's instance. */
export function cachePath(ctx: CliContext): CommandOutput {
  const target = resolveTarget(ctx);
  const file = catalogFilePath(ctx.configPath, target.server);
  return {
    human: file,
    json: JSON.stringify({ file, server: target.server }, null, 2),
  };
}
