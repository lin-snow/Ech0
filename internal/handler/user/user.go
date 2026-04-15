package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	service "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type UserHandler struct {
	userService service.Service
}

func NewUserHandler(userService service.Service) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register 用户注册
//
//	@Summary		用户注册
//	@Description	注册新用户账号
//	@Tags			用户
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.RegisterDto	true	"注册信息"
//	@Success		200		{object}	handler.Response	"注册成功"
//	@Failure		200		{object}	handler.Response	"注册失败"
//	@Router			/register [post]
func (userHandler *UserHandler) Register() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var registerDto authModel.RegisterDto
		if err := ctx.ShouldBindJSON(&registerDto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := userHandler.userService.Register(&registerDto); err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.REGISTER_SUCCESS}
	})
}

// UpdateUser 更新当前用户信息
//
//	@Summary		更新用户信息
//	@Description	更新当前已认证用户的个人信息
//	@Tags			用户
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.UserInfoDto	true	"用户信息"
//	@Success		200		{object}	handler.Response	"更新成功"
//	@Failure		200		{object}	handler.Response	"更新失败"
//	@Router			/user [put]
func (userHandler *UserHandler) UpdateUser() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var userdto model.UserInfoDto
		if err := ctx.ShouldBindJSON(&userdto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		if err := userHandler.userService.UpdateUser(ctx.Request.Context(), userdto); err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.UPDATE_USER_SUCCESS}
	})
}

// UpdateUserAdmin 切换用户管理员权限
//
//	@Summary		切换用户管理员权限
//	@Description	由管理员切换指定用户的管理员状态
//	@Tags			用户
//	@Produce		json
//	@Param			id	path		string				true	"用户 ID (UUID)"
//	@Success		200	{object}	handler.Response	"更新成功"
//	@Failure		200	{object}	handler.Response	"更新失败"
//	@Router			/user/admin/{id} [put]
func (userHandler *UserHandler) UpdateUserAdmin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		idStr := ctx.Param("id")
		if _, err := uuid.Parse(idStr); err != nil {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		if err := userHandler.userService.UpdateUserAdmin(ctx.Request.Context(), idStr); err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.UPDATE_USER_SUCCESS}
	})
}

// GetAllUsers 获取所有用户列表
//
//	@Summary		获取所有用户
//	@Description	管理员获取系统中所有用户的列表
//	@Tags			用户
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=[]model.User}	"获取成功"
//	@Failure		200	{object}	handler.Response					"获取失败"
//	@Router			/users [get]
func (userHandler *UserHandler) GetAllUsers() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		allusers, err := userHandler.userService.GetAllUsers(ctx.Request.Context())
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: allusers, Msg: commonModel.GET_USER_SUCCESS}
	})
}

// DeleteUser 删除用户
//
//	@Summary		删除用户
//	@Description	管理员根据 ID 删除指定用户
//	@Tags			用户
//	@Produce		json
//	@Param			id	path		string				true	"用户 ID (UUID)"
//	@Success		200	{object}	handler.Response	"删除成功"
//	@Failure		200	{object}	handler.Response	"删除失败"
//	@Router			/user/{id} [delete]
func (userHandler *UserHandler) DeleteUser() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		idStr := ctx.Param("id")
		if _, err := uuid.Parse(idStr); err != nil {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		if err := userHandler.userService.DeleteUser(ctx.Request.Context(), idStr); err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.DELETE_USER_SUCCESS}
	})
}

// GetUserInfo 获取当前用户信息
//
//	@Summary		获取当前用户信息
//	@Description	获取当前已认证用户的详细信息
//	@Tags			用户
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=model.User}	"获取成功"
//	@Failure		200	{object}	handler.Response					"获取失败"
//	@Router			/user [get]
func (userHandler *UserHandler) GetUserInfo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := viewer.MustFromContext(ctx.Request.Context()).UserID()
		user, err := userHandler.userService.GetUserByID(userid)
		user.Password = ""
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: user, Msg: commonModel.GET_USER_INFO_SUCCESS}
	})
}
