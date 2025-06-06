package handler

import "github.com/gin-gonic/gin"

type ConnectHandlerInterface interface {
	AddConnect(ctx *gin.Context)
	DeleteConnect(ctx *gin.Context)
	GetConnectsInfo(ctx *gin.Context)
	GetConnect(ctx *gin.Context)
	GetConnects(ctx *gin.Context)
}
