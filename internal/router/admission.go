// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"errors"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/pkg/role"
)

// Access is how an operation is reached. It is declared rather than inferred
// from the scopes: "interactive session only" is not a capability, and a public
// operation has no scopes yet still has to justify being public.
type Access uint8

const (
	// AccessPublic requires no credential. The declaration must carry a reason.
	AccessPublic Access = iota
	// AccessOptional admits anonymous callers and attaches a principal when a
	// credential is present, so the handler can widen what it returns.
	AccessOptional
	// AccessSession requires an interactive browser session. Access tokens are
	// rejected regardless of scopes: account security and token administration
	// are never delegated.
	AccessSession
	// AccessToken requires authentication and at least one declared scope.
	// Sessions satisfy the scope check through their role-derived scope set.
	AccessToken
)

// endpoint is one operation's admission declaration: who may enter and with
// which capabilities. It is a value with no wiring in it — the middleware chain
// and the OpenAPI security requirement are both derived from it by the registry,
// so the document and the gate cannot disagree.
//
// The rank gate is not declared: a capability set already determines which
// roles can hold it. Resource-level decisions (including rank over a target)
// still run inside the service; this is the coarse gate.
type endpoint struct {
	access Access
	scopes []string
	auds   []string
	reason string
}

// public admits anonymous callers. The reason is required and checked at boot:
// an operation open to the internet is a decision, not an omission. It is also
// published as the operation description, so the rationale reaches readers of
// the spec instead of living only in code.
func public(reason string) endpoint {
	return endpoint{access: AccessPublic, reason: reason}
}

// optional admits anonymous callers but resolves a principal when a credential
// is present.
func optional() endpoint {
	return endpoint{access: AccessOptional}
}

// session admits only interactive browser sessions.
func session() endpoint {
	return endpoint{access: AccessSession}
}

// scoped admits credentials holding every one of scopes.
func scoped(scopes ...string) endpoint {
	return endpoint{access: AccessToken, scopes: append([]string(nil), scopes...)}
}

// audience restricts an operation to tokens issued for one of auds. Sessions
// pass: their audience is the instance itself.
func (e endpoint) audience(auds ...string) endpoint {
	e.auds = append([]string(nil), auds...)
	return e
}

// security is the OpenAPI security requirement that mirrors the gate.
func (e endpoint) security() []map[string][]string {
	switch e.access {
	case AccessPublic:
		return nil
	case AccessSession:
		return humares.Secured()
	default:
		return humares.Secured(e.scopes...)
	}
}

// rankGate is the minimum role the declared scopes imply, or "" when they imply
// nothing above member. Requiring member grants nothing: every authenticated
// principal already holds it.
func (e endpoint) rankGate() role.Role {
	var floor role.Role
	for _, scope := range e.scopes {
		if r := authz.MinimumRole(scope); r.Rank() > floor.Rank() {
			floor = r
		}
	}
	if floor.Rank() <= role.Member.Rank() {
		return ""
	}
	return floor
}

// chain is the middleware that enforces the declaration.
func (a *apiRegistry) chain(e endpoint) huma.Middlewares {
	switch e.access {
	case AccessPublic:
		return nil
	case AccessOptional:
		return huma.Middlewares{
			humares.Bridge(middleware.NoCache()),
			humares.Bridge(middleware.OptionalAuth(a.revoker)),
		}
	case AccessSession:
		return huma.Middlewares{
			humares.Bridge(middleware.NoCache()),
			humares.Bridge(middleware.RequireAuth(a.revoker)),
			humares.Bridge(middleware.RequireSession()),
		}
	default:
		chain := huma.Middlewares{
			humares.Bridge(middleware.NoCache()),
			humares.Bridge(middleware.RequireAuth(a.revoker)),
			humares.Bridge(middleware.RequireScopes(e.scopes...)),
		}
		if gate := e.rankGate(); gate != "" {
			chain = append(chain, humares.Bridge(middleware.RequireMinRank(gate)))
		}
		if len(e.auds) > 0 {
			chain = append(chain, humares.Bridge(middleware.RequireAudience(e.auds...)))
		}
		return chain
	}
}

// validate reports declaration mistakes that would otherwise become silent
// weakenings of the gate: a field used by the wrong access class, an unknown
// scope, a public operation without a rationale.
func (e endpoint) validate(key string) []error {
	var problems []error

	if e.access == AccessPublic && e.reason == "" {
		problems = append(problems, fmt.Errorf("public endpoint %s must state a reason", key))
	}
	if e.access != AccessPublic && e.reason != "" {
		problems = append(problems, fmt.Errorf("endpoint %s declares a reason but is not public", key))
	}
	if e.access != AccessToken {
		if len(e.scopes) > 0 {
			problems = append(problems, fmt.Errorf("endpoint %s declares scopes outside AccessToken", key))
		}
		if len(e.auds) > 0 {
			problems = append(problems, fmt.Errorf("endpoint %s declares an audience outside AccessToken", key))
		}
		return problems
	}

	if len(e.scopes) == 0 {
		problems = append(problems, fmt.Errorf("token endpoint %s declares no scope", key))
	}
	for _, scope := range e.scopes {
		if !authz.IsValidScope(scope) {
			problems = append(problems, fmt.Errorf("endpoint %s declares unknown scope %q", key, scope))
		}
	}
	for _, audience := range e.auds {
		if !authz.IsValidAudience(audience) {
			problems = append(problems, fmt.Errorf("endpoint %s declares unknown audience %q", key, audience))
		}
	}
	return problems
}

// verify proves the declarations are honest. It runs at boot and when the spec
// is generated, so a misdeclared operation fails startup instead of shipping a
// gate that is weaker than it reads. Every problem is reported at once: fixing
// declarations should not require one restart per mistake.
func (a *apiRegistry) verify() error {
	var problems []error
	seen := make(map[string]struct{}, len(a.endpoints))
	required := make(map[string]struct{}, authz.ScopeCount())

	for _, d := range a.endpoints {
		key := d.method + " " + d.path
		if _, dup := seen[key]; dup {
			problems = append(problems, fmt.Errorf("duplicate endpoint %s", key))
			continue
		}
		seen[key] = struct{}{}

		problems = append(problems, d.endpoint.validate(key)...)
		for _, scope := range d.endpoint.scopes {
			required[scope] = struct{}{}
		}
	}

	// An administrative capability that gates no operation is either a
	// declaration gap or dead vocabulary. Read-side scopes are exempt: broad
	// reads are public at the transport layer, so `echo:read` and friends earn
	// their keep in resource decisions and in MCP tool bindings.
	for _, scope := range authz.AllScopes() {
		if !authz.MinimumRole(scope).AtLeast(role.Admin) {
			continue
		}
		if _, ok := required[scope]; !ok {
			problems = append(problems, fmt.Errorf("administrative scope %q is never required by an endpoint", scope))
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("router: %w", errors.Join(problems...))
	}
	return nil
}
