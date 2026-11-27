// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTagUsageService(t *testing.T, confirmTempFiles bool) (*echoService.EchoService, *echoRepository.EchoRepository, *gorm.DB) {
	t.Helper()

	db := helpers.NewTestDB(t)
	repo := echoRepository.NewEchoRepository(func() *gorm.DB { return db }, helpers.NewTestCache())
	tx := transaction.NewGormTransactor(func() *gorm.DB { return db })
	bus := helpers.NewTestBus(t)

	common := commonmock.NewMockService(t)
	common.EXPECT().CommonGetUserByUserId(mock.Anything, "user-test-0001").Return(helpers.NewUser(), nil).Once()
	file := filemock.NewMockService(t)
	if confirmTempFiles {
		file.EXPECT().ConfirmTempFiles(mock.Anything, mock.Anything).Return(nil).Once()
	}

	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	return svc, repo, db
}

func seedTaggedEcho(t *testing.T, svc *echoService.EchoService, repo *echoRepository.EchoRepository) *echoModel.Echo {
	t.Helper()

	ctx := helpers.CtxAsUser("user-test-0001")
	echo := &echoModel.Echo{
		ID:      "echo-1",
		Content: "first",
		UserID:  "user-test-0001",
		Layout:  "waterfall",
		Tags:    []echoModel.Tag{{Name: "foo"}},
	}
	require.NoError(t, svc.ProcessEchoTags(ctx, echo))
	require.NoError(t, repo.CreateEcho(ctx, echo))
	return echo
}

// An edit used to increment every tag again, so the count grew without any
// new echo using the tag.
func TestUpdateEchoKeepsTagUsageCount(t *testing.T) {
	svc, repo, db := newTagUsageService(t, true)
	seedTaggedEcho(t, svc, repo)

	_, err := svc.UpdateEcho(helpers.CtxAsUser("user-test-0001"), &echoModel.Echo{
		ID:      "echo-1",
		Content: "edited",
		Layout:  "waterfall",
		Tags:    []echoModel.Tag{{Name: "foo"}},
	})
	require.NoError(t, err)

	var tag echoModel.Tag
	require.NoError(t, db.Where("name = ?", "foo").First(&tag).Error)
	require.Equal(t, 1, tag.UsageCount)
}

// Deleting the only echo with a tag used to leave the count and the join row
// behind.
func TestDeleteEchoClearsTagUsage(t *testing.T) {
	svc, repo, db := newTagUsageService(t, false)
	seedTaggedEcho(t, svc, repo)

	require.NoError(t, svc.DeleteEchoById(helpers.CtxAsUser("user-test-0001"), "echo-1"))

	var tag echoModel.Tag
	require.NoError(t, db.Where("name = ?", "foo").First(&tag).Error)
	require.Zero(t, tag.UsageCount)

	var links int64
	require.NoError(t, db.Model(&echoModel.EchoTag{}).Where("tag_id = ?", tag.ID).Count(&links).Error)
	require.Zero(t, links)
}
