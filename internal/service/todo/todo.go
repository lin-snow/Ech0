package service

import (
	"errors"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/model/todo"
	"github.com/lin-snow/ech0/internal/repository/todo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type TodoService struct {
	todoRepository repository.TodoRepositoryInterface
	commonService  commonService.CommonServiceInterface
}

func NewTodoService(todoRepository repository.TodoRepositoryInterface, commonService commonService.CommonServiceInterface) TodoServiceInterface {
	return &TodoService{
		todoRepository: todoRepository,
		commonService:  commonService,
	}
}

func (todoService *TodoService) GetTodoList(userid uint) ([]model.Todo, error) {
	user, err := todoService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	todos, err := todoService.todoRepository.GetTodosByUserID(userid)
	if err != nil {
		return nil, err
	}

	// 如果没有找到，返回空切片
	if len(todos) == 0 {
		return []model.Todo{}, nil
	}
	// 返回查询到的 todos
	return todos, nil
}

func (todoService *TodoService) AddTodo(userid uint, todo *model.Todo) error {
	user, err := todoService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 设置TO DO
	todo.UserID = userid
	todo.Username = user.Username

	if err := todoService.todoRepository.CreateTodo(todo); err != nil {
		return err
	}
	return nil
}

func (todoService *TodoService) UpdateTodo(userid uint, id int64) error {
	user, err := todoService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 获取 To do
	theTodo, err := todoService.todoRepository.GetTodoByID(id)
	if err != nil {
		return err
	}

	// 检查该 To do 是否属于当前用户
	if theTodo.UserID != userid {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 设置To do的状态
	if theTodo.Status == model.NotDone {
		theTodo.Status = model.Done
	} else {
		theTodo.Status = model.NotDone
	}

	if err := todoService.todoRepository.UpdateTodo(theTodo); err != nil {
		return err
	}

	return nil
}

func (todoService *TodoService) DeleteTodo(userid uint, id int64) error {
	user, err := todoService.commonService.CommonGetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 获取 To do
	theTodo, err := todoService.todoRepository.GetTodoByID(id)
	if err != nil {
		return err
	}

	// 检查该 To do 是否属于当前用户
	if theTodo.UserID != userid {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if err := todoService.todoRepository.DeleteTodo(id); err != nil {
		return err
	}

	return nil
}
