package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	githubUtil "github.com/lin-snow/ech0/internal/util/github"
	"golang.org/x/mod/semver"
	"gorm.io/gorm"
)

type InboxDispatcher struct {
	inboxRepo    inboxRepository.InboxRepositoryInterface
	keyvalueRepo keyvalueRepository.KeyValueRepositoryInterface
}

func NewInboxDispatcher(inboxRepo inboxRepository.InboxRepositoryInterface, keyvalueRepo keyvalueRepository.KeyValueRepositoryInterface) *InboxDispatcher {
	return &InboxDispatcher{inboxRepo: inboxRepo, keyvalueRepo: keyvalueRepo}
}

func (id *InboxDispatcher) Handle(ctx context.Context, e *Event) error {
	switch e.Type {
	case EventTypeEch0UpdateCheck:
		return id.handleEch0UpdateCheck(ctx)
	case EventTypeInboxClear:
		return id.handleInboxClear(ctx)
	}

	return nil
}

func (id *InboxDispatcher) handleEch0UpdateCheck(ctx context.Context) error {
	// 检查 Ech0 版本更新
	currentVersion := commonModel.Version

	// 获取最新版本
	latestVersion, err := githubUtil.GetLatestVersion()
	if err != nil {
		return err
	}

	// 检查是否已发送过更新通知
	releaseVersion, err := id.keyvalueRepo.GetKeyValue(commonModel.ReleaseVersionKey)
	if err != nil {
		// 首次运行时该 key 可能不存在，视为“未通知过”
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	releaseVersionStr, _ := releaseVersion.(string)
	if releaseVersionStr != "" && releaseVersionStr == latestVersion {
		return nil
	}

	// 语义化版本比较
	cur := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(currentVersion), "v"))
	lat := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(latestVersion), "v"))
	// 任意一方不是合法 semver，则跳过比较（避免误报/漏报）
	if cur != "" && lat != "" && semver.Compare(lat, cur) > 0 {
		// 有新版本，发送更新通知
		meta, _ := json.Marshal(map[string]string{
			"latest_version":  latestVersion,
			"current_version": currentVersion,
		})
		err = id.inboxRepo.PostInbox(ctx, &inboxModel.Inbox{
			Source:    string(commonModel.SystemSource),
			Content:   fmt.Sprintf("有新版本可用，请更新：%s", latestVersion),
			Type:      string(commonModel.NotificationInboxType),
			Read:      false,
			ReadCount: 0,
			ReadAt:    0, // 首次发送时未读
			Meta:      string(meta),
			CreatedAt: time.Now().UTC().Unix(),
		})
		if err != nil {
			return err
		}
	}

	// 更新存储键 ReleaseVersionKey 值
	err = id.keyvalueRepo.AddOrUpdateKeyValue(ctx, commonModel.ReleaseVersionKey, latestVersion)
	if err != nil {
		return err
	}

	return nil
}

func (id *InboxDispatcher) handleInboxClear(ctx context.Context) error {
	// 获取所有未读消息
	unreadInbox, err := id.inboxRepo.GetUnreadInbox(ctx)
	if err != nil {
		return err
	}

	// 清理已读的存在超过五天的消息
	var unreadInboxIDs []uint
	for _, inbox := range unreadInbox {
		// 如果消息已读并且创建时间超过七天，则清理
		if inbox.Read && inbox.ReadCount > 2 && time.Now().UTC().Unix()-inbox.ReadAt > 5*24*60*60 {
			unreadInboxIDs = append(unreadInboxIDs, inbox.ID)
		}
	}

	// 清理掉
	if err := id.inboxRepo.ClearReadInboxByIds(ctx, unreadInboxIDs); err != nil {
		return err
	}

	return nil
}
