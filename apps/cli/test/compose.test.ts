// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { Client } from "../src/api/http.ts";
import type { LoadedCatalog } from "../src/catalog/load.ts";
import type { CatalogOp } from "../src/catalog/types.ts";
import { invokeOperation } from "../src/commands/operation.ts";
import { parsePayload } from "../src/commands/payload.ts";
import type { Session } from "../src/session.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

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

const CREATE: CatalogOp = baseOp({
  path: "/echo",
  input: {
    type: "object",
    properties: {
      content: { type: "string" },
      tags: { type: "array" },
      echo_files: { type: "array" },
    },
    required: ["content"],
  },
  bind: { content: "body", tags: "body", echo_files: "body" },
});

const UPLOAD: CatalogOp = baseOp({
  path: "/files/upload",
  input: {
    type: "object",
    properties: {
      file: { type: "string", format: "binary" },
      category: { type: "string" },
      storage_type: { type: "string" },
    },
  },
  bind: { file: "form", category: "form", storage_type: "form" },
});

function sessionFor(baseUrl: string): Session {
  return {
    profile: "test",
    server: baseUrl,
    config: { profiles: {} },
    token: "tok",
    tokenSource: "flag",
    client: new Client({ server: baseUrl, token: "tok" }),
    refresh: async () => false,
  };
}

function catalogWith(ops: Record<string, CatalogOp>): LoadedCatalog {
  return {
    doc: { ops },
    bytes: new Uint8Array(),
    fingerprint: "fp",
    file: "/tmp/catalog.json",
    source: "cache",
  };
}

function tempFile(name: string, bytes: string): string {
  const dir = mkdtempSync(join(tmpdir(), "ech0-compose-"));
  const path = join(dir, name);
  writeFileSync(path, bytes);
  return path;
}

interface Recorded {
  uploads: { name: string; category: string; storage: string }[];
  body?: Record<string, unknown>;
}

let server: ReturnType<typeof Bun.serve> | undefined;

afterEach(() => {
  server?.stop(true);
  server = undefined;
});

function startServer(record: Recorded): string {
  server = Bun.serve({
    port: 0,
    async fetch(request) {
      const url = new URL(request.url);
      if (url.pathname === "/api/files/upload") {
        const form = await request.formData();
        const file = form.get("file") as File;
        record.uploads.push({
          name: file.name,
          category: String(form.get("category")),
          storage: String(form.get("storage_type")),
        });
        return Response.json({ code: 1, msg: "", data: { id: `f${record.uploads.length}` } });
      }
      if (url.pathname === "/api/echo") {
        record.body = (await request.json()) as Record<string, unknown>;
        return Response.json({ code: 1, msg: "发布Echo成功！", data: null });
      }
      return new Response("not found", { status: 404 });
    },
  });
  return `http://127.0.0.1:${server.port}`;
}

describe("composed Echo payload", () => {
  test("merges payload file, flags, tags and uploaded attachments", async () => {
    const record: Recorded = { uploads: [] };
    const baseUrl = startServer(record);
    const png = tempFile("cover.png", "png-bytes");

    const result = await invokeOperation({
      session: sessionFor(baseUrl),
      op: CREATE,
      values: { content: "flag content", tags: ["cli", "demo"] },
      payload: {
        content: "payload content",
        echo_files: [{ file_id: "old", sort_order: 3 }],
        tags: [{ name: "from-file" }],
      },
      compose: true,
      tags: ["extra"],
      attach: [png],
      catalog: catalogWith({ "echo.create": CREATE, "file.upload": UPLOAD }),
      refresh: async () => catalogWith({}),
      json: false,
    });

    expect(record.uploads).toEqual([{ name: "cover.png", category: "image", storage: "local" }]);
    expect(record.body).toMatchObject({
      content: "flag content",
      tags: [{ name: "from-file" }, { name: "cli" }, { name: "demo" }, { name: "extra" }],
      echo_files: [
        { file_id: "old", sort_order: 3 },
        { file_id: "f1", sort_order: 4 },
      ],
    });
    expect(result.human).toContain("发布Echo成功");
  });

  test("normalizes bare tag strings from --tags", async () => {
    const record: Recorded = { uploads: [] };
    const baseUrl = startServer(record);

    await invokeOperation({
      session: sessionFor(baseUrl),
      op: CREATE,
      values: { content: "x", tags: ["a", "b"] },
      compose: true,
      catalog: catalogWith({ "echo.create": CREATE }),
      refresh: async () => catalogWith({}),
      json: false,
    });

    expect(record.body?.tags).toEqual([{ name: "a" }, { name: "b" }]);
  });

  test("rejects storage types that cannot upload", async () => {
    const record: Recorded = { uploads: [] };
    const baseUrl = startServer(record);
    const png = tempFile("a.png", "x");

    await expect(
      invokeOperation({
        session: sessionFor(baseUrl),
        op: CREATE,
        values: { content: "x" },
        compose: true,
        attach: [png],
        attachStorage: "external",
        catalog: catalogWith({ "echo.create": CREATE, "file.upload": UPLOAD }),
        refresh: async () => catalogWith({}),
        json: false,
      }),
    ).rejects.toThrow("external");
    expect(record.uploads).toHaveLength(0);
  });

  test("refuses mixed attachment categories before uploading anything", async () => {
    const record: Recorded = { uploads: [] };
    const baseUrl = startServer(record);
    const png = tempFile("a.png", "x");
    const mp3 = tempFile("b.mp3", "y");

    await expect(
      invokeOperation({
        session: sessionFor(baseUrl),
        op: CREATE,
        values: { content: "x" },
        compose: true,
        attach: [png, mp3],
        catalog: catalogWith({ "echo.create": CREATE, "file.upload": UPLOAD }),
        refresh: async () => catalogWith({}),
        json: false,
      }),
    ).rejects.toThrow("同一类别");
    expect(record.uploads).toHaveLength(0);
  });
});

describe("parsePayload", () => {
  test("requires a JSON object", () => {
    expect(parsePayload('{"content":"x"}', "p.json")).toEqual({ content: "x" });
    expect(() => parsePayload("[]", "p.json")).toThrow("JSON 对象");
    expect(() => parsePayload("nope", "p.json")).toThrow("不是合法 JSON");
  });
});
