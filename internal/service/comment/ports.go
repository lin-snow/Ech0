package service

import (
	"context"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	model "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetFormMeta(ctx context.Context, clientIP string) (model.FormMeta, error)
	CreateComment(
		ctx context.Context,
		clientIP,
		userAgent string,
		dto *model.CreateCommentDto,
	) (model.CreateCommentResult, error)
	ListPublicByEchoID(ctx context.Context, echoID string) ([]model.Comment, error)
	ListPublicComments(ctx context.Context, limit int) ([]model.Comment, error)
	ListPanelComments(ctx context.Context, query model.ListCommentQuery) (model.PageResult[model.Comment], error)
	GetCommentByID(ctx context.Context, id string) (model.Comment, error)
	UpdateCommentStatus(ctx context.Context, id string, status model.Status) error
	UpdateCommentHot(ctx context.Context, id string, hot bool) error
	DeleteComment(ctx context.Context, id string) error
	BatchAction(ctx context.Context, action string, ids []string) error
	GetSystemSetting(ctx context.Context) (model.SystemSetting, error)
	UpdateSystemSetting(ctx context.Context, setting model.SystemSetting) error
}

type Repository interface {
	CreateComment(ctx context.Context, c *model.Comment) error
	ListPublicByEchoID(ctx context.Context, echoID string) ([]model.Comment, error)
	ListPublicComments(ctx context.Context, limit int) ([]model.Comment, error)
	ListComments(ctx context.Context, query model.ListCommentQuery) (model.PageResult[model.Comment], error)
	GetCommentByID(ctx context.Context, id string) (model.Comment, error)
	UpdateCommentStatus(ctx context.Context, id string, status model.Status) error
	UpdateCommentHot(ctx context.Context, id string, hot bool) error
	DeleteComment(ctx context.Context, id string) error
	BatchUpdateStatus(ctx context.Context, ids []string, status model.Status) error
	BatchDelete(ctx context.Context, ids []string) error
	CountByIPWithin(ctx context.Context, ipHash string, seconds int64) (int64, error)
	CountByEmailWithin(ctx context.Context, email string, seconds int64) (int64, error)
	CountByUserWithin(ctx context.Context, userID string, seconds int64) (int64, error)
	ExistsRecentDuplicate(
		ctx context.Context,
		echoID, content, email, ipHash, userID string,
		seconds int64,
	) (bool, error)
	GetEchoByID(ctx context.Context, id string) (echoModel.Echo, error)
}

type CommonService = commonService.Service

type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddKeyValue(ctx context.Context, key, value string) error
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
}

type UserContext struct {
	User  userModel.User
	Valid bool
}

type EventPublisher interface {
	CommentCreated(ctx context.Context, evt contracts.CommentCreatedEvent) error
	CommentStatusUpdated(ctx context.Context, evt contracts.CommentStatusUpdatedEvent) error
	CommentDeleted(ctx context.Context, evt contracts.CommentDeletedEvent) error
}
