package repository

import (
	"context"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type CommonRepository struct {
	db func() *gorm.DB
}

var _ commonService.CommonRepository = (*CommonRepository)(nil)

func NewCommonRepository(dbProvider func() *gorm.DB) *CommonRepository {
	return &CommonRepository{
		db: dbProvider,
	}
}

func (commonRepository *CommonRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return commonRepository.db()
}

func (commonRepository *CommonRepository) GetUserByUserId(ctx context.Context, userId string) (userModel.User, error) {
	var user userModel.User
	if err := commonRepository.getDB(ctx).Where("id = ?", userId).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (commonRepository *CommonRepository) GetOwner(ctx context.Context) (userModel.User, error) {
	user := userModel.User{}
	err := commonRepository.getDB(ctx).Where("is_owner = ?", true).First(&user).Error
	if err != nil {
		return userModel.User{}, err
	}
	return user, nil
}

func (commonRepository *CommonRepository) GetAllUsers(ctx context.Context) ([]userModel.User, error) {
	var users []userModel.User
	err := commonRepository.getDB(ctx).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (commonRepository *CommonRepository) GetAllEchos(ctx context.Context, showPrivate bool) ([]echoModel.Echo, error) {
	var echos []echoModel.Echo

	query := commonRepository.getDB(ctx).
		Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
			return db.Order("echo_files.sort_order ASC")
		}).
		Preload("EchoFiles.File").
		Preload("Tags").
		Order("created_at DESC")

	if !showPrivate {
		query = query.Where("private = ?", false)
	}

	if err := query.Find(&echos).Error; err != nil {
		return nil, err
	}

	return echos, nil
}

func (commonRepository *CommonRepository) GetHeatMap(
	ctx context.Context,
	startUTC, endUTC int64,
) ([]int64, error) {
	var results []int64

	err := commonRepository.getDB(ctx).
		Table("echos").
		Where("created_at >= ? AND created_at < ?", startUTC, endUTC).
		Order("created_at ASC").
		Pluck("created_at", &results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (commonRepository *CommonRepository) TrackRSSCacheKey(cacheKey string) {
	echoRepository.TrackRSSCacheKey(cacheKey)
}
