package model

// Panic Constants
const (
	INIT_LOGGER_PANIC    = "Logger 初始化失败"
	READ_CONFIG_PANIC    = "读取配置文件失败"
	CREATE_DB_PATH_PANIC = "创建数据库路径失败"
	INIT_DATABASE_PANIC  = "数据库初始化失败"
	MIGRATE_DB_PANIC     = "数据库迁移失败"
	INIT_HANDLERS_PANIC  = "Handlers 初始化失败"
	GIN_RUN_FAILED       = "GIN 启动失败"
)
