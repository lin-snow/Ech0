package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	repository "github.com/lin-snow/ech0/internal/repository/connect"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type ConnectService struct {
	txManager         transaction.TransactionManager
	connectRepository repository.ConnectRepositoryInterface
	echoRepository    echoRepository.EchoRepositoryInterface
	commonService     commonService.CommonServiceInterface
	settingService    settingService.SettingServiceInterface
}

func NewConnectService(
	tm transaction.TransactionManager,
	connectRepository repository.ConnectRepositoryInterface,
	echoRepository echoRepository.EchoRepositoryInterface,
	commonService commonService.CommonServiceInterface,
	settingService settingService.SettingServiceInterface,
) ConnectServiceInterface {
	return &ConnectService{
		txManager:         tm,
		connectRepository: connectRepository,
		echoRepository:    echoRepository,
		commonService:     commonService,
		settingService:    settingService,
	}
}

// AddConnect 添加连接
func (connectService *ConnectService) AddConnect(userid uint, connected model.Connected) error {
	return connectService.txManager.Run(func(ctx context.Context) error {
		user, err := connectService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}

		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 检查连接地址是否为空
		if connected.ConnectURL == "" {
			return errors.New(commonModel.INVALID_CONNECTION_URL)
		}

		// 去除连接地址前后的空格和斜杠
		connected.ConnectURL = httpUtil.TrimURL(connected.ConnectURL)

		// 检查连接地址是否已存在
		connectedList, err := connectService.connectRepository.GetAllConnects()
		if err != nil {
			return err
		}

		// 检查连接地址是否已存在
		for _, conn := range connectedList {
			if conn.ConnectURL == connected.ConnectURL {
				return errors.New(commonModel.CONNECT_HAS_EXISTS)
			}
		}

		// 添加连接地址
		if err := connectService.connectRepository.CreateConnect(ctx, &connected); err != nil {
			return err
		}

		return nil
	})
}

// DeleteConnect 删除连接
func (connectService *ConnectService) DeleteConnect(userid, id uint) error {
	return connectService.txManager.Run(func(ctx context.Context) error {
		user, err := connectService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}

		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 删除连接地址
		if err := connectService.connectRepository.DeleteConnect(ctx, id); err != nil {
			return err
		}

		return nil
	})
}

// GetConnect 提供当前实例的连接信息
func (connectService *ConnectService) GetConnect() (model.Connect, error) {
	var connect model.Connect

	// 获取系统设置
	var setting settingModel.SystemSetting
	if err := connectService.settingService.GetSetting(&setting); err != nil {
		return connect, err
	}

	// 获取系统状态
	status, err := connectService.commonService.GetStatus()
	if err != nil {
		return connect, err
	}

	// 统计当天发布的数量
	todayEchos := connectService.echoRepository.GetTodayEchos(true, "UTC")

	// 设置 Connect 信息
	connect.ServerName = setting.ServerName
	connect.ServerURL = setting.ServerURL
	connect.TotalEchos = status.TotalEchos
	connect.TodayEchos = len(todayEchos)
	connect.SysUsername = status.Username

	// 处理 Logo URL，避免出现重复的斜杠
	trimmedServerURL := setting.ServerURL
	if len(trimmedServerURL) > 0 && trimmedServerURL[len(trimmedServerURL)-1] == '/' {
		trimmedServerURL = trimmedServerURL[:len(trimmedServerURL)-1]
	}

	if setting.ServerLogo != "" {
		// 如果 Logo URL 以 / 开头，去掉一个 /
		logoPath := setting.ServerLogo
		if len(logoPath) > 0 && logoPath[0] == '/' {
			logoPath = logoPath[1:]
		}
		connect.Logo = fmt.Sprintf("%s/api/%s", trimmedServerURL, logoPath)
	} else {
		connect.Logo = fmt.Sprintf("%s/Ech0.svg", trimmedServerURL)
	}

	return connect, nil
}

// GetConnectsInfo 获取实例获取到的其它实例的连接信息
func (connectService *ConnectService) GetConnectsInfo() ([]model.Connect, error) {
	// 总超时时间：给予足够的缓冲，避免单个慢连接导致整体超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 获取所有连接地址
	connects, err := connectService.connectRepository.GetAllConnects()
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

	seenURLs := make(map[string]struct{})
	var seenMutex sync.Mutex

	// 重试配置：平衡速度和可靠性
	const maxRetries = 3
	const baseDelay = 1 * time.Second
	const requestTimeout = 3 * time.Second // 单个请求超时时间（降低到3秒，加快失败检测）

	for _, conn := range connects {
		wg.Add(1)
		go func(conn model.Connected) {
			defer wg.Done()
			url := httpUtil.TrimURL(conn.ConnectURL) + "/api/connect"

			var lastErr error
			for attempt := 0; attempt < maxRetries; attempt++ {
				select {
				case <-ctx.Done():
					logUtil.GetLogger().
						Info("[连接信息获取取消]", zap.String("地址", conn.ConnectURL), zap.Error(ctx.Err()))
					return // 总体超时直接退出
				default:
				}

				// 计算当前重试的延迟时间（指数退避）
				if attempt > 0 {
					delay := baseDelay * time.Duration(1<<(attempt-1)) // 1s, 2s, 4s...
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return
					}
				}

				resp, err := httpUtil.SendRequest(url, "GET", struct {
					Header  string
					Content string
				}{
					Header:  "Ech0_URL",
					Content: conn.ConnectURL,
				}, requestTimeout) // 传入自定义超时时间
				if err != nil {
					lastErr = err
					logUtil.GetLogger().Error("[连接信息获取失败]",
						zap.String("地址", conn.ConnectURL),
						zap.Int("尝试次数", attempt+1),
						zap.Error(err),
					)

					// 如果是最后一次重试，记录最终失败
					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("[连接信息最终失败]",
							zap.String("地址", conn.ConnectURL),
							zap.Int("已重试次数", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				var connectInfo commonModel.Result[model.Connect]
				if err := json.Unmarshal(resp, &connectInfo); err != nil {
					lastErr = fmt.Errorf("JSON解析失败: %w", err)
					logUtil.GetLogger().Error("[连接信息解析失败]",
						zap.String("地址", conn.ConnectURL),
						zap.Int("尝试次数", attempt+1),
						zap.Error(lastErr),
					)

					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("[连接信息最终失败]",
							zap.String("地址", conn.ConnectURL),
							zap.Int("已重试次数", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				// 验证响应数据
				if connectInfo.Code != 1 {
					lastErr = fmt.Errorf("响应码无效: %d, 消息: %s", connectInfo.Code, connectInfo.Message)
					logUtil.GetLogger().Error("[连接信息校验失败]",
						zap.String("地址", conn.ConnectURL),
						zap.Int("尝试次数", attempt+1),
						zap.Error(lastErr),
					)

					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("[连接信息最终失败]",
							zap.String("地址", conn.ConnectURL),
							zap.Int("已重试次数", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				if connectInfo.Data.ServerURL == "" {
					lastErr = fmt.Errorf("服务器URL为空")
					logUtil.GetLogger().Error("[连接信息校验失败]",
						zap.String("地址", conn.ConnectURL),
						zap.Int("尝试次数", attempt+1),
						zap.Error(lastErr),
					)

					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("[连接信息最终失败]",
							zap.String("地址", conn.ConnectURL),
							zap.Int("已重试次数", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				// 成功获取有效数据，检查重复并发送
				seenMutex.Lock()
				if _, exists := seenURLs[connectInfo.Data.ServerURL]; exists {
					seenMutex.Unlock()
					logUtil.GetLogger().Info("[连接信息重复]",
						zap.String("地址", conn.ConnectURL),
						zap.String("ServerURL", connectInfo.Data.ServerURL),
					)
					return // 重复数据，直接返回
				}
				seenURLs[connectInfo.Data.ServerURL] = struct{}{}
				seenMutex.Unlock()

				logUtil.GetLogger().Info("[连接信息获取成功]",
					zap.String("地址", conn.ConnectURL),
					zap.String("服务器", connectInfo.Data.ServerName),
				)
				connectChan <- connectInfo.Data
				return // 成功处理，退出重试循环
			}
		}(conn)
	}

	// 使用带缓冲的通道来避免goroutine泄漏
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(connectChan)
		close(done)
	}()

	// 在单独的 goroutine 中收集结果，使用 mutex 保护并发写入
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

	// 等待完成或超时
	select {
	case <-done:
		// 所有 goroutine 完成，等待收集完成
		<-collectDone
		mu.Lock()
		count := len(connectList)
		mu.Unlock()
		logUtil.GetLogger().Info("[连接信息收集完成]", zap.Int("有效连接数", count))
	case <-ctx.Done():
		// 超时，等待收集器完成或超时
		logUtil.GetLogger().Info("[连接信息收集超时] 等待处理已收集到的数据")
		select {
		case <-collectDone:
			// 收集器已完成
			logUtil.GetLogger().Info("[收集器已完成]")
		case <-time.After(200 * time.Millisecond):
			// 给收集器额外的时间处理缓冲区中的数据
			logUtil.GetLogger().Info("[收集器处理超时]")
		}
		mu.Lock()
		count := len(connectList)
		mu.Unlock()
		logUtil.GetLogger().Info("[连接信息收集超时完成]", zap.Int("有效连接数", count))
	}

	// 安全地返回结果
	mu.Lock()
	defer mu.Unlock()
	return connectList, nil
}

// GetConnects 获取当前实例添加的所有连接
func (connectService *ConnectService) GetConnects() ([]model.Connected, error) {
	// 获取所有连接地址
	connects, err := connectService.connectRepository.GetAllConnects()
	if err != nil {
		return nil, err
	}

	// 如果没有找到，返回空切片
	if len(connects) == 0 {
		return []model.Connected{}, nil
	}

	// 返回查询到的 connects
	return connects, nil
}
