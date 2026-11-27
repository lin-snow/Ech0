// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"errors"
	"fmt"

	"github.com/lin-snow/ech0/internal/apperr"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// Resource kinds. The kind is part of the decision, so a file can never
// satisfy a rule written for an echo.
const (
	KindInstance = "instance"
	KindEcho     = "echo"
	KindFile     = "file"
	KindUser     = "user"
)

// Resource is what the action targets. Author is the principal that controls
// the resource; an empty Author means the resource is instance-owned and shared
// by every signed-in member. Role and Assign describe a user target: its
// current role, and the role a re-role operation would grant (empty when the
// operation is not a re-role).
//
// Construct resources only through the helpers below. The helpers keep field
// semantics in one place, where a model can drift: fileModel.File.OwnerID is a
// storage-layout slot (for content files it holds the echo ID), while the
// uploader — the principal that authorizes — is fileModel.File.UploaderID.
type Resource struct {
	Kind   string
	ID     string
	Author string
	Role   role.Role
	Assign role.Role
}

// Instance targets the deployment itself: settings, tokens, migrations.
func Instance() Resource {
	return Resource{Kind: KindInstance}
}

// ForEcho targets an echo; its author owns it.
func ForEcho(e echoModel.Echo) Resource {
	return Resource{Kind: KindEcho, ID: e.ID, Author: e.UserID}
}

// ForFile targets a file; its uploader owns it.
func ForFile(f fileModel.File) Resource {
	return Resource{Kind: KindFile, ID: f.ID, Author: f.UploaderID}
}

// ForUser targets a user account. assign is the role the caller wants to
// grant; leave it empty for operations that do not change the role.
func ForUser(u userModel.User, assign role.Role) Resource {
	return Resource{Kind: KindUser, ID: u.ID, Role: u.Role, Assign: assign}
}

// ErrDenied is the single denial value: every transport renders it as 403 with
// the same message, so no second lookup table is needed.
var ErrDenied = apperr.ErrDenied

// ErrNoPolicy reports that an action reached the decision without a rule. It is
// a programming error, not a denial: fail closed, and let the test suite catch
// the missing classification.
var ErrNoPolicy = errors.New("authz: action is not classified")

// Decide answers whether p may perform act on res. It is pure: no IO, no cache,
// no policy store. The whole rule set is the switch below, so it can be read in
// one screen, tested exhaustively, and audited by search.
func Decide(p viewer.Context, act Action, res Resource) error {
	if p == nil || p.UserID() == "" {
		return deny(act, res)
	}

	switch act {
	case EchoWrite, EchoDelete, FileRead, FileDelete:
		if err := decideOwned(p, act, res); err != nil {
			return err
		}

	case FileManage, UserList, SettingWrite, TokenManage, DataManage,
		ConnectManage, CopilotUse, OAuthManage, CommentManage, TagManage:
		if !p.Role().AtLeast(role.Admin) {
			return deny(act, res)
		}

	case UserManage:
		if err := decideUserManage(p, act, res); err != nil {
			return err
		}

	case UserTransfer:
		if p.Role() != role.Owner {
			return deny(act, res)
		}

	default:
		return fmt.Errorf("%w: %q", ErrNoPolicy, act)
	}

	return nil
}

// decideOwned allows the resource owner, an administrator, or any signed-in
// member when the resource is instance-owned (empty Author).
func decideOwned(p viewer.Context, act Action, res Resource) error {
	switch res.Kind {
	case KindEcho, KindFile:
	default:
		return fmt.Errorf("%w: %q has no owner rule for %q", ErrNoPolicy, act, res.Kind)
	}

	if res.Author == "" || res.Author == p.UserID() || p.Role().AtLeast(role.Admin) {
		return nil
	}

	return deny(act, res)
}

// decideUserManage bounds account administration by rank: the actor must be
// an administrator, strictly outrank the target, and may only grant a role
// below its own. Owners therefore manage admins and members; admins manage
// members only; nobody manages their own rank.
func decideUserManage(p viewer.Context, act Action, res Resource) error {
	if !p.Role().AtLeast(role.Admin) {
		return deny(act, res)
	}

	if res.Kind != KindUser || res.ID == "" || res.Role == "" {
		return fmt.Errorf("%w: %q needs a user target with a role", ErrNoPolicy, act)
	}

	if p.Role().Rank() <= res.Role.Rank() {
		return deny(act, res)
	}

	if res.Assign != "" && p.Role().Rank() <= res.Assign.Rank() {
		return deny(act, res)
	}

	return nil
}

// deny reports the decision with the action that failed, so logs name the
// missing permission instead of a bare "no permission".
func deny(act Action, res Resource) error {
	return fmt.Errorf("%w: %s on %s %q", ErrDenied, act, res.Kind, res.ID)
}
