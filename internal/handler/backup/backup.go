package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	response "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/backup"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

type BackupHandler struct {
	backupService *service.BackupService
}

// NewBackupHandler BackupHandler 的构造函数
func NewBackupHandler(backupService *service.BackupService) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
	}
}

// Backup 执行数据备份
//
//	@Summary		执行数据备份
//	@Description	用户触发数据备份操作，成功后返回备份成功信息
//	@Tags			系统备份
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response	"备份成功"
//	@Failure		200	{object}	response.Response	"备份失败"
//	@Router			/backup [get]
func (backupHandler *BackupHandler) Backup() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		userId := ctx.MustGet("userid").(uint)
		if err := backupHandler.backupService.Backup(userId); err != nil {
			return response.Response{
				Msg: "",
				Err: err,
			}
		}

		return response.Response{
			Msg: commonModel.BACKUP_SUCCESS,
		}
	})
}

// ExportBackup 导出数据备份
//
//	@Summary		导出数据备份
//	@Description	用户导出备份文件，成功后触发文件下载
//	@Tags			系统备份
//	@Accept			json
//	@Produce		application/octet-stream
//	@Success		200	{object}	response.Response	"导出备份成功，返回文件下载"
//	@Failure		200	{object}	response.Response	"导出备份失败"
//	@Router			/backup/export [get]
func (backupHandler *BackupHandler) ExportBackup() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		token := ctx.Query("token")
		if token == "" {
			return response.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
			}
		}

		token = strings.Trim(token, `"`) // 去掉可能的双引号

		// 使用 JWT Util进行处理
		claims, err := jwtUtil.ParseToken(token)
		if err != nil {
			return response.Response{
				Msg: commonModel.TOKEN_NOT_VALID,
				Err: err,
			}
		}

		// 从 Claims中提取 UserID
		userId := uint(claims.Userid)
		if err := backupHandler.backupService.ExportBackup(ctx, userId); err != nil {
			return response.Response{
				Msg: "",
				Err: err,
			}
		}

		return response.Response{
			Msg: commonModel.EXPORT_BACKUP_SUCCESS,
		}
	})
}

// ImportBackup 恢复数据备份
//
//	@Summary		恢复数据备份
//	@Description	用户上传备份文件，成功后恢复数据
//	@Tags			系统备份
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file				true	"备份文件"
//	@Success		200		{object}	response.Response	"导入备份成功"
//	@Failure		200		{object}	response.Response	"导入备份失败"
//	@Router			/backup/import [post]
func (backupHandler *BackupHandler) ImportBackup() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		// 提取userid
		userId := ctx.MustGet("userid").(uint)

		// 提取上传的 File数据
		file, err := ctx.FormFile("file")
		if err != nil {
			return response.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := backupHandler.backupService.ImportBackup(ctx, userId, file); err != nil {
			return response.Response{
				Msg: "",
				Err: err,
			}
		}

		return response.Response{
			Msg: commonModel.IMPORT_BACKUP_SUCCESS,
		}
	})
}
