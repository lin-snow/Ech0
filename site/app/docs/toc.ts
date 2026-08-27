// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type TocItem = {
  depth: 2 | 3;
  text: string;
  id: string;
};

function slugifySegment(text: string): string {
  const s = text
    .trim()
    .replace(/\s+/g, "-")
    .replace(/[^\p{L}\p{N}_-]/gu, "");
  return s || "section";
}

function makeUniqueId(text: string, used: Map<string, number>): string {
  const base = slugifySegment(text);
  const n = (used.get(base) ?? 0) + 1;
  used.set(base, n);
  return n === 1 ? base : `${base}-${n}`;
}

export function extractTocFromMarkdown(md: string): TocItem[] {
  const lines = md.split(/\r?\n/);
  let inFence = false;
  const toc: TocItem[] = [];
  const used = new Map<string, number>();

  for (const line of lines) {
    const t = line.trimStart();
    if (t.startsWith("```")) {
      inFence = !inFence;
      continue;
    }
    if (inFence) continue;

    const h2 = line.match(/^##\s+(.+)$/);
    if (h2) {
      const raw = h2[1].trim();
      toc.push({
        depth: 2,
        text: raw,
        id: makeUniqueId(raw, used),
      });
      continue;
    }
    const h3 = line.match(/^###\s+(.+)$/);
    if (h3) {
      const raw = h3[1].trim();
      toc.push({
        depth: 3,
        text: raw,
        id: makeUniqueId(raw, used),
      });
    }
  }

  return toc;
}
