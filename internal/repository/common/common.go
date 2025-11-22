package repository

import (
	"context"

	"gorm.io/gorm"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
)

type CommonRepository struct {
	db func() *gorm.DB
}

func NewCommonRepository(dbProvider func() *gorm.DB) CommonRepositoryInterface {
	return &CommonRepository{
		db: dbProvider,
	}
}

// getDB 从上下文中获取事务
func (commonRepository *CommonRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return commonRepository.db()
}

// GetUserByUserId 根据用户ID获取用户信息
func (commonRepository *CommonRepository) GetUserByUserId(userId uint) (userModel.User, error) {
	var user userModel.User
	if err := commonRepository.db().First(&user, userId).Error; err != nil {
		return user, err
	}
	return user, nil
}

// GetSysAdmin 获取系统管理员信息
func (commonRepository *CommonRepository) GetSysAdmin() (userModel.User, error) {
	// 获取系统管理员（首个注册的用户）
	user := userModel.User{}
	err := commonRepository.db().Where("is_admin = ?", true).First(&user).Error
	if err != nil {
		return userModel.User{}, err
	}

	return user, nil
}

// GetAllUsers 获取所有用户信息
func (commonRepository *CommonRepository) GetAllUsers() ([]userModel.User, error) {
	var users []userModel.User
	err := commonRepository.db().Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetAllEchos 获取所有Echo
func (commonRepository *CommonRepository) GetAllEchos(showPrivate bool) ([]echoModel.Echo, error) {
	var echos []echoModel.Echo

	// 是否将私密内容也查询出来
	if showPrivate {
		if err := commonRepository.db().Preload("Images").Joins("User").Order("created_at DESC").Find(&echos).Error; err != nil {
			return nil, err
		}
	} else {
		if err := commonRepository.db().Preload("Images").Joins("User").Where("private = ?", false).Find(&echos).Error; err != nil {
			return nil, err
		}
	}

	return echos, nil
}

// GetHeatMap 获取热力图数据
func (commonRepository *CommonRepository) GetHeatMap(startDate, endDate string) ([]commonModel.Heatmap, error) {
	var results []commonModel.Heatmap

	// 查询数据
	// 执行查询
	err := commonRepository.db().Table("echos").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("DATE(created_at) >= ? AND DATE(created_at) <= ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// SaveTempFile 保存临时文件记录
func (commonRepository *CommonRepository) SaveTempFile(ctx context.Context, file commonModel.TempFile) error {
	return commonRepository.getDB(ctx).Create(&file).Error
}

// DeleteTempFile 删除临时文件记录
func (commonRepository *CommonRepository) DeleteTempFile(ctx context.Context, id uint) error {
	return commonRepository.getDB(ctx).Model(&commonModel.TempFile{}).Where("id = ?", id).Update("deleted", true).Error
}

// DeleteTempFilePermanently 永久删除临时文件记录
func (commonRepository *CommonRepository) DeleteTempFilePermanently(ctx context.Context, id uint) error {
	return commonRepository.getDB(ctx).Delete(&commonModel.TempFile{}, id).Error
}

// DeleteTempFileByObjectKey 根据对象键删除临时文件记录（有则删除，没有则跳过）
func (commonRepository *CommonRepository) DeleteTempFileByObjectKey(ctx context.Context, objectKey string) error {
	result := commonRepository.getDB(ctx).Where("object_key = ?", objectKey).Delete(&commonModel.TempFile{})
	if result.Error != nil {
		return result.Error
	}
	// 没有找到记录也不算错误，直接返回 nil
	return nil
}

// GetAllTempFiles 获取所有未删除的临时文件
func (commonRepository *CommonRepository) GetAllTempFiles() ([]commonModel.TempFile, error) {
	var files []commonModel.TempFile
	err := commonRepository.db().Where("deleted = ?", false).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// UpdateTempFileAccessTime 更新临时文件的最后访问时间
func (commonRepository *CommonRepository) UpdateTempFileAccessTime(
	ctx context.Context,
	id uint,
	accessTime int64,
) error {
	return commonRepository.getDB(ctx).
		Model(&commonModel.TempFile{}).
		Where("id = ?", id).
		Update("last_accessed_at", accessTime).
		Error
}
