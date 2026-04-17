package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"todo_crud/internal/app/http/dto"
	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type ItemService interface {
	GetAll(ctx context.Context, userID, listID int64) ([]model.TodoItem, error)
	Create(ctx context.Context, userID, listID int64, title, description string) (model.TodoItem, error)
	GetByID(ctx context.Context, userID, listID, itemID int64) (model.TodoItem, error)
	Update(ctx context.Context, userID, listID, itemID int64, title, description *string, done *bool) (model.TodoItem, error)
	Delete(ctx context.Context, userID, listID, itemID int64) error
}

type ItemHandler struct {
	items ItemService
}

func NewItemHandler(items ItemService) *ItemHandler {
	return &ItemHandler{items: items}
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := parseID(chi.URLParam(r, "listId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}

	items, err := h.items.GetAll(r.Context(), userID, listID)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid listId")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to list todo items")
		}
		return
	}

	resp := make([]dto.TodoItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, mapTodoItem(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := parseID(chi.URLParam(r, "listId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}

	var req dto.CreateTodoItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	item, err := h.items.Create(r.Context(), userID, listID, req.Title, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "title is required")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create todo item")
		}
		return
	}

	writeJSON(w, http.StatusCreated, mapTodoItem(item))
}

func (h *ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := parseID(chi.URLParam(r, "listId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}
	itemID, err := parseID(chi.URLParam(r, "itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid itemId")
		return
	}

	item, err := h.items.GetByID(r.Context(), userID, listID, itemID)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid itemId")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		case errors.Is(err, domainerr.ErrItemNotFound):
			writeError(w, http.StatusNotFound, "item not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get todo item")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapTodoItem(item))
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := parseID(chi.URLParam(r, "listId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}
	itemID, err := parseID(chi.URLParam(r, "itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid itemId")
		return
	}

	var req dto.UpdateTodoItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	item, err := h.items.Update(r.Context(), userID, listID, itemID, req.Title, req.Description, req.Done)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid update payload")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		case errors.Is(err, domainerr.ErrItemNotFound):
			writeError(w, http.StatusNotFound, "item not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update todo item")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapTodoItem(item))
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	listID, err := parseID(chi.URLParam(r, "listId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}
	itemID, err := parseID(chi.URLParam(r, "itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid itemId")
		return
	}

	if err := h.items.Delete(r.Context(), userID, listID, itemID); err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid itemId")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		case errors.Is(err, domainerr.ErrItemNotFound):
			writeError(w, http.StatusNotFound, "item not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete todo item")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapTodoItem(item model.TodoItem) dto.TodoItemResponse {
	return dto.TodoItemResponse{
		ID:          item.ID,
		ListID:      item.ListID,
		Title:       item.Title,
		Description: item.Description,
		Done:        item.Done,
	}
}
