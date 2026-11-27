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
  test("returns the decoded body and sends auth + locale headers", async () => {
    api = startFakeApi({
      "GET /api/echo/abc": () => ({ body: { id: "abc", content: "hi" } }),
    });
    const client = new Client({ server: api.baseUrl, token: "tok", lang: "en-US" });
    const result = await client.call<{ id: string; content: string }>({ path: "/echo/abc", query: { page: 2 } });

    expect(result).toEqual({ id: "abc", content: "hi" });
    const call = api.calls[0];
    expect(call?.headers.get("Authorization")).toBe("Bearer tok");
    expect(call?.headers.get("X-Locale")).toBe("en-US");
  });

  test("keeps the raw body for --json output", async () => {
    api = startFakeApi({
      "GET /api/echo/abc": () => ({ body: { id: "abc" } }),
    });
    const client = new Client({ server: api.baseUrl });
    const result = await client.callWithMeta<{ id: string }>({ path: "/echo/abc" });

    expect(result.body.id).toBe("abc");
    expect(result.text).toBe('{"id":"abc"}');
  });

  test("posts JSON bodies", async () => {
    api = startFakeApi({
      "POST /api/echo": (request) => ({ body: request.body }),
    });
    const client = new Client({ server: api.baseUrl });
    await client.call({ method: "POST", path: "/echo", body: { content: "hello" } });

    expect(api.calls[0]?.body).toEqual({ content: "hello" });
    expect(api.calls[0]?.headers.get("Content-Type")).toBe("application/json");
  });

  test("maps problem responses to Ech0Error with i18n fields", async () => {
    api = startFakeApi({
      "GET /api/user": () => ({
        status: 401,
        body: {
          title: "Unauthorized",
          status: 401,
          detail: "token 无效",
          type: "https://ech0.app/problems/account-disabled",
        },
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
      expect(failure.details.type).toBe("https://ech0.app/problems/account-disabled");
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
