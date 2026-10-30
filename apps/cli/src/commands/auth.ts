// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { User } from "../api/generated/types.gen.ts";
import { Client } from "../api/http.ts";
import {
  REFRESH_COOKIE_NAME,
  deleteRefreshCookie,
  deleteToken,
  readRefreshCookie,
  readToken,
  writeRefreshCookie,
  writeToken,
} from "../config/credentials.ts";
import { activeProfileName, saveConfig, withProfile, withoutProfile } from "../config/store.ts";
import { Ech0Error, UsageError, colors } from "../output/console.ts";
import { setCookieValue } from "../session.ts";
import { resolveTarget, type CliContext } from "../session.ts";
import type { CommandOutput } from "./types.ts";

/** Least privilege for CLI commands; widen with `--scopes`. */
export const DEFAULT_CLI_SCOPES = ["echo:read", "echo:write", "profile:read"];
const DEVICE_CLIENT_NAME = "ech0-cli";

interface TokenPair {
  access_token?: string;
  expires_in?: number;
}

function resolveProfile(ctx: CliContext, flag?: string): string {
  return flag ?? activeProfileName(ctx.config, undefined, ctx.env) ?? "default";
}

function resolveLoginServer(ctx: CliContext, flag: string | undefined, profile: string): string {
  const server = flag ?? ctx.env.ECH0_SERVER ?? ctx.config.profiles[profile]?.server;
  if (!server) {
    throw new UsageError("缺少实例地址：--server、ECH0_SERVER 或已有 profile");
  }
  return server;
}

function hostOf(server: string): string {
  try {
    return new URL(server).host;
  } catch {
    return server;
  }
}

function language(ctx: CliContext): string | undefined {
  return ctx.flags.lang ?? ctx.env.ECH0_LANG;
}

interface PersistOptions {
  profile: string;
  server: string;
  token: string;
  cookie?: string;
  user: User;
}

async function persistLogin(ctx: CliContext, options: PersistOptions): Promise<string> {
  const source = await writeToken(options.profile, options.token, ctx.credentialsPath);
  if (options.cookie) {
    await writeRefreshCookie(options.profile, options.cookie, ctx.credentialsPath);
  } else {
    await deleteRefreshCookie(options.profile, ctx.credentialsPath);
  }
  saveConfig(
    withProfile(ctx.config, options.profile, {
      server: options.server,
      user: options.user.username,
      isAdmin: options.user.is_admin,
    }),
    ctx.configPath,
  );
  return source;
}

function loginMessage(kind: string, user: User, server: string, profile: string, source: string): string {
  const { green, dim } = colors();
  return [
    `${green("✓")} 已登录 ${user.username ?? "-"} @ ${server}`,
    dim(`profile: ${profile} · ${kind} · 存储: ${source}`),
  ].join("\n");
}

export interface LoginArgs {
  server?: string;
  token: string;
  profile?: string;
}

/** Verifies a pasted access token against `/api/user`, then persists it. */
export async function login(ctx: CliContext, args: LoginArgs): Promise<CommandOutput> {
  const profile = resolveProfile(ctx, args.profile);
  const server = resolveLoginServer(ctx, args.server, profile);
  const client = new Client({ server, token: args.token, lang: language(ctx) });
  const result = await client.call<User>({ path: "/user" });
  const source = await persistLogin(ctx, {
    profile,
    server,
    token: args.token,
    user: result.data,
  });

  return {
    human: loginMessage("access token", result.data, server, profile, source),
    json: result.text,
  };
}

export interface PasswordLoginArgs {
  server?: string;
  profile?: string;
  username: string;
  password: string;
  scopes?: string[];
}

/**
 * Password login. Admins also mint a long-lived `audience=cli` access token so
 * later commands never depend on the session cookie; other users keep the
 * session token plus the refresh cookie and renew transparently.
 */
export async function loginWithPassword(
  ctx: CliContext,
  args: PasswordLoginArgs,
): Promise<CommandOutput> {
  const profile = resolveProfile(ctx, args.profile);
  const server = resolveLoginServer(ctx, args.server, profile);
  const lang = language(ctx);
  const client = new Client({ server, lang });

  const loginResult = await client.call<TokenPair>({
    method: "POST",
    path: "/login",
    body: { username: args.username, password: args.password },
  });
  const sessionToken = loginResult.data.access_token ?? "";
  if (sessionToken === "") {
    throw new Ech0Error("登录响应缺少 access_token");
  }
  const cookie = setCookieValue(loginResult.headers, REFRESH_COOKIE_NAME);

  const authed = new Client({ server, token: sessionToken, lang });
  const user = (await authed.call<User>({ path: "/user" })).data;

  let token = sessionToken;
  let storedCookie = cookie;
  let kind = "会话令牌（refresh cookie）";
  if (user.is_admin) {
    const minted = await authed.call<string>({
      method: "POST",
      path: "/access-tokens",
      body: {
        name: `cli:${hostOf(server)}`,
        expiry: "1_month",
        scopes: args.scopes && args.scopes.length > 0 ? args.scopes : DEFAULT_CLI_SCOPES,
        audience: "cli",
      },
    });
    if (typeof minted.data === "string" && minted.data !== "") {
      token = minted.data;
      storedCookie = undefined;
      kind = "CLI 令牌（1 个月）";
    }
  }

  const source = await persistLogin(ctx, { profile, server, token, cookie: storedCookie, user });
  return {
    human: loginMessage(kind, user, server, profile, source),
    json: JSON.stringify({ profile, server, user: user.username ?? null, is_admin: user.is_admin ?? null }, null, 2),
  };
}

export interface DeviceCodeInfo {
  device_code: string;
  user_code: string;
  verification_uri: string;
  verification_uri_complete: string;
  expires_in: number;
  interval: number;
}

export interface DeviceLoginArgs {
  server?: string;
  profile?: string;
  scopes?: string[];
  /** Called once the browser code is known. */
  onCode: (code: DeviceCodeInfo) => void;
  sleep?: (milliseconds: number) => Promise<void>;
}

function defaultSleep(milliseconds: number): Promise<void> {
  const { promise, resolve } = Promise.withResolvers<void>();
  setTimeout(resolve, milliseconds);
  return promise;
}

/** Device authorization: show a code, then poll until the browser approves. */
export async function loginWithDevice(ctx: CliContext, args: DeviceLoginArgs): Promise<CommandOutput> {
  const profile = resolveProfile(ctx, args.profile);
  const server = resolveLoginServer(ctx, args.server, profile);
  const lang = language(ctx);
  const client = new Client({ server, lang });
  const sleep = args.sleep ?? defaultSleep;

  const started = await client.call<DeviceCodeInfo>({
    method: "POST",
    path: "/auth/device/code",
    body: {
      client_name: DEVICE_CLIENT_NAME,
      scopes: args.scopes && args.scopes.length > 0 ? args.scopes : DEFAULT_CLI_SCOPES,
    },
  });
  const code = started.data;
  args.onCode(code);

  let interval = Math.max(code.interval, 1) * 1000;
  const deadline = Date.now() + Math.max(code.expires_in, 1) * 1000;

  for (;;) {
    if (Date.now() > deadline) {
      throw new Ech0Error("设备码已过期，请重新登录");
    }
    await sleep(interval);
    try {
      const granted = await client.call<{ access_token?: string }>({
        method: "POST",
        path: "/auth/device/token",
        body: { device_code: code.device_code },
      });
      const token = granted.data.access_token ?? "";
      if (token === "") throw new Ech0Error("授权响应缺少 access_token");

      const authed = new Client({ server, token, lang });
      const user = (await authed.call<User>({ path: "/user" })).data;
      const source = await persistLogin(ctx, { profile, server, token, user });
      return {
        human: loginMessage("设备授权令牌（1 个月）", user, server, profile, source),
        json: granted.text,
      };
    } catch (error) {
      if (!(error instanceof Ech0Error)) throw error;
      switch (error.details.errorCode) {
        case "DEVICE_AUTHORIZATION_PENDING":
          continue;
        case "DEVICE_SLOW_DOWN":
          interval += 5000;
          continue;
        case "DEVICE_EXPIRED":
          throw new Ech0Error("设备码已过期，请重新登录");
        case "DEVICE_ACCESS_DENIED":
          throw new Ech0Error("设备授权已被拒绝");
        default:
          throw error;
      }
    }
  }
}

export interface StatusArgs {
  verify: boolean;
}

export async function status(ctx: CliContext, args: StatusArgs): Promise<CommandOutput> {
  const target = resolveTarget(ctx);
  const stored = await readToken(target.profile, ctx.credentialsPath);
  const cookie = await readRefreshCookie(target.profile, ctx.credentialsPath);
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
    refresh_cookie: cookie?.source ?? null,
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
        `${bold("refresh")}    ${cookie?.source ?? "-"}`,
        `${bold("user")}       ${cached?.user ?? "-"}`,
        `${bold("admin")}      ${cached?.isAdmin ? "yes" : "no"}`,
      ].join("\n"),
      json: JSON.stringify(payload, null, 2),
    };
  }

  const token = ctx.flags.token ?? ctx.env.ECH0_TOKEN ?? stored?.value;
  if (!token) throw new UsageError(`profile "${target.profile}" 没有可用 token`);
  const client = new Client({ server: target.server, token, lang: language(ctx) });
  const result = await client.call<User>({ path: "/user" });
  return {
    human: JSON.stringify(result.data, null, 2),
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
      `${green("✓")} 已清除 profile "${target.profile}" 的凭据` +
      (args.purge ? dim("（profile 已删除）") : ""),
    json: JSON.stringify(payload, null, 2),
  };
}
