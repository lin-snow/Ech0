// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/filemock"
	"github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/internal/test/mocks/usermock"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type userMocks struct {
	repo *usermock.MockRepository
	tx   *txmock.MockTransactor
	kv   *kvmock.MockStore
	file *filemock.MockService
}

func newUserSvc(t *testing.T) (*userService.UserService, *userMocks) {
	t.Helper()
	m := &userMocks{
		repo: usermock.NewMockRepository(t),
		tx:   txmock.NewMockTransactor(t),
		kv:   kvmock.NewMockStore(t),
		file: filemock.NewMockService(t),
	}
	bus := busen.New()
	svc := userService.NewUserService(m.tx, m.repo, m.kv, m.file, noopEchoCache{}, noopPurger{}, func() *busen.Bus { return bus })
	return svc, m
}

type noopEchoCache struct{}

func (noopEchoCache) InvalidateEchoCaches(...string) {}

type noopPurger struct{}

func (noopPurger) Delete(context.Context, string) error { return nil }

func (m *userMocks) expectTxPassthrough() {
	m.tx.EXPECT().
		Run(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func withID(id string) func(*userModel.User) {
	return func(u *userModel.User) { u.ID = id }
}

func mustMarshalSystem(t *testing.T, s settingModel.SystemSetting) string {
	t.Helper()
	raw, err := json.Marshal(s)
	require.NoError(t, err)
	return string(raw)
}

func TestInitOwner_InputGuards(t *testing.T) {
	cases := []struct {
		name string
		dto  authModel.RegisterDto
		want string
	}{
		{
			name: "empty username",
			dto:  authModel.RegisterDto{Username: "", Password: "pw", Email: "a@b.com"},
			want: apperr.ErrCredentialsEmpty.Error(),
		},
		{
			name: "empty password",
			dto:  authModel.RegisterDto{Username: "owner", Password: "", Email: "a@b.com"},
			want: apperr.ErrCredentialsEmpty.Error(),
		},
		{
			name: "empty email",
			dto:  authModel.RegisterDto{Username: "owner", Password: "pw", Email: "  "},
			want: "email is required",
		},
		{
			name: "invalid email",
			dto:  authModel.RegisterDto{Username: "owner", Password: "pw", Email: "not-an-email"},
			want: "invalid email address",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newUserSvc(t)
			err := svc.InitOwner(&tc.dto)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestInitOwner_AlreadyInitialized(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()

	err := svc.InitOwner(&authModel.RegisterDto{Username: "owner", Password: "pw", Email: "a@b.com"})

	require.ErrorIs(t, err, apperr.ErrAlreadyInitialized)
}

func TestInitOwner_OwnerAlreadyExists(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(false, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).
		Return([]userModel.User{helpers.NewUser()}, nil).Once()

	err := svc.InitOwner(&authModel.RegisterDto{Username: "owner", Password: "pw", Email: "a@b.com"})

	require.ErrorIs(t, err, apperr.ErrOwnerExists)
}

func TestInitOwner_Success(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(false, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()
	m.repo.EXPECT().GetUserByUsername(mock.Anything, "owner").
		Return(userModel.User{}, errors.New("not found")).Once()

	var created userModel.User
	m.repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
		Run(func(_ context.Context, u *userModel.User) { u.ID = "owner-id"; created = *u }).
		Return(nil).Once()
	var localAuth userModel.UserLocalAuth
	m.repo.EXPECT().UpsertLocalAuth(mock.Anything, mock.Anything).
		Run(func(_ context.Context, la *userModel.UserLocalAuth) { localAuth = *la }).
		Return(nil).Once()

	err := svc.InitOwner(&authModel.RegisterDto{Username: "owner", Password: "secret", Email: " owner@ech0.com "})
	require.NoError(t, err)

	assert.Equal(t, userModel.RoleOwner, created.Role, "初始化者必须成为空间所有者")
	assert.Equal(t, "owner@ech0.com", created.Email)
	assert.Equal(t, "owner-id", localAuth.UserID)
	assert.Equal(t, cryptoUtil.AlgoBcrypt, localAuth.PasswordAlgo)
	assert.True(t, cryptoUtil.CheckPassword(localAuth.PasswordAlgo, localAuth.PasswordHash, "secret"))
}

func TestRegister_NotInitialized(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(false, nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw"})

	require.ErrorIs(t, err, apperr.ErrSignupFirst)
}

func TestRegister_UserCountExceedLimit(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()

	over := make([]userModel.User, authModel.MAX_USER_COUNT+1)
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(over, nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw"})
	require.EqualError(t, err, apperr.ErrUserLimit.Error())
}

func TestRegister_UsernameExists(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()
	m.repo.EXPECT().GetUserByUsername(mock.Anything, "dup").
		Return(helpers.NewUser(withID("u-existing")), nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "dup", Password: "pw"})
	require.EqualError(t, err, apperr.ErrUsernameExists.Error())
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw", Email: "bad-email"})
	require.EqualError(t, err, "invalid email address")
}

func TestRegister_RegisterNotAllowed(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()
	m.repo.EXPECT().GetUserByUsername(mock.Anything, "u").
		Return(userModel.User{}, errors.New("not found")).Once()
	m.kv.EXPECT().Get(mock.Anything, mock.Anything).
		Return(mustMarshalSystem(t, settingModel.SystemSetting{AllowRegister: false}), nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw"})
	require.EqualError(t, err, apperr.ErrRegistrationDisabled.Error())
}

func TestRegister_Success(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()
	m.repo.EXPECT().GetUserByUsername(mock.Anything, "newbie").
		Return(userModel.User{}, errors.New("not found")).Once()
	m.kv.EXPECT().Get(mock.Anything, mock.Anything).
		Return(mustMarshalSystem(t, settingModel.SystemSetting{AllowRegister: true}), nil).Once()
	m.expectTxPassthrough()

	var created userModel.User
	m.repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
		Run(func(_ context.Context, u *userModel.User) { u.ID = "newbie-id"; created = *u }).
		Return(nil).Once()
	var localAuth userModel.UserLocalAuth
	m.repo.EXPECT().UpsertLocalAuth(mock.Anything, mock.Anything).
		Run(func(_ context.Context, la *userModel.UserLocalAuth) { localAuth = *la }).
		Return(nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "newbie", Password: "pw"})
	require.NoError(t, err)

	assert.Equal(t, userModel.RoleMember, created.Role, "注册用户默认是成员")
	assert.Equal(t, "newbie-id", localAuth.UserID)
	assert.Equal(t, cryptoUtil.AlgoBcrypt, localAuth.PasswordAlgo)
	assert.True(t, cryptoUtil.CheckPassword(localAuth.PasswordAlgo, localAuth.PasswordHash, "pw"))
}

func TestRegister_PasswordTooLong(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()

	longPw := strings.Repeat("a", cryptoUtil.MaxPasswordBytes+1)
	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: longPw})
	require.EqualError(t, err, apperr.ErrPasswordTooLong.Error())
}

func TestUpdateUser_PasswordTooLong(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()

	longPw := strings.Repeat("a", cryptoUtil.MaxPasswordBytes+1)
	err := svc.UpdateUser(helpers.CtxAsUser("admin-1"), userModel.UserInfoDto{Password: longPw})
	require.EqualError(t, err, apperr.ErrPasswordTooLong.Error())
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	svc, _ := newUserSvc(t)

	err := svc.UpdateUserRole(helpers.CtxAsUser("owner-1"), "u-2", userModel.RoleOwner)
	require.EqualError(t, err, apperr.ErrInvalidParams.Error())
}

func TestUpdateUserRole_MemberCannotAssignRoles(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()

	err := svc.UpdateUserRole(helpers.CtxAsUser("u-1"), "u-2", userModel.RoleAdmin)
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestUpdateUserRole_AdminCannotAssignRoles(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()

	err := svc.UpdateUserRole(helpers.CtxAsUser("admin-1", role.Admin), "u-2", userModel.RoleAdmin)
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestUpdateUserRole_TargetLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("target gone")
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(userModel.User{}, sentinel).Once()

	err := svc.UpdateUserRole(helpers.CtxAsUser("owner-1", role.Owner), "u-2", userModel.RoleAdmin)
	require.ErrorIs(t, err, sentinel)
}

func TestUpdateUserRole_OwnerIsNotAssignable(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()

	err := svc.UpdateUserRole(helpers.CtxAsUser("owner-1", role.Owner), "owner-1", userModel.RoleMember)
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestUpdateUserRole_SameRoleIsNoop(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()

	require.NoError(t, svc.UpdateUserRole(helpers.CtxAsUser("owner-1", role.Owner), "u-2", userModel.RoleMember))
}

func TestUpdateUserRole_OwnerPromotesMember(t *testing.T) {
	svc, m := newUserSvc(t)
	target := helpers.NewUser(withID("u-2"))
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").Return(target, nil).Once()
	m.expectTxPassthrough()
	m.repo.EXPECT().UpdateUserRole(mock.Anything, "u-2", userModel.RoleAdmin).Return(nil).Once()

	require.NoError(t, svc.UpdateUserRole(helpers.CtxAsUser("owner-1", role.Owner), "u-2", userModel.RoleAdmin))
}

func TestUpdateUserRole_OwnerDemotesAdmin(t *testing.T) {
	svc, m := newUserSvc(t)
	target := helpers.NewUser(withID("admin-2"), helpers.AsAdmin)
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-2").Return(target, nil).Once()
	m.expectTxPassthrough()
	m.repo.EXPECT().UpdateUserRole(mock.Anything, "admin-2", userModel.RoleMember).Return(nil).Once()

	require.NoError(t, svc.UpdateUserRole(helpers.CtxAsUser("owner-1", role.Owner), "admin-2", userModel.RoleMember))
}

func TestTransferOwnership_NonOwnerIsDenied(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()

	err := svc.TransferOwnership(helpers.CtxAsUser("admin-1", role.Admin), "admin-2")
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestTransferOwnership_CannotTransferToSelf(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()

	err := svc.TransferOwnership(helpers.CtxAsUser("owner-1", role.Owner), "owner-1")
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestTransferOwnership_RequiresActiveAdminTarget(t *testing.T) {
	cases := []struct {
		name   string
		target userModel.User
	}{
		{name: "member", target: helpers.NewUser(withID("u-2"))},
		{name: "disabled admin", target: func() userModel.User {
			u := helpers.NewUser(withID("admin-2"), helpers.AsAdmin)
			u.Status = userModel.StatusDisabled
			return u
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newUserSvc(t)
			m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
				Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
			m.repo.EXPECT().GetUserByID(mock.Anything, tc.target.ID).Return(tc.target, nil).Once()

			err := svc.TransferOwnership(helpers.CtxAsUser("owner-1", role.Owner), tc.target.ID)
			require.EqualError(t, err, apperr.ErrInvalidParams.Error())
		})
	}
}

func TestTransferOwnership_Success(t *testing.T) {
	svc, m := newUserSvc(t)
	owner := helpers.NewUser(withID("owner-1"), helpers.AsOwner)
	successor := helpers.NewUser(withID("admin-2"), helpers.AsAdmin)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").Return(owner, nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-2").Return(successor, nil).Once()
	m.repo.EXPECT().TransferOwnership(mock.Anything, "owner-1", "admin-2").Return(nil).Once()

	require.NoError(t, svc.TransferOwnership(helpers.CtxAsUser("owner-1", role.Owner), "admin-2"))
}

func TestUpdateUserStatus_InvalidStatus(t *testing.T) {
	svc, _ := newUserSvc(t)

	err := svc.UpdateUserStatus(helpers.CtxAsUser("admin-1", role.Admin), "u-2", userModel.Status("BROKEN"))
	require.EqualError(t, err, apperr.ErrInvalidParams.Error())
}

func TestUpdateUserStatus_CannotTargetSelf(t *testing.T) {
	svc, _ := newUserSvc(t)

	err := svc.UpdateUserStatus(helpers.CtxAsUser("admin-1", role.Admin), "admin-1", userModel.StatusDisabled)
	require.EqualError(t, err, apperr.ErrInvalidParams.Error())
}

func TestUpdateUserStatus_CannotDisableOwner(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()

	err := svc.UpdateUserStatus(helpers.CtxAsUser("admin-1", role.Admin), "owner-1", userModel.StatusDisabled)
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestUpdateUserStatus_AdminDisablesMember(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()
	m.repo.EXPECT().UpdateUserStatus(mock.Anything, "u-2", userModel.StatusDisabled).Return(nil).Once()

	require.NoError(t, svc.UpdateUserStatus(helpers.CtxAsUser("admin-1", role.Admin), "u-2", userModel.StatusDisabled))
}

func TestDeleteUser_NotAdmin(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("u-1"), "u-2")
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("admin-1", role.Admin), "admin-1")
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestDeleteUser_CannotDeletePeerAdmin(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	target := helpers.NewUser(withID("admin-2"), helpers.AsAdmin)
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-2").Return(target, nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("admin-1", role.Admin), "admin-2")
	require.ErrorIs(t, err, authz.ErrDenied)
}

func TestDeleteUser_TargetLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	sentinel := errors.New("no such user")
	m.repo.EXPECT().GetUserByID(mock.Anything, "ghost").
		Return(userModel.User{}, sentinel).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("admin-1", role.Admin), "ghost")
	require.ErrorIs(t, err, sentinel)
}

func TestDeleteUser_Success(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()
	m.repo.EXPECT().PurgeUserContent(mock.Anything, "u-2").
		Return(userModel.PurgeResult{}, nil).Once()
	m.file.EXPECT().PurgeUploader(mock.Anything, "u-2").Return(0, nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("admin-1", role.Admin), "u-2")
	require.NoError(t, err)
}

func TestGetAllUsers_CallerLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("caller gone")
	m.repo.EXPECT().GetUserByID(mock.Anything, "x").Return(userModel.User{}, sentinel).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("x", role.Admin))
	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, got)
}

func TestGetAllUsers_NotAdmin(t *testing.T) {
	svc, _ := newUserSvc(t)

	got, err := svc.GetAllUsers(helpers.CtxAsUser("u-2"))
	require.ErrorIs(t, err, authz.ErrDenied)
	assert.Nil(t, got)
}

func TestGetAllUsers_ExcludesCaller(t *testing.T) {
	svc, m := newUserSvc(t)

	otherAdmin := helpers.NewUser(withID("admin-2"), helpers.AsAdmin)
	caller := helpers.NewUser(withID("admin-1"), helpers.AsAdmin)
	normal := helpers.NewUser(withID("u-2"))

	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").Return(caller, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).
		Return([]userModel.User{otherAdmin, caller, normal}, nil).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("admin-1", role.Admin))
	require.NoError(t, err)

	require.Len(t, got, 2, "调用者自身必须被剔除")
	for _, u := range got {
		assert.NotEqual(t, caller.ID, u.ID, "结果中不应出现调用者")
	}
}

func TestGetAllUsers_RepoError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("list failed")
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, sentinel).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("admin-1", role.Admin))
	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, got)
}
