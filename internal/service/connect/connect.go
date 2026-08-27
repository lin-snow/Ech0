// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/internal/util/egress"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"golang.org/x/sync/singleflight"
)

const (
	connectsInfoCacheTTL        = 30 * time.Minute
	connectFanoutMaxConcurrency = 8
	connectsInfoSingleflightKey = "connects_info"
	siteMetricsTimezone         = "UTC"
	healthProbeTimeout          = 5 * time.Second
	healthOverallTimeout        = 30 * time.Second
	connectRetryBaseDelay       = 1 * time.Second
)

type ConnectService struct {
	transactor        transaction.Transactor
	connectRepository Repository
	echoRepository    EchoRepository
	commonService     CommonService
	durableKV         kvstore.Store

	connectsInfoCacheMu      sync.RWMutex
	connectsInfoCache        []model.Connect
	connectsInfoCacheExpires time.Time
	connectsInfoCacheValid   bool
	connectsInfoFetcher      singleflight.Group

	peerFetcher func(peerConnectURL string, requestTimeout time.Duration) (model.Connect, error)

	retryBaseDelay time.Duration
}

func NewConnectService(
	tx transaction.Transactor,
	connectRepository Repository,
	echoRepository EchoRepository,
	commonService CommonService,
	durableKV kvstore.Store,
) *ConnectService {
	return &ConnectService{
		transactor:        tx,
		connectRepository: connectRepository,
		echoRepository:    echoRepository,
		commonService:     commonService,
		durableKV:         durableKV,
		peerFetcher:       fetchPeerConnectInfo,
		retryBaseDelay:    connectRetryBaseDelay,
	}
}

func (connectService *ConnectService) WithPeerFetcher(
	f func(peerConnectURL string, requestTimeout time.Duration) (model.Connect, error),
) *ConnectService {
	connectService.peerFetcher = f
	return connectService
}

func (connectService *ConnectService) WithRetryBaseDelay(d time.Duration) *ConnectService {
	connectService.retryBaseDelay = d
	return connectService
}

func (connectService *ConnectService) AddConnect(ctx context.Context, connected model.Connected) error {
	userid := viewer.MustFromContext(ctx).UserID()
	if err := connectService.transactor.Run(ctx, func(txCtx context.Context) error {
		user, err := connectService.commonService.CommonGetUserByUserId(txCtx, userid)
		if err != nil {
			return err
		}

		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		if connected.ConnectURL == "" {
			return errors.New(commonModel.INVALID_CONNECTION_URL)
		}

		connected.ConnectURL = urlUtil.TrimURL(connected.ConnectURL)

		if err := egress.Validate(connected.ConnectURL + "/api/connect"); err != nil {
			return errors.New(commonModel.INVALID_CONNECTION_URL)
		}

		connectedList, err := connectService.connectRepository.GetAllConnects(txCtx)
		if err != nil {
			return err
		}

		for _, conn := range connectedList {
			if conn.ConnectURL == connected.ConnectURL {
				return errors.New(commonModel.CONNECT_HAS_EXISTS)
			}
		}

		if err := connectService.connectRepository.CreateConnect(txCtx, &connected); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	connectService.invalidateConnectsInfoCache()
	return nil
}

func (connectService *ConnectService) DeleteConnect(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	if err := connectService.transactor.Run(ctx, func(txCtx context.Context) error {
		user, err := connectService.commonService.CommonGetUserByUserId(txCtx, userid)
		if err != nil {
			return err
		}

		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		if err := connectService.connectRepository.DeleteConnect(txCtx, id); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	connectService.invalidateConnectsInfoCache()
	return nil
}

func (connectService *ConnectService) GetConnect() (model.Connect, error) {
	var connect model.Connect

	setting, err := coreSetting.Get(context.Background(), connectService.durableKV, coreSetting.System)
	if err != nil {
		return connect, err
	}

	owner, err := connectService.commonService.GetOwner()
	if err != nil {
		return connect, err
	}

	todayEchos := connectService.echoRepository.GetTodayEchos(true, siteMetricsTimezone)
	_, totalEchos := connectService.echoRepository.GetEchosByPage(1, 1, "", true)

	connect.ServerName = setting.ServerName
	connect.ServerURL = setting.ServerURL
	connect.TotalEchos = int(totalEchos)
	connect.TodayEchos = len(todayEchos)
	connect.SysUsername = owner.Username
	connect.Version = versionPkg.Version

	trimmedServerURL := strings.TrimRight(setting.ServerURL, "/")
	logoPath := strings.TrimSpace(setting.ServerLogo)

	if logoPath == "" || logoPath == "Ech0.svg" || logoPath == "/Ech0.svg" {
		connect.Logo = fmt.Sprintf("%s/Ech0.svg", trimmedServerURL)
	} else if strings.HasPrefix(logoPath, "http://") || strings.HasPrefix(logoPath, "https://") {
		connect.Logo = logoPath
	} else if strings.HasPrefix(logoPath, "/") {
		connect.Logo = fmt.Sprintf("%s%s", trimmedServerURL, logoPath)
	} else {
		connect.Logo = fmt.Sprintf("%s/%s", trimmedServerURL, logoPath)
	}

	return connect, nil
}

func (connectService *ConnectService) GetConnectsInfo() ([]model.Connect, error) {
	if cached, ok := connectService.getCachedConnectsInfo(); ok {
		return cached, nil
	}

	result, err, _ := connectService.connectsInfoFetcher.Do(connectsInfoSingleflightKey, func() (any, error) {
		if cached, ok := connectService.getCachedConnectsInfo(); ok {
			return cached, nil
		}

		connectList, fetchErr := connectService.fetchConnectsInfo()
		if fetchErr != nil {
			return nil, fetchErr
		}

		connectService.setCachedConnectsInfo(connectList)
		return cloneConnects(connectList), nil
	})
	if err != nil {
		return nil, err
	}

	connects, ok := result.([]model.Connect)
	if !ok {
		return nil, fmt.Errorf("invalid cache result type")
	}

	return cloneConnects(connects), nil
}

func fetchPeerConnectInfo(peerConnectURL string, requestTimeout time.Duration) (model.Connect, error) {
	url := urlUtil.TrimURL(peerConnectURL) + "/api/connect"
	resp, err := egress.Fetch(url, "GET", egress.Header{
		Header:  "Ech0_URL",
		Content: peerConnectURL,
	}, requestTimeout)
	if err != nil {
		return model.Connect{}, err
	}

	var connectInfo commonModel.Result[model.Connect]
	if err := json.Unmarshal(resp, &connectInfo); err != nil {
		return model.Connect{}, fmt.Errorf("JSON解析失败: %w", err)
	}
	if connectInfo.Code != 1 {
		return model.Connect{}, fmt.Errorf("响应码无效: %d, 消息: %s", connectInfo.Code, connectInfo.Message)
	}
	if connectInfo.Data.ServerURL == "" {
		return model.Connect{}, fmt.Errorf("服务器URL为空")
	}
	return connectInfo.Data, nil
}

func (connectService *ConnectService) fetchConnectsInfo() ([]model.Connect, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	connects, err := connectService.connectRepository.GetAllConnects(context.Background())
	if err != nil {
		return nil, err
	}

	if len(connects) == 0 {
		return []model.Connect{}, nil
	}

	var connectList []model.Connect
	connectList = make([]model.Connect, 0, len(connects))

	var wg sync.WaitGroup
	connectChan := make(chan model.Connect, len(connects))
	semaphore := make(chan struct{}, connectFanoutMaxConcurrency)

	seenURLs := make(map[string]struct{})
	var seenMutex sync.Mutex

	const maxRetries = 3
	baseDelay := connectService.retryBaseDelay
	const requestTimeout = 3 * time.Second

	for _, conn := range connects {
		wg.Add(1)
		go func(conn model.Connected) {
			defer wg.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-semaphore }()

			var lastErr error
			for attempt := range maxRetries {
				select {
				case <-ctx.Done():
					logUtil.GetLogger().
						Info(
							"fetch connection info cancelled",
							slog.String("module", "connect"),
							slog.String("connect_url", conn.ConnectURL),
							logUtil.Err(ctx.Err()),
						)
					return
				default:
				}

				if attempt > 0 {
					delay := baseDelay * time.Duration(1<<(attempt-1))
					if delay > 0 {
						select {
						case <-time.After(delay):
						case <-ctx.Done():
							return
						}
					}
				}

				data, err := connectService.peerFetcher(conn.ConnectURL, requestTimeout)
				if err != nil {
					lastErr = err
					logUtil.GetLogger().Error("fetch connection info failed",
						slog.String("module", "connect"),
						slog.String("connect_url", conn.ConnectURL),
						slog.Int("attempt", attempt+1),
						logUtil.Err(lastErr),
					)
					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("fetch connection info exhausted retries",
							slog.String("module", "connect"),
							slog.String("connect_url", conn.ConnectURL),
							slog.Int("retries", maxRetries),
							logUtil.Err(lastErr),
						)
					}
					continue
				}

				seenMutex.Lock()
				if _, exists := seenURLs[data.ServerURL]; exists {
					seenMutex.Unlock()
					logUtil.GetLogger().Info("connection info duplicated",
						slog.String("module", "connect"),
						slog.String("connect_url", conn.ConnectURL),
						slog.String("server_url", data.ServerURL),
					)
					return
				}
				seenURLs[data.ServerURL] = struct{}{}
				seenMutex.Unlock()

				logUtil.GetLogger().Info("fetch connection info succeeded",
					slog.String("module", "connect"),
					slog.String("connect_url", conn.ConnectURL),
					slog.String("server_name", data.ServerName),
				)
				connectChan <- data
				return
			}
		}(conn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(connectChan)
		close(done)
	}()

	var mu sync.Mutex
	collectDone := make(chan struct{})
	go func() {
		for connect := range connectChan {
			if connect.ServerURL != "" {
				mu.Lock()
				connectList = append(connectList, connect)
				mu.Unlock()
			}
		}
		close(collectDone)
	}()

	select {
	case <-done:
		<-collectDone
		mu.Lock()
		count := len(connectList)
		mu.Unlock()
		logUtil.GetLogger().Info("collect connection info completed", slog.String("module", "connect"), slog.Int("valid_count", count))
	case <-ctx.Done():
		logUtil.GetLogger().Info("collect connection info timeout, waiting collector", slog.String("module", "connect"))
		select {
		case <-collectDone:
			logUtil.GetLogger().Info("collector completed", slog.String("module", "connect"))
		case <-time.After(200 * time.Millisecond):
			logUtil.GetLogger().Info("collector timeout", slog.String("module", "connect"))
		}
		mu.Lock()
		count := len(connectList)
		mu.Unlock()
		logUtil.GetLogger().Info("collect connection info timeout completed", slog.String("module", "connect"), slog.Int("valid_count", count))
	}

	mu.Lock()
	defer mu.Unlock()
	return connectList, nil
}

func (connectService *ConnectService) invalidateConnectsInfoCache() {
	connectService.connectsInfoCacheMu.Lock()
	defer connectService.connectsInfoCacheMu.Unlock()
	connectService.connectsInfoCache = nil
	connectService.connectsInfoCacheExpires = time.Time{}
	connectService.connectsInfoCacheValid = false
}

func (connectService *ConnectService) getCachedConnectsInfo() ([]model.Connect, bool) {
	connectService.connectsInfoCacheMu.RLock()
	defer connectService.connectsInfoCacheMu.RUnlock()

	if !connectService.connectsInfoCacheValid {
		return nil, false
	}
	if time.Now().After(connectService.connectsInfoCacheExpires) {
		return nil, false
	}
	return cloneConnects(connectService.connectsInfoCache), true
}

func (connectService *ConnectService) setCachedConnectsInfo(connects []model.Connect) {
	connectService.connectsInfoCacheMu.Lock()
	defer connectService.connectsInfoCacheMu.Unlock()
	connectService.connectsInfoCache = cloneConnects(connects)
	connectService.connectsInfoCacheExpires = time.Now().Add(connectsInfoCacheTTL)
	connectService.connectsInfoCacheValid = true
}

func cloneConnects(connects []model.Connect) []model.Connect {
	if len(connects) == 0 {
		return []model.Connect{}
	}
	cloned := make([]model.Connect, len(connects))
	copy(cloned, connects)
	return cloned
}

func (connectService *ConnectService) GetConnects() ([]model.Connected, error) {
	connects, err := connectService.connectRepository.GetAllConnects(context.Background())
	if err != nil {
		return nil, err
	}

	if len(connects) == 0 {
		return []model.Connected{}, nil
	}

	return connects, nil
}

func (connectService *ConnectService) GetConnectsHealth() ([]model.ConnectedHealth, error) {
	connects, err := connectService.connectRepository.GetAllConnects(context.Background())
	if err != nil {
		return nil, err
	}
	if len(connects) == 0 {
		return []model.ConnectedHealth{}, nil
	}

	out := make([]model.ConnectedHealth, len(connects))
	probeCtx, cancelProbe := context.WithTimeout(context.Background(), healthOverallTimeout)
	defer cancelProbe()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, connectFanoutMaxConcurrency)

	for i := range connects {
		wg.Add(1)
		go func(i int, conn model.Connected) {
			defer wg.Done()
			select {
			case semaphore <- struct{}{}:
			case <-probeCtx.Done():
				out[i] = model.ConnectedHealth{
					ID: conn.ID, ConnectURL: conn.ConnectURL,
					Status: "offline", Version: "",
				}
				return
			}
			defer func() { <-semaphore }()

			h := model.ConnectedHealth{
				ID: conn.ID, ConnectURL: conn.ConnectURL,
				Status: "offline", Version: "",
			}

			data, err := connectService.peerFetcher(conn.ConnectURL, healthProbeTimeout)
			if err != nil {
				out[i] = h
				return
			}
			h.Status = "online"
			h.Version = data.Version
			out[i] = h
		}(i, connects[i])
	}

	wg.Wait()
	return out, nil
}
