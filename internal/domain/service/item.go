package service

import (
	"context"
	"errors"
	"strings"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type itemRepository interface {
	Create(ctx context.Context, item model.TodoItem) (int64, error)
	GetAllByList(ctx context.Context, listID int64) ([]model.TodoItem, error)
	GetByID(ctx context.Context, id int64) (model.TodoItem, error)
	Update(ctx context.Context, item model.TodoItem) error
	Delete(ctx context.Context, id int64) error
}

type listReader interface {
	GetByID(ctx context.Context, id int64) (model.TodoList, error)
}

type itemService struct {
	items itemRepository
	lists listReader
}

func NewItemService(items itemRepository, lists listReader) *itemService {
	return &itemService{items: items, lists: lists}
}

func (s *itemService) GetAll(ctx context.Context, userID, listID int64) ([]model.TodoItem, error) {
	if err := s.ensureListAccess(ctx, userID, listID); err != nil {
		return nil, err
	}
	return s.items.GetAllByList(ctx, listID)
}

func (s *itemService) Create(ctx context.Context, userID, listID int64, title, description string) (model.TodoItem, error) {
	if err := s.ensureListAccess(ctx, userID, listID); err != nil {
		return model.TodoItem{}, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return model.TodoItem{}, domainerr.ErrInvalidInput
	}

	item := model.TodoItem{
		ListID:      listID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Done:        false,
	}

	id, err := s.items.Create(ctx, item)
	if err != nil {
		return model.TodoItem{}, err
	}
	item.ID = id
	return item, nil
}

func (s *itemService) GetByID(ctx context.Context, userID, listID, itemID int64) (model.TodoItem, error) {
	if err := s.ensureListAccess(ctx, userID, listID); err != nil {
		return model.TodoItem{}, err
	}
	if itemID <= 0 {
		return model.TodoItem{}, domainerr.ErrInvalidInput
	}

	item, err := s.items.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, domainerr.ErrItemNotFound) {
			return model.TodoItem{}, domainerr.ErrItemNotFound
		}
		return model.TodoItem{}, err
	}
	if item.ListID != listID {
		return model.TodoItem{}, domainerr.ErrItemNotFound
	}

	return item, nil
}

func (s *itemService) Update(ctx context.Context, userID, listID, itemID int64, title, description *string, done *bool) (model.TodoItem, error) {
	item, err := s.GetByID(ctx, userID, listID, itemID)
	if err != nil {
		return model.TodoItem{}, err
	}

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return model.TodoItem{}, domainerr.ErrInvalidInput
		}
		item.Title = trimmed
	}
	if description != nil {
		item.Description = strings.TrimSpace(*description)
	}
	if done != nil {
		item.Done = *done
	}

	if err := s.items.Update(ctx, item); err != nil {
		if errors.Is(err, domainerr.ErrItemNotFound) {
			return model.TodoItem{}, domainerr.ErrItemNotFound
		}
		return model.TodoItem{}, err
	}

	return item, nil
}

func (s *itemService) Delete(ctx context.Context, userID, listID, itemID int64) error {
	if _, err := s.GetByID(ctx, userID, listID, itemID); err != nil {
		return err
	}

	if err := s.items.Delete(ctx, itemID); err != nil {
		if errors.Is(err, domainerr.ErrItemNotFound) {
			return domainerr.ErrItemNotFound
		}
		return err
	}
	return nil
}

func (s *itemService) ensureListAccess(ctx context.Context, userID, listID int64) error {
	if userID <= 0 || listID <= 0 {
		return domainerr.ErrInvalidInput
	}

	list, err := s.lists.GetByID(ctx, listID)
	if err != nil {
		if errors.Is(err, domainerr.ErrListNotFound) {
			return domainerr.ErrListNotFound
		}
		return err
	}
	if list.UserID != userID {
		return domainerr.ErrListNotFound
	}

	return nil
}
