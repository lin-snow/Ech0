// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import { collectEntries, commandIds, helpModel, type HelpEntry } from "../help/model.ts";
import { renderHelp, renderListing } from "../help/render.ts";
import { resolvePath, searchEntries, suggestions } from "../help/search.ts";
import { colors, print, printErr } from "../output/console.ts";
import { yamlOf } from "../output/yaml.ts";
import { readGlobalFlags } from "../session.ts";

function emitDescriptor(target: Command, json: boolean, yaml: boolean): void {
  const model = helpModel(target);
  if (json) {
    print(JSON.stringify(model, null, 2));
    return;
  }
  if (yaml) {
    process.stdout.write(yamlOf(model));
    return;
  }
  process.stdout.write(renderHelp(model));
}

function emitSearch(
  root: Command,
  words: readonly string[],
  entries: readonly HelpEntry[],
  json: boolean,
  yaml: boolean,
): void {
  const query = words.join(" ");
  const { matches, total } = searchEntries(entries, query);
  if (json || yaml) {
    const document = {
      query,
      total,
      entries: matches.map(({ id, summary, tags }) => ({ id, summary, tags })),
    };
    process.stdout.write(json ? `${JSON.stringify(document, null, 2)}\n` : yamlOf(document));
    return;
  }
  if (matches.length === 0) {
    const hints = suggestions(commandIds(root), words);
    const hint = hints.length === 0 ? "" : `；是否想找：${hints.map((id) => `ech0 ${id.replaceAll(".", " ")}`).join("、")}`;
    printErr(`${colors().yellow("提示")} 没有匹配「${query}」的命令${hint}`);
    return;
  }
  print(colors().dim(`匹配「${query}」· ${total} 个`));
  process.stdout.write(renderListing(matches, total));
}

function emitFull(entries: readonly HelpEntry[], json: boolean, yaml: boolean): void {
  if (json || yaml) {
    const document = { commands: entries.map((entry) => helpModel(entry.command)) };
    process.stdout.write(json ? `${JSON.stringify(document, null, 2)}\n` : yamlOf(document));
    return;
  }
  process.stdout.write(entries.map((entry) => renderHelp(helpModel(entry.command))).join("\n"));
}

/** `ech0 help [words...]`: the map, one namespace, a search, or one command. */
export function createHelpCommand(): Command {
  return new Command("help")
    .description("查看命令地图，按关键词搜索，或打印一条命令的完整说明")
    .argument("[words...]", "命令路径或关键词，如 `echo create`、`page`")
    .option("--full", "打印每条命令的完整说明", false)
    .action((words: string[], options: { full?: boolean }, command: Command) => {
      const root = command.parent ?? command;
      const flags = readGlobalFlags(command);
      const { json, yaml } = flags;
      const entries = collectEntries(root);
      const target = resolvePath(root, words);

      if (words.length > 0 && target !== undefined) {
        emitDescriptor(target, json, yaml);
        return;
      }
      if (options.full === true) {
        emitFull(entries, json, yaml);
        return;
      }
      if (target !== undefined) {
        emitDescriptor(target, json, yaml);
        return;
      }
      emitSearch(root, words, entries, json, yaml);
    });
}
