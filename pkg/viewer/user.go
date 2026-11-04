// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

type UserViewer struct {
	userID    string
	tokenType string
	scopes    []string
	audience  []string
	tokenID   string
	isAdmin   bool
}

func NewUserViewer(userID string) *UserViewer {
	return &UserViewer{userID: userID}
}

func NewUserViewerWithToken(
	userID string,
	tokenType string,
	scopes []string,
	audience []string,
	tokenID string,
) *UserViewer {
	return &UserViewer{
		userID:    userID,
		tokenType: tokenType,
		scopes:    append([]string(nil), scopes...),
		audience:  append([]string(nil), audience...),
		tokenID:   tokenID,
	}
}

func (v *UserViewer) UserID() string    { return v.userID }
func (v *UserViewer) TokenType() string { return v.tokenType }
func (v *UserViewer) Scopes() []string  { return append([]string(nil), v.scopes...) }
func (v *UserViewer) Audience() []string {
	return append([]string(nil), v.audience...)
}
func (v *UserViewer) TokenID() string { return v.tokenID }
func (v *UserViewer) IsAdmin() bool   { return v.isAdmin }

// WithAdmin marks the viewer as an administrator. Session tokens carry the
// role claim; access-token authorization remains scope-driven.
func (v *UserViewer) WithAdmin(isAdmin bool) *UserViewer {
	v.isAdmin = isAdmin
	return v
}

var _ Context = (*UserViewer)(nil)
