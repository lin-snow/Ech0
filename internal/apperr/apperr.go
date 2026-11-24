// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package apperr is the process-wide set of expected failures.
//
// Callers return these sentinels and wrap them with %w. They do not choose an
// HTTP status. internal/httperr is the only reader of Classes, and the only
// place that turns a sentinel into an RFC 9457 problem document.
package apperr

import (
	"errors"
	"net/http"
)

// Problem type URIs. A client must tell these failures apart from every other
// one carrying the same status, so the URIs are part of the error contract.
const (
	TypeAccountDisabled = "https://ech0.app/problems/account-disabled"
	TypeDevicePending   = "https://ech0.app/problems/device-authorization-pending"
	TypeDeviceSlowDown  = "https://ech0.app/problems/device-slow-down"
	TypeDeviceExpired   = "https://ech0.app/problems/device-code-expired"
	TypeDeviceDenied    = "https://ech0.app/problems/device-access-denied"
)

// Class is one row of the transport table: the sentinel, the HTTP status it
// renders as, and the problem type when a status is not enough.
type Class struct {
	Err    error
	Status int
	Type   string
}

var classes []Class

// Classes is the complete table. httperr projects it; nothing else reads it.
func Classes() []Class { return classes }

func must(msg string, status int, problemType string) error {
	err := errors.New(msg)
	classes = append(classes, Class{Err: err, Status: status, Type: problemType})
	return err
}

var (
	ErrInvalidFilePath    = must("invalid file path", http.StatusBadRequest, "")
	ErrInvalidRequestBody = must("invalid request body", http.StatusBadRequest, "")
	ErrInvalidParams      = must("invalid parameters", http.StatusBadRequest, "")
	ErrInvalidQuery       = must("invalid query parameters", http.StatusBadRequest, "")
	ErrInvalidMethod      = must("invalid request method", http.StatusBadRequest, "")

	ErrCredentialsEmpty     = must("username and password must not be empty", http.StatusBadRequest, "")
	ErrCredentialsIncorrect = must("incorrect username or password", http.StatusUnauthorized, "")
	ErrPasswordTooLong      = must("password must be at most 72 bytes", http.StatusBadRequest, "")
	ErrUserNotFound         = must("user not found", http.StatusNotFound, "")
	ErrUserLimit            = must("user limit reached on this instance", http.StatusForbidden, "")
	ErrUsernameExists       = must("username already exists", http.StatusConflict, "")
	ErrTokenMissing         = must("no token in this request: sign in first", http.StatusUnauthorized, "")
	ErrTokenInvalid         = must("token is invalid: sign in again", http.StatusUnauthorized, "")
	ErrTokenUnparsable      = must("token could not be parsed: sign in again", http.StatusUnauthorized, "")
	ErrTokenRevoked         = must("token has been revoked: sign in again", http.StatusUnauthorized, "")
	ErrAccountDisabled      = must("this account has been disabled by an administrator", http.StatusForbidden, TypeAccountDisabled)
	ErrRefreshTokenInvalid  = must("refresh token is invalid or expired", http.StatusUnauthorized, "")
	ErrExchangeCodeInvalid  = must("authorization code is invalid or expired", http.StatusUnauthorized, "")
	ErrTokenIssueFailed     = must("failed to issue a token", http.StatusInternalServerError, "")
	ErrInternal             = must("internal server error", http.StatusInternalServerError, "")
	ErrDevicePending        = must("waiting for approval in the browser", http.StatusBadRequest, TypeDevicePending)
	ErrDeviceSlowDown       = must("polling too fast: slow down", http.StatusBadRequest, TypeDeviceSlowDown)
	ErrDeviceExpired        = must("device code expired: sign in again", http.StatusBadRequest, TypeDeviceExpired)
	ErrDeviceDenied         = must("device authorization was denied", http.StatusBadRequest, TypeDeviceDenied)
	ErrDeviceUnknown        = must("device code is invalid", http.StatusBadRequest, "")
	ErrRegistrationDisabled = must("registration is disabled on this instance", http.StatusForbidden, "")
	ErrRateLimited          = must("too many attempts: wait a moment and try again", http.StatusTooManyRequests, "")
	ErrEmailRequired        = must("email is required", http.StatusBadRequest, "")
	ErrInvalidEmail         = must("invalid email address", http.StatusBadRequest, "")

	ErrDenied                = must("permission denied: contact the instance administrator", http.StatusForbidden, "")
	ErrScopeForbidden        = must("this token does not carry the required scope", http.StatusForbidden, "")
	ErrAudienceForbidden     = must("this token is not valid for this audience", http.StatusForbidden, "")
	ErrTokenQueryForbidden   = must("access tokens cannot be sent in the query string: use the Authorization header", http.StatusForbidden, "")
	ErrEchoEmpty             = must("echo content must not be empty", http.StatusBadRequest, "")
	ErrEchoNotFound          = must("echo not found", http.StatusNotFound, "")
	ErrEchoMixedFiles        = must("an echo may only contain files of a single category", http.StatusBadRequest, "")
	ErrEchoAuthorOnly        = must("only the author may edit or delete this echo", http.StatusForbidden, "")
	ErrInvalidCursor         = must("invalid pagination cursor", http.StatusBadRequest, "")
	ErrCursorSortUnsupported = must("cursor pagination does not support sorting by interaction", http.StatusBadRequest, "")

	ErrNoUpload           = must("no uploaded file in this request", http.StatusBadRequest, "")
	ErrUnknownStorage     = must("unknown storage backend", http.StatusBadRequest, "")
	ErrFileTypeNotAllowed = must("file type is not allowed", http.StatusUnsupportedMediaType, "")
	ErrFileSize           = must("file exceeds the size limit", http.StatusRequestEntityTooLarge, "")
	ErrImageNotFound      = must("image not found", http.StatusNotFound, "")
	ErrSignupFirst        = must("initialize the administrator account first", http.StatusBadRequest, "")
	ErrS3Disabled         = must("object storage is not enabled", http.StatusBadRequest, "")
	ErrS3NotConfigured    = must("object storage is not configured", http.StatusBadRequest, "")
	ErrS3Config           = must("object storage configuration is invalid", http.StatusBadRequest, "")
	ErrAlreadyInitialized = must("instance is already initialized", http.StatusConflict, "")
	ErrOwnerExists        = must("administrator already exists", http.StatusConflict, "")

	ErrOAuthNotConfigured = must("OAuth2 is not configured", http.StatusBadRequest, "")
	ErrOAuthDisabled      = must("OAuth2 is not enabled", http.StatusForbidden, "")
	ErrBindGitHub         = must("binding a GitHub account is not allowed", http.StatusForbidden, "")
	ErrBindGoogle         = must("binding a Google account is not allowed", http.StatusForbidden, "")
	ErrBindQQ             = must("binding a QQ account is not allowed", http.StatusForbidden, "")
	ErrBindCustom         = must("binding a custom OAuth2 account is not allowed", http.StatusForbidden, "")

	ErrConnectURL    = must("connect url must not be empty", http.StatusBadRequest, "")
	ErrConnectExists = must("connect already exists", http.StatusConflict, "")

	ErrWebhookRequired = must("webhook name and url are required", http.StatusBadRequest, "")
	ErrWebhookURL      = must("webhook url is invalid or unsafe", http.StatusBadRequest, "")
	ErrInvalidCron     = must("invalid cron expression", http.StatusBadRequest, "")

	ErrS3SettingIncomplete    = must("object storage requires both an endpoint and a bucket", http.StatusBadRequest, "")
	ErrS3Credentials          = must("access key and secret key must be set together or left empty together", http.StatusBadRequest, "")
	ErrOAuthSettingIncomplete = must("OAuth2 login requires client id, client secret, auth url and token url", http.StatusBadRequest, "")

	ErrSnapshotRestore      = must("snapshot restore failed", http.StatusInternalServerError, "")
	ErrDatabaseClose        = must("database close failed", http.StatusInternalServerError, "")
	ErrMigrationJobNotFound = must("migration job not found", http.StatusNotFound, "")
	ErrCheckUpdate          = must("failed to check for updates", http.StatusInternalServerError, "")

	ErrAgentDisabled = must("agent is not enabled", http.StatusBadRequest, "")
	ErrAgentProtocol = must("no agent protocol matches this endpoint", http.StatusBadRequest, "")
	ErrAgentKey      = must("agent api key is not configured", http.StatusBadRequest, "")
	ErrAgentModel    = must("agent model name is not configured", http.StatusBadRequest, "")
	ErrAskNotPending = must("this question expired or was already answered", http.StatusConflict, "")

	ErrFileNotFound     = must("file: not found", http.StatusNotFound, "")
	ErrFileUnsupported  = must("file: unsupported", http.StatusBadRequest, "")
	ErrFileInvalid      = must("file: invalid request", http.StatusBadRequest, "")
	ErrFileTooLarge     = must("file: too large", http.StatusRequestEntityTooLarge, "")
	ErrFileTypeRejected = must("file: type not allowed", http.StatusUnsupportedMediaType, "")
	ErrFileNotPending   = must("file: not pending", http.StatusConflict, "")
	ErrFileRange        = must("file: range not satisfiable", http.StatusRequestedRangeNotSatisfiable, "")
	ErrFileReferenced   = must("file: still referenced", http.StatusConflict, "")

	ErrCommentHidden          = must("comment target is not visible", http.StatusBadRequest, "")
	ErrCommentRejected        = must("submission rejected", http.StatusBadRequest, "")
	ErrCommentForm            = must("submitted too fast or the form expired", http.StatusBadRequest, "")
	ErrCommentsDisabled       = must("comments are disabled", http.StatusBadRequest, "")
	ErrCaptcha                = must("captcha verification failed", http.StatusBadRequest, "")
	ErrCommentEmpty           = must("comment content must not be empty", http.StatusBadRequest, "")
	ErrCommentTooLong         = must("comment content must be at most 200 characters", http.StatusBadRequest, "")
	ErrCommentIdentity        = must("nickname and email are required", http.StatusBadRequest, "")
	ErrInvalidWebsite         = must("invalid website url", http.StatusBadRequest, "")
	ErrDuplicateComment       = must("duplicate comment", http.StatusConflict, "")
	ErrReplyMissing           = must("the comment being replied to does not exist", http.StatusNotFound, "")
	ErrReplyForbidden         = must("this comment cannot be replied to", http.StatusBadRequest, "")
	ErrCommentRate            = must("comment rate limit exceeded", http.StatusTooManyRequests, "")
	ErrIntegrationCommentRate = must("integration comment rate limit exceeded", http.StatusTooManyRequests, "")
	ErrCommentStatus          = must("invalid comment status", http.StatusBadRequest, "")
	ErrCommentBulk            = must("invalid bulk action", http.StatusBadRequest, "")

	ErrAdminEmailRequired = must("admin email is required", http.StatusBadRequest, "")
	ErrAdminEmailInvalid  = must("invalid admin email", http.StatusBadRequest, "")
	ErrSMTPHost           = must("SMTP host is required", http.StatusBadRequest, "")
	ErrSMTPPort           = must("invalid SMTP port", http.StatusBadRequest, "")
	ErrSMTPUser           = must("SMTP username is required", http.StatusBadRequest, "")
	ErrSMTPFrom           = must("invalid sender address: check the SMTP from address", http.StatusBadRequest, "")
	ErrSMTPPassword       = must("SMTP password is required", http.StatusBadRequest, "")
	ErrAdminEmailMissing  = must("admin email is not configured", http.StatusBadRequest, "")
	ErrMailerUnavailable  = must("mailer unavailable", http.StatusBadRequest, "")

	ErrOrigin                 = must("origin not allowed", http.StatusForbidden, "")
	ErrNoSuchEndpoint         = must("no such endpoint", http.StatusNotFound, "")
	ErrMaintenance            = must("service is in maintenance mode: writes are temporarily disabled", http.StatusServiceUnavailable, "")
	ErrCatalogStale           = must("operation catalog is stale, refresh it from X-Ech0-Catalog", http.StatusPreconditionFailed, "")
	ErrCatalogUnavailable     = must("operation catalog unavailable", http.StatusServiceUnavailable, "")
	ErrAuthenticationRequired = must("authentication required", http.StatusUnauthorized, "")
	ErrSessionRequired        = must("a browser session is required", http.StatusForbidden, "")
	ErrAccountState           = must("account state lookup failed", http.StatusInternalServerError, "")

	ErrMigrationBusy    = must("finish or clean up the current migration first", http.StatusConflict, "")
	ErrMigrationRunning = must("cannot clean up while a migration is running", http.StatusConflict, "")
	ErrExportRunning    = must("an export is already running", http.StatusConflict, "")
	ErrNoExport         = must("no artifact to download: create an export first", http.StatusNotFound, "")
)
