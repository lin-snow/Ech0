// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import { COMMAND_NAMES, ECHO_COMPOSE_OPERATIONS, inScope } from "../catalog/scope.ts";
import { assertUploadStorage, categoryForFile, UPLOAD_OPERATION } from "./files.ts";
import { editPayload, readPayload } from "./payload.ts";
import type { Catalog, CatalogOp } from "../catalog/types.ts";
import { exampleFor, pathOf, setCommandExamples, setCommandHints, setCommandOp } from "../help/model.ts";
import { Ech0Error, UsageError } from "../output/console.ts";
import { printResult } from "../output/result.ts";
import { readGlobalFlags } from "../session.ts";
import {
  arrayItemType,
  fieldOptions,
  kebab,
  schemaType,
  type FieldOption,
  type JSONSchema,
} from "./fields.ts";
import type { FieldValues } from "./operation.ts";
import type { CommandOutput } from "./types.ts";

function parseJSON(flag: string, value: string): unknown {
  try {
    return JSON.parse(value) as unknown;
  } catch (error) {
    throw new UsageError(`--${flag} 不是合法 JSON：${(error as Error).message}`);
  }
}

function optionParser(schema: JSONSchema, flag: string): ((value: string) => unknown) | undefined {
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
        // Object elements need JSON; plain values keep the comma shorthand.
        if (trimmed.startsWith("[")) {
          const parsed = parseJSON(flag, trimmed);
          if (!Array.isArray(parsed)) throw new UsageError(`--${flag} 不是 JSON 数组：${value}`);
          return parsed;
        }
        if (arrayItemType(schema) !== "string") {
          throw new UsageError(`--${flag} 的元素是对象，需要 JSON 数组：${value}`);
        }
        return trimmed
          .split(",")
          .map((item) => item.trim())
          .filter((item) => item !== "");
      };
    case "object":
      return (value) => parseJSON(flag, value);
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
  const parser = optionParser(field.schema, field.flag);

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

function composeExamples(id: string): string[] {
  const update = id === "echo.update";
  if (update) {
    return [
      "ech0 echo update --id <echo_id> --content \"改过的正文\"",
      "ech0 echo update --id <echo_id> --content \"改过的正文\" --tag daily --tag work",
      "ech0 echo update --id <echo_id> --from echo.json",
    ];
  }
  return [
    "ech0 echo create --content \"今天的想法\" --tag daily",
    "ech0 file upload --file photo.png",
    "ech0 echo create --content \"配图\" --file-id <file_id>",
    "ech0 echo create --from echo.json",
  ];
}

/** Compose flags are CLI-only, so their semantics cannot come from the catalog. */
function composeHints(id: string, op: CatalogOp): string[] {
  const fields = new Set(Object.keys((op.input as { properties?: object }).properties ?? {}));
  const hints: string[] = [];

  if (id === "echo.update") {
    hints.push("更新是整体覆盖：--content / --echo-files / --extension / --tags 没传的部分会被清空；先用 ech0 echo get --id <id> 取回再改");
    hints.push("--id 必填，且需要管理员账号");
  }
  if (fields.has("tags")) {
    hints.push('--tag 追加标签（可重复）；--tags 接收 JSON 数组：[{"name":"daily"}]');
  }
  if (fields.has("echo_files")) {
    hints.push('--file-id 引用已上传文件（先运行 ech0 file upload）；--echo-files 接收 JSON：[{"file_id":"<id>","sort_order":0}]');
  }
  if (fields.has("extension")) {
    hints.push('--extension 接收 JSON 对象：{"type":"weather","payload":{...}}，用于渲染扩展卡片');
  }
  hints.push("--from/--edit 提交完整 payload：同名逐项参数被 payload 覆盖，tags 与 echo_files 累加");
  return hints;
}

/** Server schemas stay generic; these distinctions are command-level UX. */
function fileHints(id: string): string[] {
  if (id === "file.upload") return ["只上传本地字节；已有 URL 的资源用 ech0 file external 登记，不下载不复制"];
  if (id === "file.external") return ["只登记 URL，不上传字节；本地文件用 ech0 file upload"];
  if (id === "file.list") return ["--storage-type 取 local / s3 / object；不传时包含 external 记录"];
  if (id === "file.tree") return ["--storage-type 取 local / s3 / object，需与实例存储配置一致"];
  return [];
}

/** Server rejected the input: point at the descriptor that lists its fields. */
function withCommandHint(error: unknown, command: Command): unknown {
  if (
    error instanceof Ech0Error &&
    error.details.hint === undefined &&
    (error.details.status === 400 || error.details.status === 422)
  ) {
    return new Ech0Error(error.message, {
      ...error.details,
      hint: `请求参数未被服务端接受：运行 \`ech0 ${pathOf(command).join(" ")} --help\` 核对字段与取值`,
    });
  }
  return error;
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
    fileIds?: string[];
    uploadFileName?: string;
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
    const commandPath = COMMAND_NAMES[id] ?? id;
    const pathSegments = commandPath.split(".").filter((segment) => segment !== "");
    const segments = pathSegments.map(kebab);
    if (segments.length === 0) continue;

    // A callable command cannot also be a namespace: commander would parse the
    // parent's flags out of the child's argv. Such a child moves up a level and
    // joins its remaining segments (`comment.create.integration` becomes
    // `ech0 comment create-integration`), so a leaf never has subcommands.
    const conflict = callablePrefix(pathSegments);
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
    if (id === UPLOAD_OPERATION) {
      command.option("--file-name <name>", "上传时使用的文件名（stdin 上传务必带扩展名）");
    }
    const compose = ECHO_COMPOSE_OPERATIONS.has(id);
    if (compose) {
      command.option("--from <file>", "从 JSON 文件读取 payload（- 表示 stdin）");
      command.option("--edit", "在 $EDITOR 中编辑 payload 后提交");
      command.option("--tag <name>", "标签，可重复；与 --tags 合并", collect, []);
      command.option("--file-id <id>", "已上传的文件 id，可重复；与 --echo-files 合并", collect, []);
      setCommandExamples(command, composeExamples(id));
    }
    const hints = [...(compose ? composeHints(id, op) : []), ...fileHints(id)];
    if (hints.length > 0) setCommandHints(command, hints);
    setCommandOp(command, op);

    command.action(async () => {
      const parsed = command.opts();
      const values: FieldValues = {};
      for (const field of fields) {
        const value = parsed[field.optionKey];
        if (value !== undefined) values[field.field] = value;
      }
      // The upload API infers a category from the filename when none is given;
      // do the same here so `ech0 file upload --file a.png` just works.
      if (id === UPLOAD_OPERATION) {
        if (values.category === undefined && typeof values.file === "string") {
          values.category = categoryForFile(values.file);
        }
        if (typeof values.storage_type === "string") {
          assertUploadStorage(values.storage_type);
        }
      }
      const missing = fields
        .filter((field) => field.required && values[field.field] === undefined)
        .map((field) => `--${field.flag}`);
      if (missing.length > 0) {
        const usage = exampleFor(command);
        throw new UsageError(
          `缺少必填参数 ${missing.join(", ")}`,
          usage === undefined
            ? undefined
            : `用法：${usage}（运行 \`ech0 help ${pathOf(command).join(" ")}\` 查看完整说明）`,
        );
      }
      let payload: Record<string, unknown> | undefined;
      if (parsed.from !== undefined && parsed.edit === true) {
        throw new UsageError("--from 与 --edit 不能同时使用");
      }
      if (parsed.edit === true) payload = await editPayload();
      else if (typeof parsed.from === "string") payload = await readPayload(parsed.from);

      const flags = readGlobalFlags(command);
      const result = await options
        .run(command, op, values, {
          output: typeof parsed.output === "string" ? parsed.output : undefined,
          compose,
          payload,
          tags: Array.isArray(parsed.tag) ? (parsed.tag as string[]) : undefined,
          fileIds: Array.isArray(parsed.fileId) ? (parsed.fileId as string[]) : undefined,
          uploadFileName: typeof parsed.fileName === "string" ? parsed.fileName : undefined,
        })
        .catch((error: unknown) => {
          throw withCommandHint(error, command);
        });
      // Streaming and binary commands have already written their output.
      if (result.human !== "" || result.json !== "") emit(result, flags);
    });
  }
}
