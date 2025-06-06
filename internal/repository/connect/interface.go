package repository

import model "github.com/lin-snow/ech0/internal/model/connect"

type ConnectRepositoryInterface interface {
	GetAllConnects() ([]model.Connected, error)
	CreateConnect(connect *model.Connected) error
	DeleteConnect(id uint) error
}
