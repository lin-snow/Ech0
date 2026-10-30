// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { z } from "zod";

/** One named event an SSE operation may emit (advisory vocabulary). */
export const SseEventSchema = z.object({
  name: z.string(),
  data: z.record(z.string(), z.string()).optional(),
});

/** Non-JSON request/response encodings; absent means plain JSON both ways. */
export const TransportSchema = z.object({
  request: z
    .object({
      content_type: z.string().optional(),
      /** `query` means the access token travels as `?token=` (websocket). */
      auth: z.string().optional(),
    })
    .optional(),
  response: z
    .object({
      kind: z.string(),
      content_type: z.string().optional(),
      filename: z.string().optional(),
      events: z.array(SseEventSchema).optional(),
    })
    .optional(),
});

/**
 * Shape of the server's operation catalog (`GET /api/_catalog`), generated from
 * the Huma document and the hand-written raw operations. Unknown fields are
 * preserved so a newer server can add metadata without breaking older clients.
 */
export const CatalogOpSchema = z
  .object({
    summary: z.string(),
    method: z.string(),
    path: z.string(),
    /** `api` (default) or `ws`; websocket paths hang under `/ws`. */
    mount: z.string().optional(),
    scopes: z
      .array(z.string())
      .nullish()
      .transform((value) => value ?? []),
    effect: z.string().default("write"),
    kind: z.string().default("json"),
    input: z.record(z.string(), z.unknown()).default({}),
    bind: z.record(z.string(), z.string()).default({}),
    transport: TransportSchema.optional(),
    deprecated: z.boolean().default(false),
  })
  .catchall(z.unknown());

export const CatalogSchema = z.object({
  ops: z.record(z.string(), CatalogOpSchema),
});

export type CatalogOp = z.infer<typeof CatalogOpSchema>;
export type CatalogTransport = z.infer<typeof TransportSchema>;
export type SseEvent = z.infer<typeof SseEventSchema>;
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
