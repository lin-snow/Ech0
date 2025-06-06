package handler

import (
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	service "github.com/lin-snow/ech0/internal/service/echo"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
	"strconv"
)

type EchoHandler struct {
	echoService service.EchoServiceInterface
}

func NewEchoHandler(echoService service.EchoServiceInterface) *EchoHandler {
	return &EchoHandler{
		echoService: echoService,
	}
}

func (echoHandler *EchoHandler) PostEcho(ctx *gin.Context) {
	var newEcho model.Echo
	if err := ctx.ShouldBindJSON(&newEcho); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
		return
	}

	userId := ctx.MustGet("userid").(uint)
	if err := echoHandler.echoService.PostEcho(userId, &newEcho); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[string](commonModel.POST_ECHO_SUCCESS))
}

func (echoHandler *EchoHandler) GetEchosByPage(ctx *gin.Context) {
	// 获取分页参数
	var pageRequest commonModel.PageQueryDto
	if err := ctx.ShouldBindQuery(&pageRequest); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
		return
	}

	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)
	result, err := echoHandler.echoService.GetEchosByPage(userid, pageRequest)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[commonModel.PageQueryResult[[]model.Echo]](result, commonModel.GET_ECHOS_BY_PAGE_SUCCESS))
}

func (echoHandler *EchoHandler) DeleteEcho(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	// 从 URL 参数获取留言 ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](commonModel.INVALID_PARAMS))
		return
	}

	if err := echoHandler.echoService.DeleteEchoById(userid, uint(id)); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.DELETE_ECHO_SUCCESS))
}
