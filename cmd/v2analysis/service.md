
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

### Analyzing Relationships
1. Find related entities: `RPCGetNeighbors` for a specific URN
2. Trace attack paths: `RPCFindPath` between a CVE and ATT&CK technique
3. List all vulnerabilities: `RPCGetNodesByType` with type="cve"

### Monitoring UEE
1. Check UEE status: `RPCGetUEEStatus`
2. Build graph after UEE completes data population
3. Perform analysis on fresh data

## Notes
- All graph operations are in-memory; graph is not persisted
- Service is readonly for external data sources (local, meta services)
- Graph modifications are only through explicit RPC calls
- Thread-safe for concurrent read/write operations
- Service runs as a subprocess managed by the broker
- All communication is broker-mediated via RPC

## Dependencies
- **Subprocess Framework**: `pkg/proc/subprocess` for lifecycle management
- **Graph Database**: `pkg/graph` for in-memory graph operations
- **URN Package**: `pkg/urn` for URN validation and parsing
- **RPC Client**: `pkg/rpc` for communicating with other services

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

See implementation progress and planned features in the development logs.

Potential improvements:
- Graph persistence to disk
- More sophisticated graph algorithms (centrality, clustering)
- Real-time graph updates from UEE events
- Graph query language (Cypher-like)
- Graph visualization export formats
- Temporal graph analysis (time-based relationships)
