#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * Release helper for the `ech0` CLI, wrapped by `just cli bump` and
 * `just cli release`.
 *
 *   bun run scripts/release.ts bump 0.1.4
 *   bun run scripts/release.ts release 0.1.4 [--dry-run]
 *
 * `bump` validates the tree, typechecks and rewrites `package.json`, then prints
 * the commit/tag/push steps for a manual release.
 *
 * `release` runs the full check set, bumps, commits `chore(cli-release): vX.Y.Z`,
 * creates `cli-vX.Y.Z` and pushes the branch and tag. The tag starts
 * `.github/workflows/release-cli.yml`, which builds the binaries and publishes
 * the npm packages; `--dry-run` stops after the checks.
 */

import { join, resolve } from "node:path";

const CLI_DIR = resolve(import.meta.dir, "..");
const REPO_DIR = resolve(CLI_DIR, "../..");
const PACKAGE_FILE = join(CLI_DIR, "package.json");
const TAG_PREFIX = "cli-v";
const SEMVER = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?$/;

function fail(message: string): never {
  console.error(`release: ✘ ${message}`);
  process.exit(1);
}

function info(message: string): void {
  console.log(`release: ${message}`);
}

function run(command: string[], cwd = REPO_DIR): void {
  const result = Bun.spawnSync(command, { cwd, stdout: "inherit", stderr: "inherit" });
  if (result.exitCode !== 0) {
    fail(`${command.join(" ")} exited with ${result.exitCode}`);
  }
}

function output(command: string[], cwd = REPO_DIR): string {
  const result = Bun.spawnSync(command, { cwd, stdout: "pipe", stderr: "pipe" });
  if (result.exitCode !== 0) {
    fail(`${command.join(" ")} exited with ${result.exitCode}`);
  }
  return result.stdout.toString().trim();
}

const [mode, version, ...flags] = process.argv.slice(2);
if (mode !== "bump" && mode !== "release") {
  fail("usage: scripts/release.ts <bump|release> <version> [--dry-run]");
}
if (version === undefined || !SEMVER.test(version)) {
  fail(`"${version ?? ""}" is not a semver version (expected X.Y.Z[-prerelease])`);
}

const dryRun = flags.includes("--dry-run");
const tag = `${TAG_PREFIX}${version}`;
const pkg = (await Bun.file(PACKAGE_FILE).json()) as { version?: string };

if (pkg.version === undefined) {
  fail("apps/cli/package.json has no version");
}
if (pkg.version === version) {
  fail(`apps/cli/package.json is already at ${version} — nothing to release`);
}
if (output(["git", "status", "--porcelain"]) !== "") {
  fail("working tree is dirty — commit or stash first so the release commit is clean");
}
if (mode === "release") {
  if (output(["git", "tag", "--list", tag]) !== "") {
    fail(`tag ${tag} already exists`);
  }
  if (output(["git", "rev-parse", "--abbrev-ref", "HEAD"]) === "HEAD") {
    fail("HEAD is detached — check out the branch that should carry the release commit");
  }
}

const changelog = await Bun.file(join(REPO_DIR, "CHANGELOG.md")).text();
if (!changelog.includes(`## [${version}]`)) {
  console.warn(`release: warning: CHANGELOG.md has no "## [${version}]" section yet`);
}

info("typecheck");
run(["bun", "run", "typecheck"], CLI_DIR);

if (mode === "release") {
  info("test");
  run(["bun", "test"], CLI_DIR);
  info("generated types");
  run(["bun", "run", "gen:api"], CLI_DIR);
  run(["git", "diff", "--exit-code", "--", "src/api/generated"], CLI_DIR);
}

if (dryRun) {
  info(`dry run: would bump ${pkg.version} -> ${version}`);
  info(`dry run: would commit "chore(cli-release): v${version}"`);
  info(`dry run: would tag ${tag} and push the branch + tag`);
  process.exit(0);
}

await Bun.write(PACKAGE_FILE, `${JSON.stringify({ ...pkg, version }, null, 2)}\n`);
info(`bumped ${pkg.version} -> ${version}`);

if (mode === "bump") {
  console.log(`
Next steps:
  # 1. Update CHANGELOG.md: rename [Unreleased] to [${version}] and open a new [Unreleased]
  # 2. Commit the bump (plus the changelog):
       git commit -am 'chore(cli-release): v${version}'
  # 3. Tag and push:
       git tag -a ${tag} -m 'Release ech0 CLI v${version}'
       git push origin main
       git push origin ${tag}
`);
  process.exit(0);
}

const branch = output(["git", "rev-parse", "--abbrev-ref", "HEAD"]);
if (branch !== "main") {
  console.warn(`release: warning: releasing from ${branch}; the bump commit is pushed there`);
}

run(["git", "add", "apps/cli/package.json"]);
run(["git", "commit", "-m", `chore(cli-release): v${version}`]);
run(["git", "tag", "-a", tag, "-m", `Release ech0 CLI v${version}`]);
info(`pushing ${branch} and ${tag}`);
run(["git", "push", "origin", branch]);
run(["git", "push", "origin", tag]);
info(`pushed ${tag}; watch release-cli.yml in GitHub Actions for the build and npm publish`);
