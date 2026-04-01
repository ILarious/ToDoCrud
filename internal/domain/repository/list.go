package repository

import (
	"context"

	"todo_crud/internal/domain/model"
)

type TodoListRepository interface {
	Create(ctx context.Context, list model.TodoList) (int64, error)
	GetAllByUser(ctx context.Context, userID int64) ([]model.TodoList, error)
	GetByID(ctx context.Context, id int64) (model.TodoList, error)
	Update(ctx context.Context, list model.TodoList) error
	Delete(ctx context.Context, id int64) error
}
