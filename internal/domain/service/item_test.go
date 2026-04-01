package service

import (
	"context"
	"errors"
	"testing"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/infrastructure/memory"
)

func TestItemService_CreateGetUpdateDelete(t *testing.T) {
	listRepo := memory.NewListRepository()
	itemRepo := memory.NewItemRepository()
	listSvc := NewListService(listRepo)
	itemSvc := NewItemService(itemRepo, listRepo)

	list, err := listSvc.Create(context.Background(), 1, "Inbox", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}

	created, err := itemSvc.Create(context.Background(), 1, list.ID, "Task 1", "first")
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected positive item ID, got %d", created.ID)
	}

	got, err := itemSvc.GetByID(context.Background(), 1, list.ID, created.ID)
	if err != nil {
		t.Fatalf("get item by id: %v", err)
	}
	if got.Title != "Task 1" {
		t.Fatalf("expected title Task 1, got %q", got.Title)
	}

	newTitle := "Task updated"
	done := true
	updated, err := itemSvc.Update(context.Background(), 1, list.ID, created.ID, &newTitle, nil, &done)
	if err != nil {
		t.Fatalf("update item: %v", err)
	}
	if updated.Title != "Task updated" || !updated.Done {
		t.Fatalf("unexpected updated item: %+v", updated)
	}

	if err := itemSvc.Delete(context.Background(), 1, list.ID, created.ID); err != nil {
		t.Fatalf("delete item: %v", err)
	}

	_, err = itemSvc.GetByID(context.Background(), 1, list.ID, created.ID)
	if !errors.Is(err, domainerr.ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound after delete, got %v", err)
	}
}

func TestItemService_CreateInvalidInput(t *testing.T) {
	listRepo := memory.NewListRepository()
	itemRepo := memory.NewItemRepository()
	listSvc := NewListService(listRepo)
	itemSvc := NewItemService(itemRepo, listRepo)

	list, err := listSvc.Create(context.Background(), 1, "Inbox", "")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}

	_, err = itemSvc.Create(context.Background(), 1, list.ID, "   ", "")
	if !errors.Is(err, domainerr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
