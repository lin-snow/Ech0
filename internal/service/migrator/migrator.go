// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/migrator/artifact"
	snapshot "github.com/lin-snow/ech0/internal/migrator/snapshot"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// MigratorService 是迁移领域服务的 HTTP 生命周期编排：导入(start/status/cancel/cleanup)、
// 导出(start/status/cancel 走 job.Manager;download 同步取回最新产物)。实际导入/导出由引擎执行体
// (migrator.ImportEngine / migrator.ExportEngine)承担;本层只做 auth + 作业生命周期 + DTO 转发。
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

// DownloadExport 流式下发「上一次导出作业产出的产物」(GET /migration/export/download)。
// 与导入的 upload 对称:重活(打包/S3)在异步 export 作业里完成,这里只同步取回产物,不再现打包。
// format 决定取哪个槽位——快照与胶囊各自「只保留最新一份」,混用会互删,详见 migrator/artifact。
// 无可用产物时报错提示先创建导出。下发后发 SystemExport 事件。
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
		return errors.New("暂无可下载的产物，请先创建导出")
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
		return migratorModel.UploadMigrationSourceZipResponse{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}

	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, err
	}
	if !user.IsAdmin {
		return migratorModel.UploadMigrationSourceZipResponse{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	// 非 idle（存在任何作业行，无论在跑还是终态）则要求先清理，沿用旧语义。
	if _, err := s.jobManager.Get(ctx, jobModel.TypeMigration); err == nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, errors.New("请先结束/清理当前迁移")
	} else if !errors.Is(err, job.ErrNotFound) {
		return migratorModel.UploadMigrationSourceZipResponse{}, err
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		return migratorModel.UploadMigrationSourceZipResponse{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}

	baseTmpDir := filepath.Join("data", coreMigrator.TmpRelativeDir)
	if err := os.MkdirAll(baseTmpDir, 0o755); err != nil {
		return migratorModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("create migration tmp dir: %w", err)
	}

	uploadID := uuidUtil.MustNewV7()
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

	relativeTmpDir := filepath.ToSlash(filepath.Join(coreMigrator.TmpRelativeDir, folderName))
	return migratorModel.UploadMigrationSourceZipResponse{
		SourceType:    sourceType,
		TmpDir:        relativeTmpDir,
		SourcePayload: map[string]any{"tmp_dir": relativeTmpDir},
	}, nil
}

// StartGlobalMigration 提交一次迁移作业（互斥 + 持久化 + goroutine 生命周期由 job.Manager 负责）。
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
	raw, err := json.Marshal(migratorModel.MigrationPayload{
		SourceType:    strings.TrimSpace(req.SourceType),
		SourcePayload: sourcePayload,
	})
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}

	jb, err := s.jobManager.Submit(ctx, jobModel.TypeMigration, raw)
	if err != nil {
		// 提交失败（如同类型互斥）：清理刚上传的 tmp，沿用旧「请先结束/清理当前迁移」。
		_ = coreMigrator.CleanupTmpDirFromPayload(req.SourcePayload)
		if errors.Is(err, job.ErrAlreadyRunning) {
			return migratorModel.GlobalMigrationStateDTO{}, errors.New("请先结束/清理当前迁移")
		}
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

// GetGlobalMigrationStatus 查询当前状态；查无作业行时合成 idle 哨兵。
func (s *MigratorService) GetGlobalMigrationStatus(ctx context.Context) (migratorModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeMigration)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.GlobalMigrationStateDTO{Version: 1, Status: migratorModel.MigrationStatusIdle}, nil
	}
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

// CancelGlobalMigration 协作式取消在跑迁移；返回当前状态（前端轮询收敛到 cancelled）。
func (s *MigratorService) CancelGlobalMigration(ctx context.Context) (migratorModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeMigration)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.GlobalMigrationStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	if jb.Status != jobModel.StatusPending && jb.Status != jobModel.StatusRunning {
		return migratorModel.GlobalMigrationStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	_ = s.jobManager.Cancel(jobModel.TypeMigration)
	jb, err = s.jobManager.Get(ctx, jobModel.TypeMigration)
	if err != nil {
		return migratorModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

// CleanupGlobalMigration 清理 tmp 目录并删除作业行（复位 idle）。
func (s *MigratorService) CleanupGlobalMigration(ctx context.Context) error {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeMigration)
	if errors.Is(err, job.ErrNotFound) {
		return nil // 已是 idle，幂等
	}
	if err != nil {
		return err
	}
	if jb.Status == jobModel.StatusPending || jb.Status == jobModel.StatusRunning {
		return errors.New("迁移进行中，无法清理")
	}
	var payload migratorModel.MigrationPayload
	if jb.Payload != "" {
		_ = json.Unmarshal([]byte(jb.Payload), &payload)
	}
	if err := coreMigrator.CleanupTmpDirFromPayload(payload.SourcePayload); err != nil {
		return fmt.Errorf("cleanup migration tmp dir: %w", err)
	}
	return s.jobManager.Delete(ctx, jobModel.TypeMigration)
}

// StartExport 提交一次导出作业（手动导出的异步出口），格式由请求决定：快照或胶囊。
// 互斥 + 持久化 + goroutine 生命周期由 job.Manager 负责；快照完成由 ExportRunner 发
// SystemSnapshot 事件，无需 service 介入。
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
	raw, err := json.Marshal(migratorModel.ExportPayload{
		Format: format,
		// 快照本就整库带走，包含与否无从谈起；只让胶囊携带这个意图，避免落库的 payload 出现
		// 「快照 + 不含私密」这种读起来自相矛盾的组合。
		IncludePrivate: format == migratorModel.ExportFormatCapsule && req.IncludePrivate,
	})
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	jb, err := s.jobManager.Submit(ctx, jobModel.TypeExport, raw)
	if err != nil {
		if errors.Is(err, job.ErrAlreadyRunning) {
			return migratorModel.ExportStateDTO{}, errors.New("导出进行中，请稍候")
		}
		return migratorModel.ExportStateDTO{}, err
	}
	return s.jobExportToDTO(jb), nil
}

// GetExportStatus 查询当前导出状态；查无作业行时合成 idle 哨兵（与迁移状态机一致）。
func (s *MigratorService) GetExportStatus(ctx context.Context) (migratorModel.ExportStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeExport)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.ExportStateDTO{Version: 1, Status: migratorModel.MigrationStatusIdle}, nil
	}
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	return s.jobExportToDTO(jb), nil
}

// CancelExport 协作式取消在跑导出；返回当前状态（前端轮询收敛到 cancelled）。
func (s *MigratorService) CancelExport(ctx context.Context) (migratorModel.ExportStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeExport)
	if errors.Is(err, job.ErrNotFound) {
		return migratorModel.ExportStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	if jb.Status != jobModel.StatusPending && jb.Status != jobModel.StatusRunning {
		return migratorModel.ExportStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	_ = s.jobManager.Cancel(jobModel.TypeExport)
	jb, err = s.jobManager.Get(ctx, jobModel.TypeExport)
	if err != nil {
		return migratorModel.ExportStateDTO{}, err
	}
	return s.jobExportToDTO(jb), nil
}

// jobExportToDTO 把通用 Job 映射回前端契约的 ExportStateDTO。终态成功时 Payload 为
// migrator.ExportOutcome 的 JSON（含 file_name/size），此处解析补出。
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
		// 作业行的 Payload 列先存输入 ExportPayload、成功后被 ExportOutcome 覆盖。两者都带
		// format，故一次解析就能覆盖「运行中」与「已完成」两种形态。
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
	// 改版前落库的作业行没有 format，缺省即快照——下载出口据此取槽位，不能留空。
	if dto.Format == "" {
		dto.Format = migratorModel.ExportFormatSnapshot
	}
	if jb.UpdatedAt != 0 {
		updatedAt := jb.UpdatedAt
		dto.UpdatedAt = &updatedAt
	}
	return dto
}

// jobToDTO 把通用 Job 映射回前端契约的 GlobalMigrationStateDTO（适配层，不污染框架）。
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
		return errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	return nil
}

func (s *MigratorService) ensureAdmin(ctx context.Context) (string, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return userID, nil
}

func validateSourceType(sourceType string) error {
	switch strings.TrimSpace(sourceType) {
	case migratorModel.MigrationSourceMemos,
		migratorModel.MigrationSourceEch0,
		migratorModel.MigrationSourceCapsule:
		return nil
	default:
		return errors.New(commonModel.INVALID_REQUEST_BODY)
	}
}

// normalizeExportFormat 收口导出格式：空值即快照，既是缺省语义，也让不带 format 的旧客户端
// 保持原行为。未知取值直接拒绝而非静默回落——悄悄给出另一种格式的产物，用户会拿错东西。
func normalizeExportFormat(format string) (string, error) {
	switch strings.TrimSpace(format) {
	case "", migratorModel.ExportFormatSnapshot:
		return migratorModel.ExportFormatSnapshot, nil
	case migratorModel.ExportFormatCapsule:
		return migratorModel.ExportFormatCapsule, nil
	default:
		return "", errors.New(commonModel.INVALID_REQUEST_BODY)
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
	for k, v := range input {
		out[k] = v
	}
	return out
}
