// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export interface RecordedRequest {
  method: string;
  path: string;
  headers: Headers;
  body?: unknown;
}

export interface RouteReply {
  status?: number;
  body: unknown;
}

export interface FakeApi {
  baseUrl: string;
  calls: RecordedRequest[];
  stop(): void;
}

/** In-memory HTTP server keyed by `"METHOD /path"`, recording every call. */
export function startFakeApi(routes: Record<string, (request: RecordedRequest) => RouteReply>): FakeApi {
  const calls: RecordedRequest[] = [];
  const server = Bun.serve({
    port: 0,
    async fetch(request) {
      const url = new URL(request.url);
      const text = await request.text();
      let body: unknown;
      if (text !== "") {
        try {
          body = JSON.parse(text);
        } catch {
          body = text;
        }
      }
      const recorded: RecordedRequest = {
        method: request.method,
        path: url.pathname,
        headers: request.headers,
        body,
      };
      calls.push(recorded);

      const key = `${request.method} ${url.pathname}`;
      const handler = routes[key];
      const reply = handler
        ? handler(recorded)
        : { status: 404, body: { code: 0, msg: `no route: ${key}` } };
      return new Response(JSON.stringify(reply.body), {
        status: reply.status ?? 200,
        headers: { "Content-Type": "application/json" },
      });
    },
  });

  return {
    baseUrl: `http://127.0.0.1:${server.port}`,
    calls,
    stop: () => server.stop(true),
  };
}
