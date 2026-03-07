package repository

import (
	"context"

	"github.com/lin-snow/ech0/internal/cache"
	model "github.com/lin-snow/ech0/internal/model/todo"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type TodoRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewTodoRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) TodoRepositoryInterface {
	return &TodoRepository{
		db:    dbProvider,
		cache: cache,
	}
}

// getDB 从上下文中获取事务
func (todoRepository *TodoRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return todoRepository.db()
}

// GetTodosByUserID 根据用户ID获取待办事项
func (todoRepository *TodoRepository) GetTodosByUserID(ctx context.Context, userid uint) ([]model.Todo, error) {
	var todos []model.Todo
	// 查询数据库(按创建时间，最新的在前)
	if err := todoRepository.getDB(ctx).Where("user_id = ?", userid).Order("created_at DESC").Find(&todos).Error; err != nil {
		return nil, err
	}
	// 如果没有找到，返回空切片
	if len(todos) == 0 {
		return []model.Todo{}, nil
	}

	// 返回查询到的 todos
	return todos, nil
}

// CreateTodo 创建一个新的待办事项
func (todoRepository *TodoRepository) CreateTodo(ctx context.Context, todo *model.Todo) error {
	if err := todoRepository.getDB(ctx).Create(todo).Error; err != nil {
		return err
	}

	return nil
}

// GetTodoByID 根据ID获取待办事项
func (todoRepository *TodoRepository) GetTodoByID(ctx context.Context, todoID int64) (*model.Todo, error) {
	var todo model.Todo
	// 根据 ID 查找 To do
	if err := todoRepository.getDB(ctx).Where("id = ?", todoID).First(&todo).Error; err != nil {
		return nil, err
	}

	return &todo, nil
}

// UpdateTodo 更新待办事项
func (todoRepository *TodoRepository) UpdateTodo(ctx context.Context, todo *model.Todo) error {
	// 根据 ID 查找 To do 并更新
	if err := todoRepository.getDB(ctx).Model(&model.Todo{}).Where("id = ?", todo.ID).Updates(todo).Error; err != nil {
		return err
	}

	return nil
}

// DeleteTodo 删除待办事项
func (todoRepository *TodoRepository) DeleteTodo(ctx context.Context, id int64) error {
	// 根据 ID 删除 To do
	if err := todoRepository.getDB(ctx).Delete(&model.Todo{}, id).Error; err != nil {
		return err
	}

	return nil
}
