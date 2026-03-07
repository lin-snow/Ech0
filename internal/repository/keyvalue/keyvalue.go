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
) KeyValueRepositoryInterface {
	return &KeyValueRepository{
		db:    dbProvider,
		cache: cache,
	}
}

// getDB 从上下文中获取事务
func (keyvalueRepository *KeyValueRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return keyvalueRepository.db()
}

// GetKeyValue 根据键获取值
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

// AddKeyValue 添加键值对
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

	// 添加新的缓存
	keyvalueRepository.cache.Set(cacheKey, value, 1)

	return nil
}

// DeleteKeyValue 删除键值对
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

// UpdateKeyValue 更新键值对
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

	// 添加新的缓存
	keyvalueRepository.cache.Set(cacheKey, value, 1)

	return nil
}

// AddOrUpdateKeyValue 添加或更新键值对
func (keyvalueRepository *KeyValueRepository) AddOrUpdateKeyValue(
	ctx context.Context,
	key string,
	value string,
) error {
	// 先尝试更新
	result := keyvalueRepository.getDB(ctx).
		Model(&model.KeyValue{}).
		Where("key = ?", key).
		Update("value", value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// 如果没有行被更新，说明该键不存在，执行添加操作
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
