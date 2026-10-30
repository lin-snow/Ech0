// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import { ECHO_COMPOSE_OPERATIONS, inScope } from "../catalog/scope.ts";
import { editPayload, readPayload } from "./payload.ts";
import type { Catalog, CatalogOp } from "../catalog/types.ts";
import { setCommandOp } from "../help/model.ts";
import { UsageError } from "../output/console.ts";
import { printResult } from "../output/result.ts";
import { readGlobalFlags } from "../session.ts";
import { fieldOptions, kebab, schemaType, type FieldOption, type JSONSchema } from "./fields.ts";
import type { FieldValues } from "./operation.ts";
import type { CommandOutput } from "./types.ts";

function optionParser(schema: JSONSchema): ((value: string) => unknown) | undefined {
  switch (schemaType(schema)) {
    case "boolean":
      return undefined;
    case "integer":
      return (value) => Number.parseInt(value, 10);
    case "number":
      return (value) => Number(value);
    case "array":
      return (value) => {
        const trimmed = value.trim();
        // Objects need JSON; plain values keep the comma shorthand.
        if (trimmed.startsWith("[")) {
          const parsed: unknown = JSON.parse(trimmed);
          if (!Array.isArray(parsed)) throw new UsageError(`不是 JSON 数组：${value}`);
          return parsed;
        }
        return trimmed
          .split(",")
          .map((item) => item.trim())
          .filter((item) => item !== "");
      };
    case "object":
      return (value) => JSON.parse(value) as unknown;
    default:
      return undefined;
  }
}

function defineField(command: Command, field: FieldOption): void {
  if (field.binary) {
    command.option(`--${field.flag} <path>`, `上传文件：路径，或 - 表示 stdin（字段 ${field.field}）`);
    return;
  }

  const kind = schemaType(field.schema);
  const flags = kind === "boolean" ? `--${field.flag}` : `--${field.flag} <value>`;
  const description = typeof field.schema["description"] === "string" ? field.schema["description"] : field.field;
  const parser = optionParser(field.schema);

  // Required fields are checked in the action, not with requiredOption: an
  // operation can be both a leaf and a namespace (`file.stream`), and
  // commander enforces a parent's required options even when a child runs.
  if (parser) command.option(flags, description, parser);
  else command.option(flags, description);
}

function emit(result: CommandOutput, flags: { json: boolean; yaml: boolean }): void {
  printResult(result, flags);
}

/** Repeatable option collector: `--tag a --tag b` becomes `["a", "b"]`. */
function collect(value: string, previous: string[] = []): string[] {
  return [...previous, value];
}

export interface DynamicOptions {
  run: (
    command: Command,
    op: CatalogOp,
    values: FieldValues,
    extra: {
    output?: string;
    compose?: boolean;
    payload?: Record<string, unknown>;
    tags?: string[];
    attach?: string[];
    attachStorage?: string;
  },
  ) => Promise<CommandOutput>;
  /** Overrides the curated operation allowlist (tests, future widening). */
  scope?: (id: string) => boolean;
}

/**
 * Turns every catalog operation into a command: `echo.create` becomes
 * `ech0 echo create`, and the operation's input schema becomes its flags.
 */
export function registerCatalogCommands(program: Command, catalog: Catalog, options: DynamicOptions): void {
  const groups = new Map<string, Command>();
  const ids = Object.keys(catalog.ops);

  const groupFor = (segments: string[]): Command => {
    let current = program;
    let key = "";
    for (const segment of segments) {
      key = key === "" ? segment : `${key}.${segment}`;
      const existing = current.commands.find((child) => child.name() === segment) ?? groups.get(key);
      let next = existing;
      if (!next) {
        // A group is a pure container: the help renderer counts its operations,
        // and the built-in `help` child would shadow `ech0 help <group>`.
        next = current.command(segment).addHelpCommand(false);
        groups.set(key, next);
      }
      current = next;
    }
    return current;
  };

  /** A callable prefix of `segments`, longest first (`comment.create` for `comment.create.integration`). */
  const callablePrefix = (segments: string[]): number => {
    for (let length = segments.length - 1; length > 0; length -= 1) {
      if (ids.includes(segments.slice(0, length).join("."))) return length;
    }
    return 0;
  };

  // Shallow ids first, so a group exists before anything nests under it.
  const visible = options.scope ?? inScope;
  const entries = Object.entries(catalog.ops)
    .filter(([id]) => visible(id))
    .sort(
    ([left], [right]) => left.split(".").length - right.split(".").length,
  );

  for (const [id, op] of entries) {
    const segments = id
      .split(".")
      .filter((segment) => segment !== "")
      .map(kebab);
    if (segments.length === 0) continue;

    // A callable command cannot also be a namespace: commander would parse the
    // parent's flags out of the child's argv. Such a child moves up a level and
    // joins its remaining segments (`comment.create.integration` becomes
    // `ech0 comment create-integration`), so a leaf never has subcommands.
    const conflict = callablePrefix(id.split("."));
    const parentSegments = conflict > 0 ? segments.slice(0, conflict - 1) : segments.slice(0, -1);
    const name = segments.slice(parentSegments.length).join("-");
    if (name === "") continue;

    const parent = parentSegments.length > 0 ? groupFor(parentSegments) : program;
    if (parent.commands.some((existing) => existing.name() === name)) continue;

    const command = parent.command(name).description(op.summary || id);
    const fields = fieldOptions(op);
    for (const field of fields) defineField(command, field);
    if ((op.transport?.response?.kind ?? op.kind) === "binary") {
      command.option("--output <path>", "写入文件；- 表示 stdout（非终端默认 stdout）");
    }
    const compose = ECHO_COMPOSE_OPERATIONS.has(id);
    if (compose) {
      command.option("--from <file>", "从 JSON 文件读取 payload（- 表示 stdin）");
      command.option("--edit", "在 $EDITOR 中编辑 payload 后提交");
      command.option("--tag <name>", "标签，可重复", collect, []);
      command.option("--attach <path>", "附件，可重复：先上传并挂到 Echo", collect, []);
      command.option("--attach-storage <type>", "附件存储类型：local / s3（默认 local；external 走 URL 注册）");
    }
    setCommandOp(command, op);

    command.action(async () => {
      const parsed = command.opts();
      const values: FieldValues = {};
      for (const field of fields) {
        const value = parsed[field.optionKey];
        if (value !== undefined) values[field.field] = value;
      }
      for (const field of fields) {
        if (field.required && values[field.field] === undefined) {
          throw new UsageError(`缺少必填参数 --${field.flag}`);
        }
      }
      let payload: Record<string, unknown> | undefined;
      if (parsed.from !== undefined && parsed.edit === true) {
        throw new UsageError("--from 与 --edit 不能同时使用");
      }
      if (parsed.edit === true) payload = await editPayload();
      else if (typeof parsed.from === "string") payload = await readPayload(parsed.from);

      const flags = readGlobalFlags(command);
      const result = await options.run(command, op, values, {
        output: typeof parsed.output === "string" ? parsed.output : undefined,
        compose,
        payload,
        tags: Array.isArray(parsed.tag) ? (parsed.tag as string[]) : undefined,
        attach: Array.isArray(parsed.attach) ? (parsed.attach as string[]) : undefined,
        attachStorage: typeof parsed.attachStorage === "string" ? parsed.attachStorage : undefined,
      });
      // Streaming and binary commands have already written their output.
      if (result.human !== "" || result.json !== "") emit(result, flags);
    });
  }
}
