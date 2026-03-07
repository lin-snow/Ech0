package repository

import (
	"context"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
)

type UserRepositoryInterface interface {
	// GetUserByID 根据用户ID获取用户
	GetUserByID(ctx context.Context, id int) (model.User, error)

	// GetUserByUsername 根据用户名获取用户
	GetUserByUsername(ctx context.Context, username string) (model.User, error)

	// GetAllUsers 获取所有用户
	GetAllUsers(ctx context.Context) ([]model.User, error)

	// CreateUser 创建一个新的用户
	CreateUser(ctx context.Context, newUser *model.User) error

	// GetSysAdmin 获取系统管理员
	GetSysAdmin(ctx context.Context) (model.User, error)

	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, user *model.User) error

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, id uint) error

	// BindOAuth 绑定 OAuth 账号
	BindOAuth(ctx context.Context, userID uint, provider, oauthID, issuer, authType string) error

	// GetUserByOAuthID 根据 OAuth 提供商和 OAuth ID 获取用户
	GetUserByOAuthID(ctx context.Context, provider, oauthID string) (model.User, error)

	// GetUserByOIDC 根据 OIDC 提供商、issuer 与 sub 获取用户
	GetUserByOIDC(ctx context.Context, provider, oauthID, issuer string) (model.User, error)

	// GetOAuthInfo 获取 OAuth2 信息
	GetOAuthInfo(userId uint, provider string) (model.OAuthBinding, error)

	// GetOAuthOIDCInfo 获取 OIDC 信息
	GetOAuthOIDCInfo(userId uint, provider string, issuer string) (model.OAuthBinding, error)

	// Passkey / WebAuthn
	CreatePasskey(ctx context.Context, passkey *authModel.Passkey) error
	ListPasskeysByUserID(userID uint) ([]authModel.Passkey, error)
	GetPasskeyByCredentialID(credentialID string) (authModel.Passkey, error)
	UpdatePasskeyUsage(
		ctx context.Context,
		passkeyID uint,
		signCount uint32,
		lastUsedAt time.Time,
	) error
	UpdatePasskeyDeviceName(ctx context.Context, userID, passkeyID uint, deviceName string) error
	DeletePasskeyByID(ctx context.Context, userID, passkeyID uint) error

	// Passkey 会话缓存（challenge/session）
	CacheSetPasskeySession(key string, val any, ttl time.Duration)
	CacheGetPasskeySession(key string) (any, error)
	CacheDeletePasskeySession(key string)
}
