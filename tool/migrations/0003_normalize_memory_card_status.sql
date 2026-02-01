-- Normalize legacy memory card status string values to canonical set
-- Canonical statuses: 'new', 'learning', 'due', 'reviewed', 'mastered', 'archived'
-- Legacy mappings applied by this migration:
--   'active'        -> 'new'
--   'in-progress'    -> 'learning'
--   'in progress'    -> 'learning'
--   'to-review'/'to_review' -> 'learning'
--   'archive'        -> 'archived'

-- Safety: run inside a transaction when possible. Back up DB file before applying.
BEGIN TRANSACTION;

-- Normalize known legacy tokens (case-insensitive, trimmed)
UPDATE memory_card_models
SET status = CASE lower(trim(status))
    WHEN 'active' THEN 'new'
    WHEN 'in-progress' THEN 'learning'
    WHEN 'in progress' THEN 'learning'
    WHEN 'to-review' THEN 'learning'
    WHEN 'to_review' THEN 'learning'
    WHEN 'archive' THEN 'archived'
    ELSE status
END
WHERE status IS NOT NULL AND trim(status) != '';

-- Convert empty/NULL status values to 'new' (safe default)
UPDATE memory_card_models
SET status = 'new'
WHERE status IS NULL OR trim(status) = '';

-- Ensure version column is populated (optimistic concurrency col)
UPDATE memory_card_models
SET version = 1
WHERE version IS NULL;

COMMIT;

-- Validation queries (run after applying):
-- SELECT status, COUNT(*) FROM memory_card_models GROUP BY status ORDER BY status;

-- Rollback plan (manual):
-- 1) Restore from DB backup file made before running this migration.
-- 2) If you must revert without a backup, there's no easy SQL rollback for a normalization
--    mapping; you'd need to know previous values. Always take backups.
