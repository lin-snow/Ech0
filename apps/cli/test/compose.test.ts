// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { Command } from "commander";

import { Client } from "../src/api/http.ts";
import type { LoadedCatalog } from "../src/catalog/load.ts";
import type { CatalogOp } from "../src/catalog/types.ts";
import { registerCatalogCommands } from "../src/commands/dynamic.ts";
import { categoryForFile } from "../src/commands/files.ts";
import { buildFormRequest, invokeOperation } from "../src/commands/operation.ts";
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
    required: ["file"],
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

let server: ReturnType<typeof Bun.serve> | undefined;
let body: Record<string, unknown> | undefined;

afterEach(() => {
  server?.stop(true);
  server = undefined;
  body = undefined;
});

function startServer(): string {
  server = Bun.serve({
    port: 0,
    async fetch(request) {
      const url = new URL(request.url);
      if (url.pathname === "/api/echo") {
        body = (await request.json()) as Record<string, unknown>;
        return Response.json({ code: 1, msg: "发布Echo成功！", data: null });
      }
      return new Response("not found", { status: 404 });
    },
  });
  return `http://127.0.0.1:${server.port}`;
}

describe("composed Echo payload", () => {
  test("merges payload file, flags, tags and file ids", async () => {
    const baseUrl = startServer();

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
      fileIds: ["f1", "f2"],
      catalog: catalogWith({ "echo.create": CREATE }),
      refresh: async () => catalogWith({}),
      json: false,
    });

    expect(body).toMatchObject({
      content: "flag content",
      tags: [{ name: "from-file" }, { name: "cli" }, { name: "demo" }, { name: "extra" }],
      echo_files: [
        { file_id: "old", sort_order: 3 },
        { file_id: "f1", sort_order: 4 },
        { file_id: "f2", sort_order: 5 },
      ],
    });
    expect(result.human).toContain("发布Echo成功");
  });

  test("normalizes bare tag strings from --tags", async () => {
    await invokeOperation({
      session: sessionFor(startServer()),
      op: CREATE,
      values: { content: "x", tags: ["a", "b"] },
      compose: true,
      catalog: catalogWith({ "echo.create": CREATE }),
      refresh: async () => catalogWith({}),
      json: false,
    });

    expect(body?.tags).toEqual([{ name: "a" }, { name: "b" }]);
  });
});

describe("file external command", () => {
  test("registers a URL as an external file", async () => {
    const program = new Command("ech0");
    let captured: Record<string, unknown> | undefined;
    const external: CatalogOp = baseOp({
      path: "/files/external",
      input: {
        type: "object",
        properties: { url: { type: "string" }, category: { type: "string" } },
      },
      bind: { url: "body", category: "body" },
    });
    registerCatalogCommands(
      program,
      { ops: { "file.external": external } },
      {
        run: async (_command, _op, values) => {
          captured = values;
          return { human: "", json: "" };
        },
      },
    );

    await program.parseAsync(
      ["file", "external", "--url", "https://example.com/a.png", "--category", "image"],
      { from: "user" },
    );
    expect(captured).toEqual({ url: "https://example.com/a.png", category: "image" });
  });
});

describe("file upload command", () => {
  test("rejects storage types that cannot upload", async () => {
    const program = new Command("ech0");
    registerCatalogCommands(
      program,
      { ops: { "file.upload": UPLOAD } },
      {
        run: async () => ({ human: "", json: "" }),
        scope: () => true,
      },
    );

    await expect(
      program.parseAsync(
        ["file", "upload", "--file", "/tmp/a.png", "--storage-type", "external"],
        { from: "user" },
      ),
    ).rejects.toThrow("storage-type");
  });

  test("infers the category from the filename", async () => {
    const program = new Command("ech0");
    let captured: Record<string, unknown> | undefined;
    registerCatalogCommands(
      program,
      { ops: { "file.upload": UPLOAD } },
      {
        run: async (_command, _op, values) => {
          captured = values;
          return { human: "", json: "" };
        },
        scope: () => true,
      },
    );

    await program.parseAsync(["file", "upload", "--file", "/tmp/a.png"], { from: "user" });

    expect(captured).toEqual({ file: "/tmp/a.png", category: "image" });
  });

  test("--file-name overrides the multipart filename", async () => {
    const dir = mkdtempSync(join(tmpdir(), "ech0-upload-"));
    const path = join(dir, "blob");
    writeFileSync(path, "x");

    const spec = await buildFormRequest(UPLOAD, { file: path }, { fileName: "shot.png" });
    expect(spec.form?.[0]?.filename).toBe("shot.png");
  });

  test("forwards --file-name for stdin uploads", async () => {
    const program = new Command("ech0");
    let extra: { uploadFileName?: string } | undefined;
    registerCatalogCommands(
      program,
      { ops: { "file.upload": UPLOAD } },
      {
        run: async (_command, _op, _values, capturedExtra) => {
          extra = capturedExtra;
          return { human: "", json: "" };
        },
        scope: () => true,
      },
    );

    await program.parseAsync(
      ["file", "upload", "--file", "/tmp/a.png", "--file-name", "shot.png"],
      { from: "user" },
    );
    expect(extra?.uploadFileName).toBe("shot.png");
  });

  test("maps extensions to server categories", () => {
    expect(categoryForFile("a.png")).toBe("image");
    expect(categoryForFile("clip.MOV")).toBe("video");
    expect(categoryForFile("x.m4a")).toBe("audio");
    expect(categoryForFile("doc.pdf")).toBe("pdf");
    expect(categoryForFile("readme.md")).toBe("markdown");
    expect(categoryForFile("archive.zip")).toBe("file");
  });
});

describe("parsePayload", () => {
  test("requires a JSON object", () => {
    expect(parsePayload('{"content":"x"}', "p.json")).toEqual({ content: "x" });
    expect(() => parsePayload("[]", "p.json")).toThrow("JSON 对象");
    expect(() => parsePayload("nope", "p.json")).toThrow("不是合法 JSON");
  });
});
