package service

import (
	"errors"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	model "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/user"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

type UserService struct {
	userRepository repository.UserRepositoryInterface
	settingService settingService.SettingServiceInterface
}

func NewUserService(userRepository repository.UserRepositoryInterface, settingService settingService.SettingServiceInterface) UserServiceInterface {
	return &UserService{
		userRepository: userRepository,
		settingService: settingService,
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
	if len(users) > authModel.MAX_USER_COUNT {
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
	if len(users) == 0 {
		// 第一个注册的用户为系统管理员
		newUser.IsAdmin = true
	}

	// 检查是否开放注册
	var setting settingModel.SystemSetting
	if err := userService.settingService.GetSetting(&setting); err != nil {
		return err
	}
	if len(users) != 0 && !setting.AllowRegister {
		return errors.New(commonModel.USER_REGISTER_NOT_ALLOW)
	}

	if err := userService.userRepository.CreateUser(&newUser); err != nil {
		return err
	}

	return nil
}

func (userService *UserService) UpdateUser(userid uint, userdto model.UserInfoDto) error {
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查是否需要更新用户名
	if userdto.Username != "" && userdto.Username != user.Username {
		// 检查用户名是否已存在
		existingUser, _ := userService.userRepository.GetUserByUsername(userdto.Username)
		if existingUser.ID != model.USER_NOT_EXISTS_ID {
			return errors.New(commonModel.USERNAME_ALREADY_EXISTS)
		}
		user.Username = userdto.Username
	}

	// 检查是否需要更新密码
	if userdto.Password != "" && cryptoUtil.MD5Encrypt(userdto.Password) != user.Password {
		// 检查密码是否为空
		if userdto.Password == "" {
			return errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
		}
		// 更新密码
		user.Password = cryptoUtil.MD5Encrypt(userdto.Password)
	}

	// 检查是否需要更新头像
	if userdto.Avatar != "" && userdto.Avatar != user.Avatar {
		// 更新头像
		user.Avatar = userdto.Avatar
	}
	// 更新用户信息
	if err := userService.userRepository.UpdateUser(&user); err != nil {
		return err
	}

	return nil
}

func (userService *UserService) UpdateUserAdmin(userid uint) error {
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}

	user.IsAdmin = !user.IsAdmin

	// 更新用户信息
	if err := userService.userRepository.UpdateUser(&user); err != nil {
		return err
	}

	return nil
}

func (userService *UserService) GetAllUsers() ([]model.User, error) {
	allures, err := userService.userRepository.GetAllUsers()
	if err != nil {
		return nil, err
	}

	sysadmin, err := userService.GetSysAdmin()
	if err != nil {
		return nil, err
	}

	// 处理用户信息(去掉管理员用户)
	for i := range allures {
		if allures[i].ID == sysadmin.ID {
			allures = append(allures[:i], allures[i+1:]...)
			break
		}
	}

	// 处理用户信息(去掉密码)
	for i := range allures {
		allures[i].Password = ""
	}

	return allures, nil
}

func (userService *UserService) GetSysAdmin() (model.User, error) {
	sysadmin, err := userService.userRepository.GetSysAdmin()
	if err != nil {
		return model.User{}, err
	}

	return sysadmin, nil
}

func (userService *UserService) DeleteUser(userid, id uint) error {
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}

	sysadmin, err := userService.GetSysAdmin()
	if err != nil {
		return err
	}

	if id == user.ID || id == sysadmin.ID {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	if err := userService.userRepository.DeleteUser(id); err != nil {
		return err
	}

	return nil
}

func (userService *UserService) GetUserByID(userId int) (model.User, error) {
	return userService.userRepository.GetUserByID(userId)
}
