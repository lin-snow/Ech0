// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	echomock "github.com/lin-snow/ech0/internal/test/mocks/echomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func tagNames(tags []echoModel.Tag) []string {
	t := make([]string, 0, len(tags))
	for _, tag := range tags {
		t = append(t, tag.Name)
	}
	return t
}

func TestProcessEchoTags(t *testing.T) {
	t.Run("trims whitespace strips hash and routes existing vs new", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)

		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"golang", "vue"}).
			Return([]*echoModel.Tag{{ID: "tag-go", Name: "golang", UsageCount: 3}}, nil).
			Once()
		repo.EXPECT().IncrementTagUsageCount(mock.Anything, "tag-go").Return(nil).Once()

		var created echoModel.Tag
		repo.EXPECT().
			CreateTag(mock.Anything, mock.Anything).
			Run(func(_ context.Context, tag *echoModel.Tag) { created = *tag }).
			Return(nil).
			Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "#golang"}, {Name: "  vue  "}}}

		require.NoError(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo))

		assert.Equal(t, "vue", created.Name)
		assert.Equal(t, 1, created.UsageCount)
		assert.Equal(t, []string{"golang", "vue"}, tagNames(echo.Tags))
	})

	t.Run("skips empty and hash-only names", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)

		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"kept"}).
			Return([]*echoModel.Tag{}, nil).
			Once()
		repo.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(nil).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "#"}, {Name: "   "}, {Name: "kept"}}}

		require.NoError(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo))
		assert.Equal(t, []string{"kept"}, tagNames(echo.Tags))
	})

	t.Run("rejects unsafe tag name before touching repo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "<script>"}}}

		err := svc.ProcessEchoTags(helpers.CtxAnonymous(), echo)
		require.EqualError(t, err, commonModel.INVALID_PARAMS)
	})

	t.Run("all existing increments only", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"a", "b"}).
			Return([]*echoModel.Tag{{ID: "ta", Name: "a"}, {ID: "tb", Name: "b"}}, nil).
			Once()
		repo.EXPECT().IncrementTagUsageCount(mock.Anything, "ta").Return(nil).Once()
		repo.EXPECT().IncrementTagUsageCount(mock.Anything, "tb").Return(nil).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "a"}, {Name: "b"}}}

		require.NoError(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo))
		assert.Equal(t, []string{"a", "b"}, tagNames(echo.Tags))
	})

	t.Run("GetTagsByNames error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		boom := errors.New("db down")
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return(nil, boom).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "x"}}}

		require.ErrorIs(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo), boom)
	})

	t.Run("IncrementTagUsageCount error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		boom := errors.New("increment failed")
		repo.EXPECT().
			GetTagsByNames(mock.Anything, mock.Anything).
			Return([]*echoModel.Tag{{ID: "ta", Name: "a"}}, nil).
			Once()
		repo.EXPECT().IncrementTagUsageCount(mock.Anything, "ta").Return(boom).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "a"}}}

		require.ErrorIs(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo), boom)
	})

	t.Run("CreateTag error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		boom := errors.New("create failed")
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
		repo.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(boom).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{Tags: []echoModel.Tag{{Name: "fresh"}}}

		require.ErrorIs(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo), boom)
	})

	t.Run("no tags is a no-op create-wise", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		echo := &echoModel.Echo{}

		require.NoError(t, svc.ProcessEchoTags(helpers.CtxAnonymous(), echo))
		assert.Empty(t, echo.Tags)
	})
}
