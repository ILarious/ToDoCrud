package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	httpapp "todo_crud/internal/app/http"
	"todo_crud/internal/app/http/handler"
	"todo_crud/internal/domain/service"
	"todo_crud/internal/infrastructure/postgres"
	pg "todo_crud/pkg/postgres"
)

func main() {
	db, err := pg.New(pg.Config{
		Host:     env("POSTGRES_HOST", "127.0.0.1"),
		Port:     envInt("POSTGRES_PORT", 5432),
		User:     env("POSTGRES_USER", "postgres"),
		Password: env("POSTGRES_PASSWORD", "postgres"),
		Database: env("POSTGRES_DB", "todo"),
		SSLMode:  env("POSTGRES_SSLMODE", "disable"),
	})
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer func() {
		_ = pg.Close(db)
	}()

	if err := postgres.ApplyMigrations(context.Background(), db, env("MIGRATIONS_DIR", "migrations")); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	users := postgres.NewUserRepository(db)
	lists := postgres.NewListRepository(db)
	items := postgres.NewItemRepository(db)

	tokenKey := os.Getenv("AUTH_SIGNING_KEY")
	authService := service.NewAuthService(users, tokenKey, 24*time.Hour)
	listService := service.NewListService(lists)
	itemService := service.NewItemService(items, lists)
	authHandler := handler.NewAuthHandler(authService)
	listHandler := handler.NewListHandler(listService)
	itemHandler := handler.NewItemHandler(itemService)

	router := httpapp.NewRouter(authHandler, listHandler, itemHandler, authService)
	srv := httpapp.NewServer(router)

	if err := srv.Run(env("APP_PORT", "8080")); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
