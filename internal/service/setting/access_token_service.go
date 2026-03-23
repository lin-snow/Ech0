package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// ListAccessTokens 列出访问令牌
func (settingService *SettingService) ListAccessTokens(
	ctx context.Context,
) ([]model.AccessTokenSetting, error) {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	tokens, err := settingService.settingRepository.ListAccessTokens(ctx, user.ID)
	if err != nil {
		return []model.AccessTokenSetting{}, nil
	}

	// 处理tokens,过滤并删除过期的token
	var validTokens []model.AccessTokenSetting
	currentTime := time.Now().UTC()

	for _, token := range tokens {
		if token.Expiry == nil || token.Expiry.After(currentTime) {
			// nil 表示永不过期，或者还没过期
			validTokens = append(validTokens, token)
		} else {
			// 删除过期 token
			_ = settingService.transactor.Run(ctx, func(txCtx context.Context) error {
				return settingService.settingRepository.DeleteAccessTokenByID(txCtx, token.ID)
			})
		}
	}

	return validTokens, nil
}

// CreateAccessToken 创建访问令牌
func (settingService *SettingService) CreateAccessToken(
	ctx context.Context,
	newToken *model.AccessTokenSettingDto,
) (string, error) {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if err := validateAccessTokenRequest(user, newToken); err != nil {
		return "", err
	}

	name := newToken.Name
	expiry := newToken.Expiry
	audience := newToken.Audience
	scopes := normalizeScopes(newToken.Scopes)
	scopeJSON, err := json.Marshal(scopes)
	if err != nil {
		return "", err
	}
	jti := uuidUtil.MustNewV7()
	var expiryDuration time.Duration

	switch expiry {
	case model.EIGHT_HOUR_EXPIRY:
		expiryDuration = 8 * time.Hour
	case model.ONE_MONTH_EXPIRY:
		expiryDuration = 30 * 24 * time.Hour
	case model.NEVER_EXPIRY:
		expiryDuration = 0
	default:
		expiryDuration = 8 * time.Hour
	}

	// 生成jwt令牌
	claims := jwtUtil.CreateAccessClaimsWithExpiry(user, int64(expiryDuration), scopes, audience, jti)
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	// 处理数据库存储的 expiry
	var expiryPtr *time.Time
	if expiry == model.NEVER_EXPIRY {
		expiryPtr = nil // 永不过期，用 NULL
	} else {
		t := time.Now().UTC().Add(expiryDuration)
		expiryPtr = &t
	}

	// 保存到数据库
	accessToken := &model.AccessTokenSetting{
		UserID:    user.ID,
		Token:     tokenString,
		Name:      name,
		TokenType: authModel.TokenTypeAccess,
		Scopes:    string(scopeJSON),
		Audience:  audience,
		JTI:       jti,
		Expiry:    expiryPtr,
		CreatedAt: time.Now().UTC(),
	}

	if err := settingService.transactor.Run(ctx, func(txCtx context.Context) error {
		return settingService.settingRepository.CreateAccessToken(txCtx, accessToken)
	}); err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateAccessTokenRequest(user userModel.User, dto *model.AccessTokenSettingDto) error {
	if dto == nil {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if strings.TrimSpace(dto.Name) == "" {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if !authModel.IsValidAudience(dto.Audience) {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if len(dto.Scopes) == 0 {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	for _, scope := range dto.Scopes {
		if !authModel.IsValidScope(scope) {
			return errors.New(commonModel.INVALID_PARAMS_BODY)
		}
	}
	if authModel.HasAdminScope(dto.Scopes) && !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return nil
}

func normalizeScopes(scopes []string) []string {
	seen := make(map[string]struct{}, len(scopes))
	result := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		result = append(result, scope)
	}
	return result
}

// DeleteAccessToken 删除访问令牌
func (settingService *SettingService) DeleteAccessToken(ctx context.Context, id string) error {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(ctx, func(txCtx context.Context) error {
		return settingService.settingRepository.DeleteAccessTokenByID(txCtx, id)
	})
}
