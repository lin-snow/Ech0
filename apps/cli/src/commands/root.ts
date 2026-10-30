// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import pkg from "../../package.json" with { type: "json" };
import { activeProfileName } from "../config/store.ts";
import { setColorEnabled, shouldUseColor, UsageError, print } from "../output/console.ts";
import { loadContext, openSession, type GlobalFlags } from "../session.ts";
import { readStdin } from "../ui/prompt.ts";
import * as auth from "./auth.ts";
import * as config from "./config.ts";
import * as echo from "./echo.ts";
import * as me from "./me.ts";
import type { CommandOutput } from "./types.ts";
import * as version from "./version.ts";

function readFlags(command: Command): GlobalFlags {
  const opts = command.optsWithGlobals();
  return {
    profile: typeof opts.profile === "string" ? opts.profile : undefined,
    server: typeof opts.server === "string" ? opts.server : undefined,
    token: typeof opts.token === "string" ? opts.token : undefined,
    lang: typeof opts.lang === "string" ? opts.lang : undefined,
    json: opts.json === true,
  };
}

function emit(result: CommandOutput, json: boolean): void {
  print(json ? result.json : result.human);
}

function intArg(value: string): number {
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed)) throw new Error(`不是整数：${value}`);
  return parsed;
}

function authCommand(): Command {
  const group = new Command("auth").description("登录、查看与清除实例凭据");

  group
    .command("login")
    .description("用 access token 登录实例（首次会写入 profile）")
    .action(async (_options, command: Command) => {
      const flags = readFlags(command);
      const ctx = loadContext(flags);
      const rawToken = flags.token ?? ctx.env.ECH0_TOKEN;
      if (!rawToken) {
        throw new UsageError("缺少 token：--token <token>（或在面板铸造后通过 ECH0_TOKEN 传入）");
      }
      const token = rawToken === "-" ? (await readStdin()).trim() : rawToken;
      const result = await auth.login(ctx, {
        server: flags.server,
        token,
        profile: flags.profile,
      });
      emit(result, flags.json);
    });

  group
    .command("status")
    .description("显示当前 profile、实例与 token 来源")
    .option("--verify", "请求 /api/user 校验 token 是否有效")
    .action(async (_options, command: Command) => {
      const flags = readFlags(command);
      const ctx = loadContext(flags);
      const options = command.opts();
      emit(await auth.status(ctx, { verify: options.verify === true }), flags.json);
    });

  group
    .command("logout")
    .description("删除当前 profile 的 token")
    .option("--purge", "同时删除该 profile 配置")
    .action(async (_options, command: Command) => {
      const flags = readFlags(command);
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
      const flags = readFlags(command);
      emit(config.list(loadContext(flags)), flags.json);
    });

  group
    .command("use <name>")
    .description("切换默认 profile")
    .action((name: string, _options, command: Command) => {
      const flags = readFlags(command);
      emit(config.use(loadContext(flags), name), flags.json);
    });

  group
    .command("unset <name>")
    .description("删除 profile 及其 token")
    .action(async (name: string, _options, command: Command) => {
      const flags = readFlags(command);
      emit(await config.unset(loadContext(flags), name), flags.json);
    });

  group
    .command("path")
    .description("打印配置与凭据文件路径")
    .action((_options, command: Command) => {
      const flags = readFlags(command);
      emit(config.path(loadContext(flags)), flags.json);
    });

  return group;
}

function echoCommand(): Command {
  const group = new Command("echo").description("发布与查看 Echo");

  group
    .command("add [content]")
    .description("发布一条 Echo（`-` 从 stdin 读取）")
    .option("--file <path>", "从文件读取正文")
    .option("--private", "设为私密", false)
    .option("--tags <list>", "逗号分隔的标签名")
    .action(async (content: string | undefined, _options, command: Command) => {
      const flags = readFlags(command);
      const session = await openSession(loadContext(flags));
      const options = command.opts();
      const tags =
        typeof options.tags === "string"
          ? options.tags.split(",").map((tag: string) => tag.trim()).filter((tag: string) => tag !== "")
          : undefined;
      emit(
        await echo.add(session, {
          content,
          file: typeof options.file === "string" ? options.file : undefined,
          private: options.private === true,
          tags,
        }),
        flags.json,
      );
    });

  group
    .command("ls")
    .description("列出 Echo（按时间倒序）")
    .option("--limit <n>", "每页条数，1-100", intArg, 20)
    .option("--page <n>", "页码，从 1 开始", intArg, 1)
    .option("--search <text>", "按正文搜索")
    .option("--all", "自动翻页取全部", false)
    .action(async (_options, command: Command) => {
      const flags = readFlags(command);
      const session = await openSession(loadContext(flags));
      const options = command.opts();
      emit(
        await echo.list(session, {
          limit: options.limit,
          page: options.page,
          search: typeof options.search === "string" ? options.search : undefined,
          all: options.all === true,
        }),
        flags.json,
      );
    });

  group
    .command("show <id>")
    .description("查看单条 Echo")
    .action(async (id: string, _options, command: Command) => {
      const flags = readFlags(command);
      const session = await openSession(loadContext(flags));
      emit(await echo.show(session, id), flags.json);
    });

  group
    .command("rm <id>")
    .description("删除 Echo")
    .option("--yes", "跳过确认", false)
    .action(async (id: string, _options, command: Command) => {
      const flags = readFlags(command);
      const session = await openSession(loadContext(flags));
      const options = command.opts();
      emit(await echo.remove(session, id, options.yes === true), flags.json);
    });

  return group;
}

/** Builds the `ech0` command tree. */
export function createRootCommand(): Command {
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

  program.addCommand(authCommand());
  program.addCommand(configCommand());
  program.addCommand(echoCommand());

  program
    .command("me")
    .description("显示当前 token 对应的账号")
    .action(async (_options, command: Command) => {
      const flags = readFlags(command);
      emit(await me.me(await openSession(loadContext(flags))), flags.json);
    });

  program
    .command("version")
    .description("显示 CLI 版本；--remote 显示服务端版本")
    .option("--remote", "同时请求服务端 /api/hello", false)
    .action(async (_options, command: Command) => {
      const flags = readFlags(command);
      const ctx = loadContext(flags);
      const profile = activeProfileName(ctx.config, flags.profile, ctx.env);
      const server = flags.server ?? ctx.env.ECH0_SERVER ?? (profile ? ctx.config.profiles[profile]?.server : undefined);
      const options = command.opts();
      emit(await version.version(pkg.version, { remote: options.remote === true, server }), flags.json);
    });

  program.action(() => program.outputHelp());

  return program;
}
