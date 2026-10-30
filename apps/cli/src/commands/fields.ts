// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { CatalogOp } from "../catalog/types.ts";

export type JSONSchema = Record<string, unknown>;

/** Flags the CLI owns; an operation field with the same name becomes `--input-<name>`. */
export const RESERVED_FLAGS: Record<string, true> = {
  json: true,
  profile: true,
  server: true,
  token: true,
  lang: true,
  color: true,
  help: true,
  version: true,
};

export interface FieldOption {
  field: string;
  /** camelCase key commander stores the parsed value under. */
  optionKey: string;
  flag: string;
  schema: JSONSchema;
  required: boolean;
  /** A multipart file part: the value is a path, `-` reads stdin. */
  binary: boolean;
}

export function schemaType(schema: JSONSchema): string {
  const type = schema["type"];
  if (typeof type === "string") return type;
  if (Array.isArray(type)) {
    for (const candidate of type) {
      if (typeof candidate === "string" && candidate !== "null") return candidate;
    }
  }
  return "string";
}

export function kebab(name: string): string {
  return name
    .replace(/_/g, "-")
    .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
    .toLowerCase();
}

export function camel(name: string): string {
  return name.replace(/-([a-z0-9])/g, (_, char: string) => char.toUpperCase());
}

export function enumValues(schema: JSONSchema): string[] {
  const values = schema["enum"];
  if (!Array.isArray(values)) return [];
  return values.filter((value): value is string => typeof value === "string");
}

/** A multipart file part: the CLI turns a filesystem path into an upload. */
export function isFileField(schema: JSONSchema): boolean {
  return schema["format"] === "binary";
}

export function fieldOptions(op: CatalogOp): FieldOption[] {
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
      binary: isFileField(rawSchema ?? {}),
    };
  });
}

/** The value a flag takes on the command line: `--page <integer>`, `--private` alone. */
export function flagTerm(field: FieldOption): string {
  if (schemaType(field.schema) === "boolean") return `--${field.flag}`;
  return `--${field.flag} <${valuePlaceholder(field)}>`;
}

function valuePlaceholder(field: FieldOption): string {
  if (isFileField(field.schema)) return "path";
  const type = schemaType(field.schema);
  if (type === "array") return "a,b";
  if (type === "object") return "json";
  if (type === "integer" || type === "number") return "number";
  return field.field.replace(/_/g, "-");
}
