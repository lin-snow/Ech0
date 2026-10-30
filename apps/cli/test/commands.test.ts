// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import * as auth from "../src/commands/auth.ts";
import { readToken } from "../src/config/credentials.ts";
import { loadConfig } from "../src/config/store.ts";
import { UsageError } from "../src/output/console.ts";
import { loadContext, openSession } from "../src/session.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

describe("auth login", () => {
  test("verifies the token, stores profile and token", async () => {
    api = startFakeApi({
      "GET /api/user": () => ({
        body: { code: 1, msg: "ok", data: { id: "u1", username: "owner", is_admin: true, is_owner: true } },
      }),
    });
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-login-"));
    const env = { ...process.env, ECH0_CONFIG: join(dir, "config.json"), ECH0_NO_KEYCHAIN: "1" };
    const ctx = loadContext({ json: false, yaml: false }, env);

    const output = await auth.login(ctx, { server: api.baseUrl, token: "tok-login", profile: "memo" });

    expect(api.calls[0]?.headers.get("Authorization")).toBe("Bearer tok-login");
    expect(output.human).toContain("已登录 owner");

    const saved = loadConfig(join(dir, "config.json"));
    expect(saved.default).toBe("memo");
    expect(saved.profiles.memo?.user).toBe("owner");
    expect(saved.profiles.memo?.isAdmin).toBe(true);
    expect((await readToken("memo", ctx.credentialsPath))?.value).toBe("tok-login");
  });
});

describe("session", () => {
  test("env token wins over the stored token", async () => {
    const dir = mkdtempSync(join(tmpdir(), "ech0-cli-session-"));
    const env = {
      ...process.env,
      ECH0_CONFIG: join(dir, "config.json"),
      ECH0_SERVER: "https://env.test",
      ECH0_TOKEN: "env-token",
      ECH0_NO_KEYCHAIN: "1",
    };
    const ctx = loadContext({ json: false, yaml: false, profile: "memo" }, env);
    const opened = await openSession(ctx);

    expect(opened.server).toBe("https://env.test");
    expect(opened.token).toBe("env-token");
    expect(opened.tokenSource).toBe("env");
  });

  test("missing profile is a usage error", async () => {
    const ctx = loadContext({ json: false, yaml: false }, { ECH0_CONFIG: join(mkdtempSync(join(tmpdir(), "e-")), "c.json") });
    await expect(openSession(ctx)).rejects.toBeInstanceOf(UsageError);
  });
});
