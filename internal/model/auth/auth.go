// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/golang-jwt/jwt/v5"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type MyClaims struct {
	Userid   string   `json:"user_id"`
	Username string   `json:"username"`
	Type     string   `json:"typ"`
	Scopes   []string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

const (
	MAX_USER_COUNT  = 5
	AnonymousUserID = ""
)

type (
	OAuth2Action string
	AuthType     string
)

const (
	OAuth2ActionLogin    OAuth2Action = "login"
	OAuth2ActionRegister OAuth2Action = "register"
	OAuth2ActionBind     OAuth2Action = "bind"

	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeOIDC   AuthType = "oidc"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
	ExpiresIn    int    `json:"expires_in"`
}

type ExchangeCodeReq struct {
	Code string `json:"code" binding:"required"`
}

type OAuthState struct {
	Action   string `json:"action"`
	UserID   string `json:"user_id,omitempty"`
	Nonce    string `json:"nonce"`
	Redirect string `json:"redirect,omitempty"`
	Exp      int64  `json:"exp"`
	Provider string `json:"provider,omitempty"`
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

type GoogleUser struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type QQTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid,omitempty"`
}

type QQOpenIDResponse struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
}

type QQUser struct {
	Nickname     string `json:"nickname"`
	FigureURL    string `json:"figureurl"`
	FigureURL1   string `json:"figureurl_1"`
	FigureURL2   string `json:"figureurl_2"`
	FigureURLQQ1 string `json:"figureurl_qq_1"`
	FigureURLQQ2 string `json:"figureurl_qq_2"`
	Gender       string `json:"gender"`
}

type Passkey struct {
	ID             string `gorm:"type:char(36);primaryKey"`
	UserID         string `gorm:"type:char(36);not null;index"`
	CredentialID   string `gorm:"size:255;not null;uniqueIndex:uid_cred"`
	CredentialJSON string `gorm:"type:text;not null"`
	PublicKey      string `gorm:"type:text"`
	SignCount      uint32 `gorm:"not null;default:0"`
	LastUsedAt     int64
	DeviceName     string `gorm:"size:128"`
	AAGUID         string `gorm:"size:36"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
	UpdatedAt      int64  `gorm:"autoUpdateTime"`
}

type PasskeyRegisterBeginReq struct {
	DeviceName string `json:"device_name"`
}

type PasskeyRegisterBeginResp struct {
	Nonce     string                                       `json:"nonce"`
	PublicKey *protocol.PublicKeyCredentialCreationOptions `json:"publicKey"`
}

type PasskeyFinishReq struct {
	Nonce      string          `json:"nonce"      binding:"required"`
	Credential json.RawMessage `json:"credential" binding:"required"`
}

type PasskeyLoginBeginResp struct {
	Nonce     string                                      `json:"nonce"`
	PublicKey *protocol.PublicKeyCredentialRequestOptions `json:"publicKey"`
}

type PasskeyDeviceDto struct {
	ID         string `json:"id"`
	DeviceName string `json:"device_name"`
	AAGUID     string `json:"aaguid"`
	LastUsedAt int64  `json:"last_used_at"`
	CreatedAt  int64  `json:"created_at"`
}

type PasskeyUpdateDeviceNameReq struct {
	DeviceName string `json:"device_name" binding:"required"`
}

func (p *Passkey) BeforeCreate(_ *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuidUtil.NewV7()
	}
	return nil
}
