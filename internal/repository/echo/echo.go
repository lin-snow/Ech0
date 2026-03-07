package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/transaction"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"gorm.io/gorm"
)

type EchoRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewEchoRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) EchoRepositoryInterface {
	return &EchoRepository{db: dbProvider, cache: cache}
}

func (echoRepository *EchoRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return echoRepository.db()
}

func (echoRepository *EchoRepository) CreateEcho(ctx context.Context, echo *model.Echo) error {
	echo.Content = strings.TrimSpace(echo.Content)

	result := echoRepository.getDB(ctx).Create(echo)
	if result.Error != nil {
		return result.Error
	}

	ClearEchoPageCache(echoRepository.cache)
	ClearTodayEchosCache(echoRepository.cache)

	return nil
}

func (echoRepository *EchoRepository) GetEchosByPage(
	page, pageSize int,
	search string,
	showPrivate bool,
) ([]model.Echo, int64) {
	cacheKey := GetEchoPageCacheKey(page, pageSize, search, showPrivate)
	pageResult, err := cache.ReadThroughTyped[commonModel.PageQueryResult[[]model.Echo]](
		echoRepository.cache,
		cacheKey,
		1,
		func() (commonModel.PageQueryResult[[]model.Echo], error) {
			offset := (page - 1) * pageSize
			var echos []model.Echo
			var total int64

			query := echoRepository.db().Model(&model.Echo{})
			if search != "" {
				query = query.Where("content LIKE ?", "%"+search+"%")
			}
			if !showPrivate {
				query = query.Where("private = ?", false)
			}

			if dbErr := query.Count(&total).
				Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
					return db.Order("echo_files.sort_order ASC")
				}).
				Preload("EchoFiles.File").
				Preload("Tags").
				Limit(pageSize).
				Offset(offset).
				Order("created_at DESC").
				Find(&echos).Error; dbErr != nil {
				return commonModel.PageQueryResult[[]model.Echo]{}, dbErr
			}

			TrackEchoPageCacheKey(cacheKey)
			return commonModel.PageQueryResult[[]model.Echo]{
				Items: echos,
				Total: total,
			}, nil
		},
	)
	if err != nil {
		return []model.Echo{}, 0
	}
	return pageResult.Items, pageResult.Total
}

func (echoRepository *EchoRepository) GetEchosById(ctx context.Context, id uint) (*model.Echo, error) {
	cacheKey := GetEchoByIDCacheKey(id)
	echo, err := cache.ReadThroughTypedUnlessTx[*model.Echo](
		ctx,
		echoRepository.cache,
		cacheKey,
		1,
		func(ctx context.Context) (*model.Echo, error) {
			var row model.Echo
			result := echoRepository.getDB(ctx).
				Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
					return db.Order("echo_files.sort_order ASC")
				}).
				Preload("EchoFiles.File").
				Preload("Tags").
				First(&row, id)
			if result.Error != nil {
				return nil, result.Error
			}
			return &row, nil
		},
		func() (*model.Echo, error) {
			var row model.Echo
			result := echoRepository.db().
				Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
					return db.Order("echo_files.sort_order ASC")
				}).
				Preload("EchoFiles.File").
				Preload("Tags").
				First(&row, id)
			if result.Error != nil {
				return nil, result.Error
			}
			return &row, nil
		})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return echo, nil
}

func (echoRepository *EchoRepository) DeleteEchoById(ctx context.Context, id uint) error {
	var echo model.Echo
	echoRepository.getDB(ctx).Where("echo_id = ?", id).Delete(&fileModel.EchoFile{})

	result := echoRepository.getDB(ctx).Delete(&echo, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	echoRepository.cache.Delete(GetEchoByIDCacheKey(id))
	ClearTodayEchosCache(echoRepository.cache)
	ClearEchoPageCache(echoRepository.cache)

	return nil
}

func (echoRepository *EchoRepository) GetTodayEchos(showPrivate bool, timezone string) []model.Echo {
	normalizedTimezone := timezoneUtil.NormalizeTimezone(timezone)

	cacheKey := GetTodayEchosCacheKey(showPrivate, normalizedTimezone)
	todayEchos, err := cache.ReadThroughTypedWithStore[[]model.Echo](
		echoRepository.cache,
		cacheKey,
		func(echos []model.Echo) {
			loc := timezoneUtil.LoadLocationOrUTC(normalizedTimezone)
			nowUser := time.Now().UTC().In(loc)
			startOfDayUser := time.Date(nowUser.Year(), nowUser.Month(), nowUser.Day(), 0, 0, 0, 0, loc)
			endOfDayUser := startOfDayUser.Add(24 * time.Hour)
			ttl := time.Until(endOfDayUser)
			if ttl <= 0 {
				ttl = time.Minute
			}
			TrackTodayEchosCacheKey(cacheKey)
			echoRepository.cache.SetWithTTL(cacheKey, echos, 1, ttl)
		},
		func() ([]model.Echo, error) {
			var echos []model.Echo

			loc := timezoneUtil.LoadLocationOrUTC(normalizedTimezone)
			nowUser := time.Now().UTC().In(loc)
			startOfDayUser := time.Date(nowUser.Year(), nowUser.Month(), nowUser.Day(), 0, 0, 0, 0, loc)
			endOfDayUser := startOfDayUser.Add(24 * time.Hour)
			startOfDayUTC := startOfDayUser.UTC()
			endOfDayUTC := endOfDayUser.UTC()

			query := echoRepository.db().Model(&model.Echo{})
			if !showPrivate {
				query = query.Where("private = ?", false)
			}
			query = query.Where("created_at >= ? AND created_at < ?", startOfDayUTC, endOfDayUTC)
			if err := query.
				Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
					return db.Order("echo_files.sort_order ASC")
				}).
				Preload("EchoFiles.File").
				Preload("Tags").
				Order("created_at DESC").
				Find(&echos).Error; err != nil {
				return nil, err
			}

			return echos, nil
		})
	if err != nil {
		return []model.Echo{}
	}
	return todayEchos
}

func (echoRepository *EchoRepository) UpdateEcho(ctx context.Context, echo *model.Echo) error {
	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetEchoByIDCacheKey(echo.ID))
	ClearTodayEchosCache(echoRepository.cache)

	if err := echoRepository.getDB(ctx).Where("echo_id = ?", echo.ID).Delete(&fileModel.EchoFile{}).Error; err != nil {
		return err
	}

	if err := echoRepository.getDB(ctx).Model(&model.Echo{}).
		Where("id = ?", echo.ID).
		Updates(map[string]interface{}{
			"content":        echo.Content,
			"private":        echo.Private,
			"layout":         echo.Layout,
			"extension":      echo.Extension,
			"extension_type": echo.ExtensionType,
		}).Error; err != nil {
		return err
	}

	if len(echo.EchoFiles) > 0 {
		var echoFiles []fileModel.EchoFile
		for _, ef := range echo.EchoFiles {
			ef.EchoID = echo.ID
			echoFiles = append(echoFiles, ef)
		}
		if err := echoRepository.getDB(ctx).Create(&echoFiles).Error; err != nil {
			return err
		}
	}

	if err := echoRepository.getDB(ctx).Model(echo).Association("Tags").Replace(echo.Tags); err != nil {
		return err
	}

	return nil
}

func (echoRepository *EchoRepository) LikeEcho(ctx context.Context, id uint) error {
	var exists bool
	if err := echoRepository.getDB(ctx).
		Model(&model.Echo{}).
		Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return errors.New(commonModel.ECHO_NOT_FOUND)
	}

	if err := echoRepository.getDB(ctx).
		Model(&model.Echo{}).
		Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + ?", 1)).Error; err != nil {
		return err
	}

	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetEchoByIDCacheKey(id))
	ClearTodayEchosCache(echoRepository.cache)

	return nil
}

func (echoRepository *EchoRepository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	result := echoRepository.db().Order("usage_count DESC, created_at DESC").Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

func (echoRepository *EchoRepository) DeleteTagById(ctx context.Context, id uint) error {
	var tag model.Tag

	if err := echoRepository.getDB(ctx).Where("tag_id = ?", id).Delete(&model.EchoTag{}).Error; err != nil {
		return err
	}

	result := echoRepository.getDB(ctx).Delete(&tag, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (echoRepository *EchoRepository) GetTagByName(name string) (*model.Tag, error) {
	var tag model.Tag
	result := echoRepository.db().Where("name = ?", name).First(&tag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tag, nil
}

func (echoRepository *EchoRepository) GetTagsByNames(ctx context.Context, names []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	result := echoRepository.getDB(ctx).Where("name IN ?", names).Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

func (echoRepository *EchoRepository) CreateTag(ctx context.Context, tag *model.Tag) error {
	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		return errors.New("标签名称不能为空")
	}

	result := echoRepository.getDB(ctx).Create(tag)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (echoRepository *EchoRepository) IncrementTagUsageCount(
	ctx context.Context,
	tagID uint,
) error {
	return echoRepository.getDB(ctx).Model(&model.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

func (echoRepository *EchoRepository) GetEchosByTagId(
	tagId uint,
	page, pageSize int,
	search string,
	showPrivate bool,
) ([]model.Echo, int64, error) {
	var (
		echos []model.Echo
		total int64
	)

	applyFilters := func(db *gorm.DB) *gorm.DB {
		db = db.Joins("JOIN echo_tags ON echo_tags.echo_id = echos.id").
			Where("echo_tags.tag_id = ?", tagId)

		if !showPrivate {
			db = db.Where("echos.private = ?", false)
		}

		if search != "" {
			db = db.Where("echos.content LIKE ?", "%"+search+"%")
		}

		return db
	}

	countQuery := applyFilters(echoRepository.db().Model(&model.Echo{}))

	if err := countQuery.Distinct("echos.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	var echoIDs []uint
	idsQuery := applyFilters(echoRepository.db().Model(&model.Echo{}))
	if err := idsQuery.
		Distinct("echos.id").
		Order("echos.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Pluck("echos.id", &echoIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(echoIDs) == 0 {
		return []model.Echo{}, total, nil
	}

	if err := echoRepository.db().
		Where("id IN ?", echoIDs).
		Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
			return db.Order("echo_files.sort_order ASC")
		}).
		Preload("EchoFiles.File").
		Preload("Tags").
		Order("created_at DESC").
		Find(&echos).Error; err != nil {
		return nil, 0, err
	}

	return echos, total, nil
}
