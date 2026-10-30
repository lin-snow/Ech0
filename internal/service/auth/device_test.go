// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type fakeTokenIssuer struct {
	token string
	err   error
	dto   *settingModel.AccessTokenSettingDto
}

func (issuer *fakeTokenIssuer) CreateAccessToken(
	_ context.Context,
	dto *settingModel.AccessTokenSettingDto,
) (string, error) {
	issuer.dto = dto
	return issuer.token, issuer.err
}

func newDeviceService() (*AuthService, *fakeTokenIssuer) {
	issuer := &fakeTokenIssuer{token: "device-access-token"}
	return NewAuthService(nil, nil, nil, nil, issuer), issuer
}

func TestStartDeviceAuthorization(t *testing.T) {
	service, _ := newDeviceService()

	code, err := service.StartDeviceAuthorization("ech0-cli", nil)
	if err != nil {
		t.Fatalf("StartDeviceAuthorization: %v", err)
	}
	if !regexp.MustCompile(`^[A-Z2-9]{4}-[A-Z2-9]{4}$`).MatchString(code.UserCode) {
		t.Fatalf("unexpected user code %q", code.UserCode)
	}
	if code.DeviceCode == "" || code.ExpiresIn != int(deviceCodeTTL.Seconds()) || code.Interval != int(devicePollInterval.Seconds()) {
		t.Fatalf("unexpected response %+v", code)
	}
}

func TestStartDeviceAuthorizationRejectsUnknownScope(t *testing.T) {
	service, _ := newDeviceService()
	if _, err := service.StartDeviceAuthorization("cli", []string{"admin:root"}); err == nil {
		t.Fatal("expected an error for an unknown scope")
	}
}

func TestDeviceAuthorizationInfoNormalizesUserCode(t *testing.T) {
	service, _ := newDeviceService()
	code, err := service.StartDeviceAuthorization("ech0-cli", []string{authModel.ScopeEchoRead})
	if err != nil {
		t.Fatalf("StartDeviceAuthorization: %v", err)
	}

	info, err := service.DeviceAuthorizationInfo(" " + regexp.MustCompile(`-`).ReplaceAllString(code.UserCode, "") + " ")
	if err != nil {
		t.Fatalf("DeviceAuthorizationInfo: %v", err)
	}
	if info.ClientName != "ech0-cli" || len(info.Scopes) != 1 || info.Scopes[0] != authModel.ScopeEchoRead {
		t.Fatalf("unexpected info %+v", info)
	}
}

func TestPollDeviceAuthorizationPendingSlowDownAndApproval(t *testing.T) {
	service, issuer := newDeviceService()
	code, err := service.StartDeviceAuthorization("ech0-cli", []string{authModel.ScopeEchoRead, authModel.ScopeEchoWrite})
	if err != nil {
		t.Fatalf("StartDeviceAuthorization: %v", err)
	}

	if _, err := service.PollDeviceAuthorization(code.DeviceCode); !errors.Is(err, ErrDevicePending) {
		t.Fatalf("first poll = %v, want pending", err)
	}
	if _, err := service.PollDeviceAuthorization(code.DeviceCode); !errors.Is(err, ErrDeviceSlowDown) {
		t.Fatalf("second poll = %v, want slow_down", err)
	}

	if err := service.ApproveDeviceAuthorization(context.Background(), code.UserCode); err != nil {
		t.Fatalf("ApproveDeviceAuthorization: %v", err)
	}
	if issuer.dto == nil {
		t.Fatal("issuer was not called")
	}
	if issuer.dto.Audience != authModel.AudienceCLI || issuer.dto.Expiry != settingModel.ONE_MONTH_EXPIRY {
		t.Fatalf("unexpected token request %+v", issuer.dto)
	}
	if len(issuer.dto.Scopes) != 2 {
		t.Fatalf("requested scopes = %v", issuer.dto.Scopes)
	}

	token, err := service.PollDeviceAuthorization(code.DeviceCode)
	if err != nil {
		t.Fatalf("poll after approval: %v", err)
	}
	if token.AccessToken != "device-access-token" || token.TokenType != "Bearer" || len(token.Scopes) != 2 {
		t.Fatalf("unexpected token response %+v", token)
	}

	// The pairing is single use.
	if _, err := service.PollDeviceAuthorization(code.DeviceCode); !errors.Is(err, ErrDeviceUnknown) {
		t.Fatalf("reuse poll = %v, want unknown", err)
	}
}

func TestApproveDeviceAuthorizationPropagatesIssuerError(t *testing.T) {
	service, issuer := newDeviceService()
	issuer.err = errors.New("no permission")
	code, err := service.StartDeviceAuthorization("ech0-cli", nil)
	if err != nil {
		t.Fatalf("StartDeviceAuthorization: %v", err)
	}

	if err := service.ApproveDeviceAuthorization(context.Background(), code.UserCode); err == nil || err.Error() != "no permission" {
		t.Fatalf("approve error = %v", err)
	}
}

func TestPollDeviceAuthorizationExpired(t *testing.T) {
	service, _ := newDeviceService()
	code, err := service.StartDeviceAuthorization("ech0-cli", nil)
	if err != nil {
		t.Fatalf("StartDeviceAuthorization: %v", err)
	}

	service.deviceMu.Lock()
	service.devices[code.DeviceCode].expiresAt = time.Now().UTC().Add(-time.Second)
	service.deviceMu.Unlock()

	if _, err := service.PollDeviceAuthorization(code.DeviceCode); !errors.Is(err, ErrDeviceExpired) {
		t.Fatalf("poll = %v, want expired", err)
	}
}

func TestPollDeviceAuthorizationUnknown(t *testing.T) {
	service, _ := newDeviceService()
	if _, err := service.PollDeviceAuthorization("nope"); !errors.Is(err, ErrDeviceUnknown) {
		t.Fatalf("poll = %v, want unknown", err)
	}
}
