// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const raw = import.meta.glob<string>("../../docs/**/*.md", {
  query: "?raw",
  import: "default",
  eager: true,
});

function normalizePath(p: string): string {
  return p.replace(/\\/g, "/");
}

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

export function listDocCards(): DocCard[] {
  const cards: DocCard[] = [];
  for (const key of Object.keys(raw)) {
    const nk = normalizePath(key);
    const slug = nk.replace(/^.*\/docs\//, "").replace(/\.md$/, "");
    if (!slug || slug === "README") continue;
    const parsed = parseDocFile(raw[key] as string, slug);
    cards.push({
      slug,
      title: parsed.title,
      description: parsed.description,
    });
  }
  cards.sort(
    (a, b) =>
      docOrder(a.slug) - docOrder(b.slug) ||
      a.title.localeCompare(b.title, "zh-Hans-CN"),
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
): { title: string; description: string; body: string } | null {
  const normalized = slug.replace(/^\/+|\/+$/g, "");
  if (normalized === "") return null;
  const key = Object.keys(raw).find((k) => {
    const nk = normalizePath(k);
    return nk.endsWith(`/${normalized}.md`) || nk.endsWith(`${normalized}.md`);
  });
  if (!key) return null;
  return parseDocFile(raw[key] as string, normalized);
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
