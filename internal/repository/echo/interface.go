package repository

import model "github.com/lin-snow/ech0/internal/model/echo"

type EchoRepositoryInterface interface {
	CreateEcho(echo *model.Echo) error
}
