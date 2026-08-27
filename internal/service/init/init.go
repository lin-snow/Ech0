// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	initModel "github.com/lin-snow/ech0/internal/model/init"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type InitService struct {
	repository     Repository
	userService    UserService
	settingService SettingService
}

func NewInitService(repository Repository, userService UserService, settingService SettingService) *InitService {
	return &InitService{
		repository:     repository,
		userService:    userService,
		settingService: settingService,
	}
}

func (s *InitService) GetStatus() (initModel.Status, error) {
	initialized, err := s.repository.IsInitialized()
	if err != nil {
		return initModel.Status{}, err
	}

	_, ownerErr := s.repository.GetOwner()
	ownerExists := ownerErr == nil

	return initModel.Status{
		Initialized: initialized,
		OwnerExists: ownerExists,
	}, nil
}

func (s *InitService) InitOwner(registerDto *authModel.RegisterDto) error {
	initialized, err := s.repository.IsInitialized()
	if err != nil {
		return err
	}
	if initialized {
		return commonModel.NewBizError(commonModel.ErrCodeInitAlreadyDone, commonModel.SYSTEM_ALREADY_INITED)
	}
	if err := s.userService.InitOwner(registerDto); err != nil {
		return err
	}

	if err := s.settingService.BootstrapDefaultLocale(context.Background(), registerDto.Locale); err != nil {
		logUtil.GetLogger().Warn("Failed to bootstrap default locale on init owner", logUtil.Err(err))
	}
	return nil
}
