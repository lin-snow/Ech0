// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Command } from "commander";

import { Client } from "./api/http.ts";
import { credentialsFilePath, configFilePath } from "./config/paths.ts";
import {
  REFRESH_COOKIE_NAME,
  readRefreshCookie,
  readToken,
  writeRefreshCookie,
  type TokenSource,
} from "./config/credentials.ts";
import { activeProfileName, loadConfig, type Config } from "./config/store.ts";
import { UsageError } from "./output/console.ts";

/** Flags accepted by every command (declared on the root command). */
export interface GlobalFlags {
  profile?: string;
  server?: string;
  token?: string;
  lang?: string;
  json: boolean;
  yaml: boolean;
}

export interface CliContext {
  flags: GlobalFlags;
  configPath: string;
  credentialsPath: string;
  config: Config;
  env: NodeJS.ProcessEnv;
}

export interface ResolvedTarget {
  profile: string;
  server: string;
  config: Config;
}

export interface Session extends ResolvedTarget {
  token: string;
  tokenSource: "flag" | "env" | TokenSource;
  client: Client;
  /** Rotates the access token from the stored refresh cookie; false when it cannot. */
  refresh: () => Promise<boolean>;
}

/** Reads the global flags declared on the root command, from any subcommand. */
export function readGlobalFlags(command: Command): GlobalFlags {
  const options = command.optsWithGlobals();
  const json = options.json === true;
  const yaml = options.yaml === true;
  if (json && yaml) throw new UsageError("--json 与 --yaml 只能选一个");
  return {
    profile: typeof options.profile === "string" ? options.profile : undefined,
    server: typeof options.server === "string" ? options.server : undefined,
    token: typeof options.token === "string" ? options.token : undefined,
    lang: typeof options.lang === "string" ? options.lang : undefined,
    json,
    yaml,
  };
}

export function loadContext(
  flags: GlobalFlags,
  env: NodeJS.ProcessEnv = process.env,
): CliContext {
  const configPath = configFilePath(env);
  return {
    flags,
    configPath,
    credentialsPath: credentialsFilePath(env),
    config: loadConfig(configPath),
    env,
  };
}

/** Resolves the profile name + instance URL without requiring a token. */
export function resolveTarget(ctx: CliContext): ResolvedTarget {
  const profile = activeProfileName(ctx.config, ctx.flags.profile, ctx.env);
  if (!profile) {
    throw new UsageError("未配置任何 profile：先运行 `ech0 auth login --server <url> --token <token>`");
  }
  const server = ctx.flags.server ?? ctx.env.ECH0_SERVER ?? ctx.config.profiles[profile]?.server;
  if (!server) {
    throw new UsageError(`profile "${profile}" 缺少 server：用 --server 指定或重新登录`);
  }
  return { profile, server, config: ctx.config };
}

/** Extracts one cookie value from a response's Set-Cookie headers. */
export function setCookieValue(headers: Headers, name: string): string | undefined {
  const cookies = typeof headers.getSetCookie === "function" ? headers.getSetCookie() : [];
  for (const cookie of cookies) {
    const [pair = ""] = cookie.split(";");
    const separator = pair.indexOf("=");
    if (separator <= 0) continue;
    if (pair.slice(0, separator).trim() === name) {
      return pair.slice(separator + 1);
    }
  }
  return undefined;
}

interface RefreshedToken {
  token: string;
}

async function refreshAccessToken(
  ctx: CliContext,
  profile: string,
  server: string,
): Promise<RefreshedToken | undefined> {
  const stored = await readRefreshCookie(profile, ctx.credentialsPath);
  if (!stored) return undefined;

  const client = new Client({ server, lang: ctx.flags.lang ?? ctx.env.ECH0_LANG });
  try {
    const result = await client.call<{ access_token?: string }>({
      method: "POST",
      path: "/auth/refresh",
      headers: { Cookie: `${REFRESH_COOKIE_NAME}=${stored.value}` },
    });
    const token = typeof result.data?.access_token === "string" ? result.data.access_token : "";
    if (token === "") return undefined;

    const rotated = setCookieValue(result.headers, REFRESH_COOKIE_NAME);
    if (rotated) {
      await writeRefreshCookie(profile, rotated, ctx.credentialsPath);
    }
    return { token };
  } catch {
    return undefined;
  }
}

/** Resolves profile + server + token and builds a request client. */
export async function openSession(ctx: CliContext): Promise<Session> {
  const target = resolveTarget(ctx);
  const lang = ctx.flags.lang ?? ctx.env.ECH0_LANG;

  let token: string;
  let tokenSource: Session["tokenSource"];
  if (ctx.flags.token) {
    token = ctx.flags.token;
    tokenSource = "flag";
  } else if (ctx.env.ECH0_TOKEN) {
    token = ctx.env.ECH0_TOKEN;
    tokenSource = "env";
  } else {
    const stored = await readToken(target.profile, ctx.credentialsPath);
    if (!stored) {
      throw new UsageError(`profile "${target.profile}" 没有可用 token：先运行 \`ech0 auth login\``);
    }
    token = stored.value;
    tokenSource = stored.source;
  }

  const session: Session = {
    ...target,
    token,
    tokenSource,
    client: new Client({ server: target.server, token, lang }),
    refresh: async () => false,
  };

  // Explicit flags/env tokens are the caller's to manage; only stored
  // credentials can be renewed from the refresh cookie.
  if (!ctx.flags.token && !ctx.env.ECH0_TOKEN) {
    session.refresh = async () => {
      const refreshed = await refreshAccessToken(ctx, target.profile, target.server);
      if (!refreshed) return false;
      session.token = refreshed.token;
      session.client.setToken(refreshed.token);
      return true;
    };
  }

  return session;
}
