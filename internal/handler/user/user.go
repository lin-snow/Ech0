package handler

import (
	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/user"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Login 用户登陆
func (userHandler *UserHandler) Login(ctx *gin.Context) {
	// 从请求体获取用户名和密码
	var loginDto authModel.LoginDto
	if err := ctx.ShouldBindJSON(&loginDto); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
		return
	}

	// 调用 Service 层处理登陆
	token, err := userHandler.userService.Login(&loginDto)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
	}

	// 返回成功响应， 包含 JWT Token
	ctx.JSON(http.StatusOK, commonModel.OK(token, commonModel.LOGIN_SUCCESS))
}

// Register 用户注册
func (userHandler *UserHandler) Register(ctx *gin.Context) {
	var registerDto authModel.RegisterDto
	if err := ctx.ShouldBindJSON(&registerDto); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
	}

	// 调用 Service 层处理注册
	if err := userHandler.userService.Register(&registerDto); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.REGISTER_SUCCESS))
}
