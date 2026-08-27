// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/database"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
)

var readOnlySafeMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
}

func WriteGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !database.IsWriteLocked() {
			c.Next()
			return
		}

		if _, ok := readOnlySafeMethods[c.Request.Method]; ok {
			c.Next()
			return
		}

		c.Header("Retry-After", "30")
		c.AbortWithStatusJSON(
			http.StatusServiceUnavailable,
			commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
				Msg: "服务维护中，暂时不可写入",
				Err: nil,
			})),
		)
	}
}
