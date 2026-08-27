// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/job"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
)

const reindexStatusIdle = "idle"

type EmbeddingHandler struct {
	jobManager *job.Manager
}

func NewEmbeddingHandler(jobManager *job.Manager) *EmbeddingHandler {
	return &EmbeddingHandler{
		jobManager: jobManager,
	}
}

type ReindexStatusResponse struct {
	Status     string          `json:"status" doc:"作业状态：idle/pending/running/succeeded/failed/cancelled" example:"running"`
	Phase      string          `json:"phase,omitempty" doc:"当前阶段"`
	Error      string          `json:"error,omitempty" doc:"失败原因（status=failed 时）"`
	Payload    json.RawMessage `json:"payload,omitempty" doc:"回填结果 BackfillResult: total/indexed/skipped/failed"`
	StartedAt  *int64          `json:"started_at,omitempty" doc:"开始时间（Unix 秒）"`
	FinishedAt *int64          `json:"finished_at,omitempty" doc:"结束时间（Unix 秒）"`
}

type (
	ReindexInput       struct{}
	ReindexStatusInput struct{}
	CancelReindexInput struct{}
)

type ReindexOutput = commonModel.Result[ReindexStatusResponse]

func mapJobToReindexStatus(jb jobModel.Job) ReindexStatusResponse {
	resp := ReindexStatusResponse{
		Status:     string(jb.Status),
		Phase:      jb.Phase,
		Error:      jb.Error,
		StartedAt:  jb.StartedAt,
		FinishedAt: jb.FinishedAt,
	}
	if jb.Payload != "" {
		resp.Payload = json.RawMessage(jb.Payload)
	}
	return resp
}

func (embeddingHandler *EmbeddingHandler) Reindex(ctx context.Context, _ *ReindexInput) (ReindexOutput, error) {
	jb, err := embeddingHandler.jobManager.Submit(ctx, jobModel.TypeReindex, nil)
	if err != nil {
		return ReindexOutput{}, err
	}
	return commonModel.OK(mapJobToReindexStatus(jb)), nil
}

func (embeddingHandler *EmbeddingHandler) ReindexStatus(ctx context.Context, _ *ReindexStatusInput) (ReindexOutput, error) {
	jb, err := embeddingHandler.jobManager.Get(ctx, jobModel.TypeReindex)
	if errors.Is(err, job.ErrNotFound) {
		return commonModel.OK(ReindexStatusResponse{Status: reindexStatusIdle}), nil
	}
	if err != nil {
		return ReindexOutput{}, err
	}
	return commonModel.OK(mapJobToReindexStatus(jb)), nil
}

func (embeddingHandler *EmbeddingHandler) CancelReindex(ctx context.Context, _ *CancelReindexInput) (ReindexOutput, error) {
	_ = embeddingHandler.jobManager.Cancel(jobModel.TypeReindex)
	jb, err := embeddingHandler.jobManager.Get(ctx, jobModel.TypeReindex)
	if errors.Is(err, job.ErrNotFound) {
		return commonModel.OK(ReindexStatusResponse{Status: reindexStatusIdle}), nil
	}
	if err != nil {
		return ReindexOutput{}, err
	}
	return commonModel.OK(mapJobToReindexStatus(jb)), nil
}
