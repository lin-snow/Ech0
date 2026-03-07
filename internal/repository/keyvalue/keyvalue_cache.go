package keyvalue

const (
	keyValueCachePrefix = "keyvalue:item"
)

func GetKeyValueCacheKey(key string) string {
	return keyValueCachePrefix + ":" + key
}
