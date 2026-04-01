package postgres

import (
	"context"
	"database/sql"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type ListRepository struct {
	db *sql.DB
}

func NewListRepository(db *sql.DB) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Create(ctx context.Context, list model.TodoList) (int64, error) {
	const q = `
		INSERT INTO todo_lists (user_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, list.UserID, list.Title, list.Description).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ListRepository) GetAllByUser(ctx context.Context, userID int64) ([]model.TodoList, error) {
	const q = `
		SELECT id, user_id, title, description
		FROM todo_lists
		WHERE user_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.TodoList, 0)
	for rows.Next() {
		var list model.TodoList
		if err := rows.Scan(&list.ID, &list.UserID, &list.Title, &list.Description); err != nil {
			return nil, err
		}
		out = append(out, list)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ListRepository) GetByID(ctx context.Context, id int64) (model.TodoList, error) {
	const q = `
		SELECT id, user_id, title, description
		FROM todo_lists
		WHERE id = $1
	`

	var list model.TodoList
	err := r.db.QueryRowContext(ctx, q, id).Scan(&list.ID, &list.UserID, &list.Title, &list.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.TodoList{}, domainerr.ErrListNotFound
		}
		return model.TodoList{}, err
	}
	return list, nil
}

func (r *ListRepository) Update(ctx context.Context, list model.TodoList) error {
	const q = `
		UPDATE todo_lists
		SET title = $1,
		    description = $2,
		    updated_at = NOW()
		WHERE id = $3
	`

	res, err := r.db.ExecContext(ctx, q, list.Title, list.Description, list.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domainerr.ErrListNotFound
	}
	return nil
}

func (r *ListRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM todo_lists WHERE id = $1`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domainerr.ErrListNotFound
	}
	return nil
}
