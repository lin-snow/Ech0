// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { fileURLToPath } from "node:url";

/**
 * Client types + Zod schemas are generated from the Huma spec, which is the
 * single source of truth (`just openapi` / `just openapi-check` in the repo root).
 * Raw-Gin endpoints (login, refresh, SSE, uploads) are not in the spec and stay
 * hand-written under `src/api/`.
 */
export default {
  input: fileURLToPath(new URL("../../internal/openapi/openapi.yaml", import.meta.url)),
  output: {
    path: "src/api/generated",
    header: [
      "// SPDX-License-Identifier: AGPL-3.0-or-later",
      "// Copyright (C) 2025-2026 lin-snow",
    ],
  },
  plugins: ["@hey-api/typescript", "zod"],
};
