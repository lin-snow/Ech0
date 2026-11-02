// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { nodeAt } from "../src/help/model.ts";
import { buildProgram, helpText } from "./dynamic-harness.ts";

/** The core comment subset, plus ids that must stay out of scope. */
const CATALOG: { ops: Record<string, unknown> } = {
  ops: {
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
    "comment.create.integration": {
      summary: "经访问令牌创建评论（集成）",
      method: "POST",
      path: "/comments/integration",
      scopes: ["comment:write"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: {
          echo_id: { type: "string" },
          content: { type: "string" },
          nickname: { type: "string" },
          metadata: { type: "string" },
        },
        required: ["echo_id", "content"],
      },
      bind: { echo_id: "body", content: "body", nickname: "body", metadata: "body" },
      deprecated: false,
    },
    "comment.list.public": {
      summary: "列出最新公开评论",
      method: "GET",
      path: "/comments/public",
      scopes: [],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: { limit: { type: "integer" } } },
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
        properties: { id: { type: "string" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
  },
};

describe("comment commands", () => {
  test("exposes only create and list", () => {
    const { program } = buildProgram(CATALOG);
    const group = nodeAt(program, ["comment"]);

    expect((group?.commands ?? []).map((command) => command.name()).sort()).toEqual(["create", "list"]);
    expect(nodeAt(program, ["comment", "public"])).toBeUndefined();
    expect(nodeAt(program, ["comment", "panel"])).toBeUndefined();
  });

  test("comment create posts the integration body and takes no form token", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(
      [
        "comment",
        "create",
        "--echo-id",
        "e1",
        "--content",
        "你好",
        "--nickname",
        "小明",
        "--metadata",
        "from-test",
      ],
      { from: "user" },
    );

    expect(invocations).toEqual([
      {
        request: "POST /comments/integration",
        values: { echo_id: "e1", content: "你好", nickname: "小明", metadata: "from-test" },
      },
    ]);

    const text = helpText(program, ["comment", "create"]);
    expect(text).toContain("--echo-id");
    expect(text).toContain("--content");
    expect(text).toContain("source=integration");
    expect(text).toContain("pending");
    for (const dropped of ["--form-token", "--captcha-token", "--email", "--website", "--parent-id", "--hp-field"]) {
      expect(text).not.toContain(dropped);
    }
  });

  test("required fields are checked before the request", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await expect(
      program.parseAsync(["comment", "create", "--echo-id", "e1"], { from: "user" }),
    ).rejects.toThrow("缺少必填参数 --content");
    await expect(program.parseAsync(["comment", "list"], { from: "user" })).rejects.toThrow(
      "缺少必填参数 --echo-id",
    );
    expect(invocations).toHaveLength(0);
  });

  test("comment list binds echo_id to the query", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(["comment", "list", "--echo-id", "e1"], { from: "user" });
    expect(invocations).toEqual([{ request: "GET /comments", values: { echo_id: "e1" } }]);
  });
});
