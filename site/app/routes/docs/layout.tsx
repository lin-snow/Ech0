// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Outlet, useLocation } from "react-router";
import { useEffect } from "react";
import type { Route } from "./+types/layout";
import {
  DEFAULT_LOCALE,
  HTML_LANG,
  localeFromPathname,
} from "../../docs/locale";
import { DOCS_UI } from "../../docs/ui";

export function meta({ location }: Route.MetaArgs) {
  const locale = localeFromPathname(location.pathname);
  return [{ title: DOCS_UI[locale].docsTitle }];
}

export default function DocsLayout() {
  const { pathname } = useLocation();
  const locale = localeFromPathname(pathname);

  useEffect(() => {
    document.documentElement.lang = HTML_LANG[locale];
    return () => {
      document.documentElement.lang = HTML_LANG[DEFAULT_LOCALE];
    };
  }, [locale]);

  return (
    <div className="min-h-screen bg-app">
      <div className="mx-auto w-full max-w-[min(100%,30rem)] px-5 pb-24 lg:max-w-[min(100%,72rem)] lg:px-8">
        <Outlet />
      </div>
    </div>
  );
}
