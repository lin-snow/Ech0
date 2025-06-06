package repository

import (
	"errors"
	"github.com/lin-snow/ech0/internal/database"
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

func (echoRepository *EchoRepository) GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询数据库
	var messages []model.Echo
	var total int64

	query := database.DB.Model(&model.Echo{})

	// 如果 search 不为空，添加模糊查询条件
	if search != "" || len(search) > 0 {
		searchPattern := "%" + search + "%" // 模糊匹配模式
		query = query.Where("content LIKE ?", searchPattern)
	}

	// 如果不是管理员，过滤私密留言
	if !showPrivate {
		query = query.Where("private = ?", false)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	query.Preload("Images").Limit(pageSize).Offset(offset).Order("created_at DESC").Find(&messages)

	// 返回结果
	return messages, total
}

func (echoRepository *EchoRepository) GetEchosById(id uint) (*model.Echo, error) {
	var echo model.Echo
	result := database.DB.Preload("Images").First(&echo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 如果未找到记录，则返回 nil
		}
		return nil, result.Error // 其他错误返回
	}

	return &echo, nil
}

func (echoRepository *EchoRepository) DeleteEchoById(id uint) error {
	var echo model.Echo
	result := database.DB.Delete(&echo, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 如果没有找到记录
	}
	return nil
}
