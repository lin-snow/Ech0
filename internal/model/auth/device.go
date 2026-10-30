// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// DeviceCodeRequest starts a device authorization (RFC 8628 style).
type DeviceCodeRequest struct {
	ClientName string   `json:"client_name"`
	Scopes     []string `json:"scopes"`
}

// DeviceCodeResponse is what the CLI prints before polling.
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceTokenRequest polls a started device authorization.
type DeviceTokenRequest struct {
	DeviceCode string `json:"device_code" binding:"required"`
}

// DeviceTokenResponse is the issued access token.
type DeviceTokenResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	Scopes      []string `json:"scopes"`
}

// DeviceApproveRequest approves a pending code in the browser.
type DeviceApproveRequest struct {
	UserCode string `json:"user_code" binding:"required"`
}

// DevicePendingInfo lets the approval page show what is being authorized.
type DevicePendingInfo struct {
	ClientName string   `json:"client_name"`
	Scopes     []string `json:"scopes"`
	ExpiresIn  int      `json:"expires_in"`
}

// DefaultDeviceScopes are requested when the client sends none.
func DefaultDeviceScopes() []string {
	return []string{ScopeEchoRead, ScopeEchoWrite, ScopeProfileRead}
}
