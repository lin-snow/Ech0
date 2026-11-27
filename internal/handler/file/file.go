// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/apperr"
	res "github.com/lin-snow/ech0/internal/handler/response"
	"github.com/lin-snow/ech0/internal/httperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	service "github.com/lin-snow/ech0/internal/service/file"
)

// FileHandler is transport only. Every decision — who may read, what may be
// cached, which storage answers — belongs to the service; this file parses the
// request, calls it, and writes the answer.
type FileHandler struct {
	files service.Service
}

func NewFileHandler(files service.Service) *FileHandler {
	return &FileHandler{files: files}
}

type (
	CommitInput struct {
		ID string `path:"id" doc:"文件 ID"`
	}
	GetFileInput struct {
		ID string `path:"id" doc:"文件 ID"`
	}
	ListFilesInput struct {
		Page      int    `query:"page" doc:"页码，从 1 开始"`
		PageSize  int    `query:"pageSize" doc:"每页数量，1-100"`
		Search    string `query:"search" doc:"文件名称模糊匹配"`
		Kind      string `query:"kind" doc:"owned / blob / ref"`
		Status    string `query:"status" doc:"pending / ready"`
		Access    string `query:"access" doc:"public / owner / private"`
		OwnerKind string `query:"ownerKind" doc:"引用方类型，如 echo"`
		OwnerID   string `query:"ownerId" doc:"引用方 ID"`
		Role      string `query:"role" doc:"引用角色，如 attachment"`
	}
	RenameFileInput struct {
		ID   string `path:"id" doc:"文件 ID"`
		Body struct {
			Name string `json:"name" doc:"展示名称（可为空）"`
		}
	}
	DeleteFileInput struct {
		ID    string `path:"id" doc:"文件 ID"`
		Force bool   `query:"force" doc:"管理员强制删除仍有引用的文件"`
	}
)

type (
	FileOutput   = struct{ Body fileModel.FileDTO }
	FileListPage = struct {
		Body commonModel.List[fileModel.FileDTO]
	}
	// FileCreatedOutput adds the canonical URI of the committed file, so a 201
	// can point at it (RFC 9110 6.4.2).
	FileCreatedOutput struct {
		Location string `header:"Location"`
		Body     fileModel.FileDTO
	}
	// UploadOutput is the raw multipart answer, not a Huma output: the upload
	// route renders it through response.Created, which Huma never sees.
	UploadOutput = service.CreateResult
	EmptyOutput  = struct{}
)

func (fileHandler *FileHandler) Commit(ctx context.Context, in *CommitInput) (*FileCreatedOutput, error) {
	dto, err := fileHandler.files.Commit(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &FileCreatedOutput{Location: "/api/files/" + in.ID, Body: dto}, nil
}

func (fileHandler *FileHandler) Get(ctx context.Context, in *GetFileInput) (*FileOutput, error) {
	dto, err := fileHandler.files.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &FileOutput{Body: dto}, nil
}

func (fileHandler *FileHandler) List(ctx context.Context, in *ListFilesInput) (*FileListPage, error) {
	page, err := fileHandler.files.Browse(ctx, service.BrowseInput{
		Page:     in.Page,
		PageSize: in.PageSize,
		Search:   in.Search,
		Kind:     in.Kind,
		Status:   in.Status,
		Access:   in.Access,
		Owner:    in.OwnerKind,
		OwnerID:  in.OwnerID,
		Role:     in.Role,
	})
	if err != nil {
		return nil, err
	}
	total := page.Total
	return &FileListPage{Body: commonModel.List[fileModel.FileDTO]{Items: page.Items, Total: &total}}, nil
}

func (fileHandler *FileHandler) Rename(ctx context.Context, in *RenameFileInput) (*FileOutput, error) {
	dto, err := fileHandler.files.Rename(ctx, in.ID, in.Body.Name)
	if err != nil {
		return nil, err
	}
	return &FileOutput{Body: dto}, nil
}

func (fileHandler *FileHandler) Delete(ctx context.Context, in *DeleteFileInput) (*EmptyOutput, error) {
	if _, err := fileHandler.files.Delete(ctx, in.ID, in.Force); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

// Upload is the one multipart endpoint. Bytes, a reference, or a direct-upload
// request all arrive here; `profile` picks the validation and access rule.
func (fileHandler *FileHandler) Upload() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		in := service.CreateInput{
			Profile:     strings.TrimSpace(ctx.PostForm("profile")),
			ExternalURL: strings.TrimSpace(ctx.PostForm("external_url")),
			Name:        ctx.PostForm("name"),
			ContentType: strings.TrimSpace(ctx.PostForm("content_type")),
			Direct:      ctx.PostForm("mode") == "direct",
			Width:       atoiOrZero(ctx.PostForm("width")),
			Height:      atoiOrZero(ctx.PostForm("height")),
		}
		if fh, err := ctx.FormFile("file"); err == nil {
			in.File = fh
		} else if in.ExternalURL == "" && !in.Direct {
			return res.Failed(apperr.ErrInvalidRequestBody)
		}
		created, err := fileHandler.files.Create(ctx.Request.Context(), in)
		if err != nil {
			return res.Failed(err)
		}
		out := *created
		return res.Created(out)
	})
}

func atoiOrZero(raw string) int {
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// Content is the only byte exit. GET and HEAD share it so a HEAD can never
// answer differently about existence than a GET (#7).
func (fileHandler *FileHandler) Content(ctx *gin.Context) {
	id := strings.TrimSpace(ctx.Param("id"))
	if id == "" {
		writeStatus(ctx, http.StatusNotFound)
		return
	}
	rng, ok := parseRange(ctx.GetHeader("Range"))
	if !ok {
		ctx.Header("Content-Range", "bytes */0")
		writeStatus(ctx, http.StatusRequestedRangeNotSatisfiable)
		return
	}
	opened, err := fileHandler.files.Open(ctx.Request.Context(), id, rng)
	if err != nil {
		fileHandler.writeError(ctx, err)
		return
	}
	if opened.Reader != nil {
		defer func() { _ = opened.Reader.Close() }()
	}
	if opened.Redirect != "" {
		// A signed or external URL is a one-off capability: never cache the
		// redirect that carries it.
		ctx.Header("Cache-Control", "no-store")
		ctx.Redirect(http.StatusFound, opened.Redirect)
		return
	}

	etag := `"` + id + `"`
	ctx.Header("ETag", etag)
	ctx.Header("Cache-Control", opened.CacheControl)
	if opened.VaryAuthorization {
		ctx.Header("Vary", "Authorization")
	}
	if ctx.GetHeader("If-None-Match") == etag {
		writeStatus(ctx, http.StatusNotModified)
		return
	}
	if opened.ContentType != "" {
		ctx.Header("Content-Type", opened.ContentType)
	}
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Content-Disposition", disposition(inlineable(opened.ContentType), opened.File.Name))
	if opened.Rangeable {
		ctx.Header("Accept-Ranges", "bytes")
	}
	status := http.StatusOK
	if rng != nil {
		status = http.StatusPartialContent
		last := rng.Start + opened.Size - 1
		total := opened.Total
		if total <= 0 {
			total = opened.Size
		}
		ctx.Header("Content-Range", "bytes "+strconv.FormatInt(rng.Start, 10)+"-"+
			strconv.FormatInt(last, 10)+"/"+strconv.FormatInt(total, 10))
	}
	if !opened.ModTime.IsZero() {
		ctx.Header("Last-Modified", opened.ModTime.UTC().Format(http.TimeFormat))
	}
	writeStatus(ctx, status)
	if ctx.Request.Method == http.MethodHead || opened.Reader == nil {
		return
	}
	_, _ = io.Copy(ctx.Writer, opened.Reader)
}

// writeStatus sets the code and flushes headers. gin's WriteHeader only records
// the code; a body-less answer (HEAD, 304, 416) must flush it explicitly.
func writeStatus(ctx *gin.Context, status int) {
	ctx.Status(status)
	ctx.Writer.WriteHeaderNow()
}

func (fileHandler *FileHandler) writeError(ctx *gin.Context, err error) {
	httperr.Write(ctx, err)
}

// parseRange understands the single byte range a media element needs:
// "bytes=start-end", "bytes=start-", and the suffix form "bytes=-length".
// Anything else, or a multi-range request, is refused rather than guessed at.
func parseRange(header string) (*service.ByteRange, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil, true
	}
	raw, ok := strings.CutPrefix(header, "bytes=")
	if !ok || strings.Contains(raw, ",") {
		return nil, false
	}
	startRaw, endRaw, ok := strings.Cut(raw, "-")
	if !ok {
		return nil, false
	}
	if startRaw == "" {
		length, err := strconv.ParseInt(endRaw, 10, 64)
		if err != nil || length <= 0 {
			return nil, false
		}
		return &service.ByteRange{Suffix: length}, true
	}
	start, err := strconv.ParseInt(startRaw, 10, 64)
	if err != nil || start < 0 {
		return nil, false
	}
	if endRaw == "" {
		return &service.ByteRange{Start: start}, true
	}
	end, err := strconv.ParseInt(endRaw, 10, 64)
	if err != nil || end < start {
		return nil, false
	}
	return &service.ByteRange{Start: start, Length: end - start + 1}, true
}

func inlineable(contentType string) bool {
	base, _, _ := strings.Cut(contentType, ";")
	switch strings.TrimSpace(base) {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/avif",
		"audio/mpeg", "audio/ogg", "video/mp4", "video/webm",
		"application/pdf", "text/plain":
		return true
	}
	return false
}

func disposition(inline bool, name string) string {
	kind := "attachment"
	if inline {
		kind = "inline"
	}
	if name == "" {
		return kind
	}
	safe := strings.Map(func(r rune) rune {
		switch r {
		case '"', '\\', '\r', '\n':
			return -1
		}
		return r
	}, name)
	return kind + `; filename="` + safe + `"`
}
