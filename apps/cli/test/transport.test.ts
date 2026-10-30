// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { Client } from "../src/api/http.ts";
import { parseSse } from "../src/api/sse.ts";
import type { LoadedCatalog } from "../src/catalog/load.ts";
import type { CatalogOp } from "../src/catalog/types.ts";
import { buildFormRequest, invokeOperation } from "../src/commands/operation.ts";
import type { Session } from "../src/session.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

function streamOf(chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder();
  return new ReadableStream({
    start(controller) {
      for (const chunk of chunks) controller.enqueue(encoder.encode(chunk));
      controller.close();
    },
  });
}

function baseOp(overrides: Partial<CatalogOp>): CatalogOp {
  return {
    summary: "test",
    method: "POST",
    path: "/test",
    scopes: [],
    effect: "write",
    kind: "json",
    input: { type: "object", properties: {} },
    bind: {},
    deprecated: false,
    ...overrides,
  };
}

function sessionFor(baseUrl: string, token = "tok"): Session {
  return {
    profile: "test",
    server: baseUrl,
    config: { profiles: {} },
    token,
    tokenSource: "flag",
    client: new Client({ server: baseUrl, token }),
    refresh: async () => false,
  };
}

function catalogStub(fingerprint = "fp"): LoadedCatalog {
  return {
    doc: { ops: {} },
    bytes: new Uint8Array(),
    fingerprint,
    file: "/tmp/catalog.json",
    source: "cache",
  };
}

async function captureStdout<T>(task: () => Promise<T>): Promise<{ value: T; output: string }> {
  const original = process.stdout.write.bind(process.stdout);
  let output = "";
  process.stdout.write = ((chunk: unknown) => {
    output += typeof chunk === "string" ? chunk : new TextDecoder().decode(chunk as Uint8Array);
    return true;
  }) as typeof process.stdout.write;
  try {
    const value = await task();
    return { value, output };
  } finally {
    process.stdout.write = original;
  }
}

describe("parseSse", () => {
  test("reassembles frames split across chunks", async () => {
    const frames = [];
    for await (const frame of parseSse(streamOf(['event: del', "ta\ndata: {\"text\":\"h", 'i"}\n\n']))) {
      frames.push(frame);
    }
    expect(frames).toEqual([{ event: "delta", data: '{"text":"hi"}' }]);
  });

  test("joins multi-line data, tolerates CRLF, comments and unknown fields", async () => {
    const frames = [];
    for await (const frame of parseSse(streamOf([": keep-alive\r\n\r\nevent: x\r\ndata: a\r\ndata: b\r\n\r\n"]))) {
      frames.push(frame);
    }
    expect(frames).toEqual([{ event: "x", data: "a\nb" }]);
  });

  test("defaults to the message event and skips frames without data", async () => {
    const frames = [];
    for await (const frame of parseSse(streamOf(["event: ping\n\n", "data: plain\n\n"]))) {
      frames.push(frame);
    }
    expect(frames).toEqual([{ event: "message", data: "plain" }]);
  });
});

describe("multipart upload", () => {
  test("builds parts for text fields and files", async () => {
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-"));
    const path = join(dir, "note.txt");
    writeFileSync(path, "hello");

    const op = baseOp({
      path: "/files/upload",
      input: {
        type: "object",
        properties: {
          file: { type: "string", format: "binary" },
          category: { type: "string" },
        },
      },
      bind: { file: "form", category: "form" },
      transport: { request: { content_type: "multipart/form-data" } },
    });

    const spec = await buildFormRequest(op, { file: path, category: "images" });
    expect(spec.form?.map((part) => part.name)).toEqual(["file", "category"]);
    expect(spec.form?.[0]?.filename).toBe("note.txt");
    expect(spec.form?.[1]).toEqual({ name: "category", value: "images" });
  });

  test("reports a missing file as a usage error", async () => {
    const op = baseOp({
      bind: { file: "form" },
      input: { type: "object", properties: { file: { type: "string", format: "binary" } } },
    });
    await expect(buildFormRequest(op, { file: "/nonexistent/file.bin" })).rejects.toThrow("文件不存在");
  });

  test("posts the file and returns the JSON envelope", async () => {
    let received: { name: string; size: number; category: string | null } | undefined;
    const server = Bun.serve({
      port: 0,
      async fetch(request) {
        const form = await request.formData();
        const file = form.get("file") as File;
        received = { name: file.name, size: file.size, category: form.get("category") as string | null };
        return Response.json({ code: 1, msg: "上传成功", data: { id: "f1" } });
      },
    });
    try {
      const dir = mkdtempSync(join(tmpdir(), "ech0-cli-"));
      const path = join(dir, "tiny.txt");
      writeFileSync(path, "abc");
      const op = baseOp({
        path: "/files/upload",
        input: {
          type: "object",
          properties: { file: { type: "string", format: "binary" }, category: { type: "string" } },
        },
        bind: { file: "form", category: "form" },
      });

      const result = await invokeOperation({
        session: sessionFor(`http://127.0.0.1:${server.port}`),
        op,
        values: { file: path, category: "doc" },
        catalog: catalogStub(),
        refresh: async () => catalogStub(),
        json: false,
      });

      expect(received).toEqual({ name: "tiny.txt", size: 3, category: "doc" });
      expect(result.human).toContain('"id": "f1"');
    } finally {
      server.stop(true);
    }
  });
});

describe("streaming", () => {
  const op = baseOp({
    path: "/chat",
    bind: { question: "body" },
    kind: "stream",
    transport: {
      request: { content_type: "application/json" },
      response: { kind: "stream", content_type: "text/event-stream", events: [{ name: "delta" }, { name: "done" }] },
    },
  });

  function sseServer(): ReturnType<typeof Bun.serve> {
    return Bun.serve({
      port: 0,
      fetch() {
        const body = [
          'event: delta\ndata: {"text":"Hel"}\n\n',
          'event: delta\ndata: {"text":"lo"}\n\n',
          'event: done\ndata: {"done":true}\n\n',
        ].join("");
        return new Response(body, { headers: { "Content-Type": "text/event-stream" } });
      },
    });
  }

  test("renders deltas and stops at done", async () => {
    const server = sseServer();
    try {
      const { output, value } = await captureStdout(() =>
        invokeOperation({
          session: sessionFor(`http://127.0.0.1:${server.port}`),
          op,
          values: { question: "hi" },
          catalog: catalogStub(),
          refresh: async () => catalogStub(),
          json: false,
        }),
      );
      expect(output).toBe("Hello\n");
      expect(value.human).toBe("");
    } finally {
      server.stop(true);
    }
  });

  test("--json emits one NDJSON line per frame", async () => {
    const server = sseServer();
    try {
      const { output } = await captureStdout(() =>
        invokeOperation({
          session: sessionFor(`http://127.0.0.1:${server.port}`),
          op,
          values: { question: "hi" },
          catalog: catalogStub(),
          refresh: async () => catalogStub(),
          json: true,
        }),
      );
      const lines = output.trim().split("\n").map((line) => JSON.parse(line));
      expect(lines).toEqual([
        { event: "delta", data: { text: "Hel" } },
        { event: "delta", data: { text: "lo" } },
        { event: "done", data: { done: true } },
      ]);
    } finally {
      server.stop(true);
    }
  });

  test("refreshes a stale catalog and retries once", async () => {
    let calls = 0;
    const server = Bun.serve({
      port: 0,
      fetch() {
        calls += 1;
        if (calls === 1) {
          return Response.json({ code: 0, msg: "catalog stale", error_code: "catalog_stale" }, { status: 412 });
        }
        return new Response('event: done\ndata: {"done":true}\n\n', {
          headers: { "Content-Type": "text/event-stream" },
        });
      },
    });
    try {
      let refreshed = 0;
      await captureStdout(() =>
        invokeOperation({
          session: sessionFor(`http://127.0.0.1:${server.port}`),
          op,
          values: { question: "hi" },
          catalog: catalogStub("old"),
          refresh: async () => {
            refreshed += 1;
            return catalogStub("new");
          },
          json: true,
        }),
      );
      expect(refreshed).toBe(1);
      expect(calls).toBe(2);
    } finally {
      server.stop(true);
    }
  });

  test("surfaces an error event through the exit code", async () => {
    const server = Bun.serve({
      port: 0,
      fetch() {
        return new Response('event: error\ndata: {"message":"boom"}\n\n', {
          headers: { "Content-Type": "text/event-stream" },
        });
      },
    });
    const previous = process.exitCode;
    try {
      await captureStdout(() =>
        invokeOperation({
          session: sessionFor(`http://127.0.0.1:${server.port}`),
          op,
          values: { question: "hi" },
          catalog: catalogStub(),
          refresh: async () => catalogStub(),
          json: true,
        }),
      );
      expect(process.exitCode).toBe(1);
    } finally {
      process.exitCode = previous ?? 0;
      server.stop(true);
    }
  });
});

describe("binary download", () => {
  test("writes the body to --output and reports size", async () => {
    const payload = new TextEncoder().encode("binary-bytes");
    const server = Bun.serve({
      port: 0,
      fetch() {
        return new Response(payload, {
          headers: { "Content-Type": "application/zip", "Content-Disposition": 'attachment; filename="snap.zip"' },
        });
      },
    });
    try {
      const dir = mkdtempSync(join(tmpdir(), "ech0-cli-"));
      const target = join(dir, "out.zip");
      const op = baseOp({
        method: "GET",
        path: "/migration/export/download",
        kind: "binary",
        transport: { response: { kind: "binary" } },
      });

      const result = await invokeOperation({
        session: sessionFor(`http://127.0.0.1:${server.port}`),
        op,
        values: {},
        catalog: catalogStub(),
        refresh: async () => catalogStub(),
        json: false,
        output: target,
      });

      expect(readFileSync(target).toString()).toBe("binary-bytes");
      expect(result.human).toContain(target);
      const meta = JSON.parse(result.json) as { file: string; bytes: number; content_type: string };
      expect(meta).toEqual({ file: target, bytes: payload.byteLength, content_type: "application/zip" });
    } finally {
      server.stop(true);
    }
  });

  test("names the saved file from Content-Disposition when --output is omitted", async () => {
    const payload = new TextEncoder().encode("zip-bytes");
    const server = Bun.serve({
      port: 0,
      fetch() {
        return new Response(payload, {
          headers: {
            "Content-Type": "application/zip",
            "Content-Disposition": "inline; filename*=utf-8''%E6%8A%A5%E5%91%8A.zip",
          },
        });
      },
    });
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-name-"));
    const previousCwd = process.cwd();
    const previousTty = process.stdout.isTTY;
    const op = baseOp({
      method: "GET",
      path: "/file/{id}/stream",
      kind: "binary",
      transport: { response: { kind: "binary" } },
    });
    try {
      Object.defineProperty(process.stdout, "isTTY", { value: true, configurable: true });
      process.chdir(dir);
      const result = await invokeOperation({
        session: sessionFor(`http://127.0.0.1:${server.port}`),
        op,
        values: { id: "f1" },
        catalog: catalogStub(),
        refresh: async () => catalogStub(),
        json: false,
      });

      expect(readFileSync(join(dir, "报告.zip")).toString()).toBe("zip-bytes");
      expect(result.human).toContain(join(dir, "报告.zip"));
    } finally {
      process.chdir(previousCwd);
      Object.defineProperty(process.stdout, "isTTY", { value: previousTty, configurable: true });
      server.stop(true);
    }
  });

  test("strips directories from a server-supplied filename", async () => {
    const server = Bun.serve({
      port: 0,
      fetch: () => new Response("bytes", { headers: { "Content-Disposition": 'attachment; filename="../../evil.zip"' } }),
    });
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-name-"));
    const previousCwd = process.cwd();
    const previousTty = process.stdout.isTTY;
    const op = baseOp({
      method: "GET",
      path: "/file/{id}/stream",
      kind: "binary",
      transport: { response: { kind: "binary" } },
    });
    try {
      Object.defineProperty(process.stdout, "isTTY", { value: true, configurable: true });
      process.chdir(dir);
      await invokeOperation({
        session: sessionFor(`http://127.0.0.1:${server.port}`),
        op,
        values: { id: "f1" },
        catalog: catalogStub(),
        refresh: async () => catalogStub(),
        json: false,
      });

      expect(readFileSync(join(dir, "evil.zip")).toString()).toBe("bytes");
    } finally {
      process.chdir(previousCwd);
      Object.defineProperty(process.stdout, "isTTY", { value: previousTty, configurable: true });
      server.stop(true);
    }
  });

  test("falls back to download when the response carries no filename", async () => {
    const server = Bun.serve({ port: 0, fetch: () => new Response("bytes") });
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-name-"));
    const previousCwd = process.cwd();
    const previousTty = process.stdout.isTTY;
    const op = baseOp({
      method: "GET",
      path: "/file/{id}/stream",
      kind: "binary",
      transport: { response: { kind: "binary" } },
    });
    try {
      Object.defineProperty(process.stdout, "isTTY", { value: true, configurable: true });
      process.chdir(dir);
      await invokeOperation({
        session: sessionFor(`http://127.0.0.1:${server.port}`),
        op,
        values: { id: "f1" },
        catalog: catalogStub(),
        refresh: async () => catalogStub(),
        json: false,
      });

      expect(readFileSync(join(dir, "download")).toString()).toBe("bytes");
    } finally {
      process.chdir(previousCwd);
      Object.defineProperty(process.stdout, "isTTY", { value: previousTty, configurable: true });
      server.stop(true);
    }
  });

  test("rejects --json when the bytes would go to stdout", async () => {
    const server = Bun.serve({ port: 0, fetch: () => new Response("x") });
    const op = baseOp({
      method: "GET",
      path: "/file/stream",
      kind: "binary",
      transport: { response: { kind: "binary" } },
    });
    const previous = process.stdout.isTTY;
    try {
      Object.defineProperty(process.stdout, "isTTY", { value: false, configurable: true });
      await expect(
        invokeOperation({
          session: sessionFor(`http://127.0.0.1:${server.port}`),
          op,
          values: {},
          catalog: catalogStub(),
          refresh: async () => catalogStub(),
          json: true,
        }),
      ).rejects.toThrow("--json 需要 --output");
    } finally {
      Object.defineProperty(process.stdout, "isTTY", { value: previous, configurable: true });
      server.stop(true);
    }
  });
});

describe("websocket", () => {
  test("appends the token as a query parameter and prints frames", async () => {
    let seenToken: string | null = null;
    const server = Bun.serve({
      port: 0,
      fetch(request, socketServer) {
        seenToken = new URL(request.url).searchParams.get("token");
        if (socketServer.upgrade(request)) return undefined;
        return new Response("upgrade failed", { status: 400 });
      },
      websocket: {
        open(socket) {
          socket.send(JSON.stringify({ code: 1, msg: "log", data: { level: "info" } }));
          socket.close();
        },
        message() {},
      },
    });
    try {
      const op = baseOp({
        method: "GET",
        path: "/system/logs",
        mount: "ws",
        kind: "websocket",
        bind: { level: "query" },
        input: { type: "object", properties: { level: { type: "string" } } },
        transport: { request: { auth: "query" }, response: { kind: "websocket" } },
      });

      const { output } = await captureStdout(() =>
        invokeOperation({
          session: sessionFor(`http://127.0.0.1:${server.port}`, "secret"),
          op,
          values: { level: "info" },
          catalog: catalogStub(),
          refresh: async () => catalogStub(),
          json: true,
        }),
      );

      expect(seenToken as string | null).toBe("secret");
      expect(output.trim()).toBe(JSON.stringify({ code: 1, msg: "log", data: { level: "info" } }));
    } finally {
      server.stop(true);
    }
  });
});
