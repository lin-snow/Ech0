package service

import (
	"errors"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/user"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

type UserService struct {
	userRepository repository.UserRepositoryInterface
}

func NewUserService(userRepository repository.UserRepositoryInterface) UserServiceInterface {
	return &UserService{
		userRepository: userRepository,
	}
}

func (userService *UserService) Login(loginDto *authModel.LoginDto) (string, error) {
	if loginDto.Username == "" || loginDto.Password == "" {
		return "", errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
	}

	loginDto.Password = cryptoUtil.MD5Encrypt(loginDto.Password)

	user, err := userService.userRepository.GetUserByUsername(loginDto.Username)
	if err != nil {
		return "", err
	}

	if user.Password != loginDto.Password {
		return "", errors.New(commonModel.PASSWORD_INCORRECT)
	}

	// 生成 Token
	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (userService *UserService) Register(registerDto *authModel.RegisterDto) error {
	// 检查用户数量是否超过限制
	users, err := userService.userRepository.GetAllUsers()
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return errors.New(commonModel.USER_COUNT_EXCEED_LIMIT)
	}

	// 将密码进行 MD5 加密
	registerDto.Password = cryptoUtil.MD5Encrypt(registerDto.Password)

	newUser := model.User{
		Username: registerDto.Username,
		Password: registerDto.Password,
		IsAdmin:  false,
	}

	// 检查用户是否已经存在
	user, err := userService.userRepository.GetUserByUsername(newUser.Username)
	if err == nil && user.ID != model.USER_NOT_EXISTS_ID {
		return errors.New(commonModel.USERNAME_HAS_EXISTS)
	}

	// 检查是否该系统第一次注册用户
	if len(users) > 0 {
		// 第一个注册的用户为系统管理员
		newUser.IsAdmin = true
	}

	if err := userService.userRepository.CreateUser(&newUser); err != nil {
		return err
	}

	return nil
}

func (userService *UserService) GetUserByID(userId int) (*model.User, error) {
	return userService.userRepository.GetUserByID(userId)
}
