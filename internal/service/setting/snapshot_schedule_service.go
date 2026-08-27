// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"

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
	v, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.Snapshot)
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

	if err := coreSetting.Set(ctx, settingService.durableKV, coreSetting.Snapshot, updated); err != nil {
		return err
	}

	eventbus.Notify(context.Background(), settingService.bus, event.UpdateSnapshotSchedule{Schedule: updated})
	return nil
}
