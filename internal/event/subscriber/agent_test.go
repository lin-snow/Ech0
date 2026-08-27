// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package subscriber_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/event/subscriber"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validAgentJSON = `{"enable":true,"protocol":"openai","model":"gpt-4o-mini"}`

func TestAgentProcessor_Handlers(t *testing.T) {
	invoke := map[string]func(*subscriber.AgentProcessor, context.Context) error{
		"echo.created": func(ap *subscriber.AgentProcessor, ctx context.Context) error {
			return ap.HandleEchoCreated(ctx, event.EchoCreated{Echo: helpers.NewEcho()})
		},
		"echo.updated": func(ap *subscriber.AgentProcessor, ctx context.Context) error {
			return ap.HandleEchoUpdated(ctx, event.EchoUpdated{Echo: helpers.NewEcho()})
		},
		"user.deleted": func(ap *subscriber.AgentProcessor, ctx context.Context) error {
			return ap.HandleUserDeleted(ctx, event.UserDeleted{User: helpers.NewUser()})
		},
	}

	for name, call := range invoke {
		t.Run(name+"/valid setting clears gen cache", func(t *testing.T) {
			kv := kvmock.NewMockStore(t)
			kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return(validAgentJSON, nil).Once()

			var deletedKey string
			kv.EXPECT().Delete(mock.Anything, agent.GEN_RECENT).
				Run(func(_ context.Context, key string) { deletedKey = key }).
				Return(nil).Once()

			ap := subscriber.NewAgentProcessor(kv)
			require.NoError(t, call(ap, helpers.CtxAnonymous()))
			assert.Equal(t, agent.GEN_RECENT, deletedKey)
		})
	}
}

func TestAgentProcessor_GetError(t *testing.T) {
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return("", errBoom).Once()

	ap := subscriber.NewAgentProcessor(kv)
	err := ap.HandleEchoCreated(helpers.CtxAnonymous(), event.EchoCreated{Echo: helpers.NewEcho()})
	require.ErrorIs(t, err, errBoom)
}

func TestAgentProcessor_InvalidJSON(t *testing.T) {
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return("{not-json", nil).Once()

	ap := subscriber.NewAgentProcessor(kv)
	err := ap.HandleEchoUpdated(helpers.CtxAnonymous(), event.EchoUpdated{Echo: helpers.NewEcho()})
	require.Error(t, err)
}

func TestAgentProcessor_NotFoundStillClears(t *testing.T) {
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return("", kvstore.ErrNotFound).Once()
	kv.EXPECT().Delete(mock.Anything, agent.GEN_RECENT).Return(nil).Once()

	ap := subscriber.NewAgentProcessor(kv)
	require.NoError(t, ap.HandleUserDeleted(helpers.CtxAnonymous(), event.UserDeleted{User: helpers.NewUser()}))
}

func TestAgentProcessor_Registrations(t *testing.T) {
	kv := kvmock.NewMockStore(t)
	ap := subscriber.NewAgentProcessor(kv)
	regs := ap.Registrations()
	require.Len(t, regs, 3)
	for i, r := range regs {
		assert.NotNil(t, r, "registration %d should be non-nil", i)
	}
}

func TestAgentProcessor_DeleteError(t *testing.T) {
	delErr := errors.New("delete failed")
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return(validAgentJSON, nil).Once()
	kv.EXPECT().Delete(mock.Anything, agent.GEN_RECENT).Return(delErr).Once()

	ap := subscriber.NewAgentProcessor(kv)
	err := ap.HandleEchoCreated(helpers.CtxAnonymous(), event.EchoCreated{Echo: helpers.NewEcho()})
	require.ErrorIs(t, err, delErr)
}
