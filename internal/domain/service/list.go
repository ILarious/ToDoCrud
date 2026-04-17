package service

import (
	"context"
	"errors"
	"strings"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type listRepository interface {
	Create(ctx context.Context, list model.TodoList) (int64, error)
	GetAllByUser(ctx context.Context, userID int64) ([]model.TodoList, error)
	GetByID(ctx context.Context, id int64) (model.TodoList, error)
	Update(ctx context.Context, list model.TodoList) error
	Delete(ctx context.Context, id int64) error
}

type listService struct {
	repo listRepository
}

func NewListService(repo listRepository) *listService {
	return &listService{repo: repo}
}

func (s *listService) GetAll(ctx context.Context, userID int64) ([]model.TodoList, error) {
	if userID <= 0 {
		return nil, domainerr.ErrInvalidInput
	}
	return s.repo.GetAllByUser(ctx, userID)
}

func (s *listService) Create(ctx context.Context, userID int64, title, description string) (model.TodoList, error) {
	if userID <= 0 {
		return model.TodoList{}, domainerr.ErrInvalidInput
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return model.TodoList{}, domainerr.ErrInvalidInput
	}

	list := model.TodoList{
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(description),
	}

	id, err := s.repo.Create(ctx, list)
	if err != nil {
		return model.TodoList{}, err
	}
	list.ID = id

	return list, nil
}

func (s *listService) GetByID(ctx context.Context, userID, listID int64) (model.TodoList, error) {
	if userID <= 0 || listID <= 0 {
		return model.TodoList{}, domainerr.ErrInvalidInput
	}

	list, err := s.repo.GetByID(ctx, listID)
	if err != nil {
		return model.TodoList{}, err
	}
	if list.UserID != userID {
		return model.TodoList{}, domainerr.ErrListNotFound
	}

	return list, nil
}

func (s *listService) Update(ctx context.Context, userID, listID int64, title, description *string) (model.TodoList, error) {
	list, err := s.GetByID(ctx, userID, listID)
	if err != nil {
		return model.TodoList{}, err
	}

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return model.TodoList{}, domainerr.ErrInvalidInput
		}
		list.Title = trimmed
	}
	if description != nil {
		list.Description = strings.TrimSpace(*description)
	}

	if err := s.repo.Update(ctx, list); err != nil {
		if errors.Is(err, domainerr.ErrListNotFound) {
			return model.TodoList{}, domainerr.ErrListNotFound
		}
		return model.TodoList{}, err
	}

	return list, nil
}

func (s *listService) Delete(ctx context.Context, userID, listID int64) error {
	if _, err := s.GetByID(ctx, userID, listID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, listID); err != nil {
		if errors.Is(err, domainerr.ErrListNotFound) {
			return domainerr.ErrListNotFound
		}
		return err
	}

	return nil
}
