package service

import (
	"context"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type Service interface {
	CommonGetUserByUserId(ctx context.Context, userId string) (userModel.User, error)
	GetOwner() (userModel.User, error)
	GetHeatMap(timezone string) ([]commonModel.Heatmap, error)
	GenerateRSS(ctx *gin.Context) (string, error)
	GetWebsiteTitle(websiteURL string) (string, error)
}

type CommonRepository interface {
	GetUserByUserId(ctx context.Context, id string) (userModel.User, error)
	GetOwner(ctx context.Context) (userModel.User, error)
	GetAllEchos(ctx context.Context, showPrivate bool) ([]echoModel.Echo, error)
	GetHeatMap(ctx context.Context, startTime, endTime int64) ([]int64, error)
	TrackRSSCacheKey(cacheKey string)
}
