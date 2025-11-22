// Package service 提供用户相关的业务逻辑服务
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/event"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	model "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/user"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

// OAuthUserProfile 统一的OAuth用户信息结构
// 用于从不同OAuth提供商的原始用户数据中提取标准化的用户资料
type OAuthUserProfile struct {
	Username string // 用户名，从OAuth提供商获取或生成
	Avatar   string // 头像URL，完整的https地址
}

// UserService 用户服务结构体，提供用户相关的业务逻辑处理
type UserService struct {
	txManager      transaction.TransactionManager         // 事务管理器
	userRepository repository.UserRepositoryInterface     // 用户数据层接口
	settingService settingService.SettingServiceInterface // 系统设置数据层接口
	eventBus       event.IEventBus                        // 事件总线
}

// NewUserService 创建并返回新的用户服务实例
//
// 参数:
//   - userRepository: 用户数据层接口实现
//   - settingService: 系统设置数据层接口实现
//
// 返回:
//   - UserServiceInterface: 用户服务接口实现
func NewUserService(
	tm transaction.TransactionManager,
	userRepository repository.UserRepositoryInterface,
	settingService settingService.SettingServiceInterface,
	eventBusProvider func() event.IEventBus,
) UserServiceInterface {
	return &UserService{
		txManager:      tm,
		userRepository: userRepository,
		settingService: settingService,
		eventBus:       eventBusProvider(),
	}
}

// Login 用户登录验证
// 验证用户名和密码，成功后生成JWT token
//
// 参数:
//   - loginDto: 登录数据传输对象，包含用户名和密码
//
// 返回:
//   - string: 生成的JWT token
//   - error: 登录过程中的错误信息
func (userService *UserService) Login(loginDto *authModel.LoginDto) (string, error) {
	// 合法性校验
	if loginDto.Username == "" || loginDto.Password == "" {
		return "", errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
	}

	// 将密码进行 MD5 加密
	loginDto.Password = cryptoUtil.MD5Encrypt(loginDto.Password)

	// 检查用户是否存在
	user, err := userService.userRepository.GetUserByUsername(loginDto.Username)
	if err != nil {
		return "", errors.New(commonModel.USER_NOTFOUND)
	}

	// 进行密码验证,查看外界传入的密码是否与数据库一致
	if user.Password != loginDto.Password {
		return "", errors.New(commonModel.PASSWORD_INCORRECT)
	}

	// 生成 Token
	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		return "", err
	}

	return token, nil
}

// Register 用户注册
// 注册新用户，包括用户数量限制检查、注册权限检查等
// 第一个注册的用户自动设置为系统管理员
//
// 参数:
//   - registerDto: 注册数据传输对象，包含用户名和密码
//
// 返回:
//   - error: 注册过程中的错误信息
func (userService *UserService) Register(registerDto *authModel.RegisterDto) error {
	// 检查用户数量是否超过限制
	users, err := userService.userRepository.GetAllUsers()
	if err != nil {
		return err
	}
	if len(users) > authModel.MAX_USER_COUNT {
		return errors.New(commonModel.USER_COUNT_EXCEED_LIMIT)
	}

	// 将密码进行 MD5 加密
	registerDto.Password = cryptoUtil.MD5Encrypt(registerDto.Password)

	newUser := model.User{
		Username: registerDto.Username,
		Password: registerDto.Password,
		IsAdmin:  false,
	}

	// 检查用户是否已经存在
	user, err := userService.userRepository.GetUserByUsername(newUser.Username)
	if err == nil && user.ID != model.USER_NOT_EXISTS_ID {
		return errors.New(commonModel.USERNAME_HAS_EXISTS)
	}

	// 检查是否该系统第一次注册用户
	if len(users) == 0 {
		// 第一个注册的用户为系统管理员
		newUser.IsAdmin = true
	}

	// 检查是否开放注册
	var setting settingModel.SystemSetting
	if err := userService.settingService.GetSetting(&setting); err != nil {
		return err
	}
	if len(users) != 0 && !setting.AllowRegister {
		return errors.New(commonModel.USER_REGISTER_NOT_ALLOW)
	}
	if err := userService.txManager.Run(func(ctx context.Context) error {
		if err := userService.userRepository.CreateUser(ctx, &newUser); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// 发布用户注册事件
	newUser.Password = "" // 不包含密码信息
	userService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeUserCreated,
			event.EventPayload{
				event.EventPayloadUser: newUser,
			},
		),
	)

	return nil
}

// UpdateUser 更新用户信息
// 只有管理员可以更新用户信息，支持更新用户名、密码和头像
//
// 参数:
//   - userid: 执行更新操作的用户ID（必须为管理员）
//   - userdto: 用户信息数据传输对象，包含要更新的用户信息
//
// 返回:
//   - error: 更新过程中的错误信息
func (userService *UserService) UpdateUser(userid uint, userdto model.UserInfoDto) error {
	// 检查执行操作的用户是否为管理员
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查是否需要更新用户名
	if userdto.Username != "" && userdto.Username != user.Username {
		// 检查用户名是否已存在
		existingUser, _ := userService.userRepository.GetUserByUsername(userdto.Username)
		if existingUser.ID != model.USER_NOT_EXISTS_ID {
			return errors.New(commonModel.USERNAME_ALREADY_EXISTS)
		}
		user.Username = userdto.Username
	}

	// 检查是否需要更新密码
	if userdto.Password != "" && cryptoUtil.MD5Encrypt(userdto.Password) != user.Password {
		// 检查密码是否为空
		if userdto.Password == "" {
			return errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
		}
		// 更新密码
		user.Password = cryptoUtil.MD5Encrypt(userdto.Password)
	}

	// 检查是否需要更新头像
	if userdto.Avatar != "" && userdto.Avatar != user.Avatar {
		// 更新头像
		user.Avatar = userdto.Avatar
	}
	if err := userService.txManager.Run(func(ctx context.Context) error {
		// 更新用户信息
		return userService.userRepository.UpdateUser(ctx, &user)
	}); err != nil {
		return err
	}

	// 发布用户更新事件
	user.Password = "" // 不包含密码信息
	userService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeUserUpdated,
			event.EventPayload{
				event.EventPayloadUser: user,
			},
		),
	)

	return nil
}

// UpdateUserAdmin 更新用户的管理员权限
// 只有系统管理员、管理员可以修改其他用户的管理员权限，不能修改自己和系统管理员的权限
//
// 参数:
//   - userid: 执行操作的用户ID（必须为管理员）
//   - id: 要修改权限的用户ID
//
// 返回:
//   - error: 更新过程中的错误信息
func (userService *UserService) UpdateUserAdmin(userid uint, id uint) error {
	// 检查执行操作的用户是否为管理员
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查要修改权限的用户是否存在
	user, err = userService.userRepository.GetUserByID(int(id))
	if err != nil {
		return err
	}

	// 检查系统管理员信息
	sysadmin, err := userService.GetSysAdmin()
	if err != nil {
		return err
	}

	// 检查是否尝试修改自己或系统管理员的权限
	if userid == user.ID || id == sysadmin.ID {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	user.IsAdmin = !user.IsAdmin

	if err := userService.txManager.Run(func(ctx context.Context) error {
		// 更新用户信息
		return userService.userRepository.UpdateUser(ctx, &user)
	}); err != nil {
		return err
	}

	// 发布用户更新事件
	user.Password = "" // 不包含密码信息
	userService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeUserUpdated,
			event.EventPayload{
				event.EventPayloadUser: user,
			},
		),
	)

	return nil
}

// GetAllUsers 获取所有用户列表
// 返回除系统管理员外的所有用户，并移除密码信息
//
// 返回:
//   - []model.User: 用户列表（不包含密码信息）
//   - error: 获取过程中的错误信息
func (userService *UserService) GetAllUsers() ([]model.User, error) {
	allures, err := userService.userRepository.GetAllUsers()
	if err != nil {
		return nil, err
	}

	sysadmin, err := userService.GetSysAdmin()
	if err != nil {
		return nil, err
	}

	// 处理用户信息(去掉管理员用户)
	for i := range allures {
		if allures[i].ID == sysadmin.ID {
			allures = append(allures[:i], allures[i+1:]...)
			break
		}
	}

	// 处理用户信息(去掉密码)
	for i := range allures {
		allures[i].Password = ""
	}

	return allures, nil
}

// GetSysAdmin 获取系统管理员信息
//
// 返回:
//   - model.User: 系统管理员用户信息
//   - error: 获取过程中的错误信息
func (userService *UserService) GetSysAdmin() (model.User, error) {
	sysadmin, err := userService.userRepository.GetSysAdmin()
	if err != nil {
		return model.User{}, err
	}

	return sysadmin, nil
}

// DeleteUser 删除用户
// 只有管理员可以删除用户，不能删除自己和系统管理员
//
// 参数:
//   - userid: 执行删除操作的用户ID（必须为管理员）
//   - id: 要删除的用户ID
//
// 返回:
//   - error: 删除过程中的错误信息
func (userService *UserService) DeleteUser(userid, id uint) error {
	return userService.txManager.Run(func(ctx context.Context) error {
		// 检查执行操作的用户是否为管理员
		user, err := userService.userRepository.GetUserByID(int(userid))
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 检查要删除的用户是否存在
		user, err = userService.userRepository.GetUserByID(int(id))
		if err != nil {
			return err
		}

		sysadmin, err := userService.GetSysAdmin()
		if err != nil {
			return err
		}

		if userid == user.ID || id == sysadmin.ID {
			return errors.New(commonModel.INVALID_PARAMS_BODY)
		}

		if err := userService.userRepository.DeleteUser(ctx, id); err != nil {
			return err
		}

		return nil
	})
}

// GetUserByID 根据用户ID获取用户信息
//
// 参数:
//   - userId: 用户ID
//
// 返回:
//   - model.User: 用户信息
//   - error: 获取过程中的错误信息
func (userService *UserService) GetUserByID(userId int) (model.User, error) {
	return userService.userRepository.GetUserByID(userId)
}

// BindOAuth 为已登录用户生成OAuth账号绑定URL
// 只有管理员可以绑定OAuth账号
//
// 参数:
//   - userID: 当前用户ID
//   - provider: OAuth提供商名称
//   - redirectURI: 绑定成功后的前端回调地址
//
// 返回:
//   - string: OAuth授权URL
//   - error: 生成失败时返回错误
func (userService *UserService) BindOAuth(userID uint, provider string, redirectURI string) (string, error) {
	user, err := userService.userRepository.GetUserByID(int(userID))
	if err != nil {
		return "", err
	}

	if !user.IsAdmin {
		return "", bindingPermissionError(provider)
	}

	setting, err := userService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	state, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionBind),
		userID,
		redirectURI,
		provider,
	)
	if err != nil {
		return "", err
	}

	authorizeURL := userService.buildOAuthAuthorizeURL(setting, provider, state)
	if authorizeURL == "" {
		return "", errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return authorizeURL, nil
}

// GetOAuthLoginURL 获取OAuth登录URL
// 生成OAuth授权URL，用于用户登录
//
// 参数:
//   - provider: OAuth提供商名称
//   - redirectURI: 登录成功后的前端回调地址
//
// 返回:
//   - string: OAuth授权URL
//   - error: 生成失败时返回错误
func (userService *UserService) GetOAuthLoginURL(provider string, redirectURI string) (string, error) {
	setting, err := userService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	state, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin),
		authModel.NO_USER_LOGINED,
		redirectURI,
		provider,
	)
	if err != nil {
		return "", err
	}

	authorizeURL := userService.buildOAuthAuthorizeURL(setting, provider, state)
	if authorizeURL == "" {
		return "", errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return authorizeURL, nil
}

// validateOAuthState 验证OAuth state参数的有效性
// 检查state的签名、过期时间和provider匹配性，确保OAuth回调的安全性
//
// 参数:
//   - state: OAuth回调中的state参数（JWT格式）
//   - provider: 当前OAuth提供商名称
//
// 返回:
//   - *authModel.OAuthState: 解析后的state对象
//   - error: 验证失败时返回错误
func validateOAuthState(state, provider string) (*authModel.OAuthState, error) {
	// 解析并验证state的JWT签名
	oauthState, err := jwtUtil.ParseOAuthState(state)
	if err != nil {
		fmt.Printf("[WARN] [OAuth:%s] State签名验证失败: %v\n", provider, err)
		return nil, fmt.Errorf("state签名验证失败: %w", err)
	}

	// 检查state是否过期
	currentTime := time.Now().Unix()
	if oauthState.Exp < currentTime {
		fmt.Printf("[WARN] [OAuth:%s] State已过期\n", provider)
		return nil, errors.New("state已过期")
	}

	// 检查provider是否匹配，防止state被用于错误的OAuth提供商
	if oauthState.Provider != provider {
		fmt.Printf("[WARN] [OAuth:%s] Provider不匹配: 期望%s, 实际%s\n", provider, oauthState.Provider, provider)
		return nil, fmt.Errorf("provider不匹配: 期望%s, 实际%s", oauthState.Provider, provider)
	}

	return oauthState, nil
}

// HandleOAuthCallback 处理OAuth2回调
// 这是OAuth登录和绑定流程的核心处理函数，统一处理所有OAuth提供商的回调
//
// 处理流程:
//  1. 验证state参数（签名、过期时间、provider匹配）
//  2. 获取OAuth配置
//  3. 根据provider交换token并获取用户信息
//  4. 根据action类型执行登录或绑定操作
//
// 参数:
//   - provider: OAuth提供商名称（"github", "google", "qq", "custom"）
//   - code: OAuth授权码
//   - state: OAuth状态参数（JWT格式，包含action、userID、redirect等信息）
//
// 返回:
//   - string: 重定向URL（包含token、error或bind参数）
func (userService *UserService) HandleOAuthCallback(provider string, code string, state string) string {
	// 1. 验证state参数的有效性和安全性
	oauthState, err := validateOAuthState(state, provider)
	if err != nil {
		return buildErrorRedirect("", commonModel.QQ_OAUTH_STATE_INVALID)
	}

	// 2. 获取OAuth配置信息
	setting, err := userService.getOAuthSetting(provider)
	if err != nil {
		return buildErrorRedirect(oauthState.Redirect, "OAuth配置错误")
	}

	// 3. 处理不同OAuth提供商的token交换和用户信息获取
	externalID, userInfo, err := userService.processOAuthProvider(provider, setting, code)
	if err != nil {
		return buildErrorRedirect(oauthState.Redirect, err.Error())
	}

	// 4. 根据action类型处理回调逻辑（登录或绑定）
	return userService.resolveOAuthCallback(oauthState, provider, externalID, userInfo)
}

// processOAuthProvider 处理不同OAuth提供商的token交换和用户信息获取
// 统一各OAuth提供商的处理流程，返回外部用户ID和用户信息
//
// 参数:
//   - provider: OAuth提供商名称
//   - setting: OAuth配置信息
//   - code: OAuth授权码
//
// 返回:
//   - externalID: 第三方平台的用户唯一标识
//   - userInfo: 用户信息（不同provider类型不同）
//   - error: 处理过程中的错误
func (userService *UserService) processOAuthProvider(
	provider string,
	setting *settingModel.OAuth2Setting,
	code string,
) (externalID string, userInfo interface{}, err error) {
	switch provider {
	case string(commonModel.OAuth2GITHUB):
		return userService.processGitHubOAuth(setting, code)

	case string(commonModel.OAuth2GOOGLE):
		return userService.processGoogleOAuth(setting, code)

	case string(commonModel.OAuth2QQ):
		return userService.processQQOAuth(setting, code)

	case string(commonModel.OAuth2CUSTOM):
		return userService.processCustomOAuth(setting, code)

	default:
		return "", nil, errors.New("不支持的OAuth提供商")
	}
}

// processGitHubOAuth 处理GitHub OAuth流程
func (userService *UserService) processGitHubOAuth(
	setting *settingModel.OAuth2Setting,
	code string,
) (string, interface{}, error) {
	tokenResp, err := exchangeGithubCodeForToken(setting, code)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:GitHub] Token交换失败: %v\n", err)
		return "", nil, err
	}

	githubUser, err := fetchGitHubUserInfo(setting, tokenResp.AccessToken)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:GitHub] 获取用户信息失败: %v\n", err)
		return "", nil, err
	}

	return fmt.Sprint(githubUser.ID), githubUser, nil
}

// processGoogleOAuth 处理Google OAuth流程
func (userService *UserService) processGoogleOAuth(
	setting *settingModel.OAuth2Setting,
	code string,
) (string, interface{}, error) {
	tokenResp, err := exchangeGoogleCodeForToken(setting, code)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:Google] Token交换失败: %v\n", err)
		return "", nil, err
	}

	googleUser, err := fetchGoogleUserInfo(setting, tokenResp.AccessToken)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:Google] 获取用户信息失败: %v\n", err)
		return "", nil, err
	}

	return googleUser.Sub, googleUser, nil
}

// processQQOAuth 处理QQ OAuth流程
func (userService *UserService) processQQOAuth(
	setting *settingModel.OAuth2Setting,
	code string,
) (string, interface{}, error) {
	tokenResp, err := exchangeQQCodeForToken(setting, code)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:QQ] Token交换失败: %v\n", err)
		return "", nil, err
	}

	openIDResp, err := fetchQQOpenID(setting, tokenResp.AccessToken)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:QQ] 获取OpenID失败: %v\n", err)
		return "", nil, err
	}

	// 获取QQ用户信息（可选，失败时使用空对象）
	qqUserInfo, err := fetchQQUserInfo(setting, tokenResp.AccessToken, openIDResp.OpenID)
	if err != nil {
		qqUserInfo = &authModel.QQUser{}
	}

	return openIDResp.OpenID, qqUserInfo, nil
}

// processCustomOAuth 处理自定义OAuth流程
func (userService *UserService) processCustomOAuth(
	setting *settingModel.OAuth2Setting,
	code string,
) (string, interface{}, error) {
	accessToken, err := exchangeCustomCodeForToken(setting, code)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:Custom] Token交换失败: %v\n", err)
		return "", nil, err
	}

	customUserID, err := fetchCustomUserInfo(setting, accessToken)
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:Custom] 获取用户信息失败: %v\n", err)
		return "", nil, err
	}

	return customUserID, nil, nil
}

// getOAuthSetting 获取并验证OAuth配置
// 检查OAuth配置是否完整且已启用
//
// 参数:
//   - provider: OAuth提供商名称
//
// 返回:
//   - *settingModel.OAuth2Setting: OAuth配置信息
//   - error: 配置不存在、未启用或不完整时返回错误
func (userService *UserService) getOAuthSetting(provider string) (*settingModel.OAuth2Setting, error) {
	// 获取OAuth配置
	var setting settingModel.OAuth2Setting
	if err := userService.settingService.GetOAuth2Setting(0, &setting, true); err != nil {
		return nil, err
	}

	// 验证provider是否匹配
	if setting.Provider != provider {
		return nil, errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	// 检查是否已启用
	if !setting.Enable {
		return nil, errors.New(commonModel.OAUTH2_NOT_ENABLED)
	}

	// 验证必需字段是否完整
	if setting.ClientID == "" || setting.RedirectURI == "" || setting.AuthURL == "" || setting.TokenURL == "" ||
		setting.UserInfoURL == "" || setting.ClientSecret == "" {
		return nil, errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return &setting, nil
}

// buildOAuthAuthorizeURL 构建OAuth授权URL
// 根据不同的OAuth提供商构建相应的授权请求URL
//
// 参数:
//   - setting: OAuth2配置信息
//   - provider: OAuth提供商名称
//   - state: OAuth state参数（用于防止CSRF攻击）
//
// 返回:
//   - string: 授权URL（如果provider不支持，返回空字符串）
func (userService *UserService) buildOAuthAuthorizeURL(
	setting *settingModel.OAuth2Setting,
	provider, state string,
) string {
	// 处理scope参数
	scope := ""
	if len(setting.Scopes) > 0 {
		scope = strings.Join(setting.Scopes, " ")
	}

	switch provider {
	case string(commonModel.OAuth2GITHUB):
		// GitHub OAuth授权URL
		return fmt.Sprintf(
			"%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
			setting.AuthURL,
			url.QueryEscape(setting.ClientID),
			url.QueryEscape(setting.RedirectURI),
			url.QueryEscape(scope),
			url.QueryEscape(state),
		)

	case string(commonModel.OAuth2GOOGLE):
		// Google OAuth授权URL
		params := url.Values{}
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("response_type", "code")
		params.Set("state", state)
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
		if scope != "" {
			params.Set("scope", scope)
		}
		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())

	case string(commonModel.OAuth2QQ):
		// QQ互联OAuth授权URL
		// QQ互联使用固定的scope: get_user_info
		params := url.Values{}
		params.Set("response_type", "code")
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("state", state)
		params.Set("scope", "get_user_info")
		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())

	case string(commonModel.OAuth2CUSTOM):
		// 自定义OAuth授权URL
		params := url.Values{}
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("response_type", "code")
		params.Set("state", state)
		if scope != "" {
			params.Set("scope", scope)
		}
		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())

	default:
		return ""
	}
}

func bindingPermissionError(provider string) error {
	switch provider {
	case string(commonModel.OAuth2GITHUB):
		return errors.New(commonModel.NO_PERMISSION_BINDING_GITHUB)
	case string(commonModel.OAuth2GOOGLE):
		return errors.New(commonModel.NO_PERMISSION_BINDING_GOOGLE)
	case string(commonModel.OAuth2QQ):
		return errors.New(commonModel.NO_PERMISSION_BINDING_QQ)
	case string(commonModel.OAuth2CUSTOM):
		return errors.New(commonModel.NO_PERMISSION_BINDING_CUSTOM)
	default:
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
}

// buildSuccessRedirect 构建包含token的成功重定向URL
// 在OAuth登录成功时使用，将JWT token添加到重定向URL的查询参数中
//
// 参数:
//   - redirectURL: 重定向目标URL
//   - token: JWT token
//
// 返回:
//   - string: 包含token参数的重定向URL（如果redirectURL为空或解析失败，返回空字符串）
func buildSuccessRedirect(redirectURL string, token string) string {
	// 解析URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return ""
	}

	// 添加token参数
	query := parsedURL.Query()
	query.Set("token", token)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

// buildErrorRedirect 构建包含错误信息的重定向URL
// 在OAuth回调失败时使用，将错误信息添加到重定向URL的查询参数中
//
// 参数:
//   - redirectURL: 重定向目标URL
//   - errorMsg: 错误消息
//
// 返回:
//   - string: 包含error参数的重定向URL（如果redirectURL为空或解析失败，返回空字符串）
func buildErrorRedirect(redirectURL string, errorMsg string) string {
	// 如果重定向URL为空，返回空字符串
	if redirectURL == "" {
		return ""
	}

	// 解析URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return ""
	}

	// 添加error参数
	query := parsedURL.Query()
	query.Set("error", errorMsg)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

// extractUserProfile 从不同OAuth提供商的用户信息中提取统一的用户资料
// 该函数处理各OAuth提供商返回的不同格式的用户数据，提取出用户名和头像URL
//
// 参数:
//   - provider: OAuth提供商名称（"github", "google", "qq", "custom"）
//   - userInfo: 原始用户信息（interface{}类型，需要类型断言）
//   - externalID: 外部用户ID，用于生成默认用户名
//
// 返回:
//   - OAuthUserProfile: 提取后的用户资料，包含用户名和头像URL
func extractUserProfile(provider string, userInfo interface{}, externalID string) OAuthUserProfile {
	// 默认用户资料，使用provider和随机字符串生成默认用户名
	profile := OAuthUserProfile{
		Username: fmt.Sprintf("%s_user_%s", provider, cryptoUtil.GenerateRandomString(6)),
		Avatar:   "",
	}

	switch provider {
	case string(commonModel.OAuth2GITHUB):
		// GitHub用户信息提取
		if githubUser, ok := userInfo.(*authModel.GitHubUser); ok && githubUser != nil {
			if githubUser.Login != "" {
				profile.Username = githubUser.Login
			}
			profile.Avatar = githubUser.AvatarURL
		}

	case string(commonModel.OAuth2GOOGLE):
		// Google用户信息提取
		if googleUser, ok := userInfo.(*authModel.GoogleUser); ok && googleUser != nil {
			if googleUser.Name != "" {
				profile.Username = googleUser.Name
			} else if googleUser.Email != "" {
				// 如果没有name，使用邮箱前缀作为用户名
				profile.Username = strings.Split(googleUser.Email, "@")[0]
			}
			profile.Avatar = googleUser.Picture
		}

	case string(commonModel.OAuth2QQ):
		// QQ用户信息提取
		if qqUser, ok := userInfo.(*authModel.QQUser); ok && qqUser != nil {
			if qqUser.Nickname != "" {
				profile.Username = qqUser.Nickname
			}
			// QQ头像优先级: FigureURLQQ2(100x100) > FigureURLQQ1(40x40) > FigureURL2(100x100) > FigureURL1(50x50)
			// 优先选择最高质量的头像
			if qqUser.FigureURLQQ2 != "" {
				profile.Avatar = qqUser.FigureURLQQ2
			} else if qqUser.FigureURLQQ1 != "" {
				profile.Avatar = qqUser.FigureURLQQ1
			} else if qqUser.FigureURL2 != "" {
				profile.Avatar = qqUser.FigureURL2
			} else if qqUser.FigureURL1 != "" {
				profile.Avatar = qqUser.FigureURL1
			}
		}

	case string(commonModel.OAuth2CUSTOM):
		// 自定义OAuth用户信息提取
		// 尝试从多个可能的字段名中提取用户名和头像
		if customData, ok := userInfo.(map[string]interface{}); ok {
			// 尝试提取用户名，按优先级尝试多个字段
			for _, key := range []string{"name", "username", "nickname", "display_name"} {
				if val, exists := customData[key]; exists {
					if name := fmt.Sprint(val); name != "" && name != "<nil>" {
						profile.Username = name
						break
					}
				}
			}
			// 尝试提取头像，按优先级尝试多个字段
			for _, key := range []string{"avatar", "avatar_url", "picture", "photo"} {
				if val, exists := customData[key]; exists {
					if avatar := fmt.Sprint(val); avatar != "" && avatar != "<nil>" {
						profile.Avatar = avatar
						break
					}
				}
			}
		}
	}

	return profile
}

// resolveUsernameConflict 解决用户名冲突
// 检查用户名是否已存在，如果存在则添加随机后缀
//
// 参数:
//   - username: 原始用户名
//
// 返回:
//   - string: 可用的用户名（可能添加了随机后缀）
func (userService *UserService) resolveUsernameConflict(username string) string {
	existingUser, _ := userService.userRepository.GetUserByUsername(username)
	if existingUser.ID != model.USER_NOT_EXISTS_ID {
		// 用户名冲突，添加随机后缀
		return fmt.Sprintf("%s_%s", username, cryptoUtil.GenerateRandomString(6))
	}
	return username
}

// createOAuthUser 创建OAuth用户
// 从OAuth提供商的用户信息中提取资料，创建新用户并绑定OAuth账号
//
// 参数:
//   - provider: OAuth提供商名称
//   - externalID: 第三方平台的用户唯一标识
//   - userInfo: 用户信息
//
// 返回:
//   - model.User: 创建的用户信息（失败时返回空User，ID为0）
func (userService *UserService) createOAuthUser(
	provider, externalID string,
	userInfo interface{},
) model.User {
	// 提取用户信息
	profile := extractUserProfile(provider, userInfo, externalID)

	// 处理用户名冲突
	username := userService.resolveUsernameConflict(profile.Username)

	// 创建新用户
	newUser := model.User{
		Username: username,
		Password: cryptoUtil.MD5Encrypt(cryptoUtil.GenerateRandomString(32)), // 随机密码
		IsAdmin:  false,
		Avatar:   profile.Avatar,
	}

	// 在事务中创建用户并绑定OAuth
	err := userService.txManager.Run(func(ctx context.Context) error {
		// 创建用户
		if err := userService.userRepository.CreateUser(ctx, &newUser); err != nil {
			fmt.Printf("[ERROR] [OAuth:%s] 创建用户失败: %v\n", provider, err)
			return err
		}
		fmt.Printf("[INFO] [OAuth:%s] 用户创建成功 (userID=%d), 开始绑定OAuth (externalID=%s)\n",
			provider, newUser.ID, externalID)

		// 绑定OAuth
		if err := userService.userRepository.BindOAuth(ctx, newUser.ID, provider, externalID); err != nil {
			fmt.Printf("[ERROR] [OAuth:%s] 绑定OAuth失败: %v\n", provider, err)
			return err
		}
		fmt.Printf("[INFO] [OAuth:%s] OAuth绑定成功 (userID=%d, provider=%s, externalID=%s)\n",
			provider, newUser.ID, provider, externalID)

		return nil
	})

	if err != nil {
		fmt.Printf("[ERROR] [OAuth:%s] 创建用户并绑定失败: %v\n", provider, err)
		return model.User{}
	}

	return newUser
}

// handleOAuthBind 处理OAuth绑定逻辑
// 将OAuth账号绑定到已登录的用户，检查OAuth ID是否已被其他用户绑定
//
// 参数:
//   - oauthState: OAuth状态信息
//   - provider: OAuth提供商名称
//   - externalID: 第三方平台的用户唯一标识
//
// 返回:
//   - string: 重定向URL（包含bind=success或bind=error参数）
func (userService *UserService) handleOAuthBind(
	oauthState *authModel.OAuthState,
	provider, externalID string,
) string {
	// 检查OAuth ID是否已被其他用户绑定
	existingUser, err := userService.userRepository.GetUserByOAuthID(
		context.Background(),
		provider,
		externalID,
	)
	if err == nil && existingUser.ID != model.USER_NOT_EXISTS_ID && existingUser.ID != oauthState.UserID {
		// OAuth ID已被其他用户绑定
		errorMsg := ""
		if provider == string(commonModel.OAuth2QQ) {
			errorMsg = commonModel.QQ_OAUTH_ALREADY_BOUND
		} else {
			errorMsg = "该账号已被其他用户绑定"
		}

		redirectURL, parseErr := url.Parse(oauthState.Redirect)
		if parseErr != nil {
			return ""
		}
		query := redirectURL.Query()
		query.Set("bind", "error")
		query.Set("error", errorMsg)
		redirectURL.RawQuery = query.Encode()
		return redirectURL.String()
	}

	// 执行绑定操作
	if err := userService.txManager.Run(func(ctx context.Context) error {
		return userService.userRepository.BindOAuth(ctx, oauthState.UserID, provider, externalID)
	}); err != nil {
		fmt.Printf("[ERROR] [OAuth:%s] 绑定失败: %v\n", provider, err)
		redirectURL, parseErr := url.Parse(oauthState.Redirect)
		if parseErr != nil {
			return ""
		}
		query := redirectURL.Query()
		query.Set("bind", "error")
		query.Set("error", "绑定失败")
		redirectURL.RawQuery = query.Encode()
		return redirectURL.String()
	}

	return oauthState.Redirect + "?bind=success"
}

// handleOAuthLogin 处理OAuth登录逻辑
// 查询用户是否已绑定OAuth账号，如果未绑定则创建新用户
//
// 参数:
//   - oauthState: OAuth状态信息
//   - provider: OAuth提供商名称
//   - externalID: 第三方平台的用户唯一标识
//   - userInfo: 用户信息
//
// 返回:
//   - string: 重定向URL（包含token或error参数）
func (userService *UserService) handleOAuthLogin(
	oauthState *authModel.OAuthState,
	provider, externalID string,
	userInfo interface{},
) string {
	// 查询是否已存在OAuth绑定
	user, err := userService.userRepository.GetUserByOAuthID(
		context.Background(),
		provider,
		externalID,
	)

	if err != nil {
		// 记录查询失败的详细信息，帮助诊断问题
		fmt.Printf("[INFO] [OAuth:%s] 未找到已绑定用户 (provider=%s, externalID=%s, error=%v)，准备创建新用户\n",
			provider, provider, externalID, err)

		// 用户不存在，创建新用户
		user = userService.createOAuthUser(provider, externalID, userInfo)
		if user.ID == 0 {
			fmt.Printf("[ERROR] [OAuth:%s] 创建用户失败\n", provider)
			return buildErrorRedirect(oauthState.Redirect, "创建用户失败")
		}

		fmt.Printf("[INFO] [OAuth:%s] 成功创建新用户 (userID=%d, username=%s)\n",
			provider, user.ID, user.Username)
	} else {
		// 找到已存在的用户
		fmt.Printf("[INFO] [OAuth:%s] 找到已绑定用户 (userID=%d, username=%s)\n",
			provider, user.ID, user.Username)
	}

	// 生成JWT token
	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		fmt.Printf("[ERROR] [OAuth:%s] 生成token失败: %v\n", provider, err)
		return buildErrorRedirect(oauthState.Redirect, "生成token失败")
	}

	// 构建成功重定向URL
	return buildSuccessRedirect(oauthState.Redirect, token)
}

// resolveOAuthCallback 根据action类型分发OAuth回调处理
// 将OAuth回调分发到登录或绑定处理函数
//
// 参数:
//   - oauthState: OAuth状态信息
//   - provider: OAuth提供商名称
//   - externalID: 第三方平台的用户唯一标识
//   - userInfo: 用户信息
//
// 返回:
//   - string: 重定向URL
func (userService *UserService) resolveOAuthCallback(
	oauthState *authModel.OAuthState,
	provider, externalID string,
	userInfo interface{},
) string {
	switch oauthState.Action {
	case string(authModel.OAuth2ActionLogin):
		// 登录操作：userID必须为0（未登录状态）
		if oauthState.UserID != authModel.NO_USER_LOGINED {
			return ""
		}
		return userService.handleOAuthLogin(oauthState, provider, externalID, userInfo)

	case string(authModel.OAuth2ActionBind):
		// 绑定操作：userID必须不为0（已登录状态）
		if oauthState.UserID == authModel.NO_USER_LOGINED {
			return ""
		}
		return userService.handleOAuthBind(oauthState, provider, externalID)

	default:
		return ""
	}
}

// exchangeGithubCodeForToken 用授权码换取GitHub访问令牌
// GitHub使用JSON格式的POST请求交换token
//
// 参数:
//   - setting: OAuth2配置信息
//   - code: OAuth授权码
//
// 返回:
//   - *authModel.GitHubTokenResponse: GitHub token响应
//   - error: 交换失败时返回错误
func exchangeGithubCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.GitHubTokenResponse, error) {
	// 构建请求数据
	data := map[string]string{
		"client_id":     setting.ClientID,
		"client_secret": setting.ClientSecret,
		"code":          code,
		"redirect_uri":  setting.RedirectURI,
	}
	jsonData, _ := json.Marshal(data)

	// 创建POST请求
	req, _ := http.NewRequest("POST", setting.TokenURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("GitHub token 响应错误: " + string(body))
	}

	// 解析JSON响应
	var tokenResp authModel.GitHubTokenResponse
	_ = json.Unmarshal(body, &tokenResp)
	return &tokenResp, nil
}

// fetchGitHubUserInfo 获取GitHub用户信息
// 使用Bearer token认证获取用户的公开信息
//
// 参数:
//   - setting: OAuth2配置信息
//   - accessToken: 访问令牌
//
// 返回:
//   - *authModel.GitHubUser: GitHub用户信息
//   - error: 获取失败时返回错误
func fetchGitHubUserInfo(setting *settingModel.OAuth2Setting, accessToken string) (*authModel.GitHubUser, error) {
	// 创建GET请求
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("GitHub 用户信息请求失败: " + string(body))
	}

	// 解析JSON响应
	var user authModel.GitHubUser
	_ = json.Unmarshal(body, &user)
	return &user, nil
}

// exchangeGoogleCodeForToken 用授权码换取Google访问令牌
// Google使用URL编码格式的POST请求交换token
//
// 参数:
//   - setting: OAuth2配置信息
//   - code: OAuth授权码
//
// 返回:
//   - *authModel.GoogleTokenResponse: Google token响应
//   - error: 交换失败时返回错误
func exchangeGoogleCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.GoogleTokenResponse, error) {
	// 构建请求数据（URL编码格式）
	data := url.Values{}
	data.Set("client_id", setting.ClientID)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", setting.RedirectURI)
	data.Set("grant_type", "authorization_code")

	// 创建POST请求
	req, _ := http.NewRequest("POST", setting.TokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Google token 响应错误: " + string(body))
	}

	// 解析JSON响应
	var tokenResp authModel.GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// fetchGoogleUserInfo 获取Google用户信息
// 使用Bearer token认证获取用户的公开信息
//
// 参数:
//   - setting: OAuth2配置信息
//   - accessToken: 访问令牌
//
// 返回:
//   - *authModel.GoogleUser: Google用户信息
//   - error: 获取失败时返回错误
func fetchGoogleUserInfo(setting *settingModel.OAuth2Setting, accessToken string) (*authModel.GoogleUser, error) {
	// 创建GET请求
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Google 用户信息请求失败: " + string(body))
	}

	// 解析JSON响应
	var user authModel.GoogleUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// exchangeQQCodeForToken 用授权码换取QQ访问令牌
// QQ互联的token响应可能是JSON或URL编码格式，需要兼容处理
//
// 参数:
//   - setting: OAuth2配置信息
//   - code: OAuth授权码
//
// 返回:
//   - *authModel.QQTokenResponse: QQ token响应
//   - error: 交换失败时返回错误
func exchangeQQCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.QQTokenResponse, error) {
	// 构建请求参数
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("client_id", setting.ClientID)
	params.Set("client_secret", setting.ClientSecret)
	params.Set("code", code)
	params.Set("redirect_uri", setting.RedirectURI)

	tokenURL := fmt.Sprintf("%s?%s", setting.TokenURL, params.Encode())

	// 创建HTTP客户端，设置30秒超时
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return nil, errors.New(commonModel.QQ_TOKEN_EXCHANGE_FAILED)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "timeout") {
			return nil, errors.New("QQ登录请求超时，请稍后重试")
		}
		return nil, errors.New(commonModel.QQ_TOKEN_EXCHANGE_FAILED)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(commonModel.QQ_TOKEN_EXCHANGE_FAILED)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(commonModel.QQ_TOKEN_EXCHANGE_FAILED)
	}

	// 尝试解析JSON格式响应
	var tokenResp authModel.QQTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		// JSON解析失败，尝试URL编码格式
		vals, parseErr := url.ParseQuery(string(body))
		if parseErr != nil {
			return nil, errors.New(commonModel.QQ_TOKEN_EXCHANGE_FAILED)
		}

		tokenResp.AccessToken = vals.Get("access_token")
		if expiresIn := vals.Get("expires_in"); expiresIn != "" {
			fmt.Sscanf(expiresIn, "%d", &tokenResp.ExpiresIn)
		}
		tokenResp.RefreshToken = vals.Get("refresh_token")
	}

	// 验证access_token是否存在
	if tokenResp.AccessToken == "" {
		return nil, errors.New(commonModel.QQ_TOKEN_EXCHANGE_FAILED)
	}

	return &tokenResp, nil
}

// fetchQQOpenID 获取QQ用户的OpenID
// OpenID是QQ用户在当前应用的唯一标识，需要单独调用API获取
// QQ互联的响应可能包含JSONP格式的callback包装，需要去除
//
// 参数:
//   - setting: OAuth2配置信息（未使用，保留以保持函数签名一致）
//   - accessToken: 访问令牌
//
// 返回:
//   - *authModel.QQOpenIDResponse: OpenID响应
//   - error: 获取失败时返回错误
func fetchQQOpenID(
	setting *settingModel.OAuth2Setting,
	accessToken string,
) (*authModel.QQOpenIDResponse, error) {
	// 构建请求参数
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("fmt", "json") // 请求JSON格式响应

	openIDURL := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?%s", params.Encode())

	// 创建HTTP客户端，设置30秒超时
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", openIDURL, nil)
	if err != nil {
		return nil, errors.New(commonModel.QQ_OPENID_FETCH_FAILED)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "timeout") {
			return nil, errors.New("获取QQ用户标识超时，请稍后重试")
		}
		return nil, errors.New(commonModel.QQ_OPENID_FETCH_FAILED)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(commonModel.QQ_OPENID_FETCH_FAILED)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(commonModel.QQ_OPENID_FETCH_FAILED)
	}

	// 去除JSONP格式的callback包装（如果存在）
	bodyStr := string(body)
	if strings.HasPrefix(bodyStr, "callback(") && strings.HasSuffix(bodyStr, ");") {
		bodyStr = strings.TrimPrefix(bodyStr, "callback(")
		bodyStr = strings.TrimSuffix(bodyStr, ");")
		bodyStr = strings.TrimSpace(bodyStr)
	}

	// 解析JSON响应
	var openIDResp authModel.QQOpenIDResponse
	if err := json.Unmarshal([]byte(bodyStr), &openIDResp); err != nil {
		return nil, errors.New(commonModel.QQ_OPENID_FETCH_FAILED)
	}

	// 验证OpenID是否存在
	if openIDResp.OpenID == "" {
		return nil, errors.New(commonModel.QQ_OPENID_FETCH_FAILED)
	}

	return &openIDResp, nil
}

// fetchQQUserInfo 获取QQ用户信息
// 获取用户的昵称、头像等公开信息，此接口调用失败不影响登录流程
// 失败时返回空的QQUser对象，使用默认用户名和空头像
//
// 参数:
//   - setting: OAuth2配置信息
//   - accessToken: 访问令牌
//   - openID: QQ用户的OpenID
//
// 返回:
//   - *authModel.QQUser: QQ用户信息（失败时返回空对象）
//   - error: 获取失败时返回错误（但不影响登录流程）
func fetchQQUserInfo(
	setting *settingModel.OAuth2Setting,
	accessToken string,
	openID string,
) (*authModel.QQUser, error) {
	// 构建请求参数
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("oauth_consumer_key", setting.ClientID)
	params.Set("openid", openID)

	userInfoURL := fmt.Sprintf("https://graph.qq.com/user/get_user_info?%s", params.Encode())

	// 创建HTTP客户端，设置30秒超时
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return &authModel.QQUser{}, nil
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return &authModel.QQUser{}, nil
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &authModel.QQUser{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &authModel.QQUser{}, nil
	}

	// 解析JSON响应
	var userInfo authModel.QQUser
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return &authModel.QQUser{}, nil
	}

	// 检查QQ互联API的返回码（0表示成功）
	if userInfo.Ret != 0 {
		return &authModel.QQUser{}, nil
	}

	return &userInfo, nil
}

// exchangeCustomCodeForToken 用授权码换取自定义OAuth访问令牌
// 自定义OAuth使用URL编码格式的POST请求交换token
//
// 参数:
//   - setting: OAuth2配置信息
//   - code: OAuth授权码
//
// 返回:
//   - string: 访问令牌
//   - error: 交换失败时返回错误
func exchangeCustomCodeForToken(setting *settingModel.OAuth2Setting, code string) (string, error) {
	// 构建请求数据（URL编码格式）
	data := url.Values{}
	data.Set("client_id", setting.ClientID)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", setting.RedirectURI)
	data.Set("grant_type", "authorization_code")

	// 创建POST请求
	req, _ := http.NewRequest("POST", setting.TokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", errors.New("Custom token 响应错误: " + string(body))
	}

	// 解析JSON响应
	var tokenResp map[string]any
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	// 提取access_token
	if accessToken, ok := tokenResp["access_token"]; ok {
		if tokenStr := fmt.Sprint(accessToken); tokenStr != "" && tokenStr != "<nil>" {
			return tokenStr, nil
		}
	}

	return "", errors.New("Custom token 响应缺少 access_token")
}

// fetchCustomUserInfo 获取自定义OAuth用户信息
// 尝试从多个可能的字段中提取用户唯一标识
//
// 参数:
//   - setting: OAuth2配置信息
//   - accessToken: 访问令牌
//
// 返回:
//   - string: 用户唯一标识
//   - error: 获取失败时返回错误
func fetchCustomUserInfo(setting *settingModel.OAuth2Setting, accessToken string) (string, error) {
	// 创建GET请求
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Custom 用户信息请求失败: " + string(body))
	}

	// 解析JSON响应
	var userData map[string]any
	if err := json.Unmarshal(body, &userData); err != nil {
		return "", err
	}

	// 尝试从多个可能的字段中提取用户ID
	for _, key := range []string{"id", "sub", "user_id", "uid", "openid"} {
		if val, ok := userData[key]; ok {
			if id := fmt.Sprint(val); id != "" && id != "<nil>" {
				return id, nil
			}
		}
	}

	return "", errors.New("Custom 用户信息缺少唯一标识字段 (id/sub/user_id/uid)")
}

// GetOAuthInfo 获取用户的OAuth绑定信息
// 只有管理员可以查看OAuth绑定信息
//
// 参数:
//   - userId: 用户ID
//   - provider: OAuth提供商名称
//
// 返回:
//   - model.OAuthInfoDto: OAuth绑定信息
//   - error: 获取失败时返回错误
func (userService *UserService) GetOAuthInfo(userId uint, provider string) (model.OAuthInfoDto, error) {
	var oauthInfo model.OAuthInfoDto

	// 检查当前用户是否存在
	user, err := userService.userRepository.GetUserByID(int(userId))
	if err != nil {
		return oauthInfo, err
	}

	// 检查用户是否为管理员
	if !user.IsAdmin {
		return oauthInfo, bindingPermissionError(provider)
	}

	oauthInfoBinding, err := userService.userRepository.GetOAuthInfo(userId, provider)
	if err != nil {
		return oauthInfo, err
	}

	oauthInfo = model.OAuthInfoDto{
		Provider: oauthInfoBinding.Provider,
		UserID:   oauthInfoBinding.UserID,
		OAuthID:  oauthInfoBinding.OAuthID,
	}

	return oauthInfo, nil
}
