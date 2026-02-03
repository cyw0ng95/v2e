# CWE Views (V) — Design

This design documents the CWE "View" feature and how it will be implemented in the local service.

Purpose
- Persist and serve CWE view resources (OpenAPI `V` views) for UI and API consumers.
- Provide CRUD and paginated listing; reserve job-controller integration for future website operations.

Storage
- Normalized SQLite tables prefixed `cwe_*`:
  - `cwe_views` (id TEXT PK, name, type, status, objective, raw BLOB)
  - `cwe_view_members`, `cwe_view_audience`, `cwe_view_references`, `cwe_view_notes`, `cwe_view_content_history`
- Nested arrays stored in separate tables linked by `view_id`.
- `raw` JSON blob stored on `cwe_views` for forward compatibility.

RPC Surface (local subprocess)
- `RPCSaveCWEView` (payload: `CWEView`)
- `RPCGetCWEViewByID` (payload: `{id}`)
- `RPCListCWEViews` (payload: `{offset,limit}`)
- `RPCDeleteCWEView` (payload: `{id}`)

Job Controller (future)
- A `pkg/cwe/job` controller will be added in a later tier to handle long-running view-generation/import jobs.
- It will persist session/progress and invoke local RPCs via the broker.

Testing
- Unit tests for store methods and handlers are provided (`pkg/cwe/local_views_test.go` and `cmd/local/cwe_handlers_views_test.go`).
- Integration with website and meta job orchestration will be tested in later tiers; integration tests remain unchanged.

Notes
- To enable migrations, call `AutoMigrateViews(db)` (function provided in `pkg/cwe/local_views.go`) from `NewLocalCWEStore`'s `AutoMigrate` list or manually where appropriate.
- Handler registration helper `RegisterCWEViewHandlers(sp, store, logger)` is provided; add calls in `cmd/local/main.go` where `sp` is available.
