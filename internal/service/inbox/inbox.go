package service

import (
	"context"
	"errors"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type InboxService struct {
	txManager       transaction.TransactionManager
	commonService   commonService.CommonServiceInterface
	inboxRepository inboxRepository.InboxRepositoryInterface
}

func NewInboxService(
	tm transaction.TransactionManager,
	commonSvc commonService.CommonServiceInterface,
	inboxRepo inboxRepository.InboxRepositoryInterface,
) InboxServiceInterface {
	return &InboxService{
		txManager:       tm,
		commonService:   commonSvc,
		inboxRepository: inboxRepo,
	}
}

// GetInboxList 获取收件箱消息列表
func (inboxService *InboxService) GetInboxList(
	userid uint,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]*inboxModel.Inbox], error) {
	if err := inboxService.ensureAdmin(userid); err != nil {
		return commonModel.PageQueryResult[[]*inboxModel.Inbox]{}, err
	}

	if pageQueryDto.Page < 1 {
		pageQueryDto.Page = 1
	}
	if pageQueryDto.PageSize < 1 || pageQueryDto.PageSize > 100 {
		pageQueryDto.PageSize = 10
	}
	pageQueryDto.Search = strings.TrimSpace(pageQueryDto.Search)

	offset := (pageQueryDto.Page - 1) * pageQueryDto.PageSize

	inboxes, total, err := inboxService.inboxRepository.GetInboxList(
		context.Background(),
		offset,
		pageQueryDto.PageSize,
		pageQueryDto.Search,
	)
	if err != nil {
		return commonModel.PageQueryResult[[]*inboxModel.Inbox]{}, err
	}

	return commonModel.PageQueryResult[[]*inboxModel.Inbox]{
		Items: inboxes,
		Total: total,
	}, nil
}

// GetUnreadInbox 获取所有未读消息
func (inboxService *InboxService) GetUnreadInbox(userid uint) ([]*inboxModel.Inbox, error) {
	if err := inboxService.ensureAdmin(userid); err != nil {
		return nil, err
	}

	return inboxService.inboxRepository.GetUnreadInbox(context.Background())
}

// MarkAsRead 将消息标记为已读
func (inboxService *InboxService) MarkAsRead(userid, inboxID uint) error {
	if err := inboxService.ensureAdmin(userid); err != nil {
		return err
	}

	// 获取消息
	inbox, err := inboxService.inboxRepository.GetInboxById(context.Background(), inboxID)
	if err != nil {
		return inboxService.handleRepoError(err)
	}

	// 如果消息未读，则增加已读次数和已读时间
	if !inbox.Read {
		inbox.ReadCount++
		inbox.ReadAt = time.Now().UTC().Unix()
	} else {
		// 如果消息已读，则增加已读次数
		inbox.ReadCount++
	}

	return inboxService.txManager.Run(func(ctx context.Context) error {
		if err := inboxService.inboxRepository.UpdateInbox(ctx, inbox); err != nil {
			return inboxService.handleRepoError(err)
		}
		return nil
	})
}

// DeleteInbox 删除指定的收件箱消息
func (inboxService *InboxService) DeleteInbox(userid, inboxID uint) error {
	if err := inboxService.ensureAdmin(userid); err != nil {
		return err
	}

	return inboxService.txManager.Run(func(ctx context.Context) error {
		if err := inboxService.inboxRepository.DeleteInbox(ctx, inboxID); err != nil {
			return inboxService.handleRepoError(err)
		}
		return nil
	})
}

// ClearInbox 清空收件箱
func (inboxService *InboxService) ClearInbox(userid uint) error {
	if err := inboxService.ensureAdmin(userid); err != nil {
		return err
	}

	return inboxService.txManager.Run(func(ctx context.Context) error {
		return inboxService.inboxRepository.ClearInbox(ctx)
	})
}

func (inboxService *InboxService) ensureAdmin(userid uint) error {
	user, err := inboxService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return nil
}

func (inboxService *InboxService) handleRepoError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New(commonModel.INBOX_NOT_FOUND)
	}
	return err
}
