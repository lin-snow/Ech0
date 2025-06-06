package repository

import (
	model "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type CommonRepositoryInterface interface {
	GetUserByUserId(userid uint) (userModel.User, error)
	GetSysAdmin() (userModel.User, error)
	GetStatus() (model.Status, error)
	GetAllUsers() ([]userModel.User, error)
	GetAllEchos(showPrivate bool) ([]echoModel.Echo, error)
}
