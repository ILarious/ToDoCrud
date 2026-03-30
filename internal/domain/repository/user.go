package repository

import (
	"context"

	"todo_crud/internal/domain/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (int64, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
}
