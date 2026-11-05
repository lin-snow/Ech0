// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("privacy", "routes/privacy.tsx", { id: "privacy-en" }),
  route("zh/privacy", "routes/privacy.tsx", { id: "privacy-zh" }),
  route("docs", "routes/docs/layout.tsx", { id: "docs-en" }, [
    index("routes/docs/index.tsx", { id: "docs-en-index" }),
    route("*", "routes/docs/doc.tsx", { id: "docs-en-doc" }),
  ]),
  route("zh/docs", "routes/docs/layout.tsx", { id: "docs-zh" }, [
    index("routes/docs/index.tsx", { id: "docs-zh-index" }),
    route("*", "routes/docs/doc.tsx", { id: "docs-zh-doc" }),
  ]),
] satisfies RouteConfig;
