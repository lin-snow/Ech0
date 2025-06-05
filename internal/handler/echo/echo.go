package handler

import (
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	service "github.com/lin-snow/ech0/internal/service/echo"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
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
