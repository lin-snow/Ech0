// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { basename, join } from "node:path";
import { rename, unlink } from "node:fs/promises";
import { once } from "node:events";

import {
  toEch0Error,
  parseEnvelope,
  type FormPart,
  type QueryValue,
  type RequestSpec,
} from "../api/http.ts";
import { parseSse } from "../api/sse.ts";
import type { LoadedCatalog } from "../catalog/load.ts";
import { CATALOG_HEADER } from "../catalog/store.ts";
import type { CatalogOp } from "../catalog/types.ts";
import { Ech0Error, UsageError, colors, print, printErr } from "../output/console.ts";
import { createStreamRenderer } from "../output/stream.ts";
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
 * Multipart operations go through `buildFormRequest` instead.
 */
export function buildRequest(op: CatalogOp, values: FieldValues): RequestSpec {
  let path = op.path;
  const query: Record<string, QueryValue> = {};
  const body: Record<string, unknown> = {};
  let hasBody = false;

  for (const [field, location] of Object.entries(op.bind)) {
    const value = values[field];
    if (location === "path") {
      if (value !== undefined) path = path.replace(`{${field}}`, encodeURIComponent(String(value)));
    } else if (location === "query") {
      if (value !== undefined) query[field] = queryValue(value);
    } else if (location === "body") {
      // Huma marks a struct body required even when every field is optional;
      // send `{}` so `ech0 echo list` without flags is a valid request.
      hasBody = true;
      if (value !== undefined) body[field] = value;
    }
  }

  const spec: RequestSpec = { method: requestMethod(op.method), path };
  if (Object.keys(query).length > 0) spec.query = query;
  if (hasBody) spec.body = body;
  return spec;
}

function isBinaryField(schema: unknown): boolean {
  return (
    typeof schema === "object" &&
    schema !== null &&
    (schema as Record<string, unknown>).format === "binary"
  );
}

/** Reads a `--file` value into a multipart part; `-` streams stdin. */
async function filePart(name: string, raw: unknown, fileName?: string): Promise<FormPart> {
  const path = String(raw);
  if (path === "-") {
    const bytes = await Bun.stdin.bytes();
    return { name, value: new Blob([bytes]), filename: fileName ?? "stdin" };
  }
  const file = Bun.file(path);
  if (!(await file.exists())) {
    throw new UsageError(`文件不存在：${path}`);
  }
  return { name, value: file, filename: fileName ?? basename(path) };
}

/** Builds a multipart request: form fields travel as parts, files as blobs. */
export async function buildFormRequest(
  op: CatalogOp,
  values: FieldValues,
  options: { fileName?: string } = {},
): Promise<RequestSpec> {
  const properties = (op.input as { properties?: Record<string, unknown> }).properties ?? {};
  let path = op.path;
  const query: Record<string, QueryValue> = {};
  const form: FormPart[] = [];

  for (const [field, location] of Object.entries(op.bind)) {
    const value = values[field];
    if (value === undefined) continue;
    if (location === "path") {
      path = path.replace(`{${field}}`, encodeURIComponent(String(value)));
      continue;
    }
    if (location === "query") {
      query[field] = queryValue(value);
      continue;
    }
    if (location !== "form") continue;

    if (isBinaryField(properties[field])) {
      for (const item of Array.isArray(value) ? value : [value]) {
        form.push(await filePart(field, item, options.fileName));
      }
      continue;
    }
    for (const item of Array.isArray(value) ? value : [value]) {
      form.push({ name: field, value: typeof item === "string" ? item : JSON.stringify(item) });
    }
  }

  const spec: RequestSpec = { method: requestMethod(op.method), path, form };
  if (Object.keys(query).length > 0) spec.query = query;
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
  /** `--yaml` re-encodes the result document; streams and bytes have none. */
  yaml?: boolean;
  /** `--output` for downloads: a path, or `-` for stdout. */
  output?: string;
  /** `--from` / `--edit`: a payload that flags override. */
  payload?: Record<string, unknown>;
  /** `--tag`, repeated. */
  tags?: string[];
  /** `--file-id`, repeated: attachments uploaded by `ech0 file upload`. */
  fileIds?: string[];
  /** `--file-name`: multipart filename, so stdin uploads keep an extension. */
  uploadFileName?: string;
  /** True for echo.create / echo.update: enable payload composition. */
  compose?: boolean;
}

/** Calls one operation, dispatching on the catalog's transport description. */
export async function invokeOperation(options: InvokeOptions): Promise<CommandOutput> {
  const kind = options.op.transport?.response?.kind ?? options.op.kind;
  if (options.yaml === true && kind !== "json") {
    throw new UsageError("--yaml 只适用于 JSON 结果：流式用 --json，二进制用 --output <path>");
  }
  switch (kind) {
    case "stream":
      return streamOperation(options);
    case "binary":
      return downloadOperation(options);
    case "websocket":
      return websocketOperation(options);
    default:
      return callOperation(options);
  }
}

/** Operations that authenticate with `?token=` instead of the header. */
function withQueryAuth(spec: RequestSpec, op: CatalogOp, session: Session): RequestSpec {
  if (op.transport?.request?.auth !== "query") return spec;
  return { ...spec, query: { ...(spec.query ?? {}), token: session.token } };
}

function hasFormFields(op: CatalogOp): boolean {
  return Object.values(op.bind).includes("form");
}

/** Calls one operation, refreshing the catalog or the token at most once each. */
async function callWithRetry(
  options: InvokeOptions,
  spec: RequestSpec,
): Promise<{ data: unknown; text: string; message: string }> {
  let catalog = options.catalog;
  let catalogRetried = false;
  let authRetried = false;

  for (;;) {
    try {
      const result = await options.session.client.call<unknown>({
        ...spec,
        headers: { [CATALOG_HEADER]: catalog.fingerprint },
      });
      return { data: result.data, text: result.text, message: result.message };
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

/** Accepts both `[{"name":"x"}]` and `["x"]`; strings become tag objects. */
function normalizeTags(raw: unknown): Record<string, unknown>[] {
  if (raw === undefined || raw === null) return [];
  if (!Array.isArray(raw)) {
    throw new UsageError('tags 需要是数组：[{"name":"x"}] 或 ["x"]');
  }
  return raw.map((item) => {
    if (typeof item === "string") return { name: item };
    if (typeof item === "object" && item !== null) return item as Record<string, unknown>;
    throw new UsageError("tags 的元素需要是字符串或 {name: ...} 对象");
  });
}

function existingEchoFiles(raw: unknown): Record<string, unknown>[] {
  if (raw === undefined || raw === null) return [];
  if (!Array.isArray(raw)) {
    throw new UsageError('echo_files 需要是数组：[{"file_id":"...","sort_order":0}]');
  }
  return raw.filter(
    (item): item is Record<string, unknown> => typeof item === "object" && item !== null,
  );
}

function nextSortOrder(files: Record<string, unknown>[]): number {
  const orders = files
    .map((file) => (typeof file.sort_order === "number" ? file.sort_order : -1))
    .filter((order) => order >= 0);
  return orders.length === 0 ? 0 : Math.max(...orders) + 1;
}

/** Payload file first, flags override, `--tag` / `--file-id` append. */
async function composeEchoValues(options: InvokeOptions): Promise<FieldValues> {
  const payload = options.payload ?? {};
  const values: FieldValues = { ...payload };
  for (const [field, value] of Object.entries(options.values)) {
    // tags and echo_files accumulate: the payload keeps what the file says,
    // then flags append, then --tag / --file-id append.
    if (value === undefined || field === "tags" || field === "echo_files") continue;
    values[field] = value;
  }

  const tags = [...normalizeTags(payload.tags), ...normalizeTags(options.values.tags)];
  for (const name of options.tags ?? []) tags.push({ name });
  if (tags.length > 0) values.tags = tags;

  const files = [...existingEchoFiles(payload.echo_files), ...existingEchoFiles(options.values.echo_files)];
  let order = nextSortOrder(files);
  for (const fileID of options.fileIds ?? []) {
    files.push({ file_id: fileID, sort_order: order });
    order += 1;
  }
  if (files.length > 0) values.echo_files = files;

  return values;
}

async function callOperation(options: InvokeOptions): Promise<CommandOutput> {
  const values = options.compose === true ? await composeEchoValues(options) : options.values;
  const spec = withQueryAuth(
    hasFormFields(options.op)
      ? await buildFormRequest(options.op, values, { fileName: options.uploadFileName })
      : buildRequest(options.op, values),
    options.op,
    options.session,
  );

  const result = await callWithRetry(options, spec);
  return { human: render(result.data, result.message), json: result.text };
}

/** Runs `task` with Ctrl-C turned into an abort, so streams stop cleanly. */
async function withInterrupt<T>(task: (signal: AbortSignal) => Promise<T>): Promise<T> {
  const controller = new AbortController();
  const onSignal = (): void => {
    controller.abort();
    process.exitCode = 130;
  };
  process.once("SIGINT", onSignal);
  try {
    return await task(controller.signal);
  } finally {
    process.off("SIGINT", onSignal);
  }
}

/** Sends a streaming request, retrying once on stale catalog / expired token. */
async function openStream(
  options: InvokeOptions,
  spec: RequestSpec,
  accept: string,
): Promise<{ response: Response; catalog: LoadedCatalog }> {
  let catalog = options.catalog;
  let catalogRetried = false;
  let authRetried = false;

  for (;;) {
    const response = await options.session.client.request({
      ...spec,
      accept,
      headers: { [CATALOG_HEADER]: catalog.fingerprint },
    });
    const status = response.status;
    if (!catalogRetried && (status === 412 || status === 404)) {
      catalogRetried = true;
      catalog = await options.refresh();
      continue;
    }
    if (!authRetried && status === 401 && (await options.session.refresh())) {
      authRetried = true;
      continue;
    }
    return { response, catalog };
  }
}

async function streamOperation(options: InvokeOptions): Promise<CommandOutput> {
  const spec = withQueryAuth(buildRequest(options.op, options.values), options.op, options.session);
  const renderer = createStreamRenderer({
    json: options.json,
    events: (options.op.transport?.response?.events ?? []).map((event) => event.name),
  });

  let failed = false;
  await withInterrupt(async (signal) => {
    const { response } = await openStream(options, spec, "text/event-stream");
    if (!response.ok) {
      const text = await response.text();
      throw toEch0Error(response.status, parseEnvelope(text), text);
    }
    if (response.body === null) {
      throw new Ech0Error("服务端没有返回流式响应");
    }
    try {
      for await (const frame of parseSse(response.body)) {
        renderer.onFrame(frame);
      }
    } catch (error) {
      if (signal.aborted) return;
      throw error;
    }
  });

  renderer.finish();
  failed = renderer.failed;
  if (failed && process.exitCode !== 130) process.exitCode = 1;
  return { human: "", json: "" };
}

/**
 * Names the saved file from `Content-Disposition`: RFC 6266 says `filename*`
 * (RFC 5987 percent-encoding) wins over plain `filename`. There is no useful
 * fallback in the catalog path — its last segment is the verb `stream` — so
 * callers that get no name end up with `download`.
 */
function filenameOf(response: Response): string {
  const disposition = response.headers.get("content-disposition") ?? "";
  const extended = /filename\*\s*=\s*(?:"([^"]*)"|([^;]*))/i.exec(disposition);
  const encoded = (extended?.[1] ?? extended?.[2] ?? "").trim().replace(/^[^']*'[^']*'/, "");
  if (encoded) {
    try {
      return safeFileName(decodeURIComponent(encoded));
    } catch {
      return safeFileName(encoded);
    }
  }
  const plain = /filename\s*=\s*(?:"((?:[^"\\]|\\.)*)"|([^;]*))/i.exec(disposition);
  const name = (plain?.[1] ?? plain?.[2])?.replace(/\\(.)/g, "$1");
  return safeFileName(name);
}

/** Header values are remote input: keep the write inside the chosen directory. */
function safeFileName(raw: string | undefined): string {
  const name = basename((raw ?? "").replace(/\\/g, "/").replace(/[\u0000-\u001f\u007f]/g, "").trim());
  return name === "" || name === "." || name === ".." ? "download" : name;
}

function formatBytes(bytes: number): string {
  const units = ["B", "KiB", "MiB", "GiB"];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit += 1;
  }
  return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

function renderProgress(written: number, total: number): void {
  const percent = total > 0 ? Math.floor((written / total) * 100) : 0;
  process.stderr.write(`\r${formatBytes(written)} / ${formatBytes(total)} (${percent}%)`);
}

async function pipeToStdout(body: ReadableStream<Uint8Array>): Promise<number> {
  let written = 0;
  for await (const chunk of body) {
    written += chunk.byteLength;
    if (!process.stdout.write(chunk)) await once(process.stdout, "drain");
  }
  return written;
}

async function pipeToFile(body: ReadableStream<Uint8Array>, target: string): Promise<number> {
  const total = Number(process.env.ECH0_TRANSFER_TOTAL ?? "");
  const progress = process.stderr.isTTY === true && Number.isFinite(total) && total > 0;
  const temp = `${target}.part`;
  const sink = Bun.file(temp).writer();
  let written = 0;
  try {
    for await (const chunk of body) {
      sink.write(chunk);
      written += chunk.byteLength;
      if (progress) renderProgress(written, total);
    }
    await sink.end();
    await rename(temp, target);
  } catch (error) {
    try {
      await sink.end();
    } catch {
      // The sink is already unusable.
    }
    await unlink(temp).catch(() => undefined);
    throw error;
  }
  if (progress) process.stderr.write("\r\u001b[K");
  return written;
}

async function downloadOperation(options: InvokeOptions): Promise<CommandOutput> {
  const spec = withQueryAuth(buildRequest(options.op, options.values), options.op, options.session);
  return withInterrupt(async (signal) => {
    const { response } = await openStream(options, spec, "*/*");
    if (!response.ok) {
      const text = await response.text();
      throw toEch0Error(response.status, parseEnvelope(text), text);
    }
    if (response.body === null) {
      throw new Ech0Error("服务端没有返回响应体");
    }

    const output = options.output;
    const toStdout = output === "-" || (output === undefined && process.stdout.isTTY !== true);
    if (toStdout && options.json) {
      throw new UsageError("--json 需要 --output <文件>：二进制内容不能与 JSON 输出混在一起");
    }

    if (toStdout) {
      await pipeToStdout(response.body);
      return { human: "", json: "" };
    }

    const target = output ?? join(process.cwd(), filenameOf(response));
    const total = Number(response.headers.get("content-length") ?? "");
    if (Number.isFinite(total) && total > 0) process.env.ECH0_TRANSFER_TOTAL = String(total);
    const written = await pipeToFile(response.body, target);
    delete process.env.ECH0_TRANSFER_TOTAL;

    const summary = { file: target, bytes: written, content_type: response.headers.get("content-type") ?? "" };
    return {
      human: `${colors().green("✓")} 已保存 ${target}（${formatBytes(written)}）`,
      json: JSON.stringify(summary, null, 2),
    };
  });
}

function websocketOperation(options: InvokeOptions): Promise<CommandOutput> {
  const spec = withQueryAuth(buildRequest(options.op, options.values), options.op, options.session);
  const url = options.session.client.websocketUrl(options.op.path, spec.query ?? {});

  return new Promise<CommandOutput>((resolve, reject) => {
    const socket = new WebSocket(url);
    socket.onmessage = (event: MessageEvent) => {
      const raw = typeof event.data === "string" ? event.data : new TextDecoder().decode(event.data as ArrayBuffer);
      if (options.json) {
        print(raw);
        return;
      }
      const payload = parseEnvelope(raw);
      print(JSON.stringify(payload?.data ?? raw));
    };
    socket.onerror = () => reject(new Ech0Error(`WebSocket 连接失败：${url}`));
    socket.onclose = () => resolve({ human: "", json: "" });
  });
}
