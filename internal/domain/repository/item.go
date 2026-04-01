package repository

import (
	"context"

	"todo_crud/internal/domain/model"
)

type TodoItemRepository interface {
	Create(ctx context.Context, item model.TodoItem) (int64, error)
	GetAllByList(ctx context.Context, listID int64) ([]model.TodoItem, error)
	GetByID(ctx context.Context, id int64) (model.TodoItem, error)
	Update(ctx context.Context, item model.TodoItem) error
	Delete(ctx context.Context, id int64) error
}
