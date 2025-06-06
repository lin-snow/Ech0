package repository

import model "github.com/lin-snow/ech0/internal/model/todo"

type TodoRepositoryInterface interface {
	GetTodosByUserID(userid uint) ([]model.Todo, error)
	CreateTodo(todo *model.Todo) error
	GetTodoByID(todoID int64) (*model.Todo, error)
	UpdateTodo(todo *model.Todo) error
	DeleteTodo(id int64) error
}
