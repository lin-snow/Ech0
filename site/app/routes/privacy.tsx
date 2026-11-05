// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Link, useLocation } from "react-router";
import type { Route } from "./+types/privacy";
import { MarkdownDoc } from "../docs/MarkdownDoc";
import { LocaleSwitch } from "../docs/LocaleSwitch";
import { parseDocFile } from "../docs/registry";
import {
  alternateLocalePath,
  localeFromPathname,
  OG_LOCALE,
} from "../docs/locale";
import { DOCS_UI } from "../docs/ui";
import { absoluteUrl } from "../site";
import privacyEn from "../../content/privacy/en.md?raw";
import privacyZh from "../../content/privacy/zh.md?raw";

export function meta({ loaderData, location }: Route.MetaArgs) {
  const locale = loaderData?.locale ?? localeFromPathname(location.pathname);
  if (!loaderData) {
    return [{ title: "Privacy — Ech0" }];
  }
  return [
    { title: `${loaderData.title} — Ech0` },
    {
      name: "description",
      content: loaderData.description || loaderData.title,
    },
    { property: "og:title", content: `${loaderData.title} — Ech0` },
    {
      property: "og:url",
      content: absoluteUrl(alternateLocalePath("/privacy", locale)),
    },
    { property: "og:locale", content: OG_LOCALE[locale] },
  ];
}

export async function clientLoader({ request }: Route.ClientLoaderArgs) {
  const locale = localeFromPathname(new URL(request.url).pathname);
  const source = locale === "zh" ? privacyZh : privacyEn;
  return { locale, ...parseDocFile(source, "privacy") };
}

export default function PrivacyPage({ loaderData }: Route.ComponentProps) {
  const { pathname } = useLocation();
  const { body, locale } = loaderData;

  return (
    <div className="min-h-screen bg-app">
      <div className="mx-auto w-full max-w-[min(100%,30rem)] px-5 pb-24">
        <header className="flex items-center justify-between pt-8">
          <Link
            to="/"
            className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
          >
            ← {DOCS_UI[locale].home}
          </Link>
          <LocaleSwitch locale={locale} pathname={pathname} />
        </header>

        <article className="mt-8">
          <MarkdownDoc content={body} locale={locale} />
        </article>
      </div>
    </div>
  );
}
