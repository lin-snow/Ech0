// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import pkg from "../../package.json" with { type: "json" };
import { loadCatalog, type LoadedCatalog } from "../catalog/load.ts";
import { activeProfileName } from "../config/store.ts";
import {
  colors,
  print,
  printErr,
  setColorEnabled,
  shouldUseColor,
  UsageError,
} from "../output/console.ts";
import { loadContext, openSession, readGlobalFlags, type GlobalFlags } from "../session.ts";
import { readStdin } from "../ui/prompt.ts";
import * as auth from "./auth.ts";
import * as catalogCommands from "./catalog.ts";
import * as config from "./config.ts";
import { registerCatalogCommands } from "./dynamic.ts";
import { invokeOperation } from "./operation.ts";
import type { CommandOutput } from "./types.ts";
import * as version from "./version.ts";

function emit(result: CommandOutput, json: boolean): void {
  print(json ? result.json : result.human);
}

function authCommand(): Command {
  const group = new Command("auth").description("登录、查看与清除实例凭据");

  group
    .command("login")
    .description("用 access token 登录实例（首次会写入 profile）")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      const ctx = loadContext(flags);
      const rawToken = flags.token ?? ctx.env.ECH0_TOKEN;
      if (!rawToken) {
        throw new UsageError("缺少 token：--token <token>（或在面板铸造后通过 ECH0_TOKEN 传入）");
      }
      const token = rawToken === "-" ? (await readStdin()).trim() : rawToken;
      const result = await auth.login(ctx, { server: flags.server, token, profile: flags.profile });
      emit(result, flags.json);
    });

  group
    .command("status")
    .description("显示当前 profile、实例与 token 来源")
    .option("--verify", "请求 /api/user 校验 token 是否有效")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      const ctx = loadContext(flags);
      const options = command.opts();
      emit(await auth.status(ctx, { verify: options.verify === true }), flags.json);
    });

  group
    .command("logout")
    .description("删除当前 profile 的 token")
    .option("--purge", "同时删除该 profile 配置")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      const ctx = loadContext(flags);
      const options = command.opts();
      emit(await auth.logout(ctx, { purge: options.purge === true }), flags.json);
    });

  return group;
}

function configCommand(): Command {
  const group = new Command("config").description("管理 profile 与配置文件位置");

  group
    .command("list")
    .description("列出所有 profile")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(config.list(loadContext(flags)), flags.json);
    });

  group
    .command("use <name>")
    .description("切换默认 profile")
    .action((name: string, _options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(config.use(loadContext(flags), name), flags.json);
    });

  group
    .command("unset <name>")
    .description("删除 profile 及其 token")
    .action(async (name: string, _options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(await config.unset(loadContext(flags), name), flags.json);
    });

  group
    .command("path")
    .description("打印配置与凭据文件路径")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(config.path(loadContext(flags)), flags.json);
    });

  return group;
}

function catalogCommand(refresh: (flags: GlobalFlags) => Promise<LoadedCatalog>): Command {
  const group = new Command("catalog").description("服务端操作清单：命令树由它生成");

  group
    .command("refresh")
    .description("重新拉取 catalog 并写入本地缓存")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(catalogCommands.summarize(await refresh(flags)), flags.json);
    });

  group
    .command("path")
    .description("打印当前实例的 catalog 缓存路径")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(catalogCommands.cachePath(loadContext(flags)), flags.json);
    });

  return group;
}

function versionCommand(): Command {
  return new Command("version")
    .description("显示 CLI 版本")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(version.version(pkg.version), flags.json);
    });
}

/** Cache-first: the command tree must exist before argv is parsed. */
async function bootstrapCatalog(): Promise<LoadedCatalog | undefined> {
  const ctx = loadContext({ json: false });
  const profile = activeProfileName(ctx.config, undefined, ctx.env);
  const server = ctx.env.ECH0_SERVER ?? (profile ? ctx.config.profiles[profile]?.server : undefined);
  if (!server) return undefined;
  try {
    return await loadCatalog({
      server,
      token: ctx.env.ECH0_TOKEN,
      lang: ctx.env.ECH0_LANG,
      configPath: ctx.configPath,
    });
  } catch (error) {
    printErr(
      `${colors().yellow("提示")} catalog 不可用：${error instanceof Error ? error.message : String(error)}`,
    );
    return undefined;
  }
}

/** Builds the `ech0` command tree: local commands plus one command per server operation. */
export async function createRootCommand(): Promise<Command> {
  const program = new Command("ech0")
    .description(pkg.description)
    .version(pkg.version, "-v, --version")
    .option("--profile <name>", "使用指定 profile")
    .option("--server <url>", "覆盖实例地址")
    .option("--token <token>", "覆盖 access token")
    .option("--lang <locale>", "请求语言（X-Locale），如 zh-CN / en-US")
    .option("--json", "输出原始 JSON 响应", false)
    .option("--no-color", "关闭 ANSI 颜色")
    .showHelpAfterError()
    .showSuggestionAfterError()
    .configureHelp({ sortSubcommands: true });

  program.hook("preAction", (thisCommand) => {
    const opts = thisCommand.optsWithGlobals();
    setColorEnabled(opts.color !== false && shouldUseColor());
  });

  let current: LoadedCatalog | undefined;
  const refresh = async (flags: GlobalFlags): Promise<LoadedCatalog> => {
    const loaded = await catalogCommands.refreshCatalog(flags);
    current = loaded;
    return loaded;
  };

  program.addCommand(authCommand());
  program.addCommand(configCommand());
  program.addCommand(catalogCommand(refresh));
  program.addCommand(versionCommand());

  current = await bootstrapCatalog();
  if (current !== undefined) {
    registerCatalogCommands(program, current.doc, {
      run: async (command, op, values) => {
        const flags = readGlobalFlags(command);
        const session = await openSession(loadContext(flags));
        const loaded = current ?? (await refresh(flags));
        return invokeOperation({
          session,
          op,
          values,
          catalog: loaded,
          refresh: () => refresh(flags),
          json: flags.json,
        });
      },
    });
  }

  program.action(() => program.outputHelp());

  return program;
}
