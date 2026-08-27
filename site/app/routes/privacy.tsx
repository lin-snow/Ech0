import { Link } from "react-router";
import type { Route } from "./+types/privacy";
import { MarkdownDoc } from "../docs/MarkdownDoc";
import { parseDocFile } from "../docs/registry";
import { absoluteUrl } from "../site";
import privacyMd from "../../content/privacy.md?raw";

export function meta({ loaderData }: Route.MetaArgs) {
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
    { property: "og:url", content: absoluteUrl("/privacy") },
  ];
}

export const links: Route.LinksFunction = () => [
  { rel: "canonical", href: absoluteUrl("/privacy") },
];

export async function clientLoader() {
  return parseDocFile(privacyMd, "privacy");
}

export default function PrivacyPage({ loaderData }: Route.ComponentProps) {
  const { body } = loaderData;

  return (
    <div className="min-h-screen bg-app">
      <div className="mx-auto w-full max-w-[min(100%,30rem)] px-5 pb-24">
        <header className="flex items-center justify-between pt-8">
          <Link
            to="/"
            className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
          >
            ← Home
          </Link>
        </header>

        <article className="mt-8">
          <MarkdownDoc content={body} />
        </article>
      </div>
    </div>
  );
}
