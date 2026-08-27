// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
)

func TestBuildS3PathURLResolver_Addressing(t *testing.T) {
	cases := []struct {
		name string
		cfg  config.StorageConfig
		path string
		want string
	}{
		{
			name: "other provider defaults to virtual-hosted (COS/OSS)",
			cfg:  config.StorageConfig{Provider: "other", Endpoint: "cos.ap-guangzhou.myqcloud.com", BucketName: "mybucket-125", UseSSL: true},
			path: "images/a.png",
			want: "https://mybucket-125.cos.ap-guangzhou.myqcloud.com/images/a.png",
		},
		{
			name: "other with use_path_style forces path-style",
			cfg:  config.StorageConfig{Provider: "other", Endpoint: "s3.selfhosted.example", BucketName: "b", UseSSL: true, UsePathStyle: true},
			path: "images/a.png",
			want: "https://s3.selfhosted.example/b/images/a.png",
		},
		{
			name: "minio stays path-style",
			cfg:  config.StorageConfig{Provider: "minio", Endpoint: "minio.local:9000", BucketName: "b", UseSSL: false},
			path: "images/a.png",
			want: "http://minio.local:9000/b/images/a.png",
		},
		{
			name: "r2 stays path-style",
			cfg:  config.StorageConfig{Provider: "r2", Endpoint: "acc.r2.cloudflarestorage.com", BucketName: "b", UseSSL: true},
			path: "images/a.png",
			want: "https://acc.r2.cloudflarestorage.com/b/images/a.png",
		},
		{
			name: "cdn domain takes precedence over addressing style",
			cfg:  config.StorageConfig{Provider: "other", Endpoint: "cos.ap-guangzhou.myqcloud.com", BucketName: "b", CDNURL: "https://cdn.example.com", UseSSL: true},
			path: "images/a.png",
			want: "https://cdn.example.com/images/a.png",
		},
		{
			name: "path prefix is applied under virtual-hosted",
			cfg:  config.StorageConfig{Provider: "other", Endpoint: "cos.ap-guangzhou.myqcloud.com", BucketName: "mybucket", PathPrefix: "uploads", UseSSL: true},
			path: "images/a.png",
			want: "https://mybucket.cos.ap-guangzhou.myqcloud.com/uploads/images/a.png",
		},
		{
			name: "empty endpoint falls back to relative path-style shape",
			cfg:  config.StorageConfig{Provider: "other", Endpoint: "", BucketName: "b", UseSSL: true},
			path: "images/a.png",
			want: "/b/images/a.png",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildS3PathURLResolver(tc.cfg)(tc.path)
			if got != tc.want {
				t.Fatalf("URL mismatch\n got: %s\nwant: %s", got, tc.want)
			}
		})
	}
}
