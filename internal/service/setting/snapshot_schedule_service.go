// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"github.com/lin-snow/ech0/internal/migrator/snapshot"
	"github.com/lin-snow/ech0/internal/serverless/platform"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"time"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetSnapshotScheduleSetting(
	setting *model.SnapshotSchedule,
) error {
	v, err := coreSetting.Snapshot.Get(context.Background(), settingService.durableKV)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

func (settingService *SettingService) UpdateSnapshotScheduleSetting(
	ctx context.Context,
	newSetting *model.SnapshotScheduleDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	updated := model.SnapshotSchedule{
		Enable:         newSetting.Enable,
		CronExpression: newSetting.CronExpression,
	}

	if err := fmtUtil.ValidateCrontabExpression(updated.CronExpression); err != nil {
		return errors.New(commonModel.INVALID_CRON_EXPRESSION)
	}

	if err := coreSetting.Snapshot.Set(ctx, settingService.durableKV, updated); err != nil {
		return err
	}

	if platform.IsFunction(ctx) {
		next, err := snapshot.NextOccurrence(updated, time.Now())
		if err != nil {
			return err
		}
		if next != nil {
			if err := platform.Publish(ctx, next, next.ScheduleKey(), max(int64(0), next.SnapshotDue-time.Now().Unix())); err != nil {
				logUtil.GetLogger().Warn("snapshot schedule saved for hourly reconciliation")
			}
		}
	}
	eventbus.Notify(context.Background(), settingService.bus, event.UpdateSnapshotSchedule{Schedule: updated})
	return nil
}
