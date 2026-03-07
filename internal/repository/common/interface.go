package repository

import (
	"context"
	"time"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type CommonRepositoryInterface interface {
	GetUserByUserId(ctx context.Context, userid uint) (userModel.User, error)
	GetSysAdmin(ctx context.Context) (userModel.User, error)
	GetAllUsers(ctx context.Context) ([]userModel.User, error)
	GetAllEchos(ctx context.Context, showPrivate bool) ([]echoModel.Echo, error)
	GetHeatMap(ctx context.Context, startUTC, endUTC time.Time) ([]time.Time, error)
}
