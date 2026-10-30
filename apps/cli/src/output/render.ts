// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

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
