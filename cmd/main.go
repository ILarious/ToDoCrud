package main

import (
	"log"

	httpapp "todo_crud/internal/app/http"
	"todo_crud/internal/app/http/handler"
)

func main() {
	authHandler := handler.NewAuthHandler()
	listHandler := handler.NewListHandler()
	itemHandler := handler.NewItemHandler()

	router := httpapp.NewRouter(authHandler, listHandler, itemHandler)
	srv := httpapp.NewServer(router)

	if err := srv.Run("8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
