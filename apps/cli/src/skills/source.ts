// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * The skill ships inside the CLI: the Markdown files under `skills/ech0/` are
 * imported as text, so `bun build` (and `--compile`) embeds them. `--from` is
 * the escape hatch for installing a folder from disk instead.
 */

import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

import authReference from "../../../../skills/ech0/references/auth.md" with { type: "text" };
import echoReference from "../../../../skills/ech0/references/echo.md" with { type: "text" };
import operationsReference from "../../../../skills/ech0/references/operations.md" with { type: "text" };
import skillEntry from "../../../../skills/ech0/SKILL.md" with { type: "text" };

import { UsageError } from "../output/console.ts";

export const SKILL_NAME = "ech0";
export const SKILL_FILE = "SKILL.md";

const EMBEDDED: Record<string, string> = {
  [SKILL_FILE]: skillEntry,
  "references/auth.md": authReference,
  "references/echo.md": echoReference,
  "references/operations.md": operationsReference,
};

const FILE_MODE = 0o644;
const EXECUTABLE_MODE = 0o755;

export interface SkillFile {
  bytes: Uint8Array;
  mode: number;
}

export interface SkillSource {
  paths: readonly string[];
  read: (path: string) => SkillFile;
}

/** Embedded files carry no permission bits; a shebang marks a script. */
function embeddedMode(bytes: Uint8Array): number {
  return bytes[0] === 0x23 && bytes[1] === 0x21 ? EXECUTABLE_MODE : FILE_MODE;
}

function embeddedSource(): SkillSource {
  const paths = Object.keys(EMBEDDED).sort();
  return {
    paths,
    read: (path) => {
      const text = EMBEDDED[path];
      if (text === undefined) throw new UsageError(`内置 skill 缺少 ${path}`);
      const bytes = new TextEncoder().encode(text);
      return { bytes, mode: embeddedMode(bytes) };
    },
  };
}

function walk(root: string, dir: string, found: string[]): void {
  for (const entry of readdirSync(join(root, dir), { withFileTypes: true })) {
    const path = dir === "" ? entry.name : `${dir}/${entry.name}`;
    if (entry.isDirectory()) walk(root, path, found);
    else if (entry.isFile()) found.push(path);
  }
}

/** Installs a folder that has a SKILL.md; everything under it is copied. */
export function openDir(root: string): SkillSource {
  if (!statSync(join(root, SKILL_FILE), { throwIfNoEntry: false })?.isFile()) {
    throw new UsageError(`${root} 不是 skill 文件夹：缺少 ${SKILL_FILE}`);
  }
  const paths: string[] = [];
  walk(root, "", paths);
  paths.sort();
  return {
    paths,
    read: (path) => {
      const abs = join(root, path);
      return {
        bytes: new Uint8Array(readFileSync(abs)),
        mode: statSync(abs).mode & 0o777,
      };
    },
  };
}

export function openSource(from: string | undefined): SkillSource {
  return from === undefined ? embeddedSource() : openDir(from);
}
