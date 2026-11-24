// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

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
		Body initModel.InitOwnerDto
	}
)

type (
	StatusOutput struct{ Body initModel.Status }
	EmptyOutput  = struct{}
)

func (h *InitHandler) GetInitStatus(_ context.Context, _ *GetInitStatusInput) (*StatusOutput, error) {
	status, err := h.initService.GetStatus()
	if err != nil {
		return nil, err
	}
	return &StatusOutput{Body: status}, nil
}

func (h *InitHandler) InitOwner(_ context.Context, in *InitOwnerInput) (*EmptyOutput, error) {
	if err := h.initService.InitOwner(&in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}
