// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

type Provider interface {
	Complete(ctx context.Context, req Request) (Response, error)
	Stream(ctx context.Context, req Request) (<-chan Event, error)
}

func providerFor(setting model.AgentSetting) (Provider, error) {
	switch setting.Protocol {
	case string(commonModel.OpenAI):
		return &openaiProvider{setting: setting}, nil
	case string(commonModel.OpenAIResponses):
		return &openaiResponsesProvider{setting: setting}, nil
	case string(commonModel.Anthropic):
		return &anthropicProvider{setting: setting}, nil
	default:
		return nil, errors.New(commonModel.AGENT_PROTOCOL_NOT_FOUND)
	}
}
