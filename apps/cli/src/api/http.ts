// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Ech0Error } from "../output/console.ts";

/** `commonModel.DEFAULT_SUCCESS_CODE` — every non-failed response uses it. */
export const SUCCESS_CODE = 1;

export interface Envelope<T> {
  code: number;
  msg?: string;
  error_code?: string;
  message_key?: string;
  message_params?: Record<string, unknown>;
  data: T;
}

export type QueryValue = string | number | boolean | readonly string[] | undefined;

export interface RequestSpec {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  path: string;
  query?: Record<string, QueryValue>;
  body?: unknown;
  /** Extra headers, merged over the defaults. */
  headers?: Record<string, string>;
}

export interface RawResponse {
  status: number;
  ok: boolean;
  text: string;
  headers: Headers;
}

export interface CallResult<T> {
  data: T;
  /** Envelope message, e.g. the success text for a body-less operation. */
  message: string;
  /** Response body as received. `--json` prints this verbatim (no re-encoding). */
  text: string;
  status: number;
}

export interface ClientOptions {
  /** Instance root, with or without the `/api` suffix. */
  server: string;
  token?: string;
  lang?: string;
  fetch?: typeof fetch;
}

/** `https://host` → `https://host/api`; keeps an explicit `/api`. */
export function normalizeBaseUrl(server: string): string {
  const trimmed = server.trim().replace(/\/+$/, "");
  if (trimmed === "") {
    throw new Ech0Error("服务地址为空（--server / ECH0_SERVER / profile.server）");
  }
  return trimmed.endsWith("/api") ? trimmed : `${trimmed}/api`;
}

function buildQuery(query: Record<string, QueryValue> | undefined): string {
  if (!query) return "";
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined) continue;
    if (Array.isArray(value)) {
      for (const item of value) params.append(key, String(item));
      continue;
    }
    params.set(key, String(value));
  }
  const encoded = params.toString();
  return encoded === "" ? "" : `?${encoded}`;
}

function parseEnvelope(text: string): Envelope<unknown> | undefined {
  if (text.trim() === "") return undefined;
  try {
    return JSON.parse(text) as Envelope<unknown>;
  } catch {
    return undefined;
  }
}

function toEch0Error(status: number, envelope: Envelope<unknown> | undefined, text: string): Ech0Error {
  const fallback = text.trim() === "" ? `HTTP ${status}` : text.trim();
  const message = envelope?.msg?.trim() || fallback;
  return new Ech0Error(message, {
    status,
    errorCode: envelope?.error_code,
    messageKey: envelope?.message_key,
    messageParams: envelope?.message_params,
  });
}

/** Thin JSON client for a single Ech0 instance. */
export class Client {
  readonly baseUrl: string;

  private readonly fetchImpl: typeof fetch;

  constructor(private readonly opts: ClientOptions) {
    this.baseUrl = normalizeBaseUrl(opts.server);
    this.fetchImpl = opts.fetch ?? fetch;
  }

  /** Sends a request without interpreting the body (catalog fetch, binary). */
  async raw(spec: RequestSpec): Promise<RawResponse> {
    const url = `${this.baseUrl}${spec.path}${buildQuery(spec.query)}`;
    const headers: Record<string, string> = {
      Accept: "application/json",
      ...(spec.body === undefined ? {} : { "Content-Type": "application/json" }),
    };
    if (this.opts.token) headers.Authorization = `Bearer ${this.opts.token}`;
    if (this.opts.lang) headers["X-Locale"] = this.opts.lang;
    Object.assign(headers, spec.headers);

    let response: Response;
    try {
      response = await this.fetchImpl(url, {
        method: spec.method ?? "GET",
        headers,
        body: spec.body === undefined ? undefined : JSON.stringify(spec.body),
      });
    } catch (error) {
      const reason = error instanceof Error ? error.message : String(error);
      throw new Ech0Error(`无法连接 ${this.baseUrl}：${reason}`);
    }

    return {
      status: response.status,
      ok: response.ok,
      text: await response.text(),
      headers: response.headers,
    };
  }

  async call<T>(spec: RequestSpec): Promise<CallResult<T>> {
    const response = await this.raw(spec);
    const envelope = parseEnvelope(response.text);
    if (!response.ok || envelope === undefined || envelope.code !== SUCCESS_CODE) {
      throw toEch0Error(response.status, envelope, response.text);
    }
    return {
      data: envelope.data as T,
      message: envelope.msg ?? "",
      text: response.text,
      status: response.status,
    };
  }
}
