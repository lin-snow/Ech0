// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * Operations the CLI exposes. The catalog describes every callable operation
 * the server offers; the CLI deliberately surfaces the core Echo CRUD set, the
 * file surface (upload, external registration, browsing and download), the
 * public comment surface plus comment moderation, and the Connect surface so
 * the command tree stays small and focused. Widen this list to add a
 * capability — the command generation needs no other change.
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
  "comment.create",
  "comment.list.by.echo",
  "comment.list.public",
  "comment.panel.delete",
  "connect.self",
  "connect.list",
  "connect.add",
  "connect.delete",
  "connect.health",
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
  "comment.list.by.echo": "comment.list",
  "comment.list.public": "comment.public",
  "comment.panel.delete": "comment.delete",
  "connect.self": "connect.status",
  "connect.delete": "connect.remove",
};

/** Echo operations that accept composed payloads (`--from`, `--tag`, `--file-id`). */
export const ECHO_COMPOSE_OPERATIONS: ReadonlySet<string> = new Set(["echo.create", "echo.update"]);
