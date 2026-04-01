package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user model.User) (int64, error) {
	const q = `
		INSERT INTO users (name, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, q, user.Name, normalizeUsername(user.Username), user.Password).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, domainerr.ErrUserExists
		}
		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	const q = `
		SELECT id, name, username, password_hash
		FROM users
		WHERE username = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, q, normalizeUsername(username)).Scan(&user.ID, &user.Name, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, domainerr.ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

func normalizeUsername(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func isUniqueViolation(err error) bool {
	// Keep infra layer decoupled from specific driver error types.
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "unique")
}
