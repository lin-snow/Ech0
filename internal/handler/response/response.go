// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
)

type Response struct {
	Code int `json:"code"`

	Data any `json:"data,omitempty"`

	Msg string `json:"msg"`

	ErrorCode string `json:"error_code,omitempty"`

	MessageKey string `json:"message_key,omitempty"`

	MessageParams map[string]any `json:"message_params,omitempty"`

	Err error `json:"-"`
}

// retryAfterSeconds renders the retry_after message parameter as a header
// value, so raw gin responses keep the same contract as middleware refusals.
func retryAfterSeconds(params map[string]any) string {
	value, ok := params["retry_after"]
	if !ok {
		return ""
	}
	switch seconds := value.(type) {
	case int:
		if seconds > 0 {
			return strconv.Itoa(seconds)
		}
	case int64:
		if seconds > 0 {
			return strconv.FormatInt(seconds, 10)
		}
	case float64:
		if seconds >= 1 {
			return strconv.Itoa(int(seconds))
		}
	}
	return ""
}

func Execute(fn func(ctx *gin.Context) Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res := fn(ctx)
		if res.Err != nil {
			msg := errorUtil.HandleError(&commonModel.ServerError{
				Msg: res.Msg,
				Err: res.Err,
			})
			localizer := i18nUtil.LocalizerFromGin(ctx)

			code, messageKey, params := commonModel.ResolveFailureFields(res.Err, msg)

			if code == "" && res.ErrorCode != "" {
				code = res.ErrorCode
				messageKey = strings.TrimSpace(res.MessageKey)
				if messageKey == "" {
					messageKey = commonModel.MessageKeyFromErrorCode(res.ErrorCode)
				}
				params = res.MessageParams
			}

			if code == "" && messageKey == "" {
				ctx.JSON(http.StatusBadRequest, commonModel.Fail[string](msg))
				return
			}

			msg = i18nUtil.Localize(localizer, messageKey, msg, params)
			status := http.StatusBadRequest
			switch code {
			case commonModel.ErrCodeRateLimited:
				status = http.StatusTooManyRequests
				if seconds := retryAfterSeconds(params); seconds != "" {
					ctx.Header("Retry-After", seconds)
				}
			case commonModel.ErrCodeAccountDisabled:
				status = http.StatusForbidden
			}
			ctx.JSON(status, commonModel.FailWithLocalized[string](msg, code, messageKey, params))
			return
		}

		successMsg := res.Msg
		messageKey := strings.TrimSpace(res.MessageKey)
		if messageKey == "" {
			messageKey = commonModel.MessageKeyFromMessage(res.Msg)
		}
		if messageKey != "" {
			successMsg = i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), messageKey, res.Msg, res.MessageParams)
		}

		if res.Code != 0 {
			ctx.JSON(http.StatusOK, commonModel.OKWithCode(res.Data, res.Code, successMsg))
		} else {
			ctx.JSON(http.StatusOK, commonModel.OK(res.Data, successMsg))
		}
	}
}
