// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import type { LoadedCatalog } from "../src/catalog/load.ts";
import { parseCatalog } from "../src/catalog/types.ts";
import { invokeOperation } from "../src/commands/operation.ts";
import { yamlFromJson, yamlOf } from "../src/output/yaml.ts";
import type { Session } from "../src/session.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

const ENTRY = join(import.meta.dir, "..", "src", "index.ts");

const CATALOG = {
  ops: {
    "echo.list": {
      summary: "分页/筛选 Echo 列表",
      method: "POST",
      path: "/echo/query",
      scopes: ["echo:read"],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: { page: { type: "integer", description: "页码" } },
      },
      bind: { page: "body" },
      deprecated: false,
    },
  },
};

const STREAM_CATALOG = {
  ops: {
    "chat.ask": {
      summary: "提问",
      method: "POST",
      path: "/chat",
      scopes: [],
      effect: "write",
      kind: "stream",
      input: { type: "object", properties: { question: { type: "string" } }, required: ["question"] },
      bind: { question: "body" },
      transport: { response: { kind: "stream", content_type: "text/event-stream" } },
      deprecated: false,
    },
  },
};

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

async function run(
  args: string[],
  env: NodeJS.ProcessEnv = {},
): Promise<{ stdout: string; stderr: string; exitCode: number }> {
  const base = { ...process.env };
  for (const key of ["ECH0_SERVER", "ECH0_TOKEN", "ECH0_PROFILE", "ECH0_CONFIG", "ECH0_LANG"]) {
    delete base[key];
  }
  const child = Bun.spawn({
    cmd: [process.execPath, "run", ENTRY, ...args],
    env: {
      ...base,
      ECH0_NO_KEYCHAIN: "1",
      ECH0_CONFIG: join(mkdtempSync(join(tmpdir(), "ech0-cli-yaml-")), "config.json"),
      ...env,
    },
    stdout: "pipe",
    stderr: "pipe",
  });
  const [stdout, stderr, exitCode] = await Promise.all([
    new Response(child.stdout).text(),
    new Response(child.stderr).text(),
    child.exited,
  ]);
  return { stdout, stderr, exitCode };
}

describe("yaml output", () => {
  test("a rendered document parses back to the same value", () => {
    const value = {
      code: 1,
      msg: "ok",
      data: {
        items: [{ id: "e1", content: "hi", tags: ["a", "y"] }],
        empty: [],
        none: null,
        text: "line1\nline2",
        ratio: 0.5,
      },
    };
    expect(Bun.YAML.parse(yamlOf(value))).toEqual(value);
  });

  test("a body that is not JSON passes through unchanged", () => {
    expect(yamlFromJson("upstream said no")).toBe("upstream said no\n");
  });

  test("--yaml renders the same document --json serves", async () => {
    api = startFakeApi({
      "GET /api/_catalog": () => ({ body: CATALOG }),
      "POST /api/echo/query": () => ({
        body: { code: 1, msg: "ok", data: { items: [{ id: "e1" }], total: 1 } },
      }),
    });
    const env = {
      ECH0_SERVER: api.baseUrl,
      ECH0_TOKEN: "tok",
      ECH0_PROFILE: "test",
    };
    const json = await run(["echo", "list", "--json"], env);
    const yaml = await run(["echo", "list", "--yaml"], env);

    expect(Bun.YAML.parse(yaml.stdout)).toEqual(JSON.parse(json.stdout) as unknown);
  });

  test("help --yaml emits the descriptor document", async () => {
    api = startFakeApi({ "GET /api/_catalog": () => ({ body: CATALOG }) });
    const result = await run(["help", "echo", "list", "--yaml"], { ECH0_SERVER: api.baseUrl });
    const parsed = Bun.YAML.parse(result.stdout) as { path: string[]; hints: string[] };

    expect(parsed.path).toEqual(["echo", "list"]);
    expect(parsed.hints[0]).toContain("echo:read");
  });

  test("--json and --yaml are mutually exclusive", async () => {
    const result = await run(["version", "--json", "--yaml"]);

    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain("--json 与 --yaml 只能选一个");
  });

  test("streams and bytes have no YAML document", async () => {
    const bytes = new TextEncoder().encode(JSON.stringify(STREAM_CATALOG));
    const doc = parseCatalog(bytes);
    const catalog: LoadedCatalog = {
      doc,
      bytes,
      fingerprint: "test",
      file: "",
      source: "network",
    };
    const session: Session = {
      profile: "test",
      server: "http://127.0.0.1:1",
      config: { profiles: {} },
      token: "tok",
      tokenSource: "env",
      client: {} as Session["client"],
      refresh: async () => false,
    };

    await expect(
      invokeOperation({
        session,
        op: doc.ops["chat.ask"]!,
        values: { question: "hi" },
        catalog,
        refresh: async () => catalog,
        json: false,
        yaml: true,
      }),
    ).rejects.toThrow("--yaml 只适用于 JSON 结果");
  });
});
