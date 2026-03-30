package handler

import (
	"encoding/json"
	"net/http"

	"todo_crud/internal/app/http/dto"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req dto.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "name, username and password are required")
		return
	}

	writeJSON(w, http.StatusCreated, dto.AuthResponse{Token: "stub-token"})
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req dto.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	writeJSON(w, http.StatusOK, dto.AuthResponse{Token: "stub-token"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}
