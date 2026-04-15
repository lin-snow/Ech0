package model

import "fmt"

// ServerError 定义服务器错误信息
type ServerError struct {
	Msg string
	Err error
}

type BizError struct {
	Code       string
	Msg        string
	MessageKey string
	Params     map[string]any
	Err        error
}

func (e *BizError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func NewBizError(code, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

func NewBizErrorWithMessageKey(code, msg, messageKey string, params map[string]any) *BizError {
	return &BizError{
		Code:       code,
		Msg:        msg,
		MessageKey: messageKey,
		Params:     params,
	}
}

// 失败相关的常量
const (
	INVALID_FILE_PATH      = "无效的文件路径"
	INVALID_REQUEST_BODY   = "无效的请求体"
	INVALID_PARAMS_BODY    = "无效参数"
	INVALID_QUERY_PARAMS   = "无效的查询参数"
	INVALID_REQUEST_METHOD = "无效的请求方法"
)

// 业务错误码
const (
	ErrCodeInvalidRequest          = "INVALID_REQUEST"
	ErrCodePermissionDenied        = "PERMISSION_DENIED"
	ErrCodeInitAlreadyDone         = "INIT_ALREADY_DONE"
	ErrCodeInitOwnerExists         = "INIT_OWNER_EXISTS"
	ErrCodeInitInvalidState        = "INIT_INVALID_STATE"
	ErrCodeInvalidQuery            = "INVALID_QUERY"
	ErrCodeTokenMissing            = "TOKEN_MISSING"
	ErrCodeTokenInvalid            = "TOKEN_INVALID"
	ErrCodeTokenParse              = "TOKEN_PARSE_ERROR"
	ErrCodeScopeForbidden          = "SCOPE_FORBIDDEN"
	ErrCodeAudienceForbidden       = "AUDIENCE_FORBIDDEN"
	ErrCodeTokenTransportForbidden = "TOKEN_TRANSPORT_FORBIDDEN"
	ErrCodeTokenRevoked            = "TOKEN_REVOKED"
	ErrCodeRefreshTokenInvalid     = "REFRESH_TOKEN_INVALID"
	ErrCodeExchangeCodeInvalid     = "EXCHANGE_CODE_INVALID"
	ErrCodeTokenGenerateFailed     = "TOKEN_GENERATE_FAILED"
)

// Auth 错误相关常量
const (
	USERNAME_OR_PASSWORD_NOT_BE_EMPTY = "用户名或密码不能为空"
	PASSWORD_INCORRECT                = "密码错误"
	USER_NOTFOUND                     = "用户不存在"
	USER_COUNT_EXCEED_LIMIT           = "用户数量超过限制"
	USERNAME_HAS_EXISTS               = "用户名已存在"
	TOKEN_NOT_FOUND                   = "未找到令牌,请点击右上角登录"
	TOKEN_NOT_VALID                   = "令牌无效，请重新登录"
	TOKEN_PARSE_ERROR                 = "令牌解析失败，请尝试重新登陆"
	TOKEN_REVOKED                     = "令牌已被吊销，请重新登录"
	REFRESH_TOKEN_INVALID             = "刷新令牌无效或已过期"
	EXCHANGE_CODE_INVALID             = "授权码无效或已过期"
	TOKEN_GENERATE_FAILED             = "令牌生成失败"
	USER_REGISTER_NOT_ALLOW           = "当前系统禁止注册新用户"
)

// Echo 错误相关常量
const (
	NO_PERMISSION_DENIED  = "没有权限,请联系系统管理员"
	ECHO_CAN_NOT_BE_EMPTY = "ECHO 内容不能为空"
	ECHO_NOT_FOUND        = "找不到Echo"
)

// Common 错误相关常量
const (
	NO_FILE_UPLOAD_ERROR   = "找不到上传的文件"
	NO_FILE_STORAGE_ERROR  = "未知存储方式"
	FILE_TYPE_NOT_ALLOWED  = "不支持的文件类型"
	FILE_SIZE_EXCEED_LIMIT = "文件大小超过限制"
	IMAGE_NOT_FOUND        = "图片未找到"
	INVALID_PARAMS         = "错误的参数"
	SIGNUP_FIRST           = "请先初始化Owner账号"
	S3_NOT_ENABLED         = "S3存储未启用"
	S3_NOT_CONFIGURED      = "S3存储未配置"
	S3_CONFIG_ERROR        = "S3存储配置错误"
	SYSTEM_ALREADY_INITED  = "系统已初始化"
	OWNER_ALREADY_EXISTS   = "Owner已存在"
	ONLY_OWNER_CAN_MANAGE  = "仅Owner可管理管理员权限"
)

// User 错误相关常量
const (
	USERNAME_ALREADY_EXISTS        = "用户名已存在"
	FAILED_TO_GET_GITHUB_LOGIN_URL = "获取 GitHub 登录 URL 失败"
	FAILED_TO_GET_GOOGLE_LOGIN_URL = "获取 Google 登录 URL 失败"
	FAILED_TO_GET_QQ_LOGIN_URL     = "获取 QQ 登录 URL 失败"
	FAILED_TO_GET_CUSTOM_LOGIN_URL = "获取自定义登录 URL 失败"
	OAUTH2_NOT_CONFIGURED          = "OAuth2 未配置"
	OAUTH2_NOT_ENABLED             = "OAuth2 未启用"
	NO_PERMISSION_BINDING_GITHUB   = "没有权限绑定 GitHub 账号"
	NO_PERMISSION_BINDING_GOOGLE   = "没有权限绑定 Google 账号"
	NO_PERMISSION_BINDING_QQ       = "没有权限绑定 QQ 账号"
	NO_PERMISSION_BINDING_CUSTOM   = "没有权限绑定自定义 OAuth2 账号"
)

// Connect 错误相关常量
const (
	INVALID_CONNECTION_URL = "connect url不能为空"
	CONNECT_HAS_EXISTS     = "connect 已经存在"
)

// Setting 错误相关常量
const (
	WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY = "未填写 Webhook 名称或 URL"
	INVALID_WEBHOOK_URL                 = "webhook URL 不合法或不安全"
	INVALID_CRON_EXPRESSION             = "无效的 Cron 表达式"
)

// Backup 错误相关常量
const (
	SNAPSHOT_UPLOAD_FAILED  = "快照上传失败"
	SNAPSHOT_RESTORE_FAILED = "快照恢复失败"
	DATABASE_CLOSE_FAILED   = "数据库关闭失败"
)

// Migration 错误相关常量
const (
	MIGRATION_JOB_NOT_FOUND = "迁移任务不存在"
)

// Agent 错误相关常量
const (
	AGENT_NOT_ENABLED        = "未启用 Agent "
	AGENT_PROVIDER_NOT_FOUND = "未找到对应的 Agent 提供商"
	AGENT_API_KEY_MISSING    = "未配置 Agent API Key 或 API Key 为空"
	AGENT_MODEL_MISSING      = "未配置 Agent 模型名称或模型名称不能为空"
	AGENT_SETTING_NOT_FOUND  = "未找到 Agent 设置"
)
