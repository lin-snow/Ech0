// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { UsageError } from "../output/console.ts";

/** Upload operation the CLI calls when a file must become a server file id. */
export const UPLOAD_OPERATION = "file.upload";

/** Upload targets the server accepts; `external` is URL registration, not an upload. */
const UPLOAD_STORAGE_TYPES = new Set(["local", "s3", "object"]);

export function assertUploadStorage(value: string): void {
  if (!UPLOAD_STORAGE_TYPES.has(value)) {
    throw new UsageError(
      `--storage-type 只支持 ${[...UPLOAD_STORAGE_TYPES].join(" / ")}；external 文件请用 \`ech0 file external --url <url>\` 注册`,
    );
  }
}

const CATEGORY_BY_EXTENSION: Record<string, string> = {
  png: "image",
  jpg: "image",
  jpeg: "image",
  gif: "image",
  webp: "image",
  avif: "image",
  bmp: "image",
  svg: "image",
  mp4: "video",
  mov: "video",
  webm: "video",
  mkv: "video",
  avi: "video",
  mp3: "audio",
  wav: "audio",
  flac: "audio",
  aac: "audio",
  ogg: "audio",
  m4a: "audio",
  pdf: "pdf",
  md: "markdown",
  markdown: "markdown",
};

/** Mirrors the server's category buckets so limits and layout stay right. */
export function categoryForFile(path: string): string {
  const extension = path.split(".").pop()?.toLowerCase() ?? "";
  return CATEGORY_BY_EXTENSION[extension] ?? "file";
}
