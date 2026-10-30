// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** One dispatched server-sent event. */
export interface SseFrame {
  /** `message` when the server omits `event:`. */
  event: string;
  data: string;
  id?: string;
}

type Field = { name: string; value: string };

/** Splits one raw frame into `field: value` pairs, honouring comments. */
function parseFields(raw: string): Field[] {
  const fields: Field[] = [];
  for (const line of raw.split("\n")) {
    if (line === "" || line.startsWith(":")) continue;
    const colon = line.indexOf(":");
    const name = colon < 0 ? line : line.slice(0, colon);
    let value = colon < 0 ? "" : line.slice(colon + 1);
    if (value.startsWith(" ")) value = value.slice(1);
    fields.push({ name, value });
  }
  return fields;
}

/** Builds a frame from its fields; `undefined` when it carries no data. */
function dispatch(raw: string): SseFrame | undefined {
  let event = "message";
  let id: string | undefined;
  const data: string[] = [];

  for (const field of parseFields(raw)) {
    switch (field.name) {
      case "event":
        event = field.value === "" ? "message" : field.value;
        break;
      case "data":
        data.push(field.value);
        break;
      case "id":
        id = field.value;
        break;
      default:
        // `retry:` and unknown fields are advisory only.
        break;
    }
  }

  if (data.length === 0) return undefined;
  return id === undefined ? { event, data: data.join("\n") } : { event, data: data.join("\n"), id };
}

/**
 * Parses a `text/event-stream` body. Frames are split on blank lines and may
 * arrive split across chunks; `\r\n` and bare comments (keep-alives) are
 * tolerated.
 */
export async function* parseSse(stream: ReadableStream<Uint8Array>): AsyncGenerator<SseFrame> {
  const decoder = new TextDecoder();
  let buffer = "";
  let current = "";

  const consume = (): SseFrame[] => {
    const frames: SseFrame[] = [];
    for (;;) {
      const boundary = buffer.indexOf("\n\n");
      if (boundary < 0) break;
      current = buffer.slice(0, boundary);
      buffer = buffer.slice(boundary + 2);
      const frame = dispatch(current);
      if (frame) frames.push(frame);
    }
    return frames;
  };

  for await (const chunk of stream) {
    buffer += decoder.decode(chunk, { stream: true }).replace(/\r\n/g, "\n");
    for (const frame of consume()) yield frame;
  }

  buffer += decoder.decode();
  if (buffer.trim() !== "") {
    const frame = dispatch(buffer);
    if (frame) yield frame;
  }
}
