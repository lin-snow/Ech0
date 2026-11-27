// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/migrator/artifact"
	snapshot "github.com/lin-snow/ech0/internal/migrator/snapshot"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type MigratorService struct {
	commonService CommonService
	jobManager    *job.Manager
	bus           *busen.Bus
}

func NewMigratorService(
	commonService CommonService,
	jobManager *job.Manager,
	busProvider func() *busen.Bus,
) *MigratorService {
	return &MigratorService{
		commonService: commonService,
		jobManager:    jobManager,
		bus:           busProvider(),
	}
}

func (s *MigratorService) DownloadExport(ctx *gin.Context, reqCtx context.Context, format string) error {
	if _, err := s.ensureAdmin(reqCtx); err != nil {
		return err
	}

	format, err := normalizeExportFormat(format)
	if err != nil {
		return err
	}

	slot := artifact.Snapshots()
	if format == migratorModel.ExportFormatCapsule {
		slot = artifact.Capsules()
	}

	artifactPath, err := slot.Latest()
	if errors.Is(err, artifact.ErrNone) {
		return apperr.ErrNoExport
	}
	if err != nil {
		return err
	}

	info, err := os.Stat(artifactPath)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("ech0-%s-%s.zip", format, time.Now().UTC().Format("2006-01-02-150405"))

	ctx.Writer.Header().Set("Content-Type", "application/zip")
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	ctx.Writer.Header().Set("Accept-Ranges", "bytes")
	ctx.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Writer.WriteHeader(200)
	ctx.File(artifactPath)

	eventbus.Notify(context.Background(), s.bus, event.SystemExport{Info: "System export completed", Size: info.Size()})

	return nil
}

func (s *MigratorService) UploadSourceZip(
	ctx context.Context,
	sourceType string,
	file *multipart.FileHeader,
) (migratorModel.UploadMigrationSourceZipResponse, error) {
	if err := validateSourceType(sourceType); err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, err
	}
	if file == nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, apperr.ErrInvalidRequestBody
	}

	if err := authz.Decide(viewer.MustFromContext(ctx), authz.DataManage, authz.Instance()); err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, err
	}
	if _, err := s.jobManager.Get(ctx, jobModel.Migration); err == nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, apperr.ErrMigrationBusy
	} else if !errors.Is(err, job.ErrNotFound) {
		return migratorModel.UploadMigrationSourceZipResponse{}, err
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		return migratorModel.UploadMigrationSourceZipResponse{}, apperr.ErrInvalidRequestBody
	}

	baseTmpDir := artifact.TmpRoot()
	if err := os.MkdirAll(baseTmpDir, 0o755); err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("create migration tmp dir: %w", err)
	}

	uploadID := uuidUtil.NewV7()
	folderName := fmt.Sprintf("%s_%s", strings.TrimSpace(sourceType), uploadID)
	zipPath := filepath.Join(baseTmpDir, folderName+".zip")
	extractDir := filepath.Join(baseTmpDir, folderName)

	if err := saveMultipartFile(file, zipPath); err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("save uploaded zip: %w", err)
	}
	defer func() {
		_ = os.Remove(zipPath)
	}()

	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("create extract dir: %w", err)
	}
	if err := snapshot.Unpack(zipPath, extractDir); err != nil {
		_ = os.RemoveAll(extractDir)
		return migratorModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("unpack migration zip: %w", err)
	}

	relativeTmpDir := filepath.ToSlash(filepath.Join(artifact.TmpDir, folderName))
	return migratorModel.UploadMigrationSourceZipResponse{
		SourceType:    sourceType,
		TmpDir:        relativeTmpDir,
		SourcePayload: map[string]any{"tmp_dir": relativeTmpDir},
	}, nil
}

func (s *MigratorService) StartGlobalMigration(
	ctx context.Context,
	req migratorModel.StartGlobalMigrationRequest,
) (migratorModel.GlobalMigrationStateDTO, error) {
	if err := validateStartRequest(req); err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	adminUserID, err := s.ensureAdmin(ctx)
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}

	sourcePayload := cloneMap(req.SourcePayload)
	if _, ok := sourcePayload["created_by"]; !ok {
		sourcePayload["created_by"] = adminUserID
	}
	jb, err := s.jobManager.Submit(ctx, jobModel.Migration, migratorModel.MigrationPayload{
		SourceType:    strings.TrimSpace(req.SourceType),
		SourcePayload: sourcePayload,
	})
	if err != nil {
		_ = coreMigrator.CleanupTmpDirFromPayload(req.SourcePayload)
		if errors.Is(err, job.ErrAlreadyRunning) {
			return migratorModel.GlobalMigrationStateDTO{}, apperr.ErrMigrationBusy
		}
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

func (s *MigratorService) GetGlobalMigrationStatus(ctx context.Context) (migratorModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.Migration)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.GlobalMigrationStateDTO{Version: 1, Status: migratorModel.MigrationStatusIdle}, nil
	}
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

func (s *MigratorService) CancelGlobalMigration(ctx context.Context) (migratorModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.Migration)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.GlobalMigrationStateDTO{}, apperr.ErrInvalidRequestBody
	}
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	if jb.Status != jobModel.StatusPending && jb.Status != jobModel.StatusRunning {
		return migratorModel.GlobalMigrationStateDTO{}, apperr.ErrInvalidRequestBody
	}
	_ = s.jobManager.Cancel(jobModel.Migration)
	jb, err = s.jobManager.Get(ctx, jobModel.Migration)
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

func (s *MigratorService) CleanupGlobalMigration(ctx context.Context) error {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.Migration)
	if errors.Is(err, job.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if jb.Status == jobModel.StatusPending || jb.Status == jobModel.StatusRunning {
		return apperr.ErrMigrationRunning
	}
	var payload migratorModel.MigrationPayload
	if jb.Payload != "" {
		_ = json.Unmarshal([]byte(jb.Payload), &payload)
	}
	if err := coreMigrator.CleanupTmpDirFromPayload(payload.SourcePayload); err != nil {
		return fmt.Errorf("cleanup migration tmp dir: %w", err)
	}
	return s.jobManager.Delete(ctx, jobModel.Migration)
}

func (s *MigratorService) StartExport(
	ctx context.Context,
	req migratorModel.StartExportRequest,
) (migratorModel.ExportStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	format, err := normalizeExportFormat(req.Format)
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	jb, err := s.jobManager.Submit(ctx, jobModel.Export, migratorModel.ExportPayload{
		Format:         format,
		IncludePrivate: format == migratorModel.ExportFormatCapsule && req.IncludePrivate,
	})
	if err != nil {
		if errors.Is(err, job.ErrAlreadyRunning) {
			return migratorModel.ExportStateDTO{}, apperr.ErrExportRunning
		}
		return migratorModel.ExportStateDTO{}, err
	}
	return s.jobExportToDTO(jb), nil
}

func (s *MigratorService) GetExportStatus(ctx context.Context) (migratorModel.ExportStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.Export)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.ExportStateDTO{Version: 1, Status: migratorModel.MigrationStatusIdle}, nil
	}
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	return s.jobExportToDTO(jb), nil
}

func (s *MigratorService) CancelExport(ctx context.Context) (migratorModel.ExportStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.Export)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.ExportStateDTO{}, apperr.ErrInvalidRequestBody
	}
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	if jb.Status != jobModel.StatusPending && jb.Status != jobModel.StatusRunning {
		return migratorModel.ExportStateDTO{}, apperr.ErrInvalidRequestBody
	}
	_ = s.jobManager.Cancel(jobModel.Export)
	jb, err = s.jobManager.Get(ctx, jobModel.Export)
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	return s.jobExportToDTO(jb), nil
}

func (s *MigratorService) jobExportToDTO(jb jobModel.Job) migratorModel.ExportStateDTO {
	dto := migratorModel.ExportStateDTO{
		Version:      1,
		Status:       string(jb.Status),
		Phase:        jb.Phase,
		ErrorMessage: jb.Error,
		StartedAt:    jb.StartedAt,
		FinishedAt:   jb.FinishedAt,
	}
	if jb.Payload != "" {
		var outcome struct {
			FileName string `json:"file_name"`
			Size     int64  `json:"size"`
			Format   string `json:"format"`
		}
		if err := json.Unmarshal([]byte(jb.Payload), &outcome); err == nil {
			dto.FileName = outcome.FileName
			dto.Size = outcome.Size
			dto.Format = outcome.Format
		}
	}
	if dto.Format == "" {
		dto.Format = migratorModel.ExportFormatSnapshot
	}
	if jb.UpdatedAt != 0 {
		updatedAt := jb.UpdatedAt
		dto.UpdatedAt = &updatedAt
	}
	return dto
}

func (s *MigratorService) jobToDTO(jb jobModel.Job) migratorModel.GlobalMigrationStateDTO {
	var payload migratorModel.MigrationPayload
	if jb.Payload != "" {
		_ = json.Unmarshal([]byte(jb.Payload), &payload)
	}
	dto := migratorModel.GlobalMigrationStateDTO{
		Version:       1,
		SourceType:    payload.SourceType,
		Status:        string(jb.Status),
		Phase:         jb.Phase,
		ErrorMessage:  jb.Error,
		SourcePayload: payload.SourcePayload,
		StartedAt:     jb.StartedAt,
		FinishedAt:    jb.FinishedAt,
	}
	if jb.UpdatedAt != 0 {
		updatedAt := jb.UpdatedAt
		dto.UpdatedAt = &updatedAt
	}
	return dto
}

func validateStartRequest(req migratorModel.StartGlobalMigrationRequest) error {
	if err := validateSourceType(req.SourceType); err != nil {
		return err
	}
	tmpDir, ok := req.SourcePayload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDir) == "" {
		return apperr.ErrInvalidRequestBody
	}
	return nil
}

func (s *MigratorService) ensureAdmin(ctx context.Context) (string, error) {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.DataManage, authz.Instance()); err != nil {
		return "", err
	}
	return viewer.MustFromContext(ctx).UserID(), nil
}

func validateSourceType(sourceType string) error {
	switch strings.TrimSpace(sourceType) {
	case migratorModel.MigrationSourceMemos,
		migratorModel.MigrationSourceEch0,
		migratorModel.MigrationSourceCapsule:
		return nil
	default:
		return apperr.ErrInvalidRequestBody
	}
}

func normalizeExportFormat(format string) (string, error) {
	switch strings.TrimSpace(format) {
	case "", migratorModel.ExportFormatSnapshot:
		return migratorModel.ExportFormatSnapshot, nil
	case migratorModel.ExportFormatCapsule:
		return migratorModel.ExportFormatCapsule, nil
	default:
		return "", apperr.ErrInvalidRequestBody
	}
}

func saveMultipartFile(file *multipart.FileHeader, dstPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return err
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	maps.Copy(out, input)
	return out
}
