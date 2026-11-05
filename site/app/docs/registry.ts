// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { DEFAULT_LOCALE, type DocLocale } from "./locale";

const raw = import.meta.glob<string>("../../docs/**/*.md", {
  query: "?raw",
  import: "default",
  eager: true,
});

function normalizePath(p: string): string {
  return p.replace(/\\/g, "/");
}

const LOCALE_DOC = /\/docs\/(en|zh)\/(.+)\.md$/;

function buildDocIndex(): Record<DocLocale, Map<string, string>> {
  const index: Record<DocLocale, Map<string, string>> = {
    en: new Map(),
    zh: new Map(),
  };
  for (const [key, value] of Object.entries(raw)) {
    const match = normalizePath(key).match(LOCALE_DOC);
    if (!match) continue;
    index[match[1] as DocLocale].set(match[2], value);
  }
  return index;
}

const DOCS = buildDocIndex();

const FRONTMATTER = /^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/;

export type DocCard = {
  slug: string;
  title: string;
  description: string;
};

function extractTitleFromBody(md: string): string | null {
  const m = md.match(/^#\s+(.+)$/m);
  return m?.[1]?.trim() ?? null;
}

function labelFromSlug(slug: string): string {
  const base = slug.split("/").pop() ?? slug;
  return base.replace(/-/g, " ");
}

function parseFrontMatterBlock(fm: string): {
  title: string;
  description: string;
} {
  const titleM = fm.match(/^title:\s*(.+)$/m);
  const descM = fm.match(/^description:\s*(.+)$/m);
  return {
    title: titleM?.[1]?.trim() ?? "",
    description: descM?.[1]?.trim() ?? "",
  };
}

export function parseDocFile(
  raw: string,
  slug: string,
): {
  title: string;
  description: string;
  body: string;
} {
  const m = raw.match(FRONTMATTER);
  if (!m) {
    const body = prepareMarkdownContent(raw);
    return {
      title: extractTitleFromBody(body) ?? labelFromSlug(slug),
      description: "",
      body,
    };
  }
  const { title: fmTitle, description } = parseFrontMatterBlock(m[1]);
  const body = prepareMarkdownContent(m[2]);
  const title = fmTitle || extractTitleFromBody(body) || labelFromSlug(slug);
  return { title, description, body };
}

function rawDoc(slug: string, locale: DocLocale): string | undefined {
  return DOCS[locale].get(slug) ?? DOCS[DEFAULT_LOCALE].get(slug);
}

export function listDocCards(locale: DocLocale = DEFAULT_LOCALE): DocCard[] {
  const slugs = new Set<string>([
    ...DOCS[locale].keys(),
    ...DOCS[DEFAULT_LOCALE].keys(),
  ]);
  const cards: DocCard[] = [];
  const collator = new Intl.Collator(locale === "zh" ? "zh-Hans-CN" : "en");
  for (const slug of slugs) {
    if (!slug || slug === "README") continue;
    const source = rawDoc(slug, locale);
    if (!source) continue;
    const parsed = parseDocFile(source, slug);
    cards.push({
      slug,
      title: parsed.title,
      description: parsed.description,
    });
  }
  cards.sort(
    (a, b) =>
      docOrder(a.slug) - docOrder(b.slug) || collator.compare(a.title, b.title),
  );
  return cards;
}

export const DOC_HERO_SLUGS: readonly string[] = [
  "guide/overview",
  "start/getting-started",
  "start/installation",
  "start/update",
];

export function partitionDocCards(cards: DocCard[]): {
  featured: DocCard[];
  rest: DocCard[];
} {
  const heroSet = new Set(DOC_HERO_SLUGS);
  const featured: DocCard[] = [];
  const rest: DocCard[] = [];
  for (const c of cards) {
    if (heroSet.has(c.slug)) featured.push(c);
    else rest.push(c);
  }
  featured.sort(
    (a, b) => DOC_HERO_SLUGS.indexOf(a.slug) - DOC_HERO_SLUGS.indexOf(b.slug),
  );
  return { featured, rest };
}

const DOC_ORDER: readonly string[] = [
  "guide/overview",
  "start/getting-started",
  "start/installation",
  "start/update",
  "guide/editor",
  "guide/preferences",
  "start/faq",
  "guide/federation",
  "guide/sso",
  "guide/comment",
  "guide/agent",
  "guide/chat",
  "guide/embedding",
  "guide/webhook",
  "guide/accesstoken",
  "guide/cli",
  "guide/mcp",
  "guide/s3",
  "guide/datacontrol",
  "guide/capsule",
  "dev/guide",
  "start/community",
];

function docOrder(slug: string): number {
  const i = DOC_ORDER.indexOf(slug);
  return i === -1 ? 1000 : i;
}

export function getDoc(
  slug: string,
  locale: DocLocale = DEFAULT_LOCALE,
): { title: string; description: string; body: string } | null {
  const normalized = slug.replace(/^\/+|\/+$/g, "");
  if (normalized === "") return null;
  const source = rawDoc(normalized, locale);
  if (!source) return null;
  return parseDocFile(source, normalized);
}

export function prepareMarkdownContent(md: string): string {
  return md
    .replace(/\]\(imgs\//g, "](/docs-assets/imgs/")
    .replace(/\]\(\.\/imgs\//g, "](/docs-assets/imgs/");
}

export function extractTitle(md: string): string | null {
  const m = md.match(/^#\s+(.+)$/m);
  return m?.[1]?.trim() ?? null;
}
