import { Link, useLocation } from "react-router";
import type { Route } from "./+types/index";
import { listDocCards, partitionDocCards } from "../../docs/registry";
import { LocaleSwitch } from "../../docs/LocaleSwitch";
import { docPath, localeFromPathname, OG_LOCALE } from "../../docs/locale";
import { DOCS_UI } from "../../docs/ui";

export function meta({ loaderData, location }: Route.MetaArgs) {
  const locale = loaderData?.locale ?? localeFromPathname(location.pathname);
  return [
    { title: DOCS_UI[locale].docsTitle },
    {
      name: "description",
      content: DOCS_UI[locale].docsDescription,
    },
    { property: "og:locale", content: OG_LOCALE[locale] },
  ];
}

export async function clientLoader({ request }: Route.ClientLoaderArgs) {
  const locale = localeFromPathname(new URL(request.url).pathname);
  return { locale, ...partitionDocCards(listDocCards(locale)) };
}

function DocHeroIcon({ slug }: { slug: string }) {
  if (slug === "guide/overview") {
    return (
      <svg
        className="size-5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden
      >
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
        <path d="M8 7h8M8 11h6" />
      </svg>
    );
  }
  if (slug === "start/getting-started") {
    return (
      <svg
        className="size-5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden
      >
        <path d="M12 2v4" />
        <path d="m4.93 4.93 2.83 2.83" />
        <path d="M2 12h4" />
        <path d="m4.93 19.07 2.83-2.83" />
        <path d="M12 18v4" />
        <path d="m19.07 19.07-2.83-2.83" />
        <path d="M18 12h4" />
        <path d="m19.07 4.93-2.83 2.83" />
        <circle cx="12" cy="12" r="4" />
      </svg>
    );
  }
  if (slug === "start/installation") {
    return (
      <svg
        className="size-5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden
      >
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="7 10 12 15 17 10" />
        <line x1="12" x2="12" y1="15" y2="3" />
      </svg>
    );
  }
  if (slug === "start/update") {
    return (
      <svg
        className="size-5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden
      >
        <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
        <path d="M3 3v5h5" />
        <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
        <path d="M16 16h5v5" />
      </svg>
    );
  }
  return (
    <svg
      className="size-5"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="7 10 12 15 17 10" />
      <line x1="12" x2="12" y1="15" y2="3" />
    </svg>
  );
}

export default function DocsIndex({ loaderData }: Route.ComponentProps) {
  const { pathname } = useLocation();
  const { featured, rest, locale } = loaderData;
  const ui = DOCS_UI[locale];

  return (
    <div className="mx-auto w-full max-w-[min(100%,30rem)]">
      <header className="flex items-center justify-between pt-8">
        <Link
          to="/"
          className="text-xs font-medium text-sand-11 transition-colors hover:text-sand-12"
        >
          ← {ui.home}
        </Link>
        <LocaleSwitch locale={locale} pathname={pathname} />
      </header>

      <h1 className="mt-8 font-serif text-xl font-normal leading-snug text-sand-12">
        {ui.documentation}
      </h1>
      <p className="mt-2 max-w-[42ch] text-[0.8125rem] leading-relaxed text-sand-11">
        {ui.intro}
      </p>

      {featured.length > 0 ? (
        <section className="mt-10" aria-labelledby="docs-start-here">
          <h2
            id="docs-start-here"
            className="text-[0.6875rem] font-semibold uppercase tracking-[0.12em] text-sand-11"
          >
            {ui.startHere}
          </h2>
          <ul className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
            {featured.map((card) => (
              <li key={card.slug} className="flex min-h-0">
                <Link
                  to={docPath(locale, card.slug)}
                  className="group flex h-full min-h-[6.75rem] w-full gap-3.5 rounded-xl border border-sand-6 bg-gradient-to-br from-sand-2 via-sand-2 to-sand-2/70 px-4 py-3.5 text-left shadow-[0_1px_2px_rgba(33,32,28,0.04)] ring-1 ring-sand-6/35 transition hover:border-sand-11/25 hover:ring-sand-11/10"
                >
                  <span
                    className="flex h-11 w-11 shrink-0 items-center justify-center self-start rounded-lg bg-sand-12/[0.06] text-sand-11 transition group-hover:bg-sand-12/[0.09] group-hover:text-sand-12"
                    aria-hidden
                  >
                    <DocHeroIcon slug={card.slug} />
                  </span>
                  <span className="flex min-w-0 flex-1 flex-col justify-center">
                    <span className="block font-serif text-[0.9375rem] font-semibold leading-snug text-sand-12">
                      {card.title}
                    </span>
                    <span className="mt-1.5 block min-h-[2.75rem] text-[0.8125rem] leading-relaxed text-sand-11">
                      {card.description ? (
                        <span className="line-clamp-2">{card.description}</span>
                      ) : null}
                    </span>
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      <section
        className={featured.length > 0 ? "mt-12" : "mt-10"}
        aria-labelledby="docs-more"
      >
        <h2
          id="docs-more"
          className="text-[0.6875rem] font-semibold uppercase tracking-[0.12em] text-sand-11"
        >
          {featured.length > 0 ? ui.allGuides : ui.browse}
        </h2>
        <ul className="mt-4 grid grid-cols-2 gap-2">
          {rest.map((card) => (
            <li key={card.slug}>
              <Link
                to={docPath(locale, card.slug)}
                className="flex min-h-[3.25rem] flex-col justify-center gap-0.5 rounded-lg border border-sand-6/90 bg-sand-2/40 px-2 py-1.5 text-left shadow-[0_1px_2px_rgba(33,32,28,0.03)] transition-colors hover:border-sand-11/25 hover:bg-sand-2/80"
              >
                <span className="font-serif text-[0.6875rem] font-semibold leading-tight text-sand-12 line-clamp-2">
                  {card.title}
                </span>
                {card.description ? (
                  <span className="line-clamp-1 text-[0.625rem] leading-tight text-sand-11">
                    {card.description}
                  </span>
                ) : null}
              </Link>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
