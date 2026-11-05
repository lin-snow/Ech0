// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Link } from "react-router";
import {
  alternateLocalePath,
  HTML_LANG,
  LOCALE_LABEL,
  type DocLocale,
  otherLocale,
} from "./locale";
import { DOCS_UI } from "./ui";

export function LocaleSwitch({
  locale,
  pathname,
}: {
  locale: DocLocale;
  pathname: string;
}) {
  const alternative = otherLocale(locale);
  return (
    <nav
      className="flex items-center gap-2 text-xs"
      aria-label={DOCS_UI[locale].language}
    >
      <span
        className="font-medium text-sand-12"
        aria-current="true"
        lang={HTML_LANG[locale]}
      >
        {LOCALE_LABEL[locale]}
      </span>
      <span className="text-sand-6" aria-hidden>
        /
      </span>
      <Link
        to={alternateLocalePath(pathname, alternative)}
        hrefLang={alternative}
        lang={HTML_LANG[alternative]}
        className="font-medium text-sand-11 transition-colors hover:text-sand-12"
      >
        {LOCALE_LABEL[alternative]}
      </Link>
    </nav>
  );
}
