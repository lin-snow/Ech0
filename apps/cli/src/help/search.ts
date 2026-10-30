// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Command } from "commander";

import { nodeAt, type HelpEntry } from "./model.ts";

const SEARCH_LIMIT = 20;

export interface SearchResult {
  matches: HelpEntry[];
  total: number;
}

/** ASCII tokens match whole words, so `file` does not hit `profile`; CJK matches inline. */
function hasToken(text: string, token: string): boolean {
  if (/[^\x00-\x7f]/.test(token)) return text.includes(token);
  for (let from = 0; ; from += 1) {
    const at = text.indexOf(token, from);
    if (at === -1) return false;
    const before = text[at - 1];
    const after = text[at + token.length];
    const boundary = (character: string | undefined): boolean =>
      character === undefined || !/[a-z0-9]/.test(character);
    if (boundary(before) && boundary(after)) return true;
  }
}

function scoreOf(entry: HelpEntry, token: string): number {
  let score = 0;
  const words = entry.id.toLowerCase().split(".");
  if (words.includes(token)) score += 10;
  else if (words.some((word) => word.startsWith(token))) score += 6;
  else if (entry.id.toLowerCase().includes(token)) score += 4;

  if (hasToken(entry.summary.toLowerCase(), token)) score += 3;
  if (hasToken(entry.haystack.toLowerCase(), token)) score += 1;
  return score;
}

/** Ranked keyword search over ids, summaries, field names and field descriptions. */
export function searchEntries(entries: readonly HelpEntry[], query: string): SearchResult {
  const terms = query
    .toLowerCase()
    .split(/[^\p{L}\p{N}]+/u)
    .filter((token) => token !== "");
  if (terms.length === 0) return { matches: [], total: 0 };

  const scored = entries
    .map((entry) => {
      let score = 0;
      for (const term of terms) score += scoreOf(entry, term);
      return { entry, score };
    })
    .filter(({ score }) => score > 0)
    .sort((left, right) => {
      if (left.score !== right.score) return right.score - left.score;
      return left.entry.id.localeCompare(right.entry.id);
    });

  return {
    matches: scored.slice(0, SEARCH_LIMIT).map(({ entry }) => entry),
    total: scored.length,
  };
}

function editDistance(left: string, right: string): number {
  const rows = left.length + 1;
  const columns = right.length + 1;
  let previous = Array.from({ length: columns }, (_, index) => index);
  for (let row = 1; row < rows; row++) {
    const current = [row];
    for (let column = 1; column < columns; column++) {
      const substitution = (previous[column - 1] ?? 0) + (left[row - 1] === right[column - 1] ? 0 : 1);
      current[column] = Math.min(
        (previous[column] ?? 0) + 1,
        (current[column - 1] ?? 0) + 1,
        substitution,
      );
    }
    previous = current;
  }
  return previous[columns - 1] ?? 0;
}

/** Close names for a miss, so a typo costs one retry instead of a docs dive. */
export function suggestions(ids: readonly string[], words: readonly string[]): string[] {
  const typed = words.join(" ").toLowerCase();
  const last = (words.at(-1) ?? "").toLowerCase();
  if (last === "") return [];

  const scored: Array<{ id: string; distance: number }> = [];
  for (const id of ids) {
    const lower = id.toLowerCase();
    if (lower === typed) continue;
    const leaf = lower.split(".").at(-1) ?? lower;
    const distance = Math.min(editDistance(last, leaf), editDistance(typed, lower));
    const prefix = lower.startsWith(typed) || leaf.startsWith(last);
    if (!prefix && distance > (last.length <= 3 ? 1 : 2)) continue;
    scored.push({ id, distance: prefix ? 0 : distance });
  }

  return scored
    .sort((left, right) => left.distance - right.distance || left.id.localeCompare(right.id))
    .slice(0, 3)
    .map(({ id }) => id);
}

/** Walks dotted ids and spaced words alike: `echo.create` names `ech0 echo create`. */
export function resolvePath(root: Command, words: readonly string[]): Command | undefined {
  const path = words.flatMap((word) => word.split(".")).filter((word) => word !== "");
  if (path.length === 0) return root;
  return nodeAt(root, path);
}
