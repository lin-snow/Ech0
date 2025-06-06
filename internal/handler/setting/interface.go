package handler

import "github.com/gin-gonic/gin"

type SettingHandlerInterface interface {
	GetSettings(ctx *gin.Context)
	UpdateSettings(ctx *gin.Context)
}
