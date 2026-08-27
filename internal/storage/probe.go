// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lin-snow/ech0/internal/config"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/pkg/virefs"
)

var errBucketRequired = errors.New("S3 Bucket 不能为空")

func (m *Manager) TestS3Connection(ctx context.Context, setting settingModel.S3Setting) error {
	return probeS3(ctx, storageConfigFromSetting(setting, m.defaultCfg))
}

func probeS3(ctx context.Context, cfg config.StorageConfig) error {
	if cfg.BucketName == "" {
		return errBucketRequired
	}

	client, err := virefs.NewS3Client(ctx, virefsS3ConfigFromStorage(cfg))
	if err != nil {
		return err
	}

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.BucketName)})
	return err
}
