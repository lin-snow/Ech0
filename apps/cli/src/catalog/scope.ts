// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * Operations the CLI exposes. The catalog describes every callable operation
 * the server offers; the CLI deliberately surfaces only the core Echo CRUD set
 * so the command tree stays small and focused. Widen this list to add a
 * capability — the command generation needs no other change.
 */
export const CLI_OPERATIONS: readonly string[] = [
  "file.upload",
  "file.external",
  "echo.create",
  "echo.page.get",
  "echo.get",
  "echo.update",
  "echo.delete",
];

const allowed = new Set(CLI_OPERATIONS);

export function inScope(id: string): boolean {
  return allowed.has(id);
}

/** Echo operations that accept composed payloads (`--from`, `--tag`, `--file-id`). */
export const ECHO_COMPOSE_OPERATIONS: ReadonlySet<string> = new Set(["echo.create", "echo.update"]);
