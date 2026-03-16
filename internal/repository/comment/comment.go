package repository

import (
	"context"
	"strings"
	"time"

	model "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db func() *gorm.DB
}

var _ commentService.Repository = (*CommentRepository)(nil)

func NewCommentRepository(dbProvider func() *gorm.DB) *CommentRepository {
	return &CommentRepository{db: dbProvider}
}

func (r *CommentRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *CommentRepository) CreateComment(ctx context.Context, c *model.Comment) error {
	return r.getDB(ctx).Create(c).Error
}

func (r *CommentRepository) ListPublicByEchoID(ctx context.Context, echoID string) ([]model.Comment, error) {
	var out []model.Comment
	err := r.getDB(ctx).
		Where("echo_id = ? AND status = ?", echoID, model.StatusApproved).
		Order("created_at asc").
		Find(&out).Error
	return out, err
}

func (r *CommentRepository) ListPublicComments(ctx context.Context, limit int) ([]model.Comment, error) {
	var out []model.Comment
	err := r.getDB(ctx).
		Where("status = ?", model.StatusApproved).
		Order("created_at desc").
		Limit(limit).
		Find(&out).Error
	return out, err
}

func (r *CommentRepository) ListComments(
	ctx context.Context,
	query model.ListCommentQuery,
) (model.PageResult[model.Comment], error) {
	db := r.getDB(ctx).Model(&model.Comment{})
	if query.EchoID != "" {
		db = db.Where("echo_id = ?", query.EchoID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Hot != nil {
		db = db.Where("hot = ?", *query.Hot)
	}
	if strings.TrimSpace(query.Keyword) != "" {
		kw := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where(
			"(nickname LIKE ? OR email LIKE ? OR content LIKE ?)",
			kw,
			kw,
			kw,
		)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return model.PageResult[model.Comment]{}, err
	}

	var items []model.Comment
	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at desc").
		Offset(offset).
		Limit(query.PageSize).
		Find(&items).Error
	if err != nil {
		return model.PageResult[model.Comment]{}, err
	}

	return model.PageResult[model.Comment]{
		Items: items,
		Total: total,
	}, nil
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, id string) (model.Comment, error) {
	var item model.Comment
	err := r.getDB(ctx).Where("id = ?", id).First(&item).Error
	return item, err
}

func (r *CommentRepository) UpdateCommentStatus(
	ctx context.Context,
	id string,
	status model.Status,
) error {
	return r.getDB(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *CommentRepository) UpdateCommentHot(
	ctx context.Context,
	id string,
	hot bool,
) error {
	return r.getDB(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Update("hot", hot).Error
}

func (r *CommentRepository) DeleteComment(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&model.Comment{}).Error
}

func (r *CommentRepository) BatchUpdateStatus(
	ctx context.Context,
	ids []string,
	status model.Status,
) error {
	if len(ids) == 0 {
		return nil
	}
	return r.getDB(ctx).
		Model(&model.Comment{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

func (r *CommentRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.getDB(ctx).Where("id IN ?", ids).Delete(&model.Comment{}).Error
}

func (r *CommentRepository) CountByIPWithin(ctx context.Context, ipHash string, seconds int64) (int64, error) {
	return r.countByFieldWithin(ctx, "ip_hash", ipHash, seconds)
}

func (r *CommentRepository) CountByEmailWithin(ctx context.Context, email string, seconds int64) (int64, error) {
	return r.countByFieldWithin(ctx, "email", email, seconds)
}

func (r *CommentRepository) CountByUserWithin(ctx context.Context, userID string, seconds int64) (int64, error) {
	return r.countByFieldWithin(ctx, "user_id", userID, seconds)
}

func (r *CommentRepository) countByFieldWithin(
	ctx context.Context,
	field string,
	value string,
	seconds int64,
) (int64, error) {
	var count int64
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	since := time.Now().Add(-time.Duration(seconds) * time.Second)
	err := r.getDB(ctx).
		Model(&model.Comment{}).
		Where(field+" = ? AND created_at >= ?", value, since).
		Count(&count).Error
	return count, err
}

func (r *CommentRepository) ExistsRecentDuplicate(
	ctx context.Context,
	echoID, content, email, ipHash, userID string,
	seconds int64,
) (bool, error) {
	since := time.Now().Add(-time.Duration(seconds) * time.Second)
	db := r.getDB(ctx).Model(&model.Comment{}).
		Where(
			"echo_id = ? AND content = ? AND created_at >= ?",
			strings.TrimSpace(echoID),
			strings.TrimSpace(content),
			since,
		)

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID != "" {
		db = db.Where("user_id = ?", trimmedUserID)
	} else {
		trimmedEmail := strings.TrimSpace(email)
		trimmedIPHash := strings.TrimSpace(ipHash)
		if trimmedEmail != "" {
			db = db.Where("email = ?", trimmedEmail)
		} else if trimmedIPHash != "" {
			db = db.Where("ip_hash = ?", trimmedIPHash)
		}
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CommentRepository) GetEchoByID(ctx context.Context, id string) (echoModel.Echo, error) {
	var echo echoModel.Echo
	err := r.getDB(ctx).Where("id = ?", id).First(&echo).Error
	return echo, err
}
