package handler

import "github.com/gin-gonic/gin"

type CommonHandlerInterface interface {
	UploadImage(ctx *gin.Context)
	DeleteImage(ctx *gin.Context)
	GetStatus(ctx *gin.Context)
}
