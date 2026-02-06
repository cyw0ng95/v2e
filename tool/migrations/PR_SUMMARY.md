Migration normalization: PR summary
=================================

Summary
-------
This PR introduces a normalization migration for memory card status values and
supporting documentation and tests. It standardizes legacy, free-form status
strings into the canonical set used by the backend: `new`, `learning`, `due`,
`reviewed`, `mastered`, `archived`.

Files changed (proposed)
------------------------
- tool/migrations/0003_normalize_memory_card_status.sql  (new)
- tool/migrations/README.md                             (updated)
- tool/migrations/RUNBOOK.md                            (new)
- pkg/notes/migration_test.go                           (new)

Why
---
There are legacy and UI-driven status strings in the codebase and tests such as
`active`, `in-progress`, `archive`, and `to_review`/`to-review`. ParseCardStatus
normalizes some synonyms, but persisted DB values remain inconsistent. Normalizing
them simplifies validation, removes edge-cases where ParseCardStatus would reject
persisted values, and enforces a single canonical representation for downstream
business logic.

What the migration does
-----------------------
- Maps legacy tokens to canonical values:
  - `active` -> `new`
  - `in-progress` / `in progress` -> `learning`
  - `to-review` / `to_review` -> `learning`
  - `archive` -> `archived`
- Replaces empty or NULL status with `new` as a safe default
- Backfills `version` to `1` where NULL

Tests & verification
--------------------
- Unit/integration test included at `pkg/notes/migration_test.go` seeds an
  in-memory SQLite DB with legacy tokens, runs the SQL migration, and asserts
  canonical statuses exist.
- Manual verification (instructions in RUNBOOK.md): run SQL to check status
  counts and ensure `version` backfill.

Rollback and safety
-------------------
- The migration is destructive to original literal values; create a filesystem
  backup of the SQLite DB first. The RUNBOOK.md includes explicit backup and
  rollback steps.

Proposed PR description (short)
------------------------------
Add a normalization migration for memory_card_models.status to convert legacy
status strings into canonical values and backfill version. Includes migration
tests and a runbook describing application and verification steps.
