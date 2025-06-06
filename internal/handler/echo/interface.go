package handler

import "github.com/gin-gonic/gin"

type EchoHandlerInterface interface {
	PostEcho(ctx *gin.Context)
	GetEchosByPage(ctx *gin.Context)
	DeleteEcho(ctx *gin.Context)
}
