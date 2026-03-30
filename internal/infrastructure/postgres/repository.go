package postgres

import "database/sql"

type Authorization interface {
}

type TodoList interface {
}

type TodoItem interface {
}

type Repository struct {
	DB *sql.DB
	Authorization
	TodoList
	TodoItem
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
