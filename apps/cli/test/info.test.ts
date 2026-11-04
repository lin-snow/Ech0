// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { fingerprintOf } from "../src/catalog/store.ts";
import { info } from "../src/commands/info.ts";
import type { CliContext } from "../src/session.ts";
import { startFakeApi, type RouteReply } from "./helpers.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

const CATALOG = {
  ops: {
    "echo.list": {
      summary: "分页获取 Echo",
      method: "GET",
      path: "/echo",
      scopes: ["echo:read"],
      effect: "read",
      kind: "json",
      input: { type: "object", properties: {} },
      bind: {},
      deprecated: false,
    },
  },
};

function catalogFingerprint(): string {
  return fingerprintOf(new TextEncoder().encode(JSON.stringify(CATALOG)));
}

function routes(extra: Record<string, () => RouteReply> = {}): Record<string, () => RouteReply> {
  const fingerprint = catalogFingerprint();
  return {
    "GET /api/hello": () => ({
      body: {
        code: 1,
        data: {
          hello: "Hello, Ech0! 👋",
          version: "v1.4.2",
          commit: "abcdef123456",
          build_time: "2026-10-01T00:00:00Z",
          license: "AGPL-3.0-or-later",
          repo_url: "https://github.com/lin-snow/Ech0",
        },
      },
      headers: { "X-Ech0-Catalog": fingerprint },
    }),
    "GET /api/connect": () => ({
      body: {
        code: 1,
        data: {
          server_name: "memo",
          server_url: "https://memo.example",
          total_echos: 12,
          today_echos: 3,
          sys_username: "lin-snow",
          version: "v1.4.2",
        },
      },
      headers: { "X-Ech0-Catalog": fingerprint },
    }),
    "GET /api/settings": () => ({
      body: {
        code: 1,
        data: {
          site_title: "MyEch0",
          server_name: "memo",
          server_url: "https://memo.example",
          allow_register: true,
          default_locale: "zh-CN",
        },
      },
      headers: { "X-Ech0-Catalog": fingerprint },
    }),
    "GET /api/init/status": () => ({
      body: { code: 1, data: { initialized: true, owner_exists: true } },
      headers: { "X-Ech0-Catalog": fingerprint },
    }),
    "GET /api/_catalog": () => ({ body: CATALOG, headers: { "X-Ech0-Catalog": fingerprint } }),
    ...extra,
  };
}

function context(server: string | undefined, dir: string): CliContext {
  return {
    flags: { json: false, yaml: false, server },
    configPath: join(dir, "config.json"),
    credentialsPath: join(dir, "credentials.json"),
    config: server === undefined ? { profiles: {} } : { default: "test", profiles: { test: { server } } },
    env: {},
  };
}

describe("info", () => {
  test("aggregates the public endpoints without a token", async () => {
    const api = startFakeApi(routes());
    try {
      const dir = mkdtempSync(join(tmpdir(), "ech0-info-"));
      const result = await info(context(api.baseUrl, dir));
      const payload = JSON.parse(result.json) as {
        server: string;
        reachable: boolean;
        instance: Record<string, unknown>;
        auth: Record<string, unknown>;
        catalog: { fingerprint: string; ops: number; stale: boolean } | null;
        errors: string[];
      };

      expect(payload.server).toBe(api.baseUrl);
      expect(payload.reachable).toBe(true);
      expect(payload.instance.site_title).toBe("MyEch0");
      expect(payload.instance.version).toBe("v1.4.2");
      expect(payload.instance.commit).toBe("abcdef123456");
      expect(payload.instance.total_echos).toBe(12);
      expect(payload.instance.initialized).toBe(true);
      expect(payload.instance.allow_register).toBe(true);
      expect(payload.auth.user).toBeNull();
      expect(payload.auth.valid).toBeNull();
      expect(payload.catalog?.ops).toBe(1);
      expect(payload.catalog?.fingerprint).toBe(catalogFingerprint());
      expect(payload.catalog?.stale).toBe(false);
      expect(payload.errors).toEqual([]);
      expect(api.calls.map((call) => call.path)).not.toContain("/api/user");

      expect(result.human).toContain(api.baseUrl);
      expect(result.human).toContain("MyEch0");
      expect(result.human).toContain("v1.4.2");
      expect(result.human).toContain("12（今天 3）");
      expect(result.human).toContain("initialized");
    } finally {
      api.stop();
    }
  });

  test("flags a stale catalog cache against the server fingerprint", async () => {
    const fingerprint = "f".repeat(64);
    const api = startFakeApi(routes({ "GET /api/hello": () => ({ body: { code: 1, data: {} }, headers: { "X-Ech0-Catalog": fingerprint } }) }));
    try {
      const dir = mkdtempSync(join(tmpdir(), "ech0-info-"));
      const result = await info(context(api.baseUrl, dir));
      const payload = JSON.parse(result.json) as {
        catalog: { stale: boolean; server_fingerprint: string | null } | null;
      };

      expect(payload.catalog?.stale).toBe(true);
      expect(payload.catalog?.server_fingerprint).toBe(fingerprint);
      expect(result.human).toContain("ech0 catalog refresh");
    } finally {
      api.stop();
    }
  });

  test("reports an unreachable instance instead of failing", async () => {
    const api = startFakeApi({});
    try {
      const dir = mkdtempSync(join(tmpdir(), "ech0-info-"));
      const result = await info(context(api.baseUrl, dir));
      const payload = JSON.parse(result.json) as { reachable: boolean; errors: string[] };

      expect(payload.reachable).toBe(false);
      expect(payload.errors.length).toBeGreaterThan(0);
      expect(result.human).toContain("不可达");
    } finally {
      api.stop();
    }
  });

  test("prints local state when no instance is configured", async () => {
    const dir = mkdtempSync(join(tmpdir(), "ech0-info-"));
    const result = await info(context(undefined, dir));
    const payload = JSON.parse(result.json) as { server: null; reachable: null; errors: string[] };

    expect(payload.server).toBeNull();
    expect(payload.reachable).toBeNull();
    expect(payload.errors.join("\n")).toContain("未配置实例");
  });
});
