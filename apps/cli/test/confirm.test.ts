// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";

import { Command } from "commander";

import { parseCatalog, type CatalogOp } from "../src/catalog/types.ts";
import { confirmDestructive, destructiveTarget } from "../src/commands/confirm.ts";
import { registerCatalogCommands } from "../src/commands/dynamic.ts";
import { helpModel, nodeAt } from "../src/help/model.ts";
import { renderHelp } from "../src/help/render.ts";
import { UsageError } from "../src/output/console.ts";

const CATALOG = {
  ops: {
    "echo.delete": {
      summary: "删除 Echo",
      method: "DELETE",
      path: "/echo/{id}",
      scopes: ["echo:write"],
      effect: "destructive",
      kind: "json",
      input: {
        type: "object",
        properties: { id: { type: "string", description: "Echo ID" } },
        required: ["id"],
      },
      bind: { id: "path" },
      deprecated: false,
    },
  },
};

function deleteOp(): CatalogOp {
  const op = parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG))).ops["echo.delete"];
  if (op === undefined) throw new Error("catalog is missing echo.delete");
  return op;
}

describe("destructive confirmation", () => {
  test("fills path values into the target", () => {
    expect(destructiveTarget(deleteOp(), { id: "e/1" })).toBe("DELETE /echo/e%2F1");
  });

  test("--yes skips the prompt", async () => {
    let asked = false;
    await confirmDestructive(
      deleteOp(),
      { id: "e1" },
      {
        yes: true,
        json: false,
        yaml: false,
        interactive: true,
        ask: async () => {
          asked = true;
          return false;
        },
      },
    );
    expect(asked).toBe(false);
  });

  test("machine output demands --yes", async () => {
    await expect(
      confirmDestructive(deleteOp(), { id: "e1" }, {
        yes: false,
        json: true,
        yaml: false,
        interactive: true,
      }),
    ).rejects.toThrow("--yes");
  });

  test("non-interactive stdin demands --yes", async () => {
    await expect(
      confirmDestructive(deleteOp(), { id: "e1" }, {
        yes: false,
        json: false,
        yaml: false,
        interactive: false,
      }),
    ).rejects.toThrow("--yes");
  });

  test("interactive no cancels, yes proceeds", async () => {
    const options = { yes: false, json: false, yaml: false, interactive: true };
    const questions: string[] = [];

    await expect(
      confirmDestructive(deleteOp(), { id: "e1" }, {
        ...options,
        ask: async (question) => {
          questions.push(question);
          return false;
        },
      }),
    ).rejects.toBeInstanceOf(UsageError);

    await confirmDestructive(deleteOp(), { id: "e1" }, {
      ...options,
      ask: async (question) => {
        questions.push(question);
        return true;
      },
    });

    expect(questions[0]).toContain("DELETE /echo/e1");
  });
});

describe("destructive commands", () => {
  test("help lists --yes and the confirmation hint", () => {
    const program = new Command("ech0");
    registerCatalogCommands(program, parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG))), {
      run: async () => ({ human: "", json: "" }),
    });
    const node = nodeAt(program, ["echo", "delete"]);
    if (node === undefined) throw new Error("no echo delete command");
    const text = renderHelp(helpModel(node));

    expect(text).toContain("-y, --yes");
    expect(text).toContain("交互确认");
    expect(text).toContain("--yes");
  });

  test("without --yes the operation is not invoked and --json fails fast", async () => {
    const program = new Command("ech0")
      .exitOverride()
      .option("--json", "输出 JSON", false)
      .option("--yaml", "输出 YAML", false);
    let invoked = 0;
    registerCatalogCommands(program, parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG))), {
      run: async () => {
        invoked += 1;
        return { human: "", json: "" };
      },
    });

    await expect(program.parseAsync(["echo", "delete", "--id", "e1", "--json"], { from: "user" })).rejects.toThrow("--yes");
    expect(invoked).toBe(0);
  });

  test("--yes invokes the operation", async () => {
    const program = new Command("ech0")
      .exitOverride()
      .option("--json", "输出 JSON", false)
      .option("--yaml", "输出 YAML", false);
    let invoked = 0;
    registerCatalogCommands(program, parseCatalog(new TextEncoder().encode(JSON.stringify(CATALOG))), {
      run: async () => {
        invoked += 1;
        return { human: "", json: "" };
      },
    });

    await program.parseAsync(["echo", "delete", "--id", "e1", "--yes", "--json"], { from: "user" });
    expect(invoked).toBe(1);
  });
});
