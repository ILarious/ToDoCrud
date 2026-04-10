ALTER TABLE todo_items
    ADD COLUMN IF NOT EXISTS planning_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS planning_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS planning_claimed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS planning_completed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS planning_last_error TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'todo_items_planning_status_check'
    ) THEN
        ALTER TABLE todo_items
            ADD CONSTRAINT todo_items_planning_status_check
            CHECK (planning_status IN ('pending', 'processing', 'completed', 'failed'));
    END IF;
END $$;

UPDATE todo_items
SET planning_status = CASE
        WHEN BTRIM(COALESCE(description, '')) = '' THEN 'pending'
        ELSE 'completed'
    END,
    planning_completed_at = CASE
        WHEN BTRIM(COALESCE(description, '')) = '' THEN NULL
        ELSE NOW()
    END
WHERE planning_status = 'pending' OR planning_status = 'completed';

CREATE INDEX IF NOT EXISTS idx_todo_items_planning_status
    ON todo_items (planning_status, id)
    WHERE done = FALSE;

CREATE INDEX IF NOT EXISTS idx_todo_items_processing_claimed_at
    ON todo_items (planning_claimed_at)
    WHERE planning_status = 'processing';
