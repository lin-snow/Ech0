// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/authz"
)

func TestEndpointValidate(t *testing.T) {
	tests := []struct {
		name    string
		e       endpoint
		wantMsg string
	}{
		{
			name:    "public without reason",
			e:       endpoint{access: AccessPublic},
			wantMsg: "must state a reason",
		},
		{
			name:    "public with reason is valid",
			e:       public("liveness probe"),
			wantMsg: "",
		},
		{
			name:    "reason on a token endpoint",
			e:       endpoint{access: AccessToken, scopes: []string{authz.ScopeEchoRead}, reason: "stray"},
			wantMsg: "declares a reason but is not public",
		},
		{
			name:    "scopes on a session endpoint",
			e:       endpoint{access: AccessSession, scopes: []string{authz.ScopeEchoRead}},
			wantMsg: "scopes outside AccessToken",
		},
		{
			name:    "audience on a public endpoint",
			e:       endpoint{access: AccessPublic, reason: "x", auds: []string{authz.AudienceCLI}},
			wantMsg: "audience outside AccessToken",
		},
		{
			name:    "token endpoint without scope",
			e:       scoped(),
			wantMsg: "declares no scope",
		},
		{
			name:    "unknown scope",
			e:       scoped("echo:everything"),
			wantMsg: `unknown scope "echo:everything"`,
		},
		{
			name:    "unknown audience",
			e:       scoped(authz.ScopeEchoRead).audience("carrier-pigeon"),
			wantMsg: `unknown audience "carrier-pigeon"`,
		},
		{
			name:    "scoped declaration is valid",
			e:       scoped(authz.ScopeAdminUser).audience(authz.AudienceCLI),
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := tt.e.validate("GET /test")
			got := ""
			for _, p := range problems {
				got += p.Error() + "\n"
			}
			if tt.wantMsg == "" {
				if len(problems) != 0 {
					t.Fatalf("expected no problems, got %v", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantMsg) {
				t.Fatalf("expected problem containing %q, got %v", tt.wantMsg, got)
			}
		})
	}
}

func TestVerifyReportsEveryProblem(t *testing.T) {
	api := &apiRegistry{endpoints: []declaredEndpoint{
		{method: "GET", path: "/a", endpoint: endpoint{access: AccessPublic}},
		{method: "GET", path: "/b", endpoint: scoped()},
		{method: "GET", path: "/a", endpoint: scoped(authz.ScopeEchoRead)},
	}}

	err := api.verify()
	if err == nil {
		t.Fatal("expected verification error")
	}
	for _, want := range []string{
		"public endpoint GET /a must state a reason",
		"token endpoint GET /b declares no scope",
		"duplicate endpoint GET /a",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("verification error missing %q:\n%v", want, err)
		}
	}
}

func TestVerifyRequiresEveryAdministrativeScope(t *testing.T) {
	api := &apiRegistry{}
	if err := api.verify(); err == nil || !strings.Contains(err.Error(), "administrative scope") {
		t.Fatalf("expected missing-scope error, got %v", err)
	}
}
