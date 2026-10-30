// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { deleteToken } from "../config/credentials.ts";
import { saveConfig, withoutProfile } from "../config/store.ts";
import { UsageError, colors } from "../output/console.ts";
import { renderTable } from "../output/render.ts";
import type { CliContext } from "../session.ts";
import type { CommandOutput } from "./types.ts";

export function list(ctx: CliContext): CommandOutput {
  const entries = Object.entries(ctx.config.profiles);
  if (entries.length === 0) {
    return {
      human: "还没有任何 profile：运行 `ech0 auth login --server <url> --token <token>`",
      json: "[]",
    };
  }

  const rows = entries.map(([name, profile]) => [
    name === ctx.config.default ? `* ${name}` : name,
    profile.server,
    profile.user ?? "-",
    profile.isAdmin ? "yes" : "no",
  ]);
  const payload = entries.map(([name, profile]) => ({
    name,
    default: name === ctx.config.default,
    server: profile.server,
    user: profile.user ?? null,
    is_admin: profile.isAdmin ?? null,
  }));
  return {
    human: renderTable(["profile", "server", "user", "admin"], rows),
    json: JSON.stringify(payload, null, 2),
  };
}

export function use(ctx: CliContext, name: string): CommandOutput {
  if (!ctx.config.profiles[name]) {
    throw new UsageError(`profile "${name}" 不存在`);
  }
  saveConfig({ ...ctx.config, default: name }, ctx.configPath);
  const { green } = colors();
  return {
    human: `${green("✓")} 默认 profile 已切换为 "${name}"`,
    json: JSON.stringify({ default: name }, null, 2),
  };
}

export async function unset(ctx: CliContext, name: string): Promise<CommandOutput> {
  if (!ctx.config.profiles[name]) {
    throw new UsageError(`profile "${name}" 不存在`);
  }
  saveConfig(withoutProfile(ctx.config, name), ctx.configPath);
  await deleteToken(name, ctx.credentialsPath);
  const { green } = colors();
  return {
    human: `${green("✓")} 已删除 profile "${name}" 及其 token`,
    json: JSON.stringify({ removed: name }, null, 2),
  };
}

/** Prints where the config and credentials live (support/debugging aid). */
export function path(ctx: CliContext): CommandOutput {
  const payload = { config: ctx.configPath, credentials: ctx.credentialsPath };
  return {
    human: `config       ${ctx.configPath}\ncredentials  ${ctx.credentialsPath}`,
    json: JSON.stringify(payload, null, 2),
  };
}
