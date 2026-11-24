// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/httperr"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	service "github.com/lin-snow/ech0/internal/service/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

// stubService inherits every Service method as a nil-returning trap; the tests
// below only ever reach the methods they set.
type stubService struct {
	service.Service
	open   func(ctx context.Context, id string, rng *service.ByteRange) (*service.OpenResult, error)
	browse func(ctx context.Context, in service.BrowseInput) (fileModel.Page, error)
	create func(ctx context.Context, in service.CreateInput) (*service.CreateResult, error)
}

func (s stubService) Open(ctx context.Context, id string, rng *service.ByteRange) (*service.OpenResult, error) {
	return s.open(ctx, id, rng)
}

func (s stubService) Browse(ctx context.Context, in service.BrowseInput) (fileModel.Page, error) {
	return s.browse(ctx, in)
}

func (s stubService) Create(ctx context.Context, in service.CreateInput) (*service.CreateResult, error) {
	return s.create(ctx, in)
}

// postForm runs a raw route against an urlencoded body.
func postForm(t *testing.T, h gin.HandlerFunc, target string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h(c)
	return rec
}

func serveContent(t *testing.T, svc service.Service, method, target, rangeHeader, ifNoneMatch string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, nil)
	if rangeHeader != "" {
		c.Request.Header.Set("Range", rangeHeader)
	}
	if ifNoneMatch != "" {
		c.Request.Header.Set("If-None-Match", ifNoneMatch)
	}
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	NewFileHandler(svc).Content(c)
	return rec
}

func TestContentUnknownFileAnswers404(t *testing.T) {
	svc := stubService{open: func(context.Context, string, *service.ByteRange) (*service.OpenResult, error) {
		return nil, service.ErrNotFound
	}}
	rec := serveContent(t, svc, http.MethodGet, "/api/files/abc/content", "", "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get("ETag"), "a miss must not disclose an identity")
}

func TestContentRedirectIsNeverCached(t *testing.T) {
	svc := stubService{open: func(context.Context, string, *service.ByteRange) (*service.OpenResult, error) {
		return &service.OpenResult{Redirect: "https://cdn.example/a.png?sig=1"}, nil
	}}
	rec := serveContent(t, svc, http.MethodGet, "/api/files/abc/content", "", "")
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "https://cdn.example/a.png?sig=1", rec.Header().Get("Location"))
}

func TestContentPublicBytesAreImmutableAndRevalidateById(t *testing.T) {
	body := []byte("png-bytes")
	svc := stubService{open: func(_ context.Context, id string, rng *service.ByteRange) (*service.OpenResult, error) {
		assert.Nil(t, rng)
		return &service.OpenResult{
			File:         fileModel.FileDTO{ID: id, Name: "a.png"},
			Reader:       io.NopCloser(bytes.NewReader(body)),
			Size:         int64(len(body)),
			ContentType:  "image/png",
			CacheControl: "public, max-age=31536000, immutable",
			Rangeable:    true,
		}, nil
	}}

	rec := serveContent(t, svc, http.MethodGet, "/api/files/abc/content", "", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, body, rec.Body.Bytes())
	assert.Equal(t, `"abc"`, rec.Header().Get("ETag"))
	assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, `inline; filename="a.png"`, rec.Header().Get("Content-Disposition"))
	assert.Equal(t, "bytes", rec.Header().Get("Accept-Ranges"))

	rec = serveContent(t, svc, http.MethodGet, "/api/files/abc/content", "", `"abc"`)
	assert.Equal(t, http.StatusNotModified, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestContentRangeServesPartialBytes(t *testing.T) {
	body := []byte("0123456789")
	svc := stubService{open: func(_ context.Context, _ string, rng *service.ByteRange) (*service.OpenResult, error) {
		start, length := int64(0), int64(len(body))
		if rng != nil {
			if rng.Start < 0 || rng.Start >= int64(len(body)) {
				return nil, service.ErrRangeNotSatisfiable
			}
			start, length = rng.Start, rng.Length
			if length <= 0 || start+length > int64(len(body)) {
				length = int64(len(body)) - start
			}
		}
		return &service.OpenResult{
			File:         fileModel.FileDTO{ID: "abc", Name: "a.bin"},
			Reader:       io.NopCloser(bytes.NewReader(body[start : start+length])),
			Size:         length,
			Total:        int64(len(body)),
			ContentType:  "application/octet-stream",
			CacheControl: "public, max-age=31536000, immutable",
			Rangeable:    true,
		}, nil
	}}

	rec := serveContent(t, svc, http.MethodGet, "/api/files/abc/content", "bytes=2-5", "")
	require.Equal(t, http.StatusPartialContent, rec.Code)
	assert.Equal(t, []byte("2345"), rec.Body.Bytes())
	assert.Equal(t, "bytes 2-5/10", rec.Header().Get("Content-Range"))
	assert.Equal(t, "attachment; filename=\"a.bin\"", rec.Header().Get("Content-Disposition"))

	rec = serveContent(t, svc, http.MethodHead, "/api/files/abc/content", "", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestContentRejectsUnparseableRange(t *testing.T) {
	svc := stubService{open: func(context.Context, string, *service.ByteRange) (*service.OpenResult, error) {
		t.Fatal("service must not be consulted for a malformed range")
		return nil, nil
	}}
	rec := serveContent(t, svc, http.MethodGet, "/api/files/abc/content", "bytes=0-1,4-5", "")
	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rec.Code)
}

func TestListCarriesOffsetTotal(t *testing.T) {
	svc := stubService{browse: func(_ context.Context, in service.BrowseInput) (fileModel.Page, error) {
		require.Equal(t, 2, in.Page)
		require.Equal(t, 10, in.PageSize)
		return fileModel.Page{Total: 7, Items: []fileModel.FileDTO{{ID: "f1"}}}, nil
	}}

	out, err := NewFileHandler(svc).List(context.Background(), &ListFilesInput{Page: 2, PageSize: 10})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Body.Total)
	assert.Equal(t, int64(7), *out.Body.Total)
	assert.Len(t, out.Body.Items, 1)
}

func TestListMapsDomainErrors(t *testing.T) {
	svc := stubService{browse: func(context.Context, service.BrowseInput) (fileModel.Page, error) {
		return fileModel.Page{}, service.ErrInvalid
	}}

	out, err := NewFileHandler(svc).List(context.Background(), &ListFilesInput{})
	require.ErrorIs(t, err, service.ErrInvalid)
	assert.Nil(t, out)
	assert.Equal(t, http.StatusBadRequest, httperr.From(err).Status)
}

func TestUploadAnswers201WithTheFile(t *testing.T) {
	svc := stubService{create: func(_ context.Context, in service.CreateInput) (*service.CreateResult, error) {
		require.Equal(t, "https://cdn.example/a.png", in.ExternalURL)
		return &service.CreateResult{File: fileModel.FileDTO{ID: "f1", Name: "a.png"}}, nil
	}}

	rec := postForm(t, NewFileHandler(svc).Upload(), "/api/files", url.Values{
		"profile":      {"image"},
		"external_url": {"https://cdn.example/a.png"},
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	var got service.CreateResult
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "f1", got.File.ID)
	assert.Equal(t, "a.png", got.File.Name)
}

func TestUploadMapsDomainStatuses(t *testing.T) {
	svc := stubService{create: func(context.Context, service.CreateInput) (*service.CreateResult, error) {
		return nil, service.ErrTooLarge
	}}

	rec := postForm(t, NewFileHandler(svc).Upload(), "/api/files", url.Values{
		"external_url": {"https://cdn.example/a.png"},
	})
	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")
}

func TestUploadRejectsMissingInput(t *testing.T) {
	svc := stubService{create: func(context.Context, service.CreateInput) (*service.CreateResult, error) {
		t.Fatal("service must not be consulted without a file, url or direct request")
		return nil, nil
	}}

	rec := postForm(t, NewFileHandler(svc).Upload(), "/api/files", url.Values{"profile": {"image"}})
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestParseRange(t *testing.T) {
	cases := []struct {
		header string
		want   *service.ByteRange
		ok     bool
	}{
		{header: "", want: nil, ok: true},
		{header: "bytes=0-99", want: &service.ByteRange{Start: 0, Length: 100}, ok: true},
		{header: "bytes=100-", want: &service.ByteRange{Start: 100}, ok: true},
		{header: "bytes=-500", want: &service.ByteRange{Suffix: 500}, ok: true},
		{header: "bytes=5-4", ok: false},
		{header: "bytes=-0", ok: false},
		{header: "items=0-1", ok: false},
		{header: "bytes=0-1,3-4", ok: false},
		{header: "bytes=abc-def", ok: false},
	}
	for _, tc := range cases {
		got, ok := parseRange(tc.header)
		assert.Equal(t, tc.ok, ok, tc.header)
		if tc.ok {
			assert.Equal(t, tc.want, got, tc.header)
		}
	}
}
