package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

const (
	blacklistPrefix = "token_blacklist:"
	oauthCodePrefix = "oauth_code:"
)

type AuthRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

var (
	_ authService.Repository = (*AuthRepository)(nil)
	_ authService.AuthRepo   = (*AuthRepository)(nil)
)

func NewAuthRepository(dbProvider func() *gorm.DB, c cache.ICache[string, any]) *AuthRepository {
	return &AuthRepository{
		db:    dbProvider,
		cache: c,
	}
}

func (r *AuthRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *AuthRepository) RevokeToken(jti string, remainTTL time.Duration) {
	if jti == "" || remainTTL <= 0 {
		return
	}
	r.cache.SetWithTTL(fmt.Sprintf("%s%s", blacklistPrefix, jti), true, 1, remainTTL)
}

func (r *AuthRepository) IsTokenRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	_, found, _ := r.cache.Get(fmt.Sprintf("%s%s", blacklistPrefix, jti))
	return found
}

func (r *AuthRepository) StoreOAuthCode(code string, pair *authModel.TokenPair, ttl time.Duration) {
	if code == "" || pair == nil || ttl <= 0 {
		return
	}
	r.cache.SetWithTTL(oauthCodePrefix+code, pair, 1, ttl)
}

func (r *AuthRepository) GetAndDeleteOAuthCode(code string) (*authModel.TokenPair, error) {
	if code == "" {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}

	key := oauthCodePrefix + code
	val, found, _ := r.cache.Get(key)
	if !found {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}

	r.cache.Delete(key)

	pair, ok := val.(*authModel.TokenPair)
	if !ok {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}

	return pair, nil
}

func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	user := model.User{}
	if err := r.getDB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id string) (model.User, error) {
	user := model.User{}
	if err := r.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *AuthRepository) BindOAuth(
	ctx context.Context,
	userID string,
	provider, oauthID, issuer, authType string,
) error {
	protocol := string(authModel.AuthTypeOAuth2)
	issuerVal := ""
	if authType == string(authModel.AuthTypeOIDC) {
		protocol = string(authModel.AuthTypeOIDC)
		issuerVal = strings.TrimSpace(issuer)
	}

	var identity model.UserExternalIdentity
	err := r.getDB(ctx).
		Where("user_id = ? AND provider = ? AND issuer = ? AND protocol = ?", userID, provider, issuerVal, protocol).
		First(&identity).Error
	if err == nil {
		identity.Subject = oauthID
		return r.getDB(ctx).Save(&identity).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	identity = model.UserExternalIdentity{
		UserID:   userID,
		Provider: provider,
		Subject:  oauthID,
		Issuer:   issuerVal,
		Protocol: protocol,
	}
	return r.getDB(ctx).Create(&identity).Error
}

func (r *AuthRepository) GetUserByOAuthID(
	ctx context.Context,
	provider, oauthID string,
) (model.User, error) {
	var binding model.UserExternalIdentity
	err := r.getDB(ctx).
		Where("provider = ? AND subject = ? AND protocol = ?", provider, oauthID, string(authModel.AuthTypeOAuth2)).
		First(&binding).
		Error
	if err != nil {
		return model.User{}, err
	}

	return r.GetUserByID(ctx, binding.UserID)
}

func (r *AuthRepository) GetUserByOIDC(
	ctx context.Context,
	provider, oauthID, issuer string,
) (model.User, error) {
	var binding model.UserExternalIdentity
	err := r.getDB(ctx).
		Where(
			"provider = ? AND subject = ? AND issuer = ? AND protocol = ?",
			provider,
			oauthID,
			issuer,
			string(authModel.AuthTypeOIDC),
		).
		First(&binding).Error
	if err != nil {
		return model.User{}, err
	}

	return r.GetUserByID(ctx, binding.UserID)
}

func (r *AuthRepository) GetOAuthInfo(userId string, provider string) (model.UserExternalIdentity, error) {
	var identity model.UserExternalIdentity
	err := r.db().
		Where("user_id = ? AND provider = ? AND protocol = ?", userId, provider, string(authModel.AuthTypeOAuth2)).
		First(&identity).Error
	if err != nil {
		return model.UserExternalIdentity{}, err
	}
	return identity, nil
}

func (r *AuthRepository) GetOAuthOIDCInfo(
	userId string,
	provider string,
	issuer string,
) (model.UserExternalIdentity, error) {
	var identity model.UserExternalIdentity
	err := r.db().
		Where("user_id = ? AND provider = ? AND issuer = ? AND protocol = ?", userId, provider, issuer, string(authModel.AuthTypeOIDC)).
		First(&identity).
		Error
	if err != nil {
		return model.UserExternalIdentity{}, err
	}
	return identity, nil
}

func (r *AuthRepository) CreatePasskey(ctx context.Context, passkey *authModel.Passkey) error {
	return r.getDB(ctx).Create(&model.WebAuthnCredential{
		ID:             passkey.ID,
		UserID:         passkey.UserID,
		CredentialID:   passkey.CredentialID,
		CredentialJSON: passkey.CredentialJSON,
		PublicKey:      passkey.PublicKey,
		SignCount:      passkey.SignCount,
		LastUsedAt:     passkey.LastUsedAt,
		DeviceName:     passkey.DeviceName,
		AAGUID:         passkey.AAGUID,
		CreatedAt:      passkey.CreatedAt,
		UpdatedAt:      passkey.UpdatedAt,
	}).Error
}

func (r *AuthRepository) ListPasskeysByUserID(userID string) ([]authModel.Passkey, error) {
	var rows []model.WebAuthnCredential
	if err := r.db().
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	passkeys := make([]authModel.Passkey, 0, len(rows))
	for _, row := range rows {
		passkeys = append(passkeys, authModel.Passkey{
			ID:             row.ID,
			UserID:         row.UserID,
			CredentialID:   row.CredentialID,
			CredentialJSON: row.CredentialJSON,
			PublicKey:      row.PublicKey,
			SignCount:      row.SignCount,
			LastUsedAt:     row.LastUsedAt,
			DeviceName:     row.DeviceName,
			AAGUID:         row.AAGUID,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}
	return passkeys, nil
}

func (r *AuthRepository) GetPasskeyByCredentialID(credentialID string) (authModel.Passkey, error) {
	var row model.WebAuthnCredential
	if err := r.db().
		Where("credential_id = ?", credentialID).
		First(&row).Error; err != nil {
		return authModel.Passkey{}, err
	}
	return authModel.Passkey{
		ID:             row.ID,
		UserID:         row.UserID,
		CredentialID:   row.CredentialID,
		CredentialJSON: row.CredentialJSON,
		PublicKey:      row.PublicKey,
		SignCount:      row.SignCount,
		LastUsedAt:     row.LastUsedAt,
		DeviceName:     row.DeviceName,
		AAGUID:         row.AAGUID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func (r *AuthRepository) UpdatePasskeyUsage(
	ctx context.Context,
	passkeyID string,
	signCount uint32,
	lastUsedAt int64,
) error {
	return r.getDB(ctx).
		Model(&model.WebAuthnCredential{}).
		Where("id = ?", passkeyID).
		Updates(map[string]any{
			"sign_count":   signCount,
			"last_used_at": lastUsedAt,
		}).Error
}

func (r *AuthRepository) UpdatePasskeyDeviceName(
	ctx context.Context,
	userID, passkeyID string,
	deviceName string,
) error {
	return r.getDB(ctx).
		Model(&model.WebAuthnCredential{}).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Update("device_name", deviceName).Error
}

func (r *AuthRepository) DeletePasskeyByID(
	ctx context.Context,
	userID, passkeyID string,
) error {
	return r.getDB(ctx).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Delete(&model.WebAuthnCredential{}).Error
}

func (r *AuthRepository) CacheSetPasskeySession(
	key string,
	val any,
	ttl time.Duration,
) {
	_ = r.cache.SetWithTTL(key, val, 1, ttl)
}

func (r *AuthRepository) CacheGetPasskeySession(key string) (any, error) {
	value, found, err := r.cache.Get(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (r *AuthRepository) CacheDeletePasskeySession(key string) {
	r.cache.Delete(key)
}
