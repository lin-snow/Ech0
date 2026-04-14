package service

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
)

type Service interface {
	InitOwner(registerDto *authModel.RegisterDto) error
	Register(registerDto *authModel.RegisterDto) error
	UpdateUser(ctx context.Context, userdto model.UserInfoDto) error
	UpdateUserAdmin(ctx context.Context, id string) error
	GetAllUsers(ctx context.Context) ([]model.User, error)
	GetOwner() (model.User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByID(userId string) (model.User, error)
}

type (
	SettingService = settingService.Service
	FileService    = fileService.Service
)

type UserRepo interface {
	GetUserByID(ctx context.Context, id string) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetAllUsers(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, newUser *model.User) error
	GetOwner(ctx context.Context) (model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
}

type InstallStateRepo interface {
	IsInitialized(ctx context.Context) (bool, error)
	MarkInitialized(ctx context.Context) error
}

type Repository interface {
	UserRepo
	InstallStateRepo
}
