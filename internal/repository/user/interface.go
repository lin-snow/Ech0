package repository

import model "github.com/lin-snow/ech0/internal/model/user"

type UserRepositoryInterface interface {
	GetUserByID(id int) (model.User, error)
	GetUserByUsername(username string) (model.User, error)
	GetAllUsers() ([]model.User, error)
	CreateUser(newUser *model.User) error
	GetSysAdmin() (model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id uint) error
}
