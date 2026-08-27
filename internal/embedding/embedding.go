// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package embedding

import (
	"context"
	"errors"
	"fmt"
	"strings"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	openai "github.com/sashabaranov/go-openai"
)

var (
	ErrNotEnabled    = errors.New("embedding: not enabled")
	ErrModelMissing  = errors.New("embedding: model missing")
	ErrEmptyResponse = errors.New("embedding: empty response")
)

const defaultBatchSize = 64

func Embed(
	ctx context.Context,
	setting settingModel.EmbeddingSetting,
	inputs []string,
) ([][]float32, error) {
	if !setting.Enable {
		return nil, ErrNotEnabled
	}
	if setting.Model == "" {
		return nil, ErrModelMissing
	}
	if len(inputs) == 0 {
		return nil, nil
	}

	batchSize := setting.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	cfg := openai.DefaultConfig(setting.ApiKey)
	if setting.BaseURL != "" {
		cfg.BaseURL = setting.BaseURL
	}
	client := openai.NewClientWithConfig(cfg)

	sendDim := setting.Dim

	out := make([][]float32, 0, len(inputs))
	for start := 0; start < len(inputs); start += batchSize {
		end := min(start+batchSize, len(inputs))
		batch := inputs[start:end]

		resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model:      openai.EmbeddingModel(setting.Model),
			Input:      batch,
			Dimensions: sendDim,
		})
		if err != nil && sendDim != 0 && isDimensionsRejected(err) {
			sendDim = 0
			resp, err = client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
				Model: openai.EmbeddingModel(setting.Model),
				Input: batch,
			})
		}
		if err != nil {
			return nil, err
		}
		if len(resp.Data) != len(batch) {
			return nil, ErrEmptyResponse
		}
		for i := range resp.Data {
			vec := resp.Data[i].Embedding
			if setting.Dim > 0 && len(vec) != setting.Dim {
				return nil, fmt.Errorf(
					"embedding: 模型 %s 返回维度 %d，与配置维度 %d 不一致，"+
						"请调整 dim 或换用支持 dimensions 参数的模型",
					setting.Model, len(vec), setting.Dim,
				)
			}
			out = append(out, vec)
		}
	}
	return out, nil
}

func isDimensionsRejected(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "dimension")
}

func EmbedOne(
	ctx context.Context,
	setting settingModel.EmbeddingSetting,
	input string,
) ([]float32, error) {
	vecs, err := Embed(ctx, setting, []string{input})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, ErrEmptyResponse
	}
	return vecs[0], nil
}

type Client struct{}

func (Client) Embed(ctx context.Context, setting settingModel.EmbeddingSetting, inputs []string) ([][]float32, error) {
	return Embed(ctx, setting, inputs)
}

func (Client) EmbedOne(ctx context.Context, setting settingModel.EmbeddingSetting, input string) ([]float32, error) {
	return EmbedOne(ctx, setting, input)
}
