package model

type ServerError struct {
	Msg string
	Err error
}

// 失败相关的常量
const (
	INVALID_REQUEST_BODY = "无效的请求体"
)

// Auth 错误相关常量
const (
	USERNAME_OR_PASSWORD_NOT_BE_EMPTY = "用户名或密码不能为空"
	PASSWORD_INCORRECT                = "密码错误"
	USER_COUNT_EXCEED_LIMIT           = "用户数量超过限制"
	USERNAME_HAS_EXISTS               = "用户名已存在"
)
