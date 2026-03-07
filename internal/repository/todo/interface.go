package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/todo"
)

type TodoRepositoryInterface interface {
	// GetTodosByUserID 根据用户ID获取待办事项
	GetTodosByUserID(ctx context.Context, userid uint) ([]model.Todo, error)

	// CreateTodo 创建一个新的待办事项
	CreateTodo(ctx context.Context, todo *model.Todo) error

	// GetTodoByID 根据ID获取待办事项
	GetTodoByID(ctx context.Context, todoID int64) (*model.Todo, error)

	// UpdateTodo 更新待办事项
	UpdateTodo(ctx context.Context, todo *model.Todo) error

	// DeleteTodo 删除待办事项
	DeleteTodo(ctx context.Context, id int64) error
}
