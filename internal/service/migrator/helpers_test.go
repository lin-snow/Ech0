// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"encoding/json"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSourceType(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"ech0", migratorModel.MigrationSourceEch0, false},
		{"memos", migratorModel.MigrationSourceMemos, false},
		{"ech0 with surrounding whitespace", "  ech0  ", false},
		{"memos with tabs", "\tmemos\t", false},
		{"unknown source", "notion", true},
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"case mismatch", "Ech0", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSourceType(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateStartRequest(t *testing.T) {
	cases := []struct {
		name    string
		req     migratorModel.StartGlobalMigrationRequest
		wantErr bool
	}{
		{
			name: "valid ech0 with tmp_dir",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    migratorModel.MigrationSourceEch0,
				SourcePayload: map[string]any{"tmp_dir": "migration_tmp/ech0_x"},
			},
			wantErr: false,
		},
		{
			name: "invalid source type fails before tmp_dir check",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    "bogus",
				SourcePayload: map[string]any{"tmp_dir": "migration_tmp/x"},
			},
			wantErr: true,
		},
		{
			name: "missing tmp_dir key",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    migratorModel.MigrationSourceMemos,
				SourcePayload: map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "nil source payload",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    migratorModel.MigrationSourceMemos,
				SourcePayload: nil,
			},
			wantErr: true,
		},
		{
			name: "tmp_dir empty string",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    migratorModel.MigrationSourceEch0,
				SourcePayload: map[string]any{"tmp_dir": ""},
			},
			wantErr: true,
		},
		{
			name: "tmp_dir whitespace only",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    migratorModel.MigrationSourceEch0,
				SourcePayload: map[string]any{"tmp_dir": "   "},
			},
			wantErr: true,
		},
		{
			name: "tmp_dir wrong type",
			req: migratorModel.StartGlobalMigrationRequest{
				SourceType:    migratorModel.MigrationSourceEch0,
				SourcePayload: map[string]any{"tmp_dir": 123},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStartRequest(tc.req)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCloneMap(t *testing.T) {
	t.Run("nil input yields non-nil empty map", func(t *testing.T) {
		got := cloneMap(nil)
		require.NotNil(t, got)
		assert.Empty(t, got)
	})
	t.Run("empty input yields non-nil empty map", func(t *testing.T) {
		got := cloneMap(map[string]any{})
		require.NotNil(t, got)
		assert.Empty(t, got)
	})
	t.Run("copies all entries", func(t *testing.T) {
		in := map[string]any{"a": 1, "b": "two", "c": true}
		got := cloneMap(in)
		assert.Equal(t, in, got)
	})
	t.Run("returns an independent copy", func(t *testing.T) {
		in := map[string]any{"a": 1}
		got := cloneMap(in)
		got["b"] = 2
		_, mutatedSource := in["b"]
		assert.False(t, mutatedSource, "mutating clone must not affect source")
		assert.Len(t, in, 1)
	})
}

func TestJobToDTO(t *testing.T) {
	s := &MigratorService{}

	t.Run("maps fields and parses payload", func(t *testing.T) {
		started := int64(100)
		finished := int64(200)
		payload, err := json.Marshal(migratorModel.MigrationPayload{
			SourceType:    migratorModel.MigrationSourceEch0,
			SourcePayload: map[string]any{"tmp_dir": "migration_tmp/ech0_x"},
		})
		require.NoError(t, err)

		jb := jobModel.Job{
			Type:       jobModel.TypeMigration,
			Status:     jobModel.StatusRunning,
			Phase:      migratorModel.MigrationPhaseLoading,
			Error:      "boom",
			Payload:    string(payload),
			StartedAt:  &started,
			FinishedAt: &finished,
			UpdatedAt:  150,
		}

		dto := s.jobToDTO(jb)
		assert.Equal(t, 1, dto.Version)
		assert.Equal(t, migratorModel.MigrationSourceEch0, dto.SourceType)
		assert.Equal(t, string(jobModel.StatusRunning), dto.Status)
		assert.Equal(t, migratorModel.MigrationPhaseLoading, dto.Phase)
		assert.Equal(t, "boom", dto.ErrorMessage)
		assert.Equal(t, map[string]any{"tmp_dir": "migration_tmp/ech0_x"}, dto.SourcePayload)
		assert.Equal(t, &started, dto.StartedAt)
		assert.Equal(t, &finished, dto.FinishedAt)
		require.NotNil(t, dto.UpdatedAt)
		assert.Equal(t, int64(150), *dto.UpdatedAt)
	})

	t.Run("empty payload leaves source fields zero", func(t *testing.T) {
		dto := s.jobToDTO(jobModel.Job{Status: jobModel.StatusPending})
		assert.Equal(t, 1, dto.Version)
		assert.Empty(t, dto.SourceType)
		assert.Nil(t, dto.SourcePayload)
		assert.Equal(t, string(jobModel.StatusPending), dto.Status)
	})

	t.Run("malformed payload is ignored", func(t *testing.T) {
		dto := s.jobToDTO(jobModel.Job{Status: jobModel.StatusFailed, Payload: "{not-json"})
		assert.Empty(t, dto.SourceType)
		assert.Nil(t, dto.SourcePayload)
	})

	t.Run("zero UpdatedAt yields nil pointer", func(t *testing.T) {
		dto := s.jobToDTO(jobModel.Job{Status: jobModel.StatusSuccess, UpdatedAt: 0})
		assert.Nil(t, dto.UpdatedAt)
	})
}

func TestJobExportToDTO(t *testing.T) {
	s := &MigratorService{}

	t.Run("parses file_name and size from payload", func(t *testing.T) {
		started := int64(10)
		finished := int64(20)
		jb := jobModel.Job{
			Type:       jobModel.TypeExport,
			Status:     jobModel.StatusSuccess,
			Phase:      migratorModel.ExportPhaseCompleted,
			Error:      "",
			Payload:    `{"file_name":"snap.zip","size":4096}`,
			StartedAt:  &started,
			FinishedAt: &finished,
			UpdatedAt:  15,
		}
		dto := s.jobExportToDTO(jb)
		assert.Equal(t, 1, dto.Version)
		assert.Equal(t, string(jobModel.StatusSuccess), dto.Status)
		assert.Equal(t, migratorModel.ExportPhaseCompleted, dto.Phase)
		assert.Equal(t, "snap.zip", dto.FileName)
		assert.Equal(t, int64(4096), dto.Size)
		assert.Equal(t, &started, dto.StartedAt)
		assert.Equal(t, &finished, dto.FinishedAt)
		require.NotNil(t, dto.UpdatedAt)
		assert.Equal(t, int64(15), *dto.UpdatedAt)
	})

	t.Run("empty payload leaves file fields zero", func(t *testing.T) {
		dto := s.jobExportToDTO(jobModel.Job{Status: jobModel.StatusRunning})
		assert.Empty(t, dto.FileName)
		assert.Zero(t, dto.Size)
		assert.Equal(t, string(jobModel.StatusRunning), dto.Status)
	})

	t.Run("malformed payload is ignored", func(t *testing.T) {
		dto := s.jobExportToDTO(jobModel.Job{Status: jobModel.StatusFailed, Payload: "{bad"})
		assert.Empty(t, dto.FileName)
		assert.Zero(t, dto.Size)
	})

	t.Run("zero UpdatedAt yields nil pointer", func(t *testing.T) {
		dto := s.jobExportToDTO(jobModel.Job{Status: jobModel.StatusFailed, UpdatedAt: 0})
		assert.Nil(t, dto.UpdatedAt)
	})
}
