import type { TocItem } from "./toc";

function TocLinks({ items }: { items: TocItem[] }) {
  return (
    <ul className="space-y-1.5 border-l border-sand-6/90 pl-3">
      {items.map((item) => (
        <li key={item.id}>
          <a
            href={`#${encodeURIComponent(item.id)}`}
            className={
              item.depth === 2
                ? "block text-[0.75rem] font-medium leading-snug text-sand-11 transition-colors hover:text-sand-12"
                : "block pl-2 text-[0.6875rem] leading-snug text-sand-11/90 transition-colors hover:text-sand-12"
            }
          >
            {item.text}
          </a>
        </li>
      ))}
    </ul>
  );
}

export function DocToc({ items }: { items: TocItem[] }) {
  if (items.length === 0) return null;

  return (
    <>
      {}
      <nav className="mb-6 lg:hidden" aria-label="本页目录">
        <details className="group rounded-lg border border-sand-6/90 bg-sand-2/50 px-3 py-2">
          <summary className="cursor-pointer list-none text-[0.6875rem] font-semibold tracking-wide text-sand-11 [&::-webkit-details-marker]:hidden">
            <span className="inline-flex items-center gap-1.5">
              <span
                className="inline-block transition-transform group-open:rotate-90"
                aria-hidden
              >
                ▸
              </span>
              本页目录
            </span>
          </summary>
          <div className="mt-3 pb-1">
            <TocLinks items={items} />
          </div>
        </details>
      </nav>

      {}
      <nav className="hidden lg:block" aria-label="本页目录">
        <p className="mb-3 text-[0.625rem] font-semibold uppercase tracking-[0.14em] text-sand-11/90">
          本页目录
        </p>
        <div className="sticky top-24 max-h-[calc(100vh-8rem)] overflow-y-auto overscroll-contain pb-8">
          <TocLinks items={items} />
        </div>
      </nav>
    </>
  );
}
