// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"errors"
	"testing"
	_ "time/tzdata"

	handler "github.com/lin-snow/ech0/internal/handler/echo"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/echomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func newPage(items []echoModel.Echo) commonModel.PageQueryResult[[]echoModel.Echo] {
	return commonModel.PageQueryResult[[]echoModel.Echo]{
		Items: items,
		Total: int64(len(items)),
	}
}

func TestEchoHandler_PostEcho(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().
			PostEcho(mock.Anything, mock.MatchedBy(func(e *echoModel.Echo) bool {
				return e != nil && e.Content == "hi" && e.Private
			})).
			Return(nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.PostEcho(helpers.CtxAsUser("u1"), &handler.EchoUpsertInput{
			Body: echoModel.EchoUpsertDto{Content: "hi", Private: true},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.POST_ECHO_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().PostEcho(mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.PostEcho(helpers.CtxAsUser("u1"), &handler.EchoUpsertInput{
			Body: echoModel.EchoUpsertDto{Content: "hi"},
		})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.EmptyOutput{}, out)
	})
}

func TestEchoHandler_UpdateEcho(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().
			UpdateEcho(mock.Anything, mock.MatchedBy(func(e *echoModel.Echo) bool {
				return e != nil && e.ID == "e1" && e.Content == "edited"
			})).
			Return(nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.UpdateEcho(helpers.CtxAsUser("u1"), &handler.EchoUpsertInput{
			Body: echoModel.EchoUpsertDto{ID: "e1", Content: "edited"},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.UPDATE_ECHO_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().UpdateEcho(mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.UpdateEcho(helpers.CtxAsUser("u1"), &handler.EchoUpsertInput{
			Body: echoModel.EchoUpsertDto{ID: "e1"},
		})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_DeleteEcho(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().DeleteEchoById(mock.Anything, "e1").Return(nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.DeleteEcho(helpers.CtxAsUser("u1"), &handler.EchoIDInput{ID: "e1"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DELETE_ECHO_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().DeleteEchoById(mock.Anything, "e1").Return(errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.DeleteEcho(helpers.CtxAsUser("u1"), &handler.EchoIDInput{ID: "e1"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_LikeEcho(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().LikeEcho(mock.Anything, "e1").Return(nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.LikeEcho(helpers.CtxAnonymous(), &handler.LikeEchoInput{ID: "e1"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.LIKE_ECHO_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().LikeEcho(mock.Anything, "e1").Return(errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.LikeEcho(helpers.CtxAnonymous(), &handler.LikeEchoInput{ID: "e1"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetEchoById(t *testing.T) {
	t.Run("success returns the echo", func(t *testing.T) {
		want := helpers.NewEcho()
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchoById(mock.Anything, "e1").Return(&want, nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetEchoById(helpers.CtxAnonymous(), &handler.EchoIDInput{ID: "e1"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_ECHO_BY_ID_SUCCESS, out.Message)
		require.NotNil(t, out.Data)
		assert.Equal(t, want.ID, out.Data.ID)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchoById(mock.Anything, "e1").Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetEchoById(helpers.CtxAnonymous(), &handler.EchoIDInput{ID: "e1"})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.EchoOutput{}, out)
	})
}

func TestEchoHandler_QueryEchos(t *testing.T) {
	t.Run("success forwards the query dto", func(t *testing.T) {
		query := commonModel.EchoQueryDto{Page: 2, PageSize: 5, Search: "go", TagIDs: []string{"t1"}}
		page := newPage([]echoModel.Echo{helpers.NewEcho()})

		svc := echomock.NewMockService(t)
		svc.EXPECT().QueryEchos(mock.Anything, query).Return(page, nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.QueryEchos(helpers.CtxAnonymous(), &handler.QueryEchosInput{Body: query})

		require.NoError(t, err)
		assert.Equal(t, commonModel.QUERY_ECHOS_SUCCESS, out.Message)
		assert.Equal(t, int64(1), out.Data.Total)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().QueryEchos(mock.Anything, mock.Anything).
			Return(commonModel.PageQueryResult[[]echoModel.Echo]{}, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.QueryEchos(helpers.CtxAnonymous(), &handler.QueryEchosInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetEchosByPageGet(t *testing.T) {
	t.Run("maps query params into PageQueryDto", func(t *testing.T) {
		wantDto := commonModel.PageQueryDto{Page: 3, PageSize: 20, Search: "kw"}
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchosByPage(mock.Anything, wantDto).Return(newPage(nil), nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetEchosByPageGet(helpers.CtxAnonymous(), &handler.EchoPageGetInput{
			Page: 3, PageSize: 20, Search: "kw",
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_ECHOS_BY_PAGE_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchosByPage(mock.Anything, mock.Anything).
			Return(commonModel.PageQueryResult[[]echoModel.Echo]{}, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetEchosByPageGet(helpers.CtxAnonymous(), &handler.EchoPageGetInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetEchosByPagePost(t *testing.T) {
	t.Run("forwards the body dto", func(t *testing.T) {
		dto := commonModel.PageQueryDto{Page: 1, PageSize: 10, Search: "x"}
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchosByPage(mock.Anything, dto).Return(newPage(nil), nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetEchosByPagePost(helpers.CtxAnonymous(), &handler.EchoPagePostInput{Body: dto})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_ECHOS_BY_PAGE_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchosByPage(mock.Anything, mock.Anything).
			Return(commonModel.PageQueryResult[[]echoModel.Echo]{}, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetEchosByPagePost(helpers.CtxAnonymous(), &handler.EchoPagePostInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetEchosByTagId(t *testing.T) {
	t.Run("maps tag id and pagination", func(t *testing.T) {
		wantDto := commonModel.PageQueryDto{Page: 2, PageSize: 15, Search: "s"}
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchosByTagId(mock.Anything, "tag-1", wantDto).Return(newPage(nil), nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetEchosByTagId(helpers.CtxAnonymous(), &handler.GetEchosByTagIDInput{
			TagID: "tag-1", Page: 2, PageSize: 15, Search: "s",
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_ECHOS_BY_TAG_ID_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetEchosByTagId(mock.Anything, mock.Anything, mock.Anything).
			Return(commonModel.PageQueryResult[[]echoModel.Echo]{}, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetEchosByTagId(helpers.CtxAnonymous(), &handler.GetEchosByTagIDInput{TagID: "tag-1"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetTodayEchos(t *testing.T) {
	cases := []struct {
		name   string
		header string
		wantTZ string
	}{
		{"valid timezone passes through", "Asia/Tokyo", "Asia/Tokyo"},
		{"invalid timezone falls back to UTC", "Not/AZone", "UTC"},
		{"empty timezone normalizes to UTC", "", "UTC"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := echomock.NewMockService(t)
			svc.EXPECT().GetTodayEchos(mock.Anything, tc.wantTZ).
				Return([]echoModel.Echo{helpers.NewEcho()}, nil).Once()

			h := handler.NewEchoHandler(svc)
			out, err := h.GetTodayEchos(helpers.CtxAnonymous(), &handler.TimezoneInput{Timezone: tc.header})

			require.NoError(t, err)
			assert.Equal(t, commonModel.GET_TODAY_ECHOS_SUCCESS, out.Message)
			assert.Len(t, out.Data, 1)
		})
	}

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetTodayEchos(mock.Anything, mock.Anything).Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetTodayEchos(helpers.CtxAnonymous(), &handler.TimezoneInput{Timezone: "UTC"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetHotEchos(t *testing.T) {
	cases := []struct {
		name      string
		input     int
		wantLimit int
	}{
		{"zero falls back to default 5", 0, 5},
		{"negative falls back to default 5", -3, 5},
		{"positive passes through", 10, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := echomock.NewMockService(t)
			svc.EXPECT().GetHotEchos(mock.Anything, tc.wantLimit).Return([]echoModel.Echo{}, nil).Once()

			h := handler.NewEchoHandler(svc)
			out, err := h.GetHotEchos(helpers.CtxAnonymous(), &handler.GetHotEchosInput{Limit: tc.input})

			require.NoError(t, err)
			assert.Equal(t, commonModel.GET_HOT_ECHOS_SUCCESS, out.Message)
		})
	}

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetHotEchos(mock.Anything, 5).Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetHotEchos(helpers.CtxAnonymous(), &handler.GetHotEchosInput{Limit: 0})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetRandomEcho(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		want := helpers.NewEcho()
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetRandomEcho(mock.Anything).Return(&want, nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetRandomEcho(helpers.CtxAnonymous(), &handler.GetRandomEchoInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_RANDOM_ECHO_SUCCESS, out.Message)
		require.NotNil(t, out.Data)
		assert.Equal(t, want.ID, out.Data.ID)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetRandomEcho(mock.Anything).Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetRandomEcho(helpers.CtxAnonymous(), &handler.GetRandomEchoInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetOnThisDayEchos(t *testing.T) {
	t.Run("normalizes timezone before delegating", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetOnThisDayEchos(mock.Anything, "UTC").Return([]echoModel.Echo{}, nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetOnThisDayEchos(helpers.CtxAnonymous(), &handler.TimezoneInput{Timezone: "bad/zone"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_ON_THIS_DAY_ECHOS_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetOnThisDayEchos(mock.Anything, mock.Anything).Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetOnThisDayEchos(helpers.CtxAnonymous(), &handler.TimezoneInput{Timezone: "UTC"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_GetAllTags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetAllTags().Return([]echoModel.Tag{{ID: "t1", Name: "go"}}, nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.GetAllTags(helpers.CtxAnonymous(), &handler.GetAllTagsInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_ALL_TAGS_SUCCESS, out.Message)
		assert.Len(t, out.Data, 1)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().GetAllTags().Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.GetAllTags(helpers.CtxAnonymous(), &handler.GetAllTagsInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestEchoHandler_CreateTag(t *testing.T) {
	t.Run("forwards tag name", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().CreateTag(mock.Anything, "golang").
			Return(&echoModel.Tag{ID: "t1", Name: "golang"}, nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.CreateTag(helpers.CtxAsUser("u1"), &handler.CreateTagInput{
			Body: echoModel.CreateTagDto{Name: "golang"},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.CREATE_TAG_SUCCESS, out.Message)
		require.NotNil(t, out.Data)
		assert.Equal(t, "golang", out.Data.Name)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(nil, errBoom).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.CreateTag(helpers.CtxAsUser("u1"), &handler.CreateTagInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.TagOutput{}, out)
	})
}

func TestEchoHandler_DeleteTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().DeleteTag(mock.Anything, "t1").Return(nil).Once()

		h := handler.NewEchoHandler(svc)
		out, err := h.DeleteTag(helpers.CtxAsUser("u1"), &handler.TagIDInput{ID: "t1"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DELETE_TAG_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := echomock.NewMockService(t)
		svc.EXPECT().DeleteTag(mock.Anything, "t1").Return(errBoom).Once()

		h := handler.NewEchoHandler(svc)
		_, err := h.DeleteTag(helpers.CtxAsUser("u1"), &handler.TagIDInput{ID: "t1"})

		require.ErrorIs(t, err, errBoom)
	})
}
