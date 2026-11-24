// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
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
	v, err := coreSetting.Agent.Get(context.Background(), settingService.durableKV)
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
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	v, err := coreSetting.Agent.Get(ctx, settingService.durableKV)
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
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
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
		Temperature:   newSetting.Temperature,
	}
	if !validTemperature(setting.Temperature) {
		return apperr.ErrInvalidParams
	}
	return coreSetting.Agent.Set(ctx, settingService.durableKV, setting)
}

func (settingService *SettingService) TestAgentConnection(
	ctx context.Context,
	newSetting *model.AgentSettingDto,
) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	setting := model.AgentSetting{
		Enable:   true,
		Protocol: normalizeAgentProtocol(newSetting.Protocol),
		Model:    newSetting.Model,
		ApiKey:   newSetting.ApiKey,
		BaseURL:  urlUtil.TrimURL(newSetting.BaseURL),
		// Tested with the temperature it will run with, so a value the model
		// rejects fails here rather than in the first chat.
		Temperature: newSetting.Temperature,
	}
	if !validTemperature(setting.Temperature) {
		return apperr.ErrInvalidParams
	}

	ctx, cancel := context.WithTimeout(ctx, agentTestTimeout)
	defer cancel()
	return agent.Ping(ctx, setting)
}

// validTemperature accepts an unset temperature or one in the range every
// provider documents (Anthropic stops at 1, and rejects above that itself).
func validTemperature(t *float64) bool {
	return t == nil || (*t >= 0 && *t <= 2)
}
