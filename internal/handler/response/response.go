// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/httperr"
)

// Response is what a raw gin handler produced. Raw routes (streams, uploads,
// downloads, redirects, MCP) cannot use Huma's typed outputs, so they report
// the same three things Huma does: a status, a payload, or an error. Execute
// renders the payload or the RFC 9457 problem document, which keeps every
// route on one error contract.
type Response struct {
	Status int
	Data   any
	Err    error
}

// Payload answers 200 with data.
func Payload(data any) Response { return Response{Data: data} }

// Created answers 201 with the created resource.
func Created(data any) Response { return Response{Status: http.StatusCreated, Data: data} }

// NoContent answers the status with an empty body.
func NoContent(status int) Response { return Response{Status: status} }

// Failed answers with the status the error itself carries.
func Failed(err error) Response { return Response{Err: err} }

// FailedWithStatus renders err with an explicit status, for callers whose
// status is a transport decision rather than a domain one.
func FailedWithStatus(status int, err error) Response {
	return Response{Status: status, Err: err}
}

// Execute runs fn and writes its answer.
func Execute(fn func(ctx *gin.Context) Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res := fn(ctx)
		if res.Err != nil {
			err := res.Err
			if res.Status > 0 {
				err = httperr.EdgeCause(res.Status, err)
			}
			httperr.Write(ctx, err)
			return
		}

		status := res.Status
		if status == 0 {
			status = http.StatusOK
		}
		if res.Data == nil {
			ctx.AbortWithStatus(status)
			return
		}
		ctx.JSON(status, res.Data)
	}
}
