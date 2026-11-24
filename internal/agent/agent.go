// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"

	"github.com/lin-snow/ech0/internal/apperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

const (
	GEN_RECENT = "gen_recent"
)

func validate(setting model.AgentSetting) error {
	if !setting.Enable {
		return apperr.ErrAgentDisabled
	}
	if setting.Model == "" {
		return apperr.ErrAgentModel
	}
	if setting.Protocol == "" {
		return apperr.ErrAgentProtocol
	}
	if setting.ApiKey == "" && !allowsEmptyAPIKey(setting.Protocol) {
		return apperr.ErrAgentKey
	}
	return nil
}

func allowsEmptyAPIKey(protocol string) bool {
	return protocol == string(commonModel.OpenAI) || protocol == string(commonModel.OpenAIResponses)
}

func applyPrompt(setting model.AgentSetting, in []Message, usePrompt bool) []Message {
	if setting.Prompt != "" && usePrompt {
		in = append(in, Message{Role: RoleUser, Content: setting.Prompt})
	}
	return in
}

func Generate(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	usePrompt bool,
) (string, error) {
	if err := validate(setting); err != nil {
		return "", err
	}

	provider, err := providerFor(setting)
	if err != nil {
		return "", err
	}

	resp, err := provider.Complete(ctx, Request{
		Messages: applyPrompt(setting, in, usePrompt),
	})
	if err != nil {
		return "", err
	}
	return stripReasoning(resp.Text), nil
}

func Ping(ctx context.Context, setting model.AgentSetting) error {
	if setting.Model == "" {
		return apperr.ErrAgentModel
	}
	if setting.Protocol == "" {
		return apperr.ErrAgentProtocol
	}
	if setting.ApiKey == "" && !allowsEmptyAPIKey(setting.Protocol) {
		return apperr.ErrAgentKey
	}

	provider, err := providerFor(setting)
	if err != nil {
		return err
	}

	_, err = provider.Complete(ctx, Request{
		Messages: []Message{{Role: RoleUser, Content: "ping"}},
	})
	return err
}
