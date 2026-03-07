package database

import (
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func ProvideDBProvider() func() *gorm.DB {
	var once sync.Once
	return func() *gorm.DB {
		once.Do(InitDatabase)
		return GetDB()
	}
}

var ProviderSet = wire.NewSet(ProvideDBProvider)
