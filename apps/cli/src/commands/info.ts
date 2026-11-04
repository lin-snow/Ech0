// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { statSync } from "node:fs";

import { Client } from "../api/http.ts";
import { loadCatalog } from "../catalog/load.ts";
import { CATALOG_HEADER } from "../catalog/store.ts";
import { readRefreshCookie, readToken } from "../config/credentials.ts";
import { activeProfileName } from "../config/store.ts";
import { colors, Ech0Error } from "../output/console.ts";
import { displayWidth } from "../output/render.ts";
import type { CliContext } from "../session.ts";
import { CLI_VERSION } from "../version-info.ts";
import type { CommandOutput } from "./types.ts";

/** One probe is a diagnostic: a slow endpoint must not hang the whole command. */
const PROBE_TIMEOUT_MS = 5_000;

/** Shapes of the public endpoints this command composes (see `internal/handler`). */
interface HelloDto {
  version?: string;
  commit?: string;
  build_time?: string;
  license?: string;
  repo_url?: string;
}

interface SelfCardDto {
  server_name?: string;
  server_url?: string;
  logo?: string;
  total_echos?: number;
  today_echos?: number;
  sys_username?: string;
  version?: string;
}

interface SettingsDto {
  site_title?: string;
  server_name?: string;
  server_url?: string;
  allow_register?: boolean;
  default_locale?: string;
}

interface InitStatusDto {
  initialized?: boolean;
  owner_exists?: boolean;
}

interface UserDto {
  username?: string;
  is_admin?: boolean;
  is_owner?: boolean;
}

interface CatalogInfo {
  fingerprint: string;
  ops: number;
  source: "cache" | "network";
  age_seconds: number | null;
  server_fingerprint: string | null;
  stale: boolean;
}

type Probe<T> = { ok: true; data: T; catalog?: string } | { ok: false; error: string; status?: number };

async function get<T>(client: Client, path: string): Promise<Probe<T>> {
  try {
    const result = await client.call<T>({ path, signal: AbortSignal.timeout(PROBE_TIMEOUT_MS) });
    const fingerprint = result.headers.get(CATALOG_HEADER)?.trim();
    return { ok: true, data: result.data, catalog: fingerprint === "" ? undefined : fingerprint };
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : String(error),
      status: error instanceof Ech0Error ? error.details.status : undefined,
    };
  }
}

function value<T>(probe: Probe<T> | undefined): T | undefined {
  return probe?.ok === true ? probe.data : undefined;
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function humanAge(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3_600) return `${Math.round(seconds / 60)}m`;
  if (seconds < 86_400) return `${Math.round(seconds / 3_600)}h`;
  return `${Math.round(seconds / 86_400)}d`;
}

function catalogLine(catalog: CatalogInfo | null): string {
  if (catalog === null) return "-";
  const parts = [`${catalog.fingerprint.slice(0, 12)}…`, `${catalog.ops} ops`];
  parts.push(catalog.source === "cache" && catalog.age_seconds !== null ? `cache ${humanAge(catalog.age_seconds)}` : catalog.source);
  if (catalog.stale && catalog.server_fingerprint !== null) {
    parts.push(`服务端 ${catalog.server_fingerprint.slice(0, 12)}…`);
  }
  return parts.join(" · ");
}

/** Two-column block; keys are ASCII, but values may be CJK, so pad by display width. */
function keyValues(rows: Array<[string, string]>): string {
  const width = Math.max(...rows.map(([key]) => displayWidth(key)));
  const { bold } = colors();
  return rows
    .map(([key, line]) => `${bold(key + " ".repeat(Math.max(0, width - displayWidth(key))))}  ${line}`)
    .join("\n");
}

function versionLine(hello: HelloDto | undefined, self: SelfCardDto | undefined): string {
  const parts: string[] = [];
  const version = hello?.version ?? self?.version;
  if (version !== undefined && version !== "") parts.push(version);
  if (hello?.commit !== undefined && hello.commit !== "" && hello.commit !== "unknown") {
    parts.push(hello.commit.slice(0, 7));
  }
  if (hello?.build_time !== undefined && hello.build_time !== "") parts.push(hello.build_time);
  return parts.length === 0 ? "-" : parts.join(" · ");
}

/**
 * `ech0 info` — one read-only view of the instance this profile points at.
 *
 * Every field comes from a public endpoint (`/hello`, `/connect`, `/settings`,
 * `/init/status`) plus local profile state, so the command needs no token and
 * no scope. A token, when present, adds a `/user` check; a failing probe is
 * reported rather than fatal, because the point is to diagnose the instance.
 */
export async function info(ctx: CliContext): Promise<CommandOutput> {
  const profile = activeProfileName(ctx.config, ctx.flags.profile, ctx.env);
  const configured = profile === undefined ? undefined : ctx.config.profiles[profile];
  const server = ctx.flags.server ?? ctx.env.ECH0_SERVER ?? configured?.server;
  const lang = ctx.flags.lang ?? ctx.env.ECH0_LANG;

  const errors: string[] = [];
  const notes: string[] = [];

  let storedToken: string | undefined;
  let tokenSource: "flag" | "env" | "keychain" | "file" | undefined;
  if (ctx.flags.token !== undefined && ctx.flags.token !== "") {
    storedToken = ctx.flags.token;
    tokenSource = "flag";
  } else if (ctx.env.ECH0_TOKEN !== undefined && ctx.env.ECH0_TOKEN !== "") {
    storedToken = ctx.env.ECH0_TOKEN;
    tokenSource = "env";
  } else if (profile !== undefined) {
    try {
      const stored = await readToken(profile, ctx.credentialsPath);
      storedToken = stored?.value;
      tokenSource = stored?.source;
    } catch (error) {
      errors.push(`凭据: ${message(error)}`);
    }
  }

  let refreshSource: string | undefined;
  if (profile !== undefined) {
    try {
      refreshSource = (await readRefreshCookie(profile, ctx.credentialsPath))?.source;
    } catch (error) {
      errors.push(`refresh cookie: ${message(error)}`);
    }
  }

  let hello: HelloDto | undefined;
  let self: SelfCardDto | undefined;
  let site: SettingsDto | undefined;
  let init: InitStatusDto | undefined;
  let user: UserDto | undefined;
  let reachable: boolean | null = null;
  let latencyMs: number | null = null;
  let serverFingerprint: string | null = null;
  let authRejected = false;

  if (server === undefined) {
    errors.push("未配置实例：用 --server / ECH0_SERVER 指定，或先运行 `ech0 auth login --server <url> --device`");
  } else {
    const client = new Client({ server, token: storedToken, lang });
    const started = Date.now();
    const [helloProbe, selfProbe, siteProbe, initProbe, userProbe] = await Promise.all([
      get<HelloDto>(client, "/hello"),
      get<SelfCardDto>(client, "/connect"),
      get<SettingsDto>(client, "/settings"),
      get<InitStatusDto>(client, "/init/status"),
      storedToken === undefined ? Promise.resolve(undefined) : get<UserDto>(client, "/user"),
    ]);
    latencyMs = Date.now() - started;

    const probes: Probe<unknown>[] = [helloProbe, selfProbe, siteProbe, initProbe];
    if (userProbe !== undefined) probes.push(userProbe);
    reachable = probes.some((probe) => probe.ok);

    hello = value(helloProbe);
    self = value(selfProbe);
    site = value(siteProbe);
    init = value(initProbe);
    user = value(userProbe);

    for (const probe of probes) {
      if (probe.ok) {
        serverFingerprint ??= probe.catalog ?? null;
      }
    }
    const labels: Array<[string, Probe<unknown>]> = [
      ["hello", helloProbe],
      ["connect", selfProbe],
      ["settings", siteProbe],
      ["init", initProbe],
    ];
    if (userProbe !== undefined) labels.push(["auth", userProbe]);
    if (!reachable) {
      const first = probes.find((probe) => !probe.ok);
      errors.push(`实例不可达：${first?.ok === false ? first.error : "所有探测请求都失败"}`);
    } else {
      for (const [label, probe] of labels) {
        if (probe.ok) continue;
        // `/connect` needs an owner to build the self card: a fresh instance
        // answering 404 there is "not initialized yet", not a broken instance.
        if (label === "connect" && init !== undefined && init.initialized === false) continue;
        errors.push(`${label}: ${probe.error}`);
      }
    }
    authRejected = userProbe?.ok === false && (userProbe.status === 401 || userProbe.status === 403);
  }

  let catalog: CatalogInfo | null = null;
  if (server !== undefined) {
    try {
      const loaded = await loadCatalog({
        server,
        token: storedToken,
        lang,
        configPath: ctx.configPath,
      });
      let age: number | null = 0;
      if (loaded.source === "cache") {
        try {
          age = Math.max(0, Math.round((Date.now() - statSync(loaded.file).mtimeMs) / 1000));
        } catch {
          age = null;
        }
      }
      catalog = {
        fingerprint: loaded.fingerprint,
        ops: Object.keys(loaded.doc.ops).length,
        source: loaded.source,
        age_seconds: age,
        server_fingerprint: serverFingerprint,
        stale: serverFingerprint !== null && serverFingerprint !== loaded.fingerprint,
      };
      if (catalog.stale) {
        notes.push("catalog 缓存落后于服务端，运行 `ech0 catalog refresh` 更新命令树");
      }
    } catch (error) {
      errors.push(`catalog: ${message(error)}`);
    }
  }

  const tokenValid = storedToken === undefined ? null : user !== undefined ? true : authRejected ? false : null;
  const payload = {
    cli: { version: CLI_VERSION },
    profile: profile ?? null,
    server: server ?? null,
    reachable,
    latency_ms: latencyMs,
    instance: {
      site_title: site?.site_title ?? null,
      server_name: self?.server_name ?? site?.server_name ?? null,
      server_url: self?.server_url ?? site?.server_url ?? null,
      logo: self?.logo ?? null,
      version: hello?.version ?? self?.version ?? null,
      commit: hello?.commit ?? null,
      build_time: hello?.build_time ?? null,
      license: hello?.license ?? null,
      repo_url: hello?.repo_url ?? null,
      initialized: init?.initialized ?? null,
      owner_exists: init?.owner_exists ?? null,
      allow_register: site?.allow_register ?? null,
      default_locale: site?.default_locale ?? null,
      total_echos: self?.total_echos ?? null,
      today_echos: self?.today_echos ?? null,
    },
    auth: {
      token_source: tokenSource ?? null,
      refresh_cookie: refreshSource ?? null,
      user: user?.username ?? configured?.user ?? null,
      is_admin: user?.is_admin ?? configured?.isAdmin ?? null,
      valid: tokenValid,
    },
    catalog,
    errors,
  };

  const rows: Array<[string, string]> = [["cli", CLI_VERSION], ["profile", profile ?? "-"]];
  if (server !== undefined) {
    const state = reachable === true ? ` · ${latencyMs}ms` : reachable === false ? " · 不可达" : "";
    rows.push(["server", `${server}${state}`]);
  } else {
    rows.push(["server", "-"]);
  }
  const siteTitle = site?.site_title ?? undefined;
  const siteName = self?.server_name ?? site?.server_name ?? undefined;
  rows.push([
    "site",
    siteTitle === undefined
      ? (siteName ?? "-")
      : siteName !== undefined && siteName !== siteTitle
        ? `${siteTitle} · ${siteName}`
        : siteTitle,
  ]);
  rows.push(["version", versionLine(hello, self)]);
  if (self?.total_echos !== undefined) {
    rows.push(["echos", `${self.total_echos}（今天 ${self.today_echos ?? 0}）`]);
  }
  if (init !== undefined) {
    rows.push(["initialized", `${init.initialized === true ? "yes" : "no"}${init.owner_exists === true ? " · owner 已创建" : " · 无 owner"}`]);
  }
  if (site !== undefined) {
    rows.push([
      "register",
      `${site.allow_register === true ? "开放" : "关闭"} · ${site.default_locale ?? "-"}`,
    ]);
  }
  const username = user?.username ?? configured?.user;
  const isAdmin = user?.is_admin ?? configured?.isAdmin;
  rows.push([
    "user",
    username === undefined ? "-" : `${username}${isAdmin === true ? "（admin）" : ""}`,
  ]);
  const tokenParts = [tokenSource ?? "none"];
  if (refreshSource !== undefined) tokenParts.push(`refresh ${refreshSource}`);
  if (tokenValid === true) tokenParts.push("有效");
  if (tokenValid === false) tokenParts.push("已失效");
  if (tokenValid === null && storedToken !== undefined) tokenParts.push("未验证");
  rows.push(["token", tokenParts.join(" · ")]);
  rows.push(["catalog", catalogLine(catalog)]);

  const { dim, yellow } = colors();
  const blocks = [keyValues(rows)];
  if (errors.length > 0) blocks.push("", ...errors.map((error) => yellow(`! ${error}`)));
  if (notes.length > 0) blocks.push("", ...notes.map((note) => dim(`· ${note}`)));

  return { human: blocks.join("\n"), json: JSON.stringify(payload, null, 2) };
}
