package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/repository/connect"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/util/http"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	"sync"
	"time"
)

type ConnectService struct {
	connectRepository repository.ConnectRepositoryInterface
	commonService     commonService.CommonServiceInterface
	settingService    settingService.SettingServiceInterface
}

func NewConnectService(connectRepository repository.ConnectRepositoryInterface, commonService commonService.CommonServiceInterface, settingService settingService.SettingServiceInterface) ConnectServiceInterface {
	return &ConnectService{
		connectRepository: connectRepository,
		commonService:     commonService,
		settingService:    settingService,
	}
}

func (connectService *ConnectService) AddConnect(userid uint, connected model.Connected) error {
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
	connected.ConnectURL = http.TrimURL(connected.ConnectURL)

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
	if err := connectService.connectRepository.CreateConnect(&connected); err != nil {
		return err
	}

	return nil
}

func (connectService *ConnectService) DeleteConnect(userid, id uint) error {
	user, err := connectService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 删除连接地址
	if err := connectService.connectRepository.DeleteConnect(id); err != nil {
		return err
	}

	return nil
}

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

	// 设置 Connect 信息
	connect.ServerName = setting.ServerName
	connect.ServerURL = setting.ServerURL
	connect.Ech0s = status.TotalMessages
	connect.SysUsername = status.Username

	// 处理 Logo URL，避免出现重复的斜杠
	trimmedServerURL := setting.ServerURL
	if len(trimmedServerURL) > 0 && trimmedServerURL[len(trimmedServerURL)-1] == '/' {
		trimmedServerURL = trimmedServerURL[:len(trimmedServerURL)-1]
	}

	if status.Logo != "" {
		// 如果 Logo URL 以 / 开头，去掉一个 /
		logoPath := status.Logo
		if len(logoPath) > 0 && logoPath[0] == '/' {
			logoPath = logoPath[1:]
		}
		connect.Logo = fmt.Sprintf("%s/api/%s", trimmedServerURL, logoPath)
	} else {
		connect.Logo = fmt.Sprintf("%s/favicon.svg", trimmedServerURL)
	}

	return connect, nil
}

func (connectService *ConnectService) GetConnectsInfo() ([]model.Connect, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second) // 8秒超时
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

	// 重试配置
	const maxRetries = 2
	const retryDelay = 500 * time.Millisecond

	for _, conn := range connects {
		wg.Add(1)
		go func(conn model.Connected) {
			defer wg.Done()
			url := httpUtil.TrimURL(conn.ConnectURL) + "/api/connect"
			for attempt := 0; attempt < maxRetries; attempt++ {
				select {
				case <-ctx.Done():
					return // 超时直接退出
				default:
				}
				resp, err := httpUtil.SendRequest(url, "GET", struct {
					Header  string
					Content string
				}{
					Header:  "Ech0_URL",
					Content: conn.ConnectURL,
				})

				if err != nil {
					// 如果不是最后一次重试，等待后继续
					if attempt < maxRetries-1 {
						time.Sleep(retryDelay * time.Duration(attempt+1))
						continue
					}
					// 最后一次重试也失败，放弃该连接
					fmt.Printf("[连接信息获取失败] 地址: %s，阶段: 发送请求，错误: %v\n", conn.ConnectURL, err)
					return
				}

				var connectInfo commonModel.Result[model.Connect]
				if err := json.Unmarshal(resp, &connectInfo); err != nil {
					// JSON 解析失败，如果不是最后一次重试，等待后继续
					if attempt < maxRetries-1 {
						time.Sleep(retryDelay * time.Duration(attempt+1))
						continue
					}
					// 最后一次重试也失败，放弃该连接
					fmt.Printf("[连接信息获取失败] 地址: %s，阶段: 解析响应，错误: %v\n", conn.ConnectURL, err)
					return
				}

				// 验证响应数据
				if connectInfo.Code != 1 || connectInfo.Data.ServerURL == "" {
					lastErr := fmt.Errorf("无效响应: code=%d, serverURL=%s", connectInfo.Code, connectInfo.Data.ServerURL)
					// 数据无效，如果不是最后一次重试，等待后继续
					if attempt < maxRetries-1 {
						time.Sleep(retryDelay * time.Duration(attempt+1))
						continue
					}
					// 最后一次重试也失败，放弃该连接
					fmt.Printf("[连接信息获取失败] 地址: %s，阶段: 校验数据，错误: %v\n", conn.ConnectURL, lastErr)
					return
				}

				// 成功获取有效数据，检查重复并发送
				seenMutex.Lock()
				if _, exists := seenURLs[connectInfo.Data.ServerURL]; exists {
					seenMutex.Unlock()
					return // 重复数据，直接返回
				}
				seenURLs[connectInfo.Data.ServerURL] = struct{}{}
				seenMutex.Unlock()

				connectChan <- connectInfo.Data
				return // 成功处理，退出重试循环
			}
		}(conn)
	}

	// 收集结果时也要支持超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(connectChan)
		close(done)
	}()

	select {
	case <-done:
		// 正常收集完毕
	case <-ctx.Done():
		// 超时，提前返回
	}

	for connect := range connectChan {
		if connect.ServerURL == "" {
			continue
		}
		connectList = append(connectList, connect)
	}

	return connectList, nil
}

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
