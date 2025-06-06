//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"gorm.io/gorm"
)

// BuildHandlers 使用wire生成的代码来构建Handlers实例
func BuildHandlers(db *gorm.DB) (*Handlers, error) {
	wire.Build(
		UserSet,
		EchoSet,
		CommonSet,
		SettingSet,
		//TodoSet,
		//ConnectSet,
		NewHandlers,
	)

	return &Handlers{}, nil
}

// UserSet 包含了构建 UserHandler 所需的所有 Provider
var UserSet = wire.NewSet(
	userRepository.NewUserRepository,
	userService.NewUserService,
	userHandler.NewUserHandler,
)

// EchoSet 包含了构建 EchoHandler 所需的所有 Provider
var EchoSet = wire.NewSet(
	echoRepository.NewEchoRepository,
	echoService.NewEchoService,
	echoHandler.NewEchoHandler,
)

// CommonSet 包含了构建 CommonHandler 所需的所有 Provider
var CommonSet = wire.NewSet(
	commonRepository.NewCommonRepository,
	commonService.NewCommonService,
	commonHandler.NewCommonHandler,
)

// SettingSet 包含了构建 SettingHandler 所需的所有 Provider
var SettingSet = wire.NewSet(
	keyvalueRepository.NewKeyValueRepository,
	settingService.NewSettingService,
	settingHandler.NewSettingHandler,
)

//// TodoSet 包含了构建 TodoHandler 所需的所有 Provider
//var TodoSet = wire.NewSet()
//
//// ConnectSet 包含了构建 ConnectHandler 所需的所有 Provider
//var ConnectSet = wire.NewSet()
