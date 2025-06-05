package service

import model "github.com/lin-snow/ech0/internal/model/user"

type UserServiceInterface interface {
	GetUserByID(userId int) (*model.User, error)
}
