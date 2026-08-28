// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

type Manifest struct {
	Path              string                     `json:"path"`
	Transport         string                     `json:"transport"`
	Audience          string                     `json:"audience"`
	ProtocolVersions  []string                   `json:"protocol_versions"`
	Tools             []ManifestTool             `json:"tools"`
	Resources         []ManifestResource         `json:"resources"`
	ResourceTemplates []ManifestResourceTemplate `json:"resource_templates"`
}

type ManifestTool struct {
	Name        string   `json:"name"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
	ReadOnly    bool     `json:"read_only"`
	Destructive bool     `json:"destructive"`
}

type ManifestResource struct {
	URI         string   `json:"uri"`
	Name        string   `json:"name"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Scopes      []string `json:"scopes"`
}

type ManifestResourceTemplate struct {
	URITemplate string   `json:"uri_template"`
	Name        string   `json:"name"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Scopes      []string `json:"scopes"`
}

func hintEnabled(hint *bool) bool {
	return hint != nil && *hint
}

func (r *Registry) Manifest() Manifest {
	manifest := Manifest{
		Path:              MCPEndpointPath,
		Transport:         TransportStreamableHTTP,
		Audience:          authModel.AudienceMCPRemote,
		ProtocolVersions:  SupportedVersions,
		Tools:             make([]ManifestTool, 0, len(r.tools)),
		Resources:         make([]ManifestResource, 0, len(r.resources)),
		ResourceTemplates: make([]ManifestResourceTemplate, 0, len(r.templates)),
	}

	for _, tool := range r.tools {
		entry := ManifestTool{
			Name:        tool.definition.Name,
			Title:       tool.definition.Title,
			Description: tool.definition.Description,
			Scopes:      tool.binding.Scopes,
		}
		if a := tool.definition.Annotations; a != nil {
			entry.ReadOnly = hintEnabled(a.ReadOnlyHint)
			entry.Destructive = hintEnabled(a.DestructiveHint)
		}
		manifest.Tools = append(manifest.Tools, entry)
	}

	for _, resource := range r.resources {
		manifest.Resources = append(manifest.Resources, ManifestResource{
			URI:         resource.definition.URI,
			Name:        resource.definition.Name,
			Title:       resource.definition.Title,
			Description: resource.definition.Description,
			Scopes:      resource.binding.Scopes,
		})
	}

	for _, template := range r.templates {
		manifest.ResourceTemplates = append(manifest.ResourceTemplates, ManifestResourceTemplate{
			URITemplate: template.definition.URITemplate,
			Name:        template.definition.Name,
			Title:       template.definition.Title,
			Description: template.definition.Description,
			Scopes:      template.binding.Scopes,
		})
	}

	return manifest
}

type (
	GetManifestInput  struct{}
	GetManifestOutput = commonModel.Result[Manifest]
)

func (h *Handler) GetManifest(_ context.Context, _ *GetManifestInput) (GetManifestOutput, error) {
	return commonModel.OK(h.server.registry.Manifest()), nil
}
