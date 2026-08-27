// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

const (
	INIT_LOGGER_PANIC          = "初始化 Logger 失败"
	READ_CONFIG_PANIC          = "读取配置文件失败"
	CREATE_DB_PATH_PANIC       = "创建数据库路径失败"
	DATABASE_NOT_INITED        = "数据库未初始化"
	INIT_DATABASE_PANIC        = "数据库初始化失败"
	MIGRATE_DB_PANIC           = "数据库迁移失败"
	INIT_HANDLERS_PANIC        = "初始化 Handlers 失败"
	INIT_TASKER_PANIC          = "初始化 Task Manager 失败"
	INIT_EVENT_REGISTRAR_PANIC = "初始化 EventRegistrar 失败"
	GIN_RUN_FAILED             = "启动 GIN 服务器失败"
)
