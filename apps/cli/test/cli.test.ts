// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const ENTRY = join(import.meta.dir, "..", "src", "index.ts");

function runCli(args: string[]): { stdout: string; stderr: string; exitCode: number } {
  const result = Bun.spawnSync({
    cmd: [process.execPath, "run", ENTRY, ...args],
    env: {
      ...process.env,
      ECH0_NO_KEYCHAIN: "1",
      ECH0_CONFIG: join(mkdtempSync(join(tmpdir(), "ech0-cli-spawn-")), "config.json"),
    },
    stdout: "pipe",
    stderr: "pipe",
  });
  return {
    stdout: result.stdout.toString(),
    stderr: result.stderr.toString(),
    exitCode: result.exitCode ?? -1,
  };
}

describe("cli wiring", () => {
  test("bare invocation prints help and exits 0", () => {
    const result = runCli([]);
    expect(result.exitCode).toBe(0);
    expect(result.stdout).toContain("Usage: ech0");
  });

  test("unknown option exits with the usage code", () => {
    const result = runCli(["echo", "ls", "--bogus"]);
    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain("unknown option");
  });

  test("commands without credentials fail as usage errors", () => {
    const result = runCli(["me"]);
    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain("未配置任何 profile");
  });
});
