// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	"github.com/lin-snow/ech0/internal/apperr"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// File failures are apperr sentinels. These names are the ones the file
// domain already uses; httperr classifies the underlying sentinel.
var (
	// ErrNotFound is returned for "no such file" and for "not allowed to see
	// it": the two are indistinguishable on the wire so an id cannot be probed.
	ErrNotFound = apperr.ErrFileNotFound
	// ErrUnsupported reports a request the profile cannot serve, e.g. a direct
	// upload into a profile whose access rule would let anyone read a half-written
	// object.
	ErrUnsupported = apperr.ErrFileUnsupported
	// ErrInvalid reports a malformed request: unknown profile, bad URL, blocked
	// extension.
	ErrInvalid = apperr.ErrFileInvalid
	// ErrTooLarge reports a body above the profile's limit.
	ErrTooLarge = apperr.ErrFileTooLarge
	// ErrTypeNotAllowed reports a content type outside the profile's allowlist.
	ErrTypeNotAllowed = apperr.ErrFileTypeRejected
	// ErrConsumed reports a commit of a file that is not pending.
	ErrConsumed = apperr.ErrFileNotPending
	// ErrRangeNotSatisfiable reports a Range the object cannot satisfy.
	ErrRangeNotSatisfiable = apperr.ErrFileRange
)

// ConflictError reports a delete that still has references. References is the
// machine-readable companion; the error unwraps to apperr.ErrFileReferenced.
type ConflictError struct {
	References []fileModel.Reference
}

func (e *ConflictError) Error() string { return apperr.ErrFileReferenced.Error() }

func (e *ConflictError) Unwrap() error { return apperr.ErrFileReferenced }

// Repository is the persistence the service needs. It is declared here, next
// to its consumer, so the repository package can be swapped without the
// service learning about it.
type Repository interface {
	CreateFile(ctx context.Context, file *fileModel.File) error
	GetFile(ctx context.Context, id string) (*fileModel.File, error)
	UpdateName(ctx context.Context, id, name string) error
	DeleteFile(ctx context.Context, id string) (int64, error)
	DeletePendingIfStill(ctx context.Context, id string) (int64, error)
	MarkReady(ctx context.Context, id string) (int64, error)
	PendingOlderThan(ctx context.Context, cutoff int64, limit int) ([]fileModel.File, error)
	ListFiles(ctx context.Context, q fileModel.ListFilter) ([]fileModel.File, int64, error)

	PutBlob(ctx context.Context, fileID string, data []byte) error
	GetBlob(ctx context.Context, fileID string) ([]byte, error)

	AddLinks(ctx context.Context, links []fileModel.Link) error
	DeleteLink(ctx context.Context, fileID, ownerKind, ownerID, role string) error
	DeleteOwnerLinks(ctx context.Context, ownerKind, ownerID, role string) error
	LinksOfFile(ctx context.Context, fileID string) ([]fileModel.Link, error)
	LinksOfOwner(ctx context.Context, ownerKind, ownerID, role string) ([]fileModel.Link, error)
	LinksForFiles(ctx context.Context, ids []string) (map[string][]fileModel.Link, error)
	FilesOfOwner(ctx context.Context, ownerKind, ownerID, role string) ([]fileModel.File, error)
	CountLinks(ctx context.Context, fileID string) (int64, error)
	UnlinkedFilesOfUploader(ctx context.Context, uploaderID string) ([]fileModel.File, error)

	AddGCTask(ctx context.Context, task *fileModel.GCTask) error
	DueGCTasks(ctx context.Context, now int64, limit int) ([]fileModel.GCTask, error)
	UpdateGCTask(ctx context.Context, task *fileModel.GCTask) error
	DeleteGCTask(ctx context.Context, id string) error
	CountGCTasks(ctx context.Context) (int64, error)

	ListStorages(ctx context.Context) ([]fileModel.StorageConfig, error)
	SaveStorage(ctx context.Context, cfg *fileModel.StorageConfig) error
	DeleteStorage(ctx context.Context, id string) error
	SetDefaultStorage(ctx context.Context, id string) error
}

// Transactor runs a closure in one database transaction.
type Transactor interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

// Authorizer answers, for one owning domain, whether a principal may read a
// file that domain attached. It is the only place the file system learns a
// domain's visibility rules, and the domain is the only writer of that answer.
type Authorizer interface {
	Kind() string
	Authorize(ctx context.Context, ownerID string, p viewer.Context) (bool, error)
}

// Service is the file domain as other domains see it: create, read, link,
// delete. Every method takes the caller's principal from the context and
// decides for itself, so no caller can skip a check by forgetting one.
type Service interface {
	Create(ctx context.Context, in CreateInput) (*CreateResult, error)
	Commit(ctx context.Context, id string) (fileModel.FileDTO, error)
	Get(ctx context.Context, id string) (fileModel.FileDTO, error)
	Browse(ctx context.Context, in BrowseInput) (fileModel.Page, error)
	Rename(ctx context.Context, id, name string) (fileModel.FileDTO, error)
	Delete(ctx context.Context, id string, force bool) ([]fileModel.Reference, error)
	Open(ctx context.Context, id string, rng *ByteRange) (*OpenResult, error)

	Replace(ctx context.Context, ids []string, ownerKind, ownerID, role string) error
	Attach(ctx context.Context, ids []string, ownerKind, ownerID, role string) error
	Detach(ctx context.Context, fileID, ownerKind, ownerID, role string) error
	ClearOwner(ctx context.Context, ownerKind, ownerID string) error
	FilesOfOwner(ctx context.Context, ownerKind, ownerID, role string) ([]fileModel.FileDTO, error)
	PurgeUploader(ctx context.Context, uploaderID string) (int, error)

	SweepPending(ctx context.Context, limit int) (int, error)
	DrainGC(ctx context.Context, limit int) (int, error)
	ReloadStorages(ctx context.Context) error
	ApplyS3(ctx context.Context, conf map[string]string, enable bool) error
	TestStorage(ctx context.Context, cfg storage.Config) error
}
