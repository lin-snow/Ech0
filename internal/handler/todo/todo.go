package handler

import (
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/todo"
	service "github.com/lin-snow/ech0/internal/service/todo"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"net/http"
	"strconv"
)

type TodoHandler struct {
	todoService service.TodoServiceInterface
}

func NewTodoHandler(todoService service.TodoServiceInterface) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

func (todoHandler *TodoHandler) AddTodo(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	var todo model.Todo
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_REQUEST_BODY,
			Err: err,
		})))
		return
	}

	if err := todoHandler.todoService.AddTodo(userid, &todo); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.ADD_TODO_SUCCESS))
}

func (todoHandler *TodoHandler) UpdateTodo(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	// 从 URL 参数获取留言 ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_PARAMS_BODY,
			Err: err,
		})))
		return
	}

	if err := todoHandler.todoService.UpdateTodo(userid, int64(id)); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.UPDATE_TODO_SUCCESS))
}

func (todoHandler *TodoHandler) DeleteTodo(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	// 从 URL 参数获取留言 ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: commonModel.INVALID_PARAMS_BODY,
			Err: err,
		})))
		return
	}

	if err := todoHandler.todoService.DeleteTodo(userid, int64(id)); err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.DELETE_TODO_SUCCESS))
}

func (todoHandler *TodoHandler) GetTodoList(ctx *gin.Context) {
	// 获取当前用户 ID
	userid := ctx.MustGet("userid").(uint)

	todos, err := todoHandler.todoService.GetTodoList(userid)
	if err != nil {
		ctx.JSON(http.StatusOK, commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
			Msg: "",
			Err: err,
		})))
		return
	}

	ctx.JSON(http.StatusOK, commonModel.OK[[]model.Todo](todos, commonModel.GET_TODO_LIST_SUCCESS))
}
