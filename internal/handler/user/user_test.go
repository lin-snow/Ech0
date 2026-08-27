// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"testing"

	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	usermock "github.com/lin-snow/ech0/internal/test/mocks/usermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func assertBizErr(t *testing.T, err error, wantCode string) {
	t.Helper()
	require.Error(t, err)
	var be *commonModel.BizError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, wantCode, be.Code)
}

func TestUserHandler_Register(t *testing.T) {
	t.Run("success passes dto through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().
			Register(mock.MatchedBy(func(dto *authModel.RegisterDto) bool {
				return dto != nil && dto.Username == "alice" && dto.Password == "pw"
			})).
			Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.Register(context.Background(), &userHandler.RegisterInput{
			Body: authModel.RegisterDto{Username: "alice", Password: "pw"},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.REGISTER_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("error when registration disallowed", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		be := commonModel.NewBizError(commonModel.ErrCodePermissionDenied, commonModel.USER_REGISTER_NOT_ALLOW)
		svc.EXPECT().Register(mock.Anything).Return(be).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.Register(context.Background(), &userHandler.RegisterInput{})

		assertBizErr(t, err, commonModel.ErrCodePermissionDenied)
		assert.Equal(t, 0, out.Code)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	t.Run("success passes dto + ctx through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		ctx := helpers.CtxAsUser("u1")
		svc.EXPECT().
			UpdateUser(ctx, mock.MatchedBy(func(dto userModel.UserInfoDto) bool {
				return dto.Username == "bob"
			})).
			Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.UpdateUser(ctx, &userHandler.UpdateUserInput{
			Body: userModel.UserInfoDto{Username: "bob"},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.UPDATE_USER_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(errors.New("db down")).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.UpdateUser(context.Background(), &userHandler.UpdateUserInput{})

		require.Error(t, err)
		assert.Equal(t, 0, out.Code)
	})
}

func TestUserHandler_UpdateUserAdmin(t *testing.T) {
	t.Run("success passes id through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().UpdateUserAdmin(mock.Anything, "uid-9").Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.UpdateUserAdmin(context.Background(), &userHandler.UpdateUserAdminInput{ID: "uid-9"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.UPDATE_USER_SUCCESS, out.Message)
	})

	t.Run("error when not owner", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		be := commonModel.NewBizError(commonModel.ErrCodePermissionDenied, commonModel.ONLY_OWNER_CAN_MANAGE)
		svc.EXPECT().UpdateUserAdmin(mock.Anything, mock.Anything).Return(be).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.UpdateUserAdmin(context.Background(), &userHandler.UpdateUserAdminInput{ID: "x"})

		assertBizErr(t, err, commonModel.ErrCodePermissionDenied)
		assert.Equal(t, 0, out.Code)
	})
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		want := []userModel.User{{ID: "u1", Username: "alice"}, {ID: "u2", Username: "bob"}}
		svc.EXPECT().GetAllUsers(mock.Anything).Return(want, nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetAllUsers(context.Background(), &userHandler.GetAllUsersInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_USER_SUCCESS, out.Message)
		assert.Equal(t, want, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().GetAllUsers(mock.Anything).Return(nil, bizErrInternal()).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetAllUsers(context.Background(), &userHandler.GetAllUsersInput{})

		assertBizErr(t, err, commonModel.ErrCodeInternal)
		assert.Nil(t, out.Data)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Run("success passes id through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().DeleteUser(mock.Anything, "uid-3").Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.DeleteUser(context.Background(), &userHandler.DeleteUserInput{ID: "uid-3"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DELETE_USER_SUCCESS, out.Message)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().DeleteUser(mock.Anything, mock.Anything).Return(errors.New("nope")).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.DeleteUser(context.Background(), &userHandler.DeleteUserInput{ID: "x"})

		require.Error(t, err)
		assert.Equal(t, 0, out.Code)
	})
}

func TestUserHandler_GetUserInfo(t *testing.T) {
	t.Run("success resolves viewer id", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().
			GetUserByID("viewer-1").
			Return(userModel.User{ID: "viewer-1", Username: "alice"}, nil).
			Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetUserInfo(helpers.CtxAsUser("viewer-1"), &userHandler.GetUserInfoInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_USER_INFO_SUCCESS, out.Message)
		assert.Equal(t, "alice", out.Data.Username)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		be := commonModel.NewBizError(commonModel.ErrCodeInternal, commonModel.USER_NOTFOUND)
		svc.EXPECT().GetUserByID("viewer-2").Return(userModel.User{}, be).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetUserInfo(helpers.CtxAsUser("viewer-2"), &userHandler.GetUserInfoInput{})

		assertBizErr(t, err, commonModel.ErrCodeInternal)
		assert.Equal(t, 0, out.Code)
	})
}

func bizErrInternal() *commonModel.BizError {
	return commonModel.NewBizError(commonModel.ErrCodeInternal, "boom")
}
