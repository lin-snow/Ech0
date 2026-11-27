// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/authz"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
	"github.com/stretchr/testify/assert"
)

// echoAuthorizerRepo answers the one lookup the authorizer makes; every other
// method stays nil so an accidental extra call panics instead of silently
// passing.
type echoAuthorizerRepo struct {
	echoService.Repository
	echo *echoModel.Echo
	err  error
}

func (r echoAuthorizerRepo) GetEchosById(context.Context, string) (*echoModel.Echo, error) {
	return r.echo, r.err
}

func TestEchoFileAuthorizer(t *testing.T) {
	deletedAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	token := func(scopes ...string) context.Context {
		return helpers.CtxAsToken("u-1", "access", scopes, []string{"ech0"}, "jti-1")
	}

	cases := []struct {
		name  string
		echo  *echoModel.Echo
		repo  error
		ctx   context.Context
		allow bool
	}{
		{name: "public echo is readable by anyone", echo: &echoModel.Echo{ID: "e1"}, ctx: helpers.CtxAnonymous(), allow: true},
		{name: "missing echo denies", echo: nil, ctx: helpers.CtxAsUser("u-1"), allow: false},
		{name: "lookup failure denies instead of leaking", repo: errors.New("db down"), ctx: helpers.CtxAsUser("u-1"), allow: false},
		{name: "deleted echo denies", echo: &echoModel.Echo{ID: "e1", DeletedAt: &deletedAt}, ctx: helpers.CtxAsUser("u-1"), allow: false},
		{name: "private echo denies anonymous", echo: &echoModel.Echo{ID: "e1", Private: true}, ctx: helpers.CtxAnonymous(), allow: false},
		{name: "private echo allows a session", echo: &echoModel.Echo{ID: "e1", Private: true}, ctx: helpers.CtxAsUser("u-1", role.Member), allow: true},
		{name: "private echo denies a token without echo:private", echo: &echoModel.Echo{ID: "e1", Private: true}, ctx: token(authz.ScopeEchoRead), allow: false},
		{name: "private echo allows a token with echo:private", echo: &echoModel.Echo{ID: "e1", Private: true}, ctx: token(authz.ScopeEchoRead, authz.ScopeEchoPrivate), allow: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authorizer := echoService.NewEchoFileAuthorizer(echoAuthorizerRepo{echo: tc.echo, err: tc.repo})
			principal := viewer.MustFromContext(tc.ctx)
			allowed, err := authorizer.Authorize(tc.ctx, "e1", principal)
			assert.NoError(t, err)
			assert.Equal(t, tc.allow, allowed)
		})
	}
}
