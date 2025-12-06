package handler

import "github.com/gin-gonic/gin"

type CommonHandlerInterface interface {
	// ShowImage 显示图片
	// ShowImage() gin.HandlerFunc

	// UploadImage 上传图片
	UploadImage() gin.HandlerFunc

	// DeleteImage 删除图片
	DeleteImage() gin.HandlerFunc

	// GetStatus 获取Echo状态
	GetStatus() gin.HandlerFunc

	// GetHeatMap 获取热力图
	GetHeatMap() gin.HandlerFunc

	// UploadAudio 上传音频
	UploadAudio() gin.HandlerFunc

	// DeleteAudio 删除音频
	DeleteAudio() gin.HandlerFunc

	// GetPlayMusic 获取可播放的音乐
	GetPlayMusic() gin.HandlerFunc

	// HelloEch0 一个简单的获取一些有关 Ech0 的信息的接口
	HelloEch0() gin.HandlerFunc

	// GetRss 获取RSS
	GetRss(ctx *gin.Context)

	// PlayMusic 播放音乐
	PlayMusic(ctx *gin.Context)

	// GetS3PresignURL 获取 S3 预签名 URL
	GetS3PresignURL() gin.HandlerFunc

	// GetMetrics 获取系统指标
	GetMetrics() gin.HandlerFunc

	// GetWebsiteTitle 获取网站标题
	GetWebsiteTitle() gin.HandlerFunc

	// UploadModel 上传3D模型
	UploadModel() gin.HandlerFunc

	// DeleteModel 删除3D模型
	DeleteModel() gin.HandlerFunc
}
