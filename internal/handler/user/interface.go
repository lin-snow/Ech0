package handler

import (
	"github.com/gin-gonic/gin"
)

type UserHandlerInterface interface {
	Login(ctx *gin.Context)
	Register(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)
	UpdateUserAdmin(ctx *gin.Context)
	GetAllUsers(ctx *gin.Context)
	DeleteUser(ctx *gin.Context)
	GetUserInfo(ctx *gin.Context)
}
