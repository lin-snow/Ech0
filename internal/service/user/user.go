package service

import (
	model "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/user"
)

type UserService struct {
	userRepository repository.UserRepositoryInterface
}

func NewUserService(userRepository repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{
		userRepository: userRepository,
	}
}

func (service *UserService) GetUserByID(userId int) (*model.User, error) {
	return service.userRepository.GetUserByID(userId)
}
