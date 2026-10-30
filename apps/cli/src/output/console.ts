// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** Exit codes shared by every command. */
export const EXIT_OK = 0;
export const EXIT_ERROR = 1;
export const EXIT_USAGE = 2;

/** Bad invocation (missing flag, unusable input): maps to EXIT_USAGE. */
export class UsageError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "UsageError";
  }
}

export interface ErrorDetails {
  status?: number;
  errorCode?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
}

/** Any server- or network-side failure: maps to EXIT_ERROR. */
export class Ech0Error extends Error {
  readonly details: ErrorDetails;

  constructor(message: string, details: ErrorDetails = {}) {
    super(message);
    this.name = "Ech0Error";
    this.details = details;
  }
}

type Style = (text: string) => string;

export interface Palette {
  bold: Style;
  dim: Style;
  red: Style;
  green: Style;
  yellow: Style;
  cyan: Style;
}

let colorEnabled = false;

function ansi(code: string): Style {
  return (text) => (colorEnabled ? `\u001b[${code}m${text}\u001b[0m` : text);
}

const palette: Palette = {
  bold: ansi("1"),
  dim: ansi("2"),
  red: ansi("31"),
  green: ansi("32"),
  yellow: ansi("33"),
  cyan: ansi("36"),
};

/** Enables or disables ANSI styling for the whole process. */
export function setColorEnabled(enabled: boolean): void {
  colorEnabled = enabled;
}

/** Color is on when the target stream is a TTY and `NO_COLOR` is unset. */
export function shouldUseColor(
  stream: { isTTY?: boolean } = process.stdout,
  env: NodeJS.ProcessEnv = process.env,
): boolean {
  return Boolean(stream.isTTY) && !env.NO_COLOR;
}

export function colors(): Palette {
  return palette;
}

/** Writes one data line to stdout. */
export function print(line = ""): void {
  process.stdout.write(`${line}\n`);
}

/** Writes one diagnostic line to stderr. */
export function printErr(line = ""): void {
  process.stderr.write(`${line}\n`);
}
