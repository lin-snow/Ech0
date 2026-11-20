// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestVerifiedDirectUpload(t *testing.T) {
	fix := newFileFix(t)
	fix.expectAdmin()
	cfg := config.Config()
	previous := cfg.Upload.VerifiedDirect
	cfg.Upload.VerifiedDirect = true
	t.Cleanup(func() { cfg.Upload.VerifiedDirect = previous })
	content := pngBytes(t, 2, 2)
	size := int64(len(content))
	var stored []byte
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			if r.Header.Get("If-None-Match") != "*" || r.ContentLength != size {
				w.WriteHeader(400)
				return
			}
			if stored != nil {
				w.WriteHeader(412)
				return
			}
			var err error
			stored, err = io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				return
			}
		case http.MethodHead, http.MethodGet:
			if stored == nil {
				w.WriteHeader(404)
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(len(stored)))
			w.Header().Set("Content-Type", "image/png")
			if r.Method == http.MethodGet {
				_, _ = w.Write(stored)
			}
		default:
			w.WriteHeader(405)
		}
	}))
	defer target.Close()
	require.NoError(t, fix.mgr.ApplyS3Setting(settingModel.S3Setting{Enable: true, Provider: "r2", Region: "auto", BucketName: "media", Endpoint: target.URL, UsePathStyle: true, AccessKey: "test", SecretKey: "test"}))
	ctx := fix.adminCtx()
	signed, err := fix.svc.GetFilePresignURL(ctx, &commonModel.GetPresignURLDto{FileName: "picture.png", ContentType: "image/png", Size: &size})
	require.NoError(t, err)
	require.True(t, signed.RequireCreateOnly)
	_, err = fix.repo.GetByID(ctx, signed.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	put, err := http.NewRequestWithContext(ctx, http.MethodPut, signed.PresignURL, bytes.NewReader(content))
	require.NoError(t, err)
	put.Header.Set("Content-Type", "image/png")
	put.Header.Set("If-None-Match", "*")
	resp, err := target.Client().Do(put)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)
	_, err = fix.svc.UpdateFileMeta(ctx, signed.ID, commonModel.UpdateFileMetaDto{Size: size + 1})
	require.Error(t, err)
	for range 2 {
		published, err := fix.svc.UpdateFileMeta(ctx, signed.ID, commonModel.UpdateFileMetaDto{Size: size})
		require.NoError(t, err)
		require.Equal(t, signed.ID, published.ID)
		require.Equal(t, size, published.Size)
	}
	_, err = fix.repo.GetIntent(ctx, signed.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
