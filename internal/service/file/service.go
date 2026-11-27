// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package service is the file resource domain: it turns an upload into a file
// row, answers who may read it, and moves links. It owns every decision about
// a file; the storage package below it only moves bytes.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	imgUtil "github.com/lin-snow/ech0/internal/util/img"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const (
	// pendingTTL is how long a direct upload may stay uncommitted.
	pendingTTL = 24 * time.Hour
	// directTTL is how long a presigned PUT stays valid. It must stay short:
	// the URL is a second key to a key, and a commit does not revoke it.
	directTTL = 15 * time.Minute
	// snapshotPrefix is how many bytes commit re-sniffs to compare with the
	// declared content type.
	sniffBytes = 512
)

// roleAdmin is the role that may act on another user's file. It is the only
// place administrator override is spelled.
const roleAdmin = role.Admin

var errTypeMismatch = errors.New("file: uploaded type does not match declaration")

type FileService struct {
	tx          Transactor
	repo        Repository
	storages    *storage.Registry
	authorizers *AuthorizerRegistry

	storageMu      sync.Mutex
	storagesLoaded bool
	storagesErr    error
}

func NewFileService(
	tx Transactor,
	repo Repository,
	storages *storage.Registry,
	authorizers *AuthorizerRegistry,
) *FileService {
	return &FileService{tx: tx, repo: repo, storages: storages, authorizers: authorizers}
}

// CreateInput is one upload request. ExternalURL and File are exclusive.
type CreateInput struct {
	Profile     string
	File        *multipart.FileHeader
	ExternalURL string
	Direct      bool
	// Name overrides the uploaded filename as the display name. ContentType is
	// only honoured for a reference: for bytes the sniffed type is the truth.
	Name        string
	ContentType string
	Width       int
	Height      int
}

// CreateResult carries either the stored file or — for a direct upload — the
// pending row plus the presigned request that fills it.
type CreateResult struct {
	File   fileModel.FileDTO       `json:"file"`
	Upload *fileModel.DirectUpload `json:"upload,omitempty"`
}

// Create turns one body into a file row. It never leaves a partial object
// behind: a failed row insert deletes the bytes just written, and a failed
// delete is enqueued for the collector.
func (s *FileService) Create(ctx context.Context, in CreateInput) (*CreateResult, error) {
	p := principalOf(ctx)
	if viewer.IsAnonymous(p) {
		return nil, ErrNotFound
	}
	prof, err := profileFor(in.Profile)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.ExternalURL) != "" {
		return s.createRef(ctx, prof, p, in)
	}
	if in.Direct && !prof.AllowDirect {
		// A presigned PUT cannot carry the image-only rule, and a public file
		// must not become writable by a client-held URL.
		return nil, ErrUnsupported
	}
	if in.File == nil && !in.Direct {
		return nil, ErrInvalid
	}
	if prof.Kind() == fileModel.KindBlob {
		return s.createBlob(ctx, prof, p, in.File)
	}
	return s.createOwned(ctx, prof, p, in)
}

// displayName prefers what the caller asked for, then the uploaded filename.
// Both are optional on the wire: a file may have no name at all.
func displayName(requested, uploaded string) string {
	if name := strings.TrimSpace(requested); name != "" {
		return name
	}
	return strings.TrimSpace(uploaded)
}

// firstNonZero keeps a caller-supplied dimension and falls back to the sniffed
// one; zero means "unknown", never a real width or height.
func firstNonZero(requested, sniffed int) int {
	if requested > 0 {
		return requested
	}
	return sniffed
}

func (s *FileService) createRef(ctx context.Context, prof Profile, p viewer.Context, in CreateInput) (*CreateResult, error) {
	parsed, err := url.Parse(strings.TrimSpace(in.ExternalURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, ErrInvalid
	}
	if parsed.User != nil {
		return nil, ErrInvalid
	}
	contentType := strings.TrimSpace(in.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(path.Ext(parsed.Path)))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ref := parsed.String()
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = path.Base(parsed.Path)
		if name == "." || name == "/" {
			name = ""
		}
	}
	row := &fileModel.File{
		ID:          uuidUtil.NewV7(),
		Kind:        string(fileModel.KindRef),
		Status:      string(fileModel.StatusReady),
		Access:      string(prof.Access),
		RefURL:      &ref,
		Name:        name,
		ContentType: contentType,
		Width:       in.Width,
		Height:      in.Height,
		UploaderID:  p.UserID(),
	}
	if err := s.repo.CreateFile(ctx, row); err != nil {
		return nil, err
	}
	return &CreateResult{File: toDTO(row, nil)}, nil
}

func (s *FileService) createBlob(ctx context.Context, prof Profile, p viewer.Context, fh *multipart.FileHeader) (*CreateResult, error) {
	up, err := inspect(fh, prof)
	if err != nil {
		return nil, err
	}
	if up.Size > prof.MaxBytes {
		return nil, ErrTooLarge
	}
	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()
	data, err := io.ReadAll(io.LimitReader(src, prof.MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > prof.MaxBytes {
		return nil, ErrTooLarge
	}
	row := &fileModel.File{
		ID:          uuidUtil.NewV7(),
		Kind:        string(fileModel.KindBlob),
		Status:      string(fileModel.StatusReady),
		Access:      string(prof.Access),
		BlobBytes:   new(int64(len(data))),
		Name:        up.Filename,
		ContentType: up.ContentType,
		Size:        int64(len(data)),
		Width:       up.Width,
		Height:      up.Height,
		UploaderID:  p.UserID(),
	}
	err = s.tx.Run(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateFile(txCtx, row); err != nil {
			return err
		}
		return s.repo.PutBlob(txCtx, row.ID, data)
	})
	if err != nil {
		return nil, err
	}
	return &CreateResult{File: toDTO(row, nil)}, nil
}

func (s *FileService) createOwned(ctx context.Context, prof Profile, p viewer.Context, in CreateInput) (*CreateResult, error) {
	up, err := inspect(in.File, prof)
	if err != nil {
		return nil, err
	}
	category := fileModel.CategoryOf(up.ContentType)
	if limit := prof.limit(category); up.Size > limit {
		return nil, ErrTooLarge
	}
	if err := s.ensureStorages(ctx); err != nil {
		return nil, err
	}
	storageID, store, err := s.storages.Default()
	if err != nil {
		return nil, err
	}
	if in.Direct {
		return s.createDirect(ctx, prof, p, up, storageID, store)
	}

	id := uuidUtil.NewV7()
	objectKey, err := storage.KeyFor(id, filepath.Ext(up.Filename))
	if err != nil {
		return nil, err
	}
	src, err := in.File.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()

	put, err := store.Put(ctx, objectKey, src, up.Size)
	if err != nil {
		return nil, err
	}
	if put.Size != up.Size {
		s.collect(ctx, storageID, objectKey)
		return nil, storage.ErrSizeMismatch
	}
	row := &fileModel.File{
		ID:          id,
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusReady),
		Access:      string(prof.Access),
		StorageID:   &storageID,
		ObjectKey:   &objectKey,
		Name:        displayName(in.Name, up.Filename),
		ContentType: up.ContentType,
		Size:        put.Size,
		Width:       firstNonZero(in.Width, up.Width),
		Height:      firstNonZero(in.Height, up.Height),
		UploaderID:  p.UserID(),
	}
	if err := s.repo.CreateFile(ctx, row); err != nil {
		s.collect(ctx, storageID, objectKey)
		return nil, err
	}
	return &CreateResult{File: toDTO(row, nil)}, nil
}

// createDirect opens a presigned PUT for a pending row. Size and content type
// are signed into the request, so the client cannot store something the row
// does not declare.
func (s *FileService) createDirect(
	ctx context.Context,
	prof Profile,
	p viewer.Context,
	up upload,
	storageID string,
	store storage.Store,
) (*CreateResult, error) {
	if !prof.AllowDirect {
		return nil, ErrUnsupported
	}
	presigner, ok := store.(storage.Presigner)
	if !ok {
		return nil, ErrUnsupported
	}
	id := uuidUtil.NewV7()
	objectKey, err := storage.KeyFor(id, filepath.Ext(up.Filename))
	if err != nil {
		return nil, err
	}
	row := &fileModel.File{
		ID:          id,
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusPending),
		Access:      string(prof.Access),
		StorageID:   &storageID,
		ObjectKey:   &objectKey,
		Name:        up.Filename,
		ContentType: up.ContentType,
		Size:        up.Size,
		Width:       up.Width,
		Height:      up.Height,
		UploaderID:  p.UserID(),
	}
	if err := s.repo.CreateFile(ctx, row); err != nil {
		return nil, err
	}
	presigned, headers, err := presigner.PresignPut(ctx, objectKey, directTTL, up.ContentType, up.Size)
	if err != nil {
		s.collect(ctx, storageID, objectKey)
		_, _ = s.repo.DeleteFile(ctx, row.ID)
		return nil, err
	}
	return &CreateResult{
		File: toDTO(row, nil),
		Upload: &fileModel.DirectUpload{
			URL:       presigned,
			Method:    "PUT",
			Headers:   headers,
			ExpiresAt: time.Now().Add(directTTL).Unix(),
		},
	}, nil
}

// Commit closes a direct upload. It is the only idempotent step: committing a
// ready file returns it, committing an unknown or foreign one is a 404.
func (s *FileService) Commit(ctx context.Context, id string) (fileModel.FileDTO, error) {
	p := principalOf(ctx)
	f, err := s.load(ctx, id)
	if err != nil {
		return fileModel.FileDTO{}, err
	}
	if f.Status == string(fileModel.StatusReady) {
		if err := s.authorize(ctx, f, p); err != nil {
			return fileModel.FileDTO{}, err
		}
		return toDTO(f, nil), nil
	}
	if f.Kind != string(fileModel.KindOwned) || f.StorageID == nil || f.ObjectKey == nil {
		return fileModel.FileDTO{}, ErrConsumed
	}
	if viewer.IsAnonymous(p) || (p.UserID() != f.UploaderID && !p.Role().AtLeast(roleAdmin)) {
		return fileModel.FileDTO{}, ErrNotFound
	}
	if err := s.ensureStorages(ctx); err != nil {
		return fileModel.FileDTO{}, err
	}
	store, err := s.storages.For(*f.StorageID)
	if err != nil {
		return fileModel.FileDTO{}, err
	}
	info, err := store.Stat(ctx, *f.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fileModel.FileDTO{}, ErrNotFound
		}
		return fileModel.FileDTO{}, err
	}
	if info.Size != f.Size {
		return fileModel.FileDTO{}, ErrInvalid
	}
	if err := s.verifyDeclaredType(ctx, store, f); err != nil {
		s.collect(ctx, *f.StorageID, *f.ObjectKey)
		if _, delErr := s.repo.DeleteFile(ctx, f.ID); delErr != nil {
			return fileModel.FileDTO{}, delErr
		}
		return fileModel.FileDTO{}, err
	}

	rows, err := s.repo.MarkReady(ctx, f.ID)
	if err != nil {
		return fileModel.FileDTO{}, err
	}
	if rows == 0 {
		return fileModel.FileDTO{}, ErrConsumed
	}
	f.Status = string(fileModel.StatusReady)
	return toDTO(f, nil), nil
}

// verifyDeclaredType re-sniffs the first block of a committed object. A client
// that signed an image content type but uploaded an HTML page would otherwise
// get that page served from the gateway with a trusted type.
func (s *FileService) verifyDeclaredType(ctx context.Context, store storage.Store, f *fileModel.File) error {
	reader, _, err := store.Open(ctx, *f.ObjectKey)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()
	buf := make([]byte, sniffBytes)
	n, err := io.ReadFull(reader, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return err
	}
	sniffed := storage.ContentTypeOf(buf[:n])
	if sniffed == "" || sniffed == "application/octet-stream" {
		return nil
	}
	if !sameMediaFamily(sniffed, f.ContentType) {
		return errTypeMismatch
	}
	return nil
}

func sameMediaFamily(sniffed, declared string) bool {
	if sniffed == declared {
		return true
	}
	return mediaFamily(sniffed) == mediaFamily(declared)
}

func mediaFamily(contentType string) string {
	if i := strings.IndexByte(contentType, '/'); i > 0 {
		return contentType[:i]
	}
	return contentType
}

// authorize turns the file's rule into a yes/no and folds "no" into the same
// not-found answer the gateway uses everywhere, so a denial cannot leak that
// the id exists.
func (s *FileService) authorize(ctx context.Context, f *fileModel.File, p viewer.Context) error {
	allowed, err := s.decide(ctx, f, p)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrNotFound
	}
	return nil
}

// Get returns the metadata of one file the principal may see.
func (s *FileService) Get(ctx context.Context, id string) (fileModel.FileDTO, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return fileModel.FileDTO{}, err
	}
	if err := s.authorize(ctx, f, principalOf(ctx)); err != nil {
		return fileModel.FileDTO{}, err
	}
	return toDTO(f, nil), nil
}

// BrowseInput is the file-browser query.
type BrowseInput struct {
	Page     int
	PageSize int
	Search   string
	Kind     string
	Access   string
	Status   string
	Owner    string
	OwnerID  string
	Role     string
}

// Browse lists files. Non-administrators only ever list their own uploads.
func (s *FileService) Browse(ctx context.Context, in BrowseInput) (fileModel.Page, error) {
	p := principalOf(ctx)
	if viewer.IsAnonymous(p) {
		return fileModel.Page{}, ErrNotFound
	}
	filter := fileModel.ListFilter{
		UploaderID: p.UserID(),
		Kind:       in.Kind,
		Access:     in.Access,
		Status:     in.Status,
		Search:     in.Search,
		Page:       in.Page,
		PageSize:   in.PageSize,
	}
	if p.Role().AtLeast(roleAdmin) {
		filter.UploaderID = ""
	}
	if in.Owner != "" && in.OwnerID != "" {
		if !p.Role().AtLeast(roleAdmin) {
			return fileModel.Page{}, ErrNotFound
		}
		filter.OwnerKind = in.Owner
		filter.OwnerID = in.OwnerID
		filter.Role = in.Role
	}
	rows, total, err := s.repo.ListFiles(ctx, filter)
	if err != nil {
		return fileModel.Page{}, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	links, err := s.repo.LinksForFiles(ctx, ids)
	if err != nil {
		return fileModel.Page{}, err
	}
	items := make([]fileModel.FileDTO, 0, len(rows))
	for i := range rows {
		items = append(items, toDTO(&rows[i], links[rows[i].ID]))
	}
	return fileModel.Page{Total: total, Items: items}, nil
}

// Rename sets the display name. It is the only mutable field.
func (s *FileService) Rename(ctx context.Context, id, name string) (fileModel.FileDTO, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return fileModel.FileDTO{}, err
	}
	p := principalOf(ctx)
	if viewer.IsAnonymous(p) || (p.UserID() != f.UploaderID && !p.Role().AtLeast(roleAdmin)) {
		return fileModel.FileDTO{}, ErrNotFound
	}
	name = strings.TrimSpace(name)
	if len(name) > 255 {
		name = name[:255]
	}
	if err := s.repo.UpdateName(ctx, id, name); err != nil {
		return fileModel.FileDTO{}, err
	}
	f.Name = name
	return toDTO(f, nil), nil
}

// Delete removes a file. A referenced file refuses unless force is set, and
// force is an administrator operation: the references belong to someone else.
func (s *FileService) Delete(ctx context.Context, id string, force bool) ([]fileModel.Reference, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	p := principalOf(ctx)
	if viewer.IsAnonymous(p) || (p.UserID() != f.UploaderID && !p.Role().AtLeast(roleAdmin)) {
		return nil, ErrNotFound
	}
	links, err := s.repo.LinksOfFile(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(links) > 0 && (!force || !p.Role().AtLeast(roleAdmin)) {
		refs := make([]fileModel.Reference, 0, len(links))
		for _, link := range links {
			refs = append(refs, fileModel.Reference{
				OwnerKind: link.OwnerKind,
				OwnerID:   link.OwnerID,
				Role:      link.Role,
			})
		}
		return refs, &ConflictError{References: refs}
	}
	err = s.tx.Run(ctx, func(txCtx context.Context) error {
		for _, link := range links {
			if err := s.repo.DeleteLink(txCtx, link.FileID, link.OwnerKind, link.OwnerID, link.Role); err != nil {
				return err
			}
		}
		if _, err := s.repo.DeleteFile(txCtx, id); err != nil {
			return err
		}
		if f.IsOwned() && f.StorageID != nil && f.ObjectKey != nil {
			return s.enqueueGC(txCtx, *f.StorageID, *f.ObjectKey)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// OpenResult is everything the gateway needs to answer one content request.
type OpenResult struct {
	File     fileModel.FileDTO
	Redirect string
	Reader   io.ReadCloser
	// Size is the bytes this response carries; Total is the whole object.
	// Content-Range needs both, and after a range request they differ.
	Size        int64
	Total       int64
	ContentType string
	ModTime     time.Time
	Rangeable   bool
	// CacheControl is decided here, not by the transport: it depends on the
	// file's access rule and on who is asking (§7.1).
	CacheControl string
	// VaryAuthorization marks the one case a shared cache may hold the
	// response: an anonymous read of an owner file.
	VaryAuthorization bool
}

// ByteRange is a parsed HTTP range.
// ByteRange is one requested byte range. Suffix holds the HTTP "bytes=-N"
// form — the last N bytes — and is resolved against the real size before any
// store is touched; after resolution Start and Length are absolute.
type ByteRange struct {
	Start  int64
	Length int64
	Suffix int64
}

// resolveRange clamps a request to the actual size, turning a suffix request
// into an absolute one. A start at or past the end is not satisfiable; an open
// end means "through the last byte".
func resolveRange(spec *ByteRange, total int64) (int64, int64, bool) {
	if total <= 0 {
		return 0, 0, false
	}
	if spec.Suffix > 0 {
		if spec.Suffix > total {
			spec.Suffix = total
		}
		spec.Start = total - spec.Suffix
		spec.Length = spec.Suffix
		spec.Suffix = 0
		return spec.Start, spec.Length, true
	}
	if spec.Start < 0 || spec.Start >= total {
		return 0, 0, false
	}
	if spec.Length <= 0 || spec.Start+spec.Length > total {
		spec.Length = total - spec.Start
	}
	return spec.Start, spec.Length, true
}

// Open authorizes and opens the bytes of one file. Redirect is set for a
// public object in a signing store and for every external reference; a private
// file is never signed.
func (s *FileService) Open(ctx context.Context, id string, rng *ByteRange) (*OpenResult, error) {
	f, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	p := principalOf(ctx)
	if err := s.authorize(ctx, f, p); err != nil {
		return nil, err
	}
	out := &OpenResult{File: toDTO(f, nil), ContentType: f.ContentType, Size: f.Size}
	out.CacheControl, out.VaryAuthorization = cacheControlFor(f, p)

	switch f.Kind {
	case string(fileModel.KindRef):
		if f.Access != string(fileModel.AccessPublic) || f.RefURL == nil {
			return nil, ErrNotFound
		}
		out.Redirect = *f.RefURL
		return out, nil
	case string(fileModel.KindBlob):
		if f.Status != string(fileModel.StatusReady) {
			return nil, ErrNotFound
		}
		data, err := s.repo.GetBlob(ctx, f.ID)
		if err != nil {
			return nil, ErrNotFound
		}
		out.Size = int64(len(data))
		out.Total = out.Size
		out.Rangeable = true
		body := data
		if rng != nil {
			start, length, ok := resolveRange(rng, int64(len(data)))
			if !ok {
				return nil, ErrRangeNotSatisfiable
			}
			body = data[start : start+length]
			out.Size = length
		}
		out.Reader = io.NopCloser(bytes.NewReader(body))
		return out, nil
	}

	if f.Status != string(fileModel.StatusReady) || f.StorageID == nil || f.ObjectKey == nil {
		return nil, ErrNotFound
	}
	if err := s.ensureStorages(ctx); err != nil {
		return nil, err
	}
	store, err := s.storages.For(*f.StorageID)
	if err != nil {
		return nil, err
	}
	if f.Access == string(fileModel.AccessPublic) && rng == nil {
		// A configured CDN origin serves the bytes without touching this
		// process; a presigned URL is the fallback when there is none.
		if public := s.storages.PublicURL(*f.StorageID, *f.ObjectKey); public != "" {
			out.Redirect = public
			return out, nil
		}
		if presigner, ok := store.(storage.Presigner); ok {
			if signed, signErr := presigner.PresignGet(ctx, *f.ObjectKey, directTTL); signErr == nil {
				out.Redirect = signed
				return out, nil
			}
		}
	}
	if rng != nil {
		info, statErr := store.Stat(ctx, *f.ObjectKey)
		if statErr != nil {
			return nil, mapStoreErr(statErr)
		}
		out.ModTime = info.ModTime
		out.Total = info.Size
		start, length, ok := resolveRange(rng, info.Size)
		if !ok {
			return nil, ErrRangeNotSatisfiable
		}
		out.Size = length
		out.Rangeable = true
		if opener, isOpener := store.(storage.RangeOpener); isOpener {
			reader, err := opener.OpenRange(ctx, *f.ObjectKey, start, length)
			if err != nil {
				return nil, mapStoreErr(err)
			}
			out.Reader = reader
			return out, nil
		}
		reader, opened, err := store.Open(ctx, *f.ObjectKey)
		if err != nil {
			return nil, mapStoreErr(err)
		}
		out.ModTime = opened.ModTime
		seeker, isSeeker := reader.(io.Seeker)
		if !isSeeker {
			_ = reader.Close()
			return nil, ErrRangeNotSatisfiable
		}
		if _, err := seeker.Seek(start, io.SeekStart); err != nil {
			_ = reader.Close()
			return nil, ErrRangeNotSatisfiable
		}
		out.Reader = reader
		return out, nil
	}
	reader, info, err := store.Open(ctx, *f.ObjectKey)
	if err != nil {
		return nil, mapStoreErr(err)
	}
	out.Reader = reader
	out.ModTime = info.ModTime
	if info.Size > 0 {
		out.Size = info.Size
	}
	out.Total = out.Size
	out.Rangeable = true
	return out, nil
}

// cacheControlFor is the whole caching decision (§7.1). It reads the access
// rule and the caller, never a stored visibility value: a row whose owner turns
// private must stop being cacheable the same instant it stops being readable.
func cacheControlFor(f *fileModel.File, p viewer.Context) (cacheControl string, vary bool) {
	switch f.Access {
	case string(fileModel.AccessPublic):
		return "public, max-age=31536000, immutable", false
	case string(fileModel.AccessOwner):
		if viewer.IsAnonymous(p) {
			return "public, max-age=0, must-revalidate", true
		}
		return "private, no-store", false
	default:
		return "private, no-store", false
	}
}

func mapStoreErr(err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// ---- links -------------------------------------------------------------

// singleSlotRoles may appear at most once per owner. The database enforces the
// same list with a partial unique index; this check turns the violation into a
// 400 instead of a driver error.
var singleSlotRoles = map[string]struct{}{
	"cover": {}, "avatar": {}, "logo": {}, "favicon": {}, "og_image": {},
}

// Replace sets the full link set of one (owner, role) slot. It is a whole-set
// write, never a merge: a document that dropped a reference must lose the
// link, or the file outlives its only reader.
func (s *FileService) Replace(ctx context.Context, ids []string, ownerKind, ownerID, role string) error {
	if ownerKind == "" || ownerID == "" || role == "" {
		return ErrInvalid
	}
	unique := dedupe(ids)
	if _, single := singleSlotRoles[role]; single && len(unique) > 1 {
		return ErrInvalid
	}
	p := principalOf(ctx)
	if viewer.IsAnonymous(p) {
		return ErrNotFound
	}
	for _, id := range unique {
		f, err := s.load(ctx, id)
		if err != nil {
			return err
		}
		if allowed, err := s.decide(ctx, f, p); err != nil || !allowed {
			return ErrNotFound
		}
	}
	existing, err := s.repo.LinksOfOwner(ctx, ownerKind, ownerID, role)
	if err != nil {
		return err
	}
	keep := make(map[string]struct{}, len(unique))
	for _, id := range unique {
		keep[id] = struct{}{}
	}
	return s.tx.Run(ctx, func(txCtx context.Context) error {
		for _, link := range existing {
			if _, ok := keep[link.FileID]; ok {
				continue
			}
			if err := s.repo.DeleteLink(txCtx, link.FileID, link.OwnerKind, link.OwnerID, link.Role); err != nil {
				return err
			}
		}
		added := make([]fileModel.Link, 0, len(unique))
		for order, id := range unique {
			added = append(added, fileModel.Link{
				FileID:    id,
				OwnerKind: ownerKind,
				OwnerID:   ownerID,
				Role:      role,
				SortOrder: order,
			})
		}
		return s.repo.AddLinks(txCtx, added)
	})
}

// Attach adds links without touching existing ones. It exists for one-shot
// attachments; document-like owners use Replace.
func (s *FileService) Attach(ctx context.Context, ids []string, ownerKind, ownerID, role string) error {
	if ownerKind == "" || ownerID == "" || role == "" {
		return ErrInvalid
	}
	return s.Replace(ctx, ids, ownerKind, ownerID, role)
}

// Detach removes one link. The file stays alive, unlinked.
func (s *FileService) Detach(ctx context.Context, fileID, ownerKind, ownerID, role string) error {
	return s.repo.DeleteLink(ctx, fileID, ownerKind, ownerID, role)
}

// ClearOwner removes every link of an owner, which is what deleting the owner
// does. Bytes stay until the collector decides they are garbage.
func (s *FileService) ClearOwner(ctx context.Context, ownerKind, ownerID string) error {
	return s.repo.DeleteOwnerLinks(ctx, ownerKind, ownerID, "")
}

// PurgeUploader removes the files a user uploaded that nothing links to any
// more, and returns how many rows went away. It is the user-deletion path: a
// file someone else attached to their own echo is not the deleted user's to
// take away, and it survives with its link.
func (s *FileService) PurgeUploader(ctx context.Context, uploaderID string) (int, error) {
	if uploaderID == "" {
		return 0, ErrInvalid
	}
	purged := 0
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		rows, err := s.repo.UnlinkedFilesOfUploader(txCtx, uploaderID)
		if err != nil {
			return err
		}
		for i := range rows {
			removed, err := s.repo.DeleteFile(txCtx, rows[i].ID)
			if err != nil {
				return err
			}
			if removed == 0 {
				continue
			}
			purged++
			if rows[i].IsOwned() && rows[i].StorageID != nil && rows[i].ObjectKey != nil {
				if err := s.enqueueGC(txCtx, *rows[i].StorageID, *rows[i].ObjectKey); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return purged, err
}

// FilesOfOwner hydrates an owner's attachments for its DTO. Authorization is
// the owner's responsibility: the caller already decided the owner is visible.
func (s *FileService) FilesOfOwner(ctx context.Context, ownerKind, ownerID, role string) ([]fileModel.FileDTO, error) {
	rows, err := s.repo.FilesOfOwner(ctx, ownerKind, ownerID, role)
	if err != nil {
		return nil, err
	}
	out := make([]fileModel.FileDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toDTO(&rows[i], nil))
	}
	return out, nil
}

// ---- internals ---------------------------------------------------------

func (s *FileService) load(ctx context.Context, id string) (*fileModel.File, error) {
	f, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return f, nil
}

func (s *FileService) collect(ctx context.Context, storageID, objectKey string) {
	if err := s.enqueueGC(ctx, storageID, objectKey); err != nil {
		logUtil.GetLogger().Warn(
			"failed to enqueue gc task",
			slog.String("module", "file.gc"),
			slog.String("object_key", objectKey),
			logUtil.Err(err),
		)
	}
}

func (s *FileService) enqueueGC(ctx context.Context, storageID, objectKey string) error {
	return s.repo.AddGCTask(ctx, &fileModel.GCTask{
		ID:        uuidUtil.NewV7(),
		StorageID: storageID,
		ObjectKey: objectKey,
		NextTryAt: time.Now().Unix(),
	})
}

// SweepPending removes direct uploads that were never committed. The delete is
// conditional on status=pending, so a commit that lands during the sweep always
// wins; a row the sweep did not delete is simply skipped.
func (s *FileService) SweepPending(ctx context.Context, limit int) (int, error) {
	cutoff := time.Now().Add(-pendingTTL).Unix()
	rows, err := s.repo.PendingOlderThan(ctx, cutoff, limit)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	if err := s.ensureStorages(ctx); err != nil {
		return 0, err
	}
	removed := 0
	for _, row := range rows {
		n, err := s.repo.DeletePendingIfStill(ctx, row.ID)
		if err != nil {
			return removed, err
		}
		if n == 0 {
			continue
		}
		if row.IsOwned() && row.StorageID != nil && row.ObjectKey != nil {
			s.collect(ctx, *row.StorageID, *row.ObjectKey)
		}
		removed++
	}
	return removed, nil
}

// DrainGC retries object deletions whose row is already gone. A failed delete
// backs off exponentially instead of blocking the queue: the bytes stay in the
// storage and the task keeps its error for the operator.
func (s *FileService) DrainGC(ctx context.Context, limit int) (int, error) {
	if err := s.ensureStorages(ctx); err != nil {
		return 0, err
	}
	tasks, err := s.repo.DueGCTasks(ctx, time.Now().Unix(), limit)
	if err != nil {
		return 0, err
	}
	drained := 0
	for i := range tasks {
		task := tasks[i]
		store, err := s.storages.For(task.StorageID)
		if err != nil {
			s.backoff(ctx, &task, err)
			continue
		}
		if err := store.Delete(ctx, task.ObjectKey); err != nil && !errors.Is(err, storage.ErrNotFound) {
			s.backoff(ctx, &task, err)
			continue
		}
		if err := s.repo.DeleteGCTask(ctx, task.ID); err != nil {
			return drained, err
		}
		drained++
	}
	return drained, nil
}

func (s *FileService) backoff(ctx context.Context, task *fileModel.GCTask, cause error) {
	task.Attempts++
	wait := time.Minute << min(task.Attempts, 6)
	task.NextTryAt = time.Now().Add(wait).Unix()
	task.LastError = cause.Error()
	if err := s.repo.UpdateGCTask(ctx, task); err != nil {
		logUtil.GetLogger().Warn(
			"failed to reschedule gc task",
			slog.String("module", "storage.gc"),
			slog.String("task_id", task.ID),
			logUtil.Err(err),
		)
	}
}

// S3StorageID is the row the settings panel owns. It is fixed so the panel can
// upsert it without first finding it in the table.
const S3StorageID = "s3"

// ApplyS3 makes the storages table follow the object-storage settings panel.
// It is the panel's only write path into the byte layer: enabling it swaps the
// default to the s3 row, disabling it removes the row and the local store takes
// the default back. A failing open leaves the previous snapshot serving.
func (s *FileService) ApplyS3(ctx context.Context, conf map[string]string, enable bool) error {
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		if !enable {
			if err := s.repo.DeleteStorage(txCtx, S3StorageID); err != nil {
				return err
			}
			return s.repo.SetDefaultStorage(txCtx, LocalStorageID)
		}
		if err := s.repo.SaveStorage(txCtx, &fileModel.StorageConfig{
			ID:        S3StorageID,
			Driver:    storage.DriverS3,
			Name:      "s3",
			IsDefault: true,
			Config:    conf,
		}); err != nil {
			return err
		}
		return s.repo.SetDefaultStorage(txCtx, S3StorageID)
	})
	if err != nil {
		return err
	}
	return s.ReloadStorages(ctx)
}

// TestStorage opens a throwaway store from cfg and checks it answers. It never
// touches the live snapshot, so a bad panel value cannot break serving.
func (s *FileService) TestStorage(ctx context.Context, cfg storage.Config) error {
	store, err := storage.OpenDriver(ctx, cfg)
	if err != nil {
		return err
	}
	_, err = store.Stat(ctx, "probe")
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}
	return nil
}

// ensureStorages loads the storage table once per process. The table is the
// truth; the environment is only its initial seed.
func (s *FileService) ensureStorages(ctx context.Context) error {
	s.storageMu.Lock()
	defer s.storageMu.Unlock()
	if s.storagesLoaded {
		return s.storagesErr
	}
	cfgs, err := LoadStorageConfigs(context.WithoutCancel(ctx), s.repo)
	if err != nil {
		s.storagesErr = err
		return err
	}
	if err := s.storages.Reload(ctx, cfgs); err != nil {
		s.storagesErr = err
		return err
	}
	s.storagesLoaded = true
	s.storagesErr = nil
	return nil
}

// LoadStorageConfigs reads the storages table and seeds the configured default
// when it is empty. Boot-time wiring and the first write both go through here,
// so they cannot disagree about which store is default.
func LoadStorageConfigs(ctx context.Context, repo Repository) ([]storage.Config, error) {
	rows, err := repo.ListStorages(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		seeded := seedStorages()
		for i := range seeded {
			if err := repo.SaveStorage(ctx, &seeded[i]); err != nil {
				return nil, err
			}
		}
		rows = seeded
	}
	cfgs := make([]storage.Config, 0, len(rows))
	for _, row := range rows {
		cfgs = append(cfgs, storage.Config{
			ID:      row.ID,
			Driver:  row.Driver,
			Name:    row.Name,
			Default: row.IsDefault,
			Config:  row.Config,
		})
	}
	return cfgs, nil
}

// ReloadStorages drops the process-wide storage snapshot and rebuilds it from
// the table. A failing reload keeps the previous snapshot serving.
func (s *FileService) ReloadStorages(ctx context.Context) error {
	s.storageMu.Lock()
	s.storagesLoaded = false
	s.storageMu.Unlock()
	return s.ensureStorages(ctx)
}

// LocalStorageID is the fixed id of the disk-backed store, seeded from
// configuration. Fixed ids keep the settings panel able to address rows.
const LocalStorageID = "local"

// seedStorages derives the initial row from configuration. It is the only
// place environment variables are read for storage.
func seedStorages() []fileModel.StorageConfig {
	cfg := config.Config().Storage
	if cfg.ObjectEnabled && cfg.Endpoint != "" {
		return []fileModel.StorageConfig{{
			ID:        S3StorageID,
			Driver:    storage.DriverS3,
			Name:      "s3",
			IsDefault: true,
			Config: map[string]string{
				"endpoint":       cfg.Endpoint,
				"access_key":     cfg.AccessKey,
				"secret_key":     cfg.SecretKey,
				"bucket":         cfg.BucketName,
				"region":         cfg.Region,
				"use_ssl":        fmt.Sprint(cfg.UseSSL),
				"use_path_style": fmt.Sprint(cfg.UsePathStyle),
				"path_prefix":    cfg.PathPrefix,
			},
		}, {
			ID:        LocalStorageID,
			Driver:    storage.DriverLocal,
			Name:      "local",
			IsDefault: false,
			Config:    map[string]string{"root": cfg.DataRoot},
		}}
	}
	return []fileModel.StorageConfig{{
		ID:        LocalStorageID,
		Driver:    storage.DriverLocal,
		Name:      "local",
		IsDefault: true,
		Config:    map[string]string{"root": cfg.DataRoot},
	}}
}

func principalOf(ctx context.Context) viewer.Context {
	return viewer.MustFromContext(ctx)
}

func dedupe(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func imgSize(fh *multipart.FileHeader) (int, int, error) {
	return imgUtil.GetImageSizeFromFile(fh)
}

func toDTO(f *fileModel.File, links []fileModel.Link) fileModel.FileDTO {
	out := fileModel.FileDTO{
		ID:          f.ID,
		Kind:        f.Kind,
		Access:      f.Access,
		Status:      f.Status,
		Name:        f.Name,
		ContentType: f.ContentType,
		Category:    fileModel.CategoryOf(f.ContentType),
		Size:        f.Size,
		Width:       f.Width,
		Height:      f.Height,
		URL:         fileModel.ContentPath(f.ID),
		UploaderID:  f.UploaderID,
		CreatedAt:   f.CreatedAt,
	}
	for _, link := range links {
		out.Links = append(out.Links, fileModel.LinkDTO{
			OwnerKind: link.OwnerKind,
			OwnerID:   link.OwnerID,
			Role:      link.Role,
			SortOrder: link.SortOrder,
		})
	}
	return out
}
