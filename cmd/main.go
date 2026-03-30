package main

import (
	"log"
	"os"
	"time"

	httpapp "todo_crud/internal/app/http"
	"todo_crud/internal/app/http/handler"
	"todo_crud/internal/domain/service"
	"todo_crud/internal/infrastructure/memory"
)

func main() {
	users := memory.NewUserRepository()

	tokenKey := os.Getenv("AUTH_SIGNING_KEY")
	authService := service.NewAuthService(users, tokenKey, 24*time.Hour)
	authHandler := handler.NewAuthHandler(authService)
	listHandler := handler.NewListHandler()
	itemHandler := handler.NewItemHandler()

	router := httpapp.NewRouter(authHandler, listHandler, itemHandler)
	srv := httpapp.NewServer(router)

	if err := srv.Run("8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
