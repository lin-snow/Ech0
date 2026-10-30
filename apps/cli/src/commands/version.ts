// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { HelloResponse } from "../api/types.gen.ts";
import { colors, UsageError } from "../output/console.ts";
import { openPublicClient } from "../session.ts";
import type { CommandOutput } from "./types.ts";

export interface VersionArgs {
  remote: boolean;
  server?: string;
}

/** Local CLI version, optionally paired with the instance's `/api/hello`. */
export async function version(localVersion: string, args: VersionArgs): Promise<CommandOutput> {
  if (!args.remote) {
    return {
      human: `ech0 cli ${localVersion}`,
      json: JSON.stringify({ cli: localVersion }, null, 2),
    };
  }
  if (!args.server) {
    throw new UsageError("--remote 需要 --server 或已配置的 profile");
  }

  const client = openPublicClient(args.server);
  const result = await client.call<HelloResponse>({ path: "/hello" });
  const hello = result.data;
  const { bold } = colors();
  return {
    human: [
      `${bold("server")}    ${args.server}`,
      `${bold("version")}   ${hello.version ?? "-"}`,
      `${bold("commit")}    ${hello.commit ?? "-"}`,
      `${bold("built")}     ${hello.build_time ?? "-"}`,
      `${bold("cli")}       ${localVersion}`,
    ].join("\n"),
    json: result.text,
  };
}
