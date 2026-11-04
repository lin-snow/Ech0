// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

type NoopViewer struct{}

func NewNoopViewer() *NoopViewer { return &NoopViewer{} }
func (v *NoopViewer) UserID() string {
	return ""
}
func (v *NoopViewer) TokenType() string { return "" }
func (v *NoopViewer) Scopes() []string  { return nil }
func (v *NoopViewer) Audience() []string {
	return nil
}
func (v *NoopViewer) TokenID() string { return "" }
func (v *NoopViewer) IsAdmin() bool   { return false }

var _ Context = (*NoopViewer)(nil)
