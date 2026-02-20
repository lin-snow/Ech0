package repository

import (
	"context"
	"errors"
	"time"

	model "github.com/lin-snow/ech0/internal/model/fediverse"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type FediverseRepository struct {
	db func() *gorm.DB
}

func NewFediverseRepository(dbProvider func() *gorm.DB) FediverseRepositoryInterface {
	return &FediverseRepository{
		db: dbProvider,
	}
}

func (r *FediverseRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return r.db()
}

func (r *FediverseRepository) GetFollowers(userID uint) ([]model.Follower, error) {
	var followers []model.Follower
	if err := r.db().Where("user_id = ?", userID).Find(&followers).Error; err != nil {
		return nil, err
	}
	return followers, nil
}

func (r *FediverseRepository) GetFollowing(userID uint) ([]model.Follow, error) {
	var following []model.Follow
	if err := r.db().Where("user_id = ?", userID).Find(&following).Error; err != nil {
		return nil, err
	}
	return following, nil
}

func (r *FediverseRepository) SaveFollower(ctx context.Context, follower *model.Follower) error {
	return r.getDB(ctx).Create(follower).Error
}

func (r *FediverseRepository) FollowerExists(
	ctx context.Context,
	userID uint,
	actor string,
) (bool, error) {
	var count int64
	if err := r.getDB(ctx).Model(&model.Follower{}).Where("user_id = ? AND actor_id = ?", userID, actor).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FediverseRepository) SaveOrUpdateFollow(ctx context.Context, follow *model.Follow) error {
	if follow == nil {
		return errors.New("follow is nil")
	}

	db := r.getDB(ctx)
	var existing model.Follow
	err := db.Where("user_id = ? AND object_id = ?", follow.UserID, follow.ObjectID).
		First(&existing).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(follow).Error
	}
	if err != nil {
		return err
	}

	existing.ActorID = follow.ActorID
	existing.ActivityID = follow.ActivityID
	existing.Status = follow.Status
	return db.Save(&existing).Error
}

func (r *FediverseRepository) GetFollowByUserAndObject(
	ctx context.Context,
	userID uint,
	objectID string,
) (*model.Follow, error) {
	var follow model.Follow
	err := r.getDB(ctx).
		Where("user_id = ? AND object_id = ?", userID, objectID).
		First(&follow).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *FediverseRepository) DeleteFollow(ctx context.Context, followID uint) error {
	return r.getDB(ctx).Delete(&model.Follow{}, followID).Error
}

func (r *FediverseRepository) UpsertInboxStatus(
	ctx context.Context,
	status *model.InboxStatus,
) error {
	if status == nil {
		return errors.New("inbox status is nil")
	}

	if status.ActivityID == "" {
		return errors.New("activity id is empty")
	}

	db := r.getDB(ctx)

	var existing model.InboxStatus
	err := db.Where("user_id = ? AND activity_id = ?", status.UserID, status.ActivityID).
		First(&existing).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if status.CreatedAt.IsZero() {
			status.CreatedAt = time.Now().UTC()
		}
		if status.UpdatedAt.IsZero() {
			status.UpdatedAt = status.CreatedAt
		}
		return db.Create(status).Error
	}
	if err != nil {
		return err
	}

	status.ID = existing.ID
	if status.UpdatedAt.IsZero() {
		status.UpdatedAt = time.Now().UTC()
	}

	updates := map[string]any{
		"actor_id":                 status.ActorID,
		"actor_preferred_username": status.ActorPreferredUsername,
		"actor_display_name":       status.ActorDisplayName,
		"actor_avatar":             status.ActorAvatar,
		"object_id":                status.ObjectID,
		"object_type":              status.ObjectType,
		"object_attributed_to":     status.ObjectAttributedTo,
		"summary":                  status.Summary,
		"content":                  status.Content,
		"to":                       status.To,
		"cc":                       status.Cc,
		"raw_activity":             status.RawActivity,
		"raw_object":               status.RawObject,
		"published_at":             status.PublishedAt,
		"updated_at":               status.UpdatedAt,
	}

	return db.Model(&existing).Updates(updates).Error
}

func (r *FediverseRepository) ListInboxStatuses(
	ctx context.Context,
	userID uint,
	page, pageSize int,
) ([]model.InboxStatus, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = int(model.DefaultCollectionPageSize)
	}

	offset := (page - 1) * pageSize

	query := r.getDB(ctx).Model(&model.InboxStatus{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	statuses := make([]model.InboxStatus, 0, pageSize)
	if total == 0 {
		return statuses, 0, nil
	}

	if err := query.Order("published_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&statuses).Error; err != nil {
		return nil, 0, err
	}

	return statuses, total, nil
}

func (r *FediverseRepository) UpdateFollowStatusByActivityID(
	ctx context.Context,
	userID uint,
	activityID, status string,
) error {
	if activityID == "" {
		return errors.New("activity id is empty")
	}
	if status == "" {
		return errors.New("status is empty")
	}

	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}

	result := r.getDB(ctx).
		Model(&model.Follow{}).
		Where("user_id = ? AND activity_id = ?", userID, activityID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
