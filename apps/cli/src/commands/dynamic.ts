// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import type { Catalog, CatalogOp } from "../catalog/types.ts";
import { colors, print } from "../output/console.ts";
import { readGlobalFlags } from "../session.ts";
import type { FieldValues } from "./operation.ts";
import type { CommandOutput } from "./types.ts";

type JSONSchema = Record<string, unknown>;

/** Flags the CLI owns; an operation field with the same name becomes `--input-<name>`. */
const RESERVED_FLAGS: Record<string, true> = {
  json: true,
  profile: true,
  server: true,
  token: true,
  lang: true,
  color: true,
  help: true,
  version: true,
};

interface FieldOption {
  field: string;
  /** camelCase key commander stores the parsed value under. */
  optionKey: string;
  flag: string;
  schema: JSONSchema;
  required: boolean;
}

function schemaType(schema: JSONSchema): string {
  const type = schema["type"];
  if (typeof type === "string") return type;
  if (Array.isArray(type)) {
    for (const candidate of type) {
      if (typeof candidate === "string" && candidate !== "null") return candidate;
    }
  }
  return "string";
}

function kebab(name: string): string {
  return name
    .replace(/_/g, "-")
    .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
    .toLowerCase();
}

function camel(name: string): string {
  return name.replace(/-([a-z0-9])/g, (_, char: string) => char.toUpperCase());
}

function fieldOptions(op: CatalogOp): FieldOption[] {
  const input = op.input as { properties?: Record<string, JSONSchema>; required?: unknown };
  const properties = input.properties ?? {};
  const required = Array.isArray(input.required)
    ? new Set(input.required.filter((name): name is string => typeof name === "string"))
    : new Set<string>();

  return Object.entries(properties).map(([field, rawSchema]) => {
    const base = kebab(field);
    const flag = RESERVED_FLAGS[base] ? `input-${base}` : base;
    return {
      field,
      flag,
      optionKey: camel(flag),
      schema: rawSchema ?? {},
      required: required.has(field),
    };
  });
}

function optionParser(schema: JSONSchema): ((value: string) => unknown) | undefined {
  switch (schemaType(schema)) {
    case "boolean":
      return undefined;
    case "integer":
      return (value) => Number.parseInt(value, 10);
    case "number":
      return (value) => Number(value);
    case "array":
      return (value) =>
        value
          .split(",")
          .map((item) => item.trim())
          .filter((item) => item !== "");
    case "object":
      return (value) => JSON.parse(value) as unknown;
    default:
      return undefined;
  }
}

function defineField(command: Command, field: FieldOption): void {
  const kind = schemaType(field.schema);
  const flags = kind === "boolean" ? `--${field.flag}` : `--${field.flag} <value>`;
  const description = typeof field.schema["description"] === "string" ? field.schema["description"] : field.field;
  const parser = optionParser(field.schema);

  if (field.required) {
    if (parser) command.requiredOption(flags, description, parser);
    else command.requiredOption(flags, description);
    return;
  }
  if (parser) command.option(flags, description, parser);
  else command.option(flags, description);
}

function metaLine(op: CatalogOp): string {
  const { dim } = colors();
  const parts = [`${op.method} ${op.path}`];
  if (op.scopes.length > 0) parts.push(`scopes: ${op.scopes.join(", ")}`);
  if (op.deprecated) parts.push("deprecated");
  return dim(parts.join(" · "));
}

function emit(result: CommandOutput, json: boolean): void {
  print(json ? result.json : result.human);
}

export interface DynamicOptions {
  run: (command: Command, op: CatalogOp, values: FieldValues) => Promise<CommandOutput>;
}

/**
 * Turns every catalog operation into a command: `echo.create` becomes
 * `ech0 echo create`, and the operation's input schema becomes its flags.
 */
export function registerCatalogCommands(program: Command, catalog: Catalog, options: DynamicOptions): void {
  const groups = new Map<string, Command>();

  // A leaf and a namespace may share a name (`comment.create` and
  // `comment.create.integration`), so an existing command is reused as a group.
  const groupFor = (segments: string[]): Command => {
    let current = program;
    let key = "";
    for (const segment of segments) {
      key = key === "" ? segment : `${key}.${segment}`;
      const existing = current.commands.find((child) => child.name() === segment) ?? groups.get(key);
      let next = existing;
      if (!next) {
        next = current.command(segment).description(`${segment} 相关操作`);
        groups.set(key, next);
      }
      current = next;
    }
    return current;
  };

  // Shallow ids first: a leaf must exist before a deeper id nests under it.
  const entries = Object.entries(catalog.ops).sort(
    ([left], [right]) => left.split(".").length - right.split(".").length,
  );

  for (const [id, op] of entries) {
    const segments = id.split(".").filter((segment) => segment !== "");
    const name = segments.pop();
    if (name === undefined) continue;

    const parent = groupFor(segments);
    if (parent.commands.some((existing) => existing.name() === name)) continue;

    const command = parent.command(name).description(op.summary || id);
    const fields = fieldOptions(op);
    for (const field of fields) defineField(command, field);
    command.addHelpText("after", `\n${metaLine(op)}`);

    command.action(async () => {
      const parsed = command.opts();
      const values: FieldValues = {};
      for (const field of fields) {
        const value = parsed[field.optionKey];
        if (value !== undefined) values[field.field] = value;
      }
      const flags = readGlobalFlags(command);
      emit(await options.run(command, op, values), flags.json);
    });
  }
}
