// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	initModel "github.com/lin-snow/ech0/internal/model/init"
	service "github.com/lin-snow/ech0/internal/service/init"
)

type InitHandler struct {
	initService service.Service
}

func NewInitHandler(initService service.Service) *InitHandler {
	return &InitHandler{initService: initService}
}

type (
	GetInitStatusInput struct{}
	InitOwnerInput     struct {
		Body authModel.RegisterDto
	}
)

type (
	StatusOutput = commonModel.Result[initModel.Status]
	EmptyOutput  = commonModel.Result[any]
)

func (h *InitHandler) GetInitStatus(ctx context.Context, _ *GetInitStatusInput) (StatusOutput, error) {
	status, err := h.initService.GetStatus()
	if err != nil {
		return StatusOutput{}, err
	}
	return commonModel.OK(status), nil
}

func (h *InitHandler) InitOwner(ctx context.Context, in *InitOwnerInput) (EmptyOutput, error) {
	if err := h.initService.InitOwner(&in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.INIT_OWNER_SUCCESS), nil
}
