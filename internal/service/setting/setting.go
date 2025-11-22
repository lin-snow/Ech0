package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/event"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	settingRepository "github.com/lin-snow/ech0/internal/repository/setting"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/transaction"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

type SettingService struct {
	txManager          transaction.TransactionManager
	commonService      commonService.CommonServiceInterface
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface
	settingRepository  settingRepository.SettingRepositoryInterface
	webhookRepository  webhookRepository.WebhookRepositoryInterface
	eventBus           event.IEventBus
}

func NewSettingService(
	tm transaction.TransactionManager,
	commonService commonService.CommonServiceInterface,
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface,
	settingRepository settingRepository.SettingRepositoryInterface,
	webhookRepository webhookRepository.WebhookRepositoryInterface,
	ebProvider func() event.IEventBus,
) SettingServiceInterface {
	return &SettingService{
		txManager:          tm,
		commonService:      commonService,
		keyvalueRepository: keyvalueRepository,
		webhookRepository:  webhookRepository,
		settingRepository:  settingRepository,
		eventBus:           ebProvider(),
	}
}

// GetSetting 获取设置
func (settingService *SettingService) GetSetting(setting *model.SystemSetting) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		systemSetting, err := settingService.keyvalueRepository.GetKeyValue(commonModel.SystemSettingsKey)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.SiteTitle = config.Config.Setting.SiteTitle
			setting.ServerName = config.Config.Setting.Servername
			setting.ServerURL = config.Config.Setting.Serverurl
			setting.AllowRegister = config.Config.Setting.AllowRegister
			setting.ICPNumber = config.Config.Setting.Icpnumber
			setting.MetingAPI = config.Config.Setting.MetingAPI
			setting.CustomCSS = config.Config.Setting.CustomCSS
			setting.CustomJS = config.Config.Setting.CustomJS
			setting.Logo = config.Config.Setting.Logo

			// 处理 URL
			setting.ServerURL = httpUtil.TrimURL(setting.ServerURL)
			setting.MetingAPI = httpUtil.TrimURL(setting.MetingAPI)

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.SystemSettingsKey, string(settingToJSON)); err != nil {
				return err
			}

			// 处理 ServerURL
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(systemSetting.(string)), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateSetting 更新设置
func (settingService *SettingService) UpdateSetting(userid uint, newSetting *model.SystemSettingDto) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		var setting model.SystemSetting
		setting.SiteTitle = newSetting.SiteTitle
		setting.ServerName = newSetting.ServerName
		setting.ServerURL = httpUtil.TrimURL(newSetting.ServerURL)
		setting.AllowRegister = newSetting.AllowRegister
		setting.ICPNumber = newSetting.ICPNumber
		setting.MetingAPI = httpUtil.TrimURL(newSetting.MetingAPI)
		setting.CustomCSS = newSetting.CustomCSS
		setting.CustomJS = newSetting.CustomJS
		setting.Logo = newSetting.Logo

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.SystemSettingsKey, settingToJSONString); err != nil {
			return err
		}

		// 更新 ServerURL
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
			return err
		}

		return nil
	})
}

// GetCommentSetting 获取评论设置
func (settingService *SettingService) GetCommentSetting(setting *model.CommentSetting) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		commentSetting, err := settingService.keyvalueRepository.GetKeyValue(commonModel.CommentSettingKey)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.EnableComment = config.Config.Comment.EnableComment
			setting.Provider = config.Config.Comment.Provider
			setting.CommentAPI = config.Config.Comment.CommentAPI

			// 处理 URL
			setting.CommentAPI = httpUtil.TrimURL(setting.CommentAPI)

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.CommentSettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(commentSetting.(string)), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateCommentSetting 更新评论设置
func (settingService *SettingService) UpdateCommentSetting(userid uint, newSetting *model.CommentSettingDto) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 检查评论服务提供者是否有效
		if newSetting.Provider != string(commonModel.TWIKOO) &&
			newSetting.Provider != string(commonModel.ARTALK) &&
			newSetting.Provider != string(commonModel.WALINE) &&
			newSetting.Provider != string(commonModel.GISCUS) {
			return errors.New(commonModel.NO_SUCH_COMMENT_PROVIDER)
		}

		commentSetting := &model.CommentSetting{
			EnableComment: newSetting.EnableComment,
			Provider:      newSetting.Provider,
			CommentAPI:    httpUtil.TrimURL(newSetting.CommentAPI),
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(commentSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.CommentSettingKey, string(settingToJSON)); err != nil {
			return err
		}

		return nil
	})
}

// GetS3Setting 获取 S3 存储设置
func (settingService *SettingService) GetS3Setting(userid uint, setting *model.S3Setting) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		s3Setting, err := settingService.keyvalueRepository.GetKeyValue(commonModel.S3SettingKey)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			setting.Provider = string(commonModel.MINIO)
			setting.Endpoint = ""
			setting.AccessKey = ""
			setting.SecretKey = ""
			setting.BucketName = ""
			setting.Region = ""
			setting.UseSSL = false
			setting.CDNURL = ""
			setting.PathPrefix = ""
			setting.PublicRead = true

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.S3SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(s3Setting.(string)), setting); err != nil {
			return err
		}

		// 如果用户未登录且不为管理员,则屏蔽 S3 设置的敏感信息
		if userid == authModel.NO_USER_LOGINED {
			setting.AccessKey = "******"
			setting.SecretKey = "******"
			setting.BucketName = "******"
			setting.Endpoint = "******"
		}

		return nil
	})
}

// UpdateS3Setting 更新 S3 存储设置
func (settingService *SettingService) UpdateS3Setting(userid uint, newSetting *model.S3SettingDto) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}
		// 检查endpoint是否为http(s)动态改变USE SSL
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "https://") {
			newSetting.UseSSL = true
		} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "http://") {
			newSetting.UseSSL = false
		}

		// 去除Endpoint的协议头（http://或https://）
		endpoint := strings.TrimSpace(newSetting.Endpoint)
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		newSetting.Endpoint = endpoint

		cdnURL := strings.TrimSpace(newSetting.CDNURL)
		if cdnURL != "" {
			cdnURL = strings.TrimRight(cdnURL, "/")
		}

		s3Setting := &model.S3Setting{
			Enable:     newSetting.Enable,
			Provider:   newSetting.Provider,
			Endpoint:   httpUtil.TrimURL(newSetting.Endpoint),
			AccessKey:  newSetting.AccessKey,
			SecretKey:  newSetting.SecretKey,
			BucketName: newSetting.BucketName,
			Region:     strings.TrimSpace(newSetting.Region),
			UseSSL:     newSetting.UseSSL,
			CDNURL:     cdnURL,
			PathPrefix: httpUtil.TrimURL(newSetting.PathPrefix),
			PublicRead: newSetting.PublicRead,
		}

		// 配置检查
		switch s3Setting.Provider {
		case string(commonModel.R2):
			if s3Setting.Region == "" {
				s3Setting.Region = "auto"
			}
			s3Setting.UseSSL = true
		case string(commonModel.AWS):
		case string(commonModel.MINIO):
		case string(commonModel.OTHER):
			// 其他 S3 兼容厂商（Backblaze、Wasabi、Ceph 等）
			if s3Setting.Region == "" {
				s3Setting.Region = "auto"
			}
		default:
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(s3Setting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.S3SettingKey, string(settingToJSON)); err != nil {
			return err
		}

		return nil
	})
}

// GetOAuth2Setting 获取 OAuth2 设置
func (settingService *SettingService) GetOAuth2Setting(
	userid uint,
	setting *model.OAuth2Setting,
	forInternal bool,
) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		if !forInternal {
			user, err := settingService.commonService.CommonGetUserByUserId(userid)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				return errors.New(commonModel.NO_PERMISSION_DENIED)
			}
		}

		oauthSetting, err := settingService.keyvalueRepository.GetKeyValue(commonModel.OAuth2SettingKey)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			setting.Provider = string(commonModel.OAuth2GITHUB)
			setting.ClientID = ""
			setting.ClientSecret = ""
			setting.AuthURL = "https://github.com/login/oauth/authorize"
			setting.TokenURL = "https://github.com/login/oauth/access_token"
			setting.UserInfoURL = "https://api.github.com/user"
			setting.RedirectURI = ""
			setting.Scopes = []string{
				"read:user",
			}

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(oauthSetting.(string)), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateOAuth2Setting 更新 OAuth2 设置
func (settingService *SettingService) UpdateOAuth2Setting(userid uint, newSetting *model.OAuth2SettingDto) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		oauthSetting := &model.OAuth2Setting{
			Enable:       newSetting.Enable,
			Provider:     newSetting.Provider,
			ClientID:     newSetting.ClientID,
			ClientSecret: newSetting.ClientSecret,
			AuthURL:      httpUtil.TrimURL(newSetting.AuthURL),
			TokenURL:     httpUtil.TrimURL(newSetting.TokenURL),
			UserInfoURL:  httpUtil.TrimURL(newSetting.UserInfoURL),
			RedirectURI:  httpUtil.TrimURL(newSetting.RedirectURI),
			Scopes:       newSetting.Scopes,
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(oauthSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
			return err
		}

		return nil
	})
}

// GetOAuth2Status 获取 OAuth2 状态
func (settingService *SettingService) GetOAuth2Status(status *model.OAuth2Status) error {
	var oauthSetting model.OAuth2Setting
	if err := settingService.GetOAuth2Setting(authModel.NO_USER_LOGINED, &oauthSetting, true); err != nil {
		return err
	}

	status.Enabled = oauthSetting.Enable
	status.Provider = oauthSetting.Provider

	return nil
}

// GetAllWebhooks 获取所有 Webhook
func (settingService *SettingService) GetAllWebhooks(userid uint) ([]webhookModel.Webhook, error) {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	webhooks, err := settingService.webhookRepository.GetAllWebhooks()
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

// DeleteWebhook 删除 Webhook
func (settingService *SettingService) DeleteWebhook(userid, id uint) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	settingService.txManager.Run(func(ctx context.Context) error {
		return settingService.webhookRepository.DeleteWebhookByID(ctx, id)
	})

	return nil
}

// UpdateWebhook 更新 Webhook
func (settingService *SettingService) UpdateWebhook(userid, id uint, newWebhook *model.WebhookDto) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 数据处理
	newWebhook.URL = httpUtil.TrimURL(newWebhook.URL)

	// 检查名称或URL是否为空
	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		ID:       id,
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	settingService.txManager.Run(func(ctx context.Context) error {
		// 先删除再创建，避免部分字段无法更新的问题
		if err := settingService.webhookRepository.DeleteWebhookByID(ctx, webhook.ID); err != nil {
			return err
		}
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})

	return nil
}

// CreateWebhook 创建 Webhook
func (settingService *SettingService) CreateWebhook(userid uint, newWebhook *model.WebhookDto) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 数据处理
	newWebhook.URL = httpUtil.TrimURL(newWebhook.URL)

	// 检查名称或URL是否为空
	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	settingService.txManager.Run(func(ctx context.Context) error {
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})

	return nil
}

// ListAccessTokens 列出访问令牌
func (settingService *SettingService) ListAccessTokens(userid uint) ([]model.AccessTokenSetting, error) {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	tokens, err := settingService.settingRepository.ListAccessTokens(user.ID)
	if err != nil {
		return []model.AccessTokenSetting{}, nil
	}

	// 处理tokens,过滤并删除过期的token
	var validTokens []model.AccessTokenSetting
	currentTime := time.Now()

	for _, token := range tokens {
		if token.Expiry == nil || token.Expiry.After(currentTime) {
			// nil 表示永不过期，或者还没过期
			validTokens = append(validTokens, token)
		} else {
			// 删除过期 token
			settingService.txManager.Run(func(ctx context.Context) error {
				return settingService.settingRepository.DeleteAccessTokenByID(ctx, uint(token.ID))
			})
		}
	}

	return validTokens, nil
}

// CreateAccessToken 创建访问令牌
func (settingService *SettingService) CreateAccessToken(
	userid uint,
	newToken *model.AccessTokenSettingDto,
) (string, error) {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	name := newToken.Name
	expiry := newToken.Expiry
	var expiryDuration time.Duration

	switch expiry {
	case model.EIGHT_HOUR_EXPIRY:
		expiryDuration = 8 * time.Hour
	case model.ONE_MONTH_EXPIRY:
		expiryDuration = 30 * 24 * time.Hour
	case model.NEVER_EXPIRY:
		expiryDuration = 0
	default:
		expiryDuration = 8 * time.Hour
	}

	// 生成jwt令牌
	claims := jwtUtil.CreateClaimsWithExpiry(user, int64(expiryDuration))
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	// 处理数据库存储的 expiry
	var expiryPtr *time.Time
	if expiry == model.NEVER_EXPIRY {
		expiryPtr = nil // 永不过期，用 NULL
	} else {
		t := time.Now().Add(expiryDuration)
		expiryPtr = &t
	}

	// 保存到数据库
	accessToken := &model.AccessTokenSetting{
		UserID:    user.ID,
		Token:     tokenString,
		Name:      name,
		Expiry:    expiryPtr,
		CreatedAt: time.Now(),
	}

	settingService.txManager.Run(func(ctx context.Context) error {
		return settingService.settingRepository.CreateAccessToken(ctx, accessToken)
	})

	return tokenString, nil
}

// DeleteAccessToken 删除访问令牌
func (settingService *SettingService) DeleteAccessToken(userid, id uint) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	settingService.txManager.Run(func(ctx context.Context) error {
		return settingService.settingRepository.DeleteAccessTokenByID(ctx, id)
	})

	return nil
}

// GetFediverseSetting 获取联邦网络设置
func (settingService *SettingService) GetFediverseSetting(userid uint, setting *model.FediverseSetting) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		fediverseSetting, err := settingService.keyvalueRepository.GetKeyValue(commonModel.FediverseSettingKey)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			setting.ServerURL = ""

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.FediverseSettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(fediverseSetting.(string)), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateFediverseSetting 更新联邦网络设置
func (settingService *SettingService) UpdateFediverseSetting(userid uint, newSetting *model.FediverseSettingDto) error {
	return settingService.txManager.Run(func(ctx context.Context) error {
		// 鉴权
		user, err := settingService.commonService.CommonGetUserByUserId(userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		var setting model.FediverseSetting
		setting.Enable = newSetting.Enable
		setting.ServerURL = httpUtil.TrimURL(newSetting.ServerURL)

		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.FediverseSettingKey, settingToJSONString); err != nil {
			return err
		}

		// 更新 ServerURL
		// if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
		// 	return err
		// }

		return nil
	})
}

// GetBackupScheduleSetting 获取备份计划
func (settingService *SettingService) GetBackupScheduleSetting(setting *model.BackupSchedule) error {
	// 鉴权
	// user, err := settingService.commonService.CommonGetUserByUserId(userid)
	// if err != nil {
	// 	return err
	// }
	// if !user.IsAdmin {
	// 	return errors.New(commonModel.NO_PERMISSION_DENIED)
	// }

	return settingService.txManager.Run(func(ctx context.Context) error {
		backupSchedule, err := settingService.keyvalueRepository.GetKeyValue(commonModel.BackupScheduleKey)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			// 默认每周日凌晨2点备份
			setting.CronExpression = "0 2 * * 0"

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.BackupScheduleKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(backupSchedule.(string)), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateBackupScheduleSetting 更新备份计划
func (settingService *SettingService) UpdateBackupScheduleSetting(
	userid uint,
	newSetting *model.BackupScheduleDto,
) error {
	// 鉴权
	user, err := settingService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.txManager.Run(func(ctx context.Context) error {
		var setting model.BackupSchedule
		setting.Enable = newSetting.Enable
		setting.CronExpression = newSetting.CronExpression

		// 验证 Cron 表达式是否合法
		if err := fmtUtil.ValidateCrontabExpression(setting.CronExpression); err != nil {
			return errors.New(commonModel.INVALID_CRON_EXPRESSION)
		}

		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.BackupScheduleKey, settingToJSONString); err != nil {
			return err
		}

		// 发送更新备份计划的事件
		settingService.eventBus.Publish(
			context.Background(),
			event.NewEvent(
				event.EventTypeUpdateBackupSchedule,
				event.EventPayload{
					event.EventPayloadSchedule: setting,
				},
			),
		)

		return nil
	})
}
