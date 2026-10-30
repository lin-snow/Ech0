// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import * as auth from "../src/commands/auth.ts";
import { readRefreshCookie, readToken } from "../src/config/credentials.ts";
import { loadConfig } from "../src/config/store.ts";
import { loadContext, openSession } from "../src/session.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

function context(dir: string) {
  const env = { ...process.env, ECH0_CONFIG: join(dir, "config.json"), ECH0_NO_KEYCHAIN: "1" };
  return loadContext({ json: false, yaml: false }, env);
}

function refreshCookie(value: string): Record<string, string> {
  return { "Set-Cookie": `ech0_refresh_token=${value}; Path=/api/auth; HttpOnly` };
}

describe("password login", () => {
  test("admin mints a cli access token and drops the cookie", async () => {
    const fake = startFakeApi({
      "POST /api/login": () => ({
        body: { code: 1, msg: "登陆成功", data: { access_token: "session-token", expires_in: 3600 } },
        headers: refreshCookie("refresh-1"),
      }),
      "GET /api/user": () => ({
        body: { code: 1, msg: "ok", data: { id: "u1", username: "owner", is_admin: true } },
      }),
      "POST /api/access-tokens": (request) => ({
        body: { code: 1, msg: "ok", data: "minted-token" },
        headers: { "X-Seen": JSON.stringify(request.body) },
      }),
    });
    api = fake;
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-pw-"));
    const ctx = context(dir);

    const output = await auth.loginWithPassword(ctx, {
      server: fake.baseUrl,
      username: "owner",
      password: "secret",
      profile: "memo",
    });

    expect(output.human).toContain("CLI 令牌");
    expect((await readToken("memo", ctx.credentialsPath))?.value).toBe("minted-token");
    expect(await readRefreshCookie("memo", ctx.credentialsPath)).toBeUndefined();
    expect(loadConfig(join(dir, "config.json")).profiles.memo?.isAdmin).toBe(true);

    const mint = fake.calls.find((call) => call.path === "/api/access-tokens");
    expect(mint?.headers.get("Authorization")).toBe("Bearer session-token");
    expect(mint?.body).toMatchObject({ expiry: "1_month", audience: "cli", scopes: ["echo:read", "echo:write", "profile:read", "file:read", "file:write"] });
  });

  test("non-admin keeps the session token and the refresh cookie", async () => {
    const fake = startFakeApi({
      "POST /api/login": () => ({
        body: { code: 1, msg: "登陆成功", data: { access_token: "session-token", expires_in: 3600 } },
        headers: refreshCookie("refresh-2"),
      }),
      "GET /api/user": () => ({
        body: { code: 1, msg: "ok", data: { id: "u2", username: "reader", is_admin: false } },
      }),
    });
    api = fake;
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-pw2-"));
    const ctx = context(dir);

    await auth.loginWithPassword(ctx, {
      server: fake.baseUrl,
      username: "reader",
      password: "secret",
      profile: "memo",
    });

    expect((await readToken("memo", ctx.credentialsPath))?.value).toBe("session-token");
    expect((await readRefreshCookie("memo", ctx.credentialsPath))?.value).toBe("refresh-2");
    expect(fake.calls.some((call) => call.path === "/api/access-tokens")).toBe(false);
  });
});

describe("device login", () => {
  test("polls until approved and stores the token", async () => {
    let polls = 0;
    const fake = startFakeApi({
      "POST /api/auth/device/code": (request) => ({
        body: {
          code: 1,
          msg: "ok",
          data: {
            device_code: "device-code",
            user_code: "ABCD-EFGH",
            verification_uri: "/device",
            verification_uri_complete: "/device?code=ABCD-EFGH",
            expires_in: 60,
            interval: 1,
          },
        },
        headers: { "X-Seen": JSON.stringify(request.body) },
      }),
      "POST /api/auth/device/token": () => {
        polls += 1;
        if (polls === 1) {
          return { status: 400, body: { code: 0, msg: "pending", error_code: "DEVICE_AUTHORIZATION_PENDING" } };
        }
        return { body: { code: 1, msg: "ok", data: { access_token: "device-token" } } };
      },
      "GET /api/user": () => ({
        body: { code: 1, msg: "ok", data: { id: "u1", username: "owner", is_admin: true } },
      }),
    });
    api = fake;
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-device-"));
    const ctx = context(dir);

    const delays: number[] = [];
    let shown: auth.DeviceCodeInfo | undefined;

    const output = await auth.loginWithDevice(ctx, {
      server: fake.baseUrl,
      profile: "memo",
      onCode: (code) => {
        shown = code;
      },
      sleep: async (ms) => {
        delays.push(ms);
      },
    });

    expect(shown?.user_code).toBe("ABCD-EFGH");
    expect(delays).toEqual([1000, 1000]);
    expect(output.human).toContain("设备授权");
    expect((await readToken("memo", ctx.credentialsPath))?.value).toBe("device-token");

    const start = fake.calls.find((call) => call.path === "/api/auth/device/code");
    expect(start?.body).toMatchObject({ client_name: "ech0-cli", scopes: ["echo:read", "echo:write", "profile:read", "file:read", "file:write"] });
  });
});

describe("session refresh", () => {
  test("rotates the access token and the cookie from the refresh cookie", async () => {
    const fake = startFakeApi({
      "POST /api/auth/refresh": () => ({
        body: { code: 1, msg: "ok", data: { access_token: "fresh-token", expires_in: 3600 } },
        headers: refreshCookie("refresh-3"),
      }),
      "GET /api/user": (request) => ({
        body: { code: 1, msg: "ok", data: { username: "owner", auth: request.headers.get("Authorization") } },
      }),
    });
    api = fake;
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-refresh-"));
    const ctx = context(dir);

    // Seed a profile + stored credentials.
    const { writeToken, writeRefreshCookie } = await import("../src/config/credentials.ts");
    const { saveConfig, withProfile } = await import("../src/config/store.ts");
    await writeToken("memo", "stale-token", ctx.credentialsPath);
    await writeRefreshCookie("memo", "refresh-2", ctx.credentialsPath);
    saveConfig(withProfile(ctx.config, "memo", { server: fake.baseUrl }), ctx.configPath);

    const session = await openSession(loadContext({ json: false, yaml: false, profile: "memo" }, ctx.env));
    expect(await session.refresh()).toBe(true);

    const result = await session.client.call<{ auth?: string }>({ path: "/user" });
    expect(result.data.auth).toBe("Bearer fresh-token");
    expect((await readRefreshCookie("memo", ctx.credentialsPath))?.value).toBe("refresh-3");
  });
});
