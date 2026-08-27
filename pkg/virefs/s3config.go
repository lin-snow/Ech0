// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Provider int

const (
	ProviderAWS Provider = iota
	ProviderMinIO
	ProviderR2
)

type S3Config struct {
	Region string

	Endpoint string

	Bucket string

	AccessKey string
	SecretKey string

	Provider Provider

	UsePathStyle *bool
}

func NewS3Client(ctx context.Context, cfg *S3Config) (*s3.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("virefs: S3Config must not be nil")
	}
	resolved := *cfg
	applyProviderDefaults(&resolved)

	var loadOpts []func(*config.LoadOptions) error

	if resolved.Region != "" {
		loadOpts = append(loadOpts, config.WithRegion(resolved.Region))
	}

	if resolved.AccessKey != "" && resolved.SecretKey != "" {
		loadOpts = append(loadOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(resolved.AccessKey, resolved.SecretKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("virefs: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if resolved.Endpoint != "" {
			o.BaseEndpoint = aws.String(resolved.Endpoint)
		}
		if resolved.UsePathStyle != nil && *resolved.UsePathStyle {
			o.UsePathStyle = true
		}
		if resolved.Provider != ProviderAWS || resolved.Endpoint != "" {
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		}
	})

	return client, nil
}

func NewObjectFSFromConfig(ctx context.Context, cfg *S3Config, opts ...ObjectOption) (*ObjectFS, error) {
	if cfg == nil {
		return nil, fmt.Errorf("virefs: S3Config must not be nil")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("virefs: S3Config.Bucket must not be empty")
	}

	client, err := NewS3Client(ctx, cfg)
	if err != nil {
		return nil, err
	}

	presignClient := s3.NewPresignClient(client)

	allOpts := make([]ObjectOption, 0, len(opts)+1)
	allOpts = append(allOpts, WithPresignClient(presignClient))
	allOpts = append(allOpts, opts...)

	return NewObjectFS(client, cfg.Bucket, allOpts...), nil
}

func applyProviderDefaults(cfg *S3Config) {
	switch cfg.Provider {
	case ProviderMinIO:
		if cfg.UsePathStyle == nil {
			cfg.UsePathStyle = aws.Bool(true)
		}
	case ProviderR2:
		if cfg.UsePathStyle == nil {
			cfg.UsePathStyle = aws.Bool(true)
		}
		if cfg.Region == "" {
			cfg.Region = "auto"
		}
	case ProviderAWS:
		if cfg.Region == "" {
			cfg.Region = "us-east-1"
		}
	}
}
