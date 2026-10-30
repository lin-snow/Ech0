// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Echo, Tag, User } from "../api/types.gen.ts";
import { colors } from "./console.ts";

/** Display width in terminal cells; CJK and fullwidth forms count as 2. */
export function displayWidth(text: string): number {
  let width = 0;
  for (const char of text) {
    const code = char.codePointAt(0) ?? 0;
    width += isWide(code) ? 2 : 1;
  }
  return width;
}

function isWide(code: number): boolean {
  return (
    (code >= 0x1100 && code <= 0x115f) ||
    (code >= 0x2e80 && code <= 0xa4cf) ||
    (code >= 0xac00 && code <= 0xd7a3) ||
    (code >= 0xf900 && code <= 0xfaff) ||
    (code >= 0xfe30 && code <= 0xfe6f) ||
    (code >= 0xff00 && code <= 0xff60) ||
    (code >= 0xffe0 && code <= 0xffe6) ||
    (code >= 0x1f300 && code <= 0x1f64f) ||
    (code >= 0x1f900 && code <= 0x1f9ff)
  );
}

function truncate(text: string, width: number): string {
  if (displayWidth(text) <= width) return text;
  let out = "";
  for (const char of text) {
    if (displayWidth(out + char) > width - 1) break;
    out += char;
  }
  return `${out}…`;
}

/** Renders rows as a left-aligned table with a dim header. */
export function renderTable(headers: string[], rows: string[][]): string {
  const widths = headers.map((header, column) =>
    Math.max(displayWidth(header), ...rows.map((row) => displayWidth(row[column] ?? ""))),
  );
  const { bold, dim } = colors();
  const cell = (text: string, column: number): string => {
    const width = widths[column] ?? 0;
    const gap = width - displayWidth(text);
    return gap > 0 ? text + " ".repeat(gap) : text;
  };
  const headerLine = headers.map((header, column) => bold(cell(header, column))).join("  ");
  const separator = dim(widths.map((width) => "─".repeat(width)).join("  "));
  const body = rows.map((row) => row.map((value, column) => cell(value ?? "", column)).join("  "));
  return [headerLine, separator, ...body].join("\n");
}

export function formatEpoch(epoch?: number): string {
  if (epoch === undefined || epoch === null || epoch <= 0) return "-";
  const date = new Date(epoch * 1000);
  const pad2 = (value: number) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}`;
}

export function shortId(id?: string): string {
  return id ? id.slice(0, 8) : "-";
}

function firstLine(content?: string): string {
  const line = (content ?? "").split("\n").find((candidate) => candidate.trim() !== "") ?? "";
  return line.trim();
}

export function tagNames(tags?: Tag[] | null): string[] {
  return (tags ?? []).flatMap((tag) => (tag.name ? [tag.name] : []));
}

/** One row per echo: time, short id, flags, first content line. */
export function renderEchoList(echos: Echo[]): string {
  const width = Math.max(40, Math.min(process.stdout.columns ?? 100, 160));
  const contentWidth = Math.max(20, width - 26);
  const rows = echos.map((echo) => [
    formatEpoch(echo.created_at),
    shortId(echo.id),
    echo.private ? "私密" : "-",
    truncate(
      (echo.content ?? "").split("\n").find((line) => line.trim() !== "")?.trim() ?? "",
      contentWidth,
    ),
  ]);
  return renderTable(["时间", "ID", "可见性", "内容"], rows);
}

export function renderEchoDetail(echo: Echo): string {
  const { bold, dim } = colors();
  const lines = [
    `${bold("ID")}        ${echo.id ?? "-"}`,
    `${bold("时间")}      ${formatEpoch(echo.created_at)}`,
    `${bold("作者")}      ${echo.username ?? "-"}`,
    `${bold("可见性")}    ${echo.private ? "私密" : "公开"}`,
    `${bold("标签")}      ${tagNames(echo.tags).join(", ") || "-"}`,
    `${bold("喜欢")}      ${echo.fav_count ?? 0}`,
  ];
  if (echo.echo_files && echo.echo_files.length > 0) {
    lines.push(`${bold("媒体")}      ${echo.echo_files.length} 个文件`);
  }
  lines.push("", dim("─".repeat(48)), echo.content ?? "");
  return lines.join("\n");
}

export function renderUser(user: User): string {
  const { bold } = colors();
  return [
    `${bold("用户")}    ${user.username ?? "-"}`,
    `${bold("ID")}      ${user.id ?? "-"}`,
    `${bold("邮箱")}    ${user.email || "-"}`,
    `${bold("角色")}    ${user.is_admin ? "管理员" : "普通用户"}${user.is_owner ? "（owner）" : ""}`,
  ].join("\n");
}
