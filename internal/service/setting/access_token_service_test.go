package service

import (
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

func TestCreateAccessToken_RejectsUnknownScope(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-scope",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{"admin:root"},
		Audience: authModel.AudiencePublic,
	}
	if err := validateAccessTokenRequest(user, dto); err == nil {
		t.Fatal("expected error for unknown scope")
	}
}

func TestCreateAccessToken_RejectsUnknownAudience(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-audience",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authModel.ScopeEchoRead},
		Audience: "unknown-client",
	}
	if err := validateAccessTokenRequest(user, dto); err == nil {
		t.Fatal("expected error for unknown audience")
	}
}

func TestCreateAccessToken_RejectsAdminScopeForNonAdminUser(t *testing.T) {
	user := userModel.User{IsAdmin: false}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-admin-scope",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authModel.ScopeAdminSettings},
		Audience: authModel.AudiencePublic,
	}
	if err := validateAccessTokenRequest(user, dto); err == nil {
		t.Fatal("expected error for admin scope on non-admin user")
	}
}
