// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetMCPManifest() {
  return request<App.Api.MCP.Manifest>({
    url: '/mcp/manifest',
    method: 'GET',
  })
}
