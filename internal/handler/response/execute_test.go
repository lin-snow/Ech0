// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	i18n "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func runExecute(t *testing.T, res Response) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(i18n.ContextLocalizerKey, i18n.NewLocalizer("en-US", ""))
	Execute(func(*gin.Context) Response { return res })(c)
	return rec
}

func TestExecute_Success(t *testing.T) {
	cases := []struct {
		name        string
		res         Response
		wantCode    int
		wantMessage string
		wantData    string
	}{
		{
			name:        "code-zero-plain-data-empty-msg",
			res:         Response{Code: 0, Data: "payload", Msg: ""},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "",
			wantData:    "payload",
		},
		{
			name:        "code-zero-unmapped-msg-passthrough",
			res:         Response{Code: 0, Data: "payload", Msg: "custom hello"},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "custom hello",
			wantData:    "payload",
		},
		{
			name:        "code-zero-known-msg-localized",
			res:         Response{Code: 0, Data: "payload", Msg: commonModel.SUCCESS_MESSAGE},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "Request succeeded",
			wantData:    "payload",
		},
		{
			name:        "code-zero-explicit-message-key-localized",
			res:         Response{Code: 0, Data: "payload", Msg: "raw", MessageKey: commonModel.MsgKeyCommonSuccess},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "Request succeeded",
			wantData:    "payload",
		},
		{
			name:        "nonzero-code-takes-okwithcode",
			res:         Response{Code: 7, Data: "payload", Msg: "custom"},
			wantCode:    7,
			wantMessage: "custom",
			wantData:    "payload",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := runExecute(t, tc.res)

			require.Equal(t, http.StatusOK, rec.Code)
			got := helpers.ParseResult(t, rec)
			assert.Equal(t, tc.wantCode, got.Code)
			assert.Equal(t, tc.wantMessage, got.Msg)
			assert.Empty(t, got.ErrorCode)
			assert.Empty(t, got.MessageKey)

			var data string
			helpers.DecodeData(t, got.Data, &data)
			assert.Equal(t, tc.wantData, data)
		})
	}
}

func TestExecute_Failure(t *testing.T) {
	cases := []struct {
		name           string
		res            Response
		wantErrorCode  string
		wantMessageKey string
		wantMessage    string
	}{
		{
			name: "bizerror-carries-code-and-key",
			res: Response{
				Err: commonModel.NewBizErrorWithMessageKey(
					commonModel.ErrCodeInvalidQuery, "raw msg",
					commonModel.MsgKeyInvalidQueryParams, nil,
				),
			},
			wantErrorCode:  commonModel.ErrCodeInvalidQuery,
			wantMessageKey: commonModel.MsgKeyInvalidQueryParams,
			wantMessage:    "Invalid query parameters",
		},
		{
			name: "plain-error-msg-text-maps-to-key",
			res: Response{
				Err: errors.New("boom"),
				Msg: commonModel.AGENT_MODEL_MISSING,
			},
			wantErrorCode:  "",
			wantMessageKey: commonModel.MsgKeyAgentModelMissing,
			wantMessage:    "Agent model name is not configured or is empty",
		},
		{
			name: "explicit-errorcode-fallback-derives-key",
			res: Response{
				Err:       errors.New("boom"),
				Msg:       "totally unmapped text",
				ErrorCode: commonModel.ErrCodeInvalidQuery,
			},
			wantErrorCode:  commonModel.ErrCodeInvalidQuery,
			wantMessageKey: commonModel.MsgKeyInvalidQueryParams,
			wantMessage:    "Invalid query parameters",
		},
		{
			name: "no-code-no-key-plain-fail",
			res: Response{
				Err: errors.New("boom"),
				Msg: "totally unmapped text",
			},
			wantErrorCode:  "",
			wantMessageKey: "",
			wantMessage:    "totally unmapped text",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := runExecute(t, tc.res)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			got := helpers.ParseResult(t, rec)
			assert.Equal(t, commonModel.DEFAULT_FAILED_CODE, got.Code)
			assert.Equal(t, tc.wantErrorCode, got.ErrorCode)
			assert.Equal(t, tc.wantMessageKey, got.MessageKey)
			assert.Equal(t, tc.wantMessage, got.Msg)
		})
	}
}

func TestExecute_Failure_NoLocalizerFallsBackToDefaultText(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	Execute(func(*gin.Context) Response {
		return Response{
			Err: commonModel.NewBizErrorWithMessageKey(
				commonModel.ErrCodeInvalidQuery, "raw msg",
				commonModel.MsgKeyInvalidQueryParams, nil,
			),
		}
	})(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	got := helpers.ParseResult(t, rec)
	assert.Equal(t, commonModel.ErrCodeInvalidQuery, got.ErrorCode)
	assert.Equal(t, commonModel.MsgKeyInvalidQueryParams, got.MessageKey)
	assert.Equal(t, "raw msg", got.Msg)
}
