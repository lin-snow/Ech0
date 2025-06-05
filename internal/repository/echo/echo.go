package repository

import (
	model "github.com/lin-snow/ech0/internal/model/echo"
	"gorm.io/gorm"
	"strings"
)

type EchoRepository struct {
	db *gorm.DB
}

func NewEchoRepository(db *gorm.DB) EchoRepositoryInterface {
	return &EchoRepository{db: db}
}

func (echoRepository *EchoRepository) CreateEcho(echo *model.Echo) error {
	echo.Content = strings.TrimSpace(echo.Content)

	result := echoRepository.db.Create(echo)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
