// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, test } from "bun:test";
import { mkdtempSync, statSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { deleteToken, readToken, writeToken } from "../src/config/credentials.ts";
import { configFilePath } from "../src/config/paths.ts";
import {
  activeProfileName,
  loadConfig,
  saveConfig,
  withProfile,
  withoutProfile,
} from "../src/config/store.ts";

process.env.ECH0_NO_KEYCHAIN = "1";

function tempDir(): string {
  return mkdtempSync(join(tmpdir(), "ech0-cli-"));
}

describe("config paths", () => {
  test("honours ECH0_CONFIG, XDG_CONFIG_HOME and HOME", () => {
    expect(configFilePath({ ECH0_CONFIG: "/tmp/x.json" })).toBe("/tmp/x.json");
    expect(configFilePath({ XDG_CONFIG_HOME: "/xdg" })).toBe("/xdg/ech0/config.json");
    expect(configFilePath({ HOME: "/home/u" })).toBe("/home/u/.config/ech0/config.json");
  });
});

describe("config store", () => {
  test("round-trips profiles and keeps the default", () => {
    const path = join(tempDir(), "config.json");
    const initial = withProfile({ profiles: {} }, "memo", { server: "https://a.test", user: "owner" });
    saveConfig(initial, path);

    const loaded = loadConfig(path);
    expect(loaded.default).toBe("memo");
    expect(loaded.profiles.memo?.server).toBe("https://a.test");
    expect(activeProfileName(loaded, undefined, {})).toBe("memo");
    expect(activeProfileName(loaded, "other", {})).toBe("other");
    expect(activeProfileName(loaded, undefined, { ECH0_PROFILE: "env" })).toBe("env");
  });

  test("missing file yields an empty config", () => {
    expect(loadConfig(join(tempDir(), "absent.json"))).toEqual({ default: undefined, profiles: {} });
  });

  test("removing the default profile promotes another one", () => {
    const config = withProfile(withProfile({ profiles: {} }, "a", { server: "https://a.test" }), "b", {
      server: "https://b.test",
    });
    const next = withoutProfile(config, "a");
    expect(next.default).toBe("b");
    expect(Object.keys(next.profiles)).toEqual(["b"]);
  });

  test("rejects malformed JSON", () => {
    const dir = tempDir();
    const path = join(dir, "config.json");
    saveConfig({ profiles: {} }, path);
    Bun.write(path, "{not json");
    expect(() => loadConfig(path)).toThrow();
  });
});

describe("credential storage", () => {
  test("falls back to a 0600 file when the keychain is disabled", async () => {
    const file = join(tempDir(), "credentials.json");
    const source = await writeToken("memo", "tok-1", file);
    expect(source).toBe("file");
    expect(statSync(file).mode & 0o777).toBe(0o600);

    const stored = await readToken("memo", file);
    expect(stored).toEqual({ value: "tok-1", source: "file" });

    await deleteToken("memo", file);
    expect(await readToken("memo", file)).toBeUndefined();
  });
});
