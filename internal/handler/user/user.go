package handler

import (
	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	service "github.com/lin-snow/ech0/internal/service/user"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
	"strconv"
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
		return
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
		return
	}

	// 调用 Service 层处理注册
	if err := userHandler.userService.Register(&registerDto); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.REGISTER_SUCCESS))
}

func (userHandler *UserHandler) UpdateUser(ctx *gin.Context) {
	// 解析用户请求体中的参数
	var userdto model.UserInfoDto
	if err := ctx.ShouldBindJSON(&userdto); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
		return
	}

	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)
	if err := userHandler.userService.UpdateUser(userid, userdto); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.UPDATE_USER_SUCCESS))
}

// UpdateUserAdmin 更新用户权限
func (userHandler *UserHandler) UpdateUserAdmin(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	if err := userHandler.userService.UpdateUserAdmin(userid); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.UPDATE_USER_SUCCESS))

}

// GetAllUsers 获取所有用户
func (userHandler *UserHandler) GetAllUsers(ctx *gin.Context) {
	allusers, err := userHandler.userService.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[[]model.User](allusers, commonModel.GET_USER_SUCCESS))
}

func (userHandler *UserHandler) DeleteUser(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_PARAMS,
			Err: err,
		})))
		return
	}

	if err := userHandler.userService.DeleteUser(userid, uint(id)); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.DELETE_USER_SUCCESS))
}

func (userHandler *UserHandler) GetUserInfo(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	// 调用 Service 层获取用户信息
	user, err := userHandler.userService.GetUserByID(int(userid))
	user.Password = "" // 不返回密码
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, commonModel.OK[model.User](user, commonModel.GET_USER_INFO_SUCCESS))
}
