package postgres

import (
	"context"
	"database/sql"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(ctx context.Context, item model.TodoItem) (int64, error) {
	const q = `
		INSERT INTO todo_items (list_id, title, description, done)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, item.ListID, item.Title, item.Description, item.Done).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ItemRepository) GetAllByList(ctx context.Context, listID int64) ([]model.TodoItem, error) {
	const q = `
		SELECT id, list_id, title, description, done
		FROM todo_items
		WHERE list_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, q, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.TodoItem, 0)
	for rows.Next() {
		var item model.TodoItem
		if err := rows.Scan(&item.ID, &item.ListID, &item.Title, &item.Description, &item.Done); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ItemRepository) GetByID(ctx context.Context, id int64) (model.TodoItem, error) {
	const q = `
		SELECT id, list_id, title, description, done
		FROM todo_items
		WHERE id = $1
	`

	var item model.TodoItem
	err := r.db.QueryRowContext(ctx, q, id).Scan(&item.ID, &item.ListID, &item.Title, &item.Description, &item.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.TodoItem{}, domainerr.ErrItemNotFound
		}
		return model.TodoItem{}, err
	}
	return item, nil
}

func (r *ItemRepository) Update(ctx context.Context, item model.TodoItem) error {
	const q = `
		UPDATE todo_items
		SET title = $1,
		    description = $2,
		    done = $3,
		    updated_at = NOW()
		WHERE id = $4
	`

	res, err := r.db.ExecContext(ctx, q, item.Title, item.Description, item.Done, item.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domainerr.ErrItemNotFound
	}
	return nil
}

func (r *ItemRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM todo_items WHERE id = $1`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domainerr.ErrItemNotFound
	}
	return nil
}
