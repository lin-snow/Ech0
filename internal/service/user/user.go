// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/kvstore"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type UserService struct {
	transactor     transaction.Transactor
	userRepository Repository
	durableKV      kvstore.Store
	fileService    FileService
	bus            *busen.Bus
}

func NewUserService(
	tx transaction.Transactor,
	userRepository Repository,
	durableKV kvstore.Store,
	fileService FileService,
	busProvider func() *busen.Bus,
) *UserService {
	return &UserService{
		transactor:     tx,
		userRepository: userRepository,
		durableKV:      durableKV,
		fileService:    fileService,
		bus:            busProvider(),
	}
}

func ensurePasswordLength(password string) error {
	if len(password) > cryptoUtil.MaxPasswordBytes {
		return errors.New(commonModel.PASSWORD_TOO_LONG)
	}
	return nil
}

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

	ownerLocale := string(commonModel.DefaultLocale)
	if requested := strings.TrimSpace(registerDto.Locale); requested != "" {
		if resolved := i18nUtil.ResolveLocale(requested); resolved != "" {
			ownerLocale = resolved
		}
	}

	if err := ensurePasswordLength(registerDto.Password); err != nil {
		return err
	}
	passwordHash, err := cryptoUtil.HashPassword(registerDto.Password)
	if err != nil {
		return err
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

		existingUser, err := userService.userRepository.GetUserByUsername(ctx, registerDto.Username)
		if err == nil && existingUser.ID != model.USER_NOT_EXISTS_ID {
			return errors.New(commonModel.USERNAME_HAS_EXISTS)
		}

		owner = model.User{
			Username: registerDto.Username,
			Email:    email,
			IsAdmin:  true,
			Locale:   ownerLocale,
		}

		if err := userService.userRepository.CreateUser(ctx, &owner); err != nil {
			return err
		}
		if err := userService.userRepository.UpsertLocalAuth(ctx, &model.UserLocalAuth{
			UserID:       owner.ID,
			PasswordHash: passwordHash,
			PasswordAlgo: cryptoUtil.AlgoBcrypt,
		}); err != nil {
			return err
		}

		return userService.userRepository.MarkInitialized(ctx)
	}); err != nil {
		return err
	}

	eventbus.Notify(context.Background(), userService.bus, event.UserCreated{User: owner})

	return nil
}

func (userService *UserService) Register(registerDto *authModel.RegisterDto) error {
	initialized, err := userService.userRepository.IsInitialized(context.Background())
	if err != nil {
		return err
	}
	if !initialized {
		return commonModel.NewBizError(commonModel.ErrCodeInitInvalidState, commonModel.SIGNUP_FIRST)
	}

	users, err := userService.userRepository.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	if len(users) > authModel.MAX_USER_COUNT {
		return errors.New(commonModel.USER_COUNT_EXCEED_LIMIT)
	}

	if err := ensurePasswordLength(registerDto.Password); err != nil {
		return err
	}
	passwordHash, err := cryptoUtil.HashPassword(registerDto.Password)
	if err != nil {
		return err
	}
	email := strings.TrimSpace(registerDto.Email)
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errors.New("邮箱格式无效")
		}
	}

	newUser := model.User{
		Username: registerDto.Username,
		Email:    email,
		IsAdmin:  false,
		Locale:   string(commonModel.DefaultLocale),
	}

	user, err := userService.userRepository.GetUserByUsername(context.Background(), newUser.Username)
	if err == nil && user.ID != model.USER_NOT_EXISTS_ID {
		return errors.New(commonModel.USERNAME_HAS_EXISTS)
	}

	sysSetting, err := coreSetting.System.Get(context.Background(), userService.durableKV)
	if err != nil {
		return err
	}
	if !sysSetting.AllowRegister {
		return errors.New(commonModel.USER_REGISTER_NOT_ALLOW)
	}
	if err := userService.transactor.Run(context.Background(), func(ctx context.Context) error {
		if err := userService.userRepository.CreateUser(ctx, &newUser); err != nil {
			return err
		}
		return userService.userRepository.UpsertLocalAuth(ctx, &model.UserLocalAuth{
			UserID:       newUser.ID,
			PasswordHash: passwordHash,
			PasswordAlgo: cryptoUtil.AlgoBcrypt,
		})
	}); err != nil {
		return err
	}

	eventbus.Notify(context.Background(), userService.bus, event.UserCreated{User: newUser})

	return nil
}

func (userService *UserService) UpdateUser(ctx context.Context, userdto model.UserInfoDto) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return err
	}

	if userdto.Username != "" && userdto.Username != user.Username {
		existingUser, err := userService.userRepository.GetUserByUsername(ctx, userdto.Username)
		if err == nil && existingUser.ID != user.ID {
			return errors.New(commonModel.USERNAME_ALREADY_EXISTS)
		}
		user.Username = userdto.Username
	}

	var newPasswordHash string
	if userdto.Password != "" {
		if err := ensurePasswordLength(userdto.Password); err != nil {
			return err
		}
		hashed, err := cryptoUtil.HashPassword(userdto.Password)
		if err != nil {
			return err
		}
		newPasswordHash = hashed
	}

	avatarChanged := false
	if userdto.Avatar != "" && userdto.Avatar != user.Avatar {
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
		if err := userService.userRepository.UpdateUser(txCtx, &user); err != nil {
			return err
		}
		if newPasswordHash != "" {
			return userService.userRepository.UpsertLocalAuth(txCtx, &model.UserLocalAuth{
				UserID:       user.ID,
				PasswordHash: newPasswordHash,
				PasswordAlgo: cryptoUtil.AlgoBcrypt,
			})
		}
		return nil
	}); err != nil {
		return err
	}
	if avatarChanged && strings.TrimSpace(userdto.AvatarFileID) != "" {
		if err := userService.fileService.ConfirmTempFiles(ctx, []string{userdto.AvatarFileID}); err != nil {
			logUtil.GetLogger().Warn("confirm temp avatar file failed", logUtil.Err(err))
		}
	}

	eventbus.Notify(context.Background(), userService.bus, event.UserUpdated{User: user})

	return nil
}

func (userService *UserService) UpdateUserAdmin(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	operator, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return err
	}
	if !operator.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	user, err := userService.userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if userid == user.ID {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	if user.IsAdmin {
		admins, err := userService.countAdmins(ctx)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return lastAdminRequiredError()
		}
	}

	user.IsAdmin = !user.IsAdmin

	if err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		return userService.userRepository.UpdateUser(txCtx, &user)
	}); err != nil {
		return err
	}

	eventbus.Notify(context.Background(), userService.bus, event.UserUpdated{User: user})

	return nil
}

func (userService *UserService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	caller, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !caller.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	allUsers, err := userService.userRepository.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// The caller manages other accounts only; self-service lives in UpdateUser.
	users := make([]model.User, 0, len(allUsers))
	for _, u := range allUsers {
		if u.ID == caller.ID {
			continue
		}
		users = append(users, u)
	}

	return users, nil
}

func (userService *UserService) GetPrimaryAdmin() (model.User, error) {
	return userService.userRepository.GetPrimaryAdmin(context.Background())
}

func (userService *UserService) countAdmins(ctx context.Context) (int, error) {
	users, err := userService.userRepository.GetAllUsers(ctx)
	if err != nil {
		return 0, err
	}
	admins := 0
	for i := range users {
		if users[i].IsAdmin {
			admins++
		}
	}
	return admins, nil
}

func lastAdminRequiredError() error {
	return commonModel.NewBizErrorWithMessageKey(
		commonModel.ErrCodePermissionDenied,
		commonModel.LAST_ADMIN_REQUIRED,
		commonModel.MsgKeyUserLastAdminRequired,
		nil,
	)
}

func (userService *UserService) DeleteUser(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	var deletedUser model.User
	err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		operator, err := userService.userRepository.GetUserByID(txCtx, userid)
		if err != nil {
			return err
		}
		if !operator.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		user, err := userService.userRepository.GetUserByID(txCtx, id)
		if err != nil {
			return err
		}

		if userid == user.ID {
			return errors.New(commonModel.INVALID_PARAMS_BODY)
		}

		if user.IsAdmin {
			admins, err := userService.countAdmins(txCtx)
			if err != nil {
				return err
			}
			if admins <= 1 {
				return lastAdminRequiredError()
			}
		}

		deletedUser = user
		return userService.userRepository.DeleteUser(txCtx, id)
	})
	if err != nil {
		return err
	}

	eventbus.Notify(context.Background(), userService.bus, event.UserDeleted{User: deletedUser})
	return nil
}

func (userService *UserService) GetUserByID(userId string) (model.User, error) {
	return userService.userRepository.GetUserByID(context.Background(), userId)
}
