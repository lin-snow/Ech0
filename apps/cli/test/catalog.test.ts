// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { Command } from "commander";

import { Client } from "../src/api/http.ts";
import { loadCatalog } from "../src/catalog/load.ts";
import { CATALOG_HEADER, catalogFilePath, fingerprintOf } from "../src/catalog/store.ts";
import { parseCatalog, type Catalog, type CatalogOp } from "../src/catalog/types.ts";
import { registerCatalogCommands } from "../src/commands/dynamic.ts";
import { buildRequest, invokeOperation, type FieldValues } from "../src/commands/operation.ts";
import type { CommandOutput } from "../src/commands/types.ts";
import { Ech0Error, UsageError } from "../src/output/console.ts";
import type { Session } from "../src/session.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

const CATALOG = {
  ops: {
    "echo.create": {
      summary: "创建新的 Echo",
      method: "POST",
      path: "/echo",
      scopes: ["echo:write"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: {
          content: { type: "string" },
          private: { type: "boolean" },
          tags: { type: "array", items: { type: "object", properties: { name: { type: "string" } } } },
        },
        required: ["content"],
      },
      bind: { content: "body", private: "body", tags: "body" },
      deprecated: false,
    },
    "echo.get": {
      summary: "获取指定 ID 的 Echo",
      method: "GET",
      path: "/echo/{id}",
      scopes: ["echo:read"],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string" }, page: { type: "integer" } },
        required: ["id"],
      },
      bind: { id: "path", page: "query" },
      deprecated: false,
    },
  },
};

function catalogBytes(): Uint8Array {
  return new TextEncoder().encode(JSON.stringify(CATALOG));
}

function parsedCatalog(): Catalog {
  return parseCatalog(catalogBytes());
}

function session(server: string, token = "tok"): Session {
  return {
    profile: "test",
    server,
    config: { profiles: { test: { server } } },
    token,
    tokenSource: "flag",
    client: new Client({ server, token }),
    refresh: async () => false,
  };
}

function loaded(server: string, configPath: string) {
  const bytes = catalogBytes();
  return {
    doc: parseCatalog(bytes),
    bytes,
    fingerprint: fingerprintOf(bytes),
    file: catalogFilePath(configPath, server),
    source: "network" as const,
  };
}

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

describe("catalog fetch and cache", () => {
  test("stores the served bytes and reuses the cache", async () => {
    api = startFakeApi({ "GET /api/_catalog": () => ({ body: CATALOG }) });
    const configPath = join(mkdtempSync(join(tmpdir(), "ech0-cli-cat-")), "config.json");

    const first = await loadCatalog({ server: api.baseUrl, configPath });
    expect(first.source).toBe("network");
    expect(first.doc.ops["echo.create"]?.path).toBe("/echo");
    expect(readFileSync(first.file, "utf8")).toContain("echo.create");

    const served = readFileSync(first.file);
    expect(first.fingerprint).toBe(fingerprintOf(served));

    const second = await loadCatalog({ server: api.baseUrl, configPath });
    expect(second.source).toBe("cache");
    expect(second.fingerprint).toBe(first.fingerprint);
    expect(api.calls).toHaveLength(1);
  });

  test("reports an old server without the catalog endpoint", async () => {
    api = startFakeApi({});
    const configPath = join(mkdtempSync(join(tmpdir(), "ech0-cli-old-")), "config.json");
    await expect(loadCatalog({ server: api.baseUrl, configPath })).rejects.toBeInstanceOf(Ech0Error);
  });
});

describe("buildRequest", () => {
  test("places fields by bind: path, query and body", () => {
    const catalog = parsedCatalog();
    const get = catalog.ops["echo.get"] as CatalogOp;
    const spec = buildRequest(get, { id: "abc/def", page: 2, unknown: "ignored" });

    expect(spec.method).toBe("GET");
    expect(spec.path).toBe("/echo/abc%2Fdef");
    expect(spec.query).toEqual({ page: 2 });
    expect(spec.body).toBeUndefined();
  });

  test("builds a JSON body from body-bound fields", () => {
    const catalog = parsedCatalog();
    const create = catalog.ops["echo.create"] as CatalogOp;
    const spec = buildRequest(create, { content: "hi", private: true, tags: ["dev"] });

    expect(spec.method).toBe("POST");
    expect(spec.body).toEqual({ content: "hi", private: true, tags: ["dev"] });
  });

  test("sends an empty body when every body field is optional", () => {
    const list: CatalogOp = {
      summary: "list",
      method: "POST",
      path: "/echo/query",
      scopes: [],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: { page: { type: "integer" }, search: { type: "string" } },
      },
      bind: { page: "body", search: "body" },
      deprecated: false,
    };

    expect(buildRequest(list, {}).body).toEqual({});
    expect(buildRequest(list, { page: 2 }).body).toEqual({ page: 2 });
  });
});

describe("dynamic commands", () => {
  function programWith(run: (command: Command, op: CatalogOp, values: FieldValues) => Promise<CommandOutput>) {
    const program = new Command("ech0").option("--json", "json", false);
    registerCatalogCommands(program, parsedCatalog(), { run });
    const override = (command: Command) => {
      command.exitOverride();
      for (const child of command.commands) override(child);
    };
    override(program);
    return program;
  }

  test("maps an operation to a nested command with typed flags", async () => {
    let captured: FieldValues | undefined;
    const program = programWith(async (_command, _op, values) => {
      captured = values;
      return { human: "", json: "" };
    });

    await program.parseAsync(
      ["echo", "create", "--content", "hi", "--private", "--tags", '[{"name":"dev"}]'],
      { from: "user" },
    );
    expect(captured).toEqual({ content: "hi", private: true, tags: [{ name: "dev" }] });
  });

  test("converts integer flags", async () => {
    let captured: FieldValues | undefined;
    const program = programWith(async (_command, _op, values) => {
      captured = values;
      return { human: "", json: "" };
    });

    await program.parseAsync(["echo", "get", "--id", "abc", "--page", "3"], { from: "user" });
    expect(captured).toEqual({ id: "abc", page: 3 });
  });

  test("only the curated Echo operations are exposed", async () => {
    const program = new Command("ech0");
    const ops = { ...CATALOG.ops } as Record<string, unknown>;
    ops["file.upload"] = {
      summary: "upload",
      method: "POST",
      path: "/files/upload",
      scopes: [],
      effect: "write",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    };
    ops["echo.like"] = {
      summary: "like",
      method: "PUT",
      path: "/echo/like/{id}",
      scopes: [],
      effect: "write",
      kind: "json",
      input: { type: "object", properties: { id: { type: "string" } } },
      bind: { id: "path" },
      deprecated: false,
    };
    registerCatalogCommands(program, { ops: ops as Catalog["ops"] }, {
      run: async () => ({ human: "", json: "" }),
    });

    const echo = program.commands.find((command) => command.name() === "echo");
    const names = (echo?.commands ?? []).map((command) => command.name());
    expect(names.sort()).toEqual(["create", "get"]);
    expect(names).not.toContain("like");

    // Only the curated operations that exist in this catalog are exposed;
    // `echo.like` is a write op outside the allowlist.
    const file = program.commands.find((command) => command.name() === "file");
    expect(file?.commands.map((command) => command.name())).toEqual(["upload"]);
  });

  test("a callable op also gets children without flag stealing", async () => {
    const program = new Command("ech0");
    const withNested: Catalog = {
      ops: {
        "comment.create": {
          summary: "create",
          method: "POST",
          path: "/comments",
          scopes: [],
          effect: "write",
          kind: "json",
          input: { type: "object", properties: {} },
          bind: {},
          deprecated: false,
        },
        "comment.create.integration": {
          summary: "integration",
          method: "POST",
          path: "/comments/integration",
          scopes: [],
          effect: "write",
          kind: "json",
          input: { type: "object", properties: {} },
          bind: {},
          deprecated: false,
        },
      },
    };
    let calls = 0;
    registerCatalogCommands(program, withNested, {
      run: async () => {
        calls += 1;
        return { human: "", json: "" };
      },
      scope: () => true,
    });

    await program.parseAsync(["comment", "create"], { from: "user" });
    await program.parseAsync(["comment", "create-integration"], { from: "user" });
    expect(calls).toBe(2);
  });

  test("missing required input is a usage error", async () => {
    const program = programWith(async () => ({ human: "", json: "" }));
    await expect(program.parseAsync(["echo", "create"], { from: "user" })).rejects.toThrow("缺少必填参数 --content");
  });

  test("missing required input points at the command usage", async () => {
    const program = programWith(async () => ({ human: "", json: "" }));
    const error = await program.parseAsync(["echo", "get"], { from: "user" }).catch((reason: unknown) => reason);

    expect(error).toBeInstanceOf(UsageError);
    expect((error as UsageError).hint).toContain("ech0 echo get --id <id>");
    expect((error as UsageError).hint).toContain("ech0 help echo get");
  });
});

describe("invokeOperation", () => {
  test("retries once with a refreshed catalog on 412", async () => {
    let attempts = 0;
    const fake = startFakeApi({
      "POST /api/echo": () => {
        attempts += 1;
        if (attempts === 1) {
          return {
            status: 412,
            body: { code: 0, msg: "catalog stale", error_code: "catalog_stale" },
          };
        }
        return { body: { code: 1, msg: "ok", data: { id: "e1" } } };
      },
    });
    api = fake;
    const configPath = join(mkdtempSync(join(tmpdir(), "ech0-cli-412-")), "config.json");
    let refreshes = 0;

    const output = await invokeOperation({
      session: session(fake.baseUrl),
      op: parsedCatalog().ops["echo.create"] as CatalogOp,
      values: { content: "hi" },
      catalog: loaded(fake.baseUrl, configPath),
      refresh: async () => {
        refreshes += 1;
        return loaded(fake.baseUrl, configPath);
      },
      json: false,
    });

    expect(attempts).toBe(2);
    expect(refreshes).toBe(1);
    expect(output.human).toContain('"id": "e1"');
    expect(fake.calls[0]?.headers.get(CATALOG_HEADER)).toBe(loaded(fake.baseUrl, configPath).fingerprint);
  });

  test("does not retry ordinary failures", async () => {
    const fake = startFakeApi({
      "POST /api/echo": () => ({ status: 403, body: { code: 0, msg: "forbidden" } }),
    });
    api = fake;
    const configPath = join(mkdtempSync(join(tmpdir(), "ech0-cli-403-")), "config.json");
    let refreshes = 0;

    await expect(
      invokeOperation({
        session: session(fake.baseUrl),
        op: parsedCatalog().ops["echo.create"] as CatalogOp,
        values: { content: "hi" },
        catalog: loaded(fake.baseUrl, configPath),
        refresh: async () => {
          refreshes += 1;
          return loaded(fake.baseUrl, configPath);
        },
        json: false,
      }),
    ).rejects.toThrow("forbidden");

    expect(refreshes).toBe(0);
  });
});
