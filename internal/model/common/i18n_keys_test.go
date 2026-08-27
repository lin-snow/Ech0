// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageKeyFromErrorCode(t *testing.T) {
	cases := []struct {
		name string
		code string
		want string
	}{
		{"invalid_query", ErrCodeInvalidQuery, MsgKeyInvalidQueryParams},
		{"token_missing", ErrCodeTokenMissing, MsgKeyAuthTokenMissing},
		{"token_invalid", ErrCodeTokenInvalid, MsgKeyAuthTokenInvalid},
		{"token_parse", ErrCodeTokenParse, MsgKeyAuthTokenParse},
		{"scope_forbidden", ErrCodeScopeForbidden, MsgKeyAuthScopeForbidden},
		{"audience_forbidden", ErrCodeAudienceForbidden, MsgKeyAuthAudienceForbidden},
		{"token_transport_forbidden", ErrCodeTokenTransportForbidden, MsgKeyAuthTokenTransportForbidden},
		{"token_revoked", ErrCodeTokenRevoked, MsgKeyAuthTokenRevoked},
		{"refresh_token_invalid", ErrCodeRefreshTokenInvalid, MsgKeyAuthRefreshTokenInvalid},
		{"exchange_code_invalid", ErrCodeExchangeCodeInvalid, MsgKeyAuthExchangeCodeInvalid},
		{"token_generate_failed", ErrCodeTokenGenerateFailed, MsgKeyAuthTokenGenerateFailed},
		{"internal_unmapped", ErrCodeInternal, ""},
		{"permission_denied_unmapped", ErrCodePermissionDenied, ""},
		{"invalid_request_unmapped", ErrCodeInvalidRequest, ""},
		{"unknown_code", "TOTALLY_UNKNOWN", ""},
		{"empty_code", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MessageKeyFromErrorCode(tc.code))
		})
	}
}

func TestMessageKeyFromMessage(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want string
	}{
		{"success", SUCCESS_MESSAGE, MsgKeyCommonSuccess},
		{"update_settings", UPDATE_SETTINGS_SUCCESS, MsgKeySettingUpdateOK},
		{"agent_model_missing", AGENT_MODEL_MISSING, MsgKeyAgentModelMissing},
		{"unmapped_message", "some random failure text", ""},
		{"empty_message", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MessageKeyFromMessage(tc.msg))
		})
	}
}

func TestResolveFailureFields(t *testing.T) {
	params := map[string]any{"name": "x", "n": 3}

	t.Run("biz_error_with_explicit_message_key_wins", func(t *testing.T) {
		err := NewBizErrorWithMessageKey(ErrCodeTokenMissing, "boom", "custom.key", params)
		code, key, gotParams := ResolveFailureFields(err, "ignored base")
		assert.Equal(t, ErrCodeTokenMissing, code)
		assert.Equal(t, "custom.key", key)
		assert.Equal(t, params, gotParams)
	})

	t.Run("biz_error_empty_key_falls_back_to_code_mapping", func(t *testing.T) {
		err := NewBizError(ErrCodeScopeForbidden, "no scope")
		code, key, gotParams := ResolveFailureFields(err, "ignored base")
		assert.Equal(t, ErrCodeScopeForbidden, code)
		assert.Equal(t, MsgKeyAuthScopeForbidden, key)
		assert.Nil(t, gotParams)
	})

	t.Run("biz_error_whitespace_key_is_trimmed_then_mapped", func(t *testing.T) {
		err := NewBizErrorWithMessageKey(ErrCodeTokenRevoked, "revoked", "   ", params)
		code, key, gotParams := ResolveFailureFields(err, "ignored base")
		assert.Equal(t, ErrCodeTokenRevoked, code)
		assert.Equal(t, MsgKeyAuthTokenRevoked, key)
		assert.Equal(t, params, gotParams)
	})

	t.Run("biz_error_unmapped_code_yields_empty_key", func(t *testing.T) {
		err := NewBizError(ErrCodeInternal, "internal")
		code, key, gotParams := ResolveFailureFields(err, "ignored base")
		assert.Equal(t, ErrCodeInternal, code)
		assert.Equal(t, "", key)
		assert.Nil(t, gotParams)
	})

	t.Run("wrapped_biz_error_is_unwrapped_via_errors_as", func(t *testing.T) {
		inner := NewBizErrorWithMessageKey(ErrCodeAudienceForbidden, "aud", "", params)
		wrapped := fmt.Errorf("layer: %w", inner)
		code, key, gotParams := ResolveFailureFields(wrapped, "ignored base")
		assert.Equal(t, ErrCodeAudienceForbidden, code)
		assert.Equal(t, MsgKeyAuthAudienceForbidden, key)
		assert.Equal(t, params, gotParams)
	})

	t.Run("plain_error_maps_message_key_from_base", func(t *testing.T) {
		code, key, gotParams := ResolveFailureFields(errors.New("whatever"), SUCCESS_MESSAGE)
		assert.Equal(t, "", code)
		assert.Equal(t, MsgKeyCommonSuccess, key)
		assert.Nil(t, gotParams)
	})

	t.Run("plain_error_unmapped_base_yields_empty_key", func(t *testing.T) {
		code, key, gotParams := ResolveFailureFields(errors.New("whatever"), "unmapped base text")
		assert.Equal(t, "", code)
		assert.Equal(t, "", key)
		assert.Nil(t, gotParams)
	})
}
