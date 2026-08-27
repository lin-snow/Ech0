// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
  readonly VITE_HUB_SUBMIT_ISSUE_URL?: string
  readonly VITE_HUB_SITE_ORIGIN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
