// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/lin-snow/ech0/internal/authz"
)

type ToolHandler func(ctx context.Context, args map[string]any) (*ToolCallResult, error)

type ResourceHandler func(ctx context.Context, uri string) (*ResourceReadResult, error)

type ToolBinding struct {
	Handler ToolHandler
	Scopes  []string
}

type ResourceBinding struct {
	Handler ResourceHandler
	Scopes  []string
	Cache   CacheInfo
}

type registeredTool struct {
	definition           ToolDefinition
	binding              ToolBinding
	requiresConfirmation bool
}

func requiresConfirmation(def ToolDefinition) bool {
	return def.Annotations != nil &&
		def.Annotations.DestructiveHint != nil &&
		*def.Annotations.DestructiveHint
}

type registeredResource struct {
	definition ResourceDefinition
	binding    ResourceBinding
}

type registeredTemplate struct {
	definition ResourceTemplateDefinition
	binding    ResourceBinding
	prefix     string
}

type Registry struct {
	tools     []registeredTool
	resources []registeredResource
	templates []registeredTemplate
	toolIndex map[string]int
}

func NewRegistry() *Registry {
	return &Registry{
		toolIndex: make(map[string]int),
	}
}

func (r *Registry) RegisterTool(def ToolDefinition, handler ToolHandler, scopes ...string) {
	r.toolIndex[def.Name] = len(r.tools)
	r.tools = append(r.tools, registeredTool{
		definition:           def,
		binding:              ToolBinding{Handler: handler, Scopes: scopes},
		requiresConfirmation: requiresConfirmation(def),
	})
}

func (r *Registry) RegisterResource(def ResourceDefinition, handler ResourceHandler, scopes ...string) {
	cache := def.Cache.normalize()
	def.Cache = cache
	r.resources = append(r.resources, registeredResource{
		definition: def,
		binding:    ResourceBinding{Handler: handler, Scopes: scopes, Cache: cache},
	})
}

func (r *Registry) RegisterResourceTemplate(def ResourceTemplateDefinition, handler ResourceHandler, scopes ...string) {
	idx := strings.Index(def.URITemplate, "{")
	if idx <= 0 {
		panic(fmt.Sprintf("mcp: resource template %q must have a literal prefix before its first placeholder", def.URITemplate))
	}
	cache := def.Cache.normalize()
	def.Cache = cache
	r.templates = append(r.templates, registeredTemplate{
		definition: def,
		binding:    ResourceBinding{Handler: handler, Scopes: scopes, Cache: cache},
		prefix:     def.URITemplate[:idx],
	})
}

func (r *Registry) ToolDefinitions() []ToolDefinition {
	defs := make([]ToolDefinition, len(r.tools))
	for i, t := range r.tools {
		defs[i] = t.definition
	}
	return defs
}

// ToolDefinitionsFor returns the tools the caller can actually invoke. Filtering
// keeps the advertised surface aligned with the token's scopes, so models never
// pick a tool that would be rejected with insufficient_scope. Registration
// order is preserved to keep tools/list deterministic.
func (r *Registry) ToolDefinitionsFor(held []string) []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		if authz.Allows(held, t.binding.Scopes) {
			defs = append(defs, t.definition)
		}
	}
	return defs
}

func (r *Registry) ResourceDefinitions() []ResourceDefinition {
	defs := make([]ResourceDefinition, len(r.resources))
	for i, res := range r.resources {
		defs[i] = res.definition
	}
	return defs
}

func (r *Registry) ResourceDefinitionsFor(held []string) []ResourceDefinition {
	defs := make([]ResourceDefinition, 0, len(r.resources))
	for _, res := range r.resources {
		if authz.Allows(held, res.binding.Scopes) {
			defs = append(defs, res.definition)
		}
	}
	return defs
}

func (r *Registry) ResourceTemplateDefinitions() []ResourceTemplateDefinition {
	defs := make([]ResourceTemplateDefinition, len(r.templates))
	for i, tpl := range r.templates {
		defs[i] = tpl.definition
	}
	return defs
}

func (r *Registry) ResourceTemplateDefinitionsFor(held []string) []ResourceTemplateDefinition {
	defs := make([]ResourceTemplateDefinition, 0, len(r.templates))
	for _, tpl := range r.templates {
		if authz.Allows(held, tpl.binding.Scopes) {
			defs = append(defs, tpl.definition)
		}
	}
	return defs
}

// ToolRequiresConfirmation reports whether a tool is annotated destructive and
// therefore needs an explicit user confirmation on modern clients.
func (r *Registry) ToolRequiresConfirmation(name string) bool {
	idx, ok := r.toolIndex[name]
	if !ok {
		return false
	}
	return r.tools[idx].requiresConfirmation
}

func (r *Registry) LookupTool(name string) (ToolBinding, bool) {
	idx, ok := r.toolIndex[name]
	if !ok {
		return ToolBinding{}, false
	}
	return r.tools[idx].binding, true
}

func (r *Registry) LookupResource(uri string) (ResourceBinding, bool) {
	for _, res := range r.resources {
		if res.definition.URI == uri {
			return res.binding, true
		}
	}
	for _, tpl := range r.templates {
		if strings.HasPrefix(uri, tpl.prefix) {
			return tpl.binding, true
		}
	}
	return ResourceBinding{}, false
}
