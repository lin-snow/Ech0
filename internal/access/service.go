// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package access

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/authz"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// Service mints access tokens for both session-bound callers (the settings
// panel and device approval) and explicit principals (OAuth code exchange).
type Service struct {
	transactor transaction.Transactor
	tokens     TokenRepository
	users      UserReader
}

func NewService(
	transactor transaction.Transactor,
	tokens TokenRepository,
	users UserReader,
) *Service {
	return &Service{
		transactor: transactor,
		tokens:     tokens,
		users:      users,
	}
}

// IssueForViewer mints a token for the principal in the request context. It is
// the entry point used by the settings panel and device approval.
func (s *Service) IssueForViewer(
	ctx context.Context,
	dto *model.AccessTokenSettingDto,
) (string, error) {
	v := viewer.MustFromContext(ctx)
	if v.TokenType() == authModel.TokenTypeAccess {
		if err := ensureDelegatedScopes(v.Scopes(), dto); err != nil {
			return "", err
		}
	}
	return s.IssueForUser(ctx, v.UserID(), dto)
}

// ensureDelegatedScopes prevents an access token from minting a token wider
// than itself. Token creation is currently session-only, but keeping the
// invariant here makes the service safe if that gate ever moves.
func ensureDelegatedScopes(held []string, dto *model.AccessTokenSettingDto) error {
	if dto == nil {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if !authz.Allows(held, dto.Scopes) {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return nil
}

// IssueForUser mints a token for an explicit principal. The OAuth token
// endpoint has no session, so it cannot rely on the viewer.
func (s *Service) IssueForUser(
	ctx context.Context,
	userID string,
	dto *model.AccessTokenSettingDto,
) (string, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if err := validateRequest(user, dto); err != nil {
		return "", err
	}
	return s.issue(ctx, user, dto)
}

func (s *Service) issue(
	ctx context.Context,
	user userModel.User,
	dto *model.AccessTokenSettingDto,
) (string, error) {
	scopes := normalizeScopes(dto.Scopes)
	scopeJSON, err := json.Marshal(scopes)
	if err != nil {
		return "", err
	}
	jti := uuidUtil.NewV7()
	// CreateAccessClaimsWithExpiry takes seconds; passing a time.Duration here
	// overflows the JWT exp claim and issues an already-expired token.
	expirySeconds := expirySeconds(dto.Expiry)

	claims := jwtUtil.CreateAccessClaimsWithExpiry(
		user, expirySeconds, scopes, dto.Audience, jti)
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	var expiryPtr *int64
	if dto.Expiry != model.NEVER_EXPIRY {
		expiresAt := time.Now().UTC().Add(time.Duration(expirySeconds) * time.Second).Unix()
		expiryPtr = &expiresAt
	}

	accessToken := &model.AccessTokenSetting{
		UserID:    user.ID,
		Token:     tokenString,
		Name:      dto.Name,
		TokenType: authModel.TokenTypeAccess,
		Scopes:    string(scopeJSON),
		Audience:  dto.Audience,
		JTI:       jti,
		Expiry:    expiryPtr,
	}

	if err := s.transactor.Run(ctx, func(txCtx context.Context) error {
		return s.tokens.CreateAccessToken(txCtx, accessToken)
	}); err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateRequest(user userModel.User, dto *model.AccessTokenSettingDto) error {
	if dto == nil {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if strings.TrimSpace(dto.Name) == "" {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if !authz.IsValidAudience(dto.Audience) {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	if len(dto.Scopes) == 0 {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	for _, scope := range dto.Scopes {
		if !authz.IsValidScope(scope) {
			return errors.New(commonModel.INVALID_PARAMS_BODY)
		}
	}
	if authz.HasAdminScope(dto.Scopes) && !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return nil
}

// expirySeconds maps a stored expiry preset to the JWT lifetime in seconds.
// Unknown presets fall back to the eight-hour default.
func expirySeconds(expiry string) int64 {
	switch expiry {
	case model.ONE_MONTH_EXPIRY:
		return int64(30 * 24 * time.Hour / time.Second)
	case model.NEVER_EXPIRY:
		return 0
	default:
		return int64(8 * time.Hour / time.Second)
	}
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
