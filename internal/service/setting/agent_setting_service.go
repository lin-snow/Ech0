// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const agentTestTimeout = 15 * time.Second

func normalizeAgentProtocol(protocol string) string {
	switch protocol {
	case string(commonModel.OpenAI), string(commonModel.OpenAIResponses), string(commonModel.Anthropic):
		return protocol
	default:
		return string(commonModel.OpenAI)
	}
}

func (settingService *SettingService) GetAgentInfo(setting *model.AgentSetting) error {
	v, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.Agent)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

func (settingService *SettingService) GetAgentSettings(
	ctx context.Context,
	setting *model.AgentSetting,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.Agent)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

func (settingService *SettingService) UpdateAgentSettings(
	ctx context.Context,
	newSetting *model.AgentSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	setting := model.AgentSetting{
		Enable:        newSetting.Enable,
		Protocol:      normalizeAgentProtocol(newSetting.Protocol),
		Model:         newSetting.Model,
		ApiKey:        newSetting.ApiKey,
		Prompt:        newSetting.Prompt,
		BaseURL:       urlUtil.TrimURL(newSetting.BaseURL),
		Multimodal:    newSetting.Multimodal,
		ContextWindow: max(0, newSetting.ContextWindow),
	}
	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.Agent, setting)
}

func (settingService *SettingService) TestAgentConnection(
	ctx context.Context,
	newSetting *model.AgentSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	setting := model.AgentSetting{
		Enable:   true,
		Protocol: normalizeAgentProtocol(newSetting.Protocol),
		Model:    newSetting.Model,
		ApiKey:   newSetting.ApiKey,
		BaseURL:  urlUtil.TrimURL(newSetting.BaseURL),
	}

	ctx, cancel := context.WithTimeout(ctx, agentTestTimeout)
	defer cancel()
	return agent.Ping(ctx, setting)
}
