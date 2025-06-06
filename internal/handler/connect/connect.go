package handler

import (
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	"github.com/lin-snow/ech0/internal/service/connect"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
	"strconv"
)

type ConnectHandler struct {
	connectService service.ConnectServiceInterface
}

func NewConnectHandler(connectService service.ConnectServiceInterface) *ConnectHandler {
	return &ConnectHandler{
		connectService: connectService,
	}
}

func (connectHandler *ConnectHandler) AddConnect(ctx *gin.Context) {
	userId := ctx.MustGet("userid").(uint)

	var connected model.Connected
	if err := ctx.ShouldBindJSON(&connected); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](commonModel.INVALID_REQUEST_BODY))
		return
	}

	if err := connectHandler.connectService.AddConnect(userId, connected); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.ADD_CONNECT_SUCCESS))
}

func (connectHandler *ConnectHandler) DeleteConnect(ctx *gin.Context) {
	userId := ctx.MustGet("userid").(uint)

	// 从 URL 参数获取 ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](commonModel.INVALID_PARAMS))
		return
	}

	if err := connectHandler.connectService.DeleteConnect(userId, uint(id)); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.DELETE_CONNECT_SUCCESS))
}

func (connectHandler *ConnectHandler) GetConnectsInfo(ctx *gin.Context) {
	// 调用 Service 层获取 Connect 信息
	connects, err := connectHandler.connectService.GetConnectsInfo()
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[any](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[[]model.Connect](connects, commonModel.GET_CONNECT_INFO_SUCCESS))
}

func (connectHandler *ConnectHandler) GetConnect(ctx *gin.Context) {
	connect, err := connectHandler.connectService.GetConnect()
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[model.Connect](connect, commonModel.CONNECT_SUCCESS))
}

func (connectHandler *ConnectHandler) GetConnects(ctx *gin.Context) {
	// 调用 Service 层获取 Connect 列表
	connects, err := connectHandler.connectService.GetConnects()
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[[]model.Connected](connects, commonModel.GET_CONNECTED_LIST_SUCCESS))
}
