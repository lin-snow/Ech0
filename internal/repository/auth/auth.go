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

func NewAuthRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) *AuthRepository {
	return &AuthRepository{
		db:    dbProvider,
		cache: cache,
	}
}

func (authRepository *AuthRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return authRepository.db()
}

func (authRepository *AuthRepository) RevokeToken(jti string, remainTTL time.Duration) {
	if jti == "" || remainTTL <= 0 {
		return
	}
	authRepository.cache.SetWithTTL(fmt.Sprintf("%s%s", blacklistPrefix, jti), true, 1, remainTTL)
}

func (authRepository *AuthRepository) IsTokenRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	_, found, _ := authRepository.cache.Get(fmt.Sprintf("%s%s", blacklistPrefix, jti))
	return found
}

func (authRepository *AuthRepository) StoreOAuthCode(code string, pair *authModel.TokenPair, ttl time.Duration) {
	if code == "" || pair == nil || ttl <= 0 {
		return
	}
	authRepository.cache.SetWithTTL(oauthCodePrefix+code, pair, 1, ttl)
}

func (authRepository *AuthRepository) GetAndDeleteOAuthCode(code string) (*authModel.TokenPair, error) {
	if code == "" {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}

	key := oauthCodePrefix + code
	val, found, _ := authRepository.cache.Get(key)
	if !found {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}

	authRepository.cache.Delete(key)

	pair, ok := val.(*authModel.TokenPair)
	if !ok {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}
	return pair, nil
}

func (authRepository *AuthRepository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	user := model.User{}
	if err := authRepository.getDB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (authRepository *AuthRepository) getUserByID(ctx context.Context, id string) (model.User, error) {
	var user model.User
	if err := authRepository.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (authRepository *AuthRepository) GetUserByID(ctx context.Context, id string) (model.User, error) {
	return authRepository.getUserByID(ctx, id)
}

func (authRepository *AuthRepository) BindOAuth(
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
	err := authRepository.getDB(ctx).
		Where("user_id = ? AND provider = ? AND issuer = ? AND protocol = ?", userID, provider, issuerVal, protocol).
		First(&identity).Error
	if err == nil {
		identity.Subject = oauthID
		return authRepository.getDB(ctx).Save(&identity).Error
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
	return authRepository.getDB(ctx).Create(&identity).Error
}

func (authRepository *AuthRepository) GetUserByOAuthID(
	ctx context.Context,
	provider, oauthID string,
) (model.User, error) {
	var binding model.UserExternalIdentity
	err := authRepository.getDB(ctx).
		Where("provider = ? AND subject = ? AND protocol = ?", provider, oauthID, string(authModel.AuthTypeOAuth2)).
		First(&binding).Error
	if err != nil {
		return model.User{}, err
	}
	return authRepository.getUserByID(ctx, binding.UserID)
}

func (authRepository *AuthRepository) GetUserByOIDC(
	ctx context.Context,
	provider, oauthID, issuer string,
) (model.User, error) {
	var binding model.UserExternalIdentity
	err := authRepository.getDB(ctx).
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
	return authRepository.getUserByID(ctx, binding.UserID)
}

func (authRepository *AuthRepository) GetOAuthInfo(
	userID string,
	provider string,
) (model.UserExternalIdentity, error) {
	var identity model.UserExternalIdentity
	err := authRepository.db().
		Where("user_id = ? AND provider = ? AND protocol = ?", userID, provider, string(authModel.AuthTypeOAuth2)).
		First(&identity).Error
	if err != nil {
		return model.UserExternalIdentity{}, err
	}
	return identity, nil
}

func (authRepository *AuthRepository) GetOAuthOIDCInfo(
	userID string,
	provider string,
	issuer string,
) (model.UserExternalIdentity, error) {
	var identity model.UserExternalIdentity
	err := authRepository.db().
		Where("user_id = ? AND provider = ? AND issuer = ? AND protocol = ?", userID, provider, issuer, string(authModel.AuthTypeOIDC)).
		First(&identity).Error
	if err != nil {
		return model.UserExternalIdentity{}, err
	}
	return identity, nil
}

func (authRepository *AuthRepository) CreatePasskey(
	ctx context.Context,
	passkey *authModel.Passkey,
) error {
	return authRepository.getDB(ctx).Create(&model.WebAuthnCredential{
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

func (authRepository *AuthRepository) ListPasskeysByUserID(userID string) ([]authModel.Passkey, error) {
	var rows []model.WebAuthnCredential
	if err := authRepository.db().
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

func (authRepository *AuthRepository) GetPasskeyByCredentialID(credentialID string) (authModel.Passkey, error) {
	var row model.WebAuthnCredential
	if err := authRepository.db().
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

func (authRepository *AuthRepository) UpdatePasskeyUsage(
	ctx context.Context,
	passkeyID string,
	signCount uint32,
	lastUsedAt int64,
) error {
	return authRepository.getDB(ctx).
		Model(&model.WebAuthnCredential{}).
		Where("id = ?", passkeyID).
		Updates(map[string]any{
			"sign_count":   signCount,
			"last_used_at": lastUsedAt,
		}).Error
}

func (authRepository *AuthRepository) UpdatePasskeyDeviceName(
	ctx context.Context,
	userID, passkeyID string,
	deviceName string,
) error {
	return authRepository.getDB(ctx).
		Model(&model.WebAuthnCredential{}).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Update("device_name", deviceName).Error
}

func (authRepository *AuthRepository) DeletePasskeyByID(
	ctx context.Context,
	userID, passkeyID string,
) error {
	return authRepository.getDB(ctx).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Delete(&model.WebAuthnCredential{}).Error
}

func (authRepository *AuthRepository) CacheSetPasskeySession(
	key string,
	val any,
	ttl time.Duration,
) {
	_ = authRepository.cache.SetWithTTL(key, val, 1, ttl)
}

func (authRepository *AuthRepository) CacheGetPasskeySession(key string) (any, error) {
	value, found, err := authRepository.cache.Get(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, gorm.ErrRecordNotFound
	}
	return value, nil
}

func (authRepository *AuthRepository) CacheDeletePasskeySession(key string) {
	authRepository.cache.Delete(key)
}
