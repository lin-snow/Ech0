// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export function siteUrl(): string {
  const raw = import.meta.env.VITE_SITE_URL ?? "https://www.ech0.app";
  return raw.replace(/\/$/, "");
}

export function absoluteUrl(path: string): string {
  const base = siteUrl();
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${base}${p}`;
}

export const SITE_NAME = "Ech0";

export const DEFAULT_DESCRIPTION =
  "Let your thoughts flow: a personal timeline on your server—self-hosted, AGPL-3.0, ad-free and platform-free.";
