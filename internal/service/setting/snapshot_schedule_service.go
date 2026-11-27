// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
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
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	updated := model.SnapshotSchedule{
		Enable:         newSetting.Enable,
		CronExpression: newSetting.CronExpression,
	}

	if err := fmtUtil.ValidateCrontabExpression(updated.CronExpression); err != nil {
		return apperr.ErrInvalidCron
	}

	if err := coreSetting.Snapshot.Set(ctx, settingService.durableKV, updated); err != nil {
		return err
	}

	eventbus.Notify(context.Background(), settingService.bus, event.UpdateSnapshotSchedule{Schedule: updated})
	return nil
}
