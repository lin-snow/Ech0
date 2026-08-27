// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package keyvalue

import (
	"context"

	"github.com/lin-snow/ech0/internal/cache"
	model "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type KeyValueRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewKeyValueRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) *KeyValueRepository {
	return &KeyValueRepository{
		db:    dbProvider,
		cache: cache,
	}
}

func (keyvalueRepository *KeyValueRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return keyvalueRepository.db()
}

func (keyvalueRepository *KeyValueRepository) GetKeyValue(ctx context.Context, key string) (string, error) {
	cacheKey := GetKeyValueCacheKey(key)
	return cache.ReadThroughTypedUnlessTx[string](
		ctx,
		keyvalueRepository.cache,
		cacheKey,
		1,
		func(ctx context.Context) (string, error) {
			var kv model.KeyValue
			if err := keyvalueRepository.getDB(ctx).Where("key = ?", key).First(&kv).Error; err != nil {
				return "", err
			}
			return kv.Value, nil
		},
		func() (string, error) {
			var kv model.KeyValue
			if err := keyvalueRepository.db().Where("key = ?", key).First(&kv).Error; err != nil {
				return "", err
			}
			return kv.Value, nil
		})
}

func (keyvalueRepository *KeyValueRepository) AddKeyValue(
	ctx context.Context,
	key string,
	value string,
) error {
	cacheKey := GetKeyValueCacheKey(key)
	cache.InvalidateKeys(keyvalueRepository.cache, cacheKey)

	if err := keyvalueRepository.getDB(ctx).Create(&model.KeyValue{
		Key:   key,
		Value: value,
	}).Error; err != nil {
		return err
	}

	keyvalueRepository.cache.Set(cacheKey, value, 1)

	return nil
}

func (keyvalueRepository *KeyValueRepository) DeleteKeyValue(
	ctx context.Context,
	key string,
) error {
	cache.InvalidateKeys(keyvalueRepository.cache, GetKeyValueCacheKey(key))

	if err := keyvalueRepository.getDB(ctx).Where("key = ?", key).Delete(&model.KeyValue{}).Error; err != nil {
		return err
	}

	return nil
}

func (keyvalueRepository *KeyValueRepository) UpdateKeyValue(
	ctx context.Context,
	key string,
	value string,
) error {
	cacheKey := GetKeyValueCacheKey(key)
	cache.InvalidateKeys(keyvalueRepository.cache, cacheKey)

	if err := keyvalueRepository.getDB(ctx).Model(&model.KeyValue{}).Where("key = ?", key).Update("value", value).Error; err != nil {
		return err
	}

	keyvalueRepository.cache.Set(cacheKey, value, 1)

	return nil
}

func (keyvalueRepository *KeyValueRepository) AddOrUpdateKeyValue(
	ctx context.Context,
	key string,
	value string,
) error {
	result := keyvalueRepository.getDB(ctx).
		Model(&model.KeyValue{}).
		Where("key = ?", key).
		Update("value", value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if err := keyvalueRepository.getDB(ctx).Create(&model.KeyValue{
			Key:   key,
			Value: value,
		}).Error; err != nil {
			return err
		}
	}

	cacheKey := GetKeyValueCacheKey(key)
	cache.InvalidateKeys(keyvalueRepository.cache, cacheKey)
	keyvalueRepository.cache.Set(cacheKey, value, 1)

	return nil
}
