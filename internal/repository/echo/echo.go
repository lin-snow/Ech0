package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
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

// getDB 从上下文中获取事务
func (echoRepository *EchoRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return echoRepository.db()
}

// CreateEcho 创建新的 Echo
func (echoRepository *EchoRepository) CreateEcho(ctx context.Context, echo *model.Echo) error {
	echo.Content = strings.TrimSpace(echo.Content)

	result := echoRepository.getDB(ctx).Create(echo)
	if result.Error != nil {
		return result.Error
	}

	// 清除相关缓存
	ClearEchoPageCache(echoRepository.cache)
	ClearTodayEchosCache(echoRepository.cache)

	return nil
}

// GetEchosByPage 获取分页的 Echo 列表
func (echoRepository *EchoRepository) GetEchosByPage(
	page, pageSize int,
	search string,
	showPrivate bool,
) ([]model.Echo, int64) {
	// 查找缓存
	cacheKey := GetEchoPageCacheKey(page, pageSize, search, showPrivate)
	if cachedResult, err := echoRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，直接返回
		// 类型断言
		cachedResultTyped, ok := cachedResult.(commonModel.PageQueryResult[[]model.Echo])
		if ok {
			return cachedResultTyped.Items, cachedResultTyped.Total
		}
	}

	// 如果缓存未命中，进行数据库查询

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询数据库
	var echos []model.Echo
	var total int64

	query := echoRepository.db().Model(&model.Echo{})

	// 如果 search 不为空，添加模糊查询条件
	if search != "" {
		searchPattern := "%" + search + "%" // 模糊匹配模式
		query = query.Where("content LIKE ?", searchPattern)
	}

	// 如果不是管理员，过滤私密Echo
	if !showPrivate {
		query = query.Where("private = ?", false)
	}

	// 获取总数并进行分页查询
	query.Count(&total).
		Preload("Images").
		Preload("Tags").
		Limit(pageSize).
		Offset(offset).
		Order("created_at DESC").
		Find(&echos)

	// 保存到缓存
	echoKeyList = append(echoKeyList, cacheKey) // 记录缓存键
	echoRepository.cache.Set(cacheKey, commonModel.PageQueryResult[[]model.Echo]{
		Items: echos,
		Total: total,
	}, 1)

	// 返回结果
	return echos, total
}

// GetEchosById 根据 ID 获取 Echo
func (echoRepository *EchoRepository) GetEchosById(id uint) (*model.Echo, error) {
	// 查询缓存
	cacheKey := GetEchoByIDCacheKey(id)
	if cachedEcho, err := echoRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，直接返回
		if echo, ok := cachedEcho.(*model.Echo); ok {
			return echo, nil
		}
	}

	// 缓存未命中，查询数据库
	// 使用 Preload 预加载关联的 Images
	var echo model.Echo
	result := echoRepository.db().Preload("Images").Preload("Tags").First(&echo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 如果未找到记录，则返回 nil
		}
		return nil, result.Error // 其他错误返回
	}

	// 保存到缓存
	echoRepository.cache.Set(cacheKey, &echo, 1)

	return &echo, nil
}

// DeleteEchoById 删除 Echo
func (echoRepository *EchoRepository) DeleteEchoById(ctx context.Context, id uint) error {
	var echo model.Echo
	// 删除外键images
	echoRepository.getDB(ctx).Where("message_id = ?", id).Delete(&model.Image{})

	result := echoRepository.getDB(ctx).Delete(&echo, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 如果没有找到记录
	}

	// 清除缓存
	echoRepository.cache.Delete(GetEchoByIDCacheKey(id))      // 删除具体 Echo 的缓存
	ClearTodayEchosCache(echoRepository.cache)

	// 清除相关缓存
	ClearEchoPageCache(echoRepository.cache)

	return nil
}

// GetTodayEchos 获取今天的 Echo 列表
func (echoRepository *EchoRepository) GetTodayEchos(showPrivate bool, timezone string) []model.Echo {
	normalizedTimezone := timezoneUtil.NormalizeTimezone(timezone)

	// 查找缓存
	if cachedTodayEchos, err := echoRepository.cache.Get(GetTodayEchosCacheKey(showPrivate, normalizedTimezone)); err == nil {
		// 缓存命中，直接返回
		if todayEchos, ok := cachedTodayEchos.([]model.Echo); ok {
			return todayEchos
		}
	}

	// 查询数据库
	var echos []model.Echo

	// 先按用户时区计算日界，再转为 UTC 查询数据库。
	loc := timezoneUtil.LoadLocationOrUTC(normalizedTimezone)
	nowUser := time.Now().UTC().In(loc)
	startOfDayUser := time.Date(nowUser.Year(), nowUser.Month(), nowUser.Day(), 0, 0, 0, 0, loc)
	endOfDayUser := startOfDayUser.Add(24 * time.Hour)
	startOfDayUTC := startOfDayUser.UTC()
	endOfDayUTC := endOfDayUser.UTC()

	query := echoRepository.db().Model(&model.Echo{})
	// 如果不是管理员，过滤私密Echo
	if !showPrivate {
		query = query.Where("private = ?", false)
	}

	// 添加当天的时间过滤
	query = query.Where("created_at >= ? AND created_at < ?", startOfDayUTC, endOfDayUTC)

	// 获取总数并进行分页查询
	query.
		Preload("Images").
		Preload("Tags").
		Order("created_at DESC").
		Find(&echos)

	// 保存到缓存，缓存到明天零点
	ttl := time.Until(endOfDayUser)
	if ttl <= 0 {
		ttl = time.Minute
	}
	cacheKey := GetTodayEchosCacheKey(showPrivate, normalizedTimezone)
	TrackTodayEchosCacheKey(cacheKey)
	echoRepository.cache.SetWithTTL(cacheKey, echos, 1, ttl)

	// 返回结果
	return echos
}

// UpdateEcho 更新 Echo
func (echoRepository *EchoRepository) UpdateEcho(ctx context.Context, echo *model.Echo) error {
	// 清空缓存
	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetEchoByIDCacheKey(echo.ID)) // 删除具体 Echo 的缓存
	ClearTodayEchosCache(echoRepository.cache)

	// 1. 先删除该 Echo 关联的所有旧图片
	if err := echoRepository.getDB(ctx).Where("message_id = ?", echo.ID).Delete(&model.Image{}).Error; err != nil {
		return err
	}

	// 2. 更新 Echo 内容（包括关联的新图片）
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

	// 3. 重新添加Images
	if len(echo.Images) > 0 {
		var images []model.Image
		for _, img := range echo.Images {
			// 确保每个图片都关联到正确的 Echo ID
			img.MessageID = echo.ID
			images = append(images, img)
		}
		// 批量插入新图片
		if err := echoRepository.getDB(ctx).Create(&images).Error; err != nil {
			return err
		}
	}

	// 4. 更新标签关联关系
	if err := echoRepository.getDB(ctx).Model(echo).Association("Tags").Replace(echo.Tags); err != nil {
		return err
	}

	return nil
}

// LikeEcho 点赞 Echo
func (echoRepository *EchoRepository) LikeEcho(ctx context.Context, id uint) error {
	// 检查是否存在（可选，防止无效点赞）
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

	// 原子自增点赞数
	if err := echoRepository.getDB(ctx).
		Model(&model.Echo{}).
		Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + ?", 1)).Error; err != nil {
		return err
	}

	// 清除相关缓存
	ClearEchoPageCache(echoRepository.cache)
	echoRepository.cache.Delete(GetEchoByIDCacheKey(id))      // 删除具体 Echo 的缓存
	ClearTodayEchosCache(echoRepository.cache)

	return nil
}

// GetAllTags 获取所有标签
func (echoRepository *EchoRepository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	result := echoRepository.db().Order("usage_count DESC, created_at DESC").Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

// DeleteTagById 删除标签
func (echoRepository *EchoRepository) DeleteTagById(ctx context.Context, id uint) error {
	var tag model.Tag

	// 删除关联的 EchoTag 关系
	if err := echoRepository.getDB(ctx).Where("tag_id = ?", id).Delete(&model.EchoTag{}).Error; err != nil {
		return err
	}

	// 删除标签
	result := echoRepository.getDB(ctx).Delete(&tag, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 如果没有找到记录
	}

	return nil
}

// GetTagByName 根据名称获取标签
func (echoRepository *EchoRepository) GetTagByName(name string) (*model.Tag, error) {
	var tag model.Tag
	result := echoRepository.db().Where("name = ?", name).First(&tag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 如果未找到记录，则返回 nil
		}
		return nil, result.Error // 其他错误返回
	}
	return &tag, nil
}

// GetTagsByNames 根据名称列表获取标签
func (echoRepository *EchoRepository) GetTagsByNames(names []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	result := echoRepository.db().Where("name IN ?", names).Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

// CreateTag 创建标签
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

// IncrementTagUsageCount 增加标签的使用计数
func (echoRepository *EchoRepository) IncrementTagUsageCount(
	ctx context.Context,
	tagID uint,
) error {
	return echoRepository.getDB(ctx).Model(&model.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

// GetEchosByTagId 根据标签ID获取关联的 Echo 列表
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
		Preload("Images").
		Preload("Tags").
		Order("created_at DESC").
		Find(&echos).Error; err != nil {
		return nil, 0, err
	}

	return echos, total, nil
}
