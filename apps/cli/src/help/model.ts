// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Command } from "commander";

import type { CatalogOp } from "../catalog/types.ts";
import { enumValues, fieldOptions, flagTerm, isFileField, type FieldOption } from "../commands/fields.ts";

/** What the root command could learn about the server before argv was parsed. */
export interface HelpContext {
  server?: string;
  loaded: boolean;
  failure?: string;
  /** Operations the catalog publishes, to explain why fewer commands exist. */
  ops?: number;
  /** False when no token is available; help then points at `auth login`. */
  authenticated?: boolean;
}

/** One aligned table row: a term and what it does, plus optional dim tags. */
export interface HelpRow {
  term: string;
  description: string;
  tags: string[];
}

/** One searchable/listed command, id relative to the scope it was listed in. */
export interface CommandRow {
  id: string;
  summary: string;
  tags: string[];
}

export interface GroupRow {
  name: string;
  count: number;
  preview: string;
}

export interface HelpModel {
  root: boolean;
  group: boolean;
  path: string[];
  usage: string;
  summary: string;
  /** `POST /api/echo · 写入 · scope: echo:write` for catalog operations. */
  facet?: string;
  fields: HelpRow[];
  options: HelpRow[];
  commands: CommandRow[];
  /** Root only: local commands that are not catalog-driven. */
  locals: CommandRow[];
  /** Root only: top-level namespaces that come from the catalog. */
  groups: GroupRow[];
  /** Root only. */
  server?: HelpContext;
  examples: string[];
  /** Actionable next steps for this command, from its catalog metadata. */
  hints: string[];
  /** Compact list of globally accepted flags, on every non-root command. */
  globalHint?: string;
  footer?: string;
}

export interface HelpEntry extends CommandRow {
  command: Command;
  /** Everything the entry declares, for keyword search. */
  haystack: string;
}

const OP = new WeakMap<Command, CatalogOp>();
const EXAMPLES = new WeakMap<Command, string[]>();
const HINTS = new WeakMap<Command, string[]>();
const CONTEXT = new WeakMap<Command, HelpContext>();

/** Remembers the catalog operation a dynamic command wraps. */
export function setCommandOp(command: Command, op: CatalogOp): void {
  OP.set(command, op);
}

export function commandOp(command: Command): CatalogOp | undefined {
  return OP.get(command);
}

/** Examples are hand-written for local commands and generated for catalog ones. */
export function setCommandExamples(command: Command, examples: string[]): void {
  EXAMPLES.set(command, examples);
}

/** CLI-only advice a catalog operation cannot express (compose flags, login modes). */
export function setCommandHints(command: Command, hints: string[]): void {
  HINTS.set(command, hints);
}

export function setHelpContext(root: Command, context: HelpContext): void {
  CONTEXT.set(root, context);
}

export function rootOf(command: Command): Command {
  let current = command;
  while (current.parent) current = current.parent;
  return current;
}

export function pathOf(command: Command): string[] {
  const path: string[] = [];
  for (let current: Command | undefined = command; current?.parent; current = current.parent) {
    path.unshift(current.name());
  }
  return path;
}

export function nodeAt(root: Command, path: readonly string[]): Command | undefined {
  let current: Command | undefined = root;
  for (const name of path) {
    current = current?.commands.find(
      (child) => child.name() === name || child.aliases().includes(name),
    );
    if (current === undefined) return undefined;
  }
  return current;
}

/** Every listed entry under `base`, the base itself excluded. An entry wraps an
 * operation, or is a leaf; a namespace with neither is only a container. */
export function entriesUnder(base: Command): Command[] {
  const found: Command[] = [];
  const walk = (node: Command): void => {
    for (const child of node.commands) {
      if (commandOp(child) !== undefined || child.commands.length === 0) found.push(child);
      walk(child);
    }
  };
  walk(base);
  return found;
}

/** The words you actually type for this command, relative to the listing scope. */
function relativeId(base: Command, command: Command): string {
  const full = pathOf(command);
  const prefix = pathOf(base);
  return full.slice(prefix.length).join(" ");
}

function tagsOf(op: CatalogOp | undefined): string[] {
  if (op === undefined) return [];
  const tags: string[] = [];
  const kind = op.transport?.response?.kind ?? op.kind;
  if (kind === "stream") tags.push("stream");
  else if (kind === "binary") tags.push("binary");
  else if (kind === "websocket") tags.push("websocket");
  if (op.deprecated) tags.push("deprecated");
  return tags;
}

function commandRow(base: Command, command: Command): CommandRow {
  const op = commandOp(command);
  return {
    id: relativeId(base, command),
    summary: op?.summary || command.description() || command.name(),
    tags: tagsOf(op),
  };
}

/** Everything a search can match: id, summary, field names, field descriptions. */
function haystackOf(command: Command): string {
  const op = commandOp(command);
  if (op === undefined) return `${pathOf(command).join(" ")} ${command.description()}`;
  const fields = fieldOptions(op)
    .map((field) => {
      const description = typeof field.schema["description"] === "string" ? field.schema["description"] : "";
      return `${field.field} ${field.flag} ${field.optionKey} ${description}`;
    })
    .join(" ");
  return `${pathOf(command).join(" ")} ${op.summary} ${op.method} ${op.path} ${fields}`;
}

/** Every node in the tree as a dotted id: namespaces included, the root excluded. */
export function commandIds(root: Command): string[] {
  const ids: string[] = [];
  const walk = (node: Command): void => {
    for (const child of node.commands) {
      ids.push(pathOf(child).join("."));
      walk(child);
    }
  };
  walk(root);
  return ids;
}

export function collectEntries(root: Command, base: Command = root): HelpEntry[] {
  return entriesUnder(base).map((command) => ({
    ...commandRow(base, command),
    command,
    haystack: haystackOf(command),
  }));
}

function argToken(command: Command): string[] {
  return command.registeredArguments.map((argument) =>
    argument.variadic
      ? `[${argument.name()}...]`
      : argument.required
        ? `<${argument.name()}>`
        : `[${argument.name()}]`,
  );
}

function usageOf(command: Command): string {
  const parts = ["ech0", ...pathOf(command)];
  // The root's operands exist only to forward unknown words into `help`.
  if (command.parent !== null) parts.push(...argToken(command));
  parts.push("[options]");
  if (command.commands.length > 0) parts.push("[command]");
  return parts.join(" ");
}

function describeField(field: FieldOption): string {
  const parts: string[] = [];
  const description = field.schema["description"];
  if (typeof description === "string" && description.trim() !== "") {
    parts.push(description.trim().split("\n")[0] ?? "");
  }
  const choices = enumValues(field.schema);
  if (choices.length > 0) parts.push(`可选：${choices.join("、")}`);
  const fallback = field.schema["default"];
  if (fallback !== undefined && fallback !== null) parts.push(`默认：${JSON.stringify(fallback)}`);
  if (isFileField(field.schema)) parts.push("文件路径，或 - 读取 stdin");
  return parts.join("；");
}

function fieldRows(op: CatalogOp): HelpRow[] {
  return fieldOptions(op).map((field) => {
    const location = op.bind[field.field];
    const tags: string[] = [];
    if (field.required) tags.push("必填");
    if (location !== undefined && location !== "body") tags.push(location);
    return { term: flagTerm(field), description: describeField(field), tags };
  });
}

function optionRows(command: Command, fields: readonly HelpRow[] = []): HelpRow[] {
  const fieldFlags = new Set(fields.map((field) => field.term.split(" ")[0]));
  const rows = command.options
    .filter((option) => !fieldFlags.has(option.long ?? option.short ?? ""))
    .map((option) => {
      const parts: string[] = [];
      if (option.description) parts.push(option.description);
      if (Array.isArray(option.argChoices) && option.argChoices.length > 0) {
        parts.push(`可选：${option.argChoices.join("、")}`);
      }
      const fallback = option.defaultValue;
      if (fallback !== undefined && fallback !== null && fallback !== false) {
        parts.push(`默认：${JSON.stringify(fallback)}`);
      }
      return { term: option.flags, description: parts.join("；"), tags: [] };
    });
  // Commander keeps its own help flag outside `command.options`.
  rows.push({ term: "-h, --help", description: "显示帮助", tags: [] });
  return rows;
}

const EFFECT_LABELS: Record<string, string> = {
  read: "只读",
  write: "写入",
  destructive: "危险操作",
};

/** The one-line server card: where it goes, what it needs, what it returns. */
function facetOf(op: CatalogOp): string {
  const mount = op.mount === "ws" ? "/ws" : "/api";
  const parts = [`${op.method} ${mount}${op.path}`];
  parts.push(EFFECT_LABELS[op.effect] ?? op.effect);
  if (op.scopes.length > 0) parts.push(`scope: ${op.scopes.join(", ")}`);

  const request = op.transport?.request;
  if (request?.content_type === "multipart/form-data") parts.push("multipart 上传");
  if (request?.auth === "query") parts.push("token 走查询参数");

  const kind = op.transport?.response?.kind ?? op.kind;
  if (kind === "stream") {
    const events = op.transport?.response?.events?.map((event) => event.name) ?? [];
    parts.push(events.length > 0 ? `SSE 流（事件：${events.join(", ")}）` : "SSE 流");
  } else if (kind === "binary") parts.push("二进制响应");
  else if (kind === "websocket") parts.push("WebSocket");

  if (op.deprecated) parts.push("已弃用");
  return parts.join(" · ");
}

/** The runnable line for a catalog operation: every required field, as itself. */
export function exampleFor(command: Command): string | undefined {
  const op = commandOp(command);
  if (op === undefined) return EXAMPLES.get(command)?.[0];
  const fields = fieldOptions(op).filter(
    (field) => field.required || isFileField(field.schema),
  );
  const parts = ["ech0", ...pathOf(command), ...fields.map(flagTerm)];
  return parts.join(" ");
}

/** What to do next: scope, transport and auth advice the catalog implies. */
function hintsOf(command: Command, op: CatalogOp | undefined): string[] {
  const hints: string[] = [...(HINTS.get(command) ?? [])];
  const context = CONTEXT.get(rootOf(command));

  if (op === undefined) {
    if (context?.loaded === true && context.authenticated === false) {
      hints.push("尚未登录：运行 ech0 auth login --server <url> --device；catalog 可匿名读取，执行操作需要 token");
    }
    return hints;
  }

  if (op.scopes.length > 0) {
    hints.push(`需要 scope：${op.scopes.join(", ")}；403 表示 token 缺少该 scope，请在面板重新铸造`);
  }
  const kind = op.transport?.response?.kind ?? op.kind;
  if (kind === "stream") hints.push("流式输出：Ctrl-C 中断；--json 时每个事件输出一行 JSON");
  if (kind === "binary") hints.push("二进制响应：--output <path> 写入文件；省略时 TTY 按响应文件名落盘、非 TTY 写 stdout；- 表示 stdout");
  if (kind === "websocket") hints.push("WebSocket 订阅：Ctrl-C 断开");
  if (Object.values(op.bind).includes("form")) {
    hints.push("上传字段：文件用 --<字段> <path>，- 表示从 stdin 读取");
  }
  if (op.effect === "destructive") hints.push("危险操作：确认 --profile / --server 指向目标实例后再执行");
  if (op.deprecated) hints.push("已弃用：服务端可能在下个版本移除该操作");
  return hints;
}

function examplesOf(command: Command): string[] {
  const explicit = EXAMPLES.get(command);
  if (explicit !== undefined) return explicit;
  const generated = exampleFor(command);
  return generated === undefined ? [] : [generated];
}

function globalHintOf(root: Command): string | undefined {
  const names = root.options
    .filter((option) => option.long !== "--help" && option.long !== "--version")
    .map((option) => option.flags);
  if (names.length === 0) return undefined;
  return `全局选项（任意位置）：${names.join(" ")}（详见 ech0 --help）`;
}

/** Top-level namespaces that exist only because the catalog published them. */
function serverGroups(root: Command): GroupRow[] {
  const rows: GroupRow[] = [];
  for (const child of root.commands) {
    if (commandOp(child) === undefined && entriesUnder(child).every((entry) => commandOp(entry) === undefined)) {
      continue;
    }
    const entries = collectEntries(root, child);
    if (entries.length === 0) continue;
    const preview = entries.slice(0, 4).map((entry) => entry.id);
    rows.push({
      name: child.name(),
      count: entries.length,
      preview: `${preview.join(", ")}${entries.length > preview.length ? ", …" : ""}`,
    });
  }
  return rows.sort((left, right) => left.name.localeCompare(right.name));
}

function localRows(root: Command): CommandRow[] {
  const dynamic = new Set(serverGroups(root).map((group) => group.name));
  return root.commands
    .filter((child) => !dynamic.has(child.name()))
    .map((child) => ({
      id: child.name(),
      summary: child.description(),
      tags: [],
    }))
    .sort((left, right) => left.id.localeCompare(right.id));
}

/** Builds everything one help rendering needs, from the live command tree. */
export function helpModel(command: Command): HelpModel {
  const root = rootOf(command);
  const op = commandOp(command);
  const isRoot = command === root;
  const isGroup = command.commands.length > 0;
  const entries = isGroup ? collectEntries(command, command) : [];

  const model: HelpModel = {
    root: isRoot,
    group: !isRoot && isGroup && op === undefined,
    path: pathOf(command),
    usage: usageOf(command),
    summary: op?.summary ?? command.description() ?? "",
    fields: op === undefined ? [] : fieldRows(op),
    options: isGroup && op === undefined && !isRoot ? [] : optionRows(command, op === undefined ? [] : fieldRows(op)),
    commands: isRoot ? [] : entries.map(({ id, summary, tags }) => ({ id, summary, tags })),
    locals: isRoot ? localRows(root) : [],
    groups: isRoot ? serverGroups(root) : [],
    examples: examplesOf(command),
    hints: hintsOf(command, op),
  };

  if (op !== undefined) model.facet = facetOf(op);
  if (isRoot) {
    model.server = CONTEXT.get(root);
  } else if (!model.group) {
    model.globalHint = globalHintOf(root);
  }
  if (isGroup && op === undefined && !isRoot) {
    if (model.summary === "") model.summary = `服务端命令组（${entries.length} 个操作）`;
    model.footer = `运行 ech0 help <关键词> 搜索命令；ech0 help ${pathOf(command).join(" ")} 与本页等价。`;
  }
  if (model.examples.some((example) => example.includes("<"))) {
    model.footer = "示例中的 <字段名> 是占位符，按实际值替换。";
  }
  return model;
}
