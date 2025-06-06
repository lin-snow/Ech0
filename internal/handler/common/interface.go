package handler

import "github.com/gin-gonic/gin"

type CommonHandlerInterface interface {
	UploadImage(ctx *gin.Context)
}
