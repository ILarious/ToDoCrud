package memory

import (
	"context"
	"strings"
	"sync"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type UserRepository struct {
	mu      sync.RWMutex
	nextID  int64
	byID    map[int64]model.User
	byUname map[string]int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		nextID:  1,
		byID:    make(map[int64]model.User),
		byUname: make(map[string]int64),
	}
}

func (r *UserRepository) Create(_ context.Context, user model.User) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	uname := normalizeUsername(user.Username)
	if _, exists := r.byUname[uname]; exists {
		return 0, domainerr.ErrUserExists
	}

	user.ID = r.nextID
	r.nextID++

	r.byID[user.ID] = user
	r.byUname[uname] = user.ID

	return user.ID, nil
}

func (r *UserRepository) GetByUsername(_ context.Context, username string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byUname[normalizeUsername(username)]
	if !ok {
		return model.User{}, domainerr.ErrUserNotFound
	}

	user, ok := r.byID[id]
	if !ok {
		return model.User{}, domainerr.ErrUserNotFound
	}

	return user, nil
}

func normalizeUsername(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
