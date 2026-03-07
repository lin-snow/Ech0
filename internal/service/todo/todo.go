package service

import (
	"context"
	"errors"
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/todo"
	repository "github.com/lin-snow/ech0/internal/repository/todo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/transaction"
)

type TodoService struct {
	transactor     transaction.Transactor             // 事务执行器
	todoRepository repository.TodoRepositoryInterface // To do数据层接口
	commonService  *commonService.CommonService       // 公共服务接口
}

func NewTodoService(
	tx transaction.Transactor,
	todoRepository repository.TodoRepositoryInterface,
	commonService *commonService.CommonService,
) *TodoService {
	return &TodoService{
		transactor:     tx,
		todoRepository: todoRepository,
		commonService:  commonService,
	}
}

// GetTodoList 获取当前用户的 To do列表
func (todoService *TodoService) GetTodoList(userid uint) ([]model.Todo, error) {
	// 检查执行操作的用户是否为管理员
	user, err := todoService.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	todos, err := todoService.todoRepository.GetTodosByUserID(context.Background(), userid)
	if err != nil {
		return nil, err
	}

	// 除去已完成的 To do
	for i := len(todos) - 1; i >= 0; i-- {
		if todos[i].Status == uint(model.Done) {
			todos = append(todos[:i], todos[i+1:]...)
		}
	}
	return todos, nil
}

// AddTodo 创建新的 To do
func (todoService *TodoService) AddTodo(userid uint, todo *model.Todo) error {
	return todoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 检查执行操作的用户是否为管理员
		user, err := todoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		todos, err := todoService.todoRepository.GetTodosByUserID(ctx, userid)
		if err != nil {
			return err
		}
		// 除去已完成的 To do
		for i := len(todos) - 1; i >= 0; i-- {
			if todos[i].Status == uint(model.Done) {
				todos = append(todos[:i], todos[i+1:]...)
			}
		}
		if len(todos) >= model.MaxTodoCount {
			fmt.Println("Current todo count:", todos)
			return errors.New(commonModel.TODO_EXCEED_LIMIT)
		}

		// 设置TO DO
		todo.UserID = userid
		todo.Username = user.Username
		todo.Status = uint(model.NotDone)

		// 创建 To do
		if err := todoService.todoRepository.CreateTodo(ctx, todo); err != nil {
			return err
		}
		return nil
	})
}

// UpdateTodo 更新指定ID的 To do
func (todoService *TodoService) UpdateTodo(userid uint, id int64) error {
	return todoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 检查执行操作的用户是否为管理员
		user, err := todoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 获取 To do
		theTodo, err := todoService.todoRepository.GetTodoByID(ctx, id)
		if err != nil {
			return err
		}

		// 检查该 To do 是否属于当前用户
		if theTodo.UserID != userid {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 设置To do的状态
		if theTodo.Status == uint(model.NotDone) {
			theTodo.Status = uint(model.Done)
		}

		if err := todoService.todoRepository.UpdateTodo(ctx, theTodo); err != nil {
			return err
		}

		return nil
	})
}

// DeleteTodo 删除指定ID的 To do
func (todoService *TodoService) DeleteTodo(userid uint, id int64) error {
	return todoService.transactor.Run(context.Background(), func(ctx context.Context) error {
		// 检查执行操作的用户是否为管理员
		user, err := todoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 获取 To do
		theTodo, err := todoService.todoRepository.GetTodoByID(ctx, id)
		if err != nil {
			return err
		}

		// 检查该 To do 是否属于当前用户
		if theTodo.UserID != userid {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		if err := todoService.todoRepository.DeleteTodo(ctx, id); err != nil {
			return err
		}

		return nil
	})
}
