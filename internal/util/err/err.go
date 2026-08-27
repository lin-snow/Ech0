// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"errors"

	model "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

func HandleError(se *model.ServerError) string {
	if se.Err != nil {
		if se.Msg == "" {
			se.Msg = se.Err.Error()
		}
		logUtil.GetLogger().Error(se.Msg, logUtil.Err(se.Err))
	}

	return se.Msg
}

func ExtractBizErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if bizErr, ok := errors.AsType[*model.BizError](err); ok {
		return bizErr.Code
	}
	return ""
}

func HandlePanicError(se *model.ServerError) {
	if se.Err != nil {
		if se.Msg == "" {
			se.Msg = se.Err.Error()
		}
		logUtil.Panic(se.Msg, logUtil.Err(se.Err))
	}

	panic(se.Msg)
}
