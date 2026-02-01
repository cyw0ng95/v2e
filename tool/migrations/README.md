SQLite migration runner
======================

This directory contains SQL migrations for the Notes subsystem and a small
helper to run them against a SQLite database.

Files
- 0002_add_memory_card_version.sql - adds a `version` column to memory_card_models
- run_sqlite_migrations.go - small CLI to execute .sql files in order against a SQLite DB

Usage (recommended, manual SQL)
-------------------------------
Before running migrations in production:

1. Backup your SQLite file:

   cp /path/to/your.db /path/to/your.db.bak

2. Run the migration SQL:

   sqlite3 /path/to/your.db < tool/migrations/0002_add_memory_card_version.sql

3. Verify schema and contents (spot-check):

   sqlite3 /path/to/your.db "PRAGMA table_info('memory_card_models');"

Usage (Go-based runner)
-----------------------
You can also run the provided Go helper which executes every `*.sql` file in
the migrations directory in filename order.

Build & run:

  go build -o tool/migrations/run_migrations tool/migrations/run_sqlite_migrations.go
  ./tool/migrations/run_migrations -db /path/to/your.db -dir tool/migrations

Notes & safety
---------------
- Always back up the DB file before running migrations.
- The SQL provided is simple (`ALTER TABLE ADD COLUMN`) and safe for SQLite, but
  cross-check in staging first.
- If you have a large DB or custom schema, test the migration in a staging
  environment before applying to production.
