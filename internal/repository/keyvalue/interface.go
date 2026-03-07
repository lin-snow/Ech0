package keyvalue

import "context"

type KeyValueRepositoryInterface interface {
	// GetKeyValue 根据键获取值
	GetKeyValue(ctx context.Context, key string) (string, error)

	// AddKeyValue 添加键值对
	AddKeyValue(ctx context.Context, key string, value string) error

	// DeleteKeyValue 删除键值对
	DeleteKeyValue(ctx context.Context, key string) error

	// UpdateKeyValue 更新键值对
	UpdateKeyValue(ctx context.Context, key string, value string) error

	// AddOrUpdateKeyValue 添加或更新键值对
	AddOrUpdateKeyValue(ctx context.Context, key string, value string) error
}
