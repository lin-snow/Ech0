// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { User } from "../api/types.gen.ts";
import { Client } from "../api/http.ts";
import { deleteToken, readToken, writeToken } from "../config/credentials.ts";
import { activeProfileName, saveConfig, withProfile, withoutProfile } from "../config/store.ts";
import { UsageError, colors } from "../output/console.ts";
import { renderUser } from "../output/render.ts";
import { resolveTarget, type CliContext } from "../session.ts";
import type { CommandOutput } from "./types.ts";

export interface LoginArgs {
  server?: string;
  token: string;
  profile?: string;
}

/**
 * Verifies the token against `/api/user`, then persists the profile and token.
 * Password and device authorization land in M3/M4 (docs/dev/cli-design.md §5).
 */
export async function login(ctx: CliContext, args: LoginArgs): Promise<CommandOutput> {
  const profile =
    args.profile ??
    activeProfileName(ctx.config, undefined, ctx.env) ??
    "default";
  const server =
    args.server ??
    ctx.env.ECH0_SERVER ??
    ctx.config.profiles[profile]?.server;
  if (!server) {
    throw new UsageError("缺少实例地址：--server、ECH0_SERVER 或已有 profile");
  }

  const lang = ctx.flags.lang ?? ctx.env.ECH0_LANG;
  const client = new Client({ server, token: args.token, lang });
  const result = await client.call<User>({ path: "/user" });
  const user = result.data;

  const source = await writeToken(profile, args.token, ctx.credentialsPath);
  saveConfig(
    withProfile(ctx.config, profile, {
      server,
      user: user.username,
      isAdmin: user.is_admin,
    }),
    ctx.configPath,
  );

  const { green, dim } = colors();
  return {
    human: [
      `${green("✓")} 已登录 ${user.username ?? "-"} @ ${server}`,
      dim(`profile: ${profile} · token: ${source}`),
    ].join("\n"),
    json: result.text,
  };
}

export interface StatusArgs {
  verify: boolean;
}

export async function status(ctx: CliContext, args: StatusArgs): Promise<CommandOutput> {
  const target = resolveTarget(ctx);
  const stored = await readToken(target.profile, ctx.credentialsPath);
  const tokenSource = ctx.flags.token
    ? "flag"
    : ctx.env.ECH0_TOKEN
      ? "env"
      : (stored?.source ?? "missing");
  const cached = target.config.profiles[target.profile];

  const payload = {
    profile: target.profile,
    server: target.server,
    token_source: tokenSource,
    user: cached?.user ?? null,
    is_admin: cached?.isAdmin ?? null,
  };

  if (!args.verify) {
    const { bold } = colors();
    return {
      human: [
        `${bold("profile")}    ${target.profile}`,
        `${bold("server")}     ${target.server}`,
        `${bold("token")}      ${tokenSource}`,
        `${bold("user")}       ${cached?.user ?? "-"}`,
        `${bold("admin")}      ${cached?.isAdmin ? "yes" : "no"}`,
      ].join("\n"),
      json: JSON.stringify(payload, null, 2),
    };
  }

  const token = ctx.flags.token ?? ctx.env.ECH0_TOKEN ?? stored?.token;
  if (!token) throw new UsageError(`profile "${target.profile}" 没有可用 token`);
  const client = new Client({ server: target.server, token, lang: ctx.flags.lang ?? ctx.env.ECH0_LANG });
  const result = await client.call<User>({ path: "/user" });
  return {
    human: renderUser(result.data),
    json: result.text,
  };
}

export interface LogoutArgs {
  purge: boolean;
}

export async function logout(ctx: CliContext, args: LogoutArgs): Promise<CommandOutput> {
  const target = resolveTarget(ctx);
  await deleteToken(target.profile, ctx.credentialsPath);
  if (args.purge) {
    saveConfig(withoutProfile(target.config, target.profile), ctx.configPath);
  }

  const { green, dim } = colors();
  const payload = { profile: target.profile, purged: args.purge };
  return {
    human:
      `${green("✓")} 已清除 profile "${target.profile}" 的 token` +
      (args.purge ? dim("（profile 已删除）") : ""),
    json: JSON.stringify(payload, null, 2),
  };
}
