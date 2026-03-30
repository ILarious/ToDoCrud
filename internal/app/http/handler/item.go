package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"todo_crud/internal/app/http/dto"
)

type ItemHandler struct{}

func NewItemHandler() *ItemHandler {
	return &ItemHandler{}
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	if _, err := parseID(chi.URLParam(r, "listId")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}

	writeJSON(w, http.StatusOK, []dto.TodoItemResponse{})
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	writeJSON(w, http.StatusCreated, dto.TodoItemResponse{
		ID:          1,
		ListID:      listID,
		Title:       req.Title,
		Description: req.Description,
		Done:        false,
	})
}

func (h *ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, dto.TodoItemResponse{
		ID:          itemID,
		ListID:      listID,
		Title:       "stub-item",
		Description: "",
		Done:        false,
	})
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	resp := dto.TodoItemResponse{
		ID:          itemID,
		ListID:      listID,
		Title:       "stub-item",
		Description: "",
		Done:        false,
	}
	if req.Title != nil {
		resp.Title = *req.Title
	}
	if req.Description != nil {
		resp.Description = *req.Description
	}
	if req.Done != nil {
		resp.Done = *req.Done
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, err := parseID(chi.URLParam(r, "listId")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}
	if _, err := parseID(chi.URLParam(r, "itemId")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid itemId")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
