// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

func NewUser(opts ...func(*userModel.User)) userModel.User {
	u := userModel.User{
		ID:       "user-test-0001",
		Username: "tester",
		Role:     userModel.RoleMember,
	}
	for _, o := range opts {
		o(&u)
	}
	return u
}

// AsAdmin marks an account that passes administrator gates. It is not the
// space owner: ownership is a single row and only AsOwner models it.
func AsAdmin(u *userModel.User) { u.Role = userModel.RoleAdmin }

// AsOwner marks the single account that owns the space.
func AsOwner(u *userModel.User) { u.Role = userModel.RoleOwner }

func NewEcho(opts ...func(*echoModel.Echo)) echoModel.Echo {
	e := echoModel.Echo{
		ID:      "echo-test-0001",
		Content: "hello world",
		UserID:  "user-test-0001",
	}
	for _, o := range opts {
		o(&e)
	}
	return e
}

func AsPrivate(e *echoModel.Echo) { e.Private = true }
