// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"testing"
	"time"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedComment(t *testing.T, db *gorm.DB, id, echoID string, status commentModel.Status) {
	t.Helper()
	c := commentModel.Comment{
		ID:       id,
		EchoID:   echoID,
		Nickname: "anon",
		Email:    "a@example.com",
		Content:  "hi",
		Status:   status,
		Source:   commentModel.SourceGuest,
	}
	require.NoError(t, db.Create(&c).Error)
}

func TestEchoRepository_GetHotEchos(t *testing.T) {
	t.Run("ranks by fav_count plus approved-comment weight", func(t *testing.T) {
		repo, db := newEchoRepo(t)
		seedEcho(t, db, "e-cold", "cold", false, 0, 100)
		seedEcho(t, db, "e-fav", "favored", false, 5, 200)
		seedEcho(t, db, "e-cmt", "discussed", false, 0, 300)
		seedComment(t, db, "c1", "e-cmt", commentModel.StatusApproved)
		seedComment(t, db, "c2", "e-cmt", commentModel.StatusApproved)
		seedComment(t, db, "c3", "e-cmt", commentModel.StatusPending)

		echos, err := repo.GetHotEchos(5, true)
		require.NoError(t, err)
		require.Len(t, echos, 3)
		assert.Equal(t, []string{"e-fav", "e-cmt", "e-cold"}, echoIDs(echos))
	})

	t.Run("non-positive limit defaults to 5", func(t *testing.T) {
		repo, db := newEchoRepo(t)
		for i := 1; i <= 6; i++ {
			seedEcho(t, db, "e"+string(rune('0'+i)), "c", false, i, int64(i*100))
		}

		echos, err := repo.GetHotEchos(0, true)
		require.NoError(t, err)
		require.Len(t, echos, 5)
		assert.NotContains(t, echoIDs(echos), "e1", "fav 最低的一条应被默认 limit=5 裁掉")
	})

	t.Run("excessive limit does not error", func(t *testing.T) {
		repo, db := newEchoRepo(t)
		seedEcho(t, db, "e1", "a", false, 1, 100)
		seedEcho(t, db, "e2", "b", false, 2, 200)

		echos, err := repo.GetHotEchos(1000, true)
		require.NoError(t, err)
		assert.Len(t, echos, 2)
	})

	t.Run("private filter hides private echos when showPrivate=false", func(t *testing.T) {
		repo, db := newEchoRepo(t)
		seedEcho(t, db, "e-pub", "public", false, 1, 100)
		seedEcho(t, db, "e-prv", "secret", true, 99, 200)

		hidden, err := repo.GetHotEchos(5, false)
		require.NoError(t, err)
		assert.Equal(t, []string{"e-pub"}, echoIDs(hidden))

		shown, err := repo.GetHotEchos(5, true)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"e-pub", "e-prv"}, echoIDs(shown))
	})

	t.Run("empty database returns empty slice", func(t *testing.T) {
		repo, _ := newEchoRepo(t)
		echos, err := repo.GetHotEchos(5, true)
		require.NoError(t, err)
		assert.Empty(t, echos)
	})
}

func TestEchoRepository_GetEchosByTagId(t *testing.T) {
	setup := func(t *testing.T) (*EchoRepository, *gorm.DB) {
		t.Helper()
		repo, db := newEchoRepo(t)
		seedEcho(t, db, "e-pub", "public golang post", false, 0, 100)
		seedEcho(t, db, "e-prv", "private golang post", true, 0, 200)
		seedEcho(t, db, "e-other", "untagged", false, 0, 300)
		seedTag(t, db, "t1", "golang")
		linkTag(t, db, "e-pub", "t1")
		linkTag(t, db, "e-prv", "t1")
		return repo, db
	}

	t.Run("showPrivate=false excludes private", func(t *testing.T) {
		repo, _ := setup(t)
		echos, total, err := repo.GetEchosByTagId("t1", 1, 10, "", false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-pub", echos[0].ID)
	})

	t.Run("showPrivate=true includes private, ordered created_at DESC", func(t *testing.T) {
		repo, _ := setup(t)
		echos, total, err := repo.GetEchosByTagId("t1", 1, 10, "", true)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, []string{"e-prv", "e-pub"}, echoIDs(echos))
	})

	t.Run("search narrows within the tag", func(t *testing.T) {
		repo, _ := setup(t)
		echos, total, err := repo.GetEchosByTagId("t1", 1, 10, "public golang", true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-pub", echos[0].ID)
	})

	t.Run("unknown tag short-circuits to empty", func(t *testing.T) {
		repo, _ := setup(t)
		echos, total, err := repo.GetEchosByTagId("ghost-tag", 1, 10, "", true)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		require.NotNil(t, echos)
		assert.Empty(t, echos)
	})

	t.Run("offset past results returns empty with real total", func(t *testing.T) {
		repo, _ := setup(t)
		echos, total, err := repo.GetEchosByTagId("t1", 2, 10, "", true)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Empty(t, echos)
	})
}

func TestEchoRepository_GetTodayEchos(t *testing.T) {
	repo, db := newEchoRepo(t)

	loc := time.UTC
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayTs := startOfDay.Add(time.Second).Unix()
	yesterdayTs := startOfDay.Add(-time.Second).Unix()

	seedEcho(t, db, "e-today-pub", "today public", false, 0, todayTs)
	seedEcho(t, db, "e-today-prv", "today private", true, 0, todayTs)
	seedEcho(t, db, "e-yesterday", "yesterday public", false, 0, yesterdayTs)

	t.Run("showPrivate=false returns only today's public echos", func(t *testing.T) {
		echos := repo.GetTodayEchos(false, "UTC")
		assert.Equal(t, []string{"e-today-pub"}, echoIDs(echos))
	})

	t.Run("showPrivate=true includes today's private echos", func(t *testing.T) {
		echos := repo.GetTodayEchos(true, "UTC")
		assert.ElementsMatch(t, []string{"e-today-pub", "e-today-prv"}, echoIDs(echos))
	})
}

func TestEchoRepository_GetOnThisDayEchos(t *testing.T) {
	loc := time.UTC
	now := time.Now().In(loc)
	if now.Month() == time.February && now.Day() == 29 {
		t.Skip("Feb 29: 去年同月日不存在，跳过该边界（onthisday_test.go 已专测该逻辑）")
	}

	repo, db := newEchoRepo(t)
	year := now.Year()
	lastYearSame := time.Date(year-1, now.Month(), now.Day(), 12, 0, 0, 0, loc).Unix()
	thisYearSame := time.Date(year, now.Month(), now.Day(), 12, 0, 0, 0, loc).Unix()
	lastYearOther := time.Date(year-1, now.Month(), now.Day(), 12, 0, 0, 0, loc).AddDate(0, 0, 1).Unix()

	seedEcho(t, db, "e-match", "那年今日", false, 0, lastYearSame)
	seedEcho(t, db, "e-this-year", "今年今日（应排除）", false, 0, thisYearSame)
	seedEcho(t, db, "e-other-day", "去年隔天（应排除）", false, 0, lastYearOther)
	seedEcho(t, db, "e-match-prv", "那年今日私密", true, 0, lastYearSame+1)

	t.Run("showPrivate=false matches only past-year same month-day public echos", func(t *testing.T) {
		echos := repo.GetOnThisDayEchos(false, "UTC")
		assert.Equal(t, []string{"e-match"}, echoIDs(echos))
	})

	t.Run("showPrivate=true also includes private", func(t *testing.T) {
		echos := repo.GetOnThisDayEchos(true, "UTC")
		assert.ElementsMatch(t, []string{"e-match", "e-match-prv"}, echoIDs(echos))
	})

	t.Run("empty database returns empty slice", func(t *testing.T) {
		emptyRepo, _ := newEchoRepo(t)
		echos := emptyRepo.GetOnThisDayEchos(true, "UTC")
		assert.Empty(t, echos)
	})
}
