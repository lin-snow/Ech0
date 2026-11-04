// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import { COMMAND_NAMES, ECHO_COMPOSE_OPERATIONS, inScope } from "../catalog/scope.ts";
import { confirmDestructive } from "./confirm.ts";
import { assertUploadStorage, categoryForFile, UPLOAD_OPERATION } from "./files.ts";
import { editPayload, readPayload } from "./payload.ts";
import type { Catalog, CatalogOp } from "../catalog/types.ts";
import { exampleFor, pathOf, setCommandExamples, setCommandHints, setCommandOp } from "../help/model.ts";
import { Ech0Error, UsageError } from "../output/console.ts";
import { printResult } from "../output/result.ts";
import { readGlobalFlags } from "../session.ts";
import {
  arrayItemType,
  enumValues,
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
  const choices = enumValues(schema);

  switch (schemaType(schema)) {
    case "boolean":
      return undefined;
    case "integer":
      return (value) => {
        const parsed = Number(value);
        if (value.trim() === "" || !Number.isInteger(parsed)) {
          throw new UsageError(`--${flag} 需要整数：${value}`);
        }
        return parsed;
      };
    case "number":
      return (value) => {
        const parsed = Number(value);
        if (value.trim() === "" || !Number.isFinite(parsed)) {
          throw new UsageError(`--${flag} 需要数字：${value}`);
        }
        return parsed;
      };
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
    case "string":
      if (choices.length === 0) return undefined;
      return (value) => {
        if (!choices.includes(value)) {
          throw new UsageError(`--${flag} 只支持 ${choices.join(" / ")}：${value}`);
        }
        return value;
      };
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
  const description = typeof field.schema["description"] === "string" ? field.schema["description"] : field.field;

  if (kind === "boolean") {
    // Tri-state: omitting the flag leaves the field unset, so filters can tell
    // "no preference" from false (`--no-private` = public only).
    command.option(`--${field.flag}`, description);
    command.option(`--no-${field.flag}`, "发送 false");
    return;
  }

  const flags = `--${field.flag} <value>`;
  const parser = optionParser(field.schema, field.flag);

  // Required fields are checked in the action, not with requiredOption: an
  // operation can be both a leaf and a namespace (`file.stream`), and
  // commander enforces a parent's required options even when a child runs.
  if (parser === undefined) {
    command.option(flags, description);
  } else if (kind === "array") {
    // Each occurrence parses on its own and concatenates, so repeated
    // `--tag-ids` behaves like the documented repeatable `--tag`.
    command.option(
      flags,
      description,
      (value: string, previous: unknown[] = []): unknown[] => {
        const parsed = parser(value);
        return [...previous, ...(Array.isArray(parsed) ? parsed : [parsed])];
      },
    );
  } else {
    command.option(flags, description, parser);
  }
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

/** Read-only Echo views; field semantics come from the server catalog. */
function echoQueryExamples(id: string): string[] | undefined {
  switch (id) {
    case "echo.list":
      return [
        "ech0 echo list --page 1 --page-size 20",
        "ech0 echo list --search daily --tag-ids <tag_id>",
        "ech0 echo list --sort-by fav_count --json",
      ];
    case "echo.get":
      return ["ech0 echo get --id <echo_id>", "ech0 echo get --id <echo_id> --json"];
    case "echo.hot":
      return ["ech0 echo hot --limit 10"];
    case "echo.today":
      return ["ech0 echo today"];
    case "echo.random":
      return ["ech0 echo random"];
    case "echo.onthisday":
      return ["ech0 echo onthisday"];
    default:
      return undefined;
  }
}

function echoQueryHints(id: string): string[] {
  if (id === "echo.list") {
    return [
      "标签 ID 来自 ech0 tag list；多个 ID 命中任意一个即可",
      "私密可见性由账号决定：管理员可用 --private / --no-private 过滤，其他账号只看公开",
    ];
  }
  if (id === "echo.hot" || id === "echo.today" || id === "echo.random" || id === "echo.onthisday") {
    return ["只返回公开 Echo；管理员 token 会包含私密"];
  }
  return [];
}

/** Server schemas stay generic; these distinctions are command-level UX. */
function fileHints(id: string): string[] {
  if (id === "file.upload") return ["只上传本地字节；已有 URL 的资源用 ech0 file external 登记，不下载不复制"];
  if (id === "file.external") return ["只登记 URL，不上传字节；本地文件用 ech0 file upload"];
  if (id === "file.list") return ["--storage-type 取 local / s3 / object；不传时包含 external 记录"];
  if (id === "file.tree") return ["--storage-type 取 local / s3 / object，需与实例存储配置一致"];
  return [];
}

/**
 * Comment commands cover the full lifecycle: publish through the trusted
 * integration route, read approved comments, and review (list / status /
 * delete) through the panel endpoints.
 */
function commentExamples(id: string): string[] | undefined {
  if (id === "comment.create.integration") {
    return [
      'ech0 comment create --echo-id <echo_id> --content "写得好"',
      'ech0 comment create --echo-id <echo_id> --content "来自流水线" --nickname "CI Bot"',
    ];
  }
  if (id === "comment.list.by.echo") return ["ech0 comment list --echo-id <echo_id>"];
  if (id === "comment.panel.list") {
    return [
      "ech0 comment panel list --status pending",
      "ech0 comment panel list --keyword 广告 --page 1 --page-size 20",
    ];
  }
  if (id === "comment.panel.status") {
    return [
      "ech0 comment panel status --id <comment_id> --status approved",
      "ech0 comment panel status --id <comment_id> --status rejected",
    ];
  }
  if (id === "comment.panel.delete") return ["ech0 comment panel delete --id <comment_id> --yes"];
  return undefined;
}

function commentHints(id: string): string[] {
  if (id === "comment.create.integration") {
    return [
      "走可信集成入口 /comments/integration：不需要 form token 或验证码，评论标记为 source=integration",
      "默认昵称是 Integration；人类评论建议显式传 --nickname",
      "实例开启审核时返回 pending；用 ech0 comment panel list --status pending 处理",
    ];
  }
  if (id === "comment.list.by.echo") {
    return ["只返回 approved 评论；pending / rejected 用 ech0 comment panel list 查看"];
  }
  if (id === "comment.panel.list") {
    return [
      "--status 取 pending / approved / rejected；--echo-id 只看某条 Echo 的评论",
      "需要 comment:review scope；审核列表与面板 /panel/comments 同一份数据",
    ];
  }
  if (id === "comment.panel.status") {
    return [
      "--status 取 pending / approved / rejected；只改状态，不动正文",
      "需要 comment:review scope；审核前先用 ech0 comment panel list --status pending 拿 id",
    ];
  }
  if (id === "comment.panel.delete") {
    return [
      "删除评论本身，不影响所属 Echo；需要 comment:delete scope（与 comment:review 分开授予）",
      "不可恢复；脚本输出必须显式传 --yes",
    ];
  }
  return [];
}

/** Connect commands: peer list/info/add/delete plus the read-only health probe. */
function connectExamples(id: string): string[] | undefined {
  if (id === "connect.list") return ["ech0 connect list", "ech0 connect list --json"];
  if (id === "connect.info") return ["ech0 connect info --json"];
  if (id === "connect.health") return ["ech0 connect health", "ech0 connect health --json"];
  if (id === "connect.add") return ["ech0 connect add --connect-url https://peer.example"];
  if (id === "connect.delete") return ["ech0 connect delete --id <uuid> --yes"];
  return undefined;
}

function connectHints(id: string): string[] {
  if (id === "connect.list") {
    return ["只返回本地保存的 {id, connect_url}；id 用于 ech0 connect delete"];
  }
  if (id === "connect.info") {
    return ["对端公开卡片（名称 / logo / 帖子统计 / 版本）；服务端缓存 30 分钟，首次可能较慢"];
  }
  if (id === "connect.health") {
    return [
      "对每个已添加连接请求 <url>/api/connect，返回 status / version",
      "需要 connect:read scope；只读探测，不会修改连接",
    ];
  }
  if (id === "connect.add") {
    return [
      "--connect-url 传对端实例根地址，服务端会请求 <url>/api/connect",
      "需要管理员账号；重复 URL 会被拒绝",
    ];
  }
  if (id === "connect.delete") return ["需要管理员账号；只删除本地连接记录，对端实例不受影响"];
  return [];
}

/** Tags are auto-created by `echo create/update --tag`; the CLI only needs lookup and cleanup. */
function tagExamples(id: string): string[] | undefined {
  if (id === "tag.list") return ["ech0 tag list", "ech0 tag list --json"];
  if (id === "tag.delete") return ["ech0 tag delete --id <tag_id> --yes"];
  return undefined;
}

function tagHints(id: string): string[] {
  if (id === "tag.list") return ["返回 id / name / usage_count；id 用于 ech0 echo list --tag-ids"];
  if (id === "tag.delete") return ["需要管理员账号；会同时移除所有 Echo 上的该标签"];
  return [];
}

/** Webhooks mirror the MCP tools: list, create, update, delete and test. */
function webhookExamples(id: string): string[] | undefined {
  if (id === "webhook.list") return ["ech0 webhook list", "ech0 webhook list --json"];
  if (id === "webhook.create") {
    return [
      "ech0 webhook create --name ci --url https://example.com/hook --is-active",
      "ech0 webhook create --name ci --url https://example.com/hook --secret <secret> --is-active",
    ];
  }
  if (id === "webhook.update") {
    return ["ech0 webhook update --id <webhook_id> --name ci --url https://example.com/hook --is-active"];
  }
  if (id === "webhook.delete") return ["ech0 webhook delete --id <webhook_id> --yes"];
  if (id === "webhook.test") return ["ech0 webhook test --id <webhook_id>"];
  return undefined;
}

function webhookHints(id: string): string[] {
  if (id === "webhook.list") {
    return ["返回 name / url / is_active / last_status / last_trigger；secret 不会返回"];
  }
  if (id === "webhook.create") {
    return [
      "需要管理员账号；--url 必须是可访问的 HTTP(S) 地址",
      "省略 --is-active 时服务端默认启用；--secret 用于 X-Ech0-Signature 签名",
    ];
  }
  if (id === "webhook.update") {
    return ["需要管理员账号；整体覆盖，省略 --is-active 会停用、省略 --secret 会清空签名密钥"];
  }
  if (id === "webhook.delete") return ["需要管理员账号；只删除本地 Webhook 配置"];
  if (id === "webhook.test") return ["需要管理员账号；向该 Webhook 的 URL 发送一条测试事件"];
  return [];
}

/** Server rejected the input: point at the descriptor that lists its fields. */
function withCommandHint(error: unknown, command: Command): unknown {
  if (!(error instanceof Ech0Error) || error.details.hint !== undefined) return error;
  if (pathOf(command).join(" ") === "webhook test" && error.details.status === 400) {
    return new Ech0Error(error.message, {
      ...error.details,
      hint: "Webhook 投递失败：检查目标 URL 是否可达、是否被实例的 egress 策略允许",
    });
  }
  if (error.details.status === 400 || error.details.status === 422) {
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
    // session_only routes 403 for every access token, so a CLI command for
    // them could never succeed.
    .filter(([id, op]) => !op.session_only && visible(id))
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
      command.option("--output <path>", "写入文件；省略时按响应文件名落到当前目录；- 表示 stdout（非终端默认 stdout）");
    }
    if (id === UPLOAD_OPERATION) {
      command.option("--file-name <name>", "上传时使用的文件名（stdin 上传务必带扩展名）");
    }
    if (op.effect === "destructive") {
      command.option("-y, --yes", "跳过危险操作确认（非交互环境与 --json/--yaml 必须显式传入）");
    }
    const compose = ECHO_COMPOSE_OPERATIONS.has(id);
    if (compose) {
      command.option("--from <file>", "从 JSON 文件读取 payload（- 表示 stdin）");
      command.option("--edit", "在 $EDITOR 中编辑 payload 后提交");
      command.option("--tag <name>", "标签，可重复；与 --tags 合并", collect, []);
      command.option("--file-id <id>", "已上传的文件 id，可重复；与 --echo-files 合并", collect, []);
      setCommandExamples(command, composeExamples(id));
    } else {
      const examples =
        echoQueryExamples(id) ??
        commentExamples(id) ??
        connectExamples(id) ??
        tagExamples(id) ??
        webhookExamples(id);
      if (examples !== undefined) setCommandExamples(command, examples);
    }
    const hints = [
      ...(compose ? composeHints(id, op) : []),
      ...echoQueryHints(id),
      ...fileHints(id),
      ...commentHints(id),
      ...connectHints(id),
      ...tagHints(id),
      ...webhookHints(id),
    ];
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

      const flags = readGlobalFlags(command);
      if (op.effect === "destructive") {
        await confirmDestructive(op, values, {
          yes: parsed.yes === true,
          json: flags.json,
          yaml: flags.yaml,
        });
      }

      let payload: Record<string, unknown> | undefined;
      if (parsed.from !== undefined && parsed.edit === true) {
        throw new UsageError("--from 与 --edit 不能同时使用");
      }
      if (parsed.edit === true) payload = await editPayload();
      else if (typeof parsed.from === "string") payload = await readPayload(parsed.from);

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
