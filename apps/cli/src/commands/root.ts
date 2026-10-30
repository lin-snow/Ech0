// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Command } from "commander";

import pkg from "../../package.json" with { type: "json" };
import { loadCatalog, type LoadedCatalog } from "../catalog/load.ts";
import { readToken } from "../config/credentials.ts";
import { activeProfileName } from "../config/store.ts";
import {
  commandIds,
  commandOp,
  entriesUnder,
  helpModel,
  nodeAt,
  setCommandExamples,
  setCommandHints,
  setHelpContext,
} from "../help/model.ts";
import { formatCommand, renderHelp } from "../help/render.ts";
import { printResult } from "../output/result.ts";
import { suggestions } from "../help/search.ts";
import {
  colors,
  printErr,
  setColorEnabled,
  shouldUseColor,
  UsageError,
} from "../output/console.ts";
import { loadContext, openSession, readGlobalFlags, type GlobalFlags } from "../session.ts";
import { readPassword, readStdin } from "../ui/prompt.ts";
import * as auth from "./auth.ts";
import { installCommand } from "./install.ts";
import * as catalogCommands from "./catalog.ts";
import * as config from "./config.ts";
import { registerCatalogCommands } from "./dynamic.ts";
import { createHelpCommand } from "./help.ts";
import { invokeOperation } from "./operation.ts";
import type { CommandOutput } from "./types.ts";
import { CLI_VERSION } from "../version-info.ts";
import * as version from "./version.ts";

function readScopes(raw: unknown): string[] | undefined {
  if (typeof raw !== "string") return undefined;
  const scopes = raw
    .split(",")
    .map((scope) => scope.trim())
    .filter((scope) => scope !== "");
  return scopes.length > 0 ? scopes : undefined;
}

function profileServer(ctx: ReturnType<typeof loadContext>, profile: string | undefined): string | undefined {
  const name = profile ?? activeProfileName(ctx.config, undefined, ctx.env);
  return name ? ctx.config.profiles[name]?.server : undefined;
}

function emit(result: CommandOutput, flags: GlobalFlags): void {
  printResult(result, flags);
}

function authCommand(): Command {
  const group = new Command("auth").description("登录、查看与清除实例凭据");

  const login = group
    .command("login")
    .description("登录实例：access token / 密码 / 设备授权")
    .option("--device", "使用浏览器设备授权登录", false)
    .option("-u, --username <name>", "用户名（密码登录）")
    .option("--password-stdin", "从 stdin 读取密码", false)
    .option("--scopes <list>", "逗号分隔的 scope（密码/设备登录铸造令牌时使用）")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      const ctx = loadContext(flags);
      const options = command.opts();
      const scopes = readScopes(options.scopes);

      const inlineToken = flags.token ?? ctx.env.ECH0_TOKEN;
      if (inlineToken !== undefined && inlineToken !== "") {
        const token = inlineToken === "-" ? (await readStdin()).trim() : inlineToken;
        emit(await auth.login(ctx, { server: flags.server, token, profile: flags.profile }), flags);
        return;
      }

      if (options.device === true) {
        const server = flags.server ?? ctx.env.ECH0_SERVER ?? profileServer(ctx, flags.profile);
        const result = await auth.loginWithDevice(ctx, {
          server: flags.server,
          profile: flags.profile,
          scopes,
          onCode: (code) => {
            const where = server ? `${server.replace(/\/+$/, "")}${code.verification_uri_complete}` : code.verification_uri_complete;
            printErr(`${colors().bold("在浏览器中确认")} ${where}`);
            printErr(`${colors().bold("设备码")} ${code.user_code}（${code.expires_in}s 内有效）`);
          },
        });
        emit(result, flags);
        return;
      }

      if (typeof options.username === "string") {
        const password =
          options.passwordStdin === true
            ? ((await readStdin()).split("\n")[0] ?? "")
            : await readPassword("密码: ");
        if (password === "") throw new UsageError("密码为空");
        emit(
          await auth.loginWithPassword(ctx, {
            server: flags.server,
            profile: flags.profile,
            username: options.username,
            password,
            scopes,
          }),
          flags,
        );
        return;
      }

      throw new UsageError("登录方式：--token <token>（access token）、-u/--username（密码）或 --device（浏览器授权）");
    });
  setCommandHints(login, [
    "--device：CLI 打印设备码与确认链接；浏览器批准后保存 1 个月设备令牌",
    "--token -：从 stdin 读取 access token（CI 也可用 ECH0_TOKEN）",
    "-u <name>：密码登录；管理员会额外铸造 audience=cli 的 1 个月令牌，普通用户保存会话令牌 + refresh cookie",
    `--scopes：只在密码/设备登录铸造令牌时生效，默认 ${auth.DEFAULT_CLI_SCOPES.join(",")}`,
    "凭据按 profile 保存（ech0 config list 查看）；--no-keychain 或 ECH0_NO_KEYCHAIN=1 可跳过系统钥匙串",
  ]);

  group
    .command("status")
    .description("显示当前 profile、实例与 token 来源")
    .option("--verify", "请求 /api/user 校验 token 是否有效")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      const ctx = loadContext(flags);
      const options = command.opts();
      emit(await auth.status(ctx, { verify: options.verify === true }), flags);
    });

  group
    .command("logout")
    .description("删除当前 profile 的 token")
    .option("--purge", "同时删除该 profile 配置")
    .action(async (_options, command: Command) => {
      const flags = readGlobalFlags(command);
      const ctx = loadContext(flags);
      const options = command.opts();
      emit(await auth.logout(ctx, { purge: options.purge === true }), flags);
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
      emit(config.list(loadContext(flags)), flags);
    });

  group
    .command("use <name>")
    .description("切换默认 profile")
    .action((name: string, _options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(config.use(loadContext(flags), name), flags);
    });

  group
    .command("unset <name>")
    .description("删除 profile 及其 token")
    .action(async (name: string, _options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(await config.unset(loadContext(flags), name), flags);
    });

  group
    .command("path")
    .description("打印配置与凭据文件路径")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(config.path(loadContext(flags)), flags);
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
      emit(catalogCommands.summarize(await refresh(flags)), flags);
    });

  group
    .command("path")
    .description("打印当前实例的 catalog 缓存路径")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(catalogCommands.cachePath(loadContext(flags)), flags);
    });

  return group;
}

function versionCommand(): Command {
  return new Command("version")
    .description("显示 CLI 版本")
    .action((_options, command: Command) => {
      const flags = readGlobalFlags(command);
      emit(version.version(CLI_VERSION), flags);
    });
}

interface CatalogBootstrap {
  catalog?: LoadedCatalog;
  server?: string;
  failure?: string;
  authenticated?: boolean;
}

/** Cache-first: the command tree must exist before argv is parsed. */
async function bootstrapCatalog(): Promise<CatalogBootstrap> {
  const ctx = loadContext({ json: false, yaml: false });
  const profile = activeProfileName(ctx.config, undefined, ctx.env);
  const server = ctx.env.ECH0_SERVER ?? (profile ? ctx.config.profiles[profile]?.server : undefined);
  const stored = profile === undefined ? undefined : await readToken(profile, ctx.credentialsPath).catch(() => undefined);
  const authenticated = Boolean(ctx.env.ECH0_TOKEN) || stored !== undefined;
  if (!server) return { failure: "未配置实例：没有 profile，也没有 ECH0_SERVER", authenticated };
  try {
    return {
      server,
      authenticated,
      catalog: await loadCatalog({
        server,
        token: ctx.env.ECH0_TOKEN,
        lang: ctx.env.ECH0_LANG,
        configPath: ctx.configPath,
      }),
    };
  } catch (error) {
    const failure = error instanceof Error ? error.message : String(error);
    printErr(`${colors().yellow("提示")} catalog 不可用：${failure}`);
    return { server, failure, authenticated };
  }
}

/** Every command in the tree, root included. */
function allCommands(root: Command): Command[] {
  const found: Command[] = [root];
  for (const child of root.commands) found.push(...allCommands(child));
  return found;
}

/** Examples local commands cannot derive from a schema. */
function localExamples(program: Command): void {
  const examples: Array<[string[], string[]]> = [
    [
      ["auth", "login"],
      [
        "ech0 auth login --server https://example.com --device",
        "ech0 auth login --server https://example.com -u admin",
        "ech0 auth login --server https://example.com --token -",
      ],
    ],
    [["auth", "status"], ["ech0 auth status --verify"]],
    [["config", "use"], ["ech0 config use memo"]],
    [["config", "list"], ["ech0 config list"]],
    [["catalog", "refresh"], ["ech0 catalog refresh"]],
    [["install", "skills"], ["ech0 install skills ~/.claude/skills"]],
    [["help"], ["ech0 help", "ech0 help echo", "ech0 help page"]],
  ];
  for (const [path, lines] of examples) {
    const command = nodeAt(program, path);
    if (command !== undefined) setCommandExamples(command, lines);
  }
}

/** Builds the `ech0` command tree: local commands plus one command per server operation. */
export async function createRootCommand(): Promise<Command> {
  const program = new Command("ech0")
    .description(pkg.description)
    .version(CLI_VERSION, "-v, --version", "显示 CLI 版本")
    .option("--profile <name>", "使用指定 profile")
    .option("--server <url>", "覆盖实例地址")
    .option("--token <token>", "覆盖 access token")
    .option("--lang <locale>", "请求语言（X-Locale），如 zh-CN / en-US")
    .option("--json", "输出原始 JSON 响应", false)
    .option("--yaml", "输出 YAML（与 --json 同一文档）", false)
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
  program.addCommand(installCommand());
  program.addCommand(createHelpCommand());

  const boot = await bootstrapCatalog();
  current = boot.catalog;
  if (current !== undefined) {
    registerCatalogCommands(program, current.doc, {
      run: async (command, op, values, extra) => {
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
          yaml: flags.yaml,
          output: extra.output,
          compose: extra.compose,
          payload: extra.payload,
          tags: extra.tags,
          fileIds: extra.fileIds,
          uploadFileName: extra.uploadFileName,
        });
      },
    });
  }

  // One help renderer for the whole tree: local commands, namespaces and
  // catalog operations read the same sections in the same order.
  for (const command of allCommands(program)) {
    command.configureHelp({ formatHelp: formatCommand });
    command.addHelpCommand(false);
  }

  setHelpContext(program, {
    server: boot.server,
    loaded: boot.catalog !== undefined,
    failure: boot.failure,
    ops: boot.catalog === undefined ? undefined : Object.keys(boot.catalog.doc.ops).length,
    authenticated: boot.authenticated,
  });
  localExamples(program);
  const serverGroup = program.commands.find((child) =>
    entriesUnder(child).some((entry) => commandOp(entry) !== undefined),
  );
  setCommandExamples(program, [
    "ech0 auth login --server https://example.com --device",
    "ech0 help",
    ...(serverGroup === undefined ? [] : [`ech0 help ${serverGroup.name()}`]),
    "ech0 catalog refresh",
    "ech0 install skills ~/.claude/skills",
  ]);

  program.argument("[words...]", "命令或关键词（与 `ech0 help` 等价）");
  program.action((words: string[]) => {
    if (words.length === 0) {
      process.stdout.write(renderHelp(helpModel(program)));
      return;
    }
    const hints = suggestions(commandIds(program), words);
    throw new UsageError(
      [
        `未知命令：${words.join(" ")}`,
        ...(hints.length === 0 ? [] : [`  是否想运行：${hints.map((id) => `ech0 ${id}`).join("、")}`]),
        "  运行 `ech0 help` 查看全部命令，或 `ech0 help <关键词>` 搜索",
      ].join("\n"),
    );
  });

  return program;
}
