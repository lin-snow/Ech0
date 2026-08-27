// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"errors"
	"strings"
)

const (
	MsgKeyCommonSuccess               = "common.success"
	MsgKeyCommonRequestFailed         = "common.request_failed"
	MsgKeyCommonInvalidRequest        = "common.invalid_request"
	MsgKeyInvalidQueryParams          = "common.invalid_query_params"
	MsgKeySettingUpdateOK             = "setting.update_success"
	MsgKeyAgentModelMissing           = "agent.model_missing"
	MsgKeyEchoMixedFileCategories     = "echo.mixed_file_categories"
	MsgKeyAuthTokenMissing            = "auth.token_missing"
	MsgKeyAuthTokenInvalid            = "auth.token_invalid"
	MsgKeyAuthTokenParse              = "auth.token_parse_error"
	MsgKeyAuthScopeForbidden          = "auth.scope_forbidden"
	MsgKeyAuthAudienceForbidden       = "auth.audience_forbidden"
	MsgKeyAuthTokenTransportForbidden = "auth.token_transport_forbidden"
	MsgKeyAuthTokenRevoked            = "auth.token_revoked"
	MsgKeyAuthRefreshTokenInvalid     = "auth.refresh_token_invalid"
	MsgKeyAuthExchangeCodeInvalid     = "auth.exchange_code_invalid"
	MsgKeyAuthTokenGenerateFailed     = "auth.token_generate_failed"
	MsgKeyDashboardLogsOk             = "dashboard.logs.success"
	MsgKeyDashboardTailBad            = "dashboard.logs.tail_invalid"
	MsgKeyDashboardCheckUpdateFailed  = "dashboard.check_update_failed"
)

func MessageKeyFromErrorCode(code string) string {
	switch code {
	case ErrCodeInvalidQuery:
		return MsgKeyInvalidQueryParams
	case ErrCodeTokenMissing:
		return MsgKeyAuthTokenMissing
	case ErrCodeTokenInvalid:
		return MsgKeyAuthTokenInvalid
	case ErrCodeTokenParse:
		return MsgKeyAuthTokenParse
	case ErrCodeScopeForbidden:
		return MsgKeyAuthScopeForbidden
	case ErrCodeAudienceForbidden:
		return MsgKeyAuthAudienceForbidden
	case ErrCodeTokenTransportForbidden:
		return MsgKeyAuthTokenTransportForbidden
	case ErrCodeTokenRevoked:
		return MsgKeyAuthTokenRevoked
	case ErrCodeRefreshTokenInvalid:
		return MsgKeyAuthRefreshTokenInvalid
	case ErrCodeExchangeCodeInvalid:
		return MsgKeyAuthExchangeCodeInvalid
	case ErrCodeTokenGenerateFailed:
		return MsgKeyAuthTokenGenerateFailed
	default:
		return ""
	}
}

func MessageKeyFromMessage(msg string) string {
	switch msg {
	case SUCCESS_MESSAGE:
		return MsgKeyCommonSuccess
	case UPDATE_SETTINGS_SUCCESS:
		return MsgKeySettingUpdateOK
	case AGENT_MODEL_MISSING:
		return MsgKeyAgentModelMissing
	case ECHO_MIXED_FILE_CATEGORIES:
		return MsgKeyEchoMixedFileCategories
	default:
		return ""
	}
}

func ResolveFailureFields(err error, base string) (code, messageKey string, params map[string]any) {
	if bizErr, ok := errors.AsType[*BizError](err); ok {
		key := strings.TrimSpace(bizErr.MessageKey)
		if key == "" {
			key = MessageKeyFromErrorCode(bizErr.Code)
		}
		return bizErr.Code, key, bizErr.Params
	}
	return "", MessageKeyFromMessage(base), nil
}
