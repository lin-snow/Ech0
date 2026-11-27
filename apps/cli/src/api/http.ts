// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Ech0Error } from "../output/console.ts";
import type { ErrorModel } from "./generated/types.gen.ts";

/** The spec's RFC 9457 model, aliased to what this module calls it. */
type Problem = ErrorModel;

export type QueryValue = string | number | boolean | readonly string[] | undefined;

/** One multipart/form-data part; a `Blob` is a file upload. */
export interface FormPart {
  name: string;
  value: string | Blob;
  filename?: string;
}

export interface RequestSpec {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
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
  /** Decoded body; `undefined` for `204 No Content`. */
  body: T;
  /** Response headers, e.g. Set-Cookie after login/refresh. */
  headers: Headers;
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

/** RFC 9457 body of a failed response; `undefined` when the body is not JSON. */
export function parseProblem(text: string): Problem | undefined {
  if (text.trim() === "") return undefined;
  try {
    const parsed: unknown = JSON.parse(text);
    return typeof parsed === "object" && parsed !== null ? (parsed as Problem) : undefined;
  } catch {
    return undefined;
  }
}

/** Maps a failed response onto the CLI's error type. */
export function toEch0Error(status: number, problem: Problem | undefined, text: string): Ech0Error {
  const fallback = text.trim() === "" ? `HTTP ${status}` : text.trim();
  const message = problem?.detail?.trim() || problem?.title?.trim() || fallback;
  return new Ech0Error(message, { status, type: problem?.type });
}

/** Success bodies are resources themselves; an empty body (`204`) decodes to `undefined`. */
function decodeBody(text: string): unknown {
  if (text.trim() === "") return undefined;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
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

  /** Decodes the body of a successful response; any non-2xx throws `Ech0Error`. */
  async call<T>(spec: RequestSpec): Promise<T> {
    return (await this.callWithMeta<T>(spec)).body;
  }

  /** `call`, plus the metadata `--json`, cookie rotation and catalog headers need. */
  async callWithMeta<T>(spec: RequestSpec): Promise<CallResult<T>> {
    const response = await this.request(spec);
    const text = await response.text();
    if (!response.ok) {
      throw toEch0Error(response.status, parseProblem(text), text);
    }
    return {
      body: decodeBody(text) as T,
      headers: response.headers,
      text,
      status: response.status,
    };
  }
}
