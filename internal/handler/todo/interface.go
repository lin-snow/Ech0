package handler

import "github.com/gin-gonic/gin"

type TodoHandlerInterface interface {
	GetTodoList(ctx *gin.Context)
	AddTodo(ctx *gin.Context)
	UpdateTodo(ctx *gin.Context)
	DeleteTodo(ctx *gin.Context)
}
