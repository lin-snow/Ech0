package service

import (
	"github.com/lin-snow/ech0/internal/model/echo"
)

type EchoServiceInterface interface {
	PostEcho(userid uint, newEcho *model.Echo) error
}
