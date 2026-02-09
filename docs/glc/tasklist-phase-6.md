# GLC Project Implementation Plan - Phase 6: Backend Integration

## Phase Overview

This phase implements backend integration for saving and restoring topology/graph data through RPC calls to the v2e broker system. This enables persistent storage, multi-user collaboration, and integration with the existing v2e vulnerability management ecosystem.

## Task 6.1: Backend Service Design

### Change Estimation (File Level)
- New files (Go): 8-10
- New files (Frontend): 3-5
- Modified files: 5-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Go code: ~1,500-2,200 lines
- TypeScript code: ~300-500 lines
- Documentation: ~400-600 lines

### Detailed Work Items

#### 6.1.1 GLC Service Specification
**File List**:
- `cmd/glc/service.md` - GLC service RPC API documentation

**Work Content**:
Define RPC API specification for GLC service:
- CreateGraph - Create new graph with metadata
- SaveGraph - Save/update graph data
- LoadGraph - Load graph by ID
- DeleteGraph - Delete graph
- ListGraphs - List user's graphs (with pagination)
- DuplicateGraph - Duplicate existing graph
- ShareGraph - Generate share token
- LoadGraphByShare - Load graph via share token
- ValidatePreset - Validate preset data

For each RPC method, document:
- Method name and description
- Request parameters (name, type, required/optional, description)
- Response fields (name, type, description)
- Error conditions and messages
- Example request/response

**Acceptance Criteria**:
1. WHEN service.md is reviewed, SHALL contain all 8 RPC methods with complete specifications
2. WHEN each RPC method is documented, SHALL include all required fields (params, response, errors, examples)
3. WHEN API is reviewed, SHALL follow v2e RPC patterns and conventions
4. WHEN specification is complete, SHALL be ready for implementation

#### 6.1.2 Database Schema Design
**File List**:
- `cmd/glc/schema/graph.sql` - Graph tables schema
- `cmd/glc/schema/001_create_glc_tables.sql` - Migration file

**Work Content**:
Design and create database schema for GLC:
- glc_graphs table (id, user_id, preset_id, title, description, metadata, created_at, updated_at, is_public, share_token, etc.)
- glc_graph_versions table (id, graph_id, version, node_data, edge_data, viewport_data, created_at, created_by)
- glc_presets table (id, user_id, name, description, preset_data, version, is_public, created_at, updated_at)
- Indexes for performance (user_id, preset_id, share_token)
- Foreign key relationships

**Acceptance Criteria**:
1. WHEN schema is reviewed, SHALL support all GLC features (graphs, versions, custom presets)
2. WHEN glc_graphs table is created, SHALL contain all required fields
3. WHEN glc_graph_versions table is created, SHALL support versioning
4. WHEN indexes are created, SHALL optimize common queries (user graphs, share loads)
5. WHEN foreign keys are defined, SHALL ensure referential integrity

#### 6.1.3 Go Data Models
**File List**:
- `cmd/glc/internal/models/graph.go` - Graph data models
- `cmd/glc/internal/models/preset.go` - Preset data models
- `cmd/glc/internal/models/requests.go` - RPC request models
- `cmd/glc/internal/models/responses.go` - RPC response models

**Work Content**:
Define Go data models matching database schema:
- Graph struct with database tags
- GraphVersion struct
- Preset struct
- Request structs for all RPC methods
- Response structs for all RPC methods
- Validation methods
- JSON serialization/deserialization

**Acceptance Criteria**:
1. WHEN Graph struct is defined, SHALL have correct database tags for all fields
2. WHEN request structs are defined, SHALL match RPC API specification
3. WHEN response structs are defined, SHALL match RPC API specification
4. WHEN models are validated, SHALL pass all validation rules
5. WHEN models are serialized, SHALL produce valid JSON

---

## Task 6.2: GLC Service Implementation

### Change Estimation (File Level)
- New files (Go): 15-20
- Modified files (Go): 5-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Go code: ~2,500-3,500 lines
- Tests (Go): ~800-1,200 lines

### Detailed Work Items

#### 6.2.1 GLC Service Main Setup
**File List**:
- `cmd/glc/main.go` - GLC service entry point
- `cmd/glc/config/config.go` - Configuration management

**Work Content**:
Implement GLC service following v2e subprocess pattern:
- Setup logging using pkg/proc/subprocess.SetupLogging()
- Create subprocess using pkg/proc/subprocess.NewSubprocess()
- Register all RPC handlers
- Run with subprocess.RunWithDefaults()
- Configuration for database path, port, logging

**Acceptance Criteria**:
1. WHEN service starts, SHALL initialize subprocess correctly
2. WHEN all RPC handlers are registered, SHALL be ready to receive requests
3. WHEN service runs, SHALL follow v2e subprocess patterns
4. WHEN configuration is loaded, SHALL use provided database path
5. WHEN service starts, SHALL log startup message with service name and version

#### 6.2.2 Database Connection Setup
**File List**:
- `cmd/glc/internal/database/database.go` - Database connection manager
- `cmd/glc/internal/database/migrations.go` - Migration runner

**Work Content**:
Implement database connection and migrations:
- SQLite database connection (using GORM)
- Connection pooling configuration
- WAL mode enablement
- Migration execution on startup
- Connection health checks
- Graceful shutdown handling

**Acceptance Criteria**:
1. WHEN service starts, SHALL connect to SQLite database
2. WHEN WAL mode is enabled, SHALL enable concurrent reads/writes
3. WHEN migrations are run, SHALL create all required tables
4. WHEN connection fails, SHALL log error and exit gracefully
5. WHEN service stops, SHALL close database connection cleanly

#### 6.2.3 Graph CRUD Operations
**File List**:
- `cmd/glc/internal/handlers/graph_handlers.go` - Graph RPC handlers
- `cmd/glc/internal/service/graph_service.go` - Graph business logic

**Work Content**:
Implement graph CRUD RPC handlers:
- CreateGraph: Create new graph with initial data
- SaveGraph: Save/update graph data (nodes, edges, viewport)
- LoadGraph: Load graph by ID with all data
- DeleteGraph: Delete graph by ID
- DuplicateGraph: Duplicate existing graph (create new version)
- ListGraphs: List user's graphs with pagination and filtering
- ShareGraph: Generate unique share token
- LoadGraphByShare: Load graph via share token (read-only)

For each handler:
- Validate request parameters
- Execute business logic
- Handle database transactions
- Return appropriate responses or errors
- Log operations

**Acceptance Criteria**:
1. WHEN CreateGraph is called, SHALL create new graph with unique ID
2. WHEN SaveGraph is called, SHALL update existing graph or create new version
3. WHEN LoadGraph is called, SHALL return complete graph data (nodes, edges, metadata)
4. WHEN DeleteGraph is called, SHALL mark graph as deleted or remove from database
5. WHEN DuplicateGraph is called, SHALL create new graph with same data
6. WHEN ListGraphs is called, SHALL return paginated list matching filters
7. WHEN ShareGraph is called, SHALL generate unique share token
8. WHEN LoadGraphByShare is called with invalid token, SHALL return error
9. WHEN handler fails, SHALL return appropriate error message

#### 6.2.4 Graph Versioning
**File List**:
- `cmd/glc/internal/service/versioning.go` - Graph versioning logic

**Work Content**:
Implement graph versioning system:
- Auto-save versions on significant changes
- Manual version creation
- Version history retrieval
- Version rollback capability
- Version cleanup (keep last N versions)

**Acceptance Criteria**:
1. WHEN graph is saved with major changes, SHALL create new version
2. WHEN user requests version history, SHALL return list of all versions
3. WHEN specific version is loaded, SHALL restore graph to that state
4. WHEN old versions exceed retention limit, SHALL be cleaned up
5. WHEN version is created, SHALL include snapshot timestamp

#### 6.2.5 Custom Preset Management
**File List**:
- `cmd/glc/internal/handlers/preset_handlers.go` - Preset RPC handlers
- `cmd/glc/internal/service/preset_service.go` - Preset business logic

**Work Content**:
Implement custom preset RPC handlers:
- SavePreset: Save/update custom preset
- LoadPreset: Load preset by ID
- ListPresets: List user's presets and built-in presets
- DeletePreset: Delete custom preset
- ValidatePreset: Validate preset structure

**Acceptance Criteria**:
1. WHEN SavePreset is called, SHALL save preset to database
2. WHEN LoadPreset is called with valid ID, SHALL return preset data
3. WHEN ListPresets is called, SHALL return built-in + user presets
4. WHEN DeletePreset is called, SHALL remove from database
5. WHEN ValidatePreset is called, SHALL validate preset structure and return validation result

#### 6.2.6 User Authorization
**File List**:
- `cmd/glc/internal/auth/authorization.go` - Authorization logic
- `cmd/glc/internal/middleware/auth.go` - Auth middleware

**Work Content**:
Implement user authorization:
- Extract user ID from request (from access service)
- Verify user owns graph/preset
- Check read/write permissions
- Handle public graphs (shared via token)
- Log authorization decisions

**Acceptance Criteria**:
1. WHEN user accesses own graph, authorization SHALL succeed
2. WHEN user accesses another user's graph, authorization SHALL fail
3. WHEN user accesses public graph via share token, authorization SHALL succeed (read-only)
4. WHEN authorization fails, SHALL return permission denied error
5. WHEN authorization check is logged, SHALL include user ID, resource ID, and result

---

## Task 6.3: Access Service Integration

### Change Estimation (File Level)
- New files (Go): 2-3
- Modified files (Go): 3-4
- Deleted files: 0

### Cost Estimation (LoC Level)
- Go code: ~400-600 lines
- Configuration: ~100-200 lines

### Detailed Work Items

#### 6.3.1 GLC Endpoints in Access Service
**File List**:
- `cmd/v2access/internal/handlers/glc_handlers.go` - GLC HTTP handlers
- `cmd/v2access/routes/glc_routes.go` - GLC route definitions

**Work Content**:
Add GLC RPC endpoints to access service:
- POST /restful/rpc with GLC methods
- Route to GLC subprocess via broker
- Forward requests and responses
- Handle errors from GLC service
- Log RPC calls

**Acceptance Criteria**:
1. WHEN POST /restful/rpc is called with GLC method, SHALL route to GLC subprocess
2. WHEN GLC service responds, SHALL return response to client
3. WHEN GLC service returns error, SHALL return error to client
4. WHEN RPC call is logged, SHALL include method name, user ID, and status
5. WHEN broker communication fails, SHALL return appropriate error

#### 6.3.2 Broker Configuration
**File List**:
- `config.json` - Updated broker configuration

**Work Content**:
Add GLC service to broker configuration:
- Process definition for glc service
- Command and arguments to start glc
- Restart policy configuration
- Logging configuration
- RPC routing rules

**Acceptance Criteria**:
1. WHEN broker starts, SHALL spawn glc subprocess
2. WHEN glc crashes, SHALL restart based on policy
3. WHEN RPC requests arrive, SHALL route to glc correctly
4. WHEN glc logs, SHALL be captured by broker logging system
5. WHEN configuration is valid, broker SHALL load without errors

---

## Task 6.4: Frontend RPC Client

### Change Estimation (File Level)
- New files (TypeScript): 6-8
- Modified files (TypeScript): 10-15
- Deleted files: 0

### Cost Estimation (LoC Level)
- TypeScript code: ~1,000-1,500 lines
- Tests (TypeScript): ~300-500 lines

### Detailed Work Items

#### 6.4.1 RPC Client Library
**File List**:
- `website/glc/lib/rpc-client.ts` - RPC client implementation
- `website/glc/lib/rpc/types.ts` - RPC type definitions
- `website/glc/lib/rpc/graph-api.ts` - Graph API methods
- `website/glc/lib/rpc/preset-api.ts` - Preset API methods

**Work Content**:
Implement RPC client for GLC backend:
- POST /restful/rpc request handling
- Request/response type definitions matching Go backend
- Authentication token management
- Error handling and retry logic
- Request/response logging (development only)
- Mock mode for development without backend

**Acceptance Criteria**:
1. WHEN RPC client makes request, SHALL POST to /restful/rpc
2. WHEN request succeeds, SHALL return typed response
3. WHEN request fails, SHALL return error with details
4. WHEN authentication token is provided, SHALL include in request
5. WHEN mock mode is enabled, SHALL return mock data
6. WHEN types are defined, SHALL match backend Go structs

#### 6.4.2 Graph Operations Integration
**File List**:
- `website/glc/lib/hooks/use-graph-rpc.ts` - Graph RPC operations hook
- `website/glc/lib/hooks/use-preset-rpc.ts` - Preset RPC operations hook

**Work Content**:
Integrate backend RPC operations into frontend:
- Auto-save to backend on changes
- Save on demand
- Load graph from backend
- List user's graphs
- Delete graph from backend
- Duplicate graph
- Share graph
- Load shared graph
- Optimistic UI updates
- Conflict resolution (if graph changed by another user)

**Acceptance Criteria**:
1. WHEN user creates new graph, SHALL save to backend with generated ID
2. WHEN user makes changes, SHALL auto-save to backend (debounced)
3. WHEN user explicitly saves, SHALL save immediately to backend
4. WHEN user loads graph, SHALL fetch from backend
5. WHEN user lists graphs, SHALL show graphs from backend
6. WHEN save fails, SHALL show error and retry option
7. WHEN graph is shared, SHALL generate shareable URL
8. WHEN backend is unavailable, SHALL fall back to localStorage

#### 6.4.3 Graph Browser UI
**File List**:
- `website/glc/components/glc/graph-browser.tsx` - Graph browser component
- `website/glc/components/glc/graph-item.tsx` - Graph item card
- `website/glc/app/glc/my-graphs/page.tsx` - My graphs page

**Work Content**:
Implement graph browser UI:
- List user's graphs with metadata
- Search and filter graphs
- Pagination for large lists
- Open graph from list
- Delete graph from list
- Duplicate graph from list
- Share graph from list
- Show last modified timestamp
- Show preset used

**Acceptance Criteria**:
1. WHEN user opens My Graphs page, SHALL see list of their graphs
2. WHEN user searches graphs, SHALL filter matching graphs
3. WHEN user opens graph, SHALL load into canvas
4. WHEN user deletes graph, SHALL remove from list after confirmation
5. WHEN user shares graph, SHALL generate shareable link
6. WHEN list is long, SHALL show pagination

#### 6.4.4 Share Link Handling
**File List**:
- `website/glc/lib/hooks/use-shared-graph.ts` - Shared graph hook
- `website/glc/app/glc/shared/[shareToken]/page.tsx` - Shared graph page

**Work Content**:
Implement shared graph loading:
- Parse share token from URL
- Load graph via LoadGraphByShare RPC
- Display shared graph in read-only mode
- Show graph owner and metadata
- Enable "Copy to My Graphs" button

**Acceptance Criteria**:
1. WHEN user opens shared link, SHALL load graph from share token
2. WHEN graph loads, SHALL display in read-only mode
3. WHEN user tries to edit, SHALL prompt to copy to their graphs
4. WHEN user copies graph, SHALL create new graph in their account
5. WHEN share token is invalid, SHALL show error message

#### 6.4.5 Preset Backend Integration
**File List**:
- `website/glc/lib/hooks/use-preset-backend.ts` - Preset backend hook

**Work Content**:
Integrate custom presets with backend:
- Save custom presets to backend
- Load user's custom presets
- List built-in + custom presets
- Delete custom presets
- Sync presets between frontend and backend

**Acceptance Criteria**:
1. WHEN user creates custom preset, SHALL save to backend
2. WHEN user opens preset picker, SHALL show built-in + user presets
3. WHEN user deletes custom preset, SHALL remove from backend
4. WHEN backend is unavailable, SHALL fallback to localStorage
5. WHEN presets are synced, SHALL merge changes correctly

---

## Task 6.5: Testing

### Change Estimation (File Level)
- New files (Go tests): 8-10
- New files (TypeScript tests): 6-8
- Modified files: 5-7
- Deleted files: 0

### Cost Estimation (LoC Level)
- Go test code: ~800-1,200 lines
- TypeScript test code: ~300-500 lines

### Detailed Work Items

#### 6.5.1 Go Unit Tests
**File List**:
- `cmd/glc/internal/models/*_test.go` - Model tests
- `cmd/glc/internal/service/*_test.go` - Service tests
- `cmd/glc/internal/handlers/*_test.go` - Handler tests

**Work Content**:
Write Go unit tests for:
- Data model validation
- Graph CRUD operations
- Preset CRUD operations
- Versioning logic
- Authorization logic
- Database operations (using in-memory SQLite)

**Acceptance Criteria**:
1. WHEN running tests, SHALL pass all tests
2. WHEN checking coverage, SHALL be >80%
3. WHEN tests fail, SHALL show clear error messages
4. WHEN database is mocked, tests SHALL run without external dependencies
5. WHEN authorization is tested, SHALL verify all permission scenarios

#### 6.5.2 Integration Tests
**File List**:
- `tests/glc_integration/test_graph_operations.py` - Graph operations integration tests
- `tests/glc_integration/test_preset_operations.py` - Preset operations integration tests
- `tests/glc_integration/test_concurrent_access.py` - Concurrent access tests

**Work Content**:
Write Python integration tests (using pytest):
- Start broker and glc service
- Test graph CRUD via RPC
- Test preset CRUD via RPC
- Test concurrent graph editing
- Test share token generation and usage
- Test versioning functionality
- Test error handling

**Acceptance Criteria**:
1. WHEN integration tests run, SHALL start broker and glc service
2. WHEN RPC calls are made, SHALL be routed correctly
3. WHEN graph is saved and loaded, data SHALL match
4. WHEN multiple users edit same graph, SHALL handle conflicts
5. WHEN share token is used, SHALL load correct graph
6. WHEN all tests complete, services SHALL shut down cleanly

#### 6.5.3 Frontend RPC Client Tests
**File List**:
- `website/glc/lib/rpc/__tests__/rpc-client.test.ts` - RPC client tests
- `website/glc/lib/hooks/__tests__/use-graph-rpc.test.ts` - Graph hook tests

**Work Content**:
Write TypeScript tests for:
- RPC client request/response handling
- Error handling and retry logic
- Mock mode functionality
- Graph operations hook
- Auto-save logic
- Conflict resolution

**Acceptance Criteria**:
1. WHEN RPC client tests run, SHALL verify request formatting
2. WHEN mock mode is tested, SHALL return correct mock data
3. WHEN auto-save is tested, SHALL debounce correctly
4. WHEN error occurs, SHALL be handled gracefully
5. WHEN tests use mocked responses, SHALL match backend format

---

## Task 6.6: Documentation and Deployment

### Change Estimation (File Level)
- New files: 6-8
- Modified files: 4-5
- Deleted files: 0

### Cost Estimation (LoC Level)
- Documentation: ~800-1,200 lines
- Configuration: ~200-300 lines

### Detailed Work Items

#### 6.6.1 Backend Documentation
**File List**:
- `cmd/glc/README.md` - GLC service documentation
- `cmd/glc/DEPLOYMENT.md` - GLC service deployment guide

**Work Content**:
Create comprehensive backend documentation:
- Service overview
- Architecture description
- RPC API reference (link to service.md)
- Database schema documentation
- Configuration options
- Deployment instructions
- Troubleshooting guide

**Acceptance Criteria**:
1. WHEN documentation is reviewed, SHALL cover all aspects of GLC service
2. WHEN API reference is checked, SHALL match service.md
3. WHEN deployment guide is followed, SHALL successfully deploy service
4. WHEN troubleshooting is needed, guide SHALL help resolve issues

#### 6.6.2 Frontend Integration Documentation
**File List**:
- `website/glc/docs/BACKEND_INTEGRATION.md` - Backend integration guide

**Work Content**:
Create frontend integration documentation:
- RPC client usage
- Authentication setup
- Auto-save behavior
- Offline mode handling
- Error handling
- Mock mode for development
- Migration from localStorage to backend

**Acceptance Criteria**:
1. WHEN developer reads guide, SHALL understand how to use RPC client
2. WHEN authentication is set up, SHALL work correctly
3. WHEN mock mode is used, SHALL simulate backend behavior
4. WHEN migration is needed, guide SHALL explain steps

#### 6.6.3 User Documentation Updates
**File List**:
- `website/glc/docs/USER_GUIDE.md` - Updated user guide
- `website/glc/docs/TUTORIAL_COLLABORATION.md` - Collaboration tutorial

**Work Content**:
Update user documentation for backend integration:
- Save to cloud
- Load saved graphs
- Share graphs
- Collaborative features (read-only)
- Graph history/versions
- My Graphs page

**Acceptance Criteria**:
1. WHEN user reads guide, SHALL understand cloud features
2. WHEN user follows tutorial, SHALL learn to share graphs
3. WHEN user uses My Graphs, SHALL understand all features
4. WHEN documentation is reviewed, SHALL be accurate and up-to-date

#### 6.6.4 Build and Deployment Scripts
**File List**:
- `build.sh` - Updated build script for GLC service
- `cmd/glc/Dockerfile` - Docker container (optional)

**Work Content**:
Update build and deployment:
- Add glc service to build.sh
- Configure glc build flags
- Add glc to deployment process
- Create Docker container (if needed)
- Update environment configuration

**Acceptance Criteria**:
1. WHEN build.sh is run with glc target, SHALL build glc service
2. WHEN glc is deployed, SHALL start as subprocess under broker
3. WHEN configuration is set, SHALL use correct database path
4. WHEN Docker is used, container SHALL build and run correctly

---

## Phase 6 Overall Acceptance Criteria

### Functional Acceptance
1. WHEN user creates graph, SHALL be saved to backend with unique ID
2. WHEN user makes changes, SHALL auto-save to backend (debounced)
3. WHEN user saves explicitly, SHALL save immediately to backend
4. WHEN user loads graph, SHALL fetch from backend
5. WHEN user shares graph, SHALL generate shareable link
6. WHEN user opens shared link, SHALL load shared graph
7. WHEN user creates custom preset, SHALL save to backend
8. WHEN user lists graphs, SHALL see all their graphs from backend

### Backend Acceptance
1. WHEN GLC service starts, SHALL connect to SQLite database
2. WHEN RPC request arrives, SHALL be processed correctly
3. WHEN graph is saved, SHALL be persisted to database
4. WHEN graph is loaded, SHALL return complete data
5. WHEN version is created, SHALL store snapshot
6. WHEN authorization check fails, SHALL return permission error
7. WHEN service crashes, SHALL be restarted by broker

### Frontend Acceptance
1. WHEN RPC client makes request, SHALL send to /restful/rpc
2. WHEN request succeeds, SHALL update UI with response
3. WHEN request fails, SHALL show error to user
4. WHEN backend is unavailable, SHALL fallback to localStorage
5. WHEN graph is loaded from backend, SHALL display in canvas
6. WHEN auto-save triggers, SHALL save changes without disrupting user

### Performance Acceptance
1. WHEN saving graph (100 nodes), SHALL complete in <500ms
2. WHEN loading graph (100 nodes), SHALL complete in <500ms
3. WHEN listing graphs (50 items), SHALL complete in <300ms
4. WHEN auto-save debounces, SHALL wait 2s after last change
5. WHEN backend responds, SHALL handle response in <100ms

### Code Quality Acceptance
1. WHEN running Go tests, SHALL pass all tests
2. WHEN checking Go coverage, SHALL be >80%
3. WHEN running TypeScript tests, SHALL pass all tests
4. WHEN running lint, SHALL have zero errors
5. WHEN reviewing code, SHALL follow v2e patterns and best practices

---

## Phase 6 Deliverables Checklist

### Backend Deliverables
- [ ] GLC service RPC API specification (service.md)
- [ ] Database schema and migrations
- [ ] Go data models
- [ ] GLC service main implementation
- [ ] Database connection setup
- [ ] Graph CRUD RPC handlers
- [ ] Graph versioning system
- [ ] Custom preset management handlers
- [ ] User authorization logic
- [ ] Go unit tests (>80% coverage)
- [ ] Integration tests

### Frontend Deliverables
- [ ] RPC client library
- [ ] Graph operations hook
- [ ] Preset operations hook
- [ ] Graph browser UI
- [ ] Shared graph loading
- [ ] Preset backend integration
- [ ] Frontend RPC client tests

### Integration Deliverables
- [ ] GLC endpoints in access service
- [ ] Broker configuration for GLC service
- [ ] Build script updates
- [ ] Deployment configuration

### Documentation Deliverables
- [ ] GLC service documentation
- [ ] Backend integration guide
- [ ] Updated user documentation
- [ ] Deployment guide

---

## Dependencies

- Phase 5 must be completed before starting Phase 6
- Task 6.1 must be completed before Task 6.2
- Task 6.2 must be completed before Task 6.3
- Task 6.3 must be completed before Task 6.4
- Task 6.4 must be completed before Task 6.5
- Task 6.5 must be completed before Task 6.6
- Frontend from Phases 1-4 must be complete before backend integration

---

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| RPC communication failures between services | High | Implement retry logic, add circuit breakers, log all failures |
| Concurrent graph editing conflicts | High | Implement optimistic updates with conflict resolution, last-write-wins with timestamps |
| Database performance with large graphs | Medium | Add indexes, implement pagination, optimize queries, consider caching |
| Data loss during service crashes | Medium | Use SQLite WAL mode, implement regular backups, graceful shutdown |
| Frontend-backend API version mismatch | Medium | Use semantic versioning, maintain backward compatibility, deprecation warnings |
| Security: unauthorized graph access | High | Implement robust authorization, validate all requests, audit access logs |
| Network latency affecting auto-save | Medium | Debounce saves, show save status, implement offline queue |
| Share token brute-force attacks | Medium | Use cryptographically secure tokens, implement rate limiting, token expiration |

---

## Time Estimation

| Task | Estimated Hours |
|------|-----------------|
| 6.1 Backend Service Design | 12-16 |
| 6.2 GLC Service Implementation | 32-40 |
| 6.3 Access Service Integration | 8-12 |
| 6.4 Frontend RPC Client | 24-32 |
| 6.5 Testing | 20-28 |
| 6.6 Documentation and Deployment | 12-16 |
| **Total** | **108-144** |

---

## Next Steps After Phase 6

### Phase 7: Advanced Collaboration Features
- Real-time collaborative editing (WebSocket/WebRTC)
- User presence indicators
- Comment and annotation system
- Change history comparison
- Merge conflict resolution UI

### Phase 8: Analytics and Insights
- Graph analytics dashboard
- Usage statistics
- Performance metrics
- User behavior analysis

### Phase 9: Advanced D3FEND Features
- D3FEND threat intelligence integration
- Automated attack path analysis
- Vulnerability correlation
- Report generation

---

## Conclusion

Phase 6 completes the backend integration for GLC, enabling persistent storage, multi-user features, and integration with the v2e ecosystem. The system now supports:

- Graph and preset persistence in SQLite database
- RPC-based communication between frontend and backend
- User authentication and authorization
- Graph sharing via share tokens
- Versioning and history tracking
- Auto-save with offline fallback

After completing Phase 6, GLC is a fully functional graph modeling platform with backend storage and can be deployed as part of the v2e system.

**Project Status**: Backend integration complete, ready for advanced collaboration features
