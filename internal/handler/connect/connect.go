package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	service "github.com/lin-snow/ech0/internal/service/connect"
)

type ConnectHandler struct {
	connectService service.Service
}

// NewConnectHandler ConnectHandler 的构造函数
func NewConnectHandler(connectService service.Service) *ConnectHandler {
	return &ConnectHandler{
		connectService: connectService,
	}
}

// AddConnect 添加连接
//
//	@Summary		添加连接
//	@Description	用户添加一个新的连接
//	@Tags			连接管理
//	@Accept			json
//	@Produce		json
//	@Param			connected	body		model.Connected		true	"连接信息"
//	@Success		200			{object}	handler.Response	"添加连接成功"
//	@Failure		200			{object}	handler.Response	"添加连接失败"
//	@Router			/connects [post]
func (connectHandler *ConnectHandler) AddConnect() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var connected connectModel.Connected
		if err := ctx.ShouldBindJSON(&connected); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
			}
		}

		if err := connectHandler.connectService.AddConnect(ctx.Request.Context(), connected); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.ADD_CONNECT_SUCCESS,
		}
	})
}

// DeleteConnect 删除连接
//
//	@Summary		删除连接
//	@Description	用户根据ID删除一个已添加的连接
//	@Tags			连接管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string				true	"连接ID (UUID)"
//	@Success		200	{object}	handler.Response	"删除连接成功"
//	@Failure		200	{object}	handler.Response	"删除连接失败"
//	@Router			/connects/{id} [delete]
func (connectHandler *ConnectHandler) DeleteConnect() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 从 URL 参数获取 ID
		id := ctx.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
			}
		}

		if err := connectHandler.connectService.DeleteConnect(ctx.Request.Context(), id); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_CONNECT_SUCCESS,
		}
	})
}

// GetConnectsInfo 获取所有添加的连接的信息
//
//	@Summary		获取所有添加的连接信息
//	@Description	获取当前用户所有已添加的连接的详细信息
//	@Tags			连接管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=[]model.Connected}	"获取连接信息成功"
//	@Failure		200	{object}	handler.Response							"获取连接信息失败"
//	@Router			/connects/info [get]
func (connectHandler *ConnectHandler) GetConnectsInfo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 调用 Service 层获取 Connect 信息
		connects, err := connectHandler.connectService.GetConnectsInfo()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: connects,
			Msg:  commonModel.GET_CONNECT_INFO_SUCCESS,
		}
	})
}

// GetConnect 提供当前实例的连接信息
//
//	@Summary		获取当前实例的连接信息
//	@Description	获取当前实例的连接详细信息
//	@Tags			连接管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=model.Connected}	"获取连接信息成功"
//	@Failure		200	{object}	handler.Response						"获取连接信息失败"
//	@Router			/connect [get]
func (connectHandler *ConnectHandler) GetConnect() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		connect, err := connectHandler.connectService.GetConnect()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: connect,
			Msg:  commonModel.CONNECT_SUCCESS,
		}
	})
}

// GetConnects 获取当前实例添加的所有连接
//
//	@Summary		获取所有连接
//	@Description	获取当前实例添加的所有连接列表
//	@Tags			连接管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=[]model.Connected}	"获取连接列表成功"
//	@Failure		200	{object}	handler.Response							"获取连接列表失败"
//	@Router			/connect/list [get]
func (connectHandler *ConnectHandler) GetConnects() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 调用 Service 层获取 Connect 列表
		connects, err := connectHandler.connectService.GetConnects()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: connects,
			Msg:  commonModel.GET_CONNECTED_LIST_SUCCESS,
		}
	})
}

// GetConnectsHealth 探测已保存互联地址的可达性及远端版本
//
//	@Summary		获取互联健康状态
//	@Description	由服务端请求各远端 /api/connect，返回每个已保存连接的状态与版本
//	@Tags			连接管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=[]model.ConnectedHealth}	"成功"
//	@Failure		200	{object}	handler.Response								"失败"
//	@Router			/connects/health [get]
func (connectHandler *ConnectHandler) GetConnectsHealth() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		rows, err := connectHandler.connectService.GetConnectsHealth()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: rows,
			Msg:  commonModel.GET_CONNECT_HEALTH_SUCCESS,
		}
	})
}
