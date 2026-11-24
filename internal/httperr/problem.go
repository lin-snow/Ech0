// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package httperr renders every failed request as an RFC 9457 problem document.
// It is the only place that maps a Go error to an HTTP status, so the Huma
// surface, the raw gin routes and the middleware gates cannot drift apart.
//
// The document carries no application-level error code. The status is the
// behaviour a client branches on, `errors` locates request fields, and `type`
// is reserved for the rare failure a status alone cannot express. `detail` is
// developer-facing English; clients never parse it.
package httperr

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// detailInternal is the only thing a client sees for a 5xx: driver strings,
// panics and stack fragments are not part of the contract.
const detailInternal = "internal server error"

// From projects an error into the document Huma sends back. A StatusError
// carries its own status and detail; anything else is an unexpected failure,
// logged in full while the client only learns that the request failed.
func From(err error) *huma.ErrorModel {
	status, problemType, detail := http.StatusInternalServerError, "", detailInternal

	var se *commonModel.StatusError
	if errors.As(err, &se) && se.Status != 0 {
		status, problemType = se.Status, se.Type
		detail = se.Error()
		if status >= http.StatusInternalServerError {
			detail = detailInternal
			logUnexpected(se.Err)
		}
	} else {
		logUnexpected(err)
	}
	if detail == "" {
		detail = http.StatusText(status)
	}

	return &huma.ErrorModel{
		Type:   problemType,
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}
}

// Write renders the document. Raw gin routes use it; Huma renders its own.
func Write(ctx *gin.Context, err error) {
	problem := From(err)
	ctx.Header("Content-Type", "application/problem+json; charset=utf-8")
	ctx.JSON(problem.Status, problem)
}

// Abort stops the gin chain and renders the document.
func Abort(ctx *gin.Context, err error) {
	ctx.Abort()
	Write(ctx, err)
}

func logUnexpected(err error) {
	// Huma builds its default error documents with an empty message, so an
	// empty text carries no diagnosis and would only add noise to the log.
	if err == nil || err.Error() == "" {
		return
	}
	logUtil.GetLogger().Error("request failed", logUtil.Err(err))
}
