// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { homedir } from "node:os";
import { join, relative, resolve } from "node:path";

import { Command } from "commander";

import { printResult } from "../output/result.ts";
import { readGlobalFlags } from "../session.ts";
import { installSkill } from "../skills/install.ts";
import { openSource, SKILL_NAME } from "../skills/source.ts";

/** `~` in an argument is the caller's home, not a literal directory name. */
function expandHome(path: string): string {
  if (path === "~") return homedir();
  return path.startsWith("~/") ? join(homedir(), path.slice(2)) : path;
}

export function installCommand(): Command {
  const group = new Command("install").description("安装 Ech0 资源到其他工具");

  group
    .command("skills <dir>")
    .description("把内置 Ech0 skill 写入 skills 根目录（如 ~/.claude/skills）")
    .option("--from <folder>", "安装本地 skill 文件夹，而不是内置版本")
    .action((dir: string, _options, command: Command) => {
      const flags = readGlobalFlags(command);
      const options = command.opts();
      const root = resolve(expandHome(dir));
      const source = openSource(typeof options.from === "string" ? options.from : undefined);
      const wrote = installSkill(root, source);
      const target = join(root, SKILL_NAME);
      const files = wrote.map((path) => relative(target, path));

      printResult(
        {
          human: [`✓ 已安装 Ech0 skill 到 ${target}`, ...files.map((path) => `  ${path}`)].join("\n"),
          json: JSON.stringify({ target, files }, null, 2),
        },
        flags,
      );
    });

  return group;
}
