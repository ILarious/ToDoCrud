package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"todo_crud/internal/infrastructure/postgres"
	pg "todo_crud/pkg/postgres"
	"todo_crud/pkg/worker_pool"
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

	planningRepo := postgres.NewPlanningRepository(db, envInt("TASK_PLAN_MAX_ATTEMPTS", 3))
	planner, err := worker_pool.NewPlanner(planningRepo, worker_pool.NewPlanGeneratorFromEnv(), worker_pool.PlannerConfig{
		PoolSize:     envInt("TASK_PLAN_WORKERS", 4),
		PollInterval: envDuration("TASK_PLAN_POLL_INTERVAL", 3*time.Second),
		StaleAfter:   envDuration("TASK_PLAN_STALE_AFTER", 30*time.Second),
		ClaimBatch:   envInt("TASK_PLAN_CLAIM_BATCH", 4),
		Logger:       log.Default(),
	})
	if err != nil {
		log.Fatalf("failed to create planner: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	planner.Start(ctx)
	<-ctx.Done()
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

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	v, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return v
}
