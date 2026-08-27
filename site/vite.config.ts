// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import type { Plugin } from "vite";
import { defineConfig } from "vite";

function reactRouterDevSpuriousRequestFilter(): Plugin {
  return {
    name: "react-router-dev-spurious-request-filter",
    enforce: "pre",
    apply: "serve",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const pathname = req.url?.split("?")[0] ?? "";
        if (
          pathname.startsWith("/@vite-plugin-pwa/") ||
          pathname.startsWith("/.well-known/") ||
          pathname === "/sw.js" ||
          /^\/src\/.*\.(ts|tsx|js|jsx|mjs|cjs)$/.test(pathname)
        ) {
          res.statusCode = 404;
          res.end();
          return;
        }
        next();
      });
    },
  };
}

export default defineConfig({
  plugins: [
    tailwindcss(),
    reactRouterDevSpuriousRequestFilter(),
    reactRouter(),
  ],
  resolve: {
    tsconfigPaths: true,
  },
});
