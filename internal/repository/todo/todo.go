package repository

import (
	"github.com/lin-snow/ech0/internal/database"
	model "github.com/lin-snow/ech0/internal/model/todo"
	"gorm.io/gorm"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepositoryInterface {
	return &TodoRepository{
		db: db,
	}
}

func (todoRepository *TodoRepository) GetTodosByUserID(userid uint) ([]model.Todo, error) {
	var todos []model.Todo
	// 查询数据库(按创建时间，最新的在前)
	if err := database.DB.Where("user_id = ?", userid).Order("created_at DESC").Find(&todos).Error; err != nil {
		return nil, err
	}
	// 如果没有找到，返回空切片
	if len(todos) == 0 {
		return []model.Todo{}, nil
	}
	// 返回查询到的 todos
	return todos, nil
}

func (todoRepository *TodoRepository) CreateTodo(todo *model.Todo) error {
	if err := database.DB.Create(todo).Error; err != nil {
		return err
	}
	return nil
}

func (todoRepository *TodoRepository) GetTodoByID(todoID int64) (*model.Todo, error) {
	var todo model.Todo
	// 根据 ID 查找 To do
	if err := database.DB.Where("id = ?", todoID).First(&todo).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (todoRepository *TodoRepository) UpdateTodo(todo *model.Todo) error {
	// 根据 ID 查找 To do 并更新
	if err := database.DB.Model(&model.Todo{}).Where("id = ?", todo.ID).Updates(todo).Error; err != nil {
		return err
	}

	return nil
}

func (todoRepository *TodoRepository) DeleteTodo(id int64) error {
	// 根据 ID 删除 To do
	if err := database.DB.Delete(&model.Todo{}, id).Error; err != nil {
		return err
	}

	return nil
}
