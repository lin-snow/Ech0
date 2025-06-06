package service

import model "github.com/lin-snow/ech0/internal/model/connect"

type ConnectServiceInterface interface {
	AddConnect(userid uint, connected model.Connected) error
	DeleteConnect(userid, id uint) error
	GetConnect() (model.Connect, error)
	GetConnectsInfo() ([]model.Connect, error)
	GetConnects() ([]model.Connected, error)
}
