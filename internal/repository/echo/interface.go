package repository

import model "github.com/lin-snow/ech0/internal/model/echo"

type EchoRepositoryInterface interface {
	CreateEcho(echo *model.Echo) error
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64)
	GetEchosById(id uint) (*model.Echo, error)
	DeleteEchoById(id uint) error
}
