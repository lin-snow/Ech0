// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { QueryValue, RequestSpec } from "../api/http.ts";
import type { LoadedCatalog } from "../catalog/load.ts";
import { CATALOG_HEADER } from "../catalog/store.ts";
import type { CatalogOp } from "../catalog/types.ts";
import { Ech0Error } from "../output/console.ts";
import type { Session } from "../session.ts";
import type { CommandOutput } from "./types.ts";

export type FieldValues = Record<string, unknown>;

const METHODS: Record<string, true> = { GET: true, POST: true, PUT: true, DELETE: true };

function requestMethod(method: string): RequestSpec["method"] {
  const upper = method.toUpperCase();
  return (METHODS[upper] ? upper : "GET") as RequestSpec["method"];
}

function queryValue(value: unknown): QueryValue {
  if (Array.isArray(value)) return value.map(String);
  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") return value;
  return JSON.stringify(value);
}

/**
 * Places each provided field where the catalog says it belongs: `{name}`
 * segments for path, query parameters for query, and a JSON object body.
 */
export function buildRequest(op: CatalogOp, values: FieldValues): RequestSpec {
  let path = op.path;
  const query: Record<string, QueryValue> = {};
  const body: Record<string, unknown> = {};

  for (const [field, location] of Object.entries(op.bind)) {
    const value = values[field];
    if (value === undefined) continue;
    if (location === "path") {
      path = path.replace(`{${field}}`, encodeURIComponent(String(value)));
    } else if (location === "query") {
      query[field] = queryValue(value);
    } else {
      body[field] = value;
    }
  }

  const spec: RequestSpec = { method: requestMethod(op.method), path };
  if (Object.keys(query).length > 0) spec.query = query;
  if (Object.keys(body).length > 0) spec.body = body;
  return spec;
}

function render(data: unknown, message: string): string {
  if (data === null || data === undefined) return message || "ok";
  return JSON.stringify(data, null, 2) ?? String(data);
}

export interface InvokeOptions {
  session: Session;
  op: CatalogOp;
  values: FieldValues;
  catalog: LoadedCatalog;
  /** Refetches the catalog; called once when the server reports 412/404. */
  refresh: () => Promise<LoadedCatalog>;
  json: boolean;
}

/** Calls one catalog operation, retrying once with a refreshed catalog. */
export async function invokeOperation(options: InvokeOptions): Promise<CommandOutput> {
  const spec = buildRequest(options.op, options.values);
  let catalog = options.catalog;
  let catalogRetried = false;
  let authRetried = false;

  for (;;) {
    try {
      const result = await options.session.client.call<unknown>({
        ...spec,
        headers: { [CATALOG_HEADER]: catalog.fingerprint },
      });
      return { human: render(result.data, result.message), json: result.text };
    } catch (error) {
      if (!(error instanceof Ech0Error)) throw error;

      const status = error.details.status;
      if (!catalogRetried && (status === 412 || status === 404)) {
        catalogRetried = true;
        catalog = await options.refresh();
        continue;
      }
      if (!authRetried && status === 401 && (await options.session.refresh())) {
        authRetried = true;
        continue;
      }
      throw error;
    }
  }
}
