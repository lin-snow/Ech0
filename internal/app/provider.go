package app

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideLifecycles(
	_ *event.RegisteredRegistrar,
	tasker *task.Tasker,
	httpServer *server.Server,
) []Lifecycle {
	// 启动顺序保持为 task -> http，确保后台任务基础能力先就绪，再对外暴露服务。
	return []Lifecycle{tasker, httpServer}
}

var ProviderSet = wire.NewSet(ProvideLifecycles, NewApp)
