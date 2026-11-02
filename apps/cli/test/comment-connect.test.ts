// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { Command } from "commander";

import { parseCatalog, type CatalogOp } from "../src/catalog/types.ts";
import { registerCatalogCommands } from "../src/commands/dynamic.ts";
import type { FieldValues } from "../src/commands/operation.ts";
import type { CommandOutput } from "../src/commands/types.ts";
import { helpModel, nodeAt } from "../src/help/model.ts";
import { renderHelp } from "../src/help/render.ts";

/** The comment/connect subset of the real catalog, plus ids outside the scope. */
const CATALOG: { ops: Record<string, unknown> } = {
  ops: {
    "comment.create": {
      summary: "创建公开评论",
      method: "POST",
      path: "/comments",
      scopes: [],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        additionalProperties: true,
        properties: {
          captcha_token: { type: "string" },
          content: { type: "string" },
          echo_id: { type: "string" },
          email: { type: "string" },
          form_token: { type: "string" },
          hp_field: { type: "string" },
          nickname: { type: "string" },
          parent_id: { type: "string" },
          website: { type: "string" },
        },
      },
      bind: {
        captcha_token: "body",
        content: "body",
        echo_id: "body",
        email: "body",
        form_token: "body",
        hp_field: "body",
        nickname: "body",
        parent_id: "body",
        website: "body",
      },
      deprecated: false,
    },
    "comment.list.by.echo": {
      summary: "按 echo 列出公开评论",
      method: "GET",
      path: "/comments",
      scopes: [],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: { echo_id: { type: "string", description: "Echo ID" } },
        required: ["echo_id"],
      },
      bind: { echo_id: "query" },
      deprecated: false,
    },
    "comment.list.public": {
      summary: "列出最新公开评论",
      method: "GET",
      path: "/comments/public",
      scopes: [],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: { limit: { type: "integer", format: "int64", default: 30, description: "返回数量上限" } },
      },
      bind: { limit: "query" },
      deprecated: false,
    },
    "comment.panel.delete": {
      summary: "删除评论",
      method: "DELETE",
      path: "/panel/comments/{id}",
      scopes: ["comment:moderate"],
      effect: "destructive",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string", description: "评论 ID" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
    "comment.form.meta": {
      summary: "获取评论表单元信息",
      method: "GET",
      path: "/comments/form",
      scopes: [],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "comment.create.integration": {
      summary: "经访问令牌创建评论（集成）",
      method: "POST",
      path: "/comments/integration",
      scopes: ["comment:write"],
      effect: "write",
      kind: "json",
      input: { type: "object", properties: { content: { type: "string" } } },
      bind: { content: "body" },
      deprecated: false,
    },
    "comment.panel.list": {
      summary: "管理面板列出评论",
      method: "GET",
      path: "/panel/comments",
      scopes: ["comment:moderate"],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "connect.self": {
      summary: "获取当前实例的连接信息",
      method: "GET",
      path: "/connect",
      scopes: [],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "connect.list": {
      summary: "获取当前实例添加的所有连接",
      method: "GET",
      path: "/connect/list",
      scopes: [],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "connect.add": {
      summary: "添加连接",
      method: "POST",
      path: "/connects",
      scopes: ["connect:write"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        additionalProperties: true,
        properties: { id: { type: "string" }, connect_url: { type: "string" } },
      },
      bind: { id: "body", connect_url: "body" },
      deprecated: false,
    },
    "connect.delete": {
      summary: "删除连接",
      method: "DELETE",
      path: "/connects/{id}",
      scopes: ["connect:write"],
      effect: "destructive",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string", format: "uuid", description: "连接 ID（UUID）" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
    "connect.health": {
      summary: "获取互联健康状态",
      method: "GET",
      path: "/connects/health",
      scopes: ["connect:read"],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "connect.info": {
      summary: "获取所有已添加连接的详细信息",
      method: "GET",
      path: "/connects/info",
      scopes: [],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
  },
};

interface Invocation {
  request: string;
  values: FieldValues;
}

function buildProgram(): { program: Command; invocations: Invocation[] } {
  const invocations: Invocation[] = [];
  const program = new Command("ech0")
    .option("--json", "输出 JSON", false)
    .option("--yaml", "输出 YAML", false);
  const catalog = parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG)));
  registerCatalogCommands(program, catalog, {
    run: async (_command: Command, op: CatalogOp, values: FieldValues): Promise<CommandOutput> => {
      invocations.push({ request: `${op.method} ${op.path}`, values });
      return { human: "", json: "" };
    },
  });
  const override = (command: Command): void => {
    command.exitOverride();
    for (const child of command.commands) override(child);
  };
  override(program);
  return { program, invocations };
}

function groupNames(program: Command, group: string): string[] {
  const node = nodeAt(program, [group]);
  return (node?.commands ?? []).map((command) => command.name()).sort();
}

function helpText(program: Command, path: string[]): string {
  const node = nodeAt(program, path);
  if (node === undefined) throw new Error(`no command ${path.join(" ")}`);
  return renderHelp(helpModel(node));
}

describe("comment and connect commands", () => {
  test("exposes both groups and keeps every other catalog op out", () => {
    const { program } = buildProgram();

    expect(groupNames(program, "comment")).toEqual(["create", "delete", "list", "public"]);
    expect(groupNames(program, "connect")).toEqual(["add", "health", "list", "remove", "status"]);

    expect(nodeAt(program, ["comment", "form"])).toBeUndefined();
    expect(nodeAt(program, ["comment", "panel"])).toBeUndefined();
    expect(nodeAt(program, ["comment", "create-integration"])).toBeUndefined();
    expect(nodeAt(program, ["connect", "info"])).toBeUndefined();
  });

  test("comment create takes the body fields from the catalog schema", async () => {
    const { program, invocations } = buildProgram();

    await program.parseAsync(
      ["comment", "create", "--echo-id", "e1", "--content", "你好", "--nickname", "小明", "--form-token", "tok"],
      { from: "user" },
    );

    expect(invocations).toEqual([
      {
        request: "POST /comments",
        values: { echo_id: "e1", content: "你好", nickname: "小明", form_token: "tok" },
      },
    ]);
  });

  test("comment reads bind echo_id and limit to the query", async () => {
    const { program, invocations } = buildProgram();

    await program.parseAsync(["comment", "list", "--echo-id", "e1"], { from: "user" });
    await program.parseAsync(["comment", "public", "--limit", "5"], { from: "user" });

    expect(invocations).toEqual([
      { request: "GET /comments", values: { echo_id: "e1" } },
      { request: "GET /comments/public", values: { limit: 5 } },
    ]);
  });

  test("missing required comment input is a usage error", async () => {
    const { program } = buildProgram();

    await expect(program.parseAsync(["comment", "list"], { from: "user" })).rejects.toThrow(
      "缺少必填参数 --echo-id",
    );
  });

  test("connect add maps the url field and connect remove maps the id", async () => {
    const { program, invocations } = buildProgram();

    await program.parseAsync(["connect", "add", "--connect-url", "https://peer.example"], { from: "user" });
    await program.parseAsync(["connect", "remove", "--id", "c-1", "--yes"], { from: "user" });

    expect(invocations).toEqual([
      { request: "POST /connects", values: { connect_url: "https://peer.example" } },
      { request: "DELETE /connects/{id}", values: { id: "c-1" } },
    ]);
  });

  test("connect status, list and health need no fields", async () => {
    const { program, invocations } = buildProgram();

    await program.parseAsync(["connect", "status"], { from: "user" });
    await program.parseAsync(["connect", "list"], { from: "user" });
    await program.parseAsync(["connect", "health"], { from: "user" });

    expect(invocations.map((invocation) => invocation.request)).toEqual([
      "GET /connect",
      "GET /connect/list",
      "GET /connects/health",
    ]);
  });
});

describe("destructive comment and connect commands", () => {
  test("help lists --yes and the confirmation hint", () => {
    const { program } = buildProgram();

    for (const path of [["comment", "delete"], ["connect", "remove"]]) {
      const text = helpText(program, path);
      expect(text).toContain("-y, --yes");
      expect(text).toContain("交互确认");
    }
  });

  test("the guard follows the catalog effect, not the command name", () => {
    const { program } = buildProgram();

    expect(helpText(program, ["comment", "create"])).not.toContain("--yes");
  });

  test("comment delete without --yes never invokes the operation", async () => {
    const { program, invocations } = buildProgram();

    await expect(
      program.parseAsync(["comment", "delete", "--id", "c1", "--json"], { from: "user" }),
    ).rejects.toThrow("--yes");
    expect(invocations).toHaveLength(0);
  });

  test("comment delete with --yes invokes the operation", async () => {
    const { program, invocations } = buildProgram();

    await program.parseAsync(["comment", "delete", "--id", "c1", "--yes", "--json"], { from: "user" });
    expect(invocations).toEqual([{ request: "DELETE /panel/comments/{id}", values: { id: "c1" } }]);
  });

  test("connect remove without --yes never invokes the operation", async () => {
    const { program, invocations } = buildProgram();

    await expect(
      program.parseAsync(["connect", "remove", "--id", "c-1", "--json"], { from: "user" }),
    ).rejects.toThrow("--yes");
    expect(invocations).toHaveLength(0);
  });

  test("connect remove with --yes invokes the operation", async () => {
    const { program, invocations } = buildProgram();

    await program.parseAsync(["connect", "remove", "--id", "c-1", "--yes", "--json"], { from: "user" });
    expect(invocations).toEqual([{ request: "DELETE /connects/{id}", values: { id: "c-1" } }]);
  });
});
