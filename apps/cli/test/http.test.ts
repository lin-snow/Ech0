// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, describe, expect, test } from "bun:test";

import { Client, normalizeBaseUrl } from "../src/api/http.ts";
import { Ech0Error } from "../src/output/console.ts";
import { startFakeApi, type FakeApi } from "./helpers.ts";

let api: FakeApi | undefined;

afterEach(() => {
  api?.stop();
  api = undefined;
});

describe("normalizeBaseUrl", () => {
  test("appends /api once", () => {
    expect(normalizeBaseUrl("https://a.test")).toBe("https://a.test/api");
    expect(normalizeBaseUrl("https://a.test/")).toBe("https://a.test/api");
    expect(normalizeBaseUrl("https://a.test/api")).toBe("https://a.test/api");
  });

  test("rejects empty server", () => {
    expect(() => normalizeBaseUrl("  ")).toThrow(Ech0Error);
  });
});

describe("Client", () => {
  test("unwraps the envelope and sends auth + locale headers", async () => {
    api = startFakeApi({
      "GET /api/echo/abc": () => ({ body: { code: 1, msg: "ok", data: { id: "abc", content: "hi" } } }),
    });
    const client = new Client({ server: api.baseUrl, token: "tok", lang: "en-US" });
    const result = await client.call<{ id: string }>({ path: "/echo/abc", query: { page: 2 } });

    expect(result.data.id).toBe("abc");
    expect(result.text).toContain('"code":1');
    const call = api.calls[0];
    expect(call?.headers.get("Authorization")).toBe("Bearer tok");
    expect(call?.headers.get("X-Locale")).toBe("en-US");
  });

  test("posts JSON bodies", async () => {
    api = startFakeApi({
      "POST /api/echo": (request) => ({ body: { code: 1, msg: "ok", data: request.body } }),
    });
    const client = new Client({ server: api.baseUrl });
    await client.call({ method: "POST", path: "/echo", body: { content: "hello" } });

    expect(api.calls[0]?.body).toEqual({ content: "hello" });
    expect(api.calls[0]?.headers.get("Content-Type")).toBe("application/json");
  });

  test("maps failure envelopes to Ech0Error with i18n fields", async () => {
    api = startFakeApi({
      "GET /api/user": () => ({
        status: 401,
        body: { code: 0, msg: "token 无效", error_code: "AUTH_INVALID", message_key: "auth.token_invalid" },
      }),
    });
    const client = new Client({ server: api.baseUrl, token: "bad" });

    try {
      await client.call({ path: "/user" });
      throw new Error("expected call to reject");
    } catch (error) {
      expect(error).toBeInstanceOf(Ech0Error);
      const failure = error as Ech0Error;
      expect(failure.message).toBe("token 无效");
      expect(failure.details.status).toBe(401);
      expect(failure.details.errorCode).toBe("AUTH_INVALID");
      expect(failure.details.messageKey).toBe("auth.token_invalid");
    }
  });

  test("keeps raw text for non-JSON error bodies", async () => {
    api = startFakeApi({
      "GET /api/echo/x": () => ({ status: 502, body: "bad gateway" }),
    });
    const client = new Client({ server: api.baseUrl });
    await expect(client.call({ path: "/echo/x" })).rejects.toThrow("bad gateway");
  });
});
