// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

// The browser holds a cryptographically random code. A copy of the database
// alone cannot recover unredeemed token pairs. Separate the lookup and AEAD keys.
func challengeKeys(code string) (string, cipher.AEAD, error) {
	digest := sha256.Sum256([]byte("ech0:lookup:" + code))
	key := sha256.Sum256([]byte("ech0:payload:" + code))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", nil, err
	}
	aead, err := cipher.NewGCMWithRandomNonce(block)
	return hex.EncodeToString(digest[:]), aead, err
}

func (r *AuthRepository) saveChallenge(code string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	id, aead, err := challengeKeys(code)
	if err != nil {
		return err
	}
	now := time.Now()
	record := authModel.Challenge{ID: id, Payload: aead.Seal(nil, nil, data, nil), ExpiresAt: now.Add(ttl).UnixMilli()}
	return r.db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at <= ?", now.UnixMilli()).Delete(&authModel.Challenge{}).Error; err != nil {
			return err
		}
		return tx.Create(&record).Error
	})
}

func (r *AuthRepository) consumeChallenge(code string, dest any) error {
	id, aead, err := challengeKeys(code)
	if err != nil {
		return err
	}
	var ciphertext []byte
	if err := r.db().Raw("DELETE FROM challenges WHERE id=? AND expires_at>? RETURNING payload", id, time.Now().UnixMilli()).Row().Scan(&ciphertext); err != nil {
		return err
	}
	data, err := aead.Open(nil, nil, ciphertext, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

type storedTokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (r *AuthRepository) StoreOAuthCode(code string, pair *authModel.TokenPair, ttl time.Duration) error {
	if code == "" || pair == nil || ttl <= 0 {
		return errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}
	return r.saveChallenge("oauth:"+code, storedTokenPair(*pair), ttl)
}

func (r *AuthRepository) GetAndDeleteOAuthCode(code string) (*authModel.TokenPair, error) {
	var pair storedTokenPair
	if err := r.consumeChallenge("oauth:"+code, &pair); err != nil {
		return nil, errors.New(commonModel.EXCHANGE_CODE_INVALID)
	}
	return (*authModel.TokenPair)(&pair), nil
}

func (r *AuthRepository) StorePasskeySession(key string, val authModel.PasskeySession, ttl time.Duration) error {
	return r.saveChallenge("webauthn:"+key, val, ttl)
}

func (r *AuthRepository) ConsumePasskeySession(key string) (authModel.PasskeySession, error) {
	var session authModel.PasskeySession
	err := r.consumeChallenge("webauthn:"+key, &session)
	return session, err
}
