// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { readFileSync } from "node:fs";

import type { Echo, EchoUpsertDto, PageQueryResultListEcho, Tag } from "../api/types.gen.ts";
import { UsageError, colors } from "../output/console.ts";
import { renderEchoDetail, renderEchoList } from "../output/render.ts";
import type { Session } from "../session.ts";
import { confirm, readStdin } from "../ui/prompt.ts";
import type { CommandOutput } from "./types.ts";

export interface AddArgs {
  content?: string;
  file?: string;
  private: boolean;
  tags?: string[];
}

export async function add(session: Session, args: AddArgs): Promise<CommandOutput> {
  let content: string | undefined;
  if (args.file) {
    if (args.content) throw new UsageError("内容与 --file 只能二选一");
    try {
      content = readFileSync(args.file, "utf8");
    } catch (error) {
      throw new UsageError(`无法读取 ${args.file}：${(error as Error).message}`);
    }
  } else if (args.content === "-") {
    content = await readStdin();
  } else {
    content = args.content;
  }

  if (!content || content.trim() === "") {
    throw new UsageError("内容为空：传入文本、`-`（stdin）或 --file");
  }

  const body: EchoUpsertDto = { content, private: args.private };
  if (args.tags && args.tags.length > 0) {
    body.tags = args.tags.map((name): Tag => ({ name }));
  }

  const result = await session.client.call<unknown>({ method: "POST", path: "/echo", body });
  const { green, dim } = colors();
  return {
    human: `${green("✓")} 已发布${args.private ? "（私密）" : ""} ${dim(`${content.length} 字`)}`,
    json: result.text,
  };
}

export interface ListArgs {
  limit: number;
  page: number;
  search?: string;
  all: boolean;
}

interface EchoPage {
  items?: Echo[] | null;
  total?: number;
}

export async function list(session: Session, args: ListArgs): Promise<CommandOutput> {
  if (args.limit < 1 || args.limit > 100) throw new UsageError("--limit 范围是 1-100");
  if (args.page < 1) throw new UsageError("--page 从 1 开始");
  const { dim } = colors();

  const fetchPage = (page: number, pageSize: number) =>
    session.client.call<EchoPage>({
      method: "POST",
      path: "/echo/query",
      body: { page, pageSize, search: args.search ?? "" },
    });

  if (!args.all) {
    const result = await fetchPage(args.page, args.limit);
    const items = result.data.items ?? [];
    const total = result.data.total ?? items.length;
    return {
      human:
        items.length === 0
          ? "没有匹配的 Echo"
          : `${renderEchoList(items)}\n${dim(`第 ${args.page} 页 · 显示 ${items.length} / 共 ${total} 条`)}`,
      json: result.text,
    };
  }

  const pageSize = Math.min(args.limit, 100);
  const collected: Echo[] = [];
  let page = args.page;
  let total = 0;
  for (;;) {
    const result = await fetchPage(page, pageSize);
    const items = result.data.items ?? [];
    total = result.data.total ?? total;
    collected.push(...items);
    if (items.length === 0 || collected.length >= total) break;
    page += 1;
  }

  return {
    human:
      collected.length === 0
        ? "没有匹配的 Echo"
        : `${renderEchoList(collected)}\n${dim(`共 ${collected.length} 条`)}`,
    json: collected.map((echo) => JSON.stringify(echo)).join("\n"),
  };
}

export async function show(session: Session, id: string): Promise<CommandOutput> {
  const result = await session.client.call<Echo>({ path: `/echo/${encodeURIComponent(id)}` });
  return {
    human: renderEchoDetail(result.data),
    json: result.text,
  };
}

export async function remove(session: Session, id: string, assumeYes: boolean): Promise<CommandOutput> {
  const confirmed = await confirm(`删除 Echo ${id}？`, assumeYes);
  if (!confirmed) {
    return { human: "已取消", json: JSON.stringify({ cancelled: true }) };
  }
  const result = await session.client.call<unknown>({
    method: "DELETE",
    path: `/echo/${encodeURIComponent(id)}`,
  });
  const { green } = colors();
  return {
    human: `${green("✓")} 已删除 ${id}`,
    json: result.text,
  };
}
