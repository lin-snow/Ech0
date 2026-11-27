// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"errors"
	"testing"

	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const (
	ownerID  = "u-owner"
	adminID  = "u-admin"
	memberID = "u-member"
)

var (
	owner  = viewer.NewSession(ownerID, role.Owner, nil)
	admin  = viewer.NewSession(adminID, role.Admin, nil)
	member = viewer.NewSession(memberID, role.Member, nil)
	anon   = viewer.Anonymous
)

func resourceFor(act Action) Resource {
	switch act {
	case EchoWrite, EchoDelete:
		return Resource{Kind: KindEcho, ID: "e-1", Author: memberID}
	case FileRead, FileDelete:
		return Resource{Kind: KindFile, ID: "f-1", Author: memberID}
	case UserManage:
		return forUserTarget(role.Member, "")
	default:
		return Instance()
	}
}

func forUserTarget(target, assign role.Role) Resource {
	return Resource{Kind: KindUser, ID: "u-target", Role: target, Assign: assign}
}

// TestEveryActionClassified is the startup exhaustiveness check: an action with
// no rule fails closed with ErrNoPolicy, and this test makes that unreachable.
func TestEveryActionClassified(t *testing.T) {
	seen := make(map[Action]bool, len(AllActions()))

	for _, act := range AllActions() {
		if seen[act] {
			t.Fatalf("action %q listed twice", act)
		}
		seen[act] = true

		err := Decide(owner, act, resourceFor(act))
		if errors.Is(err, ErrNoPolicy) {
			t.Errorf("action %q reaches the decision with no rule: %v", act, err)
		}
	}
}

func TestAnonymousDeniedEverywhere(t *testing.T) {
	for _, act := range AllActions() {
		if err := Decide(anon, act, resourceFor(act)); !errors.Is(err, ErrDenied) {
			t.Errorf("anonymous %q: want ErrDenied, got %v", act, err)
		}
	}
}

func TestOwnedResources(t *testing.T) {
	tests := []struct {
		name string
		act  Action
		p    viewer.Context
		res  Resource
		deny bool
	}{
		{"author edits own echo", EchoWrite, member, Resource{Kind: KindEcho, ID: "e1", Author: memberID}, false},
		{"member edits foreign echo", EchoWrite, member, Resource{Kind: KindEcho, ID: "e1", Author: "u-other"}, true},
		{"admin edits foreign echo", EchoWrite, admin, Resource{Kind: KindEcho, ID: "e1", Author: "u-other"}, false},
		{"author deletes own file", FileDelete, member, Resource{Kind: KindFile, ID: "f1", Author: memberID}, false},
		{"member deletes foreign file", FileDelete, member, Resource{Kind: KindFile, ID: "f1", Author: "u-other"}, true},
		{"member reads instance file", FileRead, member, Resource{Kind: KindFile, ID: "f1"}, false},
		{"admin reads private file", FileRead, admin, Resource{Kind: KindFile, ID: "f1", Author: "u-other"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertDenial(t, Decide(tc.p, tc.act, tc.res), tc.deny)
		})
	}
}

func TestAdminSurfaces(t *testing.T) {
	for _, act := range []Action{
		FileManage, UserList, SettingWrite, TokenManage, DataManage,
		ConnectManage, CopilotUse, OAuthManage, CommentManage, TagManage,
	} {
		t.Run(string(act), func(t *testing.T) {
			assertDenial(t, Decide(member, act, Instance()), true)
			assertDenial(t, Decide(admin, act, Instance()), false)
			assertDenial(t, Decide(owner, act, Instance()), false)
		})
	}
}

func TestUserManage(t *testing.T) {
	tests := []struct {
		name string
		p    viewer.Context
		res  Resource
		deny bool
	}{
		{"admin disables member", admin, forUserTarget(role.Member, ""), false},
		{"admin disables admin", admin, forUserTarget(role.Admin, ""), true},
		{"admin disables owner", admin, forUserTarget(role.Owner, ""), true},
		{"admin promotes member to admin", admin, forUserTarget(role.Member, role.Admin), true},
		{"admin demotes member", admin, forUserTarget(role.Member, role.Member), false},
		{"owner promotes member to admin", owner, forUserTarget(role.Member, role.Admin), false},
		{"owner demotes admin", owner, forUserTarget(role.Admin, role.Member), false},
		{"owner touches owner", owner, forUserTarget(role.Owner, role.Member), true},
		{"member touches member", member, forUserTarget(role.Member, ""), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertDenial(t, Decide(tc.p, UserManage, tc.res), tc.deny)
		})
	}
}

func TestUserTransfer(t *testing.T) {
	res := Resource{Kind: KindUser, ID: "u-next"}
	assertDenial(t, Decide(owner, UserTransfer, res), false)
	assertDenial(t, Decide(admin, UserTransfer, res), true)
	assertDenial(t, Decide(member, UserTransfer, res), true)
}

func TestUnclassifiedDecisions(t *testing.T) {
	tests := []struct {
		name string
		act  Action
		res  Resource
	}{
		{"unknown action", Action("nope.nope"), Instance()},
		{"owned action on instance resource", EchoDelete, Instance()},
		{"user manage without role", UserManage, Resource{Kind: KindUser, ID: "u1"}},
		{"user manage without target", UserManage, Resource{Kind: KindUser, Role: role.Member}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Decide(owner, tc.act, tc.res)
			if !errors.Is(err, ErrNoPolicy) {
				t.Fatalf("want ErrNoPolicy, got %v", err)
			}
		})
	}
}

// TestDenialIsOneSentinel pins the denial contract: every transport classifies
// the same value, so there is no second denial to keep in step.
func TestDenialIsOneSentinel(t *testing.T) {
	err := Decide(member, SettingWrite, Instance())
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("denial = %T %v", err, err)
	}
}

func assertDenial(t *testing.T, err error, wantDenied bool) {
	t.Helper()

	if wantDenied {
		if !errors.Is(err, ErrDenied) {
			t.Fatalf("want ErrDenied, got %v", err)
		}
		return
	}

	if err != nil {
		t.Fatalf("want allow, got %v", err)
	}
}
