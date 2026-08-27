import { Link, redirect } from "react-router";
import type { Route } from "./+types/doc";
import { DocToc } from "../../docs/DocToc";
import { getDoc } from "../../docs/registry";
import { MarkdownDoc } from "../../docs/MarkdownDoc";
import { extractTocFromMarkdown } from "../../docs/toc";
import { absoluteUrl } from "../../site";

export function meta({ loaderData }: Route.MetaArgs) {
  if (!loaderData?.title) {
    return [{ title: "Documentation — Ech0" }];
  }
  return [
    { title: `${loaderData.title} — Ech0 Docs` },
    {
      name: "description",
      content:
        loaderData.description || `${loaderData.title} — Ech0 documentation.`,
    },
    { property: "og:title", content: `${loaderData.title} — Ech0 Docs` },
    { property: "og:url", content: absoluteUrl(`/docs/${loaderData.slug}`) },
  ];
}

const LEGACY_DOC_REDIRECTS: Record<string, string> = {
  "guide/connect": "/docs/guide/federation",
  "guide/hub": "/docs/guide/federation",
  "guide/oauth": "/docs/guide/sso",
  "guide/passkey": "/docs/guide/sso",
  "design/palette": "/docs",
  "start/credits": "/docs/start/community",
};

export async function clientLoader({ params }: Route.ClientLoaderArgs) {
  const splat = params["*"]?.replace(/\/$/, "") ?? "";
  if (splat === "" || splat === "README") {
    return redirect("/docs");
  }
  const legacy = LEGACY_DOC_REDIRECTS[splat];
  if (legacy) {
    return redirect(legacy);
  }
  const doc = getDoc(splat);
  if (!doc) throw new Response("Not Found", { status: 404 });
  return {
    body: doc.body,
    title: doc.title,
    description: doc.description,
    slug: splat,
  };
}

export default function DocsPage({ loaderData }: Route.ComponentProps) {
  const { body, title, description } = loaderData;
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
  } as const;

  return (
    <>
      <div className="flex items-center pt-8">
        <Link
          to="/docs"
          className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          ← All docs
        </Link>
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
          <DocToc items={toc} />
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
