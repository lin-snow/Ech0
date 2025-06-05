package repository

import model "github.com/lin-snow/ech0/internal/model/user"

type UserRepositoryInterface interface {
	GetUserByID(id int) (*model.User, error)
}
