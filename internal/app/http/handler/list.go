package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"todo_crud/internal/app/http/dto"
	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type ListService interface {
	GetAll(ctx context.Context, userID int64) ([]model.TodoList, error)
	Create(ctx context.Context, userID int64, title, description string) (model.TodoList, error)
	GetByID(ctx context.Context, userID, listID int64) (model.TodoList, error)
	Update(ctx context.Context, userID, listID int64, title, description *string) (model.TodoList, error)
	Delete(ctx context.Context, userID, listID int64) error
}

type ListHandler struct {
	lists ListService
}

func NewListHandler(lists ListService) *ListHandler {
	return &ListHandler{lists: lists}
}

func (h *ListHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	lists, err := h.lists.GetAll(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domainerr.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to list todo lists")
		return
	}

	resp := make([]dto.TodoListResponse, 0, len(lists))
	for _, list := range lists {
		resp = append(resp, mapTodoList(list))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ListHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateTodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	list, err := h.lists.Create(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		if errors.Is(err, domainerr.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create todo list")
		return
	}

	writeJSON(w, http.StatusCreated, mapTodoList(list))
}

func (h *ListHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	list, err := h.lists.GetByID(r.Context(), userID, listID)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid listId")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get todo list")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapTodoList(list))
}

func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req dto.UpdateTodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	list, err := h.lists.Update(r.Context(), userID, listID, req.Title, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid update payload")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update todo list")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapTodoList(list))
}

func (h *ListHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.lists.Delete(r.Context(), userID, listID); err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid listId")
		case errors.Is(err, domainerr.ErrListNotFound):
			writeError(w, http.StatusNotFound, "list not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete todo list")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

func mapTodoList(list model.TodoList) dto.TodoListResponse {
	return dto.TodoListResponse{
		ID:          list.ID,
		UserID:      list.UserID,
		Title:       list.Title,
		Description: list.Description,
	}
}
