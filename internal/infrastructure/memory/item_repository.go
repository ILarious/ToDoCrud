package memory

import (
	"context"
	"sort"
	"sync"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type ItemRepository struct {
	mu       sync.RWMutex
	nextID   int64
	byID     map[int64]model.TodoItem
	byListID map[int64]map[int64]struct{}
}

func NewItemRepository() *ItemRepository {
	return &ItemRepository{
		nextID:   1,
		byID:     make(map[int64]model.TodoItem),
		byListID: make(map[int64]map[int64]struct{}),
	}
}

func (r *ItemRepository) Create(_ context.Context, item model.TodoItem) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.ID = r.nextID
	r.nextID++

	r.byID[item.ID] = item
	if _, ok := r.byListID[item.ListID]; !ok {
		r.byListID[item.ListID] = make(map[int64]struct{})
	}
	r.byListID[item.ListID][item.ID] = struct{}{}

	return item.ID, nil
}

func (r *ItemRepository) GetAllByList(_ context.Context, listID int64) ([]model.TodoItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byListID[listID]
	out := make([]model.TodoItem, 0, len(ids))
	for id := range ids {
		item, ok := r.byID[id]
		if ok {
			out = append(out, item)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *ItemRepository) GetByID(_ context.Context, id int64) (model.TodoItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.byID[id]
	if !ok {
		return model.TodoItem{}, domainerr.ErrItemNotFound
	}
	return item, nil
}

func (r *ItemRepository) Update(_ context.Context, item model.TodoItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[item.ID]
	if !ok {
		return domainerr.ErrItemNotFound
	}
	if existing.ListID != item.ListID {
		return domainerr.ErrItemNotFound
	}

	r.byID[item.ID] = item
	return nil
}

func (r *ItemRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[id]
	if !ok {
		return domainerr.ErrItemNotFound
	}
	delete(r.byID, id)
	if ids, ok := r.byListID[existing.ListID]; ok {
		delete(ids, id)
		if len(ids) == 0 {
			delete(r.byListID, existing.ListID)
		}
	}
	return nil
}
