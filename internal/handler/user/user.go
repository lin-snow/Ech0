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

func (userHandler *UserHandler) GetAllUsers() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		allusers, err := userHandler.userService.GetAllUsers(ctx.Request.Context())
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: allusers, Msg: commonModel.GET_USER_SUCCESS}
	})
}

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
