// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
)

type ErrorBody struct {
	Code          int            `json:"code" doc:"业务状态码，失败为 0"`
	Message       string         `json:"msg" doc:"状态描述（回退文案）"`
	ErrorCode     string         `json:"error_code,omitempty" doc:"稳定的业务错误码"`
	MessageKey    string         `json:"message_key,omitempty" doc:"i18n 翻译 key"`
	MessageParams map[string]any `json:"message_params,omitempty" doc:"i18n 模板参数"`
	Data          any            `json:"data" doc:"失败时为 null"`
}

type apiError struct {
	status int
	body   commonModel.Result[any]
}

func (e *apiError) Error() string  { return e.body.Message }
func (e *apiError) GetStatus() int { return e.status }

func (e *apiError) MarshalJSON() ([]byte, error) { return json.Marshal(e.body) }

func (e *apiError) Schema(r huma.Registry) *huma.Schema {
	return r.Schema(reflect.TypeFor[ErrorBody](), true, "ErrorBody")
}

func Err(ctx context.Context, err error) error {
	base := errUtil.HandleError(&commonModel.ServerError{Err: err})
	code, key, params := commonModel.ResolveFailureFields(err, base)

	if code == "" && key == "" {
		return &apiError{status: http.StatusBadRequest, body: commonModel.Fail[any](base)}
	}

	msg := i18nUtil.Localize(localizerFrom(ctx), key, base, params)
	return &apiError{
		status: http.StatusBadRequest,
		body:   commonModel.FailWithLocalized[any](msg, code, key, params),
	}
}

var installErrorModelOnce sync.Once

func frameworkErrorFields(status int) (code, messageKey string) {
	if status >= http.StatusInternalServerError {
		return commonModel.ErrCodeInternal, commonModel.MsgKeyCommonRequestFailed
	}
	return commonModel.ErrCodeInvalidRequest, commonModel.MsgKeyCommonInvalidRequest
}

func installErrorModel() {
	installErrorModelOnce.Do(func() {
		huma.NewErrorWithContext = func(hctx huma.Context, status int, msg string, errs ...error) huma.StatusError {
			code, key := frameworkErrorFields(status)
			loc := i18nUtil.LocalizerFromGin(humagin.Unwrap(hctx))
			localized := i18nUtil.Localize(loc, key, msg, nil) + detailSuffix(errs)
			return &apiError{
				status: status,
				body:   commonModel.FailWithLocalized[any](localized, code, key, nil),
			}
		}
		huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
			code, key := frameworkErrorFields(status)
			return &apiError{
				status: status,
				body:   commonModel.FailWithLocalized[any](msg+detailSuffix(errs), code, key, nil),
			}
		}
	})
}

func detailSuffix(errs []error) string {
	details := make([]string, 0, len(errs))
	for _, e := range errs {
		if e != nil {
			details = append(details, e.Error())
		}
	}
	if len(details) == 0 {
		return ""
	}
	return " (" + strings.Join(details, "; ") + ")"
}
