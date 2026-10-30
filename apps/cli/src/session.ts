// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Command } from "commander";

import { Client } from "./api/http.ts";
import { credentialsFilePath, configFilePath } from "./config/paths.ts";
import { readToken, type TokenSource } from "./config/credentials.ts";
import { activeProfileName, loadConfig, type Config } from "./config/store.ts";
import { UsageError } from "./output/console.ts";

/** Flags accepted by every command (declared on the root command). */
export interface GlobalFlags {
  profile?: string;
  server?: string;
  token?: string;
  lang?: string;
  json: boolean;
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
}

/** Reads the global flags declared on the root command, from any subcommand. */
export function readGlobalFlags(command: Command): GlobalFlags {
  const options = command.optsWithGlobals();
  return {
    profile: typeof options.profile === "string" ? options.profile : undefined,
    server: typeof options.server === "string" ? options.server : undefined,
    token: typeof options.token === "string" ? options.token : undefined,
    lang: typeof options.lang === "string" ? options.lang : undefined,
    json: options.json === true,
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
    token = stored.token;
    tokenSource = stored.source;
  }

  return {
    ...target,
    token,
    tokenSource,
    client: new Client({ server: target.server, token, lang }),
  };
}
