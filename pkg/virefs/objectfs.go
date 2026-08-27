// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
}

type PresignAPI interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type ObjectOption func(*ObjectFS)

func WithPrefix(prefix string) ObjectOption {
	return func(o *ObjectFS) { o.basePrefix = prefix }
}

func WithObjectKeyFunc(fn KeyFunc) ObjectOption {
	return func(o *ObjectFS) { o.keyFunc = fn }
}

func WithPresignClient(pc PresignAPI) ObjectOption {
	return func(o *ObjectFS) { o.presignClient = pc }
}

func WithBaseURL(url string) ObjectOption {
	return func(o *ObjectFS) { o.baseURL = strings.TrimRight(url, "/") }
}

func WithAccessExpires(d time.Duration) ObjectOption {
	return func(o *ObjectFS) { o.accessExpires = d }
}

func WithAccessFunc(fn AccessFunc) ObjectOption {
	return func(o *ObjectFS) { o.accessFunc = fn }
}

type ObjectFS struct {
	client        S3API
	bucket        string
	basePrefix    string
	keyFunc       KeyFunc
	presignClient PresignAPI
	baseURL       string
	accessExpires time.Duration
	accessFunc    AccessFunc
}

const defaultAccessExpires = 15 * time.Minute

func NewObjectFS(client S3API, bucket string, opts ...ObjectOption) *ObjectFS {
	o := &ObjectFS{
		client: client,
		bucket: bucket,
	}
	for _, fn := range opts {
		fn(o)
	}
	return o
}

func (o *ObjectFS) s3Key(key string) (string, error) {
	cleaned, err := CleanKey(key)
	if err != nil {
		return "", err
	}
	if o.keyFunc != nil {
		cleaned = o.keyFunc(cleaned)
	}
	return o.basePrefix + cleaned, nil
}

func (o *ObjectFS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	s3k, err := o.s3Key(key)
	if err != nil {
		return nil, &OpError{Op: "Get", Key: key, Err: err}
	}
	out, err := o.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
	})
	if err != nil {
		return nil, &OpError{Op: "Get", Key: key, Err: mapS3Error(err)}
	}
	return out.Body, nil
}

func (o *ObjectFS) Put(ctx context.Context, key string, r io.Reader, opts ...PutOption) error {
	s3k, err := o.s3Key(key)
	if err != nil {
		return &OpError{Op: "Put", Key: key, Err: err}
	}
	cfg := BuildPutConfig(opts)
	input := &s3.PutObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
		Body:   r,
	}
	if cfg.ContentType != "" {
		input.ContentType = aws.String(cfg.ContentType)
	}
	if len(cfg.Metadata) > 0 {
		input.Metadata = cfg.Metadata
	}
	_, err = o.client.PutObject(ctx, input)
	if err != nil {
		return &OpError{Op: "Put", Key: key, Err: mapS3Error(err)}
	}
	return nil
}

func (o *ObjectFS) Delete(ctx context.Context, key string) error {
	s3k, err := o.s3Key(key)
	if err != nil {
		return &OpError{Op: "Delete", Key: key, Err: err}
	}
	_, err = o.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
	})
	if err != nil {
		return &OpError{Op: "Delete", Key: key, Err: mapS3Error(err)}
	}
	return nil
}

func (o *ObjectFS) List(ctx context.Context, prefix string) (*ListResult, error) {
	cleanedPrefix, err := CleanKey(prefix)
	if err != nil {
		return nil, &OpError{Op: "List", Key: prefix, Err: err}
	}
	s3Prefix := o.basePrefix + cleanedPrefix
	if s3Prefix != "" && s3Prefix[len(s3Prefix)-1] != '/' {
		s3Prefix += "/"
	}

	delimiter := "/"
	result := &ListResult{}
	var continuationToken *string
	for {
		if err := ctx.Err(); err != nil {
			return nil, &OpError{Op: "List", Key: prefix, Err: err}
		}
		out, err := o.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(o.bucket),
			Prefix:            aws.String(s3Prefix),
			Delimiter:         aws.String(delimiter),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, &OpError{Op: "List", Key: prefix, Err: mapS3Error(err)}
		}
		for _, obj := range out.Contents {
			k := aws.ToString(obj.Key)
			if len(k) > len(o.basePrefix) {
				k = k[len(o.basePrefix):]
			}
			result.Files = append(result.Files, FileInfo{
				Key:          k,
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
		for _, cp := range out.CommonPrefixes {
			k := aws.ToString(cp.Prefix)
			if len(k) > len(o.basePrefix) {
				k = k[len(o.basePrefix):]
			}
			k = strings.TrimSuffix(k, "/")
			result.Files = append(result.Files, FileInfo{
				Key:   k,
				IsDir: true,
			})
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		continuationToken = out.NextContinuationToken
	}
	return result, nil
}

func (o *ObjectFS) Stat(ctx context.Context, key string) (*FileInfo, error) {
	s3k, err := o.s3Key(key)
	if err != nil {
		return nil, &OpError{Op: "Stat", Key: key, Err: err}
	}
	out, err := o.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
	})
	if err != nil {
		return nil, &OpError{Op: "Stat", Key: key, Err: mapS3Error(err)}
	}
	return &FileInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		LastModified: aws.ToTime(out.LastModified),
		ContentType:  aws.ToString(out.ContentType),
	}, nil
}

func (o *ObjectFS) Exists(ctx context.Context, key string) (bool, error) {
	s3k, err := o.s3Key(key)
	if err != nil {
		return false, &OpError{Op: "Exists", Key: key, Err: err}
	}
	_, err = o.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(mapS3Error(err), ErrNotFound) {
		return false, nil
	}
	return false, &OpError{Op: "Exists", Key: key, Err: mapS3Error(err)}
}

func (o *ObjectFS) PresignGet(ctx context.Context, key string, expires time.Duration) (*PresignedRequest, error) {
	if o.presignClient == nil {
		return nil, &OpError{Op: "PresignGet", Key: key, Err: fmt.Errorf("%w: presign client not configured", ErrNotSupported)}
	}
	s3k, err := o.s3Key(key)
	if err != nil {
		return nil, &OpError{Op: "PresignGet", Key: key, Err: err}
	}
	req, err := o.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return nil, &OpError{Op: "PresignGet", Key: key, Err: err}
	}
	return &PresignedRequest{
		URL:    req.URL,
		Method: req.Method,
		Header: req.SignedHeader,
	}, nil
}

func (o *ObjectFS) PresignPut(ctx context.Context, key string, expires time.Duration) (*PresignedRequest, error) {
	if o.presignClient == nil {
		return nil, &OpError{Op: "PresignPut", Key: key, Err: fmt.Errorf("%w: presign client not configured", ErrNotSupported)}
	}
	s3k, err := o.s3Key(key)
	if err != nil {
		return nil, &OpError{Op: "PresignPut", Key: key, Err: err}
	}
	req, err := o.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(s3k),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return nil, &OpError{Op: "PresignPut", Key: key, Err: err}
	}
	return &PresignedRequest{
		URL:    req.URL,
		Method: req.Method,
		Header: req.SignedHeader,
	}, nil
}

func (o *ObjectFS) Access(ctx context.Context, key string) (*AccessInfo, error) {
	s3k, err := o.s3Key(key)
	if err != nil {
		return nil, &OpError{Op: "Access", Key: key, Err: err}
	}

	if o.accessFunc != nil {
		if info := o.accessFunc(s3k); info != nil {
			return info, nil
		}
	}

	if o.presignClient != nil {
		expires := o.accessExpires
		if expires == 0 {
			expires = defaultAccessExpires
		}
		req, err := o.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(o.bucket),
			Key:    aws.String(s3k),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = expires
		})
		if err != nil {
			return nil, &OpError{Op: "Access", Key: key, Err: err}
		}
		return &AccessInfo{URL: req.URL}, nil
	}

	if o.baseURL != "" {
		return &AccessInfo{URL: o.baseURL + "/" + s3k}, nil
	}

	return nil, &OpError{Op: "Access", Key: key, Err: fmt.Errorf("%w: set WithAccessFunc, WithPresignClient, or WithBaseURL", ErrNotSupported)}
}

func (o *ObjectFS) Copy(ctx context.Context, srcKey, dstKey string) error {
	srcS3k, err := o.s3Key(srcKey)
	if err != nil {
		return &OpError{Op: "Copy", Key: srcKey, Err: err}
	}
	dstS3k, err := o.s3Key(dstKey)
	if err != nil {
		return &OpError{Op: "Copy", Key: dstKey, Err: err}
	}
	_, err = o.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(o.bucket),
		CopySource: aws.String(o.bucket + "/" + srcS3k),
		Key:        aws.String(dstS3k),
	})
	if err != nil {
		return &OpError{Op: "Copy", Key: srcKey + " -> " + dstKey, Err: mapS3Error(err)}
	}
	return nil
}

func (o *ObjectFS) BatchDelete(ctx context.Context, keys []string) error {
	const maxBatch = 1000
	for i := 0; i < len(keys); i += maxBatch {
		if err := ctx.Err(); err != nil {
			return &OpError{Op: "BatchDelete", Key: fmt.Sprintf("%d keys remaining", len(keys)-i), Err: err}
		}
		end := min(i+maxBatch, len(keys))
		objects := make([]types.ObjectIdentifier, 0, end-i)
		for _, key := range keys[i:end] {
			s3k, err := o.s3Key(key)
			if err != nil {
				return &OpError{Op: "BatchDelete", Key: key, Err: err}
			}
			objects = append(objects, types.ObjectIdentifier{Key: aws.String(s3k)})
		}
		out, err := o.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(o.bucket),
			Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return &OpError{Op: "BatchDelete", Key: fmt.Sprintf("%d keys", end-i), Err: mapS3Error(err)}
		}
		if len(out.Errors) > 0 {
			e := out.Errors[0]
			return &OpError{
				Op:  "BatchDelete",
				Key: aws.ToString(e.Key),
				Err: fmt.Errorf("s3 error %s: %s", aws.ToString(e.Code), aws.ToString(e.Message)),
			}
		}
	}
	return nil
}

var (
	_ FS           = (*ObjectFS)(nil)
	_ Presigner    = (*ObjectFS)(nil)
	_ Copier       = (*ObjectFS)(nil)
	_ BatchDeleter = (*ObjectFS)(nil)
)

func mapS3Error(err error) error {
	if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
		return ErrNotFound
	}
	if _, ok := errors.AsType[*types.NotFound](err); ok {
		return ErrNotFound
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		if apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey" {
			return ErrNotFound
		}
	}
	return err
}
