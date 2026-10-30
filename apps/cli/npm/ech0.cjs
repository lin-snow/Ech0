#!/usr/bin/env node
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Launcher for the platform-specific binary shipped as an optional dependency
// (@ech0/cli-<os>-<arch>). Kept as plain CommonJS so `npm install -g` works with
// any supported Node version and no install scripts run.

"use strict";

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const SUPPORTED = {
  darwin: ["arm64", "x64"],
  linux: ["arm64", "x64"],
};

const platform = process.platform;
const arch = process.arch;

if (!SUPPORTED[platform] || !SUPPORTED[platform].includes(arch)) {
  const list = Object.entries(SUPPORTED)
    .flatMap(([os, arches]) => arches.map((candidate) => `${os}/${candidate}`))
    .join(", ");
  console.error(
    `ech0: unsupported platform ${platform}/${arch}. Supported: ${list}. ` +
      "Build from source: https://github.com/lin-snow/ech0/tree/main/apps/cli",
  );
  process.exit(1);
}

const pkg = `@ech0/cli-${platform}-${arch}`;
let binary;
try {
  binary = path.join(path.dirname(require.resolve(`${pkg}/package.json`)), "bin", "ech0");
} catch {
  console.error(
    `ech0: the optional dependency ${pkg} is not installed. ` +
      "Reinstall without --omit=optional: npm install -g @ech0/cli",
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(`ech0: failed to start ${binary}: ${result.error.message}`);
  process.exit(1);
}
if (result.signal !== null) {
  process.kill(process.pid, result.signal);
}
process.exit(result.status ?? 1);
