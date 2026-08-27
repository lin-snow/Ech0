// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lin-snow/ech0/internal/job"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	jobRepository "github.com/lin-snow/ech0/internal/repository/job"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMapJobToReindexStatus(t *testing.T) {
	cases := []struct {
		name string
		in   jobModel.Job
		want ReindexStatusResponse
	}{
		{
			name: "all-fields-passthrough-with-payload",
			in: jobModel.Job{
				Type:       jobModel.TypeReindex,
				Status:     jobModel.StatusRunning,
				Phase:      "embedding",
				Error:      "",
				Payload:    `{"total":10,"indexed":4}`,
				StartedAt:  new(int64(1000)),
				FinishedAt: new(int64(2000)),
			},
			want: ReindexStatusResponse{
				Status:     "running",
				Phase:      "embedding",
				Error:      "",
				Payload:    json.RawMessage(`{"total":10,"indexed":4}`),
				StartedAt:  new(int64(1000)),
				FinishedAt: new(int64(2000)),
			},
		},
		{
			name: "empty-payload-leaves-payload-nil",
			in: jobModel.Job{
				Status:  jobModel.StatusPending,
				Payload: "",
			},
			want: ReindexStatusResponse{
				Status:  "pending",
				Payload: nil,
			},
		},
		{
			name: "failed-carries-error-no-timestamps",
			in: jobModel.Job{
				Status: jobModel.StatusFailed,
				Error:  "boom",
			},
			want: ReindexStatusResponse{
				Status: "failed",
				Error:  "boom",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapJobToReindexStatus(tc.in)
			assert.Equal(t, tc.want, got)
			if tc.in.Payload == "" {
				assert.Nil(t, got.Payload)
			} else {
				assert.JSONEq(t, tc.in.Payload, string(got.Payload))
			}
		})
	}
}

func newEmbeddingHandlerWithDB(t *testing.T) (*EmbeddingHandler, *jobRepository.JobRepository) {
	t.Helper()
	db := helpers.NewTestDB(t)
	repo := jobRepository.NewJobRepository(func() *gorm.DB { return db })
	mgr := job.NewManager(repo)
	return NewEmbeddingHandler(mgr), repo
}

func TestReindexStatus_NoJobSynthesizesIdle(t *testing.T) {
	h, _ := newEmbeddingHandlerWithDB(t)

	out, err := h.ReindexStatus(context.Background(), &ReindexStatusInput{})

	require.NoError(t, err)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	assert.Equal(t, reindexStatusIdle, out.Data.Status)
	assert.Nil(t, out.Data.Payload)
}

func TestReindexStatus_ExistingJobMapped(t *testing.T) {
	h, repo := newEmbeddingHandlerWithDB(t)
	require.NoError(t, repo.Upsert(context.Background(), &jobModel.Job{
		Type:      jobModel.TypeReindex,
		Status:    jobModel.StatusRunning,
		Phase:     "embedding",
		Payload:   `{"total":10}`,
		StartedAt: new(int64(1234)),
	}))

	out, err := h.ReindexStatus(context.Background(), &ReindexStatusInput{})

	require.NoError(t, err)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	assert.Equal(t, string(jobModel.StatusRunning), out.Data.Status)
	assert.Equal(t, "embedding", out.Data.Phase)
	require.NotNil(t, out.Data.StartedAt)
	assert.Equal(t, int64(1234), *out.Data.StartedAt)
	assert.JSONEq(t, `{"total":10}`, string(out.Data.Payload))
}

func TestCancelReindex_NoJobSynthesizesIdle(t *testing.T) {
	h, _ := newEmbeddingHandlerWithDB(t)

	out, err := h.CancelReindex(context.Background(), &CancelReindexInput{})

	require.NoError(t, err)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	assert.Equal(t, reindexStatusIdle, out.Data.Status)
}

func TestCancelReindex_TerminalRowMapped(t *testing.T) {
	h, repo := newEmbeddingHandlerWithDB(t)
	require.NoError(t, repo.Upsert(context.Background(), &jobModel.Job{
		Type:   jobModel.TypeReindex,
		Status: jobModel.StatusSuccess,
		Phase:  "done",
	}))

	out, err := h.CancelReindex(context.Background(), &CancelReindexInput{})

	require.NoError(t, err)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	assert.Equal(t, string(jobModel.StatusSuccess), out.Data.Status)
	assert.Equal(t, "done", out.Data.Phase)
}

func TestReindex_NoRunnerReturnsErrNoRunner(t *testing.T) {
	h, _ := newEmbeddingHandlerWithDB(t)

	out, err := h.Reindex(context.Background(), &ReindexInput{})

	require.ErrorIs(t, err, job.ErrNoRunner)
	assert.Equal(t, ReindexOutput{}, out)
}
