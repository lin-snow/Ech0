// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { Client } from "../src/api/http.ts";
import * as auth from "../src/commands/auth.ts";
import * as echo from "../src/commands/echo.ts";
import { readToken } from "../src/config/credentials.ts";
import { loadConfig } from "../src/config/store.ts";
import { UsageError } from "../src/output/console.ts";
import { loadContext, openSession, type Session } from "../src/session.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

const SAMPLE = {
  id: "11111111-1111-7111-8111-111111111111",
  content: "hello world\nsecond line",
  created_at: 1_700_000_000,
  private: false,
  username: "owner",
  fav_count: 2,
  tags: [{ name: "dev" }],
};

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

function session(server: string, token = "tok"): Session {
  return {
    profile: "test",
    server,
    config: { profiles: { test: { server } } },
    token,
    tokenSource: "flag",
    client: new Client({ server, token }),
  };
}

describe("echo commands", () => {
  test("add posts content, visibility and tags", async () => {
    api = startFakeApi({
      "POST /api/echo": () => ({ body: { code: 1, msg: "ok", data: null } }),
    });
    const output = await echo.add(session(api.baseUrl), {
      content: "hi there",
      private: true,
      tags: ["dev", "go"],
    });

    expect(api.calls[0]?.body).toEqual({
      content: "hi there",
      private: true,
      tags: [{ name: "dev" }, { name: "go" }],
    });
    expect(output.human).toContain("已发布");
    expect(output.json).toContain('"code":1');
  });

  test("add rejects empty content", async () => {
    api = startFakeApi({});
    await expect(echo.add(session(api.baseUrl), { content: "  ", private: false })).rejects.toBeInstanceOf(
      UsageError,
    );
  });

  test("list renders a table and keeps the raw envelope for --json", async () => {
    const envelope = { code: 1, msg: "ok", data: { items: [SAMPLE], total: 1 } };
    api = startFakeApi({ "POST /api/echo/query": () => ({ body: envelope }) });

    const output = await echo.list(session(api.baseUrl), { limit: 10, page: 1, all: false });

    expect(api.calls[0]?.body).toEqual({ page: 1, pageSize: 10, search: "" });
    expect(output.human).toContain("hello world");
    expect(output.human).toContain("11111111");
    expect(output.human).toContain("共 1 条");
    expect(output.json).toBe(JSON.stringify(envelope));
  });

  test("list --all walks every page", async () => {
    api = startFakeApi({
      "POST /api/echo/query": (request) => {
        const body = request.body as { page?: number };
        const page = body.page ?? 1;
        return {
          body: {
            code: 1,
            msg: "ok",
            data:
              page === 1
                ? { items: [SAMPLE, { ...SAMPLE, id: "2222", content: "second" }], total: 3 }
                : { items: [{ ...SAMPLE, id: "3333", content: "third" }], total: 3 },
          },
        };
      },
    });

    const output = await echo.list(session(api.baseUrl), { limit: 2, page: 1, all: true });

    expect(api.calls).toHaveLength(2);
    expect(output.human).toContain("共 3 条");
    expect(output.json.split("\n")).toHaveLength(3);
  });

  test("show renders a single echo", async () => {
    api = startFakeApi({ "GET /api/echo/1111": () => ({ body: { code: 1, msg: "ok", data: SAMPLE } }) });
    const output = await echo.show(session(api.baseUrl), "1111");

    expect(output.human).toContain("hello world");
    expect(output.human).toContain("dev");
  });

  test("rm deletes after confirmation", async () => {
    api = startFakeApi({ "DELETE /api/echo/1111": () => ({ body: { code: 1, msg: "ok", data: null } }) });
    const output = await echo.remove(session(api.baseUrl), "1111", true);

    expect(api.calls[0]?.method).toBe("DELETE");
    expect(output.human).toContain("已删除");
  });
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
    const ctx = loadContext({ json: false }, env);

    const output = await auth.login(ctx, { server: api.baseUrl, token: "tok-login", profile: "memo" });

    expect(api.calls[0]?.headers.get("Authorization")).toBe("Bearer tok-login");
    expect(output.human).toContain("已登录 owner");

    const saved = loadConfig(join(dir, "config.json"));
    expect(saved.default).toBe("memo");
    expect(saved.profiles.memo?.user).toBe("owner");
    expect(saved.profiles.memo?.isAdmin).toBe(true);
    expect((await readToken("memo", ctx.credentialsPath))?.token).toBe("tok-login");
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
    const ctx = loadContext({ json: false, profile: "memo" }, env);
    const opened = await openSession(ctx);

    expect(opened.server).toBe("https://env.test");
    expect(opened.token).toBe("env-token");
    expect(opened.tokenSource).toBe("env");
  });

  test("missing profile is a usage error", async () => {
    const ctx = loadContext({ json: false }, { ECH0_CONFIG: join(mkdtempSync(join(tmpdir(), "e-")), "c.json") });
    await expect(openSession(ctx)).rejects.toBeInstanceOf(UsageError);
  });
});
