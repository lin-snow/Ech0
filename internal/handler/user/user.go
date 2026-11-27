// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	service "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type UserHandler struct {
	userService service.Service
}

func NewUserHandler(userService service.Service) *UserHandler {
	return &UserHandler{userService: userService}
}

type (
	RegisterInput struct {
		Body authModel.RegisterDto
	}
	UpdateUserInput struct {
		Body model.UserInfoDto
	}
	UpdateUserRoleInput struct {
		ID   string `path:"id" format:"uuid" doc:"用户 ID（UUID）"`
		Body struct {
			Role model.Role `json:"role" doc:"目标角色：admin 或 member"`
		}
	}
	TransferOwnershipInput struct {
		ID string `path:"id" format:"uuid" doc:"接任所有者的用户 ID（UUID）"`
	}
	GetAllUsersInput struct{}
	DeleteUserInput  struct {
		ID string `path:"id" format:"uuid" doc:"用户 ID（UUID）"`
	}
	UpdateUserStatusInput struct {
		ID   string `path:"id" format:"uuid" doc:"用户 ID（UUID）"`
		Body struct {
			Status model.Status `json:"status" doc:"账户状态：ACTIVE 或 DISABLED"`
		}
	}
	GetUserInfoInput struct{}
)

type (
	UserListOutput = struct{ Body commonModel.List[model.User] }
	UserOutput     = struct{ Body model.User }
	EmptyOutput    = struct{}
)

func (userHandler *UserHandler) Register(ctx context.Context, in *RegisterInput) (*EmptyOutput, error) {
	if err := userHandler.userService.Register(&in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (userHandler *UserHandler) UpdateUser(ctx context.Context, in *UpdateUserInput) (*EmptyOutput, error) {
	if err := userHandler.userService.UpdateUser(ctx, in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (userHandler *UserHandler) UpdateUserRole(ctx context.Context, in *UpdateUserRoleInput) (*EmptyOutput, error) {
	if err := userHandler.userService.UpdateUserRole(ctx, in.ID, in.Body.Role); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (userHandler *UserHandler) TransferOwnership(ctx context.Context, in *TransferOwnershipInput) (*EmptyOutput, error) {
	if err := userHandler.userService.TransferOwnership(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (userHandler *UserHandler) UpdateUserStatus(ctx context.Context, in *UpdateUserStatusInput) (*EmptyOutput, error) {
	if err := userHandler.userService.UpdateUserStatus(ctx, in.ID, in.Body.Status); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (userHandler *UserHandler) GetAllUsers(ctx context.Context, _ *GetAllUsersInput) (*UserListOutput, error) {
	users, err := userHandler.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return &UserListOutput{Body: commonModel.List[model.User]{Items: users}}, nil
}

func (userHandler *UserHandler) DeleteUser(ctx context.Context, in *DeleteUserInput) (*EmptyOutput, error) {
	if err := userHandler.userService.DeleteUser(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (userHandler *UserHandler) GetUserInfo(ctx context.Context, _ *GetUserInfoInput) (*UserOutput, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := userHandler.userService.GetUserByID(userid)
	if err != nil {
		return nil, err
	}
	return &UserOutput{Body: user}, nil
}
