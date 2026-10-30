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
      summary: "上传文件（multipart）",
      method: "POST",
      path: "/files/upload",
      effect: "write",
      input: {
        type: "object",
        properties: {
          file: { type: "string", format: "binary" },
          category: { type: "string" },
          storage_type: { type: "string" },
        },
        required: ["file"],
      },
      bind: { file: "form", category: "form", storage_type: "form" },
    }),
    "file.external": baseOp({
      summary: "登记外链文件",
      method: "POST",
      path: "/files/external",
      effect: "write",
      input: {
        type: "object",
        properties: { url: { type: "string" }, category: { type: "string" } },
        required: ["url"],
      },
      bind: { url: "body", category: "body" },
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
          storage_type: { type: "string" },
        },
      },
      bind: { page: "query", pageSize: "query", search: "query", storage_type: "query" },
    }),
    "file.get": baseOp({
      summary: "获取文件元信息",
      path: "/file/{id}",
      input: {
        type: "object",
        properties: { id: { type: "string" } },
        required: ["id"],
      },
      bind: { id: "path" },
    }),
    "file.tree": baseOp({
      summary: "获取文件树",
      path: "/file/tree",
      input: {
        type: "object",
        properties: { storage_type: { type: "string" }, prefix: { type: "string" } },
        required: ["storage_type"],
      },
      bind: { storage_type: "query", prefix: "query" },
    }),
    "file.stream.by_id": baseOp({
      summary: "按文件 ID 读取内容（二进制）",
      path: "/file/{id}/stream",
      kind: "binary",
      input: {
        type: "object",
        properties: { id: { type: "string" } },
        required: ["id"],
      },
      bind: { id: "path" },
      transport: { response: { kind: "binary" } },
    }),
    // Present in the catalog but deliberately not exposed by the CLI.
    "file.stream": baseOp({
      summary: "按存储路径读取文件内容（二进制）",
      path: "/file/stream",
      kind: "binary",
      input: {
        type: "object",
        properties: { path: { type: "string" }, storage_type: { type: "string" } },
        required: ["path", "storage_type"],
      },
      bind: { path: "query", storage_type: "query" },
      transport: { response: { kind: "binary" } },
    }),
    "file.delete": baseOp({
      summary: "删除文件",
      method: "DELETE",
      path: "/file/{id}",
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
      "download",
      "external",
      "get",
      "list",
      "tree",
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

  test("file tree requires --storage-type", async () => {
    const invoked: Invocation[] = [];
    const program = programWith(invoked);

    await expect(program.parseAsync(["file", "tree"], { from: "user" })).rejects.toThrow(
      "缺少必填参数 --storage-type",
    );

    await program.parseAsync(["file", "tree", "--storage-type", "local"], { from: "user" });
    expect(invoked[0]?.values).toEqual({ storage_type: "local" });
  });

  test("file download maps to the binary stream-by-id operation with --output", async () => {
    const invoked: Invocation[] = [];
    const program = programWith(invoked);

    await program.parseAsync(
      ["file", "download", "--id", "f1", "--output", "/tmp/out.bin"],
      { from: "user" },
    );

    expect(invoked[0]?.op.path).toBe("/file/{id}/stream");
    expect(invoked[0]?.values).toEqual({ id: "f1" });
    expect(invoked[0]?.extra.output).toBe("/tmp/out.bin");
  });
});
