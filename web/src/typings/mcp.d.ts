// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace MCP {
      type Tool = {
        name: string
        title?: string
        description: string
        scopes: string[]
        read_only: boolean
        destructive: boolean
      }

      type Resource = {
        uri: string
        name: string
        title?: string
        description?: string
        scopes: string[]
      }

      type ResourceTemplate = {
        uri_template: string
        name: string
        title?: string
        description?: string
        scopes: string[]
      }

      type Manifest = {
        path: string
        transport: string
        audience: string
        protocol_versions: string[]
        tools: Tool[]
        resources: Resource[]
        resource_templates: ResourceTemplate[]
      }
    }
  }
}
