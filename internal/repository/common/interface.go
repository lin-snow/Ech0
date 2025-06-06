package repository

import userModel "github.com/lin-snow/ech0/internal/model/user"

type CommonRepositoryInterface interface {
	GetUserByUserId(userid uint) (userModel.User, error)
}
