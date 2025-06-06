package di

import (
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	todoHandler "github.com/lin-snow/ech0/internal/handler/todo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
)

// Handlers 聚合各个模块的Handler
type Handlers struct {
	UserHandler    *userHandler.UserHandler
	EchoHandler    *echoHandler.EchoHandler
	CommonHandler  *commonHandler.CommonHandler
	SettingHandler *settingHandler.SettingHandler
	TodoHandler    *todoHandler.TodoHandler
	ConnectHandler *connectHandler.ConnectHandler
}

// NewHandlers 创建Handlers实例
func NewHandlers(
	userHandler *userHandler.UserHandler,
	echoHandler *echoHandler.EchoHandler,
	commonHandler *commonHandler.CommonHandler,
	settingHandler *settingHandler.SettingHandler,
	todoHandler *todoHandler.TodoHandler,
	connectHandler *connectHandler.ConnectHandler,
) *Handlers {
	return &Handlers{
		UserHandler:    userHandler,
		EchoHandler:    echoHandler,
		CommonHandler:  commonHandler,
		SettingHandler: settingHandler,
		TodoHandler:    todoHandler,
		ConnectHandler: connectHandler,
	}
}
