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
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/transaction"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"gorm.io/gorm"
)

type EchoRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

var _ echoService.Repository = (*EchoRepository)(nil)

func NewEchoRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) *EchoRepository {
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

	return nil
}

func (echoRepository *EchoRepository) InvalidateEchoCaches(echoIDs ...string) {
	ClearEchoPageCache(echoRepository.cache)
	ClearTodayEchosCache(echoRepository.cache)
	ClearRSSCache(echoRepository.cache)
	for _, id := range echoIDs {
		echoRepository.cache.Delete(GetEchoByIDCacheKey(id))
	}
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
				Preload("Extension").
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

func (echoRepository *EchoRepository) GetEchosById(ctx context.Context, id string) (*model.Echo, error) {
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
				Preload("Extension").
				Preload("Tags").
				Where("id = ?", id).
				First(&row)
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
				Preload("Extension").
				Preload("Tags").
				Where("id = ?", id).
				First(&row)
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

func (echoRepository *EchoRepository) DeleteEchoById(ctx context.Context, id string) error {
	var echo model.Echo
	echoRepository.getDB(ctx).Where("echo_id = ?", id).Delete(&fileModel.EchoFile{})

	result := echoRepository.getDB(ctx).Where("id = ?", id).Delete(&echo)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

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
			startOfDayUTC := startOfDayUser.UTC().Unix()
			endOfDayUTC := endOfDayUser.UTC().Unix()

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
				Preload("Extension").
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
	if err := echoRepository.getDB(ctx).Where("echo_id = ?", echo.ID).Delete(&fileModel.EchoFile{}).Error; err != nil {
		return err
	}

	if err := echoRepository.getDB(ctx).Model(&model.Echo{}).
		Where("id = ?", echo.ID).
		Updates(map[string]interface{}{
			"content": echo.Content,
			"private": echo.Private,
			"layout":  echo.Layout,
		}).Error; err != nil {
		return err
	}

	if err := echoRepository.getDB(ctx).Where("echo_id = ?", echo.ID).Delete(&model.EchoExtension{}).Error; err != nil {
		return err
	}
	if echo.Extension != nil {
		echo.Extension.EchoID = echo.ID
		if err := echoRepository.getDB(ctx).Create(echo.Extension).Error; err != nil {
			return err
		}
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

func (echoRepository *EchoRepository) LikeEcho(ctx context.Context, id string) error {
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

func (echoRepository *EchoRepository) DeleteTagById(ctx context.Context, id string) error {
	var tag model.Tag

	if err := echoRepository.getDB(ctx).Where("tag_id = ?", id).Delete(&model.EchoTag{}).Error; err != nil {
		return err
	}

	result := echoRepository.getDB(ctx).Where("id = ?", id).Delete(&tag)
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
	tagID string,
) error {
	return echoRepository.getDB(ctx).Model(&model.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

func (echoRepository *EchoRepository) QueryEchos(
	queryDto commonModel.EchoQueryDto,
	showPrivate bool,
) ([]model.Echo, int64, error) {
	var (
		echos []model.Echo
		total int64
	)

	hasTagFilter := len(queryDto.TagIDs) > 0

	sortColumn := "echos.created_at"
	if queryDto.SortBy == "fav_count" {
		sortColumn = "echos.fav_count"
	}
	sortDir := "DESC"
	if queryDto.SortOrder == "asc" {
		sortDir = "ASC"
	}
	orderClause := sortColumn + " " + sortDir

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if hasTagFilter {
			db = db.Joins("JOIN echo_tags ON echo_tags.echo_id = echos.id").
				Where("echo_tags.tag_id IN ?", queryDto.TagIDs)
		}
		if !showPrivate {
			db = db.Where("echos.private = ?", false)
		}
		if queryDto.Search != "" {
			db = db.Where("echos.content LIKE ?", "%"+queryDto.Search+"%")
		}
		return db
	}

	countQuery := applyFilters(echoRepository.db().Model(&model.Echo{}))
	if hasTagFilter {
		if err := countQuery.Distinct("echos.id").Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := countQuery.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	offset := (queryDto.Page - 1) * queryDto.PageSize

	if hasTagFilter {
		var echoIDs []string
		idsQuery := applyFilters(echoRepository.db().Model(&model.Echo{}))
		if err := idsQuery.
			Distinct("echos.id").
			Order(orderClause).
			Limit(queryDto.PageSize).
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
			Preload("Extension").
			Preload("Tags").
			Order(orderClause).
			Find(&echos).Error; err != nil {
			return nil, 0, err
		}
	} else {
		query := applyFilters(echoRepository.db().Model(&model.Echo{}))
		if err := query.
			Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
				return db.Order("echo_files.sort_order ASC")
			}).
			Preload("EchoFiles.File").
			Preload("Extension").
			Preload("Tags").
			Limit(queryDto.PageSize).
			Offset(offset).
			Order(orderClause).
			Find(&echos).Error; err != nil {
			return nil, 0, err
		}
	}

	return echos, total, nil
}

func (echoRepository *EchoRepository) GetEchosByTagId(
	tagId string,
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

	var echoIDs []string
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
		Preload("Extension").
		Preload("Tags").
		Order("created_at DESC").
		Find(&echos).Error; err != nil {
		return nil, 0, err
	}

	return echos, total, nil
}

func (echoRepository *EchoRepository) GetHotEchos(limit int, showPrivate bool) ([]model.Echo, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	const recentPool = 10

	recentQuery := echoRepository.db().Model(&model.Echo{}).
		Select("id").
		Order("created_at DESC").
		Limit(recentPool)
	if !showPrivate {
		recentQuery = recentQuery.Where("private = ?", false)
	}

	type hotRow struct {
		ID string
	}

	hotQuery := echoRepository.db().Table("(?) AS recent", recentQuery).
		Select("recent.id, echos.fav_count + COUNT(comments.id) * 2 AS hot_score").
		Joins("JOIN echos ON echos.id = recent.id").
		Joins("LEFT JOIN comments ON comments.echo_id = recent.id AND comments.status = 'approved'").
		Group("recent.id").
		Order("hot_score DESC").
		Limit(limit)

	var rows []hotRow
	if err := hotQuery.Find(&rows).Error; err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []model.Echo{}, nil
	}

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}

	var echos []model.Echo
	if err := echoRepository.db().
		Where("id IN ?", ids).
		Preload("EchoFiles", func(db *gorm.DB) *gorm.DB {
			return db.Order("echo_files.sort_order ASC")
		}).
		Preload("EchoFiles.File").
		Preload("Extension").
		Preload("Tags").
		Find(&echos).Error; err != nil {
		return nil, err
	}

	idOrder := make(map[string]int, len(ids))
	for i, id := range ids {
		idOrder[id] = i
	}
	sorted := make([]model.Echo, len(echos))
	for _, e := range echos {
		sorted[idOrder[e.ID]] = e
	}

	return sorted, nil
}
