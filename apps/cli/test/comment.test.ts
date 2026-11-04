// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { parseCatalog } from "../src/catalog/types.ts";
import { buildRequest } from "../src/commands/operation.ts";
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
    "comment.panel.list": {
      summary: "管理面板列出评论",
      method: "GET",
      path: "/panel/comments",
      scopes: ["comment:review"],
      effect: "read",
      kind: "json",
      input: {
        type: "object",
        properties: {
          page: { type: "integer" },
          page_size: { type: "integer" },
          status: { type: "string", enum: ["pending", "approved", "rejected"] },
        },
      },
      bind: { page: "query", page_size: "query", status: "query" },
      deprecated: false,
    },
    "comment.panel.status": {
      summary: "更新评论审核状态",
      method: "PATCH",
      path: "/panel/comments/{id}/status",
      scopes: ["comment:review"],
      effect: "write",
      kind: "json",
      input: {
        type: "object",
        properties: {
          id: { type: "string" },
          status: { type: "string", enum: ["pending", "approved", "rejected"] },
        },
        required: ["id", "status"],
      },
      bind: { id: "path", status: "body" },
      deprecated: false,
    },
    "comment.panel.delete": {
      summary: "删除评论",
      method: "DELETE",
      path: "/panel/comments/{id}",
      scopes: ["comment:delete"],
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
  test("exposes publishing, reading and the review lifecycle", () => {
    const { program } = buildProgram(CATALOG);
    const group = nodeAt(program, ["comment"]);

    expect((group?.commands ?? []).map((command) => command.name()).sort()).toEqual([
      "create",
      "list",
      "panel",
    ]);
    expect(nodeAt(program, ["comment", "public"])).toBeUndefined();
    expect(
      (nodeAt(program, ["comment", "panel"])?.commands ?? []).map((command) => command.name()).sort(),
    ).toEqual(["delete", "list", "status"]);
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

  test("comment panel list, status and delete map to the review endpoints", async () => {
    const { program, invocations } = buildProgram(CATALOG);

    await program.parseAsync(["comment", "panel", "list", "--status", "pending"], { from: "user" });
    await program.parseAsync(
      ["comment", "panel", "status", "--id", "c-1", "--status", "approved"],
      { from: "user" },
    );
    await expect(
      program.parseAsync(["comment", "panel", "delete", "--id", "c-1", "--json"], { from: "user" }),
    ).rejects.toThrow("--yes");
    await program.parseAsync(["comment", "panel", "delete", "--id", "c-1", "--yes"], { from: "user" });

    expect(invocations).toEqual([
      { request: "GET /panel/comments", values: { status: "pending" } },
      { request: "PATCH /panel/comments/{id}/status", values: { id: "c-1", status: "approved" } },
      { request: "DELETE /panel/comments/{id}", values: { id: "c-1" } },
    ]);

    const listHelp = helpText(program, ["comment", "panel", "list"]);
    expect(listHelp).toContain("--status");
    expect(listHelp).toContain("comment:review");
    expect(helpText(program, ["comment", "panel", "status"])).toContain("rejected");
  });

  test("review status is a PATCH with the id in the path and status in the body", () => {
    const op = parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG))).ops["comment.panel.status"];
    if (op === undefined) throw new Error("fixture op missing");

    const spec = buildRequest(op, { id: "c-1", status: "approved" });

    expect(spec.method).toBe("PATCH");
    expect(spec.path).toBe("/panel/comments/c-1/status");
    expect(spec.body).toEqual({ status: "approved" });
  });
});
