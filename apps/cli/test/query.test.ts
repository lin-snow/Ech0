// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { UsageError } from "../src/output/console.ts";
import { buildProgram, helpText } from "./dynamic-harness.ts";

const op = (overrides: Record<string, unknown>) => ({
  summary: "test",
  method: "GET",
  path: "/test",
  scopes: [],
  effect: "read",
  kind: "json",
  input: { type: "object", properties: {} },
  bind: {},
  deprecated: false,
  ...overrides,
});

const CATALOG = {
  ops: {
    "echo.list": op({
      summary: "分页/筛选 Echo 列表",
      method: "POST",
      path: "/echo/query",
      scopes: ["echo:read"],
      input: {
        type: "object",
        properties: {
          page: { type: "integer", description: "Page number, 1-based (default 1)" },
          pageSize: { type: "integer", description: "Items per page, 1-100 (default 10)" },
          search: { type: "string" },
          tagIds: { type: ["array", "null"], items: { type: "string" } },
          sortBy: { type: "string", enum: ["created_at", "reactions"], description: "Sort field" },
          sortOrder: { type: "string", enum: ["asc", "desc"] },
          private: { type: "boolean", description: "Visibility filter" },
        },
      },
      bind: {
        page: "body",
        pageSize: "body",
        search: "body",
        tagIds: "body",
        sortBy: "body",
        sortOrder: "body",
        private: "body",
      },
    }),
    "echo.get": op({
      summary: "获取指定 ID 的 Echo",
      path: "/echo/{id}",
      input: { type: "object", properties: { id: { type: "string" } }, required: ["id"] },
      bind: { id: "path" },
    }),
    "echo.hot": op({
      summary: "获取热门 Echo",
      path: "/echo/hot",
      input: { type: "object", properties: { limit: { type: "integer" } } },
      bind: { limit: "query" },
    }),
    "echo.today": op({ summary: "获取今天的 Echo", path: "/echo/today" }),
    "echo.random": op({ summary: "随机返回一篇 Echo", path: "/echo/random" }),
    "echo.onthisday": op({ summary: "那年今日", path: "/echo/onthisday" }),
  },
};

async function valuesFor(argv: string[]): Promise<Record<string, unknown>> {
  const { program, invocations } = buildProgram(CATALOG);
  await program.parseAsync(argv, { from: "user" });
  return invocations[0]?.values ?? {};
}

describe("echo query flags", () => {
  test("accumulates repeated arrays and keeps the other filters", async () => {
    const values = await valuesFor([
      "echo",
      "list",
      "--page-size",
      "20",
      "--tag-ids",
      "a",
      "--tag-ids",
      "b",
      "--private",
    ]);

    expect(values).toEqual({ pageSize: 20, tagIds: ["a", "b"], private: true });
  });

  test("distinguishes an omitted boolean from false", async () => {
    expect(await valuesFor(["echo", "list"])).toEqual({});
    expect(await valuesFor(["echo", "list", "--no-private"])).toEqual({ private: false });
  });

  test("rejects non-numeric values instead of sending null", async () => {
    await expect(valuesFor(["echo", "list", "--page", "abc"])).rejects.toBeInstanceOf(UsageError);
    await expect(valuesFor(["echo", "list", "--page-size", "1.5"])).rejects.toBeInstanceOf(UsageError);
  });

  test("accepts only the documented sort values", async () => {
    expect(await valuesFor(["echo", "list", "--sort-by", "reactions"])).toEqual({ sortBy: "reactions" });
    await expect(valuesFor(["echo", "list", "--sort-by", "favorites"])).rejects.toBeInstanceOf(UsageError);
  });

  test("invokes the read views", async () => {
    expect(await valuesFor(["echo", "hot", "--limit", "10"])).toEqual({ limit: 10 });
    expect(await valuesFor(["echo", "today"])).toEqual({});
    expect(await valuesFor(["echo", "random"])).toEqual({});
    expect(await valuesFor(["echo", "onthisday"])).toEqual({});
  });
});

describe("echo query help", () => {
  test("renders field docs, the read effect and both boolean forms", () => {
    const { program } = buildProgram(CATALOG);
    const text = helpText(program, ["echo", "list"]);

    expect(text).toContain("Page number, 1-based (default 1)");
    expect(text).toContain("Items per page, 1-100 (default 10)");
    expect(text).toContain("只读");
    expect(text).toContain("--[no-]private");
  });
});
