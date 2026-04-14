package auth

import (
	"context"
	"encoding/json"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
)

type Service interface {
	Login(loginDto *authModel.LoginDto) (*authModel.TokenPair, error)
	BindOAuth(ctx context.Context, provider string, redirectURI string) (string, error)
	GetOAuthLoginURL(provider string, redirectURI string) (string, error)
	HandleOAuthCallback(provider string, code string, state string) (string, error)
	ExchangeOAuthCode(code string) (*authModel.TokenPair, error)
	GetOAuthInfo(ctx context.Context, provider string) (model.OAuthInfoDto, error)
	PasskeyRegisterBegin(ctx context.Context, rpID, origin, deviceName string) (authModel.PasskeyRegisterBeginResp, error)
	PasskeyRegisterFinish(ctx context.Context, rpID, origin, nonce string, credential json.RawMessage) error
	PasskeyLoginBegin(rpID, origin string) (authModel.PasskeyLoginBeginResp, error)
	PasskeyLoginFinish(rpID, origin, nonce string, credential json.RawMessage) (*authModel.TokenPair, error)
	ListPasskeys(ctx context.Context) ([]authModel.PasskeyDeviceDto, error)
	DeletePasskey(ctx context.Context, passkeyID string) error
	UpdatePasskeyDeviceName(ctx context.Context, passkeyID string, deviceName string) error
	TokenRevoker
}

type TokenRevoker interface {
	RevokeToken(jti string, remainTTL time.Duration)
	IsTokenRevoked(jti string) bool
}

type UserRepo interface {
	GetUserByID(ctx context.Context, id string) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
}

type IdentityRepo interface {
	BindOAuth(ctx context.Context, userID string, provider, oauthID, issuer, authType string) error
	GetUserByOAuthID(ctx context.Context, provider, oauthID string) (model.User, error)
	GetUserByOIDC(ctx context.Context, provider, oauthID, issuer string) (model.User, error)
	GetOAuthInfo(userId string, provider string) (model.UserExternalIdentity, error)
	GetOAuthOIDCInfo(userId string, provider string, issuer string) (model.UserExternalIdentity, error)
}

type PasskeyRepo interface {
	CreatePasskey(ctx context.Context, passkey *authModel.Passkey) error
	ListPasskeysByUserID(userID string) ([]authModel.Passkey, error)
	GetPasskeyByCredentialID(credentialID string) (authModel.Passkey, error)
	UpdatePasskeyUsage(ctx context.Context, passkeyID string, signCount uint32, lastUsedAt int64) error
	UpdatePasskeyDeviceName(ctx context.Context, userID, passkeyID string, deviceName string) error
	DeletePasskeyByID(ctx context.Context, userID, passkeyID string) error
}

type ChallengeStore interface {
	CacheSetPasskeySession(key string, val any, ttl time.Duration)
	CacheGetPasskeySession(key string) (any, error)
	CacheDeletePasskeySession(key string)
}

type Repository interface {
	UserRepo
	IdentityRepo
	PasskeyRepo
	ChallengeStore
}

type OAuthCodeStore interface {
	StoreOAuthCode(code string, pair *authModel.TokenPair, ttl time.Duration)
	GetAndDeleteOAuthCode(code string) (*authModel.TokenPair, error)
}

type AuthRepo interface {
	OAuthCodeStore
	TokenRevoker
}

type SettingService = settingService.Service
