package di

import userHandler "github.com/lin-snow/ech0/internal/handler/user"

// Handlers 聚合各个模块的Handler
type Handlers struct {
	UserHandler *userHandler.UserHandler
}

// NewHandlers 创建Handlers实例
func NewHandlers(uh *userHandler.UserHandler) *Handlers {
	return &Handlers{
		UserHandler: uh,
	}
}
