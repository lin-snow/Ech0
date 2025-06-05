package di

import (
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
)

// Handlers 聚合各个模块的Handler
type Handlers struct {
	UserHandler *userHandler.UserHandler
	EchoHandler *echoHandler.EchoHandler
}

// NewHandlers 创建Handlers实例
func NewHandlers(
	userHandler *userHandler.UserHandler,
	echoHandler *echoHandler.EchoHandler,
) *Handlers {
	return &Handlers{
		UserHandler: userHandler,
		EchoHandler: echoHandler,
	}
}
