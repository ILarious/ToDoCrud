package memory

import (
	"context"
	"sort"
	"sync"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type ListRepository struct {
	mu     sync.RWMutex
	nextID int64
	byID   map[int64]model.TodoList
}

func NewListRepository() *ListRepository {
	return &ListRepository{
		nextID: 1,
		byID:   make(map[int64]model.TodoList),
	}
}

func (r *ListRepository) Create(_ context.Context, list model.TodoList) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	list.ID = r.nextID
	r.nextID++
	r.byID[list.ID] = list

	return list.ID, nil
}

func (r *ListRepository) GetAllByUser(_ context.Context, userID int64) ([]model.TodoList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]model.TodoList, 0)
	for _, list := range r.byID {
		if list.UserID == userID {
			out = append(out, list)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *ListRepository) GetByID(_ context.Context, id int64) (model.TodoList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, ok := r.byID[id]
	if !ok {
		return model.TodoList{}, domainerr.ErrListNotFound
	}
	return list, nil
}

func (r *ListRepository) Update(_ context.Context, list model.TodoList) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[list.ID]; !ok {
		return domainerr.ErrListNotFound
	}
	r.byID[list.ID] = list
	return nil
}

func (r *ListRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[id]; !ok {
		return domainerr.ErrListNotFound
	}
	delete(r.byID, id)
	return nil
}
