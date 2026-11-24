// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/kvstore"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	model "github.com/lin-snow/ech0/internal/model/user"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type UserService struct {
	transactor      transaction.Transactor
	userRepository  Repository
	durableKV       kvstore.Store
	fileService     FileService
	echoCache       EchoCacheInvalidator
	embeddingPurger EmbeddingPurger
	bus             *busen.Bus
}

func NewUserService(
	tx transaction.Transactor,
	userRepository Repository,
	durableKV kvstore.Store,
	fileService FileService,
	echoCache EchoCacheInvalidator,
	embeddingPurger EmbeddingPurger,
	busProvider func() *busen.Bus,
) *UserService {
	return &UserService{
		transactor:      tx,
		userRepository:  userRepository,
		durableKV:       durableKV,
		fileService:     fileService,
		echoCache:       echoCache,
		embeddingPurger: embeddingPurger,
		bus:             busProvider(),
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
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email address")
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
			return commonModel.Fail(http.StatusBadRequest, commonModel.SYSTEM_ALREADY_INITED)
		}

		users, err := userService.userRepository.GetAllUsers(ctx)
		if err != nil {
			return err
		}
		if len(users) > 0 {
			return commonModel.Fail(http.StatusBadRequest, commonModel.OWNER_ALREADY_EXISTS)
		}

		existingUser, err := userService.userRepository.GetUserByUsername(ctx, registerDto.Username)
		if err == nil && existingUser.ID != model.USER_NOT_EXISTS_ID {
			return errors.New(commonModel.USERNAME_HAS_EXISTS)
		}

		owner = model.User{
			Username: registerDto.Username,
			Email:    email,
			Role:     model.RoleOwner,
			Locale:   ownerLocale,
		}

		if err := userService.userRepository.CreateUser(ctx, &owner); err != nil {
			return err
		}
		return userService.userRepository.UpsertLocalAuth(ctx, &model.UserLocalAuth{
			UserID:       owner.ID,
			PasswordHash: passwordHash,
			PasswordAlgo: cryptoUtil.AlgoBcrypt,
		})
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
		return commonModel.Fail(http.StatusBadRequest, commonModel.SIGNUP_FIRST)
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
			return errors.New("invalid email address")
		}
	}

	newUser := model.User{
		Username: registerDto.Username,
		Email:    email,
		Role:     model.RoleMember,
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
	oldAvatarFileID := user.AvatarFileID
	newAvatarFileID := strings.TrimSpace(userdto.AvatarFileID)
	if newAvatarFileID != "" {
		// Get is the authorization point: an avatar must be a file the caller
		// may read, which a stranger's private upload is not.
		avatarDTO, err := userService.fileService.Get(ctx, newAvatarFileID)
		if err != nil {
			return err
		}
		user.Avatar = avatarDTO.URL
		user.AvatarFileID = newAvatarFileID
		avatarChanged = avatarDTO.URL != ""
	} else if userdto.Avatar != "" && userdto.Avatar != user.Avatar {
		user.Avatar = userdto.Avatar
		user.AvatarFileID = ""
		avatarChanged = true
	}
	if userdto.Locale != "" {
		user.Locale = i18nUtil.ResolveLocale(userdto.Locale)
	}
	if strings.TrimSpace(userdto.Email) != "" {
		if _, err := mail.ParseAddress(strings.TrimSpace(userdto.Email)); err != nil {
			return errors.New("invalid email address")
		}
		user.Email = strings.TrimSpace(userdto.Email)
	}
	if err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		if err := userService.userRepository.UpdateUser(txCtx, &user); err != nil {
			return err
		}
		if newPasswordHash != "" {
			if err := userService.userRepository.UpsertLocalAuth(txCtx, &model.UserLocalAuth{
				UserID:       user.ID,
				PasswordHash: newPasswordHash,
				PasswordAlgo: cryptoUtil.AlgoBcrypt,
			}); err != nil {
				return err
			}
		}
		if newAvatarFileID == "" {
			return nil
		}
		// The avatar slot holds one file: replacing it drops the old link in
		// the same statement group that adds the new one.
		return userService.fileService.Replace(
			txCtx, []string{newAvatarFileID},
			fileModel.OwnerKindUser, user.ID, fileModel.RoleAvatar,
		)
	}); err != nil {
		return err
	}
	if avatarChanged && oldAvatarFileID != "" && oldAvatarFileID != newAvatarFileID {
		// The replaced avatar is the uploader's own file with no other reader;
		// a file someone else still links refuses and stays.
		if _, err := userService.fileService.Delete(ctx, oldAvatarFileID, false); err != nil {
			var conflict *fileService.ConflictError
			if !errors.As(err, &conflict) {
				logUtil.GetLogger().Warn("delete replaced avatar file failed", logUtil.Err(err))
			}
		}
	}

	eventbus.Notify(context.Background(), userService.bus, event.UserUpdated{User: user})

	return nil
}

// UpdateUserRole promotes a member to admin or demotes an admin to member.
// Assigning a role requires standing above it, so only the owner can do either.
// The owner's row is not a valid target: ownership changes only through
// TransferOwnership.
func (userService *UserService) UpdateUserRole(ctx context.Context, id string, newRole model.Role) error {
	if newRole != model.RoleAdmin && newRole != model.RoleMember {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	target, err := userService.userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if target.Role == newRole {
		return nil
	}
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.UserManage,
		authz.ForUser(target, newRole)); err != nil {
		return err
	}

	if err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		return userService.userRepository.UpdateUserRole(txCtx, id, newRole)
	}); err != nil {
		return err
	}

	target.Role = newRole
	eventbus.Notify(context.Background(), userService.bus, event.UserUpdated{User: target})

	return nil
}

// TransferOwnership hands the space to another account. It is the one
// operation not derived from the rank rule: the current owner names an active admin
// as successor and keeps acting as an admin afterwards.
func (userService *UserService) TransferOwnership(ctx context.Context, newOwnerID string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	operator, err := userService.userRepository.GetUserByID(ctx, userid)
	if err != nil {
		return err
	}
	if operator.ID == newOwnerID {
		return authz.ErrDenied
	}
	if err := authz.Decide(viewer.NewSession(operator.ID, operator.Role, authz.ScopesForRole(operator.Role)), authz.UserTransfer,
		authz.Instance()); err != nil {
		return err
	}
	target, err := userService.userRepository.GetUserByID(ctx, newOwnerID)
	if err != nil {
		return err
	}
	if target.Role != model.RoleAdmin || !target.IsActive() {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	if err := userService.userRepository.TransferOwnership(ctx, operator.ID, target.ID); err != nil {
		return err
	}

	target.Role = model.RoleOwner
	operator.Role = model.RoleAdmin
	eventbus.Notify(context.Background(), userService.bus, event.UserUpdated{User: target})
	eventbus.Notify(context.Background(), userService.bus, event.UserUpdated{User: operator})

	return nil
}

func (userService *UserService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.UserList, authz.Instance()); err != nil {
		return nil, err
	}
	caller, err := userService.userRepository.GetUserByID(ctx, viewer.MustFromContext(ctx).UserID())
	if err != nil {
		return nil, err
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

func (userService *UserService) GetOwner() (model.User, error) {
	return userService.userRepository.GetOwner(context.Background())
}

// UpdateUserStatus disables or restores an account. It cannot reach an account
// of equal or higher standing, which is why an administrator cannot lock out
// another administrator and nobody can lock out the owner.
func (userService *UserService) UpdateUserStatus(ctx context.Context, id string, status model.Status) error {
	if !status.Valid() {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}
	userid := viewer.MustFromContext(ctx).UserID()
	if userid == id {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	target, err := userService.userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.UserManage,
		authz.ForUser(target, "")); err != nil {
		return err
	}

	return userService.userRepository.UpdateUserStatus(ctx, id, status)
}

// DeleteUser permanently removes a user and everything addressed to them. It
// is the only deletion path: there is no soft-delete state to restore from.
func (userService *UserService) DeleteUser(ctx context.Context, id string) error {
	var deletedUser model.User
	var purge model.PurgeResult
	err := userService.transactor.Run(ctx, func(txCtx context.Context) error {
		user, err := userService.userRepository.GetUserByID(txCtx, id)
		if err != nil {
			return err
		}

		if err := authz.Decide(viewer.MustFromContext(txCtx), authz.UserManage,
			authz.ForUser(user, "")); err != nil {
			return err
		}

		deletedUser = user
		purge, err = userService.userRepository.PurgeUserContent(txCtx, id)
		if err != nil {
			return err
		}
		// Only the files nobody links to any more go away with the account:
		// an image someone else attached to their echo keeps its link, and its row
		// with it.
		if _, err := userService.fileService.PurgeUploader(txCtx, id); err != nil {
			return err
		}
		for _, echoID := range purge.EchoIDs {
			if err := userService.embeddingPurger.Delete(txCtx, echoID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	userService.echoCache.InvalidateEchoCaches(purge.EchoIDs...)

	eventbus.Notify(context.Background(), userService.bus, event.UserDeleted{User: deletedUser})
	for _, echoID := range purge.EchoIDs {
		eventbus.Notify(context.Background(), userService.bus, event.EchoDeleted{Echo: echoModel.Echo{ID: echoID}})
	}
	return nil
}

func (userService *UserService) GetUserByID(userId string) (model.User, error) {
	return userService.userRepository.GetUserByID(context.Background(), userId)
}
