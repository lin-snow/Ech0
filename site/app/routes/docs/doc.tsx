// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Link, redirect, useLocation } from "react-router";
import type { Route } from "./+types/doc";
import { DocToc } from "../../docs/DocToc";
import { LocaleSwitch } from "../../docs/LocaleSwitch";
import { getDoc } from "../../docs/registry";
import { MarkdownDoc } from "../../docs/MarkdownDoc";
import { extractTocFromMarkdown } from "../../docs/toc";
import {
  docPath,
  docsIndexPath,
  localeFromPathname,
  OG_LOCALE,
} from "../../docs/locale";
import { DOCS_UI } from "../../docs/ui";
import { absoluteUrl } from "../../site";

export function meta({ loaderData, location }: Route.MetaArgs) {
  const locale = loaderData?.locale ?? localeFromPathname(location.pathname);
  if (!loaderData?.title) {
    return [{ title: DOCS_UI[locale].docsTitle }];
  }
  const title = `${loaderData.title} — Ech0 Docs`;
  const description =
    loaderData.description || `${loaderData.title} — Ech0 documentation.`;
  return [
    { title },
    { name: "description", content: description },
    { property: "og:title", content: title },
    { property: "og:description", content: description },
    {
      property: "og:url",
      content: absoluteUrl(docPath(locale, loaderData.slug)),
    },
    { property: "og:locale", content: OG_LOCALE[locale] },
  ];
}

const LEGACY_DOC_REDIRECTS: Record<string, string> = {
  "guide/connect": "guide/federation",
  "guide/hub": "guide/federation",
  "guide/oauth": "guide/sso",
  "guide/passkey": "guide/sso",
  "design/palette": "",
  "start/credits": "start/community",
};

export async function clientLoader({
  params,
  request,
}: Route.ClientLoaderArgs) {
  const locale = localeFromPathname(new URL(request.url).pathname);
  const splat = params["*"]?.replace(/\/$/, "") ?? "";
  if (splat === "" || splat === "README") {
    return redirect(docsIndexPath(locale));
  }
  const legacy = LEGACY_DOC_REDIRECTS[splat];
  if (legacy !== undefined) {
    return redirect(docPath(locale, legacy));
  }
  const doc = getDoc(splat, locale);
  if (!doc) throw new Response("Not Found", { status: 404 });
  return {
    body: doc.body,
    title: doc.title,
    description: doc.description,
    slug: splat,
    locale,
  };
}

export default function DocsPage({ loaderData }: Route.ComponentProps) {
  const { pathname } = useLocation();
  const { body, title, description, locale } = loaderData;
  const ui = DOCS_UI[locale];
  const toc = extractTocFromMarkdown(body);
  const hasToc = toc.length > 0;
  const firstMeaningfulLine = body
    .split(/\r?\n/)
    .map((l) => l.trim())
    .find((l) => l !== "" && l !== "---");
  const hasTopH1 = !!firstMeaningfulLine?.match(/^#\s+.+$/);

  const mdProps = {
    content: body,
    tocItems: hasToc ? toc : undefined,
    locale,
  } as const;

  return (
    <>
      <div className="flex items-center justify-between pt-8">
        <Link
          to={docsIndexPath(locale)}
          className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          ← {ui.allDocs}
        </Link>
        <LocaleSwitch locale={locale} pathname={pathname} />
      </div>

      {!hasTopH1 ? (
        hasToc ? (
          <div className="mt-7 mb-2 lg:grid lg:grid-cols-[minmax(0,12rem)_minmax(0,30rem)_minmax(0,11rem)] lg:items-start lg:gap-x-8">
            <div className="hidden lg:block" aria-hidden />
            <header className="max-w-[30rem]">
              <h1 className="font-serif text-[1.2rem] font-semibold leading-snug text-sand-12">
                {title}
              </h1>
              {description ? (
                <p className="mt-1 text-[0.8125rem] leading-relaxed text-sand-11">
                  {description}
                </p>
              ) : null}
            </header>
            <div className="hidden lg:block" aria-hidden />
          </div>
        ) : (
          <header className="mx-auto mt-7 mb-2 max-w-[30rem]">
            <h1 className="font-serif text-[1.2rem] font-semibold leading-snug text-sand-12">
              {title}
            </h1>
            {description ? (
              <p className="mt-1 text-[0.8125rem] leading-relaxed text-sand-11">
                {description}
              </p>
            ) : null}
          </header>
        )
      ) : null}

      {hasToc ? (
        <div className="mt-8 lg:grid lg:grid-cols-[minmax(0,12rem)_minmax(0,30rem)_minmax(0,11rem)] lg:items-start lg:gap-x-8">
          <DocToc items={toc} label={ui.onThisPage} />
          <div className="min-w-0">
            <article className="lg:max-w-none">
              <MarkdownDoc {...mdProps} />
            </article>
          </div>
          <div className="hidden lg:block" aria-hidden />
        </div>
      ) : (
        <article className="mx-auto mt-8 max-w-[min(100%,30rem)]">
          <MarkdownDoc {...mdProps} />
        </article>
      )}
    </>
  );
}
