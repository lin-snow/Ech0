package handler

import (
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
)

type CommonHandler struct {
	commonService service.CommonServiceInterface
}

func NewCommonHandler(commonService service.CommonServiceInterface) *CommonHandler {
	return &CommonHandler{
		commonService: commonService,
	}
}

func (commonHandler *CommonHandler) UploadImage(ctx *gin.Context) {
	// 提取上传的 File数据
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
		return
	}

	// 提取userid
	userId := ctx.MustGet("userid").(uint)

	// 调用 CommonService 上传文件
	imageUrl, err := commonHandler.commonService.UploadImage(userId, file)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[string](imageUrl, commonModel.UPLOAD_SUCCESS))
}
