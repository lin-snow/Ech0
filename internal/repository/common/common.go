package repository

import (
	"github.com/lin-snow/ech0/internal/database"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
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

func (commonRepository *CommonRepository) GetSysAdmin() (userModel.User, error) {
	// 获取系统管理员（首个注册的用户）
	user := userModel.User{}
	err := database.DB.Where("is_admin = ?", true).First(&user).Error
	if err != nil {
		return userModel.User{}, err
	}

	return user, nil
}

func (commonRepository *CommonRepository) GetAllUsers() ([]userModel.User, error) {
	var users []userModel.User
	err := database.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (commonRepository *CommonRepository) GetAllEchos(showPrivate bool) ([]echoModel.Echo, error) {
	var echos []echoModel.Echo

	// 是否将私密内容也查询出来
	if showPrivate {
		if err := database.DB.Preload("Images").Order("created_at DESC").Find(&echos).Error; err != nil {
			return nil, err
		}
	} else {
		if err := database.DB.Preload("Images").Where("private = ?", false).Find(&echos).Error; err != nil {
			return nil, err
		}
	}

	return echos, nil
}

func (commonRepository *CommonRepository) GetStatus() (commonModel.Status, error) {
	//TODO implement me
	panic("implement me")
}
