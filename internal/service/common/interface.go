package service

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"

	model "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	storageUtil "github.com/lin-snow/ech0/internal/util/storage"
)

type CommonServiceInterface interface {
	// CommonGetUserByUserId 根据用户ID获取用户信息
	CommonGetUserByUserId(userId uint) (userModel.User, error)

	// UploadImage 上传图片
	UploadImage(userid uint, file *multipart.FileHeader) (string, error)

	// DeleteImage 删除图片
	DeleteImage(userid uint, url, source, object_key string) error

	// DirectDeleteImage 直接根据URL和来源删除图片
	DirectDeleteImage(url, source, object_key string) error

	// GetSysAdmin 获取系统管理员
	GetSysAdmin() (userModel.User, error)

	// GetStatus 获取系统状态
	GetStatus() (model.Status, error)

	// GetHeatMap 获取热力图数据
	GetHeatMap() ([]model.Heatmap, error)

	// GenerateRSS 生成RSS订阅链接
	GenerateRSS(ctx *gin.Context) (string, error)

	// UploadMusic 上传音乐文件
	UploadMusic(userId uint, file *multipart.FileHeader) (string, error)

	// DeleteMusic 删除音乐文件
	DeleteMusic(userid uint) error

	// UploadModel 上传3D模型文件
	UploadModel(userId uint, file *multipart.FileHeader) (string, error)

	// DeleteModel 删除3D模型文件
	DeleteModel(userid uint, url string) error

	// GetPlayMusicUrl 获取可播放的音乐URL
	GetPlayMusicUrl() string

	// PlayMusic 播放音乐
	PlayMusic(ctx *gin.Context)

	// GetS3PresignURL 获取 S3 预签名 URL
	GetS3PresignURL(userid uint, s3Dto *model.GetPresignURLDto, method string) (model.PresignDto, error)

	// GetS3Client 获取 S3 客户端和配置信息
	GetS3Client() (storageUtil.ObjectStorage, settingModel.S3Setting, error)

	// GetS3ObjectURL 获取 S3 对象的访问 URL
	GetS3ObjectURL(s3setting settingModel.S3Setting, objectKey string) (string, error)

	// CleanupTempFiles 清理过期的临时文件
	CleanupTempFiles() error

	// RefreshEchoImageURL 刷新 Echo 中的图片 URL
	RefreshEchoImageURL(echo *echoModel.Echo)

	// GetWebsiteTitle 获取网站标题
	GetWebsiteTitle(websiteURL string) (string, error)
}
