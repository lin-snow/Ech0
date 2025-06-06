package model

type ServerError struct {
	Msg string
	Err error
}

// 失败相关的常量
const (
	INVALID_REQUEST_BODY = "无效的请求体"
	INVALID_PARAMS_BODY  = "无效参数"
)

// Auth 错误相关常量
const (
	USERNAME_OR_PASSWORD_NOT_BE_EMPTY = "用户名或密码不能为空"
	PASSWORD_INCORRECT                = "密码错误"
	USER_COUNT_EXCEED_LIMIT           = "用户数量超过限制"
	USERNAME_HAS_EXISTS               = "用户名已存在"
	TOKEN_NOT_FOUND                   = "未找到令牌,请点击右上角登录"
	TOKEN_NOT_VALID                   = "令牌无效，请重新登录"
	TOKEN_PARSE_ERROR                 = "令牌解析失败，请尝试重新登陆"
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
	SIGNUP_FIRST           = "请先注册用户"
)

// User 错误相关常量
const (
	USERNAME_ALREADY_EXISTS = "用户名已存在"
)

// Connect 错误相关常量
const (
	INVALID_CONNECTION_URL = "connect url不能为空"
	CONNECT_HAS_EXISTS     = "connect 已经存在"
)
