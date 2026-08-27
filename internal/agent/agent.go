// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

const (
	GEN_RECENT = "gen_recent"
)

func validate(setting model.AgentSetting) error {
	if !setting.Enable {
		return errors.New(commonModel.AGENT_NOT_ENABLED)
	}
	if setting.Model == "" {
		return errors.New(commonModel.AGENT_MODEL_MISSING)
	}
	if setting.Protocol == "" {
		return errors.New(commonModel.AGENT_PROTOCOL_NOT_FOUND)
	}
	if setting.ApiKey == "" && !allowsEmptyAPIKey(setting.Protocol) {
		return errors.New(commonModel.AGENT_API_KEY_MISSING)
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
	temperature *float32,
) (string, error) {
	if err := validate(setting); err != nil {
		return "", err
	}

	provider, err := providerFor(setting)
	if err != nil {
		return "", err
	}

	resp, err := provider.Complete(ctx, Request{
		Messages:    applyPrompt(setting, in, usePrompt),
		Temperature: temperature,
	})
	if err != nil {
		return "", err
	}
	return stripReasoning(resp.Text), nil
}

func Ping(ctx context.Context, setting model.AgentSetting) error {
	if setting.Model == "" {
		return errors.New(commonModel.AGENT_MODEL_MISSING)
	}
	if setting.Protocol == "" {
		return errors.New(commonModel.AGENT_PROTOCOL_NOT_FOUND)
	}
	if setting.ApiKey == "" && !allowsEmptyAPIKey(setting.Protocol) {
		return errors.New(commonModel.AGENT_API_KEY_MISSING)
	}

	provider, err := providerFor(setting)
	if err != nil {
		return err
	}

	_, err = provider.Complete(ctx, Request{
		Messages:  []Message{{Role: RoleUser, Content: "ping"}},
		MaxTokens: 16,
	})
	return err
}
