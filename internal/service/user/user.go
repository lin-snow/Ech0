package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	model "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

// UserService 用户服务结构体，提供用户相关的业务逻辑处理
type UserService struct {
	transactor     transaction.Transactor
	userRepository Repository
	settingService SettingService
	fileService    FileService
	publisher      *publisher.Publisher
}

// NewUserService 创建并返回新的用户服务实例
//
// 参数:
//   - userRepository: 用户数据层接口实现
//   - settingService: 系统设置数据层接口实现
//
// 返回:
//   - *UserService: 用户服务实现
func NewUserService(
	tx transaction.Transactor,
	userRepository Repository,
	settingService SettingService,
	fileService FileService,
	publisher *publisher.Publisher,
) *UserService {
	return &UserService{
		transactor:     tx,
		userRepository: userRepository,
		settingService: settingService,
		fileService:    fileService,
		publisher:      publisher,
	}
}

// InitOwner 初始化 Owner 账号
//
// 参数:
//   - registerDto: 注册数据传输对象，包含用户名和密码
//
// 返回:
//   - error: 初始化过程中的错误信息
func (userService *UserService) InitOwner(registerDto *authModel.RegisterDto) error {
	if registerDto.Username == "" || registerDto.Password == "" {
		return errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
	}
	email := strings.TrimSpace(registerDto.Email)
	if email == "" {
		return errors.New("邮箱不能为空")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("邮箱格式无效")
	}

	var owner model.User
	if err := userService.transactor.Run(context.Background(), func(ctx context.Context) error {
		initialized, err := userService.userRepository.IsInitialized(ctx)
		if err != nil {
			return err
		}
		if initialized {
			return commonModel.NewBizError(commonModel.ErrCodeInitAlreadyDone, commonModel.SYSTEM_ALREADY_INITED)
		}

		users, err := userService.userRepository.GetAllUsers(ctx)
		if err != nil {
			return err
		}
		if len(users) > 0 {
			return commonModel.NewBizError(commonModel.ErrCodeInitOwnerExists, commonModel.OWNER_ALREADY_EXISTS)
		}

		// 检查用户是否已经存在
		existingUser, err := userService.userRepository.GetUserByUsername(ctx, registerDto.Username)
		if err == nil && existingUser.ID != model.USER_NOT_EXISTS_ID {
			return errors.New(commonModel.USERNAME_HAS_EXISTS)
		}

		owner = model.User{
			Username: registerDto.Username,
			Email:    email,
			Password: cryptoUtil.MD5Encrypt(registerDto.Password),
			IsAdmin:  true,
			IsOwner:  true,
			Locale:   string(commonModel.DefaultLocale),
		}

		if err := userService.userRepository.CreateUser(ctx, &owner); err != nil {
			return err
		}

		return userService.userRepository.MarkInitialized(ctx)
	}); err != nil {
		return err
	}

	// 发布用户注册事件
	owner.Password = "" // 不包含密码信息
	if err := userService.publisher.UserCreated(
		context.Background(),
		contracts.UserCreatedEvent{User: owner},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish owner created event", zap.Error(err))
	}

	return nil
}

// Register 用户注册
// 注册普通用户，包括用户数量限制检查、注册权限检查等
//
// 参数:
//   - registerDto: 注册数据传输对象，包含用户名和密码
//
// 返回:
//   - error: 注册过程中的错误信息
func (userService *UserService) Register(registerDto *authModel.RegisterDto) error {
	initialized, err := userService.userRepository.IsInitialized(context.Background())
	if err != nil {
		return err
	}
	if !initialized {
		return commonModel.NewBizError(commonModel.ErrCodeInitInvalidState, commonModel.SIGNUP_FIRST)
	}

	// 检查用户数量是否超过限制
	users, err := userService.userRepository.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	if len(users) > authModel.MAX_USER_COUNT {
		return errors.New(commonModel.USER_COUNT_EXCEED_LIMIT)
	}

	// 将密码进行 MD5 加密
	registerDto.Password = cryptoUtil.MD5Encrypt(registerDto.Password)
	email := strings.TrimSpace(registerDto.Email)
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errors.New("邮箱格式无效")
		}
	}

	newUser := model.User{
		Username: registerDto.Username,
		Email:    email,
		Password: registerDto.Password,
		IsAdmin:  false,
		IsOwner:  false,
		Locale:   string(commonModel.DefaultLocale),
	}

	// 检查用户是否已经存在
	user, err := userService.userRepository.GetUserByUsername(context.Background(), newUser.Username)
	if err == nil && user.ID != model.USER_NOT_EXISTS_ID {
		return errors.New(commonModel.USERNAME_HAS_EXISTS)
	}

	// 检查是否开放注册
	var setting settingModel.SystemSetting
	if err := userService.settingService.GetSetting(&setting); err != nil {
		return err
	}
	if !setting.AllowRegister {
		return errors.New(commonModel.USER_REGISTER_NOT_ALLOW)
	}
	if err := userService.transactor.Run(context.Background(), func(ctx context.Context) error {
		if err := userService.userRepository.CreateUser(ctx, &newUser); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// 发布用户注册事件
	newUser.Password = "" // 不包含密码信息
	if err := userService.publisher.UserCreated(
		context.Background(),
		contracts.UserCreatedEvent{User: newUser},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish user created event", zap.Error(err))
	}

	return nil
}

// UpdateUser 更新用户信息
// 只有管理员可以更新用户信息，支持更新用户名、密码和头像
//
// 参数:
//   - userid: 执行更新操作的用户ID（必须为管理员）
//   - userdto: 用户信息数据传输对象，包含要更新的用户信息
//
// 返回:
//   - error: 更新过程中的错误信息
func (userService *UserService) UpdateUser(ctx context.Context, userdto model.UserInfoDto) error {
	userid := viewer.MustFromContext(ctx).UserID()
	// 检查执行操作的用户是否为管理员
	user, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查是否需要更新用户名
	if userdto.Username != "" && userdto.Username != user.Username {
		// 检查用户名是否已存在
		existingUser, err := userService.userRepository.GetUserByUsername(ctx, userdto.Username)
		if err == nil && existingUser.ID != user.ID {
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

	avatarChanged := false
	// 检查是否需要更新头像
	if userdto.Avatar != "" && userdto.Avatar != user.Avatar {
		// 更新头像
		user.Avatar = userdto.Avatar
		avatarChanged = true
	}
	if userdto.Locale != "" {
		user.Locale = i18nUtil.ResolveLocale(userdto.Locale)
	}
	if strings.TrimSpace(userdto.Email) != "" {
		if _, err := mail.ParseAddress(strings.TrimSpace(userdto.Email)); err != nil {
			return errors.New("邮箱格式无效")
		}
		user.Email = strings.TrimSpace(userdto.Email)
	}
	if err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		// 更新用户信息
		return userService.userRepository.UpdateUser(txCtx, &user)
	}); err != nil {
		return err
	}
	if avatarChanged && strings.TrimSpace(userdto.AvatarFileID) != "" {
		if err := userService.fileService.ConfirmTempFiles(ctx, []string{userdto.AvatarFileID}); err != nil {
			logUtil.GetLogger().Warn("confirm temp avatar file failed", zap.Error(err))
		}
	}

	// 发布用户更新事件
	user.Password = "" // 不包含密码信息
	if err := userService.publisher.UserUpdated(
		context.Background(),
		contracts.UserUpdatedEvent{User: user},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish user updated event", zap.Error(err))
	}

	return nil
}

// UpdateUserAdmin 更新用户的管理员权限
// 只有 Owner 可以修改其他用户的管理员权限，不能修改自己和 Owner 的权限
//
// 参数:
//   - userid: 执行操作的用户ID（必须为管理员）
//   - id: 要修改权限的用户ID
//
// 返回:
//   - error: 更新过程中的错误信息
func (userService *UserService) UpdateUserAdmin(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	// 检查执行操作的用户是否为 Owner
	operator, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return err
	}
	if !operator.IsOwner {
		return errors.New(commonModel.ONLY_OWNER_CAN_MANAGE)
	}

	// 检查要修改权限的用户是否存在
	user, err := userService.userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否尝试修改自己或 Owner 的权限
	if userid == user.ID || user.IsOwner {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	user.IsAdmin = !user.IsAdmin

	if err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		// 更新用户信息
		return userService.userRepository.UpdateUser(txCtx, &user)
	}); err != nil {
		return err
	}

	// 发布用户更新事件
	user.Password = "" // 不包含密码信息
	if err := userService.publisher.UserUpdated(
		context.Background(),
		contracts.UserUpdatedEvent{User: user},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish user updated event", zap.Error(err))
	}

	return nil
}

// GetAllUsers 获取所有用户列表
// 返回除 Owner 外的所有用户，并移除密码信息
//
// 返回:
//   - []model.User: 用户列表（不包含密码信息）
//   - error: 获取过程中的错误信息
func (userService *UserService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	// Only Admin can get all users
	userid := viewer.MustFromContext(ctx).UserID()
	caller, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !caller.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	allures, err := userService.userRepository.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	owner, err := userService.GetOwner()
	if err != nil {
		return nil, err
	}

	// 处理用户信息(去掉Owner用户)
	for i := range allures {
		if allures[i].ID == owner.ID {
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

// GetOwner 获取 Owner 信息
//
// 返回:
//   - model.User: Owner 用户信息
//   - error: 获取过程中的错误信息
func (userService *UserService) GetOwner() (model.User, error) {
	owner, err := userService.userRepository.GetOwner(context.Background())
	if err != nil {
		return model.User{}, err
	}

	return owner, nil
}

// DeleteUser 删除用户
// 只有 Owner 可以删除用户，不能删除自己和 Owner
//
// 参数:
//   - userid: 执行删除操作的用户ID（必须为管理员）
//   - id: 要删除的用户ID
//
// 返回:
//   - error: 删除过程中的错误信息
func (userService *UserService) DeleteUser(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	var deletedUser model.User
	err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		// 检查执行操作的用户是否为 Owner
		operator, err := userService.userRepository.GetUserByID(txCtx, userid)
		if err != nil {
			return err
		}
		if !operator.IsOwner {
			return errors.New(commonModel.ONLY_OWNER_CAN_MANAGE)
		}

		// 检查要删除的用户是否存在
		user, err := userService.userRepository.GetUserByID(txCtx, id)
		if err != nil {
			return err
		}

		if userid == user.ID || user.IsOwner {
			return errors.New(commonModel.INVALID_PARAMS_BODY)
		}

		deletedUser = user
		if err := userService.userRepository.DeleteUser(txCtx, id); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	deletedUser.Password = ""
	if err := userService.publisher.UserDeleted(
		context.Background(),
		contracts.UserDeletedEvent{User: deletedUser},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish user deleted event", zap.Error(err))
	}
	return nil
}

// GetUserByID 根据用户ID获取用户信息
//
// 参数:
//   - userId: 用户ID
//
// 返回:
//   - model.User: 用户信息
//   - error: 获取过程中的错误信息
func (userService *UserService) GetUserByID(userId string) (model.User, error) {
	return userService.userRepository.GetUserByID(context.Background(), userId)
}
