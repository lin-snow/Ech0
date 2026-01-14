package model

// SUCCESS_MESSAGE 成功相关的消息常量
const (
	SUCCESS_MESSAGE = "请求成功"
)

// Auth 成功相关常量
const (
	LOGIN_SUCCESS    = "登陆成功"
	REGISTER_SUCCESS = "注册成功"
)

// Echo 成功相关常量
const (
	POST_ECHO_SUCCESS           = "发布Echo成功！"
	GET_ECHOS_BY_PAGE_SUCCESS   = "获取Echos成功！"
	DELETE_ECHO_SUCCESS         = "删除Echo成功"
	GET_TODAY_ECHOS_SUCCESS     = "获取当日Echos成功"
	UPDATE_ECHO_SUCCESS         = "更新Echo成功"
	LIKE_ECHO_SUCCESS           = "点赞Echo成功"
	GET_ECHO_BY_ID_SUCCESS      = "获取Echo成功"
	GET_ALL_TAGS_SUCCESS        = "获取所有标签成功"
	DELETE_TAG_SUCCESS          = "删除标签成功"
	GET_ECHOS_BY_TAG_ID_SUCCESS = "获取标签下的Echos成功"
)

// Common 成功相关常量
const (
	UPLOAD_SUCCESS             = "上传成功"
	DELETE_SUCCESS             = "删除成功"
	GET_STATUS_SUCCESS         = "获取状态成功"
	GET_HEATMAP_SUCCESS        = "获取热力图成功"
	GET_MUSIC_URL_SUCCESS      = "获取音乐播放链接成功"
	GET_HELLO_SUCCESS          = "获取Hello成功"
	GET_HEALTHZ_SUCCESS        = "健康检查"
	GET_S3_PRESIGN_URL_SUCCESS = "获取 S3 预签名 URL 成功"
	GET_METRICS_SUCCESS        = "获取系统指标成功"
	GET_WEBSITE_TITLE_SUCCESS  = "获取网站标题成功"
)

// Inbox 成功相关常量
const (
	GET_INBOX_LIST_SUCCESS   = "获取收件箱成功"
	GET_UNREAD_INBOX_SUCCESS = "获取未读消息成功"
	MARK_INBOX_READ_SUCCESS  = "标记消息已读成功"
	DELETE_INBOX_SUCCESS     = "删除收件箱消息成功"
	CLEAR_INBOX_SUCCESS      = "清空收件箱成功"
)

// Setting 成功相关常量
const (
	GET_SETTINGS_SUCCESS              = "获取设置成功！"
	UPDATE_SETTINGS_SUCCESS           = "更新设置成功！"
	GET_COMMENT_SETTINGS_SUCCESS      = "获取评论设置成功！"
	UPDATE_COMMENT_SETTINGS_SUCCESS   = "更新评论设置成功！"
	GET_S3_SETTINGS_SUCCESS           = "获取 S3 存储设置成功！"
	UPDATE_S3_SETTINGS_SUCCESS        = "更新 S3 存储设置成功！"
	GET_OAUTH_SETTINGS_SUCCESS        = "获取 OAuth 设置成功！"
	UPDATE_OAUTH_SETTINGS_SUCCESS     = "更新 OAuth 设置成功！"
	GET_OAUTH2_STATUS_SUCCESS         = "获取 OAuth2 状态成功"
	GET_WEBHOOK_SUCCESS               = "获取 Webhook 成功"
	DELETE_WEBHOOK_SUCCESS            = "删除 Webhook 成功"
	UPDATE_WEBHOOK_SUCCESS            = "更新 Webhook 成功"
	CREATE_WEBHOOK_SUCCESS            = "创建 Webhook 成功"
	LIST_ACCESS_TOKENS_SUCCESS        = "列出访问令牌成功"
	CREATE_ACCESS_TOKEN_SUCCESS       = "创建访问令牌成功"
	DELETE_ACCESS_TOKEN_SUCCESS       = "删除访问令牌成功"
	GET_FEDIVERSE_SETTINGS_SUCCESS    = "获取联邦网络设置成功"
	UPDATE_FEDIVERSE_SETTINGS_SUCCESS = "更新联邦网络设置成功"
	SCHEDULE_BACKUP_SUCCESS           = "设置备份计划成功"
)

// To do 成功相关常量
const (
	GET_TODO_LIST_SUCCESS = "获取Todo list 成功"
	ADD_TODO_SUCCESS      = "添加Todo成功"
	UPDATE_TODO_SUCCESS   = "更新Todo成功"
	DELETE_TODO_SUCCESS   = "删除Todo成功"
)

// User 成功相关常量
const (
	UPDATE_USER_SUCCESS       = "更新用户信息成功"
	GET_USER_SUCCESS          = "获取用户列表成功"
	GET_USER_INFO_SUCCESS     = "获取用户信息成功"
	DELETE_USER_SUCCESS       = "删除用户成功"
	BIND_GITHUB_SUCCESS       = "绑定 GitHub 账号成功"
	GET_OAUTH_BINGURL_SUCCESS = "获取绑定 URL 成功"
	GET_OAUTH_INFO_SUCCESS    = "获取 OAuth2 信息成功"
)

// Connect 成功相关常量
const (
	CONNECT_SUCCESS            = "连接成功"
	ADD_CONNECT_SUCCESS        = "添加连接成功"
	DELETE_CONNECT_SUCCESS     = "连接已取消"
	GET_CONNECT_INFO_SUCCESS   = "获取 Connect 信息成功"
	GET_CONNECTED_LIST_SUCCESS = "获取连接列表成功"
)

// Backup 成功相关常量
const (
	BACKUP_SUCCESS        = "备份成功"
	EXPORT_BACKUP_SUCCESS = "导出备份成功"
	IMPORT_BACKUP_SUCCESS = "导入备份成功"
)

// Fediverse 成功相关常量
const (
	FEDIVERSE_SEARCH_ACTOR_SUCCESS      = "搜索 Actor 成功"
	FEDIVERSE_FOLLOW_SUCCESS            = "关注请求已发送"
	FEDIVERSE_UNFOLLOW_SUCCESS          = "取消关注请求已发送"
	FEDIVERSE_LIKE_SUCCESS              = "点赞请求已发送"
	FEDIVERSE_UNDO_LIKE_SUCCESS         = "取消点赞请求已发送"
	FEDIVERSE_GET_FOLLOW_STATUS_SUCCESS = "获取关注状态成功"
	FEDIVERSE_GET_TIMELINE_SUCCESS      = "获取时间线成功"
)

// Agent 成功相关常量
const (
	AGENT_GET_RECENT_SUCCESS = "获取近期活动总结成功"
)
