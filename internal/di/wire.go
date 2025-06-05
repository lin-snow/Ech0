//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"gorm.io/gorm"
)

// BuildHandlers 使用wire生成的代码来构建Handlers实例
func BuildHandlers(db *gorm.DB) (*Handlers, error) {
	wire.Build(
		UserSet,
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
