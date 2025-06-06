package service

import model "github.com/lin-snow/ech0/internal/model/todo"

type TodoServiceInterface interface {
	GetTodoList(userid uint) ([]model.Todo, error)
	AddTodo(userid uint, todo *model.Todo) error
	UpdateTodo(userid uint, id int64) error
	DeleteTodo(userid uint, id int64) error
}
