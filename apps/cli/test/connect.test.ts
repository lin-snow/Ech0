// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { nodeAt } from "../src/help/model.ts";
import { buildProgram, helpText } from "./dynamic-harness.ts";

/** The core Connect subset, plus ids that must stay out of scope. */
const CATALOG: { ops: Record<string, unknown> } = {
  ops: {
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
    "connect.add": {
      summary: "添加连接",
      method: "POST",
      path: "/connects",
      scopes: ["connect:write"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: { connect_url: { type: "string" } },
        required: ["connect_url"],
      },
      bind: { connect_url: "body" },
      deprecated: false,
    },
    "connect.delete": {
      summary: "删除连接",
      method: "DELETE",
      path: "/connects/{id}",
      scopes: ["connect:delete"],
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
  },
};

describe("connect commands", () => {
  test("exposes add, delete, health, info and list", () => {
    const { program } = buildProgram(CATALOG);
    const group = nodeAt(program, ["connect"]);

    expect((group?.commands ?? []).map((command) => command.name()).sort()).toEqual([
      "add",
      "delete",
      "health",
      "info",
      "list",
    ]);
    expect(nodeAt(program, ["connect", "status"])).toBeUndefined();
  });

  test("connect list, info, add and delete map to their operations", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(["connect", "list"], { from: "user" });
    await program.parseAsync(["connect", "info"], { from: "user" });
    await program.parseAsync(["connect", "health"], { from: "user" });
    await program.parseAsync(["connect", "add", "--connect-url", "https://peer.example"], { from: "user" });
    await program.parseAsync(["connect", "delete", "--id", "c-1", "--yes", "--json"], { from: "user" });

    expect(invocations).toEqual([
      { request: "GET /connect/list", values: {} },
      { request: "GET /connects/info", values: {} },
      { request: "GET /connects/health", values: {} },
      { request: "POST /connects", values: { connect_url: "https://peer.example" } },
      { request: "DELETE /connects/{id}", values: { id: "c-1" } },
    ]);

    const addHelp = helpText(program, ["connect", "add"]);
    expect(addHelp).toContain("--connect-url");
    expect(addHelp).not.toContain("--id");
    expect(addHelp).toContain("/api/connect");
  });

  test("connect add requires --connect-url before the request", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await expect(program.parseAsync(["connect", "add"], { from: "user" })).rejects.toThrow(
      "缺少必填参数 --connect-url",
    );
    expect(invocations).toHaveLength(0);
  });

  test("connect delete needs --yes when non-interactive", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await expect(
      program.parseAsync(["connect", "delete", "--id", "c-1", "--json"], { from: "user" }),
    ).rejects.toThrow("--yes");
    expect(invocations).toHaveLength(0);

    const text = helpText(program, ["connect", "delete"]);
    expect(text).toContain("-y, --yes");
    expect(text).toContain("--yes");
  });
});
