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

/** One multipart/form-data part; a `Blob` is a file upload. */
export interface FormPart {
  name: string;
  value: string | Blob;
  filename?: string;
}

export interface RequestSpec {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  path: string;
  query?: Record<string, QueryValue>;
  body?: unknown;
  /** Multipart body; mutually exclusive with `body`. */
  form?: FormPart[];
  /** Extra headers, merged over the defaults. */
  headers?: Record<string, string>;
  /** Overrides the default `Accept` header (`text/event-stream`, any type). */
  accept?: string;
  /** Cancels the request, e.g. on Ctrl-C while a stream is open. */
  signal?: AbortSignal;
}

export interface RawResponse {
  status: number;
  ok: boolean;
  text: string;
  headers: Headers;
}

export interface CallResult<T> {
  data: T;
  /** Response headers, e.g. Set-Cookie after login/refresh. */
  headers: Headers;
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

export function parseEnvelope(text: string): Envelope<unknown> | undefined {
  if (text.trim() === "") return undefined;
  try {
    return JSON.parse(text) as Envelope<unknown>;
  } catch {
    return undefined;
  }
}

export function toEch0Error(status: number, envelope: Envelope<unknown> | undefined, text: string): Ech0Error {
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

  /** Swaps the bearer token after a refresh, keeping the same instance. */
  setToken(token: string): void {
    this.opts.token = token;
  }

  /** Absolute URL for one operation path, honouring the `/api` base. */
  url(spec: Pick<RequestSpec, "path" | "query">): string {
    return `${this.baseUrl}${spec.path}${buildQuery(spec.query)}`;
  }

  /** WebSocket URL for a path mounted under `/ws` (not under `/api`). */
  websocketUrl(path: string, query: Record<string, QueryValue> = {}): string {
    const origin = new URL(this.baseUrl).origin.replace(/^http/, "ws");
    return `${origin}/ws${path}${buildQuery(query)}`;
  }

  /**
   * Sends a request and hands back the unread Response, so callers can stream
   * it (SSE, downloads) instead of buffering everything.
   */
  async request(spec: RequestSpec): Promise<Response> {
    const headers: Record<string, string> = {
      Accept: spec.accept ?? "application/json",
    };
    if (spec.body !== undefined) headers["Content-Type"] = "application/json";
    if (this.opts.token) headers.Authorization = `Bearer ${this.opts.token}`;
    if (this.opts.lang) headers["X-Locale"] = this.opts.lang;
    Object.assign(headers, spec.headers);

    let payload: string | FormData | undefined;
    if (spec.form !== undefined) {
      const data = new FormData();
      for (const part of spec.form) {
        if (typeof part.value === "string") data.append(part.name, part.value);
        else data.append(part.name, part.value, part.filename);
      }
      payload = data;
    } else if (spec.body !== undefined) {
      payload = JSON.stringify(spec.body);
    }

    try {
      return await this.fetchImpl(this.url(spec), {
        method: spec.method ?? "GET",
        headers,
        body: payload,
        signal: spec.signal,
      });
    } catch (error) {
      const reason = error instanceof Error ? error.message : String(error);
      throw new Ech0Error(`无法连接 ${this.baseUrl}：${reason}`);
    }
  }

  /** Sends a request without interpreting the body (catalog fetch). */
  async raw(spec: RequestSpec): Promise<RawResponse> {
    const response = await this.request(spec);
    return {
      status: response.status,
      ok: response.ok,
      text: await response.text(),
      headers: response.headers,
    };
  }

  async call<T>(spec: RequestSpec): Promise<CallResult<T>> {
    const response = await this.request(spec);
    const text = await response.text();
    const envelope = parseEnvelope(text);
    if (!response.ok || envelope === undefined || envelope.code !== SUCCESS_CODE) {
      throw toEch0Error(response.status, envelope, text);
    }
    return {
      data: envelope.data as T,
      headers: response.headers,
      message: envelope.msg ?? "",
      text,
      status: response.status,
    };
  }
}
