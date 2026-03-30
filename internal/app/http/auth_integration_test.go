package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpapp "todo_crud/internal/app/http"
	"todo_crud/internal/app/http/dto"
	"todo_crud/internal/app/http/handler"
	"todo_crud/internal/domain/service"
	"todo_crud/internal/infrastructure/memory"
)

func newTestRouter() http.Handler {
	users := memory.NewUserRepository()
	authSvc := service.NewAuthService(users, "integration-secret", time.Hour)

	authHandler := handler.NewAuthHandler(authSvc)
	listHandler := handler.NewListHandler()
	itemHandler := handler.NewItemHandler()

	return httpapp.NewRouter(authHandler, listHandler, itemHandler)
}

func TestAuthIntegration_SignUpThenSignIn(t *testing.T) {
	router := newTestRouter()

	signUpBody, _ := json.Marshal(dto.SignUpRequest{
		Name:     "John",
		Username: "john",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewReader(signUpBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d; body=%s", http.StatusCreated, w.Code, w.Body.String())
	}
	var signUpResp dto.AuthResponse
	if err := json.NewDecoder(w.Body).Decode(&signUpResp); err != nil {
		t.Fatalf("decode sign-up response: %v", err)
	}
	if signUpResp.Token == "" {
		t.Fatal("expected sign-up token")
	}

	signInBody, _ := json.Marshal(dto.SignInRequest{Username: "john", Password: "password123"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-in", bytes.NewReader(signInBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d; body=%s", http.StatusOK, w.Code, w.Body.String())
	}
	var signInResp dto.AuthResponse
	if err := json.NewDecoder(w.Body).Decode(&signInResp); err != nil {
		t.Fatalf("decode sign-in response: %v", err)
	}
	if signInResp.Token == "" {
		t.Fatal("expected sign-in token")
	}
}

func TestAuthIntegration_SignUpDuplicate(t *testing.T) {
	router := newTestRouter()

	body, _ := json.Marshal(dto.SignUpRequest{Name: "John", Username: "john", Password: "password123"})

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewReader(body))
	firstReq.Header.Set("Content-Type", "application/json")
	firstW := httptest.NewRecorder()
	router.ServeHTTP(firstW, firstReq)
	if firstW.Code != http.StatusCreated {
		t.Fatalf("first sign-up expected %d, got %d; body=%s", http.StatusCreated, firstW.Code, firstW.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewReader(body))
	secondReq.Header.Set("Content-Type", "application/json")
	secondW := httptest.NewRecorder()
	router.ServeHTTP(secondW, secondReq)
	if secondW.Code != http.StatusConflict {
		t.Fatalf("second sign-up expected %d, got %d; body=%s", http.StatusConflict, secondW.Code, secondW.Body.String())
	}
}

func TestAuthIntegration_SignInInvalidCredentials(t *testing.T) {
	router := newTestRouter()

	body, _ := json.Marshal(dto.SignInRequest{Username: "unknown", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-in", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d; body=%s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}
