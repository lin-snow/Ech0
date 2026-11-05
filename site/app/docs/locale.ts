// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const DOC_LOCALES = ["en", "zh"] as const;

export type DocLocale = (typeof DOC_LOCALES)[number];

/** English is the default documentation language. */
export const DEFAULT_LOCALE: DocLocale = "en";

export const LOCALE_LABEL: Record<DocLocale, string> = {
  en: "English",
  zh: "中文",
};

export const HTML_LANG: Record<DocLocale, string> = {
  en: "en",
  zh: "zh-Hans",
};

export const OG_LOCALE: Record<DocLocale, string> = {
  en: "en_US",
  zh: "zh_CN",
};

export function isDocLocale(value: string): value is DocLocale {
  return (DOC_LOCALES as readonly string[]).includes(value);
}

/** `/zh` and `/zh/...` are Chinese; everything else is the default locale. */
export function localeFromPathname(pathname: string): DocLocale {
  return pathname === "/zh" || pathname.startsWith("/zh/") ? "zh" : "en";
}

export function docsIndexPath(locale: DocLocale): string {
  return locale === "zh" ? "/zh/docs" : "/docs";
}

export function docPath(locale: DocLocale, slug: string): string {
  const clean = slug.replace(/^\/+|\/+$/g, "");
  return clean === ""
    ? docsIndexPath(locale)
    : `${docsIndexPath(locale)}/${clean}`;
}

/** Path of the same page in `target`, e.g. `/docs/a` ↔ `/zh/docs/a`. */
export function alternateLocalePath(
  pathname: string,
  target: DocLocale,
): string {
  const bare = pathname.replace(/^\/zh(?=\/|$)/, "") || "/";
  if (target === "zh") return bare === "/" ? "/zh" : `/zh${bare}`;
  return bare;
}

export function otherLocale(locale: DocLocale): DocLocale {
  return locale === "zh" ? "en" : "zh";
}
