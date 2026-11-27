// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lin-snow/ech0/internal/job"
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
				Type:       string(jobModel.Reindex),
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
	require.NotNil(t, out)
	assert.Equal(t, reindexStatusIdle, out.Body.Status)
	assert.Nil(t, out.Body.Payload)
}

func TestReindexStatus_ExistingJobMapped(t *testing.T) {
	h, repo := newEmbeddingHandlerWithDB(t)
	require.NoError(t, repo.Upsert(context.Background(), &jobModel.Job{
		Type:      string(jobModel.Reindex),
		Status:    jobModel.StatusRunning,
		Phase:     "embedding",
		Payload:   `{"total":10}`,
		StartedAt: new(int64(1234)),
	}))

	out, err := h.ReindexStatus(context.Background(), &ReindexStatusInput{})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, string(jobModel.StatusRunning), out.Body.Status)
	assert.Equal(t, "embedding", out.Body.Phase)
	require.NotNil(t, out.Body.StartedAt)
	assert.Equal(t, int64(1234), *out.Body.StartedAt)
	assert.JSONEq(t, `{"total":10}`, string(out.Body.Payload))
}

func TestCancelReindex_NoJobSynthesizesIdle(t *testing.T) {
	h, _ := newEmbeddingHandlerWithDB(t)

	out, err := h.CancelReindex(context.Background(), &CancelReindexInput{})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, reindexStatusIdle, out.Body.Status)
}

func TestCancelReindex_TerminalRowMapped(t *testing.T) {
	h, repo := newEmbeddingHandlerWithDB(t)
	require.NoError(t, repo.Upsert(context.Background(), &jobModel.Job{
		Type:   string(jobModel.Reindex),
		Status: jobModel.StatusSuccess,
		Phase:  "done",
	}))

	out, err := h.CancelReindex(context.Background(), &CancelReindexInput{})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, string(jobModel.StatusSuccess), out.Body.Status)
	assert.Equal(t, "done", out.Body.Phase)
}

func TestReindex_NoRunnerReturnsErrNoRunner(t *testing.T) {
	h, _ := newEmbeddingHandlerWithDB(t)

	out, err := h.Reindex(context.Background(), &ReindexInput{})

	require.ErrorIs(t, err, job.ErrNoRunner)
	assert.Nil(t, out)
}
