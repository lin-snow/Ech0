// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { z } from "zod";

/**
 * Shape of the server's operation catalog (`GET /api/_catalog`), generated from
 * the Huma document. Unknown fields are preserved so a newer server can add
 * metadata without breaking older clients.
 */
export const CatalogOpSchema = z
  .object({
    summary: z.string(),
    method: z.string(),
    path: z.string(),
    scopes: z.array(z.string()).default([]),
    effect: z.string().default("write"),
    kind: z.string().default("json"),
    input: z.record(z.string(), z.unknown()).default({}),
    bind: z.record(z.string(), z.string()).default({}),
    deprecated: z.boolean().default(false),
  })
  .catchall(z.unknown());

export const CatalogSchema = z.object({
  ops: z.record(z.string(), CatalogOpSchema),
});

export type CatalogOp = z.infer<typeof CatalogOpSchema>;
export type Catalog = z.infer<typeof CatalogSchema>;

export function parseCatalog(bytes: Uint8Array): Catalog {
  let raw: unknown;
  try {
    raw = JSON.parse(new TextDecoder().decode(bytes));
  } catch (error) {
    throw new Error(`catalog 不是合法 JSON：${(error as Error).message}`);
  }
  const parsed = CatalogSchema.safeParse(raw);
  if (!parsed.success) {
    throw new Error(`catalog 结构无效：${parsed.error.issues.map((issue) => issue.message).join("; ")}`);
  }
  return parsed.data;
}
