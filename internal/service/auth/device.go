// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

const (
	deviceCodeTTL      = 10 * time.Minute
	devicePollInterval = 5 * time.Second
	// deviceUserCodeAlphabet omits 0/O and 1/I so codes survive being retyped.
	deviceUserCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var (
	ErrDeviceUnknown  = errors.New(commonModel.DEVICE_CODE_UNKNOWN)
	ErrDevicePending  = errors.New(commonModel.DEVICE_CODE_PENDING)
	ErrDeviceSlowDown = errors.New(commonModel.DEVICE_CODE_SLOW_DOWN)
	ErrDeviceExpired  = errors.New(commonModel.DEVICE_CODE_EXPIRED)
	ErrDeviceDenied   = errors.New(commonModel.DEVICE_CODE_DENIED)
)

type deviceStatus int

const (
	devicePending deviceStatus = iota
	deviceApproved
	deviceDenied
)

type deviceEntry struct {
	deviceCode string
	userCode   string
	clientName string
	scopes     []string
	status     deviceStatus
	token      string
	expiresAt  time.Time
	lastPoll   time.Time
}

// StartDeviceAuthorization opens a pending device authorization. The caller
// shows UserCode to the user; nothing is issued until it is approved.
func (authService *AuthService) StartDeviceAuthorization(
	clientName string,
	scopes []string,
) (authModel.DeviceCodeResponse, error) {
	name := strings.TrimSpace(clientName)
	if name == "" {
		name = "ech0-cli"
	}
	requested := normalizeDeviceScopes(scopes)
	for _, scope := range requested {
		if !authModel.IsValidScope(scope) {
			return authModel.DeviceCodeResponse{}, fmt.Errorf("invalid scope %q", scope)
		}
	}

	deviceCode, err := randomURLToken(32)
	if err != nil {
		return authModel.DeviceCodeResponse{}, err
	}
	userCode, err := randomUserCode()
	if err != nil {
		return authModel.DeviceCodeResponse{}, err
	}

	now := time.Now().UTC()
	entry := &deviceEntry{
		deviceCode: deviceCode,
		userCode:   userCode,
		clientName: name,
		scopes:     requested,
		status:     devicePending,
		expiresAt:  now.Add(deviceCodeTTL),
	}

	authService.deviceMu.Lock()
	authService.sweepExpiredLocked(now)
	authService.devices[deviceCode] = entry
	authService.deviceMu.Unlock()

	return authModel.DeviceCodeResponse{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ExpiresIn:  int(deviceCodeTTL.Seconds()),
		Interval:   int(devicePollInterval.Seconds()),
	}, nil
}

// DeviceAuthorizationInfo describes a pending code so the approval page can
// show what is being granted.
func (authService *AuthService) DeviceAuthorizationInfo(userCode string) (authModel.DevicePendingInfo, error) {
	normalized := normalizeUserCode(userCode)
	if normalized == "" {
		return authModel.DevicePendingInfo{}, ErrDeviceUnknown
	}

	authService.deviceMu.Lock()
	defer authService.deviceMu.Unlock()

	entry := authService.entryByUserCodeLocked(normalized)
	if entry == nil {
		return authModel.DevicePendingInfo{}, ErrDeviceUnknown
	}
	if time.Now().UTC().After(entry.expiresAt) || entry.status == deviceDenied {
		return authModel.DevicePendingInfo{}, ErrDeviceExpired
	}
	if entry.status == deviceApproved {
		return authModel.DevicePendingInfo{}, ErrDeviceUnknown
	}
	return authModel.DevicePendingInfo{
		ClientName: entry.clientName,
		Scopes:     append([]string(nil), entry.scopes...),
		ExpiresIn:  int(time.Until(entry.expiresAt).Seconds()),
	}, nil
}

// ApproveDeviceAuthorization mints the access token for a pending code. The
// token issuer enforces that the approver is an admin, which keeps device
// approval as restricted as minting a token in the panel.
func (authService *AuthService) ApproveDeviceAuthorization(ctx context.Context, userCode string) error {
	normalized := normalizeUserCode(userCode)
	if normalized == "" {
		return ErrDeviceUnknown
	}

	authService.deviceMu.Lock()
	entry := authService.entryByUserCodeLocked(normalized)
	if entry == nil {
		authService.deviceMu.Unlock()
		return ErrDeviceUnknown
	}
	if time.Now().UTC().After(entry.expiresAt) {
		authService.deviceMu.Unlock()
		return ErrDeviceExpired
	}
	if entry.status == deviceApproved {
		authService.deviceMu.Unlock()
		return nil
	}
	clientName := entry.clientName
	scopes := append([]string(nil), entry.scopes...)
	authService.deviceMu.Unlock()

	if authService.tokenIssuer == nil {
		return errors.New("token issuer unavailable")
	}
	token, err := authService.tokenIssuer.CreateAccessToken(ctx, &settingModel.AccessTokenSettingDto{
		Name:     "cli:" + clientName,
		Expiry:   settingModel.ONE_MONTH_EXPIRY,
		Scopes:   scopes,
		Audience: authModel.AudienceCLI,
	})
	if err != nil {
		return err
	}

	authService.deviceMu.Lock()
	if current := authService.entryByUserCodeLocked(normalized); current != nil && current.status == devicePending {
		current.status = deviceApproved
		current.token = token
	}
	authService.deviceMu.Unlock()
	return nil
}

// PollDeviceAuthorization is what the CLI calls until the user approves.
func (authService *AuthService) PollDeviceAuthorization(deviceCode string) (authModel.DeviceTokenResponse, error) {
	authService.deviceMu.Lock()
	defer authService.deviceMu.Unlock()

	entry, ok := authService.devices[deviceCode]
	if !ok {
		return authModel.DeviceTokenResponse{}, ErrDeviceUnknown
	}
	now := time.Now().UTC()
	if now.After(entry.expiresAt) {
		delete(authService.devices, deviceCode)
		return authModel.DeviceTokenResponse{}, ErrDeviceExpired
	}

	switch entry.status {
	case deviceApproved:
		delete(authService.devices, deviceCode)
		return authModel.DeviceTokenResponse{
			AccessToken: entry.token,
			TokenType:   "Bearer",
			ExpiresIn:   int((30 * 24 * time.Hour).Seconds()),
			Scopes:      append([]string(nil), entry.scopes...),
		}, nil
	case deviceDenied:
		delete(authService.devices, deviceCode)
		return authModel.DeviceTokenResponse{}, ErrDeviceDenied
	}

	if !entry.lastPoll.IsZero() && now.Sub(entry.lastPoll) < devicePollInterval {
		entry.lastPoll = now
		return authModel.DeviceTokenResponse{}, ErrDeviceSlowDown
	}
	entry.lastPoll = now
	return authModel.DeviceTokenResponse{}, ErrDevicePending
}

func (authService *AuthService) entryByUserCodeLocked(userCode string) *deviceEntry {
	for _, entry := range authService.devices {
		if entry.userCode == userCode {
			return entry
		}
	}
	return nil
}

func (authService *AuthService) sweepExpiredLocked(now time.Time) {
	for deviceCode, entry := range authService.devices {
		if now.After(entry.expiresAt) {
			delete(authService.devices, deviceCode)
		}
	}
}

func normalizeDeviceScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return authModel.DefaultDeviceScopes()
	}
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		trimmed := strings.TrimSpace(scope)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return authModel.DefaultDeviceScopes()
	}
	return out
}

func randomURLToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func randomUserCode() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	chars := make([]byte, len(buf))
	for index, value := range buf {
		chars[index] = deviceUserCodeAlphabet[int(value)%len(deviceUserCodeAlphabet)]
	}
	return string(chars[:4]) + "-" + string(chars[4:]), nil
}

// normalizeUserCode accepts "abcd efgh" / "ABCD-EFGH" and returns "ABCD-EFGH".
func normalizeUserCode(raw string) string {
	var letters []byte
	for _, char := range strings.ToUpper(raw) {
		if char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' {
			letters = append(letters, byte(char))
		}
	}
	if len(letters) != 8 {
		return ""
	}
	return string(letters[:4]) + "-" + string(letters[4:])
}
