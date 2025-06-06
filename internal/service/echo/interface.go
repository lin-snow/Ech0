package service

import (
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/model/echo"
)

type EchoServiceInterface interface {
	PostEcho(userid uint, newEcho *model.Echo) error
	GetEchosByPage(userid uint, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	DeleteEchoById(userid, id uint) error
}
