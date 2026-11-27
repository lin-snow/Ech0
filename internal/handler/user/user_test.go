// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lin-snow/ech0/internal/apperr"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	"github.com/lin-snow/ech0/internal/httperr"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	usermock "github.com/lin-snow/ech0/internal/test/mocks/usermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func assertStatusErr(t *testing.T, err error, wantStatus int) {
	t.Helper()
	require.Error(t, err)
	assert.Equal(t, wantStatus, httperr.From(err).Status)
}

func TestUserHandler_Register(t *testing.T) {
	t.Run("success emits an empty output", func(t *testing.T) {
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
		require.NotNil(t, out)
	})

	t.Run("error when registration disallowed", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		be := apperr.ErrRegistrationDisabled
		svc.EXPECT().Register(mock.Anything).Return(be).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.Register(context.Background(), &userHandler.RegisterInput{})

		assertStatusErr(t, err, http.StatusForbidden)
		assert.Nil(t, out)
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
		require.NotNil(t, out)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(errors.New("db down")).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.UpdateUser(context.Background(), &userHandler.UpdateUserInput{})

		require.Error(t, err)
		assert.Nil(t, out)
	})
}

func TestUserHandler_UpdateUserRole(t *testing.T) {
	t.Run("success passes id and role through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().UpdateUserRole(mock.Anything, "uid-9", userModel.RoleAdmin).Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		in := &userHandler.UpdateUserRoleInput{ID: "uid-9"}
		in.Body.Role = userModel.RoleAdmin
		out, err := h.UpdateUserRole(context.Background(), in)

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("error when the operator cannot manage the target", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		be := apperr.ErrDenied
		svc.EXPECT().UpdateUserRole(mock.Anything, mock.Anything, mock.Anything).Return(be).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.UpdateUserRole(context.Background(), &userHandler.UpdateUserRoleInput{ID: "x"})

		assertStatusErr(t, err, http.StatusForbidden)
		assert.Nil(t, out)
	})
}

func TestUserHandler_TransferOwnership(t *testing.T) {
	t.Run("success passes the successor through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().TransferOwnership(mock.Anything, "uid-9").Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.TransferOwnership(context.Background(), &userHandler.TransferOwnershipInput{ID: "uid-9"})

		require.NoError(t, err)
		require.NotNil(t, out)
	})
}

func TestUserHandler_UpdateUserStatus(t *testing.T) {
	t.Run("success passes id and status through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().UpdateUserStatus(mock.Anything, "uid-4", userModel.StatusDisabled).Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		in := &userHandler.UpdateUserStatusInput{ID: "uid-4"}
		in.Body.Status = userModel.StatusDisabled
		out, err := h.UpdateUserStatus(context.Background(), in)

		require.NoError(t, err)
		require.NotNil(t, out)
	})
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	t.Run("success returns the list payload", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		want := []userModel.User{{ID: "u1", Username: "alice"}, {ID: "u2", Username: "bob"}}
		svc.EXPECT().GetAllUsers(mock.Anything).Return(want, nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetAllUsers(context.Background(), &userHandler.GetAllUsersInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body.Items)
		assert.Nil(t, out.Body.Total, "an unpaged list carries no total")
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().GetAllUsers(mock.Anything).Return(nil, internalErr()).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetAllUsers(context.Background(), &userHandler.GetAllUsersInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Run("success passes id through", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().DeleteUser(mock.Anything, "uid-3").Return(nil).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.DeleteUser(context.Background(), &userHandler.DeleteUserInput{ID: "uid-3"})

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		svc.EXPECT().DeleteUser(mock.Anything, mock.Anything).Return(errors.New("nope")).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.DeleteUser(context.Background(), &userHandler.DeleteUserInput{ID: "x"})

		require.Error(t, err)
		assert.Nil(t, out)
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
		require.NotNil(t, out)
		assert.Equal(t, "alice", out.Body.Username)
	})

	t.Run("error", func(t *testing.T) {
		svc := usermock.NewMockService(t)
		be := apperr.ErrUserNotFound
		svc.EXPECT().GetUserByID("viewer-2").Return(userModel.User{}, be).Once()

		h := userHandler.NewUserHandler(svc)
		out, err := h.GetUserInfo(helpers.CtxAsUser("viewer-2"), &userHandler.GetUserInfoInput{})

		assertStatusErr(t, err, http.StatusNotFound)
		assert.Nil(t, out)
	})
}

func internalErr() error {
	return errors.New("boom")
}
