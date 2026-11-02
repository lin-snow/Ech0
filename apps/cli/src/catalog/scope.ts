// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * Operations the CLI exposes. The catalog describes every callable operation
 * the server offers; the CLI deliberately surfaces only the core Echo CRUD and
 * read views, file, tag, comment, Connect and webhook commands. The comment and
 * Connect sets mirror the MCP tool surface: post/read comments and
 * list/info/add/delete peers. There is no comment moderation, Connect health
 * probe or self card, so the default login token never needs
 * `comment:moderate` or `connect:read`. Tag management is list/delete only —
 * Echo create/update already auto-creates tags by name. Webhook management
 * mirrors the MCP tools and needs `admin:settings`, the only admin scope in the
 * default set. Widen this list to add a capability — the command generation
 * needs no other change.
 */
export const CLI_OPERATIONS: readonly string[] = [
  "file.upload",
  "file.external",
  "file.list",
  "file.get",
  "file.tree",
  "file.stream.by_id",
  "echo.create",
  "echo.list",
  "echo.get",
  "echo.update",
  "echo.delete",
  "echo.hot",
  "echo.today",
  "echo.random",
  "echo.onthisday",
  "tag.list",
  "tag.delete",
  "comment.create.integration",
  "comment.list.by.echo",
  "connect.list",
  "connect.info",
  "connect.add",
  "connect.delete",
  "webhook.list",
  "webhook.create",
  "webhook.update",
  "webhook.delete",
  "webhook.test",
];

const allowed = new Set(CLI_OPERATIONS);

export function inScope(id: string): boolean {
  return allowed.has(id);
}

/**
 * Command-path overrides for catalog ids that read poorly as a command name.
 * `file.stream.by_id` is a content download, so it becomes `ech0 file download`
 * while the catalog id stays the stable key used by help, scopes and requests.
 */
export const COMMAND_NAMES: Readonly<Record<string, string>> = {
  "file.stream.by_id": "file.download",
  "comment.create.integration": "comment.create",
  "comment.list.by.echo": "comment.list",
};

/** Echo operations that accept composed payloads (`--from`, `--tag`, `--file-id`). */
export const ECHO_COMPOSE_OPERATIONS: ReadonlySet<string> = new Set(["echo.create", "echo.update"]);
