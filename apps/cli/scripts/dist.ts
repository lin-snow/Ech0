#!/usr/bin/env bun
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * Build the distributable CLI: standalone binaries (for GitHub Releases) and
 * the npm packages (`@ech0/cli` + one `@ech0/cli-<os>-<arch>` per platform).
 *
 *   bun run scripts/dist.ts --version 0.1.0 [--out dist/release] [--targets darwin-arm64,linux-x64]
 *
 * Cross-compilation is handled by `bun build --compile --target=bun-<os>-<arch>`;
 * the version is inlined via `--define process.env.ECH0_CLI_VERSION`, so the
 * checkout stays clean (see src/version-info.ts).
 *
 * Output:
 *   <out>/bin/ech0-<os>-<arch>              raw binaries
 *   <out>/ech0-cli-<os>-<arch>.tar.gz       release assets (`ech0` + LICENSE)
 *   <out>/SHA256SUMS                        checksums of the tarballs
 *   <out>/npm/cli                           publishable @ech0/cli (launcher + optionalDeps)
 *   <out>/npm/cli-<os>-<arch>               publishable platform packages
 */

import { chmodSync, copyFileSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";

const CLI_DIR = resolve(import.meta.dir, "..");
const REPO_DIR = resolve(CLI_DIR, "../..");
const LICENSE = join(REPO_DIR, "LICENSE");

type Os = "darwin" | "linux";
type Cpu = "arm64" | "x64";

interface Target {
  os: Os;
  cpu: Cpu;
}

const ALL_TARGETS: readonly Target[] = [
  { os: "darwin", cpu: "arm64" },
  { os: "darwin", cpu: "x64" },
  { os: "linux", cpu: "arm64" },
  { os: "linux", cpu: "x64" },
];

const SEMVER = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?$/;

function fail(message: string): never {
  console.error(`dist: ${message}`);
  process.exit(1);
}

function readArg(name: string): string | undefined {
  const index = process.argv.indexOf(name);
  return index >= 0 ? process.argv[index + 1] : undefined;
}

function run(cmd: string[], cwd: string): void {
  const result = Bun.spawnSync(cmd, { cwd, stdout: "inherit", stderr: "inherit" });
  if (result.exitCode !== 0) {
    fail(`${cmd.join(" ")} exited with ${result.exitCode}`);
  }
}

function writeJSON(file: string, value: unknown): void {
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

const version = readArg("--version");
if (version === undefined || !SEMVER.test(version)) {
  fail("--version <x.y.z[-prerelease]> is required (release tag without the leading `v`)");
}

const targets = ((): Target[] => {
  const raw = readArg("--targets");
  if (raw === undefined) return [...ALL_TARGETS];
  return raw.split(",").map((entry) => {
    const [os, cpu] = entry.trim().split("-");
    const target = ALL_TARGETS.find((candidate) => candidate.os === os && candidate.cpu === cpu);
    if (target === undefined) fail(`unknown target "${entry}" (expected ${ALL_TARGETS.map(label).join(", ")})`);
    return target;
  });
})();

const outDir = resolve(CLI_DIR, readArg("--out") ?? "dist/release");
const binDir = join(outDir, "bin");
const npmDir = join(outDir, "npm");

function label(target: Target): string {
  return `${target.os}-${target.cpu}`;
}

function packageName(target: Target): string {
  return `@ech0/cli-${label(target)}`;
}

const repository = process.env.GITHUB_REPOSITORY ?? "lin-snow/ech0";

function repositoryField(): Record<string, string> {
  return {
    type: "git",
    url: `git+https://github.com/${repository}.git`,
    directory: "apps/cli",
  };
}

rmSync(outDir, { recursive: true, force: true });
mkdirSync(binDir, { recursive: true });
mkdirSync(npmDir, { recursive: true });

const tarballs: Array<{ file: string; name: string }> = [];

for (const target of targets) {
  const name = `ech0-${label(target)}`;
  const binary = join(binDir, name);

  console.log(`\n==> compiling ${label(target)}`);
  run(
    [
      "bun",
      "build",
      "--compile",
      "--target",
      `bun-${label(target)}`,
      "--define",
      `process.env.ECH0_CLI_VERSION=${JSON.stringify(version)}`,
      "--outfile",
      binary,
      "src/index.ts",
    ],
    CLI_DIR,
  );
  chmodSync(binary, 0o755);

  // Release tarball: `ech0` + LICENSE, unpackable straight into a PATH dir.
  const stage = join(outDir, ".stage", label(target));
  mkdirSync(stage, { recursive: true });
  copyFileSync(binary, join(stage, "ech0"));
  chmodSync(join(stage, "ech0"), 0o755);
  copyFileSync(LICENSE, join(stage, "LICENSE"));
  const tarball = join(outDir, `ech0-cli-${label(target)}.tar.gz`);
  run(["tar", "-czf", tarball, "-C", stage, "ech0", "LICENSE"], outDir);
  tarballs.push({ file: tarball, name: `ech0-cli-${label(target)}.tar.gz` });

  // Platform npm package: the compiled binary only.
  const dir = join(npmDir, `cli-${label(target)}`);
  mkdirSync(join(dir, "bin"), { recursive: true });
  copyFileSync(binary, join(dir, "bin", "ech0"));
  chmodSync(join(dir, "bin", "ech0"), 0o755);
  copyFileSync(LICENSE, join(dir, "LICENSE"));
  writeJSON(join(dir, "package.json"), {
    name: packageName(target),
    version,
    description: `Ech0 CLI binary for ${label(target)}`,
    license: "AGPL-3.0-or-later",
    homepage: `https://github.com/${repository}/tree/main/apps/cli`,
    repository: repositoryField(),
    bugs: { url: `https://github.com/${repository}/issues` },
    os: [target.os],
    cpu: [target.cpu],
    files: ["bin/ech0", "LICENSE"],
    exports: { "./package.json": "./package.json" },
  });
}

rmSync(join(outDir, ".stage"), { recursive: true, force: true });

// Launcher package. npm installs exactly one platform package because the
// others are excluded by their os/cpu fields.
const metaDir = join(npmDir, "cli");
mkdirSync(join(metaDir, "bin"), { recursive: true });
copyFileSync(join(CLI_DIR, "npm", "ech0.cjs"), join(metaDir, "bin", "ech0.cjs"));
chmodSync(join(metaDir, "bin", "ech0.cjs"), 0o755);
copyFileSync(join(CLI_DIR, "npm", "README.md"), join(metaDir, "README.md"));
copyFileSync(LICENSE, join(metaDir, "LICENSE"));
writeJSON(join(metaDir, "package.json"), {
  name: "@ech0/cli",
  version,
  description: "Ech0 command line — talks to a running Ech0 instance over its REST API",
  license: "AGPL-3.0-or-later",
  homepage: `https://github.com/${repository}/tree/main/apps/cli`,
  repository: repositoryField(),
  bugs: { url: `https://github.com/${repository}/issues` },
  keywords: ["ech0", "cli", "microblog", "timeline", "self-hosted"],
  bin: { ech0: "bin/ech0.cjs" },
  files: ["bin/ech0.cjs", "README.md", "LICENSE"],
  engines: { node: ">=18" },
  optionalDependencies: Object.fromEntries(targets.map((target) => [packageName(target), version])),
  publishConfig: { access: "public" },
});

// SHA256SUMS over the release tarballs.
const lines: string[] = [];
for (const { file, name } of tarballs.sort((a, b) => a.name.localeCompare(b.name))) {
  const digest = new Bun.CryptoHasher("sha256").update(await Bun.file(file).arrayBuffer()).digest("hex");
  lines.push(`${digest}  ${name}`);
}
writeFileSync(join(outDir, "SHA256SUMS"), `${lines.join("\n")}\n`);

console.log(`\n==> ${version}: ${tarballs.length} binaries, npm packages under ${npmDir}`);
for (const { name } of tarballs) console.log(`    ${name}`);
