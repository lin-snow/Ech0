// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import { parseCatalog, type CatalogOp } from "../src/catalog/types.ts";
import { registerCatalogCommands, type DynamicOptions } from "../src/commands/dynamic.ts";
import type { FieldValues } from "../src/commands/operation.ts";
import type { CommandOutput } from "../src/commands/types.ts";
import { helpModel, nodeAt } from "../src/help/model.ts";
import { renderHelp } from "../src/help/render.ts";

export interface Invocation {
  request: string;
  values: FieldValues;
}

/** Builds the real command tree from a fixture catalog and records invocations. */
export function buildProgram(
  catalogJson: { ops: Record<string, unknown> },
  run?: DynamicOptions["run"],
): {
  program: Command;
  invocations: Invocation[];
} {
  const invocations: Invocation[] = [];
  const program = new Command("ech0")
    .option("--json", "输出 JSON", false)
    .option("--yaml", "输出 YAML", false);
  const catalog = parseCatalog(new TextEncoder().encode(JSON.stringify(catalogJson)));
  registerCatalogCommands(program, catalog, {
    run:
      run ??
      (async (_command: Command, op: CatalogOp, values: FieldValues): Promise<CommandOutput> => {
        invocations.push({ request: `${op.method} ${op.path}`, values });
        return { human: "", json: "" };
      }),
  });
  const override = (command: Command): void => {
    command.exitOverride();
    for (const child of command.commands) override(child);
  };
  override(program);
  return { program, invocations };
}

export function helpText(program: Command, path: string[]): string {
  const node = nodeAt(program, path);
  if (node === undefined) throw new Error(`no command ${path.join(" ")}`);
  return renderHelp(helpModel(node));
}
