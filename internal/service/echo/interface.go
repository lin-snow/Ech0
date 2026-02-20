package service

import (
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
)

type EchoServiceInterface interface {
	// PostEcho 创建新的Echo
	PostEcho(userid uint, newEcho *model.Echo) error

	// GetEchosByPage 获取Echo列表，支持分页
	GetEchosByPage(
		userid uint,
		pageQueryDto commonModel.PageQueryDto,
	) (commonModel.PageQueryResult[[]model.Echo], error)

	// DeleteEchoById 删除指定ID的Echo
	DeleteEchoById(userid, id uint) error

	// GetTodayEchos 获取今天的Echo列表
	GetTodayEchos(userid uint, timezone string) ([]model.Echo, error)

	// UpdateEcho 更新指定ID的Echo
	UpdateEcho(userid uint, echo *model.Echo) error

	// LikeEcho 点赞指定ID的Echo
	LikeEcho(id uint) error

	// GetEchoById 获取指定 ID 的 Echo
	GetEchoById(userId, id uint) (*model.Echo, error)

	// GetAllTags 获取所有标签
	GetAllTags() ([]model.Tag, error)

	// DeleteTag 删除标签
	DeleteTag(userid, id uint) error

	GetEchosByTagId(
		userId, tagId uint,
		pageQueryDto commonModel.PageQueryDto,
	) (commonModel.PageQueryResult[[]model.Echo], error)
}
