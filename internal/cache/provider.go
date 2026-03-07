package cache

import "github.com/google/wire"

func ProvideCache() (ICache[string, any], error) {
	return NewCache[string, any]()
}

var ProviderSet = wire.NewSet(ProvideCache)
