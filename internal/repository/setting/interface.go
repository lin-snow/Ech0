package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

type SettingRepositoryInterface interface {
	// ListAccessTokens 列出访问令牌
	ListAccessTokens(ctx context.Context, userID uint) ([]model.AccessTokenSetting, error)

	// CreateAccessToken 创建访问令牌
	CreateAccessToken(ctx context.Context, token *model.AccessTokenSetting) error

	// DeleteAccessTokenByID 删除访问令牌
	DeleteAccessTokenByID(ctx context.Context, id uint) error
}
