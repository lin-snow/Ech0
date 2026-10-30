// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { User } from "../api/types.gen.ts";
import { renderUser } from "../output/render.ts";
import type { Session } from "../session.ts";
import type { CommandOutput } from "./types.ts";

/** Prints the account behind the current token. */
export async function me(session: Session): Promise<CommandOutput> {
  const result = await session.client.call<User>({ path: "/user" });
  return {
    human: renderUser(result.data),
    json: result.text,
  };
}
