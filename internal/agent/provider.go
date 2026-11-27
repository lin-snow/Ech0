// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"errors"

	"github.com/lin-snow/ech0/internal/apperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

// errContentFiltered is a response the provider's content filter stopped
// partway. What was streamed is not an answer, and passing it off as one would
// hide why it ends where it does.
var errContentFiltered = errors.New("模型输出被服务端内容过滤拦截，回答未完成（content_filter）")

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
		return nil, apperr.ErrAgentProtocol
	}
}
