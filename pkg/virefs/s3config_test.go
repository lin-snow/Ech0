// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestApplyProviderDefaults_AWS(t *testing.T) {
	cfg := S3Config{Provider: ProviderAWS}
	applyProviderDefaults(&cfg)
	if cfg.Region != "us-east-1" {
		t.Fatalf("AWS default region = %q, want %q", cfg.Region, "us-east-1")
	}
	if cfg.UsePathStyle != nil {
		t.Fatal("AWS should not force path style")
	}
}

func TestApplyProviderDefaults_AWS_RegionPreserved(t *testing.T) {
	cfg := S3Config{Provider: ProviderAWS, Region: "eu-west-1"}
	applyProviderDefaults(&cfg)
	if cfg.Region != "eu-west-1" {
		t.Fatalf("region = %q, want %q", cfg.Region, "eu-west-1")
	}
}

func TestApplyProviderDefaults_MinIO(t *testing.T) {
	cfg := S3Config{Provider: ProviderMinIO}
	applyProviderDefaults(&cfg)
	if cfg.UsePathStyle == nil || !*cfg.UsePathStyle {
		t.Fatal("MinIO should default to path style")
	}
}

func TestApplyProviderDefaults_MinIO_PathStyleOverride(t *testing.T) {
	cfg := S3Config{
		Provider:     ProviderMinIO,
		UsePathStyle: aws.Bool(false),
	}
	applyProviderDefaults(&cfg)
	if *cfg.UsePathStyle != false {
		t.Fatal("explicit UsePathStyle=false should be preserved")
	}
}

func TestApplyProviderDefaults_R2(t *testing.T) {
	cfg := S3Config{Provider: ProviderR2}
	applyProviderDefaults(&cfg)
	if cfg.Region != "auto" {
		t.Fatalf("R2 default region = %q, want %q", cfg.Region, "auto")
	}
	if cfg.UsePathStyle == nil || !*cfg.UsePathStyle {
		t.Fatal("R2 should default to path style")
	}
}

func TestApplyProviderDefaults_R2_RegionPreserved(t *testing.T) {
	cfg := S3Config{Provider: ProviderR2, Region: "wnam"}
	applyProviderDefaults(&cfg)
	if cfg.Region != "wnam" {
		t.Fatalf("region = %q, want %q", cfg.Region, "wnam")
	}
}

func TestS3Config_Validation(t *testing.T) {
	_, err := NewObjectFSFromConfig(t.Context(), &S3Config{})
	if err == nil {
		t.Fatal("NewObjectFSFromConfig with empty Bucket should fail")
	}
}

func TestNewS3Client_ChecksumBehavior(t *testing.T) {
	tests := []struct {
		name             string
		cfg              S3Config
		wantWhenRequired bool
	}{
		{
			name:             "MinIO",
			cfg:              S3Config{Provider: ProviderMinIO, Region: "us-east-1", Endpoint: "http://localhost:9000"},
			wantWhenRequired: true,
		},
		{
			name:             "R2",
			cfg:              S3Config{Provider: ProviderR2, Endpoint: "https://acct.r2.cloudflarestorage.com"},
			wantWhenRequired: true,
		},
		{
			name:             "OtherViaCustomEndpoint",
			cfg:              S3Config{Provider: ProviderAWS, Endpoint: "https://s3.example.com"},
			wantWhenRequired: true,
		},
		{
			name:             "RealAWS",
			cfg:              S3Config{Provider: ProviderAWS, Region: "us-east-1"},
			wantWhenRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			cfg.AccessKey, cfg.SecretKey = "x", "y"
			client, err := NewS3Client(t.Context(), &cfg)
			if err != nil {
				t.Fatalf("NewS3Client failed: %v", err)
			}

			wantReq := aws.RequestChecksumCalculationWhenSupported
			wantResp := aws.ResponseChecksumValidationWhenSupported
			if tt.wantWhenRequired {
				wantReq = aws.RequestChecksumCalculationWhenRequired
				wantResp = aws.ResponseChecksumValidationWhenRequired
			}

			if got := client.Options().RequestChecksumCalculation; got != wantReq {
				t.Errorf("RequestChecksumCalculation = %v, want %v", got, wantReq)
			}
			if got := client.Options().ResponseChecksumValidation; got != wantResp {
				t.Errorf("ResponseChecksumValidation = %v, want %v", got, wantResp)
			}
		})
	}
}
