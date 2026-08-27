import { Link } from "react-router";
import type { ComponentProps, ComponentPropsWithoutRef } from "react";
import { useMemo, useRef } from "react";
import "highlight.js/styles/github.css";
import ReactMarkdown from "react-markdown";
import rehypeHighlight from "rehype-highlight";
import remarkGfm from "remark-gfm";
import type { TocItem } from "./toc";

function DocLink({
  href,
  children,
}: {
  href?: string;
  children?: React.ReactNode;
}) {
  if (!href) return <span>{children}</span>;
  if (/^https?:\/\//i.test(href)) {
    return (
      <a
        href={href}
        target="_blank"
        rel="noreferrer noopener"
        className="text-sand-12 underline decoration-sand-11/45 underline-offset-2"
      >
        {children}
      </a>
    );
  }
  if (href.startsWith("/docs")) {
    return (
      <Link
        to={href}
        className="text-sand-12 underline decoration-sand-11/45 underline-offset-2"
      >
        {children}
      </Link>
    );
  }
  if (href.endsWith(".md")) {
    let path = href.replace(/\.md$/, "").replace(/^\.\//, "");
    while (path.startsWith("../")) path = path.slice(3);
    const to = path === "" || path === "README" ? "/docs" : `/docs/${path}`;
    return (
      <Link
        to={to}
        className="text-sand-12 underline decoration-sand-11/45 underline-offset-2"
      >
        {children}
      </Link>
    );
  }
  return (
    <a
      href={href}
      className="text-sand-12 underline decoration-sand-11/45 underline-offset-2"
    >
      {children}
    </a>
  );
}

const markdownClass =
  "docs-markdown max-w-none font-sans text-[0.875rem] leading-relaxed text-sand-12 [&_h1]:font-serif [&_h1]:text-xl [&_h1]:font-normal [&_h1]:leading-snug [&_h1]:text-sand-12 [&_h1]:mt-8 [&_h1]:mb-3 [&_h1]:first:mt-0 [&_h2]:font-serif [&_h2]:text-base [&_h2]:font-semibold [&_h2]:leading-snug [&_h2]:text-sand-12 [&_h2]:mt-6 [&_h2]:mb-2 [&_h3]:font-serif [&_h3]:text-[0.9375rem] [&_h3]:font-semibold [&_h3]:text-sand-12 [&_h3]:mt-5 [&_h3]:mb-1.5 [&_p]:my-2.5 [&_ul]:my-2.5 [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:my-2.5 [&_ol]:list-decimal [&_ol]:pl-5 [&_li]:my-0.5 [&_blockquote]:border-l-2 [&_blockquote]:border-sand-6 [&_blockquote]:pl-3 [&_blockquote]:text-sand-12/90 [&_blockquote]:text-[0.875rem] [&_pre]:my-3 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:border [&_pre]:border-sand-6/60 [&_pre]:bg-sand-2 [&_pre]:p-0 [&_pre]:text-[0.8125rem] [&_pre_code]:bg-transparent [&_pre_code]:p-0 [&_table]:my-3 [&_table]:w-full [&_table]:border-collapse [&_table]:text-[0.8125rem] [&_th]:border [&_th]:border-sand-6 [&_th]:bg-sand-2 [&_th]:px-2 [&_th]:py-1.5 [&_th]:text-left [&_th]:font-serif [&_th]:font-semibold [&_td]:border [&_td]:border-sand-6 [&_td]:px-2 [&_td]:py-1.5 [&_img]:my-3 [&_img]:max-w-full [&_img]:rounded-lg [&_hr]:my-6 [&_hr]:border-sand-6";

export function MarkdownDoc({
  content,
  tocItems,
}: {
  content: string;
  tocItems?: TocItem[];
}) {
  const tocPtr = useRef(0);
  tocPtr.current = 0;
  const rehypePlugins = useMemo(() => [rehypeHighlight], []);

  const components = useMemo(() => {
    const toc = tocItems ?? [];
    const base: Record<string, React.ComponentType<unknown>> = {
      a: DocLink as React.ComponentType<unknown>,
    };

    if (toc.length === 0) {
      return base;
    }

    const H2 = ({ children, ...props }: ComponentPropsWithoutRef<"h2">) => {
      const item = toc[tocPtr.current];
      if (item?.depth === 2) {
        tocPtr.current++;
        return (
          <h2 id={item.id} {...props}>
            {children}
          </h2>
        );
      }
      return <h2 {...props}>{children}</h2>;
    };

    const H3 = ({ children, ...props }: ComponentPropsWithoutRef<"h3">) => {
      const item = toc[tocPtr.current];
      if (item?.depth === 3) {
        tocPtr.current++;
        return (
          <h3 id={item.id} {...props}>
            {children}
          </h3>
        );
      }
      return <h3 {...props}>{children}</h3>;
    };

    return {
      ...base,
      h2: H2,
      h3: H3,
    };
  }, [tocItems]);

  return (
    <div className={markdownClass}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={
          rehypePlugins as ComponentProps<typeof ReactMarkdown>["rehypePlugins"]
        }
        components={components}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
