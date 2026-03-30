package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"todo_crud/internal/app/http/dto"
)

type ListHandler struct{}

func NewListHandler() *ListHandler {
	return &ListHandler{}
}

func (h *ListHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []dto.TodoListResponse{})
}

func (h *ListHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	writeJSON(w, http.StatusCreated, dto.TodoListResponse{
		ID:          1,
		UserID:      1,
		Title:       req.Title,
		Description: req.Description,
	})
}

func (h *ListHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	listID, err := parseID(chi.URLParam(r, "listId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
		return
	}

	writeJSON(w, http.StatusOK, dto.TodoListResponse{
		ID:          listID,
		UserID:      1,
		Title:       "stub-list",
		Description: "",
	})
}

func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	resp := dto.TodoListResponse{ID: listID, UserID: 1, Title: "stub-list", Description: ""}
	if req.Title != nil {
		resp.Title = *req.Title
	}
	if req.Description != nil {
		resp.Description = *req.Description
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ListHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, err := parseID(chi.URLParam(r, "listId")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid listId")
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
