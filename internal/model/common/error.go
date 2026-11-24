// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"fmt"
	"net/http"
)

// StatusError is the only failure type: the status the transport renders, the
// RFC 9457 type clients branch on when a status is not enough, a
// developer-facing English message, and the cause.
//
// The detail shown to clients is Msg when set and the cause otherwise; 5xx
// never exposes either. It unwraps, so errors.Is and errors.As see through it.
type StatusError struct {
	Status int
	Type   string
	Msg    string
	Err    error
}

func (e *StatusError) Error() string {
	if e == nil {
		return ""
	}
	switch {
	case e.Msg != "" && e.Err != nil:
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	case e.Msg != "":
		return e.Msg
	case e.Err != nil:
		return e.Err.Error()
	default:
		return http.StatusText(e.Status)
	}
}

func (e *StatusError) Unwrap() error { return e.Err }

// Fail builds a failure with the status and the message the client reads.
func Fail(status int, msg string) error {
	return &StatusError{Status: status, Msg: msg}
}

// FailWithType is Fail plus the RFC 9457 type URI. Use it only when a client
// must tell this failure apart from every other one with the same status.
func FailWithType(status int, problemType, msg string) error {
	return &StatusError{Status: status, Type: problemType, Msg: msg}
}

// WithStatus attaches a status to an existing error and keeps the cause, so
// the transport renders the cause's message while errors.Is still works. A nil
// err stays nil: the caller used it as "no error".
func WithStatus(status int, err error) error {
	if err == nil {
		return nil
	}
	return &StatusError{Status: status, Err: err}
}

// Problem type URIs. A client must tell these failures apart from every other
// one carrying the same status, so the URIs are part of the error contract.
const (
	TypeAccountDisabled = "https://ech0.app/problems/account-disabled"
	TypeDevicePending   = "https://ech0.app/problems/device-authorization-pending"
	TypeDeviceSlowDown  = "https://ech0.app/problems/device-slow-down"
	TypeDeviceExpired   = "https://ech0.app/problems/device-code-expired"
	TypeDeviceDenied    = "https://ech0.app/problems/device-access-denied"
)

const (
	INVALID_FILE_PATH      = "invalid file path"
	INVALID_REQUEST_BODY   = "invalid request body"
	INVALID_PARAMS_BODY    = "invalid parameters"
	INVALID_QUERY_PARAMS   = "invalid query parameters"
	INVALID_REQUEST_METHOD = "invalid request method"
)

const (
	USERNAME_OR_PASSWORD_NOT_BE_EMPTY = "username and password must not be empty"
	USERNAME_OR_PASSWORD_INCORRECT    = "incorrect username or password"
	PASSWORD_TOO_LONG                 = "password must be at most 72 bytes"
	USER_NOTFOUND                     = "user not found"
	USER_COUNT_EXCEED_LIMIT           = "user limit reached on this instance"
	USERNAME_HAS_EXISTS               = "username already exists"
	TOKEN_NOT_FOUND                   = "no token in this request: sign in first"
	TOKEN_NOT_VALID                   = "token is invalid: sign in again"
	TOKEN_PARSE_ERROR                 = "token could not be parsed: sign in again"
	TOKEN_REVOKED                     = "token has been revoked: sign in again"
	ACCOUNT_DISABLED                  = "this account has been disabled by an administrator"
	REFRESH_TOKEN_INVALID             = "refresh token is invalid or expired"
	EXCHANGE_CODE_INVALID             = "authorization code is invalid or expired"
	TOKEN_GENERATE_FAILED             = "failed to issue a token"
	INTERNAL_SERVER_ERROR             = "internal server error"
	DEVICE_CODE_PENDING               = "waiting for approval in the browser"
	DEVICE_CODE_SLOW_DOWN             = "polling too fast: slow down"
	DEVICE_CODE_EXPIRED               = "device code expired: sign in again"
	DEVICE_CODE_DENIED                = "device authorization was denied"
	DEVICE_CODE_UNKNOWN               = "device code is invalid"
	USER_REGISTER_NOT_ALLOW           = "registration is disabled on this instance"
	RATE_LIMIT_EXCEEDED               = "too many attempts: wait a moment and try again"
)

const (
	NO_PERMISSION_DENIED         = "permission denied: contact the instance administrator"
	SCOPE_FORBIDDEN_MSG          = "this token does not carry the required scope"
	AUDIENCE_FORBIDDEN_MSG       = "this token is not valid for this audience"
	TOKEN_QUERY_FORBIDDEN        = "access tokens cannot be sent in the query string: use the Authorization header"
	ECHO_CAN_NOT_BE_EMPTY        = "echo content must not be empty"
	ECHO_NOT_FOUND               = "echo not found"
	ECHO_MIXED_FILE_CATEGORIES   = "an echo may only contain files of a single category"
	ECHO_OTHER_USER_FORBIDDEN    = "only the author may edit or delete this echo"
	INVALID_ECHO_CURSOR          = "invalid pagination cursor"
	ECHO_CURSOR_SORT_UNSUPPORTED = "cursor pagination does not support sorting by interaction"
)

const (
	NO_FILE_UPLOAD_ERROR   = "no uploaded file in this request"
	NO_FILE_STORAGE_ERROR  = "unknown storage backend"
	FILE_TYPE_NOT_ALLOWED  = "file type is not allowed"
	FILE_SIZE_EXCEED_LIMIT = "file exceeds the size limit"
	IMAGE_NOT_FOUND        = "image not found"
	INVALID_PARAMS         = "invalid parameters"
	SIGNUP_FIRST           = "initialize the administrator account first"
	S3_NOT_ENABLED         = "object storage is not enabled"
	S3_NOT_CONFIGURED      = "object storage is not configured"
	S3_CONFIG_ERROR        = "object storage configuration is invalid"
	SYSTEM_ALREADY_INITED  = "instance is already initialized"
	OWNER_ALREADY_EXISTS   = "administrator already exists"
)

const (
	USERNAME_ALREADY_EXISTS      = "username already exists"
	OAUTH2_NOT_CONFIGURED        = "OAuth2 is not configured"
	OAUTH2_NOT_ENABLED           = "OAuth2 is not enabled"
	NO_PERMISSION_BINDING_GITHUB = "binding a GitHub account is not allowed"
	NO_PERMISSION_BINDING_GOOGLE = "binding a Google account is not allowed"
	NO_PERMISSION_BINDING_QQ     = "binding a QQ account is not allowed"
	NO_PERMISSION_BINDING_CUSTOM = "binding a custom OAuth2 account is not allowed"
)

const (
	INVALID_CONNECTION_URL = "connect url must not be empty"
	CONNECT_HAS_EXISTS     = "connect already exists"
)

const (
	WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY = "webhook name and url are required"
	INVALID_WEBHOOK_URL                 = "webhook url is invalid or unsafe"
	INVALID_CRON_EXPRESSION             = "invalid cron expression"
)

const (
	S3_SETTING_INCOMPLETE     = "object storage requires both an endpoint and a bucket"
	S3_CREDENTIALS_INCOMPLETE = "access key and secret key must be set together or left empty together"
	OAUTH2_SETTING_INCOMPLETE = "OAuth2 login requires client id, client secret, auth url and token url"
)

const (
	SNAPSHOT_RESTORE_FAILED = "snapshot restore failed"
	DATABASE_CLOSE_FAILED   = "database close failed"
)

const (
	MIGRATION_JOB_NOT_FOUND = "migration job not found"
)

const (
	CHECK_UPDATE_FAILED = "failed to check for updates"
)

const (
	AGENT_NOT_ENABLED        = "agent is not enabled"
	AGENT_PROTOCOL_NOT_FOUND = "no agent protocol matches this endpoint"
	AGENT_API_KEY_MISSING    = "agent api key is not configured"
	AGENT_MODEL_MISSING      = "agent model name is not configured"
)

const (
	CHAT_ASK_NOT_PENDING = "this question expired or was already answered"
)
