// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Command } from "commander";

import { colors, setColorEnabled, shouldUseColor } from "../output/console.ts";
import { helpModel, type CommandRow, type HelpModel, type HelpRow } from "./model.ts";

const INDENT = "  ";
const GAP = "  ";
const FALLBACK_WIDTH = 100;
const MIN_DESCRIPTION = 24;

/** The formatter commander uses for every command's help. */
export function formatCommand(command: Command): string {
  const options = command.optsWithGlobals();
  setColorEnabled(options.color !== false && shouldUseColor());
  return renderHelp(helpModel(command));
}

function terminalWidth(): number {
  const columns = process.stdout.columns;
  return columns !== undefined && columns >= 40 ? columns : FALLBACK_WIDTH;
}

/** East Asian wide characters occupy two terminal cells. */
function characterWidth(character: string): number {
  const code = character.codePointAt(0) ?? 0;
  const wide =
    (code >= 0x1100 && code <= 0x115f) ||
    code === 0x2329 ||
    code === 0x232a ||
    (code >= 0x2e80 && code <= 0xa4cf && code !== 0x303f) ||
    (code >= 0xac00 && code <= 0xd7a3) ||
    (code >= 0xf900 && code <= 0xfaff) ||
    (code >= 0xfe30 && code <= 0xfe6f) ||
    (code >= 0xff00 && code <= 0xff60) ||
    (code >= 0xffe0 && code <= 0xffe6);
  return wide ? 2 : 1;
}

function textWidth(text: string): number {
  let total = 0;
  for (const character of text) total += characterWidth(character);
  return total;
}

function pad(text: string, target: number): string {
  return text + " ".repeat(Math.max(0, target - textWidth(text)));
}

/** Word wrap that also splits a run of CJK text, which has no spaces to break at. */
function wrapText(text: string, limit: number): string[] {
  const lines: string[] = [];
  let line = "";
  const push = (): void => {
    if (line.trim() !== "") lines.push(line.trimEnd());
    line = "";
  };

  for (const word of text.split(/(\s+)/)) {
    if (word === "") continue;
    if (textWidth(line) + textWidth(word) <= limit) {
      line += word;
      continue;
    }
    push();
    if (textWidth(word) <= limit) {
      line = word.trimStart();
      continue;
    }
    let piece = "";
    for (const character of word) {
      if (textWidth(piece) + characterWidth(character) > limit && piece !== "") {
        lines.push(piece);
        piece = "";
      }
      piece += character;
    }
    line = piece;
  }
  push();
  return lines.length > 0 ? lines : [""];
}

/** Aligned term | tags | description table, wrapped to the terminal. */
function table(rows: readonly HelpRow[], limit: number): string[] {
  const palette = colors();
  const tagged = rows.some((row) => row.tags.length > 0);
  const termWidth = Math.max(...rows.map((row) => textWidth(row.term)));
  const tagWidth = tagged ? Math.max(...rows.map((row) => textWidth(row.tags.join(" ")))) : 0;
  const used = termWidth + GAP.length + (tagged ? tagWidth + GAP.length : 0);
  const descriptionWidth = Math.max(MIN_DESCRIPTION, limit - INDENT.length - used);

  const lines: string[] = [];
  for (const row of rows) {
    const cells: string[] = [palette.cyan(pad(row.term, termWidth))];
    if (tagged) cells.push(palette.dim(pad(row.tags.join(" "), tagWidth)));
    const description = row.description === "" ? [""] : wrapText(row.description, descriptionWidth);
    const [first = "", ...rest] = description;
    lines.push(`${INDENT}${[...cells, first].join(GAP)}`.trimEnd());
    const blank = INDENT + " ".repeat(used);
    for (const line of rest) lines.push(`${blank}${line}`.trimEnd());
  }
  return lines;
}

/** Listing rows reuse the table with the command id as the term. */
function listingRows(rows: readonly CommandRow[]): HelpRow[] {
  return rows.map((row) => ({ term: row.id, description: row.summary, tags: row.tags }));
}

function section(title: string, lines: readonly string[]): string {
  return lines.length === 0 ? "" : [colors().bold(title), ...lines].join("\n");
}

function serverSection(model: HelpModel): string {
  const palette = colors();
  const context = model.server;
  if (context?.loaded === true) {
    const where = context.server === undefined ? "" : `（${context.server}）`;
    const count = model.groups.reduce((total, group) => total + group.count, 0);
    const nameWidth = Math.max(0, ...model.groups.map((group) => textWidth(group.name)));
    const lines = model.groups.map((group) => {
      const preview = group.preview === "" ? "" : palette.dim(` · ${group.preview}`);
      return `${INDENT}${palette.cyan(pad(group.name, nameWidth))}${GAP}${group.count} 个操作${preview}`;
    });
    if (context.ops !== undefined && context.ops > count) {
      lines.push(
        `${INDENT}${palette.dim(`catalog 共 ${context.ops} 个操作，CLI 当前开放 ${count} 个`)}`,
      );
    }
    return section(`服务端命令${where} · ${count} 个操作`, lines);
  }
  const reason = context?.failure ?? (context === undefined ? "未配置实例" : "catalog 尚未加载");
  return section("服务端命令", [
    `${INDENT}${palette.dim(reason)}`,
    `${INDENT}${palette.dim("先运行 `ech0 auth login --server <url>`，或临时设置 ECH0_SERVER=<url>")}`,
  ]);
}

/** The full text view of one command: usage, server card, tables, examples. */
export function renderHelp(model: HelpModel): string {
  const palette = colors();
  const limit = terminalWidth();
  const blocks: string[] = [`${palette.bold("用法:")} ${model.usage}`];

  if (model.summary !== "") blocks.push(model.summary);
  if (model.facet !== undefined) {
    blocks.push(wrapText(`${palette.dim("接口:")} ${model.facet}`, limit).join("\n"));
  }

  if (model.fields.length > 0) blocks.push(section("参数", table(model.fields, limit)));
  if (model.options.length > 0 && !model.root) {
    blocks.push(section("选项", table(model.options, limit)));
  }
  if (model.commands.length > 0) {
    blocks.push(section("子命令", table(listingRows(model.commands), limit)));
  }
  if (model.root) {
    if (model.locals.length > 0) {
      blocks.push(section("本地命令", table(listingRows(model.locals), limit)));
    }
    blocks.push(serverSection(model));
    if (model.options.length > 0) blocks.push(section("全局选项", table(model.options, limit)));
  }

  if (!model.root && model.globalHint !== undefined) {
    blocks.push(wrapText(model.globalHint, limit).map((line) => palette.dim(line)).join("\n"));
  }

  if (model.examples.length > 0) {
    blocks.push(section("示例", model.examples.map((example) => `${INDENT}${example}`)));
  }
  if (model.footer !== undefined) {
    blocks.push(wrapText(model.footer, limit).map((line) => palette.dim(line)).join("\n"));
  }
  if (model.hints.length > 0) {
    const lines = model.hints.flatMap((hint) =>
      wrapText(hint, limit - INDENT.length).map((line) => palette.dim(`${INDENT}${line}`)),
    );
    blocks.push(section("提示", lines));
  }

  return `${blocks.filter((block) => block !== "").join("\n\n")}\n`;
}

/** Search results: one row per command, ids relative to the root. */
export function renderListing(rows: readonly CommandRow[], total: number): string {
  const palette = colors();
  const lines = table(listingRows(rows), terminalWidth());
  if (total > rows.length) {
    lines.push(palette.dim(`${INDENT}共 ${total} 个匹配，显示前 ${rows.length} 个…`));
  }
  return `${lines.join("\n")}\n`;
}
