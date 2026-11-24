// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package httperr renders every failed request as an RFC 9457 problem document.
// It is the only place that reads apperr.Classes, so a sentinel means the same
// status on the Huma surface, the raw gin routes and the middleware gates.
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
	"github.com/lin-snow/ech0/internal/apperr"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// Problem type URIs, re-exported so the wire contract has one import path.
const (
	TypeAccountDisabled = apperr.TypeAccountDisabled
	TypeDevicePending   = apperr.TypeDevicePending
	TypeDeviceSlowDown  = apperr.TypeDeviceSlowDown
	TypeDeviceExpired   = apperr.TypeDeviceExpired
	TypeDeviceDenied    = apperr.TypeDeviceDenied
)

// detailInternal is the only thing a client sees for a 5xx: driver strings,
// panics and stack fragments are not part of the contract.
const detailInternal = "internal server error"

// edgeError is a failure whose status was decided at the transport edge:
// Huma's own decode errors, a method-not-allowed line that names the path,
// an OAuth protocol error. Domain code returns apperr sentinels instead.
type edgeError struct {
	status int
	detail string
	cause  error
}

func (e *edgeError) Error() string {
	if e.detail != "" {
		return e.detail
	}
	if e.cause != nil {
		return e.cause.Error()
	}
	return http.StatusText(e.status)
}

func (e *edgeError) Unwrap() error { return e.cause }

// Edge is a transport failure with a fixed status and detail.
func Edge(status int, detail string) error {
	return &edgeError{status: status, detail: detail}
}

// EdgeCause forces status onto err. From redacts the text when status is 5xx.
// A nil err stays nil.
func EdgeCause(status int, err error) error {
	if err == nil {
		return nil
	}
	return &edgeError{status: status, cause: err}
}

// Code is the status From would write. It does not log.
func Code(err error) int {
	status, _, _, _ := project(err)
	return status
}

// From projects an error into the document Huma sends back. A classified
// sentinel carries its own status and detail; anything else is an unexpected
// failure, logged in full while the client only learns that the request failed.
func From(err error) *huma.ErrorModel {
	status, problemType, detail, logIt := project(err)
	if logIt {
		logUnexpected(err)
	}
	if status >= http.StatusInternalServerError {
		detail = detailInternal
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

func project(err error) (status int, problemType, detail string, logIt bool) {
	if err == nil {
		return http.StatusInternalServerError, "", detailInternal, false
	}

	var edge *edgeError
	if errors.As(err, &edge) {
		status = edge.status
		if status == 0 {
			status = http.StatusInternalServerError
		}
		detail = edge.detail
		if detail == "" && edge.cause != nil {
			detail = edge.cause.Error()
		}
		return status, "", detail, status >= http.StatusInternalServerError
	}

	if status, problemType, detail, ok := classify(err); ok {
		return status, problemType, detail, status >= http.StatusInternalServerError
	}
	return http.StatusInternalServerError, "", detailInternal, true
}

func classify(err error) (status int, problemType, detail string, ok bool) {
	for _, class := range apperr.Classes() {
		if errors.Is(err, class.Err) {
			return class.Status, class.Type, class.Err.Error(), true
		}
	}
	return 0, "", "", false
}

func logUnexpected(err error) {
	// Huma builds its default error documents with an empty message, so an
	// empty text carries no diagnosis and would only add noise to the log.
	if err == nil || err.Error() == "" {
		return
	}
	logUtil.GetLogger().Error("request failed", logUtil.Err(err))
}
