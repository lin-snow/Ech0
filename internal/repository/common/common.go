package repository

import (
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

type CommonRepository struct {
	db *gorm.DB
}

func NewCommonRepository(db *gorm.DB) CommonRepositoryInterface {
	return &CommonRepository{
		db: db,
	}
}

func (commonRepository *CommonRepository) GetUserByUserId(userId uint) (userModel.User, error) {
	var user userModel.User
	if err := commonRepository.db.First(&user, userId).Error; err != nil {
		return user, err
	}
	return user, nil
}
