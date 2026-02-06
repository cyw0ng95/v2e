Migration Runbook: Normalize memory_card status values
=====================================================

Purpose
-------
This runbook documents how to safely apply tool/migrations/0003_normalize_memory_card_status.sql
to a SQLite production or staging database, how to verify the results, and how to rollback
by restoring a backup. The migration normalizes legacy/freeform status strings into the
canonical set used by the backend (new, learning, due, reviewed, mastered, archived).

Pre-conditions
--------------
- Ensure you have an offline copy / backup of the DB file before starting.
  Example:

  cp /path/to/your.db /path/to/your.db.$(date +%Y%m%d_%H%M%S).bak

- Validate that no other migrations are running and no heavy writes are happening.
- Prefer running in a staging environment first. For very large DBs, perform the migration
  during low-traffic hours.
- Confirm you can stop and restart the service that uses the DB (if necessary).

High-level strategy
-------------------
1. Take a filesystem-level backup of the SQLite DB.
2. Run the migration script (idempotent) against the DB.
3. Run verification queries to confirm status normalization and version backfill.
4. Run smoke tests (unit/integration) against the updated DB in staging.
5. If verification passes, deploy to production during a maintenance window.

Apply (manual sqlite3)
-----------------------
1) Backup

   cp /path/to/your.db /path/to/your.db.pre-migrate.bak

2) Run migration

   sqlite3 /path/to/your.db < tool/migrations/0003_normalize_memory_card_status.sql

3) Verify counts

   sqlite3 /path/to/your.db "SELECT status, COUNT(*) FROM memory_card_models GROUP BY status ORDER BY status;"

   Expect: at minimum there are rows for 'new', 'learning', and 'archived' if legacy rows existed.

Apply (Go runner)
------------------
If you prefer the Go runner included in this repo:

1) Build the runner

   go build -o tool/migrations/run_migrations tool/migrations/run_sqlite_migrations.go

2) Execute the runner

   ./tool/migrations/run_migrations -db /path/to/your.db -dir tool/migrations

Verification checks (recommended)
-------------------------------
- Sanity: status distribution
  sqlite3 /path/to/your.db "SELECT status, COUNT(*) FROM memory_card_models GROUP BY status ORDER BY status;"

- Ensure no remaining legacy tokens exist (case-insensitive):
  sqlite3 /path/to/your.db "SELECT status, COUNT(*) FROM memory_card_models WHERE lower(trim(status)) IN ('active','in-progress','in progress','archive','to_review','to-review') GROUP BY status;"

- Ensure version backfill: 
  sqlite3 /path/to/your.db "SELECT COUNT(*) FROM memory_card_models WHERE version IS NULL OR version = 0;"
  (expected 0)

- Application-level check: run service-level tests against the DB (in staging):
  - Run `go test ./pkg/notes -run TestNormalizationMigration -v`
  - Run other note/card integration tests that exercise parsing and transitions.

Smoke tests
-----------
- Start the service pointing at the migrated DB in a staging environment and exercise common flows:
  - Create a new card and ensure status set via RPC uses ParseCardStatus successfully.
  - Update a card status via RPCUpdateMemoryCard to a canonical value and ensure no ErrInvalidTransition occurs for valid transitions.
  - Run the existing concurrency test to ensure version bump behavior remains consistent.

Rollback
--------
If something goes wrong:

1) Stop services that are writing to the DB (to avoid further divergence).
2) Restore the backup made in Pre-conditions:

   cp /path/to/your.db.pre-migrate.bak /path/to/your.db

3) Restart services and re-run smoke tests.

Notes & risks
-------------
- Normalization is destructive with respect to the original literal textual values.
  Keep backups — restoration is the only safe rollback path.
- The ParseCardStatus validation in the service layer rejects certain legacy tokens (e.g., "active").
  If existing clients still send legacy tokens, normalize the DB first, then progressively
  harden the RPC handlers to validate and reject non-canonical tokens (or add compatibility shims if needed).
- For very large DBs, the UPDATE queries may take time and should be tested against a similar-sized copy in staging.

Checklist (run before migration)
--------------------------------
- [ ] Backup DB file created
- [ ] Staging run completed and smoke tests passed
- [ ] Maintenance window scheduled (if needed)
- [ ] All writes quiesced or service stopped during migration (recommended)

Post-migration monitoring
-------------------------
- Monitor application logs for ErrInvalidTransition or ErrConcurrentUpdate errors following migration.
- Verify user-reported issues and be ready to restore backup if critical failures are observed.
