// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { nodeAt } from "../src/help/model.ts";
import { buildProgram, helpText } from "./dynamic-harness.ts";

/** The core tag subset, plus create which must stay out of scope. */
const CATALOG: { ops: Record<string, unknown> } = {
  ops: {
    "tag.list": {
      summary: "获取所有标签",
      method: "GET",
      path: "/tags",
      scopes: [],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
    "tag.delete": {
      summary: "删除标签",
      method: "DELETE",
      path: "/tag/{id}",
      scopes: ["echo:delete"],
      effect: "destructive",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string", format: "uuid", description: "标签 ID" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
    "tag.create": {
      summary: "创建标签",
      method: "POST",
      path: "/tag",
      scopes: ["echo:write"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: { name: { type: "string", description: "标签名" } },
        required: ["name"],
      },
      bind: { name: "body" },
      deprecated: false,
    },
  },
};

describe("tag commands", () => {
  test("exposes list and delete, not create", () => {
    const { program } = buildProgram(CATALOG);
    const group = nodeAt(program, ["tag"]);

    expect((group?.commands ?? []).map((command) => command.name()).sort()).toEqual(["delete", "list"]);
    expect(nodeAt(program, ["tag", "create"])).toBeUndefined();
  });

  test("tag list maps to GET /tags and points at --tag-ids", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(["tag", "list", "--json"], { from: "user" });
    expect(invocations).toEqual([{ request: "GET /tags", values: {} }]);

    const text = helpText(program, ["tag", "list"]);
    expect(text).toContain("usage_count");
    expect(text).toContain("echo list --tag-ids");
  });

  test("tag delete maps to DELETE /tag/{id} after --yes", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(["tag", "delete", "--id", "t-1", "--yes", "--json"], { from: "user" });
    expect(invocations).toEqual([{ request: "DELETE /tag/{id}", values: { id: "t-1" } }]);
  });

  test("tag delete needs --yes when non-interactive", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await expect(
      program.parseAsync(["tag", "delete", "--id", "t-1", "--json"], { from: "user" }),
    ).rejects.toThrow("--yes");
    expect(invocations).toHaveLength(0);
  });

  test("tag delete help warns about admin and cascading removal", () => {
    const { program } = buildProgram(CATALOG);
    const text = helpText(program, ["tag", "delete"]);

    expect(text).toContain("需要管理员账号");
    expect(text).toContain("所有 Echo");
    expect(text).toContain("-y, --yes");
  });
});
