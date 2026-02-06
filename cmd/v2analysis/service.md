
# v2analysis Service (UDA Framework)

## Service Type
RPC (stdin/stdout message passing)

## Description
The v2analysis service implements the UDA (Unified Data Analysis) framework for analyzing security data and creating relationship graphs between different data entities. It maintains an in-memory graph database that represents URN-based connections between CVE, CWE, CAPEC, ATT&CK, and SSG objects.

The service provides readonly access to data from other services for analysis purposes and supports:
- Building and querying URN-based relationship graphs
- Finding paths between security entities
- Monitoring UEE (Unified ETL Engine) status
- Analyzing relationships between vulnerabilities and weaknesses

## Architecture

### Graph Database
- **In-memory storage**: Fast access to relationship data
- **URN-based nodes**: Each node is identified by a unique URN (v2e::provider::type::atomic_id)
- **Directed edges**: Relationships are directional with typed edges
- **Thread-safe**: All graph operations are protected by RWMutex

### Edge Types
- `references`: One entity references another (e.g., CVE references CWE)
- `related_to`: General relationship
- `mitigates`: Mitigation relationship
- `exploits`: Exploitation relationship
- `contains`: Containment relationship

### Technical Implementation

**Graph Structure:**
- Nodes stored in a map: `map[string]*Node` (keyed by URN string)
- Outgoing edges: `map[string][]*Edge` (from URN -> list of edges)
- Incoming edges: `map[string][]*Edge` (to URN -> list of edges for reverse lookup)
- Bidirectional indexing enables efficient neighbor queries

**Concurrency:**
- Uses `sync.RWMutex` for thread-safe operations
- Read operations (GetNode, GetNeighbors) use RLock for concurrent reads
- Write operations (AddNode, AddEdge) use exclusive Lock
- All exported slices are copied to prevent concurrent modification

**Path Finding:**
- Breadth-First Search (BFS) algorithm
- Finds shortest path in directed graph
- Tracks visited nodes to prevent cycles
- Returns ordered path from source to destination

**Data Integration:**
- Communicates with local service via RPC for CVE/CWE/CAPEC/ATT&CK data
- Communicates with meta service via RPC for UEE status
- All data access is readonly to ensure integrity
- Graph can be rebuilt from fresh data at any time

## Available RPC Methods

### 1. RPCGetGraphStats
- **Description**: Returns statistics about the current graph state
- **Request Parameters**: None
- **Response**:
  - `node_count` (int): Total number of nodes in the graph
  - `edge_count` (int): Total number of edges in the graph
- **Errors**: None
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"node_count": 1500, "edge_count": 3200}`

### 2. RPCAddNode
- **Description**: Adds a node to the graph with optional properties
- **Request Parameters**:
  - `urn` (string, required): URN identifier for the node (e.g., "v2e::nvd::cve::CVE-2024-1234")
  - `properties` (object, optional): Arbitrary properties associated with the node
- **Response**:
  - `urn` (string): The URN of the added node
- **Errors**:
  - Invalid URN format: URN does not match expected pattern
- **Example**:
  - **Request**: `{"urn": "v2e::nvd::cve::CVE-2024-1234", "properties": {"severity": "HIGH"}}`
  - **Response**: `{"urn": "v2e::nvd::cve::CVE-2024-1234"}`

### 3. RPCAddEdge
- **Description**: Adds a directed edge between two existing nodes
- **Request Parameters**:
  - `from` (string, required): Source URN
  - `to` (string, required): Destination URN
  - `type` (string, required): Edge type (references, related_to, mitigates, exploits, contains)
  - `properties` (object, optional): Edge properties
- **Response**:
  - `from` (string): Source URN
  - `to` (string): Destination URN
  - `type` (string): Edge type
- **Errors**:
  - Node not found: One or both nodes don't exist in the graph
  - Invalid URN: URN format is invalid
- **Example**:
  - **Request**: `{"from": "v2e::nvd::cve::CVE-2024-1234", "to": "v2e::mitre::cwe::CWE-79", "type": "references"}`
  - **Response**: `{"from": "v2e::nvd::cve::CVE-2024-1234", "to": "v2e::mitre::cwe::CWE-79", "type": "references"}`

### 4. RPCGetNode
- **Description**: Retrieves a node and its properties by URN
- **Request Parameters**:
  - `urn` (string, required): URN of the node to retrieve
- **Response**:
  - `urn` (string): Node URN
  - `properties` (object): Node properties
- **Errors**:
  - Node not found: No node with the specified URN exists
  - Invalid URN: URN format is invalid
- **Example**:
  - **Request**: `{"urn": "v2e::nvd::cve::CVE-2024-1234"}`
  - **Response**: `{"urn": "v2e::nvd::cve::CVE-2024-1234", "properties": {"severity": "HIGH", "description": "..."}}`

### 5. RPCGetNeighbors
- **Description**: Gets all neighboring nodes (both incoming and outgoing connections)
- **Request Parameters**:
  - `urn` (string, required): URN of the node
- **Response**:
  - `neighbors` ([]string): Array of URNs of neighboring nodes
- **Errors**:
  - Invalid URN: URN format is invalid
- **Example**:
  - **Request**: `{"urn": "v2e::nvd::cve::CVE-2024-1234"}`
  - **Response**: `{"neighbors": ["v2e::mitre::cwe::CWE-79", "v2e::mitre::cwe::CWE-89"]}`

### 6. RPCFindPath
- **Description**: Finds a path between two nodes using breadth-first search
- **Request Parameters**:
  - `from` (string, required): Starting URN
  - `to` (string, required): Destination URN
- **Response**:
  - `path` ([]string): Ordered array of URNs representing the path
  - `length` (int): Number of nodes in the path
- **Errors**:
  - No path found: No connection exists between the nodes
  - Invalid URN: One or both URNs are invalid
- **Example**:
  - **Request**: `{"from": "v2e::nvd::cve::CVE-2024-1234", "to": "v2e::mitre::attack::T1566"}`
  - **Response**: `{"path": ["v2e::nvd::cve::CVE-2024-1234", "v2e::mitre::cwe::CWE-79", "v2e::mitre::capec::CAPEC-66", "v2e::mitre::attack::T1566"], "length": 4}`

### 7. RPCGetNodesByType
- **Description**: Retrieves all nodes of a specific resource type
- **Request Parameters**:
  - `type` (string, required): Resource type (cve, cwe, capec, attack, ssg)
- **Response**:
  - `nodes` ([]object): Array of nodes with their URNs and properties
  - `count` (int): Number of nodes returned
- **Errors**:
  - Invalid type: Unsupported resource type
- **Example**:
  - **Request**: `{"type": "cve"}`
  - **Response**: `{"nodes": [{"urn": "v2e::nvd::cve::CVE-2024-1234", "properties": {...}}, ...], "count": 150}`

### 8. RPCGetUEEStatus
- **Description**: Queries the meta service for UEE (Unified ETL Engine) status
- **Request Parameters**: None
- **Response**:
  - Returns the current UEE session data from the meta service
  - Varies based on meta service response format
- **Errors**:
  - Failed to query: Meta service is unavailable or returned an error
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"sessions": [{"id": "cve-123", "status": "running", ...}]}`

### 9. RPCBuildCVEGraph
- **Description**: Builds a graph from CVE data by querying the local service and creating relationships
- **Request Parameters**:
  - `limit` (int, optional): Maximum number of CVEs to process (default: 100)
- **Response**:
  - `nodes_added` (int): Number of nodes added during build
  - `edges_added` (int): Number of edges added during build
  - `total_nodes` (int): Total nodes in graph after build
  - `total_edges` (int): Total edges in graph after build
- **Errors**:
  - Failed to query: Local service is unavailable
  - Parse error: Unable to parse CVE data
- **Example**:
  - **Request**: `{"limit": 200}`
  - **Response**: `{"nodes_added": 250, "edges_added": 180, "total_nodes": 250, "total_edges": 180}`

### 10. RPCClearGraph
- **Description**: Clears all nodes and edges from the graph
- **Request Parameters**: None
- **Response**:
  - `status` (string): "cleared"
- **Errors**: None
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"status": "cleared"}`

### 11. RPCGetFSMState
- **Description**: Returns the current state of the analysis FSM and graph FSM
- **Request Parameters**: None
- **Response**:
  - `analyze_state` (string): Current state of the analysis service (BOOTSTRAPPING, IDLE, PROCESSING, PAUSED, DRAINING, TERMINATED)
  - `graph_state` (string): Current state of the graph (IDLE, BUILDING, ANALYZING, PERSISTING, READY, ERROR)
- **Errors**: None
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"analyze_state": "IDLE", "graph_state": "READY"}`

### 12. RPCPauseAnalysis
- **Description**: Pauses the analysis service (stops accepting new analysis requests)
- **Request Parameters**: None
- **Response**:
  - `status` (string): "paused"
- **Errors**:
  - Invalid state transition: Cannot pause from current state
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"status": "paused"}`

### 13. RPCResumeAnalysis
- **Description**: Resumes the analysis service after being paused
- **Request Parameters**: None
- **Response**:
  - `status` (string): "resumed"
- **Errors**:
  - Invalid state transition: Service is not paused
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"status": "resumed"}`

### 14. RPCSaveGraph
- **Description**: Saves the current graph to disk (BoltDB)
- **Request Parameters**: None
- **Response**:
  - `status` (string): "saved"
  - `node_count` (int): Number of nodes saved
  - `edge_count` (int): Number of edges saved
  - `last_saved` (string): Timestamp of save operation
- **Errors**:
  - Failed to save: Disk write error or permission issue
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"status": "saved", "node_count": 250, "edge_count": 180, "last_saved": "2026-02-06T03:45:00Z"}`

### 15. RPCLoadGraph
- **Description**: Loads the graph from disk, replacing the current in-memory graph
- **Request Parameters**: None
- **Response**:
  - `status` (string): "loaded"
  - `node_count` (int): Number of nodes loaded
  - `edge_count` (int): Number of edges loaded
- **Errors**:
  - Failed to load: No saved graph found or corrupted data
- **Example**:
  - **Request**: `{}`
  - **Response**: `{"status": "loaded", "node_count": 250, "edge_count": 180"}`

---

## URN Format

All nodes in the graph are identified by URNs following this format:

```
v2e::<provider>::<type>::<atomic_id>
```

**Supported Providers:**
- `nvd` - National Vulnerability Database
- `mitre` - MITRE Corporation
- `ssg` - SCAP Security Guide

**Supported Types:**
- `cve` - Common Vulnerabilities and Exposures
- `cwe` - Common Weakness Enumeration
- `capec` - Common Attack Pattern Enumeration and Classification
- `attack` - ATT&CK framework data
- `ssg` - SSG guide data

**Examples:**
- `v2e::nvd::cve::CVE-2024-12233`
- `v2e::mitre::cwe::CWE-79`
- `v2e::mitre::capec::CAPEC-66`
- `v2e::mitre::attack::T1566`
- `v2e::ssg::ssg::rhel9-guide-ospp`

## Usage Patterns

### Building a Graph
1. Clear existing graph: `RPCClearGraph`
2. Build from data source: `RPCBuildCVEGraph` with desired limit
3. Query graph statistics: `RPCGetGraphStats`

**Example workflow via /restful/rpc endpoint:**
```bash
# 1. Clear the graph
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{"method": "RPCClearGraph", "target": "analysis"}'

# 2. Build graph from CVE data (limit 500)
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{"method": "RPCBuildCVEGraph", "target": "analysis", "params": {"limit": 500}}'

# 3. Check graph statistics
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{"method": "RPCGetGraphStats", "target": "analysis"}'
```

### Analyzing Relationships
1. Find related entities: `RPCGetNeighbors` for a specific URN
2. Trace attack paths: `RPCFindPath` between a CVE and ATT&CK technique
3. List all vulnerabilities: `RPCGetNodesByType` with type="cve"

**Example: Finding related weaknesses for a vulnerability**
```bash
# Get neighbors of a CVE
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCGetNeighbors",
    "target": "analysis",
    "params": {"urn": "v2e::nvd::cve::CVE-2024-1234"}
  }'

# Find path between CVE and ATT&CK technique
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCFindPath",
    "target": "analysis",
    "params": {
      "from": "v2e::nvd::cve::CVE-2024-1234",
      "to": "v2e::mitre::attack::T1566"
    }
  }'
```

### Manual Graph Construction
You can also manually build custom graphs:

```bash
# Add a CVE node
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCAddNode",
    "target": "analysis",
    "params": {
      "urn": "v2e::nvd::cve::CVE-2024-1234",
      "properties": {"severity": "HIGH", "description": "XSS vulnerability"}
    }
  }'

# Add a CWE node
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCAddNode",
    "target": "analysis",
    "params": {
      "urn": "v2e::mitre::cwe::CWE-79",
      "properties": {"name": "Cross-site Scripting"}
    }
  }'

# Create a reference edge
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCAddEdge",
    "target": "analysis",
    "params": {
      "from": "v2e::nvd::cve::CVE-2024-1234",
      "to": "v2e::mitre::cwe::CWE-79",
      "type": "references"
    }
  }'
```

### Monitoring UEE
1. Check UEE status: `RPCGetUEEStatus`
2. Build graph after UEE completes data population
3. Perform analysis on fresh data

### Persisting and Loading Graphs
1. Save graph to disk: `RPCSaveGraph`
2. Load graph from disk: `RPCLoadGraph`
3. Graph is automatically loaded on service startup if available
4. Graph is automatically saved on service shutdown

### FSM State Management
1. Check FSM state: `RPCGetFSMState`
2. Pause analysis: `RPCPauseAnalysis`
3. Resume analysis: `RPCResumeAnalysis`
4. FSM ensures proper lifecycle management and resource allocation

## Notes
- Graph operations are in-memory with BoltDB persistence
- Graph is automatically loaded on startup and saved on shutdown
- Service is readonly for external data sources (local, meta services)
- Graph modifications are only through explicit RPC calls
- Thread-safe for concurrent read/write operations
- Service runs as a subprocess managed by the broker
- All communication is broker-mediated via RPC
- FSM-based lifecycle management for reliable operation
- Broker/perf optimizer provides conflict resolution with frontend and ETL services

## Dependencies
- **Subprocess Framework**: `pkg/proc/subprocess` for lifecycle management
- **Graph Database**: `pkg/graph` for in-memory graph operations
- **URN Package**: `pkg/urn` for URN validation and parsing
- **RPC Client**: `pkg/rpc` for communicating with other services
- **FSM Package**: `pkg/analysis/fsm` for state machine management
- **Storage Package**: `pkg/analysis/storage` for BoltDB persistence

## Performance Characteristics

Based on benchmarks:
- **AddNode**: ~150-200 ns/op
- **GetNode**: ~100-150 ns/op (with 1000 nodes)
- **AddEdge**: ~300-400 ns/op
- **GetNeighbors**: ~1-2 µs/op (50 neighbors)
- **FindPath**: ~10-15 µs/op (4-node path)
- **Concurrent reads**: Highly efficient with RWMutex

Graph scales well for typical security data volumes (thousands of nodes, tens of thousands of edges).

## Future Enhancements

See implementation progress and planned features below.

## Implementation Status

### Phase 1: Core Framework ✅ COMPLETE
- [x] Graph database package (`pkg/graph`)
  - In-memory graph with URN-based nodes
  - Directed edges with type support
  - Thread-safe operations with RWMutex
  - BFS path finding algorithm
  - Node filtering by type and provider
  - GetAllNodes and GetAllEdges for persistence
- [x] Analysis service (`cmd/v2analysis`)
  - RPC handlers for graph operations
  - Integration with UEE via meta service
  - Readonly data access from local service
  - Graph building from CVE data
- [x] Unit tests and benchmarks
  - Graph package: 7 tests, all passing
  - Service: 5 tests, all passing
  - Benchmarks show efficient performance
- [x] Documentation (this file)
- [x] Broker integration (added to boot configuration)

### Phase 2: FSM and Persistence ✅ COMPLETE
- [x] GraphFSM state machine (`pkg/analysis/fsm`)
  - State definitions (IDLE, BUILDING, ANALYZING, PERSISTING, READY, ERROR)
  - State transition validation
  - Event system for state changes
- [x] AnalyzeFSM integration (`pkg/analysis/fsm`)
  - Service-level state machine (BOOTSTRAPPING, IDLE, PROCESSING, PAUSED, DRAINING, TERMINATED)
  - Lifecycle management (Start, Pause, Resume, Stop)
  - Event handling from GraphFSM
- [x] BoltDB persistence (`pkg/analysis/storage`)
  - Graph storage and retrieval
  - Metadata tracking (node count, edge count, timestamps)
  - Checkpoint support
  - Automatic load on startup, save on shutdown
- [x] RPC methods for FSM control
  - RPCGetFSMState - query current state
  - RPCPauseAnalysis - pause analysis service
  - RPCResumeAnalysis - resume analysis service
  - RPCSaveGraph - save graph to disk
  - RPCLoadGraph - load graph from disk
- [x] Service integration
  - FSM integrated into analysis service
  - Storage integrated with graph operations
  - Environment variable configuration (GRAPH_DB_PATH)

### Phase 3: Broker/Perf Optimization ✅ COMPLETE
- [x] Analysis optimizer (`cmd/v2broker/perf`)
  - Service priority scheduling
  - Conflict detection (analysis vs frontend vs ETL)
  - Three conflict resolution policies (frontend_first, fair_share, weighted)
  - Service metrics tracking
  - Dynamic throttling based on load
- [x] Broker integration
  - SetAnalysisOptimizer method
  - StartConflictMonitor for automatic conflict resolution
  - Service registration by type and priority
- [x] Resource allocation policies
  - Analysis: Low priority (2 concurrent operations)
  - Frontend: High priority (10 concurrent operations)
  - ETL: Medium priority (5 concurrent operations)

### Phase 4: Planned Enhancements
- [ ] More sophisticated graph algorithms:
  - Centrality measures (PageRank, betweenness)
  - Community detection / clustering
  - Shortest path algorithms (Dijkstra, A*)
- [ ] Real-time graph updates:
  - Listen to UEE events for automatic updates
  - Incremental graph building
- [ ] Graph query capabilities:
  - Cypher-like query language
  - Pattern matching
  - Subgraph extraction
- [ ] Visualization support:
  - Export to GraphML, DOT, JSON formats
  - API for frontend graph rendering
- [ ] Advanced analysis:
  - Temporal graph analysis (time-based relationships)
  - Graph statistics and metrics
  - Anomaly detection in relationships

### Phase 5: Integration & Production Readiness
- [ ] Frontend integration
  - Graph visualization component
  - Interactive graph exploration
- [ ] Performance optimization
  - Graph indexing for faster queries
  - Lazy loading for large graphs
- [ ] Monitoring & observability
  - Graph operation metrics
  - Query performance tracking
- [ ] Documentation
  - API examples
  - Usage patterns
  - Best practices guide
