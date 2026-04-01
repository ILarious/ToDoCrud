package service

import (
	"context"
	"errors"
	"testing"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/infrastructure/memory"
)

func TestListService_CreateGetUpdateDelete(t *testing.T) {
	repo := memory.NewListRepository()
	svc := NewListService(repo)

	created, err := svc.Create(context.Background(), 1, "Inbox", "main list")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected positive list ID, got %d", created.ID)
	}

	got, err := svc.GetByID(context.Background(), 1, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Title != "Inbox" {
		t.Fatalf("expected title Inbox, got %q", got.Title)
	}

	newTitle := "Work"
	updated, err := svc.Update(context.Background(), 1, created.ID, &newTitle, nil)
	if err != nil {
		t.Fatalf("update list: %v", err)
	}
	if updated.Title != "Work" {
		t.Fatalf("expected updated title Work, got %q", updated.Title)
	}

	if err := svc.Delete(context.Background(), 1, created.ID); err != nil {
		t.Fatalf("delete list: %v", err)
	}

	_, err = svc.GetByID(context.Background(), 1, created.ID)
	if !errors.Is(err, domainerr.ErrListNotFound) {
		t.Fatalf("expected ErrListNotFound after delete, got %v", err)
	}
}

func TestListService_CreateInvalidInput(t *testing.T) {
	repo := memory.NewListRepository()
	svc := NewListService(repo)

	_, err := svc.Create(context.Background(), 1, "   ", "")
	if !errors.Is(err, domainerr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
