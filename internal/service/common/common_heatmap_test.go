// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"
	"time"
	_ "time/tzdata"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const secondsPerDay = int64(24 * 60 * 60)

func countForDate(t *testing.T, hm []commonModel.Heatmap, date string) (int, bool) {
	t.Helper()
	for _, h := range hm {
		if h.Date == date {
			return h.Count, true
		}
	}
	return 0, false
}

func sumCounts(hm []commonModel.Heatmap) int {
	total := 0
	for _, h := range hm {
		total += h.Count
	}
	return total
}

func TestGetHeatMap_BucketingStructure(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, nil)

	loc := time.UTC
	now := time.Now().In(loc)
	todayNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)

	timestamps := []int64{
		todayNoon.Unix(),
		todayNoon.Unix(),
		todayNoon.Unix(),
		todayNoon.AddDate(0, 0, -1).Unix(),
		todayNoon.AddDate(0, 0, -1).Unix(),
		todayNoon.AddDate(0, 0, -29).Unix(),
	}

	var gotStart, gotEnd int64
	repo.EXPECT().
		GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, start int64, end int64) {
			gotStart, gotEnd = start, end
		}).
		Return(timestamps, nil).
		Once()

	hm, err := svc.GetHeatMap("UTC")
	require.NoError(t, err)
	require.Len(t, hm, 30, "热力图固定返回 30 天")

	assert.Equal(t, 30*secondsPerDay, gotEnd-gotStart, "查询窗口宽度应为 30 天")
	assert.Zero(t, gotStart%secondsPerDay, "UTC 起点应对齐午夜")

	assert.Equal(t, 3, hm[29].Count, "今天应有 3 条")
	assert.Equal(t, 2, hm[28].Count, "昨天应有 2 条")
	assert.Equal(t, 1, hm[0].Count, "29 天前应有 1 条")
	assert.Equal(t, 6, sumCounts(hm), "落在窗口内的计数总和")

	for i := 0; i < len(hm)-1; i++ {
		prev, perr := time.Parse("2006-01-02", hm[i].Date)
		require.NoError(t, perr)
		assert.Equal(t, prev.AddDate(0, 0, 1).Format("2006-01-02"), hm[i+1].Date, "相邻格子应相差一天")
	}
}

func TestGetHeatMap_CrossTimezoneBucketing(t *testing.T) {
	shLoc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)

	nowUTC := time.Now().UTC()
	base := nowUTC.AddDate(0, 0, -5)
	instant := time.Date(base.Year(), base.Month(), base.Day(), 23, 0, 0, 0, time.UTC)

	utcDate := instant.Format("2006-01-02")
	shDate := instant.In(shLoc).Format("2006-01-02")
	require.NotEqual(t, utcDate, shDate, "构造的时刻应跨日，否则用例无意义")

	cases := []struct {
		name      string
		timezone  string
		hitDate   string
		emptyDate string
	}{
		{name: "UTC 归当天", timezone: "UTC", hitDate: utcDate, emptyDate: shDate},
		{name: "Shanghai 归次日", timezone: "Asia/Shanghai", hitDate: shDate, emptyDate: utcDate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := commonmock.NewMockCommonRepository(t)
			svc := commonService.NewCommonService(repo, nil)

			repo.EXPECT().
				GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
				Return([]int64{instant.Unix()}, nil).
				Once()

			hm, hmErr := svc.GetHeatMap(tc.timezone)
			require.NoError(t, hmErr)
			require.Len(t, hm, 30)

			hit, ok := countForDate(t, hm, tc.hitDate)
			require.True(t, ok, "命中日期 %s 应在窗口内", tc.hitDate)
			assert.Equal(t, 1, hit, "该时间戳应只落在 %s 这一格", tc.hitDate)

			empty, ok := countForDate(t, hm, tc.emptyDate)
			require.True(t, ok, "对照日期 %s 应在窗口内", tc.emptyDate)
			assert.Equal(t, 0, empty, "另一时区的日期 %s 不应计数", tc.emptyDate)

			assert.Equal(t, 1, sumCounts(hm), "整个窗口只有一条记录")
		})
	}
}

func TestGetHeatMap_RepositoryError(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, nil)

	wantErr := assert.AnError
	repo.EXPECT().
		GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, wantErr).
		Once()

	hm, err := svc.GetHeatMap("UTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.Nil(t, hm)
}

func TestGetHeatMap_InvalidTimezoneFallsBackToUTC(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, nil)

	var gotStart int64
	repo.EXPECT().
		GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, start int64, _ int64) { gotStart = start }).
		Return([]int64{}, nil).
		Once()

	hm, err := svc.GetHeatMap("Not/A_Zone")
	require.NoError(t, err)
	require.Len(t, hm, 30)
	assert.Zero(t, gotStart%secondsPerDay)
	assert.Zero(t, sumCounts(hm))
}
