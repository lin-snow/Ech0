package keyvalue

type KeyValueRepositoryInterface interface {
	GetKeyValue(key string) (interface{}, error)
	AddKeyValue(key string, value interface{}) error
	DeleteKeyValue(key string) error
	UpdateKeyValue(key string, value interface{}) error
}
