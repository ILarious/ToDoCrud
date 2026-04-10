DROP INDEX IF EXISTS idx_todo_items_processing_claimed_at;
DROP INDEX IF EXISTS idx_todo_items_planning_status;

ALTER TABLE todo_items
    DROP CONSTRAINT IF EXISTS todo_items_planning_status_check,
    DROP COLUMN IF EXISTS planning_last_error,
    DROP COLUMN IF EXISTS planning_completed_at,
    DROP COLUMN IF EXISTS planning_claimed_at,
    DROP COLUMN IF EXISTS planning_attempts,
    DROP COLUMN IF EXISTS planning_status;
