package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	service "github.com/lin-snow/ech0/internal/service/setting"
)

type SettingHandler struct {
	settingService *service.SettingService
}

// NewSettingHandler SettingHandler 的构造函数
func NewSettingHandler(settingService *service.SettingService) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
	}
}

// GetSettings 获取设置
//
//	@Summary		获取设置
//	@Description	获取系统的全局设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.SystemSetting}	"获取设置成功"
//	@Failure		200	{object}	res.Response							"获取设置失败"
//	@Router			/settings [get]
func (settingHandler *SettingHandler) GetSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var settings model.SystemSetting
		if err := settingHandler.settingService.GetSetting(&settings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: settings,
			Msg:  commonModel.GET_SETTINGS_SUCCESS,
		}
	})
}

// UpdateSettings 更新设置
//
//	@Summary		更新设置
//	@Description	更新系统的全局设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			settings	body		model.SystemSettingDto	true	"新的系统设置"
//	@Success		200			{object}	res.Response			"更新设置成功"
//	@Failure		200			{object}	res.Response			"更新设置失败"
//	@Router			/settings [put]
func (settingHandler *SettingHandler) UpdateSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newSettings model.SystemSettingDto
		if err := ctx.ShouldBindJSON(&newSettings); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateSetting(userid, &newSettings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_SETTINGS_SUCCESS,
		}
	})
}

// GetCommentSettings 获取评论设置
//
//	@Summary		获取评论设置
//	@Description	获取系统的评论相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.CommentSetting}	"获取评论设置成功"
//	@Failure		200	{object}	res.Response							"获取评论设置失败"
//	@Router			/comment/settings [get]
func (settingHandler *SettingHandler) GetCommentSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var commentSetting model.CommentSetting
		if err := settingHandler.settingService.GetCommentSetting(&commentSetting); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: commentSetting,
			Msg:  commonModel.GET_COMMENT_SETTINGS_SUCCESS,
		}
	})
}

// UpdateCommentSettings 更新评论设置
//
//	@Summary		更新评论设置
//	@Description	更新系统的评论相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			commentSettings	body		model.CommentSettingDto	true	"新的评论设置"
//	@Success		200				{object}	res.Response			"更新评论设置成功"
//	@Failure		200				{object}	res.Response			"更新评论设置失败"
//	@Router			/comment/settings [put]
func (settingHandler *SettingHandler) UpdateCommentSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newCommentSettings model.CommentSettingDto
		if err := ctx.ShouldBindJSON(&newCommentSettings); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateCommentSetting(userid, &newCommentSettings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_COMMENT_SETTINGS_SUCCESS,
		}
	})
}

// GetS3Settings 获取 S3 存储设置
//
//	@Summary		获取 S3 存储设置
//	@Description	获取系统的 S3 存储相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.S3Setting}	"获取 S3 存储设置成功"
//	@Failure		200	{object}	res.Response						"获取 S3 存储设置失败"
//	@Router			/s3/settings [get]
func (settingHandler *SettingHandler) GetS3Settings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		var s3Setting model.S3Setting
		if err := settingHandler.settingService.GetS3Setting(userid, &s3Setting); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: s3Setting,
			Msg:  commonModel.GET_S3_SETTINGS_SUCCESS,
		}
	})
}

// UpdateS3Settings 更新 S3 存储设置
//
//	@Summary		更新 S3 存储设置
//	@Description	更新系统的 S3 存储相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			s3Settings	body		model.S3SettingDto	true	"新的 S3 存储设置"
//	@Success		200			{object}	res.Response		"更新 S3 存储设置成功"
//	@Failure		200			{object}	res.Response		"更新 S3 存储设置失败"
//	@Router			/s3/settings [put]
func (settingHandler *SettingHandler) UpdateS3Settings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newS3Settings model.S3SettingDto
		if err := ctx.ShouldBindJSON(&newS3Settings); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateS3Setting(userid, &newS3Settings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_S3_SETTINGS_SUCCESS,
		}
	})
}

// GetOAuth2Settings 获取 OAuth2 设置
//
//	@Summary		获取 OAuth2 设置
//	@Description	获取系统的 OAuth2 相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.OAuth2Setting}	"获取 OAuth2 设置成功"
//	@Failure		200	{object}	res.Response							"获取 OAuth2 设置失败"
//	@Router			/oauth2/settings [get]
func (settingHandler *SettingHandler) GetOAuth2Settings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		var oauthSetting model.OAuth2Setting
		if err := settingHandler.settingService.GetOAuth2Setting(userid, &oauthSetting, false); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: oauthSetting,
			Msg:  commonModel.GET_OAUTH_SETTINGS_SUCCESS,
		}
	})
}

// UpdateOAuth2Settings 更新 OAuth2 设置
//
//	@Summary		更新 OAuth2 设置
//	@Description	更新系统的 OAuth2 相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			oauthSettings	body		model.OAuth2SettingDto	true	"新的 OAuth 设置"
//	@Success		200				{object}	res.Response			"更新 OAuth 设置成功"
//	@Failure		200				{object}	res.Response			"更新 OAuth 设置失败"
//	@Router			/oauth2/settings [put]
func (settingHandler *SettingHandler) UpdateOAuth2Settings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newOAuthSettings model.OAuth2SettingDto
		if err := ctx.ShouldBindJSON(&newOAuthSettings); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateOAuth2Setting(userid, &newOAuthSettings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_OAUTH_SETTINGS_SUCCESS,
		}
	})
}

// GetOAuth2Status 获取 OAuth2 状态
//
//	@Summary		获取 OAuth2 状态
//	@Description	获取系统的 OAuth2 启用状态
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=bool}	"获取 OAuth2 状态成功"
//	@Failure		200	{object}	res.Response			"获取 OAuth2 状态失败"
//	@Router			/oauth2/status [get]
func (settingHandler *SettingHandler) GetOAuth2Status() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var status model.OAuth2Status
		if err := settingHandler.settingService.GetOAuth2Status(&status); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: status,
			Msg:  commonModel.GET_OAUTH2_STATUS_SUCCESS,
		}
	})
}

// GetWebhook 获取所有 Webhook
//
//	@Summary		获取所有 Webhook
//	@Description	获取系统中配置的所有 Webhook 列表
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=[]model.Webhook}	"获取 Webhook 列表成功"
//	@Failure		200	{object}	res.Response						"获取 Webhook 列表失败"
//	@Router			/webhook [get]
func (settingHandler *SettingHandler) GetWebhook() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		result, err := settingHandler.settingService.GetAllWebhooks(userid)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.GET_WEBHOOK_SUCCESS,
		}
	})
}

// DeleteWebhook 删除 Webhook
//
//	@Summary		删除 Webhook
//	@Description	根据 ID 删除指定的 Webhook 配置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"要删除的 Webhook ID"
//	@Success		200	{object}	res.Response	"删除 Webhook 成功"
//	@Failure		200	{object}	res.Response	"删除 Webhook 失败"
//	@Router			/webhook/{id} [delete]
func (settingHandler *SettingHandler) DeleteWebhook() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 从路径参数中获取 Webhook ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		if err := settingHandler.settingService.DeleteWebhook(userid, uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_WEBHOOK_SUCCESS,
		}
	})
}

// UpdateWebhook 更新 Webhook
//
//	@Summary		更新 Webhook
//	@Description	根据 ID 更新指定的 Webhook 配置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"要更新的 Webhook ID"
//	@Param			webhook	body		model.WebhookDto	true	"新的 Webhook 配置"
//	@Success		200		{object}	res.Response		"更新 Webhook 成功"
//	@Failure		200		{object}	res.Response		"更新 Webhook 失败"
//	@Router			/webhook/{id} [put]
func (settingHandler *SettingHandler) UpdateWebhook() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 从路径参数中获取 Webhook ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		// 解析请求体中的参数
		var updatedWebhook model.WebhookDto
		if err := ctx.ShouldBindJSON(&updatedWebhook); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateWebhook(userid, uint(id), &updatedWebhook); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_WEBHOOK_SUCCESS,
		}
	})
}

// CreateWebhook 创建新的 Webhook
//
//	@Summary		创建新的 Webhook
//	@Description	创建一个新的 Webhook 配置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			webhook	body		model.WebhookDto	true	"新的 Webhook 配置"
//	@Success		200		{object}	res.Response		"创建 Webhook 成功"
//	@Failure		200		{object}	res.Response		"创建 Webhook 失败"
//	@Router			/webhook [post]
func (settingHandler *SettingHandler) CreateWebhook() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newWebhook model.WebhookDto
		if err := ctx.ShouldBindJSON(&newWebhook); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.CreateWebhook(userid, &newWebhook); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.CREATE_WEBHOOK_SUCCESS,
		}
	})
}

// ListAccessTokens 列出访问令牌
//
//	@Summary		列出访问令牌
//	@Description	列出当前用户的所有访问令牌
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=[]model.AccessTokenSetting}	"列出访问令牌成功"
//	@Failure		200	{object}	res.Response									"列出访问令牌失败"
//	@Router			/access-tokens [get]
func (settingHandler *SettingHandler) ListAccessTokens() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		result, err := settingHandler.settingService.ListAccessTokens(userid)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.LIST_ACCESS_TOKENS_SUCCESS,
		}
	})
}

// CreateAccessToken 创建访问令牌
//
//	@Summary		创建访问令牌
//	@Description	为当前用户创建一个新的访问令牌
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			accessToken	body		model.AccessTokenSettingDto					true	"新的访问令牌信息"
//	@Success		200			{object}	res.Response{data=model.AccessTokenSetting}	"创建访问令牌成功"
//	@Failure		200			{object}	res.Response								"创建访问令牌失败"
//	@Router			/access-tokens [post]
func (settingHandler *SettingHandler) CreateAccessToken() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newAccessToken model.AccessTokenSettingDto
		if err := ctx.ShouldBindJSON(&newAccessToken); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		createdToken, err := settingHandler.settingService.CreateAccessToken(
			userid,
			&newAccessToken,
		)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: createdToken,
			Msg:  commonModel.CREATE_ACCESS_TOKEN_SUCCESS,
		}
	})
}

// DeleteAccessToken 删除访问令牌
//
//	@Summary		删除访问令牌
//	@Description	根据 ID 删除指定的访问令牌
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"要删除的访问令牌 ID"
//	@Success		200	{object}	res.Response	"删除访问令牌成功"
//	@Failure		200	{object}	res.Response	"删除访问令牌失败"
//	@Router			/access-tokens/{id} [delete]
func (settingHandler *SettingHandler) DeleteAccessToken() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 从路径参数中获取 访问令牌 ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		if err := settingHandler.settingService.DeleteAccessToken(userid, uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_ACCESS_TOKEN_SUCCESS,
		}
	})
}

// GetBackupScheduleSetting 获取备份计划
//
//	@Summary		获取备份计划
//	@Description	获取系统的定期备份计划设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.BackupSchedule}	"获取备份计划成功"
//	@Failure		200	{object}	res.Response							"获取备份计划失败"
//	@Router			/backup/schedule [get]
func (settingHandler *SettingHandler) GetBackupScheduleSetting() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		// userid := ctx.MustGet("userid").(uint)

		var backupSchedule model.BackupSchedule
		if err := settingHandler.settingService.GetBackupScheduleSetting(&backupSchedule); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: backupSchedule,
			Msg:  commonModel.GET_SETTINGS_SUCCESS,
		}
	})
}

// UpdateScheduleBackupSettings 更新备份计划
//
//	@Summary		更新备份计划
//	@Description	为系统设置定期备份计划
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			backupSchedule	body		model.BackupScheduleDto	true	"备份计划设置"
//	@Success		200				{object}	res.Response			"设置备份计划成功"
//	@Failure		200				{object}	res.Response			"设置备份计划失败"
//	@Router			/backup/schedule [post]
func (settingHandler *SettingHandler) UpdateBackupScheduleSetting() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)
		// 解析请求体中的参数
		var backupSchedule model.BackupScheduleDto
		if err := ctx.ShouldBindJSON(&backupSchedule); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateBackupScheduleSetting(userid, &backupSchedule); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.SCHEDULE_BACKUP_SUCCESS,
		}
	})
}

// GetAgentInfo 获取 Agent 信息
//
//	@Summary		获取 Agent 信息
//	@Description	获取系统的 Agent 信息
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"获取 Agent 信息成功"
//	@Failure		200	{object}	res.Response	"获取 Agent 信息失败"
//	@Router			/agent/info [get]
func (settingHandler *SettingHandler) GetAgentInfo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var settings model.AgentSetting
		if err := settingHandler.settingService.GetAgentInfo(&settings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		settings.ApiKey = ""  // 不返回 API Key 信息
		settings.Prompt = ""  // 不返回 Prompt 信息
		settings.BaseURL = "" // 不返回 BaseURL 信息

		return res.Response{
			Data: settings,
			Msg:  commonModel.GET_SETTINGS_SUCCESS,
		}
	})
}

// GetAgentSettings 获取 Agent 设置
//
//	@Summary		获取 Agent 设置
//	@Description	获取系统的 Agent 相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.AgentSetting}	"获取 Agent 设置成功"
//	@Failure		200	{object}	res.Response							"获取 Agent 设置失败"
//	@Router			/agent/settings [get]
func (settingHandler *SettingHandler) GetAgentSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		var settings model.AgentSetting
		if err := settingHandler.settingService.GetAgentSettings(userid, &settings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: settings,
			Msg:  commonModel.GET_SETTINGS_SUCCESS,
		}
	})
}

// UpdateAgentSettings 更新 Agent 设置
//
//	@Summary		更新 Agent 设置
//	@Description	更新系统的 Agent 相关设置
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			settings	body		model.AgentSettingDto	true	"新的 Agent 设置"
//	@Success		200			{object}	res.Response			"更新 Agent 设置成功"
//	@Failure		200			{object}	res.Response			"更新 Agent 设置失败"
//	@Router			/agent/settings [put]
func (settingHandler *SettingHandler) UpdateAgentSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 解析请求体中的参数
		var newSettings model.AgentSettingDto
		if err := ctx.ShouldBindJSON(&newSettings); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := settingHandler.settingService.UpdateAgentSettings(userid, &newSettings); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_SETTINGS_SUCCESS,
		}
	})
}
