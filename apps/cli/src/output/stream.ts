// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { SseFrame } from "../api/sse.ts";
import { colors, print, printErr } from "./console.ts";

export interface StreamRenderer {
  onFrame(frame: SseFrame): void;
  finish(): void;
  /** True once an `error` event arrived. */
  readonly failed: boolean;
}

interface AskQuestion {
  id?: string;
  text?: string;
  header?: string;
  detail?: string;
  multi?: boolean;
  options?: { label?: string; description?: string }[];
}

function parseJson(text: string): unknown {
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return undefined;
  }
}

function asObject(value: unknown): Record<string, unknown> {
  return typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
}

function renderAsk(payload: unknown, askId: string | undefined): void {
  const palette = colors();
  const ask = asObject(payload);
  const questions = Array.isArray(ask.questions) ? (ask.questions as AskQuestion[]) : [];
  for (const question of questions) {
    const header = question.header ? `${question.header} ` : "";
    printErr(palette.bold(`${header}${question.text ?? ""}`));
    if (question.detail) printErr(palette.dim(question.detail));
    for (const [index, option] of (question.options ?? []).entries()) {
      const description = option.description ? palette.dim(` — ${option.description}`) : "";
      printErr(`  ${index + 1}. ${option.label ?? ""}${description}`);
    }
  }
  const id = typeof ask.ask_id === "string" ? ask.ask_id : askId;
  if (id) {
    printErr(palette.dim(`  用 \`ech0 copilot chat-answer --ask-id ${id} ...\` 回答`));
  }
}

/**
 * Renders a stream as it arrives. Named events (the catalog's advisory
 * vocabulary) get bespoke formatting; unknown names and unnamed streams fall
 * back to one line per frame, so a newer server stays usable.
 */
export function createStreamRenderer(options: { json: boolean; events: string[] }): StreamRenderer {
  const palette = colors();
  const named = new Set(options.events);
  let midLine = false;
  let reasoningLine = false;
  let failed = false;
  let sources: unknown[] = [];
  let coverage: unknown;

  const breakLine = (): void => {
    if (midLine) {
      print("");
      midLine = false;
    }
    if (reasoningLine) {
      printErr("");
      reasoningLine = false;
    }
  };

  const unknown = (frame: SseFrame): void => {
    breakLine();
    const payload = parseJson(frame.data);
    if (payload === undefined) {
      print(`[${frame.event}] ${frame.data}`);
      return;
    }
    const envelope = asObject(payload);
    const data = "code" in envelope ? envelope.data : payload;
    print(JSON.stringify(data));
  };

  const renderNamed = (frame: SseFrame): void => {
    const payload = parseJson(frame.data);
    const data = asObject(payload);

    switch (frame.event) {
      case "delta": {
        const text = typeof data.text === "string" ? data.text : frame.data;
        process.stdout.write(text);
        midLine = true;
        break;
      }
      case "reasoning": {
        const text = typeof data.text === "string" ? data.text : frame.data;
        process.stderr.write(palette.dim(text));
        reasoningLine = true;
        break;
      }
      case "searching": {
        breakLine();
        const name = typeof data.name === "string" ? data.name : "search";
        const query = typeof data.query === "string" && data.query !== "" ? ` ${data.query}` : "";
        printErr(palette.dim(`${palette.cyan("·")} ${name}${query}`));
        break;
      }
      case "sources":
        sources = Array.isArray(payload) ? payload : [];
        break;
      case "coverage":
        coverage = payload;
        break;
      case "reasoning_done": {
        breakLine();
        const ms = typeof data.duration_ms === "number" ? data.duration_ms : undefined;
        if (ms !== undefined) printErr(palette.dim(`思考 ${ms}ms`));
        break;
      }
      case "ask":
        breakLine();
        renderAsk(payload, typeof data.ask_id === "string" ? data.ask_id : undefined);
        break;
      case "ask_closed":
        break;
      case "done":
        breakLine();
        if (sources.length > 0) {
          const titles = sources
            .map((source) => asObject(source))
            .map((source) => (typeof source.title === "string" ? source.title : undefined))
            .filter((title): title is string => Boolean(title));
          printErr(palette.dim(`来源 ${sources.length} 条${titles.length > 0 ? `：${titles.join(" · ")}` : ""}`));
        }
        if (coverage !== undefined) printErr(palette.dim(`覆盖 ${JSON.stringify(coverage)}`));
        break;
      case "error": {
        breakLine();
        const message = typeof data.message === "string" ? data.message : frame.data;
        printErr(`${palette.red("错误")} ${message}`);
        failed = true;
        break;
      }
      default:
        unknown(frame);
        break;
    }
  };

  return {
    onFrame(frame) {
      if (options.json) {
        const payload = parseJson(frame.data);
        print(JSON.stringify({ event: frame.event, data: payload ?? frame.data }));
        if (frame.event === "error") failed = true;
        return;
      }
      if (named.size === 0 || !named.has(frame.event)) {
        unknown(frame);
        return;
      }
      renderNamed(frame);
    },
    finish() {
      breakLine();
      if (sources.length > 0 && !options.json) sources = [];
    },
    get failed() {
      return failed;
    },
  };
}
