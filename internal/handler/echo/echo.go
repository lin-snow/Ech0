package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	service "github.com/lin-snow/ech0/internal/service/echo"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type EchoHandler struct {
	echoService service.EchoServiceInterface
}

// NewEchoHandler EchoHandler 的构造函数
func NewEchoHandler(echoService service.EchoServiceInterface) *EchoHandler {
	return &EchoHandler{
		echoService: echoService,
	}
}

// PostEcho 创建新的Echo
//
//	@Summary		创建新的Echo
//	@Description	用户创建一条新的Echo动态
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			echo	body		model.Echo		true	"Echo内容"
//	@Success		200		{object}	res.Response	"创建成功"
//	@Failure		200		{object}	res.Response	"创建失败"
//	@Router			/echo [post]
func (echoHandler *EchoHandler) PostEcho() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var newEcho model.Echo
		if err := ctx.ShouldBindJSON(&newEcho); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		userId := ctx.MustGet("userid").(uint)
		if err := echoHandler.echoService.PostEcho(userId, &newEcho); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.POST_ECHO_SUCCESS,
		}
	})
}

// GetEchosByPage 获取Echo列表，支持分页, 兼容 GET Query 和 POST JSON 请求
//
//	@Summary		获取Echo列表（分页）
//	@Description	获取Echo列表，支持分页，兼容 GET Query 和 POST JSON 请求
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int							false	"页码（GET方式）"
//	@Param			pageSize	query		int							false	"每页数量（GET方式）"
//	@Param			body		body		commonModel.PageQueryDto	false	"分页参数（POST方式）"
//	@Success		200			{object}	res.Response{data=object}	"获取成功"
//	@Failure		200			{object}	res.Response				"获取失败"
//	@Router			/echo/page [get]
//	@Router			/echo/page [post]
func (echoHandler *EchoHandler) GetEchosByPage() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取分页参数
		var pageRequest commonModel.PageQueryDto

		switch ctx.Request.Method {
		case "GET":
			// 尝试从 query 中获取分页参数
			if err := ctx.ShouldBindQuery(&pageRequest); err != nil {
				return res.Response{
					Msg: commonModel.INVALID_QUERY_PARAMS,
					Err: err,
				}
			}

		case "POST":
			// 尝试从 JSON 中获取分页参数
			if err := ctx.ShouldBindJSON(&pageRequest); err != nil {
				return res.Response{
					Msg: commonModel.INVALID_REQUEST_BODY,
					Err: err,
				}
			}

		default:
			// 如果不是 GET 或 POST 请求，返回错误
			{
				return res.Response{
					Msg: commonModel.INVALID_REQUEST_METHOD,
					Err: errors.New(commonModel.INVALID_REQUEST_METHOD),
				}
			}
		}

		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)
		result, err := echoHandler.echoService.GetEchosByPage(userid, pageRequest)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.GET_ECHOS_BY_PAGE_SUCCESS,
		}
	})
}

// DeleteEcho 删除Echo
//
//	@Summary		删除Echo
//	@Description	根据ID删除指定的Echo动态
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"Echo ID"
//	@Success		200	{object}	res.Response	"删除成功"
//	@Failure		200	{object}	res.Response	"删除失败"
//	@Router			/echo/{id} [delete]
func (echoHandler *EchoHandler) DeleteEcho() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 从 URL 参数获取Echo ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		if err := echoHandler.echoService.DeleteEchoById(userid, uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_ECHO_SUCCESS,
		}
	})
}

// GetTodayEchos 获取今天的Echo列表
//
//	@Summary		获取今天的Echo列表
//	@Description	获取当前用户今天发布的所有Echo动态
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"获取成功"
//	@Failure		200	{object}	res.Response	"获取失败"
//	@Router			/echo/today [get]
func (echoHandler *EchoHandler) GetTodayEchos() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)
		timezone := timezoneUtil.NormalizeTimezone(ctx.GetHeader(timezoneUtil.DefaultTimezoneHeader))
		result, err := echoHandler.echoService.GetTodayEchos(userid, timezone)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.GET_TODAY_ECHOS_SUCCESS,
		}
	})
}

// UpdateEcho 更新Echo
//
//	@Summary		更新Echo
//	@Description	更新指定的Echo动态内容
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			echo	body		model.Echo		true	"要更新的Echo内容"
//	@Success		200		{object}	res.Response	"更新成功"
//	@Failure		200		{object}	res.Response	"更新失败"
//	@Router			/echo [put]
func (echoHandler *EchoHandler) UpdateEcho() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var updateEcho model.Echo
		if err := ctx.ShouldBindJSON(&updateEcho); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		userId := ctx.MustGet("userid").(uint)
		if err := echoHandler.echoService.UpdateEcho(userId, &updateEcho); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_ECHO_SUCCESS,
		}
	})
}

// LikeEcho 点赞Echo
//
//	@Summary		点赞Echo
//	@Description	根据ID为指定的Echo动态点赞
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"Echo ID"
//	@Success		200	{object}	res.Response	"点赞成功"
//	@Failure		200	{object}	res.Response	"点赞失败"
//	@Router			/echo/like/{id} [put]
func (echoHandler *EchoHandler) LikeEcho() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 从 URL 参数获取Echo ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		if err := echoHandler.echoService.LikeEcho(uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.LIKE_ECHO_SUCCESS,
		}
	})
}

// GetEchoById 获取指定 ID 的 Echo
//
//	@Summary		获取指定ID的Echo
//	@Description	根据ID获取指定的Echo动态详情
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"Echo ID"
//	@Success		200	{object}	res.Response	"获取成功"
//	@Failure		200	{object}	res.Response	"获取失败"
//	@Router			/echo/{id} [get]
func (echoHandler *EchoHandler) GetEchoById() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 从 URL 参数获取Echo ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		userId := ctx.MustGet("userid").(uint)

		echo, err := echoHandler.echoService.GetEchoById(userId, uint(id))
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: echo,
			Msg:  commonModel.GET_ECHO_BY_ID_SUCCESS,
		}
	})
}

// GetAllTags 获取所有标签
//
//	@Summary		获取所有标签
//	@Description	获取所有标签及其使用次数
//	@Tags			Tag
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"获取成功"
//	@Failure		200	{object}	res.Response	"获取失败"
//	@Router			/tags [get]
func (echoHandler *EchoHandler) GetAllTags() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		tags, err := echoHandler.echoService.GetAllTags()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: tags,
			Msg:  commonModel.GET_ALL_TAGS_SUCCESS,
		}
	})
}

// DeleteTag 删除标签
//
//	@Summary		删除标签
//	@Description	根据ID删除指定的标签
//	@Tags			Tag
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"标签 ID"
//	@Success		200	{object}	res.Response	"删除成功"
//	@Failure		200	{object}	res.Response	"删除失败"
//	@Router			/tag/{id} [delete]
func (echoHandler *EchoHandler) DeleteTag() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 从 URL 参数获取标签 ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		userid := ctx.MustGet("userid").(uint)

		if err := echoHandler.echoService.DeleteTag(userid, uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_TAG_SUCCESS,
		}
	})
}

// GetEchosByTagId 获取指定标签 ID 的 Echo 列表
//
//	@Summary		获取指定标签 ID 的 Echo 列表
//	@Description	根据标签 ID 获取包含该标签的 Echo 列表，支持 query 分页与搜索
//	@Tags			Echo
//	@Accept			json
//	@Produce		json
//	@Param			tagid		path		int							true	"标签 ID"
//	@Param			page		query		int							false	"页码"
//	@Param			pageSize	query		int							false	"每页数量"
//	@Param			search		query		string						false	"搜索关键字"
//	@Success		200			{object}	res.Response{data=object}	"获取成功"
//	@Failure		200			{object}	res.Response				"获取失败"
//	@Router			/echo/tag/{tagid} [get]
func (echoHandler *EchoHandler) GetEchosByTagId() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 从 URL 参数获取标签 ID
		tagIdStr := ctx.Param("tagid")
		tagId, err := strconv.ParseUint(tagIdStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		// 获取分页/搜索查询参数
		var pageRequest commonModel.PageQueryDto
		if err := ctx.ShouldBindQuery(&pageRequest); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_QUERY_PARAMS,
				Err: err,
			}
		}

		userid := ctx.MustGet("userid").(uint)

		result, err := echoHandler.echoService.GetEchosByTagId(userid, uint(tagId), pageRequest)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.GET_ECHOS_BY_TAG_ID_SUCCESS,
		}
	})
}
