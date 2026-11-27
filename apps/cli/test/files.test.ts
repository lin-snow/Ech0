// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { Command } from "commander";

import type { Catalog, CatalogOp } from "../src/catalog/types.ts";
import { registerCatalogCommands, type DynamicOptions } from "../src/commands/dynamic.ts";
import type { FieldValues } from "../src/commands/operation.ts";
import type { CommandOutput } from "../src/commands/types.ts";

function baseOp(overrides: Partial<CatalogOp>): CatalogOp {
  return {
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
  };
}

const FILE_OPS: Catalog = {
  ops: {
    "file.upload": baseOp({
      summary: "上传文件",
      method: "POST",
      path: "/files",
      effect: "write",
      input: {
        type: "object",
        properties: {
          file: { type: "string", format: "binary" },
          profile: { type: "string" },
          external_url: { type: "string" },
          mode: { type: "string" },
        },
      },
      bind: { file: "form", profile: "form", external_url: "form", mode: "form" },
    }),
    "file.list": baseOp({
      summary: "分页获取文件列表",
      path: "/files",
      input: {
        type: "object",
        properties: {
          page: { type: "integer" },
          pageSize: { type: "integer" },
          search: { type: "string" },
          kind: { type: "string" },
        },
      },
      bind: { page: "query", pageSize: "query", search: "query", kind: "query" },
    }),
    "file.get": baseOp({
      summary: "获取文件元信息",
      path: "/files/{id}",
      input: {
        type: "object",
        properties: { id: { type: "string" } },
        required: ["id"],
      },
      bind: { id: "path" },
    }),
    "file.rename": baseOp({
      summary: "重命名文件",
      method: "PATCH",
      path: "/files/{id}",
      effect: "write",
      input: {
        type: "object",
        properties: { id: { type: "string" }, body: { type: "object" } },
        required: ["id", "body"],
      },
      bind: { id: "path", body: "body" },
    }),
    "file.content": baseOp({
      summary: "读取文件内容（二进制）",
      path: "/files/{id}/content",
      kind: "binary",
      input: {
        type: "object",
        properties: { id: { type: "string" } },
        required: ["id"],
      },
      bind: { id: "path" },
      transport: { response: { kind: "binary" } },
    }),
    "file.delete": baseOp({
      summary: "删除文件",
      method: "DELETE",
      path: "/files/{id}",
      effect: "destructive",
      input: {
        type: "object",
        properties: { id: { type: "string" } },
        required: ["id"],
      },
      bind: { id: "path" },
    }),
  },
};

interface Invocation {
  op: CatalogOp;
  values: FieldValues;
  extra: Parameters<NonNullable<DynamicOptions["run"]>>[3];
}

function programWith(invoked: Invocation[]): Command {
  const program = new Command("ech0").option("--json", "json", false);
  registerCatalogCommands(program, FILE_OPS, {
    run: async (_command, op, values, extra): Promise<CommandOutput> => {
      invoked.push({ op, values, extra });
      return { human: "", json: "" };
    },
  });
  const override = (command: Command): void => {
    command.exitOverride();
    for (const child of command.commands) override(child);
  };
  override(program);
  return program;
}

describe("file command surface", () => {
  test("exposes browsing and upload, but not destructive or raw-path operations", () => {
    const program = programWith([]);
    const file = program.commands.find((command) => command.name() === "file");

    expect(file?.commands.map((command) => command.name()).sort()).toEqual([
      "delete",
      "download",
      "get",
      "list",
      "rename",
      "upload",
    ]);
  });

  test("file list forwards pagination and filters as query parameters", async () => {
    const invoked: Invocation[] = [];
    const program = programWith(invoked);

    await program.parseAsync(
      ["file", "list", "--page", "2", "--page-size", "50", "--search", "cat"],
      { from: "user" },
    );

    expect(invoked[0]?.op.path).toBe("/files");
    expect(invoked[0]?.values).toEqual({ page: 2, pageSize: 50, search: "cat" });
  });

  test("file get requires --id and sends it on the path", async () => {
    const invoked: Invocation[] = [];
    const program = programWith(invoked);

    await expect(program.parseAsync(["file", "get"], { from: "user" })).rejects.toThrow(
      "缺少必填参数 --id",
    );

    await program.parseAsync(["file", "get", "--id", "f1"], { from: "user" });
    expect(invoked[0]?.values).toEqual({ id: "f1" });
  });

  test("file list forwards kind as a query filter", async () => {
    const invoked: Invocation[] = [];
    const program = programWith(invoked);

    await program.parseAsync(["file", "list", "--kind", "blob"], { from: "user" });
    expect(invoked[0]?.values).toEqual({ kind: "blob" });
  });

  test("file download maps to the binary content gateway with --output", async () => {
    const invoked: Invocation[] = [];
    const program = programWith(invoked);

    await program.parseAsync(["file", "download", "--id", "f1", "--output", "/tmp/out.bin"], {
      from: "user",
    });

    expect(invoked[0]?.op.path).toBe("/files/{id}/content");
    expect(invoked[0]?.values).toEqual({ id: "f1" });
    expect(invoked[0]?.extra.output).toBe("/tmp/out.bin");
  });
});
