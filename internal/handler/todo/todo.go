package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/todo"
	service "github.com/lin-snow/ech0/internal/service/todo"
)

type TodoHandler struct {
	todoService *service.TodoService
}

// NewTodoHandler TodoHandler 的构造函数
func NewTodoHandler(todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

// AddTodo 添加新的待办事项
//
//	@Summary		添加新的待办事项
//	@Description	用户添加一条新的待办事项
//	@Tags			待办事项
//	@Accept			json
//	@Produce		json
//	@Param			todo	body		model.Todo		true	"待办事项内容"
//	@Success		200		{object}	res.Response	"添加成功"
//	@Failure		200		{object}	res.Response	"添加失败"
//	@Router			/todo [post]
func (todoHandler *TodoHandler) AddTodo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		var todo model.Todo
		if err := ctx.ShouldBindJSON(&todo); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := todoHandler.todoService.AddTodo(userid, &todo); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.ADD_TODO_SUCCESS,
		}
	})
}

// UpdateTodo 更新待办事项
//
//	@Summary		更新待办事项
//	@Description	根据ID更新指定的待办事项
//	@Tags			待办事项
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"待办事项ID"
//	@Success		200	{object}	res.Response	"更新成功"
//	@Failure		200	{object}	res.Response	"更新失败"
//	@Router			/todo/{id} [put]
func (todoHandler *TodoHandler) UpdateTodo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 从 URL 参数获取Echo ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS_BODY,
				Err: err,
			}
		}

		if err := todoHandler.todoService.UpdateTodo(userid, int64(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_TODO_SUCCESS,
		}
	})
}

// DeleteTodo 删除待办事项
//
//	@Summary		删除待办事项
//	@Description	根据ID删除指定的待办事项
//	@Tags			待办事项
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"待办事项ID"
//	@Success		200	{object}	res.Response	"删除成功"
//	@Failure		200	{object}	res.Response	"删除失败"
//	@Router			/todo/{id} [delete]
func (todoHandler *TodoHandler) DeleteTodo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 从 URL 参数获取Echo ID
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS_BODY,
				Err: err,
			}
		}

		if err := todoHandler.todoService.DeleteTodo(userid, int64(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_TODO_SUCCESS,
		}
	})
}

// GetTodoList 获取待办事项列表
//
//	@Summary		获取待办事项列表
//	@Description	获取当前用户的所有待办事项
//	@Tags			待办事项
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"获取成功"
//	@Failure		200	{object}	res.Response	"获取失败"
//	@Router			/todo [get]
func (todoHandler *TodoHandler) GetTodoList() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		todos, err := todoHandler.todoService.GetTodoList(userid)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: todos,
			Msg:  commonModel.GET_TODO_LIST_SUCCESS,
		}
	})
}
