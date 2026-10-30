// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { installSkill } from "../src/skills/install.ts";
import { openDir, openSource, SKILL_FILE, SKILL_NAME } from "../src/skills/source.ts";

const ENTRY = join(import.meta.dir, "..", "src", "index.ts");

function temp(prefix: string): string {
  return mkdtempSync(join(tmpdir(), prefix));
}

async function run(
  args: string[],
  env: NodeJS.ProcessEnv = {},
): Promise<{ stdout: string; stderr: string; exitCode: number }> {
  const base = { ...process.env };
  for (const key of ["ECH0_SERVER", "ECH0_TOKEN", "ECH0_PROFILE", "ECH0_CONFIG", "ECH0_LANG"]) {
    delete base[key];
  }
  const child = Bun.spawn({
    cmd: [process.execPath, "run", ENTRY, ...args],
    env: {
      ...base,
      ECH0_NO_KEYCHAIN: "1",
      ECH0_CONFIG: join(temp("ech0-cli-skill-"), "config.json"),
      ...env,
    },
    stdout: "pipe",
    stderr: "pipe",
  });
  const [stdout, stderr, exitCode] = await Promise.all([
    new Response(child.stdout).text(),
    new Response(child.stderr).text(),
    child.exited,
  ]);
  return { stdout, stderr, exitCode };
}

describe("skill source", () => {
  test("a local folder is copied wholesale, replacing the previous install", () => {
    const from = temp("ech0-skill-src-");
    mkdirSync(join(from, "references"), { recursive: true });
    writeFileSync(join(from, SKILL_FILE), "---\nname: ech0\n---\n");
    writeFileSync(join(from, "references", "setup.md"), "setup");

    const root = temp("ech0-skill-root-");
    mkdirSync(join(root, SKILL_NAME, "stale"), { recursive: true });
    writeFileSync(join(root, SKILL_NAME, "stale", "old.md"), "old");

    const wrote = installSkill(root, openDir(from));

    expect(wrote).toEqual([
      join(root, SKILL_NAME, SKILL_FILE),
      join(root, SKILL_NAME, "references", "setup.md"),
    ]);
    expect(existsSync(join(root, SKILL_NAME, "stale"))).toBe(false);
    expect(readFileSync(join(root, SKILL_NAME, "references", "setup.md"), "utf8")).toBe("setup");
  });

  test("a folder without SKILL.md is rejected", () => {
    expect(() => openDir(temp("ech0-skill-empty-"))).toThrow("缺少 SKILL.md");
  });

  test("the embedded skill ships the entry point and its references", () => {
    const source = openSource(undefined);
    const references = source.paths.filter((path) => path.startsWith("references/"));

    expect(source.paths).toContain(SKILL_FILE);
    expect(references.length).toBeGreaterThanOrEqual(3);
    expect(new TextDecoder().decode(source.read(SKILL_FILE).bytes)).toContain("name: ech0");
  });
});

describe("install skills", () => {
  test("writes the embedded skill and reports every file", async () => {
    const root = temp("ech0-skill-out-");
    const result = await run(["install", "skills", root, "--json"]);
    const parsed = JSON.parse(result.stdout) as { target: string; files: string[] };

    expect(result.exitCode).toBe(0);
    expect(parsed.target).toBe(join(root, SKILL_NAME));
    expect(parsed.files).toContain(SKILL_FILE);
    expect(readFileSync(join(root, SKILL_NAME, SKILL_FILE), "utf8")).toContain("name: ech0");
  });

  test("expands ~ in the skills root", async () => {
    const home = temp("ech0-skill-home-");
    const result = await run(["install", "skills", "~/agent-skills"], { HOME: home });

    expect(result.exitCode).toBe(0);
    expect(existsSync(join(home, "agent-skills", SKILL_NAME, SKILL_FILE))).toBe(true);
  });

  test("--from needs a folder with SKILL.md", async () => {
    const result = await run(["install", "skills", temp("ech0-skill-out-"), "--from", temp("ech0-skill-bad-")]);

    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain("缺少 SKILL.md");
  });
});
