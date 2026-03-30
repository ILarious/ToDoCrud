package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"todo_crud/internal/app/http/dto"
	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/service"
)

type AuthHandler struct {
	auth service.AuthService
}

func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req dto.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	token, err := h.auth.SignUp(r.Context(), req.Name, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid sign-up payload")
		case errors.Is(err, domainerr.ErrUsernameTaken):
			writeError(w, http.StatusConflict, "username already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto.AuthResponse{Token: token})
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req dto.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	token, err := h.auth.SignIn(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domainerr.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid sign-in payload")
		case errors.Is(err, domainerr.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid username or password")
		default:
			writeError(w, http.StatusInternalServerError, "failed to sign in")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.AuthResponse{Token: token})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}
