-- Add version column to memory_card_models and backfill existing rows
ALTER TABLE memory_card_models ADD COLUMN version INTEGER DEFAULT 1;
-- Backfill for SQLite/Postgres/MySQL compatibility
UPDATE memory_card_models SET version = 1 WHERE version IS NULL;

-- Note: For Postgres you might want to set a sequence or default nextval explicitly.
-- Run this in a maintenance window and ensure backups exist before applying.
