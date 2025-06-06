package handler

import (
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	service "github.com/lin-snow/ech0/internal/service/setting"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
)

type SettingHandler struct {
	settingService service.SettingServiceInterface
}

func NewSettingHandler(settingService service.SettingServiceInterface) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
	}
}

func (settingHandler *SettingHandler) GetSettings(ctx *gin.Context) {
	var settings model.SystemSetting
	if err := settingHandler.settingService.GetSetting(&settings); err != nil {
		ctx.JSON(http.StatusOK, errUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		}))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[model.SystemSetting](settings, commonModel.GET_SETTINGS_SUCCESS))
}

func (settingHandler *SettingHandler) UpdateSettings(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	// 解析请求体中的参数
	var newSettings model.SystemSettingDto
	if err := ctx.ShouldBindJSON(&newSettings); err != nil {
		ctx.JSON(http.StatusOK, errUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		}))
		return
	}

	if err := settingHandler.settingService.UpdateSetting(userid, &newSettings); err != nil {
		ctx.JSON(http.StatusOK, errUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		}))
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.UPDATE_SETTINGS_SUCCESS))
}
