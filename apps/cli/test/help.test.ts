// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { Command } from "commander";

import { parseCatalog } from "../src/catalog/types.ts";
import { registerCatalogCommands } from "../src/commands/dynamic.ts";
import { helpModel, nodeAt } from "../src/help/model.ts";
import { renderHelp } from "../src/help/render.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

const ENTRY = join(import.meta.dir, "..", "src", "index.ts");

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
          content: {
            type: "string",
            description:
              "Echo 的正文内容，支持 Markdown；服务端会按配置的可见性规则裁剪，私密 Echo 只对 owner 可见",
          },
          private: { type: "boolean", description: "是否私密" },
          tags: {
            type: ["array", "null"],
            description: "标签列表，元素为 {name}",
            items: { type: "object", additionalProperties: true, properties: { name: { type: "string" } } },
          },
          echo_files: {
            type: ["array", "null"],
            description: "附件引用列表：{file_id, sort_order}",
            items: {
              type: "object",
              additionalProperties: true,
              properties: { file_id: { type: "string" }, sort_order: { type: "integer" } },
            },
          },
          extension: {
            type: "object",
            description: "扩展卡片 {type, payload}",
            additionalProperties: true,
            properties: { type: { type: "string" } },
          },
        },
        required: ["content"],
      },
      bind: { content: "body", private: "body", tags: "body", echo_files: "body", extension: "body" },
      deprecated: false,
    },
    "echo.list": {
      summary: "分页/筛选 Echo 列表",
      method: "POST",
      path: "/echo/query",
      scopes: ["echo:read"],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: {
          page: { type: "integer", description: "页码", default: 1 },
          search: { type: "string", description: "关键词" },
        },
      },
      bind: { page: "body", search: "body" },
      deprecated: false,
    },
    "file.upload": {
      summary: "上传文件（multipart）",
      method: "POST",
      path: "/files/upload",
      scopes: ["file:write"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: {
          file: { type: "string", format: "binary", description: "文件" },
          category: { type: "string", description: "分类", enum: ["image", "audio"] },
        },
        required: ["file", "category"],
      },
      bind: { file: "form", category: "form" },
      transport: { request: { content_type: "multipart/form-data" } },
      deprecated: false,
    },
    "file.download": {
      summary: "按 ID 下载文件",
      method: "GET",
      path: "/file/{id}/stream",
      scopes: ["file:read"],
      effect: "read",
      kind: "binary",
      input: { type: "object", properties: { id: { type: "string", description: "文件 ID" } }, required: ["id"] },
      bind: { id: "path" },
      transport: { response: { kind: "binary" } },
      deprecated: false,
    },
    "chat.ask": {
      summary: "向 Copilot 提问（SSE 流式回答）",
      method: "POST",
      path: "/chat",
      scopes: ["admin:settings"],
      effect: "write",
      kind: "stream",
      input: {
        type: "object",
        properties: { question: { type: "string", description: "问题" } },
        required: ["question"],
      },
      bind: { question: "body" },
      transport: {
        response: {
          kind: "stream",
          content_type: "text/event-stream",
          events: [{ name: "searching" }, { name: "delta" }],
        },
      },
      deprecated: false,
    },
  },
};

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

/** Async spawn: the fake API answers on this event loop, so it cannot be blocked. */
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
      ECH0_CONFIG: join(mkdtempSync(join(tmpdir(), "ech0-cli-help-")), "config.json"),
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

/** The descriptor of one command, rendered from an unfiltered catalog tree. */
function helpText(id: string): string {
  const program = new Command("ech0");
  registerCatalogCommands(program, parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG))), {
    run: async () => ({ human: "", json: "" }),
    scope: () => true,
  });
  const node = nodeAt(program, id.split(" "));
  if (node === undefined) throw new Error(`no command ${id}`);
  return renderHelp(helpModel(node));
}

function withCatalog(): FakeApi {
  api = startFakeApi({ "GET /api/_catalog": () => ({ body: CATALOG }) });
  return api;
}

/** Terminal cells, counting East Asian wide characters twice. */
function cells(text: string): number {
  let total = 0;
  for (const character of text) {
    const code = character.codePointAt(0) ?? 0;
    total += code >= 0x2e80 && code <= 0xff60 ? 2 : 1;
  }
  return total;
}

describe("help", () => {
  test("root help maps local commands and catalog namespaces", async () => {
    const server = withCatalog();
    const result = await run(["--help"], { ECH0_SERVER: server.baseUrl });

    expect(result.exitCode).toBe(0);
    expect(result.stdout).toContain("用法: ech0 [options] [command]");
    expect(result.stdout).toContain("本地命令");
    expect(result.stdout).toContain("auth");
    expect(result.stdout).toContain("服务端命令");
    expect(result.stdout).toContain("3 个操作");
    expect(result.stdout).toContain("catalog 共 5 个操作，CLI 当前开放 3 个");
    expect(result.stdout).toContain("示例");
  });

  test("a namespace lists the words you type under it", async () => {
    const server = withCatalog();
    const result = await run(["help", "echo"], { ECH0_SERVER: server.baseUrl });

    expect(result.exitCode).toBe(0);
    expect(result.stdout).toContain("子命令");
    expect(result.stdout).toContain("list");
    expect(result.stdout).toContain("分页/筛选 Echo 列表");
  });

  test("a descriptor carries usage, fields, the server card and an example", async () => {
    const server = withCatalog();
    const result = await run(["help", "echo", "create"], { ECH0_SERVER: server.baseUrl });

    expect(result.exitCode).toBe(0);
    expect(result.stdout).toContain("用法: ech0 echo create [options]");
    expect(result.stdout).toContain("接口: POST /api/echo · 写入 · scope: echo:write");
    expect(result.stdout).toContain("--content <content>");
    expect(result.stdout).toContain("必填");
    expect(result.stdout).toContain("ech0 echo create --content \"今天的想法\" --tag daily");
    expect(result.stdout).toContain("全局选项");
    expect(result.stdout).toContain("提示");
    expect(result.stdout).toContain("需要 scope：echo:write");
  });

  test("compose flags explain themselves and object arrays demand JSON", async () => {
    const server = withCatalog();
    const help = await run(["help", "echo", "create"], { ECH0_SERVER: server.baseUrl });
    expect(help.stdout).toContain("--echo-files <json>");
    expect(help.stdout).toContain("--tags <json>");
    expect(help.stdout).toContain("--tag 追加标签（可重复）");
    expect(help.stdout).toContain("--file-id 引用已上传文件");
    expect(help.stdout).toContain("--from/--edit 提交完整 payload");

    const rejected = await run(["echo", "create", "--tags", "a,b"], { ECH0_SERVER: server.baseUrl });
    expect(rejected.exitCode).toBe(2);
    expect(rejected.stderr).toContain("--tags 的元素是对象，需要 JSON 数组");

    const accepted = await run(
      ["echo", "create", "--tags", '[{"name":"daily"}]', "--content", "hi"],
      { ECH0_SERVER: server.baseUrl },
    );
    expect(accepted.stderr).not.toContain("需要 JSON 数组");
  });

  test("the login descriptor teaches every login mode", async () => {
    const result = await run(["help", "auth", "login"]);

    expect(result.stdout).toContain("设备码与确认链接");
    expect(result.stdout).toContain("--token -");
    expect(result.stdout).toContain("audience=cli");
    expect(result.stdout).toContain("默认 echo:read,echo:write,profile:read");
  });

  test("--help and `help <id>` render the same descriptor", async () => {
    const server = withCatalog();
    const viaFlag = await run(["echo", "create", "--help"], { ECH0_SERVER: server.baseUrl });
    const viaHelp = await run(["help", "echo", "create"], { ECH0_SERVER: server.baseUrl });

    expect(viaFlag.stdout).toBe(viaHelp.stdout);
  });

  test("keywords search ids, summaries and field names", async () => {
    const server = withCatalog();
    const result = await run(["help", "list"], { ECH0_SERVER: server.baseUrl });

    expect(result.stdout).toContain("echo list");
    expect(result.stdout).toContain("匹配「list」");
  });

  test("a miss suggests the nearest command instead of failing", async () => {
    const server = withCatalog();
    const result = await run(["echos"], { ECH0_SERVER: server.baseUrl });

    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain("未知命令：echos");
    expect(result.stderr).toContain("是否想运行：ech0 echo");
  });

  test("--json emits the descriptor, not the text view", async () => {
    const server = withCatalog();
    const result = await run(["help", "echo", "create", "--json"], { ECH0_SERVER: server.baseUrl });
    const parsed = JSON.parse(result.stdout) as { path: string[]; facet: string; fields: unknown[] };

    expect(parsed.path).toEqual(["echo", "create"]);
    expect(parsed.facet).toContain("POST /api/echo");
    expect(result.stdout).not.toContain("用法:");
    expect(Array.isArray(parsed.fields)).toBe(true);
  });

  test("operations outside the curated scope stay out of the command tree", async () => {
    const server = withCatalog();
    const chat = await run(["chat", "ask"], { ECH0_SERVER: server.baseUrl });
    expect(chat.exitCode).toBe(2);
    expect(chat.stderr).toContain("未知命令");

    const search = await run(["help", "ask"], { ECH0_SERVER: server.baseUrl });
    expect(search.stdout).not.toContain("chat ask");
  });

  test("long descriptions wrap inside the terminal width", () => {
    const columns = Object.getOwnPropertyDescriptor(process.stdout, "columns");
    Object.defineProperty(process.stdout, "columns", { value: 60, configurable: true });
    try {
      const text = helpText("echo create");
      for (const line of text.split("\n")) expect(cells(line)).toBeLessThanOrEqual(60);
      expect(text).toContain("owner 可见");
    } finally {
      if (columns === undefined) delete (process.stdout as { columns?: number }).columns;
      else Object.defineProperty(process.stdout, "columns", columns);
    }
  });

  test("an unauthenticated map points at auth login", async () => {
    const server = withCatalog();
    const result = await run(["help"], { ECH0_SERVER: server.baseUrl });

    expect(result.stdout).toContain("尚未登录");
    expect(result.stdout).toContain("ech0 auth login");
  });

  test("a failed call explains the next step", async () => {
    api = startFakeApi({
      "GET /api/_catalog": () => ({ body: CATALOG }),
      "POST /api/echo": () => ({
        status: 401,
        body: { code: 0, msg: "token expired", error_code: "invalid_token" },
      }),
    });
    const result = await run(["echo", "create", "--content", "hi"], {
      ECH0_SERVER: api.baseUrl,
      ECH0_TOKEN: "expired",
      ECH0_PROFILE: "test",
    });

    expect(result.stderr).not.toBe("");
    expect(result.exitCode).toBe(1);
    expect(result.stderr).toContain("token expired");
    expect(result.stderr).toContain("提示：token 无效或已过期");
    expect(result.stderr).toContain("ech0 auth login");
  });

  test("a rejected request points at the command descriptor", async () => {
    let received: unknown;
    api = startFakeApi({
      "GET /api/_catalog": () => ({ body: CATALOG }),
      "POST /api/echo/query": (request) => {
        received = request.body;
        return {
          status: 400,
          body: { code: 0, msg: "无效的请求", error_code: "INVALID_REQUEST", message_key: "common.invalid_request" },
        };
      },
    });
    const result = await run(["echo", "list"], {
      ECH0_SERVER: api.baseUrl,
      ECH0_TOKEN: "tok",
      ECH0_PROFILE: "test",
    });

    expect(received).toEqual({});
    expect(result.exitCode).toBe(1);
    expect(result.stderr).toContain("无效的请求");
    expect(result.stderr).toContain("ech0 echo list --help");
  });

  test("missing required flags print the runnable usage", async () => {
    const server = withCatalog();
    const result = await run(["echo", "create"], {
      ECH0_SERVER: server.baseUrl,
      ECH0_TOKEN: "tok",
      ECH0_PROFILE: "test",
    });

    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain("缺少必填参数 --content");
    expect(result.stderr).toContain("ech0 echo create --content <content>");
  });

  test("without an instance the map still explains how to get one", async () => {
    const result = await run(["help"]);

    expect(result.exitCode).toBe(0);
    expect(result.stdout).toContain("服务端命令");
    expect(result.stdout).toContain("未配置实例");
    expect(result.stdout).toContain("auth login");
  });
});
