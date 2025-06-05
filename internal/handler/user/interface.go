package handler

import "github.com/gin-gonic/gin"

type UserHandlerInterface interface {
	Login(ctx *gin.Context)
	Register(ctx *gin.Context)
}
