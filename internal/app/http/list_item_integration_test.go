package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"todo_crud/internal/app/http/dto"
)

func TestListItemIntegration_CRUDFlow(t *testing.T) {
	router := newTestRouter()

	signUpBody, _ := json.Marshal(dto.SignUpRequest{
		Name:     "John",
		Username: "john_lists",
		Password: "password123",
	})
	signUpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewReader(signUpBody))
	signUpReq.Header.Set("Content-Type", "application/json")
	signUpW := httptest.NewRecorder()
	router.ServeHTTP(signUpW, signUpReq)
	if signUpW.Code != http.StatusCreated {
		t.Fatalf("sign-up expected %d, got %d; body=%s", http.StatusCreated, signUpW.Code, signUpW.Body.String())
	}

	var authResp dto.AuthResponse
	if err := json.NewDecoder(signUpW.Body).Decode(&authResp); err != nil {
		t.Fatalf("decode sign-up response: %v", err)
	}
	if authResp.Token == "" {
		t.Fatal("expected non-empty auth token")
	}

	createListBody, _ := json.Marshal(dto.CreateTodoListRequest{Title: "Inbox", Description: "main"})
	listReq := httptest.NewRequest(http.MethodPost, "/api/v1/lists/", bytes.NewReader(createListBody))
	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("Authorization", "Bearer "+authResp.Token)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusCreated {
		t.Fatalf("create list expected %d, got %d; body=%s", http.StatusCreated, listW.Code, listW.Body.String())
	}

	var listResp dto.TodoListResponse
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.ID <= 0 {
		t.Fatalf("expected list ID > 0, got %d", listResp.ID)
	}

	createItemBody, _ := json.Marshal(dto.CreateTodoItemRequest{Title: "Task", Description: "first"})
	itemReq := httptest.NewRequest(http.MethodPost, "/api/v1/lists/"+itoa(listResp.ID)+"/items/", bytes.NewReader(createItemBody))
	itemReq.Header.Set("Content-Type", "application/json")
	itemReq.Header.Set("Authorization", "Bearer "+authResp.Token)
	itemW := httptest.NewRecorder()
	router.ServeHTTP(itemW, itemReq)
	if itemW.Code != http.StatusCreated {
		t.Fatalf("create item expected %d, got %d; body=%s", http.StatusCreated, itemW.Code, itemW.Body.String())
	}

	var itemResp dto.TodoItemResponse
	if err := json.NewDecoder(itemW.Body).Decode(&itemResp); err != nil {
		t.Fatalf("decode item response: %v", err)
	}
	if itemResp.ID <= 0 {
		t.Fatalf("expected item ID > 0, got %d", itemResp.ID)
	}

	getItemReq := httptest.NewRequest(http.MethodGet, "/api/v1/lists/"+itoa(listResp.ID)+"/items/"+itoa(itemResp.ID)+"/", nil)
	getItemReq.Header.Set("Authorization", "Bearer "+authResp.Token)
	getItemW := httptest.NewRecorder()
	router.ServeHTTP(getItemW, getItemReq)
	if getItemW.Code != http.StatusOK {
		t.Fatalf("get item expected %d, got %d; body=%s", http.StatusOK, getItemW.Code, getItemW.Body.String())
	}

	deleteItemReq := httptest.NewRequest(http.MethodDelete, "/api/v1/lists/"+itoa(listResp.ID)+"/items/"+itoa(itemResp.ID)+"/", nil)
	deleteItemReq.Header.Set("Authorization", "Bearer "+authResp.Token)
	deleteItemW := httptest.NewRecorder()
	router.ServeHTTP(deleteItemW, deleteItemReq)
	if deleteItemW.Code != http.StatusNoContent {
		t.Fatalf("delete item expected %d, got %d; body=%s", http.StatusNoContent, deleteItemW.Code, deleteItemW.Body.String())
	}
}

func TestListItemIntegration_UnauthorizedWithoutToken(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/lists/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d; body=%s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
