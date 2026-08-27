// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

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

func (settingService *SettingService) ListAccessTokens(
	ctx context.Context,
) ([]model.AccessTokenSetting, error) {
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

	var validTokens []model.AccessTokenSetting
	currentTime := time.Now().UTC().Unix()

	for _, token := range tokens {
		if token.Expiry == nil || *token.Expiry > currentTime {
			validTokens = append(validTokens, token)
		} else {
			_ = settingService.transactor.Run(ctx, func(txCtx context.Context) error {
				return settingService.settingRepository.DeleteAccessTokenByID(txCtx, token.ID)
			})
		}
	}

	return validTokens, nil
}

func (settingService *SettingService) CreateAccessToken(
	ctx context.Context,
	newToken *model.AccessTokenSettingDto,
) (string, error) {
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
	jti := uuidUtil.NewV7()
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

	claims := jwtUtil.CreateAccessClaimsWithExpiry(user, int64(expiryDuration), scopes, audience, jti)
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	var expiryPtr *int64
	if expiry == model.NEVER_EXPIRY {
		expiryPtr = nil
	} else {
		t := time.Now().UTC().Add(expiryDuration).Unix()
		expiryPtr = &t
	}

	accessToken := &model.AccessTokenSetting{
		UserID:    user.ID,
		Token:     tokenString,
		Name:      name,
		TokenType: authModel.TokenTypeAccess,
		Scopes:    string(scopeJSON),
		Audience:  audience,
		JTI:       jti,
		Expiry:    expiryPtr,
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

func (settingService *SettingService) DeleteAccessToken(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if settingService.tokenRevoker != nil {
		token, getErr := settingService.settingRepository.GetAccessTokenByID(ctx, id)
		if getErr == nil && token.JTI != "" {
			ttl := remainingTTLForRevoke(token.Expiry)
			settingService.tokenRevoker.RevokeToken(token.JTI, ttl)
		}
	}

	return settingService.transactor.Run(ctx, func(txCtx context.Context) error {
		return settingService.settingRepository.DeleteAccessTokenByID(txCtx, id)
	})
}

func remainingTTLForRevoke(expiryUnix *int64) time.Duration {
	const neverFallback = 100 * 365 * 24 * time.Hour
	if expiryUnix == nil {
		return neverFallback
	}
	remaining := time.Until(time.Unix(*expiryUnix, 0).UTC())
	if remaining <= 0 {
		return time.Hour
	}
	return remaining
}
