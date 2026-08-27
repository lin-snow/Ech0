// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/lin-snow/ech0/internal/test/mocks/connectmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type peerFetch = func(peerConnectURL string, requestTimeout time.Duration) (model.Connect, error)

func cannedFetcher(t *testing.T, byURL map[string]struct {
	connect model.Connect
	err     error
},
) peerFetch {
	t.Helper()
	return func(peerConnectURL string, _ time.Duration) (model.Connect, error) {
		entry, ok := byURL[peerConnectURL]
		if !ok {
			t.Errorf("peerFetcher called with unexpected url %q", peerConnectURL)
			return model.Connect{}, errors.New("unexpected url")
		}
		return entry.connect, entry.err
	}
}

func serverURLs(connects []model.Connect) []string {
	out := make([]string, 0, len(connects))
	for _, c := range connects {
		out = append(out, c.ServerURL)
	}
	return out
}

func TestAddConnect_CreateConnectErrorPropagates(t *testing.T) {
	const userID = "u-1"
	wantErr := errors.New("insert failed")
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return([]model.Connected{}, nil).Once()
	repo.EXPECT().
		CreateConnect(mock.Anything, mock.Anything).
		Return(wantErr).
		Once()

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: "https://example.com"})

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestDeleteConnect_NonAdminDenied(t *testing.T) {
	const userID = "u-1"
	tx := passthroughTx(t)
	cs := commonmock.NewMockService(t)
	cs.EXPECT().
		CommonGetUserByUserId(mock.Anything, userID).
		Return(userModel.User{IsAdmin: false}, nil).
		Once()
	repo := connectmock.NewMockRepository(t)

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.DeleteConnect(helpers.CtxAsUser(userID), "id-1")

	require.Error(t, err)
	assert.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
}

func TestDeleteConnect_UserLookupErrorPropagates(t *testing.T) {
	const userID = "u-1"
	wantErr := errors.New("user lookup failed")
	tx := passthroughTx(t)
	cs := commonmock.NewMockService(t)
	cs.EXPECT().
		CommonGetUserByUserId(mock.Anything, userID).
		Return(userModel.User{}, wantErr).
		Once()
	repo := connectmock.NewMockRepository(t)

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.DeleteConnect(helpers.CtxAsUser(userID), "id-1")

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestDeleteConnect_RepoErrorPropagates(t *testing.T) {
	const userID = "u-1"
	wantErr := errors.New("delete failed")
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().DeleteConnect(mock.Anything, "id-1").Return(wantErr).Once()

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.DeleteConnect(helpers.CtxAsUser(userID), "id-1")

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestDeleteConnect_Success(t *testing.T) {
	const userID = "u-1"
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().DeleteConnect(mock.Anything, "id-1").Return(nil).Once()

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.DeleteConnect(helpers.CtxAsUser(userID), "id-1")

	require.NoError(t, err)
}

func TestGetConnects_Empty(t *testing.T) {
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return([]model.Connected{}, nil).Once()

	svc := connectService.NewConnectService(nil, repo, nil, nil, nil)
	got, err := svc.GetConnects()

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got)
}

func TestGetConnects_ReturnsList(t *testing.T) {
	want := []model.Connected{
		{ID: "a", ConnectURL: "https://a.example"},
		{ID: "b", ConnectURL: "https://b.example"},
	}
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(want, nil).Once()

	svc := connectService.NewConnectService(nil, repo, nil, nil, nil)
	got, err := svc.GetConnects()

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetConnects_ErrorPropagates(t *testing.T) {
	wantErr := errors.New("list failed")
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(nil, wantErr).Once()

	svc := connectService.NewConnectService(nil, repo, nil, nil, nil)
	got, err := svc.GetConnects()

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
}

func TestGetConnectsInfo_EmptyConnects(t *testing.T) {
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return([]model.Connected{}, nil).Once()

	fetcher := func(string, time.Duration) (model.Connect, error) {
		t.Errorf("peerFetcher must not be called when there are no connects")
		return model.Connect{}, nil
	}
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).WithPeerFetcher(fetcher)

	got, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got)
}

func TestGetConnectsInfo_RepoErrorPropagates(t *testing.T) {
	wantErr := errors.New("list failed")
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(nil, wantErr).Once()

	svc := connectService.NewConnectService(nil, repo, nil, nil, nil)
	got, err := svc.GetConnectsInfo()

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
}

func TestGetConnectsInfo_FanoutAggregatesPeers(t *testing.T) {
	connects := []model.Connected{
		{ID: "1", ConnectURL: "https://one.example"},
		{ID: "2", ConnectURL: "https://two.example"},
		{ID: "3", ConnectURL: "https://three.example"},
	}
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(connects, nil).Once()

	fetcher := cannedFetcher(t, map[string]struct {
		connect model.Connect
		err     error
	}{
		"https://one.example":   {connect: model.Connect{ServerName: "one", ServerURL: "https://one.srv"}},
		"https://two.example":   {connect: model.Connect{ServerName: "two", ServerURL: "https://two.srv"}},
		"https://three.example": {connect: model.Connect{ServerName: "three", ServerURL: "https://three.srv"}},
	})
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).WithPeerFetcher(fetcher)

	got, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	assert.ElementsMatch(t,
		[]string{"https://one.srv", "https://two.srv", "https://three.srv"},
		serverURLs(got),
	)
}

func TestGetConnectsInfo_DedupBySeenServerURL(t *testing.T) {
	connects := []model.Connected{
		{ID: "1", ConnectURL: "https://a.example"},
		{ID: "2", ConnectURL: "https://b.example"},
	}
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(connects, nil).Once()

	fetcher := cannedFetcher(t, map[string]struct {
		connect model.Connect
		err     error
	}{
		"https://a.example": {connect: model.Connect{ServerName: "dup", ServerURL: "https://same.srv"}},
		"https://b.example": {connect: model.Connect{ServerName: "dup", ServerURL: "https://same.srv"}},
	})
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).WithPeerFetcher(fetcher)

	got, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "https://same.srv", got[0].ServerURL)
}

func TestGetConnectsInfo_PartialFailureAggregation(t *testing.T) {
	connects := []model.Connected{
		{ID: "1", ConnectURL: "https://good.example"},
		{ID: "2", ConnectURL: "https://bad.example"},
	}
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(connects, nil).Once()

	fetcher := cannedFetcher(t, map[string]struct {
		connect model.Connect
		err     error
	}{
		"https://good.example": {connect: model.Connect{ServerName: "good", ServerURL: "https://good.srv"}},
		"https://bad.example":  {err: errors.New("peer unreachable")},
	})
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).
		WithPeerFetcher(fetcher).
		WithRetryBaseDelay(0)

	got, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "https://good.srv", got[0].ServerURL)
}

func TestGetConnectsInfo_RetryThenSuccess(t *testing.T) {
	connects := []model.Connected{{ID: "1", ConnectURL: "https://flaky.example"}}
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(connects, nil).Once()

	var calls atomic.Int32
	fetcher := func(url string, _ time.Duration) (model.Connect, error) {
		if calls.Add(1) == 1 {
			return model.Connect{}, errors.New("transient")
		}
		return model.Connect{ServerName: "flaky", ServerURL: "https://flaky.srv"}, nil
	}
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).
		WithPeerFetcher(fetcher).
		WithRetryBaseDelay(0)

	got, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "https://flaky.srv", got[0].ServerURL)
	assert.Equal(t, int32(2), calls.Load(), "expected one failed attempt then one success")
}

func TestGetConnectsInfo_CacheHitAndInvalidation(t *testing.T) {
	const userID = "u-1"
	connects := []model.Connected{{ID: "id-1", ConnectURL: "https://peer.example"}}

	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(connects, nil)
	repo.EXPECT().DeleteConnect(mock.Anything, "id-1").Return(nil).Once()

	cs := adminCommon(t, userID)
	tx := passthroughTx(t)

	var fetchCount atomic.Int32
	fetcher := func(string, time.Duration) (model.Connect, error) {
		fetchCount.Add(1)
		return model.Connect{ServerName: "peer", ServerURL: "https://peer.srv"}, nil
	}
	svc := connectService.NewConnectService(tx, repo, nil, cs, nil).WithPeerFetcher(fetcher)

	first, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.Equal(t, int32(1), fetchCount.Load())

	second, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	assert.Equal(t, first, second)
	assert.Equal(t, int32(1), fetchCount.Load(), "second call must hit cache")

	require.NoError(t, svc.DeleteConnect(helpers.CtxAsUser(userID), "id-1"))

	third, err := svc.GetConnectsInfo()
	require.NoError(t, err)
	require.Len(t, third, 1)
	assert.Equal(t, int32(2), fetchCount.Load(), "cache invalidation must force a re-fetch")
}

func TestGetConnectsInfo_SingleflightCollapsesConcurrentCalls(t *testing.T) {
	connects := []model.Connected{{ID: "1", ConnectURL: "https://peer.example"}}

	var getAllCount atomic.Int32
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().
		GetAllConnects(mock.Anything).
		RunAndReturn(func(context.Context) ([]model.Connected, error) {
			getAllCount.Add(1)
			return connects, nil
		})

	var (
		fetchCount  atomic.Int32
		startedOnce sync.Once
		started     = make(chan struct{})
		release     = make(chan struct{})
	)
	fetcher := func(string, time.Duration) (model.Connect, error) {
		fetchCount.Add(1)
		startedOnce.Do(func() { close(started) })
		<-release
		return model.Connect{ServerName: "peer", ServerURL: "https://peer.srv"}, nil
	}
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).WithPeerFetcher(fetcher)

	const callers = 6
	results := make(chan []model.Connect, callers)
	errs := make(chan error, callers)

	go func() {
		got, err := svc.GetConnectsInfo()
		results <- got
		errs <- err
	}()
	<-started

	var launched sync.WaitGroup
	launched.Add(callers - 1)
	for range callers - 1 {
		go func() {
			launched.Done()
			got, err := svc.GetConnectsInfo()
			results <- got
			errs <- err
		}()
	}
	launched.Wait()
	for range 50 {
		runtime.Gosched()
	}
	close(release)

	var first []model.Connect
	for range callers {
		require.NoError(t, <-errs)
		got := <-results
		require.Len(t, got, 1)
		if first == nil {
			first = got
		} else {
			assert.Equal(t, first, got, "all concurrent callers must observe the same snapshot")
		}
	}

	assert.Equal(t, int32(1), fetchCount.Load(), "singleflight must collapse to a single peer fetch")
	assert.Equal(t, int32(1), getAllCount.Load(), "singleflight must collapse to a single repository read")
}

func TestGetConnectsHealth_Empty(t *testing.T) {
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return([]model.Connected{}, nil).Once()

	svc := connectService.NewConnectService(nil, repo, nil, nil, nil)
	got, err := svc.GetConnectsHealth()

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got)
}

func TestGetConnectsHealth_RepoErrorPropagates(t *testing.T) {
	wantErr := errors.New("list failed")
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(nil, wantErr).Once()

	svc := connectService.NewConnectService(nil, repo, nil, nil, nil)
	got, err := svc.GetConnectsHealth()

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
}

func TestGetConnectsHealth_MixedOnlineOffline(t *testing.T) {
	connects := []model.Connected{
		{ID: "online-1", ConnectURL: "https://up.example"},
		{ID: "offline-1", ConnectURL: "https://down.example"},
	}
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().GetAllConnects(mock.Anything).Return(connects, nil).Once()

	fetcher := cannedFetcher(t, map[string]struct {
		connect model.Connect
		err     error
	}{
		"https://up.example":   {connect: model.Connect{ServerURL: "https://up.srv", Version: "1.2.3"}},
		"https://down.example": {err: errors.New("connection refused")},
	})
	svc := connectService.NewConnectService(nil, repo, nil, nil, nil).WithPeerFetcher(fetcher)

	got, err := svc.GetConnectsHealth()
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "online-1", got[0].ID)
	assert.Equal(t, "https://up.example", got[0].ConnectURL)
	assert.Equal(t, "online", got[0].Status)
	assert.Equal(t, "1.2.3", got[0].Version)

	assert.Equal(t, "offline-1", got[1].ID)
	assert.Equal(t, "https://down.example", got[1].ConnectURL)
	assert.Equal(t, "offline", got[1].Status)
	assert.Empty(t, got[1].Version)
}
