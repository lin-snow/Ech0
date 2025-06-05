package service

import (
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
)

type UserServiceInterface interface {
	Login(user *authModel.LoginDto) (string, error)
	GetUserByID(userId int) (*model.User, error)
	Register(registerDto *authModel.RegisterDto) error
}
