// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"fmt"

	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/kvstore"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

// capsuleRuntime 是胶囊类命令要用的最小运行时：库 + 存储 + 设置 KV + 事务。
// 它平铺装配 wire_gen 中 BuildApp 的同名依赖，但刻意不碰 HTTP server、事件总线、
// 作业管理器与定时任务——胶囊命令是一次性进程，起这些只会拖慢启动并制造副作用
// （尤其是事件总线：导入永不发布事件，连总线都不该存在）。
type capsuleRuntime struct {
	db      *gorm.DB
	kv      kvstore.Store
	cache   cache.ICache[string, any]
	storage *storage.Manager
	tx      transaction.Transactor
}

// newCapsuleRuntime 打开数据库并装配存储层。
//
// InitDatabase 内部失败走 panic（util.HandlePanicError 的既有约定），而 CLI 需要
// 的是可读的退出信息，故在此收口成 error。
func newCapsuleRuntime() (rt *capsuleRuntime, err error) {
	defer func() {
		if r := recover(); r != nil {
			rt, err = nil, fmt.Errorf("initialise runtime: %v", r)
		}
	}()

	dbProvider := database.ProvideDBProvider()
	db := dbProvider()
	if db == nil {
		return nil, fmt.Errorf("database is not available")
	}

	appCache, err := cache.ProvideCache()
	if err != nil {
		return nil, fmt.Errorf("initialise cache: %w", err)
	}

	// 走持久化 KV 而不是 config 默认值：面板配置的 S3 才是存储的真相来源，
	// 传 nil 会让导出在对象存储实例上静默退化成「只看本地盘」。
	durableKV := kvstore.NewPersistent(keyvalueRepository.NewKeyValueRepository(dbProvider, appCache))

	return &capsuleRuntime{
		db:      db,
		kv:      durableKV,
		cache:   appCache,
		storage: storage.NewStorageManager(durableKV),
		tx:      transaction.NewGormTransactor(dbProvider),
	}, nil
}

// selector 返回当前生效的存储后端选择器。
func (rt *capsuleRuntime) selector() *storage.StorageSelector {
	return rt.storage.GetSelector()
}
