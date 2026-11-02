// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { nodeAt } from "../src/help/model.ts";
import { Ech0Error } from "../src/output/console.ts";
import { buildProgram, helpText } from "./dynamic-harness.ts";

const BODY_PROPERTIES = {
  name: { type: "string", description: "Webhook 显示名" },
  url: { type: "string", description: "接收事件的 HTTP(S) 地址" },
  secret: { type: "string", description: "可选的签名密钥" },
  is_active: { type: "boolean", description: "是否启用" },
};
const BODY_BIND = { name: "body", url: "body", secret: "body", is_active: "body" };

const CATALOG: { ops: Record<string, unknown> } = {
  ops: {
    "webhook.list": {
      summary: "获取所有 Webhook",
      method: "GET",
      path: "/webhook",
      scopes: ["admin:settings"],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "webhook.create": {
      summary: "创建 Webhook",
      method: "POST",
      path: "/webhook",
      scopes: ["admin:settings"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: BODY_PROPERTIES,
        required: ["name", "url"],
      },
      bind: BODY_BIND,
      deprecated: false,
    },
    "webhook.update": {
      summary: "更新 Webhook",
      method: "PUT",
      path: "/webhook/{id}",
      scopes: ["admin:settings"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: {
          id: { type: "string", format: "uuid", description: "Webhook ID（UUID）" },
          ...BODY_PROPERTIES,
        },
        required: ["id", "name", "url"],
      },
      bind: { id: "path", ...BODY_BIND },
      deprecated: false,
    },
    "webhook.delete": {
      summary: "删除 Webhook",
      method: "DELETE",
      path: "/webhook/{id}",
      scopes: ["admin:settings"],
      effect: "destructive",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string", format: "uuid", description: "Webhook ID（UUID）" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
    "webhook.test": {
      summary: "测试 Webhook",
      method: "POST",
      path: "/webhook/{id}/test",
      scopes: ["admin:settings"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string", format: "uuid", description: "Webhook ID（UUID）" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
  },
};

describe("webhook commands", () => {
  test("exposes list, create, update, delete and test", () => {
    const { program } = buildProgram(CATALOG);
    const group = nodeAt(program, ["webhook"]);

    expect((group?.commands ?? []).map((command) => command.name()).sort()).toEqual([
      "create",
      "delete",
      "list",
      "test",
      "update",
    ]);
  });

  test("list, create, update, delete and test map to their operations", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(["webhook", "list"], { from: "user" });
    await program.parseAsync(
      [
        "webhook",
        "create",
        "--name",
        "ci",
        "--url",
        "https://example.com/hook",
        "--secret",
        "s3cr3t",
        "--is-active",
      ],
      { from: "user" },
    );
    await program.parseAsync(
      [
        "webhook",
        "update",
        "--id",
        "w-1",
        "--name",
        "ci",
        "--url",
        "https://example.com/hook",
        "--is-active",
      ],
      { from: "user" },
    );
    await program.parseAsync(["webhook", "delete", "--id", "w-1", "--yes", "--json"], { from: "user" });
    await program.parseAsync(["webhook", "test", "--id", "w-1"], { from: "user" });

    expect(invocations).toEqual([
      { request: "GET /webhook", values: {} },
      {
        request: "POST /webhook",
        values: {
          name: "ci",
          url: "https://example.com/hook",
          secret: "s3cr3t",
          is_active: true,
        },
      },
      {
        request: "PUT /webhook/{id}",
        values: { id: "w-1", name: "ci", url: "https://example.com/hook", is_active: true },
      },
      { request: "DELETE /webhook/{id}", values: { id: "w-1" } },
      { request: "POST /webhook/{id}/test", values: { id: "w-1" } },
    ]);
  });

  test("create requires --name and --url before the request", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await expect(program.parseAsync(["webhook", "create", "--name", "ci"], { from: "user" })).rejects.toThrow(
      "缺少必填参数 --url",
    );
    await expect(
      program.parseAsync(["webhook", "create", "--url", "https://example.com/hook"], { from: "user" }),
    ).rejects.toThrow("缺少必填参数 --name");
    expect(invocations).toHaveLength(0);
  });

  test("delete needs --yes when non-interactive", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await expect(
      program.parseAsync(["webhook", "delete", "--id", "w-1", "--json"], { from: "user" }),
    ).rejects.toThrow("--yes");
    expect(invocations).toHaveLength(0);
  });

  test("test delivery failure gets a delivery hint, not a field hint", async () => {
    const { program } = buildProgram(CATALOG, async () => {
      throw new Ech0Error("连接目标地址不被允许", { status: 400 });
    });

    const failure = await program
      .parseAsync(["webhook", "test", "--id", "w-1"], { from: "user" })
      .catch((error: unknown) => error);

    expect(failure).toBeInstanceOf(Ech0Error);
    const hint = (failure as Ech0Error).details.hint ?? "";
    expect(hint).toContain("Webhook 投递失败");
    expect(hint).not.toContain("--help");
  });

  test("help explains admin, signing and full-replace semantics", () => {
    const { program } = buildProgram(CATALOG);

    const create = helpText(program, ["webhook", "create"]);
    expect(create).toContain("需要管理员账号");
    expect(create).toContain("X-Ech0-Signature");
    expect(create).toContain("--is-active");

    const update = helpText(program, ["webhook", "update"]);
    expect(update).toContain("整体覆盖");
    expect(update).toContain("省略 --is-active");
    expect(update).toContain("省略 --secret");

    const remove = helpText(program, ["webhook", "delete"]);
    expect(remove).toContain("-y, --yes");

    const probe = helpText(program, ["webhook", "test"]);
    expect(probe).toContain("测试事件");
  });
});
