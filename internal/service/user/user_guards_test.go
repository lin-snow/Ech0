// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
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
	svc := userService.NewUserService(m.tx, m.repo, m.kv, m.file, func() *busen.Bus { return bus })
	return svc, m
}

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
			want: commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY,
		},
		{
			name: "empty password",
			dto:  authModel.RegisterDto{Username: "owner", Password: "", Email: "a@b.com"},
			want: commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY,
		},
		{
			name: "empty email",
			dto:  authModel.RegisterDto{Username: "owner", Password: "pw", Email: "  "},
			want: "邮箱不能为空",
		},
		{
			name: "invalid email",
			dto:  authModel.RegisterDto{Username: "owner", Password: "pw", Email: "not-an-email"},
			want: "邮箱格式无效",
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

	var be *commonModel.BizError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, commonModel.ErrCodeInitAlreadyDone, be.Code)
	assert.Equal(t, commonModel.SYSTEM_ALREADY_INITED, be.Msg)
}

func TestInitOwner_OwnerAlreadyExists(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(false, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).
		Return([]userModel.User{helpers.NewUser()}, nil).Once()

	err := svc.InitOwner(&authModel.RegisterDto{Username: "owner", Password: "pw", Email: "a@b.com"})

	var be *commonModel.BizError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, commonModel.ErrCodeInitOwnerExists, be.Code)
	assert.Equal(t, commonModel.OWNER_ALREADY_EXISTS, be.Msg)
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
	m.repo.EXPECT().MarkInitialized(mock.Anything).Return(nil).Once()

	err := svc.InitOwner(&authModel.RegisterDto{Username: "owner", Password: "secret", Email: " owner@ech0.com "})
	require.NoError(t, err)

	assert.True(t, created.IsOwner, "first user must be owner")
	assert.True(t, created.IsAdmin, "owner must be admin")
	assert.Equal(t, "owner@ech0.com", created.Email)
	assert.Equal(t, "owner-id", localAuth.UserID)
	assert.Equal(t, cryptoUtil.AlgoBcrypt, localAuth.PasswordAlgo)
	assert.True(t, cryptoUtil.CheckPassword(localAuth.PasswordAlgo, localAuth.PasswordHash, "secret"))
}

func TestRegister_NotInitialized(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(false, nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw"})

	var be *commonModel.BizError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, commonModel.ErrCodeInitInvalidState, be.Code)
	assert.Equal(t, commonModel.SIGNUP_FIRST, be.Msg)
}

func TestRegister_UserCountExceedLimit(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()

	over := make([]userModel.User, authModel.MAX_USER_COUNT+1)
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(over, nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw"})
	require.EqualError(t, err, commonModel.USER_COUNT_EXCEED_LIMIT)
}

func TestRegister_UsernameExists(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()
	m.repo.EXPECT().GetUserByUsername(mock.Anything, "dup").
		Return(helpers.NewUser(withID("u-existing")), nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "dup", Password: "pw"})
	require.EqualError(t, err, commonModel.USERNAME_HAS_EXISTS)
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().IsInitialized(mock.Anything).Return(true, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, nil).Once()

	err := svc.Register(&authModel.RegisterDto{Username: "u", Password: "pw", Email: "bad-email"})
	require.EqualError(t, err, "邮箱格式无效")
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
	require.EqualError(t, err, commonModel.USER_REGISTER_NOT_ALLOW)
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

	assert.False(t, created.IsAdmin, "registered user must not be admin")
	assert.False(t, created.IsOwner, "registered user must not be owner")
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
	require.EqualError(t, err, commonModel.PASSWORD_TOO_LONG)
}

func TestUpdateUser_PasswordTooLong(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()

	longPw := strings.Repeat("a", cryptoUtil.MaxPasswordBytes+1)
	err := svc.UpdateUser(helpers.CtxAsUser("admin-1"), userModel.UserInfoDto{Password: longPw})
	require.EqualError(t, err, commonModel.PASSWORD_TOO_LONG)
}

func TestUpdateUserAdmin_OperatorLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("db down")
	m.repo.EXPECT().GetUserByID(mock.Anything, "op").Return(userModel.User{}, sentinel).Once()

	err := svc.UpdateUserAdmin(helpers.CtxAsUser("op"), "target")
	require.ErrorIs(t, err, sentinel)
}

func TestUpdateUserAdmin_NotOwner(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()

	err := svc.UpdateUserAdmin(helpers.CtxAsUser("admin-1"), "u-2")
	require.EqualError(t, err, commonModel.ONLY_OWNER_CAN_MANAGE)
}

func TestUpdateUserAdmin_TargetLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("target gone")
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(userModel.User{}, sentinel).Once()

	err := svc.UpdateUserAdmin(helpers.CtxAsUser("owner-1"), "u-2")
	require.ErrorIs(t, err, sentinel)
}

func TestUpdateUserAdmin_CannotChangeSelf(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil)

	err := svc.UpdateUserAdmin(helpers.CtxAsUser("owner-1"), "owner-1")
	require.EqualError(t, err, commonModel.INVALID_PARAMS_BODY)
}

func TestUpdateUserAdmin_CannotChangeOwner(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-2").
		Return(helpers.NewUser(withID("owner-2"), helpers.AsOwner), nil).Once()

	err := svc.UpdateUserAdmin(helpers.CtxAsUser("owner-1"), "owner-2")
	require.EqualError(t, err, commonModel.INVALID_PARAMS_BODY)
}

func TestUpdateUserAdmin_Success(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()
	m.expectTxPassthrough()

	var updated userModel.User
	m.repo.EXPECT().UpdateUser(mock.Anything, mock.Anything).
		Run(func(_ context.Context, u *userModel.User) { updated = *u }).
		Return(nil).Once()

	err := svc.UpdateUserAdmin(helpers.CtxAsUser("owner-1"), "u-2")
	require.NoError(t, err)
	assert.True(t, updated.IsAdmin, "普通用户应被提升为 admin（取反）")
}

func TestDeleteUser_NotOwner(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("admin-1"), "u-2")
	require.EqualError(t, err, commonModel.ONLY_OWNER_CAN_MANAGE)
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil)

	err := svc.DeleteUser(helpers.CtxAsUser("owner-1"), "owner-1")
	require.EqualError(t, err, commonModel.INVALID_PARAMS_BODY)
}

func TestDeleteUser_CannotDeleteOwner(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-2").
		Return(helpers.NewUser(withID("owner-2"), helpers.AsOwner), nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("owner-1"), "owner-2")
	require.EqualError(t, err, commonModel.INVALID_PARAMS_BODY)
}

func TestDeleteUser_TargetLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	sentinel := errors.New("no such user")
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "ghost").
		Return(userModel.User{}, sentinel).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("owner-1"), "ghost")
	require.ErrorIs(t, err, sentinel)
}

func TestDeleteUser_Success(t *testing.T) {
	svc, m := newUserSvc(t)
	m.expectTxPassthrough()
	m.repo.EXPECT().GetUserByID(mock.Anything, "owner-1").
		Return(helpers.NewUser(withID("owner-1"), helpers.AsOwner), nil).Once()
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()
	m.repo.EXPECT().DeleteUser(mock.Anything, "u-2").Return(nil).Once()

	err := svc.DeleteUser(helpers.CtxAsUser("owner-1"), "u-2")
	require.NoError(t, err)
}

func TestGetAllUsers_CallerLookupError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("caller gone")
	m.repo.EXPECT().GetUserByID(mock.Anything, "x").Return(userModel.User{}, sentinel).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("x"))
	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, got)
}

func TestGetAllUsers_NotAdmin(t *testing.T) {
	svc, m := newUserSvc(t)
	m.repo.EXPECT().GetUserByID(mock.Anything, "u-2").
		Return(helpers.NewUser(withID("u-2")), nil).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("u-2"))
	require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
	assert.Nil(t, got)
}

func TestGetAllUsers_StripsOwner(t *testing.T) {
	svc, m := newUserSvc(t)

	owner := helpers.NewUser(withID("owner-1"), helpers.AsOwner)
	admin := helpers.NewUser(withID("admin-1"), helpers.AsAdmin)
	normal := helpers.NewUser(withID("u-2"))

	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").Return(admin, nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).
		Return([]userModel.User{owner, admin, normal}, nil).Once()
	m.repo.EXPECT().GetOwner(mock.Anything).Return(owner, nil).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("admin-1"))
	require.NoError(t, err)

	require.Len(t, got, 2, "owner 必须被剔除")
	for _, u := range got {
		assert.NotEqual(t, owner.ID, u.ID, "结果中不应出现 owner")
	}
}

func TestGetAllUsers_RepoError(t *testing.T) {
	svc, m := newUserSvc(t)
	sentinel := errors.New("list failed")
	m.repo.EXPECT().GetUserByID(mock.Anything, "admin-1").
		Return(helpers.NewUser(withID("admin-1"), helpers.AsAdmin), nil).Once()
	m.repo.EXPECT().GetAllUsers(mock.Anything).Return(nil, sentinel).Once()

	got, err := svc.GetAllUsers(helpers.CtxAsUser("admin-1"))
	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, got)
}
