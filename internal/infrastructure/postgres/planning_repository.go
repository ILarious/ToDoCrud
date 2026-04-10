package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"todo_crud/pkg/worker_pool"
)

type PlanningRepository struct {
	db          *sql.DB
	maxAttempts int
}

func NewPlanningRepository(db *sql.DB, maxAttempts int) *PlanningRepository {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	return &PlanningRepository{
		db:          db,
		maxAttempts: maxAttempts,
	}
}

func (r *PlanningRepository) ClaimTasks(ctx context.Context, limit int, staleAfter time.Duration) ([]worker_pool.Task, error) {
	if limit <= 0 {
		return nil, nil
	}
	if staleAfter <= 0 {
		staleAfter = 30 * time.Second
	}

	const q = `
		WITH candidate AS (
			SELECT id
			FROM todo_items
			WHERE done = FALSE
			  AND (
			      planning_status = 'pending'
			      OR (planning_status = 'failed' AND planning_attempts < $3)
			      OR (
			          planning_status = 'processing'
			          AND planning_claimed_at IS NOT NULL
			          AND planning_claimed_at <= NOW() - ($1 * INTERVAL '1 second')
			      )
			  )
			ORDER BY id
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE todo_items AS t
		SET planning_status = 'processing',
		    planning_claimed_at = NOW(),
		    planning_last_error = '',
		    planning_attempts = t.planning_attempts + 1
		FROM candidate
		WHERE t.id = candidate.id
		RETURNING t.id, t.list_id, t.title
	`

	rows, err := r.db.QueryContext(ctx, q, int64(staleAfter/time.Second), limit, r.maxAttempts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]worker_pool.Task, 0, limit)
	for rows.Next() {
		var task worker_pool.Task
		if err := rows.Scan(&task.ID, &task.ListID, &task.Title); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *PlanningRepository) CompleteTask(ctx context.Context, taskID int64, plan string) error {
	const q = `
		UPDATE todo_items
		SET description = $1,
		    planning_status = 'completed',
		    planning_claimed_at = NULL,
		    planning_completed_at = NOW(),
		    planning_last_error = '',
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, q, strings.TrimSpace(plan), taskID)
	return err
}

func (r *PlanningRepository) FailTask(ctx context.Context, taskID int64, reason string) error {
	const q = `
		UPDATE todo_items
		SET planning_status = 'failed',
		    planning_claimed_at = NULL,
		    planning_last_error = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, q, truncate(strings.TrimSpace(reason), 2000), taskID)
	return err
}

func truncate(s string, limit int) string {
	if limit <= 0 || len(s) <= limit {
		return s
	}
	return s[:limit]
}
