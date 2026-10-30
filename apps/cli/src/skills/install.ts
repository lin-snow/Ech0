// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { mkdirSync, rmSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";

import { SKILL_NAME, type SkillSource } from "./source.ts";

/**
 * Writes the skill into `<root>/ech0`, replacing any previous install: the
 * folder is owned by the CLI, so stale references never survive an upgrade.
 */
export function installSkill(root: string, source: SkillSource): string[] {
  const target = join(root, SKILL_NAME);
  rmSync(target, { recursive: true, force: true });

  const wrote: string[] = [];
  for (const path of source.paths) {
    const abs = join(target, path);
    mkdirSync(dirname(abs), { recursive: true });
    const file = source.read(path);
    writeFileSync(abs, file.bytes, { mode: file.mode });
    wrote.push(abs);
  }
  return wrote;
}
