// ============================================================================
// Sysmon Types (from sysmon RPC API)
// ============================================================================

/**
 * System monitoring metrics returned by the sysmon RPC service.
 * Provides real-time CPU, memory, disk, network, and swap usage statistics.
 */
export interface SysMetrics {
  /** CPU usage percentage (0-100) */
  cpuUsage: number;
  /** Memory usage percentage (0-100) */
  memoryUsage: number;
  /** Load average values (1, 5, or 15 minute averages) or single value */
  loadAvg?: number[] | number;
  /** System uptime in seconds */
  uptime?: number;
  /** Disk usage breakdown keyed by mount path */
  disk?: Record<string, { total: number; used: number }>;
  /** Total disk space across all mounts (compatibility field) */
  diskTotal?: number;
  /** Used disk space across all mounts (compatibility field) */
  diskUsage?: number;
  /** Total network bytes received */
  netRx?: number;
  /** Total network bytes transmitted */
  netTx?: number;
  /** Network breakdown by interface name */
  network?: Record<string, { rx: number; tx: number }>;
  /** Swap usage percentage (0-100) */
  swapUsage?: number;
}
/**
 * TypeScript types mirroring Go structs from the v2e backend
 * Generated from pkg/cve/types.go and RPC API specifications
 */

// ============================================================================
// CVE Data Types (from pkg/cve/types.go)
// ============================================================================

/**
 * Description text with language identifier.
 * Used for CVE descriptions and other multi-language content.
 */
export interface Description {
  /** Language code (e.g., "en", "es") */
  lang: string;
  /** Description text content */
  value: string;
}

/**
 * Tags associated with a CVE entry.
 * Provides additional metadata from the source identifier.
 */
export interface CVETag {
  /** Source that assigned the tags */
  sourceIdentifier: string;
  /** Array of tag strings */
  tags?: string[];
}

/**
 * Weakness (CWE) associated with a CVE.
 * Links a CVE to its related Common Weakness Enumeration.
 */
export interface Weakness {
  /** Source of the weakness reference */
  source: string;
  /** Type/Category of the weakness */
  type: string;
  /** Descriptions of the weakness in various languages */
  description: Description[];
}

/**
 * CPE (Common Platform Enumeration) match criteria.
 * Defines vulnerable and non-vulnerable software versions.
 */
export interface CPEMatch {
  /** Whether this CPE match is vulnerable */
  vulnerable: boolean;
  /** CPE criteria string */
  criteria: string;
  /** Unique identifier for this match criteria */
  matchCriteriaId: string;
  /** Version start (excluding) */
  versionStartExcluding?: string;
  /** Version start (including) */
  versionStartIncluding?: string;
  /** Version end (excluding) */
  versionEndExcluding?: string;
  /** Version end (including) */
  versionEndIncluding?: string;
}

/**
 * Configuration node for CPE matching.
 * Represents a single node in the vulnerability configuration tree.
 */
export interface Node {
  /** Logical operator (AND, OR) */
  operator: string;
  /** Whether to negate the result */
  negate?: boolean;
  /** Array of CPE matches for this node */
  cpeMatch: CPEMatch[];
}

/**
 * Vulnerability configuration container.
 * Holds the configuration tree for a CVE.
 */
export interface Config {
  /** Logical operator for combining nodes */
  operator?: string;
  /** Whether to negate the result */
  negate?: boolean;
  /** Array of child nodes */
  nodes: Node[];
}

/**
 * External reference for a CVE.
 * Links to advisories, patches, and other related resources.
 */
export interface Reference {
  /** URL of the reference */
  url: string;
  /** Source of the reference */
  source?: string;
  /** Tags describing the reference type */
  tags?: string[];
}

/**
 * Vendor comment associated with a CVE.
 * Provides official vendor statements about the vulnerability.
 */
export interface VendorComment {
  /** Organization providing the comment */
  organization: string;
  /** Comment text content */
  comment: string;
  /** ISO timestamp of last modification */
  lastModified: string;
}

/**
 * CVSS v3.x score data.
 * Contains the Common Vulnerability Scoring System version 3 metrics.
 */
export interface CVSSDataV3 {
  /** CVSS version (e.g., "3.1") */
  version: string;
  /** CVSS vector string */
  vectorString: string;
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Base severity rating */
  baseSeverity: string;
  /** Attack vector (NETWORK, ADJACENT, LOCAL, PHYSICAL) */
  attackVector?: string;
  /** Attack complexity (LOW, HIGH) */
  attackComplexity?: string;
  /** Privileges required */
  privilegesRequired?: string;
  /** User interaction required */
  userInteraction?: string;
  /** Scope (UNCHANGED, CHANGED) */
  scope?: string;
  /** Confidentiality impact */
  confidentialityImpact?: string;
  /** Integrity impact */
  integrityImpact?: string;
  /** Availability impact */
  availabilityImpact?: string;
  /** Exploitability score */
  exploitabilityScore?: number;
  /** Impact score */
  impactScore?: number;
}

/**
 * CVSS v3.x metric wrapper.
 * Combines CVSS data with source information.
 */
export interface CVSSMetricV3 {
  /** Source of the CVSS data */
  source: string;
  /** Type of CVSS score */
  type: string;
  /** CVSS v3 data */
  cvssData: CVSSDataV3;
  /** Exploitability score */
  exploitabilityScore?: number;
  /** Impact score */
  impactScore?: number;
}

/**
 * CVSS v2 score data.
 * Contains the Common Vulnerability Scoring System version 2 metrics.
 */
export interface CVSSDataV2 {
  /** CVSS version (e.g., "2.0") */
  version: string;
  /** CVSS vector string */
  vectorString: string;
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Access vector */
  accessVector?: string;
  /** Access complexity */
  accessComplexity?: string;
  /** Authentication required */
  authentication?: string;
  /** Confidentiality impact */
  confidentialityImpact?: string;
  /** Integrity impact */
  integrityImpact?: string;
  /** Availability impact */
  availabilityImpact?: string;
}

/**
 * CVSS v2 metric wrapper.
 * Combines CVSS v2 data with source information.
 */
export interface CVSSMetricV2 {
  /** Source of the CVSS data */
  source: string;
  /** Type of CVSS score */
  type: string;
  /** CVSS v2 data */
  cvssData: CVSSDataV2;
  /** Base severity rating */
  baseSeverity?: string;
  /** Exploitability score */
  exploitabilityScore?: number;
  /** Impact score */
  impactScore?: number;
}

/**
 * CVSS v4.0 score data.
 * Contains the Common Vulnerability Scoring System version 4.0 metrics.
 */
export interface CVSSDataV40 {
  /** CVSS version (e.g., "4.0") */
  version: string;
  /** CVSS vector string */
  vectorString: string;
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Base severity rating */
  baseSeverity: string;
  /** Attack vector */
  attackVector?: string;
  /** Attack complexity */
  attackComplexity?: string;
  /** Attack requirements */
  attackRequirements?: string;
  /** Privileges required */
  privilegesRequired?: string;
  /** User interaction required */
  userInteraction?: string;
  /** Vulnerable system confidentiality impact */
  vulnConfidentialityImpact?: string;
  /** Vulnerable system integrity impact */
  vulnIntegrityImpact?: string;
  /** Vulnerable system availability impact */
  vulnAvailabilityImpact?: string;
  /** Subsequent system confidentiality impact */
  subConfidentialityImpact?: string;
  /** Subsequent system integrity impact */
  subIntegrityImpact?: string;
  /** Subsequent system availability impact */
  subAvailabilityImpact?: string;
}

/**
 * CVSS v4.0 metric wrapper.
 * Combines CVSS v4.0 data with source information.
 */
export interface CVSSMetricV40 {
  /** Source of the CVSS data */
  source: string;
  /** Type of CVSS score */
  type: string;
  /** CVSS v4.0 data */
  cvssData: CVSSDataV40;
}

/**
 * Collection of CVSS metrics for a CVE.
 * Contains all available CVSS version scores.
 */
export interface Metrics {
  /** CVSS v4.0 metrics */
  cvssMetricV40?: CVSSMetricV40[];
  /** CVSS v3.1 metrics */
  cvssMetricV31?: CVSSMetricV3[];
  /** CVSS v3.0 metrics */
  cvssMetricV30?: CVSSMetricV3[];
  /** CVSS v2.0 metrics */
  cvssMetricV2?: CVSSMetricV2[];
}

/**
 * Complete CVE (Common Vulnerabilities and Exposures) item.
 * Contains all data for a single CVE entry.
 */
export interface CVEItem {
  /** CVE identifier (e.g., "CVE-2021-1234") */
  id: string;
  /** Organization that created the CVE */
  sourceIdentifier: string;
  /** ISO timestamp when CVE was published */
  published: string;
  /** ISO timestamp of last modification */
  lastModified: string;
  /** Current vulnerability status */
  vulnStatus: string;
  /** Evaluator comment on the vulnerability */
  evaluatorComment?: string;
  /** Evaluator suggested solution */
  evaluatorSolution?: string;
  /** Evaluator impact assessment */
  evaluatorImpact?: string;
  /** CISA exploit addition date */
  cisaExploitAdd?: string;
  /** CISA action due date */
  cisaActionDue?: string;
  /** CISA required action */
  cisaRequiredAction?: string;
  /** CISA vulnerability name */
  cisaVulnerabilityName?: string;
  /** Tags associated with this CVE */
  cveTags?: CVETag[];
  /** CVE descriptions in multiple languages */
  descriptions: Description[];
  /** CVSS severity metrics */
  metrics?: Metrics;
  /** Associated weakness (CWE) references */
  weaknesses?: Weakness[];
  /** CPE configuration for vulnerable products */
  configurations?: Config[];
  /** External references and advisories */
  references?: Reference[];
  /** Vendor comments */
  vendorComments?: VendorComment[];
}

/**
 * CVE API response wrapper.
 * Contains paginated CVE data from NVD API.
 */
export interface CVEResponse {
  /** Number of results per page */
  resultsPerPage: number;
  /** Starting index of results */
  startIndex: number;
  /** Total number of results available */
  totalResults: number;
  /** Response format */
  format: string;
  /** API version */
  version: string;
  /** Response timestamp */
  timestamp: string;
  /** Array of CVE items */
  vulnerabilities: Array<{
    cve: CVEItem;
  }>;
}

// ============================================================================
// RPC Request/Response Types
// ============================================================================

/**
 * Generic RPC request wrapper.
 * Used for all RPC calls to the backend.
 */
export interface RPCRequest<T = unknown> {
  /** RPC method name to call */
  method: string;
  /** Method parameters (typed) */
  params?: T;
  /** Target service for the request */
  target?: string;
}

/**
 * Generic RPC response wrapper.
 * All RPC responses follow this structure.
 */
export interface RPCResponse<T = unknown> {
  /** Return code (0 = success) */
  retcode: number;
  /** Response message */
  message: string;
  /** Response payload data */
  payload: T | null;
}

// ============================================================================
// CVE Meta Service RPC Types
// ============================================================================

/**
 * Request to get a single CVE by ID.
 */
export interface GetCVERequest {
  /** CVE identifier (e.g., "CVE-2021-1234") */
  cveId: string;
}

/**
 * Response containing CVE data with source indication.
 */
export interface GetCVEResponse {
  /** Complete CVE item */
  cve: CVEItem;
  /** Source of the data */
  source: 'local' | 'remote';
}

/**
 * Request to create/fetch a new CVE entry.
 */
export interface CreateCVERequest {
  /** CVE identifier to create */
  cveId: string;
}

/**
 * Response after creating a CVE entry.
 */
export interface CreateCVEResponse {
  /** Whether creation was successful */
  success: boolean;
  /** CVE identifier */
  cveId: string;
  /** Created CVE item */
  cve: CVEItem;
}

/**
 * Request to update an existing CVE entry.
 */
export interface UpdateCVERequest {
  /** CVE identifier to update */
  cveId: string;
}

/**
 * Response after updating a CVE entry.
 */
export interface UpdateCVEResponse {
  /** Whether update was successful */
  success: boolean;
  /** CVE identifier */
  cveId: string;
  /** Updated CVE item */
  finalSeverity?: CVSSSeverity;
  cve: CVEItem;
}

/**
 * Request to delete a CVE entry.
 */
export interface DeleteCVERequest {
  /** CVE identifier to delete */
  cveId: string;
}

/**
 * Response after deleting a CVE entry.
 */
export interface DeleteCVEResponse {
  /** Whether deletion was successful */
  success: boolean;
  /** Deleted CVE identifier */
  cveId: string;
}

/**
 * Request to list CVEs with pagination.
 */
export interface ListCVEsRequest {
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing paginated CVE list.
 */
export interface ListCVEsResponse {
  /** Array of CVE items */
  cves: CVEItem[];
  /** Total number of CVEs available */
  total: number;
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
}

/**
 * Response containing total CVE count.
 */
export interface CountCVEsResponse {
  /** Total number of CVEs in database */
  count: number;
}

// ============================================================================
// Job Session Types
// ============================================================================

/**
 * Request to start a new job session.
 */
export interface StartSessionRequest {
  /** Unique session identifier */
  sessionId: string;
  /** Starting index for data fetching */
  startIndex?: number;
  /** Number of results per batch */
  resultsPerBatch?: number;
}

/**
 * Response after starting a job session.
 */
export interface StartSessionResponse {
  /** Whether session started successfully */
  success: boolean;
  /** Session identifier */
  sessionId: string;
  /** Current session state */
  state: string;
  /** ISO timestamp when session was created */
  createdAt: string;
}

/**
 * Response after stopping a job session.
 */
export interface StopSessionResponse {
  /** Whether stop was successful */
  success: boolean;
  /** Session identifier */
  sessionId: string;
  /** Number of items fetched */
  fetchedCount: number;
  /** Number of items stored */
  storedCount: number;
  /** Number of errors encountered */
  errorCount: number;
}

/**
 * Status of a running job session.
 */
export interface SessionStatus {
  /** Whether a session is active */
  hasSession: boolean;
  /** Session identifier */
  sessionId?: string;
  /** Current session state */
  state?: string;
  /** Data type being processed */
  dataType?: string;
  /** Starting index for data fetching */
  startIndex?: number;
  /** Number of results per batch */
  resultsPerBatch?: number;
  /** ISO timestamp when session was created */
  createdAt?: string;
  /** ISO timestamp of last update */
  updatedAt?: string;
  /** Number of items fetched */
  fetchedCount?: number;
  /** Number of items stored */
  storedCount?: number;
  /** Number of errors encountered */
  errorCount?: number;
  /** Error message if session failed */
  errorMessage?: string;
  /** Progress breakdown by data source */
  progress?: Record<string, DataProgress>;
  /** Additional session parameters */
  params?: Record<string, unknown>;
}

/**
 * Progress data for a single data source.
 */
export interface DataProgress {
  /** Total items to process */
  totalCount: number;
  /** Number of items processed */
  processedCount: number;
  /** Number of errors encountered */
  errorCount: number;
  /** ISO timestamp when processing started */
  startTime: string;
  /** ISO timestamp of last update */
  lastUpdate: string;
  /** Error message if processing failed */
  errorMessage?: string;
}

/**
 * Response after pausing a job.
 */
export interface PauseJobResponse {
  /** Whether pause was successful */
  success: boolean;
  /** Current job state after pause */
  state: string;
}

/**
 * Response after resuming a job.
 */
export interface ResumeJobResponse {
  /** Whether resume was successful */
  success: boolean;
  /** Current job state after resume */
  state: string;
}

// ============================================================================
// CWE View Job RPC Types
// ============================================================================

/**
 * Request to start a CWE view import job.
 */
export interface StartCWEViewJobRequest {
  /** Optional session identifier */
  sessionId?: string;
  /** Starting index for data fetching */
  startIndex?: number;
  /** Number of results per batch */
  resultsPerBatch?: number;
}

/**
 * Response after starting a CWE view job.
 */
export interface StartCWEViewJobResponse {
  /** Whether job started successfully */
  success: boolean;
  /** Session identifier for the job */
  sessionId: string;
}

/**
 * Response after stopping a CWE view job.
 */
export interface StopCWEViewJobResponse {
  /** Whether stop was successful */
  success: boolean;
  /** Session identifier that was stopped */
  sessionId?: string;
}

// ============================================================================
// Utility Types
// ============================================================================

/**
 * Possible states for a job session.
 */
export type JobState = 'idle' | 'running' | 'paused';

/**
 * Health check response from backend services.
 */
export interface HealthResponse {
  /** Service health status */
  status: string;
}

// ============================================================================
// Graph Analysis RPC Types (from cmd/v2analysis/service.md)
// ============================================================================

/**
 * Statistics about a graph.
 */
export interface GraphStats {
  /** Total number of nodes in the graph */
  node_count: number;
  /** Total number of edges in the graph */
  edge_count: number;
}

/**
 * A node in the analysis graph.
 */
export interface GraphNode {
  /** Uniform Resource Name for the node */
  urn: string;
  /** Node properties as key-value pairs */
  properties: Record<string, unknown>;
}

/**
 * An edge in the analysis graph.
 */
export interface GraphEdge {
  /** Source node URN */
  from: string;
  /** Target node URN */
  to: string;
  /** Edge type/relationship */
  type: string;
  /** Optional edge properties */
  properties?: Record<string, unknown>;
}

/**
 * A path through the graph.
 */
export interface GraphPath {
  /** Array of node URNs forming the path */
  path: string[];
  /** Number of edges in the path */
  length: number;
}

/**
 * Request to get graph statistics.
 */
export interface GetGraphStatsRequest {}

/**
 * Response containing graph statistics.
 */
export interface GetGraphStatsResponse {
  /** Total number of nodes in the graph */
  node_count: number;
  /** Total number of edges in the graph */
  edge_count: number;
}

/**
 * Request to add a node to the graph.
 */
export interface AddNodeRequest {
  /** URN for the new node */
  urn: string;
  /** Optional node properties */
  properties?: Record<string, unknown>;
}

/**
 * Response after adding a node.
 */
export interface AddNodeResponse {
  /** URN of the added node */
  urn: string;
}

/**
 * Request to add an edge to the graph.
 */
export interface AddEdgeRequest {
  /** Source node URN */
  from: string;
  /** Target node URN */
  to: string;
  /** Edge type/relationship */
  type: string;
  /** Optional edge properties */
  properties?: Record<string, unknown>;
}

/**
 * Response after adding an edge.
 */
export interface AddEdgeResponse {
  /** Source node URN */
  from: string;
  /** Target node URN */
  to: string;
  /** Edge type */
  type: string;
}

/**
 * Request to get a node by URN.
 */
export interface GetNodeRequest {
  /** URN of the node to retrieve */
  urn: string;
}

/**
 * Response containing node data.
 */
export interface GetNodeResponse {
  /** URN of the node */
  urn: string;
  /** Node properties */
  properties: Record<string, unknown>;
}

/**
 * Request to get neighbors of a node.
 */
export interface GetNeighborsRequest {
  /** URN of the node */
  urn: string;
}

/**
 * Response containing neighboring node URNs.
 */
export interface GetNeighborsResponse {
  /** Array of neighbor URNs */
  neighbors: string[];
}

/**
 * Request to find a path between nodes.
 */
export interface FindPathRequest {
  /** Source node URN */
  from: string;
  /** Target node URN */
  to: string;
}

/**
 * Response containing the found path.
 */
export interface FindPathResponse {
  /** Array of node URNs forming the path */
  path: string[];
  /** Number of edges in the path */
  length: number;
}

/**
 * Request to get nodes by type.
 */
export interface GetNodesByTypeRequest {
  /** Node type to filter by */
  type: string;
}

/**
 * Response containing nodes of the specified type.
 */
export interface GetNodesByTypeResponse {
  /** Array of matching nodes */
  nodes: Array<{
    urn: string;
    properties: Record<string, unknown>;
  }>;
  /** Number of nodes found */
  count: number;
}

/**
 * Request to build a CVE graph.
 */
export interface BuildCVEGraphRequest {
  /** Optional limit on number of CVEs to process */
  limit?: number;
}

/**
 * Response after building CVE graph.
 */
export interface BuildCVEGraphResponse {
  /** Number of nodes added */
  nodes_added: number;
  /** Number of edges added */
  edges_added: number;
  /** Total nodes in graph after build */
  total_nodes: number;
  /** Total edges in graph after build */
  total_edges: number;
}

/**
 * Request to clear the graph.
 */
export interface ClearGraphRequest {}

/**
 * Response after clearing the graph.
 */
export interface ClearGraphResponse {
  /** Status message */
  status: string;
}

/**
 * Request to get FSM (Finite State Machine) state.
 */
export interface GetFSMStateRequest {}

/**
 * Response containing FSM state information.
 */
export interface GetFSMStateResponse {
  /** Analysis FSM state */
  analyze_state: string;
  /** Graph FSM state */
  graph_state: string;
}

/**
 * Request to pause graph analysis.
 */
export interface PauseAnalysisRequest {}

/**
 * Response after pausing analysis.
 */
export interface PauseAnalysisResponse {
  /** Status message */
  status: string;
}

/**
 * Request to resume graph analysis.
 */
export interface ResumeAnalysisRequest {}

/**
 * Response after resuming analysis.
 */
export interface ResumeAnalysisResponse {
  /** Status message */
  status: string;
}

/**
 * Request to save the graph to disk.
 */
export interface SaveGraphRequest {}

/**
 * Response after saving graph.
 */
export interface SaveGraphResponse {
  /** Status message */
  status: string;
  /** Number of nodes saved */
  node_count: number;
  /** Number of edges saved */
  edge_count: number;
  /** ISO timestamp when saved */
  last_saved: string;
}

/**
 * Request to load the graph from disk.
 */
export interface LoadGraphRequest {}

/**
 * Response after loading graph.
 */
export interface LoadGraphResponse {
  /** Status message */
  status: string;
  /** Number of nodes loaded */
  node_count: number;
  /** Number of edges loaded */
  edge_count: number;
}

// ============================================================================
// CWE Data Types (from pkg/cwe/types.go)
// ============================================================================

/**
 * Complete CWE (Common Weakness Enumeration) item.
 * Contains all data for a single weakness entry.
 */
export interface CWEItem {
  /** CWE identifier (e.g., "CWE-79") */
  id: string;
  /** Weakness name */
  name: string;
  /** Taxonomic classification diagram */
  diagram?: string;
  /** Abstraction level (Class, Base, Variant) */
  abstraction: string;
  /** Structure (Simple, Composite) */
  structure: string;
  /** Status (Draft, Incomplete, Deprecated, etc.) */
  status: string;
  /** Weakness description */
  description: string;
  /** Extended description */
  extendedDescription?: string;
  /** Likelihood of exploit rating */
  likelihoodOfExploit?: string;
  /** Related weaknesses */
  relatedWeaknesses?: RelatedWeakness[];
  /** Weakness ordinalities */
  weaknessOrdinalities?: WeaknessOrdinality[];
  /** Applicable platforms */
  applicablePlatforms?: ApplicablePlatform[];
  /** Background details */
  backgroundDetails?: string[];
  /** Alternate terms for this weakness */
  alternateTerms?: AlternateTerm[];
  /** Modes of introduction */
  modesOfIntroduction?: ModeOfIntroduction[];
  /** Common consequences */
  commonConsequences?: Consequence[];
  /** Detection methods */
  detectionMethods?: DetectionMethod[];
  /** Potential mitigations */
  potentialMitigations?: Mitigation[];
  /** Demonstrative examples */
  demonstrativeExamples?: DemonstrativeExample[];
  /** Observed real-world examples */
  observedExamples?: ObservedExample[];
  /** Functional areas affected */
  functionalAreas?: string[];
  /** Resources affected */
  affectedResources?: string[];
  /** Taxonomy mappings */
  taxonomyMappings?: TaxonomyMapping[];
  /** Related ATT&CK patterns */
  relatedAttackPatterns?: string[];
  /** External references */
  references?: Reference[];
  /** Mapping notes */
  mappingNotes?: MappingNotes;
  /** Additional notes */
  notes?: Note[];
  /** Content history */
  contentHistory?: ContentHistory[];
}

// ============================================================================
// ASVS Data Types (from pkg/asvs/types.go)
// ============================================================================

/**
 * ASVS (Application Security Verification Standard) requirement item.
 * Represents a single security verification requirement.
 */
export interface ASVSItem {
  /** Unique requirement identifier */
  requirementID: string;
  /** Chapter number/title */
  chapter: string;
  /** Section number/title */
  section: string;
  /** Requirement description */
  description: string;
  /** Whether requirement applies to Level 1 */
  level1: boolean;
  /** Whether requirement applies to Level 2 */
  level2: boolean;
  /** Whether requirement applies to Level 3 */
  level3: boolean;
  /** Related CWE identifier */
  cwe?: string;
}

// ============================================================================
// CAPEC Data Types
// ============================================================================

/**
 * Weakness reference in a CAPEC entry.
 */
export interface CAPECRelatedWeakness {
  /** CWE identifier */
  cweId?: string;
}

/**
 * CAPEC (Common Attack Pattern Enumeration and Classification) item.
 * Represents a single attack pattern.
 */
export interface CAPECItem {
  /** CAPEC identifier (e.g., "CAPEC-123") */
  id: string;
  /** Attack pattern name */
  name: string;
  /** Brief summary of the attack */
  summary?: string;
  /** Detailed description */
  description?: string;
  /** CAPEC entry status */
  status?: string;
  /** Likelihood of attack */
  likelihood?: string;
  /** Typical severity level */
  typicalSeverity?: string;
  /** Related weaknesses */
  relatedWeaknesses?: CAPECRelatedWeakness[];
  /** External references */
  references?: Reference[];
}

// ATT&CK Types
/**
 * ATT&CK technique entry.
 * Represents a specific adversarial technique.
 */
export interface AttackTechnique {
  /** Technique ID (e.g., "T1001") */
  id: string;
  /** Technique name */
  name: string;
  /** Technique description */
  description?: string;
  /** ATT&CK domain (enterprise, mobile, ics) */
  domain?: string;
  /** Target platforms */
  platform?: string;
  /** ISO timestamp when created */
  created?: string;
  /** ISO timestamp when last modified */
  modified?: string;
  /** Whether technique is revoked */
  revoked?: boolean;
  /** Whether technique is deprecated */
  deprecated?: boolean;
}

/**
 * ATT&CK tactic (also known as phase/matrix).
 * Represents a high-level category of techniques.
 */
export interface AttackTactic {
  /** Tactic ID (e.g., "TA0001") */
  id: string;
  /** Tactic name */
  name: string;
  /** Tactic description */
  description?: string;
  /** ATT&CK domain */
  domain?: string;
  /** ISO timestamp when created */
  created?: string;
  /** ISO timestamp when last modified */
  modified?: string;
}

/**
 * ATT&CK mitigation entry.
 * Represents a security mitigation for techniques.
 */
export interface AttackMitigation {
  /** Mitigation ID (e.g., "M1001") */
  id: string;
  /** Mitigation name */
  name: string;
  /** Mitigation description */
  description?: string;
  /** ATT&CK domain */
  domain?: string;
  /** ISO timestamp when created */
  created?: string;
  /** ISO timestamp when last modified */
  modified?: string;
}

/**
 * ATT&CK software entry.
 * Represents malware, tools, or other software used by adversaries.
 */
export interface AttackSoftware {
  /** Software ID (e.g., "S0001") */
  id: string;
  /** Software name */
  name: string;
  /** Software description */
  description?: string;
  /** Software type (malware, tool) */
  type?: string;
  /** ATT&CK domain */
  domain?: string;
  /** ISO timestamp when created */
  created?: string;
  /** ISO timestamp when last modified */
  modified?: string;
}

/**
 * ATT&CK group entry.
 * Represents threat actor groups (also known as intrusion sets).
 */
export interface AttackGroup {
  /** Group ID (e.g., "G0001") */
  id: string;
  /** Group name/alias */
  name: string;
  /** Group description */
  description?: string;
  /** ATT&CK domain */
  domain?: string;
  /** ISO timestamp when created */
  created?: string;
  /** ISO timestamp when last modified */
  modified?: string;
}

/**
 * Response containing ATT&CK data with pagination.
 */
export interface AttackListResponse {
  /** Array of techniques */
  techniques?: AttackTechnique[];
  /** Array of tactics */
  tactics?: AttackTactic[];
  /** Array of mitigations */
  mitigations?: AttackMitigation[];
  /** Pagination offset */
  offset: number;
  /** Pagination limit */
  limit: number;
  /** Total number of results */
  total: number;
}

/**
 * Related weakness reference in CWE.
 */
export interface RelatedWeakness {
  /** Nature of relationship (ChildOf, PeerOf, etc.) */
  nature: string;
  /** CWE identifier */
  cweId: string;
  /** View identifier */
  viewId: string;
  /** Ordinal position */
  ordinal?: string;
}

/**
 * Weakness ordinality information.
 */
export interface WeaknessOrdinality {
  /** Ordinality value */
  ordinality: string;
  /** Description of ordinality */
  description?: string;
}

/**
 * Platform where the weakness can occur.
 */
export interface ApplicablePlatform {
  /** Platform type (OS, Hardware, Language, etc.) */
  type: string;
  /** Platform name */
  name?: string;
  /** Platform class */
  class?: string;
  /** How common the platform is */
  prevalence: string;
}

/**
 * Alternative term for a weakness.
 */
export interface AlternateTerm {
  /** Alternate term */
  term: string;
  /** Description of the term */
  description?: string;
}

/**
 * When a weakness can be introduced.
 */
export interface ModeOfIntroduction {
  /** Phase of introduction (Architecture, Implementation, etc.) */
  phase: string;
  /** Additional notes */
  note?: string;
}

/**
 * Consequence of exploiting a weakness.
 */
export interface Consequence {
  /** Impact scope (Technical, Business) */
  scope: string[];
  /** Type of impact */
  impact?: string[];
  /** Likelihood of consequence */
  likelihood?: string[];
  /** Additional notes */
  note?: string;
}

/**
 * Method for detecting a weakness.
 */
export interface DetectionMethod {
  /** Detection method identifier */
  detectionMethodId?: string;
  /** Detection method name */
  method: string;
  /** Description of the method */
  description: string;
  /** Effectiveness rating */
  effectiveness?: string;
  /** Additional effectiveness notes */
  effectivenessNotes?: string;
}

/**
 * Mitigation strategy for a weakness.
 */
export interface Mitigation {
  /** Mitigation identifier */
  mitigationId?: string;
  /** Phases where mitigation applies */
  phase?: string[];
  /** Mitigation strategy */
  strategy: string;
  /** Description of mitigation */
  description: string;
  /** Effectiveness rating */
  effectiveness?: string;
  /** Additional effectiveness notes */
  effectivenessNotes?: string;
}

/**
 * Demonstrative example of a weakness.
 */
export interface DemonstrativeExample {
  /** Example identifier */
  id?: string;
  /** Example content entries */
  entries: DemonstrativeEntry[];
}

/**
 * Single entry in a demonstrative example.
 */
export interface DemonstrativeEntry {
  /** Introductory text */
  introText?: string;
  /** Body text content */
  bodyText?: string;
  /** Nature of the entry */
  nature?: string;
  /** Programming language */
  language?: string;
  /** Example code snippet */
  exampleCode?: string;
  /** Reference link */
  reference?: string;
}

/**
 * Real-world observed example of a weakness.
 */
export interface ObservedExample {
  /** Reference identifier */
  reference: string;
  /** Description of the example */
  description: string;
  /** Link to more information */
  link: string;
}

/**
 * Mapping to external taxonomy.
 */
export interface TaxonomyMapping {
  /** Taxonomy name (e.g., "ISO 27001") */
  taxonomyName: string;
  /** Entry name in taxonomy */
  entryName?: string;
  /** Entry identifier in taxonomy */
  entryId?: string;
  /** How well the mapping fits */
  mappingFit?: string;
}

/**
 * Mapping notes for taxonomy relationships.
 */
export interface MappingNotes {
  /** Usage description */
  usage: string;
  /** Mapping rationale */
  rationale: string;
  /** Additional comments */
  comments: string;
  /** Reasons for mapping */
  reasons: string[];
  /** Suggestions for improvement */
  suggestions?: SuggestionComment[];
}

/**
 * Suggestion for mapping improvement.
 */
export interface SuggestionComment {
  /** Suggestion text */
  comment: string;
  /** Related CWE identifier */
  cweId: string;
}

/**
 * General note attached to an entry.
 */
export interface Note {
  /** Note type */
  type: string;
  /** Note content */
  note: string;
}

/**
 * Content history entry for tracking changes.
 */
export interface ContentHistory {
  /** Type of history entry */
  type: string;
  /** Submitter name */
  submissionName?: string;
  /** Submitter organization */
  submissionOrganization?: string;
  /** Submission date */
  submissionDate?: string;
  /** Submission version */
  submissionVersion?: string;
  /** Submission release date */
  submissionReleaseDate?: string;
  /** Submission comment */
  submissionComment?: string;
  /** Modifier name */
  modificationName?: string;
  /** Modifier organization */
  modificationOrganization?: string;
  /** Modification date */
  modificationDate?: string;
  /** Modification version */
  modificationVersion?: string;
  /** Modification release date */
  modificationReleaseDate?: string;
  /** Modification comment */
  modificationComment?: string;
  /** Contributor name */
  contributionName?: string;
  /** Contributor organization */
  contributionOrganization?: string;
  /** Contribution date */
  contributionDate?: string;
  /** Contribution version */
  contributionVersion?: string;
  /** Contribution release date */
  contributionReleaseDate?: string;
  /** Contribution comment */
  contributionComment?: string;
  /** Contribution type */
  contributionType?: string;
  /** Previous entry name */
  previousEntryName?: string;
  /** Date of change */
  date?: string;
  /** Version number */
  version?: string;
}

/**
 * Request to list CWEs with pagination and search.
 */
export interface ListCWEsRequest {
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
  /** Search query string */
  search?: string;
}

/**
 * Response containing paginated CWE list.
 */
export interface ListCWEsResponse {
  /** Array of CWE items */
  cwes: CWEItem[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of CWEs */
  total: number;
}

// ============================================================================
// ASVS RPC Types
// ============================================================================

/**
 * Request to list ASVS requirements with filters.
 */
export interface ListASVSRequest {
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
  /** Filter by chapter */
  chapter?: string;
  /** Filter by verification level (1-3) */
  level?: number;
}

/**
 * Response containing paginated ASVS requirements.
 */
export interface ListASVSResponse {
  /** Array of ASVS requirement items */
  requirements: ASVSItem[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of requirements */
  total: number;
}

/**
 * Request to get an ASVS requirement by ID.
 */
export interface GetASVSByIDRequest {
  /** Requirement identifier */
  requirementId: string;
}

/**
 * Response containing ASVS requirement data.
 */
export interface GetASVSByIDResponse extends ASVSItem {}

/**
 * Request to import ASVS data from a URL.
 */
export interface ImportASVSRequest {
  /** URL to fetch ASVS data from */
  url: string;
}

/**
 * Response after importing ASVS data.
 */
export interface ImportASVSResponse {
  /** Whether import was successful */
  success: boolean;
}

// CWE View Types (from pkg/cwe/views.go)
/**
 * Member of a CWE view.
 */
export interface CWEViewMember {
  /** CWE identifier */
  cweId: string;
  /** Role of the member in the view */
  role?: string;
}

/**
 * Stakeholder for a CWE view.
 */
export interface CWEViewStakeholder {
  /** Stakeholder type */
  type: string;
  /** Stakeholder description */
  description?: string;
}

/**
 * CWE view (research perspective).
 * Represents a curated collection of weaknesses.
 */
export interface CWEView {
  /** View identifier */
  id: string;
  /** View name */
  name?: string;
  /** View type */
  type?: string;
  /** View status */
  status?: string;
  /** View objective */
  objective?: string;
  /** Target audience */
  audience?: CWEViewStakeholder[];
  /** Weakness members of the view */
  members?: CWEViewMember[];
  /** External references */
  references?: Reference[];
  /** Additional notes */
  notes?: Note[];
  /** Content history */
  contentHistory?: ContentHistory[];
  /** Raw view data */
  raw?: unknown;
}

/**
 * Request to list CWE views.
 */
export interface ListCWEViewsRequest {
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing paginated CWE views.
 */
export interface ListCWEViewsResponse {
  /** Array of CWE views */
  views: CWEView[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of views */
  total: number;
}

/**
 * Response containing a single CWE view.
 */
export interface GetCWEViewResponse {
  /** The CWE view data */
  view: CWEView;
}

// ============================================================================
// Notes Framework Types
// ============================================================================

// Bookmark Types
/**
 * Bookmark entry for saving and organizing security data.
 * Part of the ULP (Unified Learning Portal) framework.
 */
export interface Bookmark {
  /** Unique bookmark identifier */
  id: number;
  /** Global item identifier */
  global_item_id: string;
  /** Type of item (CVE, CWE, CAPEC, etc.) */
  item_type: string;
  /** Item identifier */
  item_id: string;
  /** URN reference (e.g., v2e::nvd::cve::CVE-2021-1234) */
  urn: string;
  /** Bookmark title */
  title: string;
  /** Bookmark description */
  description: string;
  /** Bookmark author */
  author?: string;
  /** Whether bookmark is private */
  is_private: boolean;
  /** ISO timestamp when created */
  created_at: string;
  /** ISO timestamp when last updated */
  updated_at: string;
  /** Additional metadata */
  metadata: Record<string, unknown>;
}

/**
 * Request to create a new bookmark.
 */
export interface CreateBookmarkRequest {
  /** Global item identifier */
  global_item_id: string;
  /** Type of item */
  item_type: string;
  /** Item identifier */
  item_id: string;
  /** Bookmark title */
  title: string;
  /** Bookmark description */
  description: string;
  /** Bookmark author */
  author?: string;
  /** Whether bookmark is private */
  is_private?: boolean;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after creating a bookmark.
 */
export interface CreateBookmarkResponse {
  /** Whether creation was successful */
  success: boolean;
  /** Created bookmark */
  bookmark: Bookmark;
  /** Optional associated memory card */
  memoryCard?: MemoryCard;
}

/**
 * Request to get a bookmark by ID.
 */
export interface GetBookmarkRequest {
  /** Bookmark identifier */
  id: number;
}

/**
 * Response containing bookmark data.
 */
export interface GetBookmarkResponse {
  /** The bookmark data */
  bookmark: Bookmark;
}

/**
 * Request to list bookmarks with filters.
 */
export interface ListBookmarksRequest {
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
  /** Filter by item type */
  item_type?: string;
  /** Filter by item ID */
  item_id?: string;
  /** Filter by author */
  author?: string;
  /** Filter by privacy status */
  is_private?: boolean;
}

/**
 * Response containing paginated bookmarks.
 */
export interface ListBookmarksResponse {
  /** Array of bookmarks */
  bookmarks: Bookmark[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of bookmarks */
  total: number;
}

/**
 * Request to update a bookmark.
 */
export interface UpdateBookmarkRequest {
  /** Bookmark identifier */
  id: number;
  /** New title */
  title?: string;
  /** New description */
  description?: string;
  /** New author */
  author?: string;
  /** New privacy status */
  is_private?: boolean;
  /** New metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after updating a bookmark.
 */
export interface UpdateBookmarkResponse {
  /** Whether update was successful */
  success: boolean;
  /** Updated bookmark */
  bookmark: Bookmark;
}

/**
 * Request to delete a bookmark.
 */
export interface DeleteBookmarkRequest {
  /** Bookmark identifier */
  id: number;
}

/**
 * Response after deleting a bookmark.
 */
export interface DeleteBookmarkResponse {
  /** Whether deletion was successful */
  success: boolean;
}

// Note Types
/**
 * Note attached to a bookmark.
 * Contains user annotations and commentary.
 */
export interface NoteModel {
  /** Unique note identifier */
  id: number;
  /** Associated bookmark identifier */
  bookmark_id: number;
  /** Note content */
  content: string;
  /** Note author */
  author?: string;
  /** Whether note is private */
  is_private: boolean;
  /** ISO timestamp when created */
  created_at: string;
  /** ISO timestamp when last updated */
  updated_at: string;
  /** Additional metadata */
  metadata: Record<string, unknown>;
}

/**
 * Request to add a note to a bookmark.
 */
export interface AddNoteRequest {
  /** Bookmark identifier */
  bookmark_id: number;
  /** Note content */
  content: string;
  /** Note author */
  author?: string;
  /** Whether note is private */
  is_private?: boolean;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after adding a note.
 */
export interface AddNoteResponse {
  /** Whether addition was successful */
  success: boolean;
  /** Created note */
  note: NoteModel;
}

/**
 * Request to get a note by ID.
 */
export interface GetNoteRequest {
  /** Note identifier */
  id: number;
}

/**
 * Response containing note data.
 */
export interface GetNoteResponse {
  /** The note data */
  note: NoteModel;
}

/**
 * Request to get notes for a bookmark.
 */
export interface GetNotesByBookmarkRequest {
  /** Bookmark identifier */
  bookmark_id: number;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing bookmark notes.
 */
export interface GetNotesByBookmarkResponse {
  /** Array of notes */
  notes: NoteModel[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of notes */
  total: number;
}

/**
 * Request to update a note.
 */
export interface UpdateNoteRequest {
  /** Note identifier */
  id: number;
  /** New content */
  content?: string;
  /** New author */
  author?: string;
  /** New privacy status */
  is_private?: boolean;
  /** New metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after updating a note.
 */
export interface UpdateNoteResponse {
  /** Whether update was successful */
  success: boolean;
  /** Updated note */
  note: NoteModel;
}

/**
 * Request to delete a note.
 */
export interface DeleteNoteRequest {
  /** Note identifier */
  id: number;
}

/**
 * Response after deleting a note.
 */
export interface DeleteNoteResponse {
  /** Whether deletion was successful */
  success: boolean;
}

// Memory Card Types
/**
 * Memory card for spaced repetition learning.
 * Part of the ULP (Unified Learning Portal) framework.
 */
export interface MemoryCard {
  /** Unique card identifier */
  id: number;
  /** Associated bookmark identifier */
  bookmark_id: number;
  /** URN reference */
  urn: string;
  /** Front content (question) */
  front: string;
  /** Back content (answer) */
  back: string;
  /** Front content (alias for frontend compatibility) */
  front_content: string;
  /** Back content (alias for frontend compatibility) */
  back_content: string;
  /** Major classification */
  major_class: string;
  /** Minor classification */
  minor_class: string;
  /** Card status */
  status: string;
  /** Rich content (TipTap JSON) */
  content: any;
  /** Card type: basic, cloze, reverse */
  card_type: string;
  /** Learning state (derived from bookmark) */
  learning_state: string;
  /** Card creator/author */
  author: string;
  /** Whether card is private */
  is_private: boolean;
  /** Days until next review */
  interval: number;
  /** SM-2 algorithm ease factor */
  ease_factor: number;
  /** Number of times reviewed */
  repetitions: number;
  /** ISO timestamp when created */
  created_at: string;
  /** ISO timestamp when last updated */
  updated_at: string;
  /** ISO timestamp for next review */
  next_review_at: string;
  /** Additional metadata */
  metadata: Record<string, unknown>;
}

/**
 * Request to create a memory card.
 */
export interface CreateMemoryCardRequest {
  /** Associated bookmark identifier */
  bookmark_id: number;
  /** Front content (question) */
  front: string;
  /** Back content (answer) */
  back: string;
  /** Major classification */
  major_class?: string;
  /** Minor classification */
  minor_class?: string;
  /** Card status */
  status?: string;
  /** Rich content (TipTap JSON) */
  content?: any;
  /** Card type: basic, cloze, reverse */
  card_type?: string;
  /** Card creator/author */
  author?: string;
  /** Whether card is private */
  is_private?: boolean;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after creating a memory card.
 */
export interface CreateMemoryCardResponse {
  /** Whether creation was successful */
  success: boolean;
  /** Created memory card */
  memory_card: MemoryCard;
}

/**
 * Request to get a memory card by ID.
 */
export interface GetMemoryCardRequest {
  /** Card identifier */
  id: number;
}

/**
 * Response containing memory card data.
 */
export interface GetMemoryCardResponse {
  /** The memory card data */
  memory_card: MemoryCard;
}

/**
 * Request to list memory cards with filters.
 */
export interface ListMemoryCardsRequest {
  /** Filter by bookmark ID */
  bookmark_id?: number;
  /** Filter by learning state */
  learning_state?: string;
  /** Filter by author */
  author?: string;
  /** Filter by privacy status */
  is_private?: boolean;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing paginated memory cards.
 */
export interface ListMemoryCardsResponse {
  /** Array of memory cards */
  memory_cards: MemoryCard[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of cards */
  total: number;
}

/**
 * Request to update a memory card.
 */
export interface UpdateMemoryCardRequest {
  /** Card identifier (backend expects 'card_id') */
  card_id: number;
  /** New front content */
  front?: string;
  /** New back content */
  back?: string;
  /** New major classification */
  major_class?: string;
  /** New minor classification */
  minor_class?: string;
  /** New status */
  status?: string;
  /** New rich content */
  content?: any;
  /** New learning state */
  learning_state?: string;
  /** New author */
  author?: string;
  /** New privacy status */
  is_private?: boolean;
  /** New review interval */
  interval?: number;
  /** New ease factor */
  ease_factor?: number;
  /** New repetition count */
  repetitions?: number;
  /** New next review time */
  next_review_at?: string;
  /** New metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after updating a memory card.
 */
export interface UpdateMemoryCardResponse {
  /** Whether update was successful */
  success: boolean;
  /** Updated memory card */
  memory_card: MemoryCard;
}

/**
 * Request to delete a memory card.
 */
export interface DeleteMemoryCardRequest {
  /** Card identifier (backend expects 'card_id') */
  card_id: number;
}

/**
 * Response after deleting a memory card.
 */
export interface DeleteMemoryCardResponse {
  /** Whether deletion was successful */
  success: boolean;
}

/**
 * Request to rate a memory card (spaced repetition).
 */
export interface RateMemoryCardRequest {
  /** Card identifier (backend expects 'card_id') */
  card_id: number;
  /** Rating: 'again', 'hard', 'good', 'easy' */
  rating: string;
}

/**
 * Response after rating a memory card.
 */
export interface RateMemoryCardResponse {
  /** Whether rating was successful */
  success: boolean;
  /** Updated memory card with new schedule */
  memory_card: MemoryCard;
}

// Cross Reference Types
/**
 * Cross-reference between two items.
 * Links related security entities together.
 */
export interface CrossReference {
  /** Unique cross-reference identifier */
  id: number;
  /** Source item identifier */
  from_item_id: string;
  /** Source item type */
  from_item_type: string;
  /** Target item identifier */
  to_item_id: string;
  /** Target item type */
  to_item_type: string;
  /** Relationship type: related_to, depends_on, similar_to, opposite_of, etc. */
  relationship_type: string;
  /** Description of the relationship */
  description?: string;
  /** Relationship strength (1-5 scale) */
  strength: number;
  /** Reference author */
  author?: string;
  /** Whether reference is private */
  is_private: boolean;
  /** ISO timestamp when created */
  created_at: string;
  /** ISO timestamp when last updated */
  updated_at: string;
  /** Additional metadata */
  metadata: Record<string, unknown>;
}

/**
 * Request to create a cross-reference.
 */
export interface CreateCrossReferenceRequest {
  /** Source item identifier */
  from_item_id: string;
  /** Source item type */
  from_item_type: string;
  /** Target item identifier */
  to_item_id: string;
  /** Target item type */
  to_item_type: string;
  /** Relationship type */
  relationship_type: string;
  /** Relationship description */
  description?: string;
  /** Relationship strength (1-5) */
  strength?: number;
  /** Reference author */
  author?: string;
  /** Whether reference is private */
  is_private?: boolean;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after creating a cross-reference.
 */
export interface CreateCrossReferenceResponse {
  /** Whether creation was successful */
  success: boolean;
  /** Created cross-reference */
  cross_reference: CrossReference;
}

/**
 * Request to get a cross-reference by ID.
 */
export interface GetCrossReferenceRequest {
  /** Cross-reference identifier */
  id: number;
}

/**
 * Response containing cross-reference data.
 */
export interface GetCrossReferenceResponse {
  /** The cross-reference data */
  cross_reference: CrossReference;
}

/**
 * Request to list cross-references with filters.
 */
export interface ListCrossReferencesRequest {
  /** Filter by source item ID */
  from_item_id?: string;
  /** Filter by source item type */
  from_item_type?: string;
  /** Filter by target item ID */
  to_item_id?: string;
  /** Filter by target item type */
  to_item_type?: string;
  /** Filter by relationship type */
  relationship_type?: string;
  /** Filter by author */
  author?: string;
  /** Filter by privacy status */
  is_private?: boolean;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing paginated cross-references.
 */
export interface ListCrossReferencesResponse {
  /** Array of cross-references */
  cross_references: CrossReference[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of cross-references */
  total: number;
}

/**
 * Request to update a cross-reference.
 */
export interface UpdateCrossReferenceRequest {
  /** Cross-reference identifier */
  id: number;
  /** New relationship type */
  relationship_type?: string;
  /** New description */
  description?: string;
  /** New strength */
  strength?: number;
  /** New author */
  author?: string;
  /** New privacy status */
  is_private?: boolean;
  /** New metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after updating a cross-reference.
 */
export interface UpdateCrossReferenceResponse {
  /** Whether update was successful */
  success: boolean;
  /** Updated cross-reference */
  cross_reference: CrossReference;
}

/**
 * Request to delete a cross-reference.
 */
export interface DeleteCrossReferenceRequest {
  /** Cross-reference identifier */
  id: number;
}

/**
 * Response after deleting a cross-reference.
 */
export interface DeleteCrossReferenceResponse {
  /** Whether deletion was successful */
  success: boolean;
}

// History Types
/**
 * History entry tracking changes to items.
 * Provides audit trail for all item modifications.
 */
export interface HistoryEntry {
  /** Unique history entry identifier */
  id: number;
  /** Item identifier that changed */
  item_id: string;
  /** Item type */
  item_type: string;
  /** Action performed: created, updated, deleted, bookmarked, rated, etc. */
  action: string;
  /** Old values before change */
  old_values?: Record<string, unknown>;
  /** New values after change */
  new_values?: Record<string, unknown>;
  /** User who made the change */
  author?: string;
  /** ISO timestamp of change */
  timestamp: string;
  /** Additional metadata */
  metadata: Record<string, unknown>;
}

/**
 * Request to add a history entry.
 */
export interface AddHistoryRequest {
  /** Item identifier */
  item_id: string;
  /** Item type */
  item_type: string;
  /** Action performed */
  action: string;
  /** Old values before change */
  old_values?: Record<string, unknown>;
  /** New values after change */
  new_values?: Record<string, unknown>;
  /** User who made the change */
  author?: string;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response after adding a history entry.
 */
export interface AddHistoryResponse {
  /** Whether addition was successful */
  success: boolean;
  /** Created history entry */
  history_entry: HistoryEntry;
}

/**
 * Request to get history for an item.
 */
export interface GetHistoryRequest {
  /** Item identifier */
  item_id: string;
  /** Item type */
  item_type: string;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing item history.
 */
export interface GetHistoryResponse {
  /** Array of history entries */
  history_entries: HistoryEntry[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of entries */
  total: number;
}

/**
 * Request to get history by action type.
 */
export interface GetHistoryByActionRequest {
  /** Action type to filter by */
  action: string;
  /** Filter by author */
  author?: string;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing filtered history entries.
 */
export interface GetHistoryByActionResponse {
  /** Array of history entries */
  history_entries: HistoryEntry[];
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
  /** Total number of entries */
  total: number;
}

// Bookmark State Reversion
/**
 * Request to revert a bookmark to a previous state.
 */
export interface RevertBookmarkStateRequest {
  /** Item identifier */
  item_id: string;
  /** Item type */
  item_type: string;
  /** ISO timestamp to revert to */
  to_timestamp: string;
  /** User requesting reversion */
  author?: string;
}

/**
 * Response after reverting bookmark state.
 */
export interface RevertBookmarkStateResponse {
  /** Whether reversion was successful */
  success: boolean;
  /** Response message */
  message: string;
}

// ============================================================================
// SSG (SCAP Security Guide) Data Types
// ============================================================================

/**
 * SSG security guide benchmark.
 * Represents a complete security benchmark document.
 */
export interface SSGGuide {
  /** Unique guide identifier */
  id: string;
  /** Product name (e.g., "rhel8") */
  product: string;
  /** Profile identifier */
  profileId: string;
  /** Short identifier */
  shortId: string;
  /** Guide title */
  title: string;
  /** HTML content of the guide */
  htmlContent: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG group (category of rules).
 */
export interface SSGGroup {
  /** Unique group identifier */
  id: string;
  /** Associated guide identifier */
  guideId: string;
  /** Parent group identifier */
  parentId: string;
  /** Group title */
  title: string;
  /** Group description */
  description: string;
  /** Nesting level */
  level: number;
  /** Number of child groups */
  groupCount: number;
  /** Number of rules in group */
  ruleCount: number;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG reference link.
 */
export interface SSGReference {
  /** Reference URL */
  href: string;
  /** Reference label */
  label: string;
  /** Reference value */
  value: string;
}

/**
 * SSG security rule.
 */
export interface SSGRule {
  /** Unique rule identifier */
  id: string;
  /** Associated guide identifier */
  guideId: string;
  /** Associated group identifier */
  groupId: string;
  /** Short identifier */
  shortId: string;
  /** Rule title */
  title: string;
  /** Rule description */
  description: string;
  /** Rule rationale */
  rationale: string;
  /** Severity level: low, medium, high */
  severity: 'low' | 'medium' | 'high';
  /** External references */
  references: SSGReference[];
  /** Requirement level */
  level: number;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG mapping table.
 */
export interface SSGTable {
  /** Unique table identifier */
  id: string;
  /** Product name */
  product: string;
  /** Table type */
  tableType: string;
  /** Table title */
  title: string;
  /** Table description */
  description: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG table entry (mapping).
 */
export interface SSGTableEntry {
  /** Unique entry identifier */
  id: number;
  /** Associated table identifier */
  tableId: string;
  /** Mapping string */
  mapping: string;
  /** Associated rule title */
  ruleTitle: string;
  /** Entry description */
  description: string;
  /** Entry rationale */
  rationale: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG manifest (product metadata).
 */
export interface SSGManifest {
  /** Unique manifest identifier */
  id: string;
  /** Product name */
  product: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG profile within a manifest.
 */
export interface SSGProfile {
  /** Unique profile identifier */
  id: string;
  /** Associated manifest identifier */
  manifestId: string;
  /** Profile identifier */
  profileId: string;
  /** Profile description */
  description: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG profile rule association.
 */
export interface SSGProfileRule {
  /** Unique association identifier */
  id: number;
  /** Profile identifier */
  profileId: string;
  /** Rule short identifier */
  ruleShortId: string;
  /** ISO timestamp when created */
  createdAt: string;
}

/**
 * Complete SSG guide tree structure.
 */
export interface SSGTree {
  /** Guide metadata */
  guide: SSGGuide;
  /** Groups in the guide */
  groups: SSGGroup[];
  /** Rules in the guide */
  rules: SSGRule[];
}

/**
 * Tree node for hierarchical display.
 */
export interface TreeNode {
  /** Node identifier */
  id: string;
  /** Parent node identifier */
  parentId: string;
  /** Nesting level */
  level: number;
  /** Node type: group or rule */
  type: 'group' | 'rule';
  /** Group data (if type is 'group') */
  group?: SSGGroup;
  /** Rule data (if type is 'rule') */
  rule?: SSGRule;
  /** Child nodes */
  children: TreeNode[];
}

// SSG RPC Request/Response Types

/**
 * Request to import an SSG guide from file.
 */
export interface SSGImportGuideRequest {
  /** File path to the guide */
  path: string;
}

/**
 * Response after importing an SSG guide.
 */
export interface SSGImportGuideResponse {
  /** Whether import was successful */
  success: boolean;
  /** Imported guide identifier */
  guideId: string;
  /** Number of groups imported */
  groupCount: number;
  /** Number of rules imported */
  ruleCount: number;
}

/**
 * Request to get an SSG guide by ID.
 */
export interface SSGGetGuideRequest {
  /** Guide identifier */
  id: string;
}

/**
 * Response containing SSG guide data.
 */
export interface SSGGetGuideResponse {
  /** The guide data */
  guide: SSGGuide;
}

/**
 * Request to list SSG guides with filters.
 */
export interface SSGListGuidesRequest {
  /** Filter by product name */
  product?: string;
  /** Filter by profile ID */
  profileId?: string;
}

/**
 * Response containing list of SSG guides.
 */
export interface SSGListGuidesResponse {
  /** Array of guides */
  guides: SSGGuide[];
  /** Number of guides */
  count: number;
}

/**
 * Request to get the tree structure of a guide.
 */
export interface SSGGetTreeRequest {
  /** Guide identifier */
  guideId: string;
}

/**
 * Response containing the guide tree.
 */
export interface SSGGetTreeResponse {
  /** Complete tree structure */
  tree: SSGTree;
}

/**
 * Request to get tree nodes for a guide.
 */
export interface SSGGetTreeNodeRequest {
  /** Guide identifier */
  guideId: string;
}

/**
 * Response containing tree nodes.
 */
export interface SSGGetTreeNodeResponse {
  /** Array of tree nodes */
  nodes: TreeNode[];
  /** Number of nodes */
  count: number;
}

/**
 * Request to get an SSG group by ID.
 */
export interface SSGGetGroupRequest {
  /** Group identifier */
  id: string;
}

/**
 * Response containing SSG group data.
 */
export interface SSGGetGroupResponse {
  /** The group data */
  group: SSGGroup;
}

/**
 * Request to get child groups.
 */
export interface SSGGetChildGroupsRequest {
  /** Parent group identifier */
  parentId?: string;
}

/**
 * Response containing child groups.
 */
export interface SSGGetChildGroupsResponse {
  /** Array of child groups */
  groups: SSGGroup[];
  /** Number of groups */
  count: number;
}

/**
 * Request to get an SSG rule by ID.
 */
export interface SSGGetRuleRequest {
  /** Rule identifier */
  id: string;
}

/**
 * Response containing SSG rule data.
 */
export interface SSGGetRuleResponse {
  /** The rule data */
  rule: SSGRule;
}

/**
 * Request to list SSG rules with filters.
 */
export interface SSGListRulesRequest {
  /** Filter by group ID */
  groupId?: string;
  /** Filter by severity */
  severity?: string;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing paginated SSG rules.
 */
export interface SSGListRulesResponse {
  /** Array of rules */
  rules: SSGRule[];
  /** Total number of rules */
  total: number;
}

/**
 * Request to get child rules for a group.
 */
export interface SSGGetChildRulesRequest {
  /** Parent group identifier */
  groupId: string;
}

/**
 * Response containing child rules.
 */
export interface SSGGetChildRulesResponse {
  /** Array of child rules */
  rules: SSGRule[];
  /** Number of rules */
  count: number;
}

// SSG Table RPC Types

/**
 * Request to list SSG tables with filters.
 */
export interface SSGListTablesRequest {
  /** Filter by product name */
  product?: string;
  /** Filter by table type */
  tableType?: string;
}

/**
 * Response containing list of SSG tables.
 */
export interface SSGListTablesResponse {
  /** Array of tables */
  tables: SSGTable[];
  /** Number of tables */
  count: number;
}

/**
 * Request to get an SSG table by ID.
 */
export interface SSGGetTableRequest {
  /** Table identifier */
  id: string;
}

/**
 * Response containing SSG table data.
 */
export interface SSGGetTableResponse {
  /** The table data */
  table: SSGTable;
}

/**
 * Request to get entries for an SSG table.
 */
export interface SSGGetTableEntriesRequest {
  /** Table identifier */
  tableId: string;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing table entries.
 */
export interface SSGGetTableEntriesResponse {
  /** Array of table entries */
  entries: SSGTableEntry[];
  /** Total number of entries */
  total: number;
}

/**
 * Request to import an SSG table from file.
 */
export interface SSGImportTableRequest {
  /** File path to table */
  path: string;
}

/**
 * Response after importing an SSG table.
 */
export interface SSGImportTableResponse {
  /** Whether import was successful */
  success: boolean;
  /** Imported table identifier */
  tableId: string;
  /** Number of entries imported */
  entryCount: number;
}

// SSG Manifest RPC Types

/**
 * Request to list SSG manifests with filters.
 */
export interface SSGListManifestsRequest {
  /** Filter by product name */
  product?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing list of SSG manifests.
 */
export interface SSGListManifestsResponse {
  /** Array of manifests */
  manifests: SSGManifest[];
  /** Number of manifests */
  count: number;
}

/**
 * Request to get an SSG manifest by ID.
 */
export interface SSGGetManifestRequest {
  /** Manifest identifier */
  manifestId: string;
}

/**
 * Response containing SSG manifest data.
 */
export interface SSGGetManifestResponse {
  /** The manifest data */
  manifest: SSGManifest;
}

/**
 * Request to list SSG profiles with filters.
 */
export interface SSGListProfilesRequest {
  /** Filter by product name */
  product?: string;
  /** Filter by profile ID */
  profileId?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing list of SSG profiles.
 */
export interface SSGListProfilesResponse {
  /** Array of profiles */
  profiles: SSGProfile[];
  /** Number of profiles */
  count: number;
}

/**
 * Request to get an SSG profile by ID.
 */
export interface SSGGetProfileRequest {
  /** Profile identifier */
  profileId: string;
}

/**
 * Response containing SSG profile data.
 */
export interface SSGGetProfileResponse {
  /** The profile data */
  profile: SSGProfile;
}

/**
 * Request to get rules for an SSG profile.
 */
export interface SSGGetProfileRulesRequest {
  /** Profile identifier */
  profileId: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing profile rules.
 */
export interface SSGGetProfileRulesResponse {
  /** Array of profile rules */
  rules: SSGProfileRule[];
  /** Number of rules */
  count: number;
}

// SSG Data Stream Types

/**
 * SSG data stream component.
 * Represents a SCAP data stream with all its components.
 */
export interface SSGDataStream {
  /** Unique data stream identifier */
  id: string;
  /** Product name */
  product: string;
  /** SCAP version */
  scapVersion: string;
  /** Generation timestamp */
  generated: string;
  /** XCCDF benchmark identifier */
  xccdfBenchmarkId: string;
  /** OVAL checks identifier */
  ovalChecksId: string;
  /** OCIL questionnaires identifier */
  ocilQuestionnairesId: string;
  /** CPE dictionary identifier */
  cpeDictId: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG benchmark within a data stream.
 */
export interface SSGBenchmark {
  /** Unique benchmark identifier */
  id: string;
  /** Associated data stream identifier */
  dataStreamId: string;
  /** XCCDF identifier */
  xccdfId: string;
  /** Benchmark title */
  title: string;
  /** Benchmark version */
  version: string;
  /** Benchmark description */
  description: string;
  /** Total number of profiles */
  totalProfiles: number;
  /** Total number of groups */
  totalGroups: number;
  /** Total number of rules */
  totalRules: number;
  /** Maximum group nesting level */
  maxGroupLevel: number;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG profile within a data stream.
 */
export interface SSGDSProfile {
  /** Unique profile identifier */
  id: string;
  /** Associated data stream identifier */
  dataStreamId: string;
  /** XCCDF profile identifier */
  xccdfProfileId: string;
  /** Profile title */
  title: string;
  /** Profile description */
  description: string;
  /** Total number of rules */
  totalRules: number;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG profile rule association within data stream.
 */
export interface SSGDSProfileRule {
  /** Unique association identifier */
  id: number;
  /** Profile identifier */
  profileId: string;
  /** Rule short identifier */
  ruleShortId: string;
  /** Whether rule is selected in profile */
  selected: boolean;
  /** ISO timestamp when created */
  createdAt: string;
}

/**
 * SSG group within a data stream.
 */
export interface SSGDSGroup {
  /** Unique group identifier */
  id: string;
  /** Associated data stream identifier */
  dataStreamId: string;
  /** XCCDF group identifier */
  xccdfGroupId: string;
  /** Parent XCCDF group identifier */
  parentXccdfGroupId: string;
  /** Group title */
  title: string;
  /** Group description */
  description: string;
  /** Nesting level */
  level: number;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * SSG rule within a data stream.
 */
export interface SSGDSRule {
  /** Unique rule identifier */
  id: string;
  /** Associated data stream identifier */
  dataStreamId: string;
  /** XCCDF rule identifier */
  xccdfRuleId: string;
  /** Group XCCDF identifier */
  groupXccdfId: string;
  /** Short identifier */
  shortId: string;
  /** Rule title */
  title: string;
  /** Rule description */
  description: string;
  /** Rule rationale */
  rationale: string;
  /** Severity level */
  severity: string;
  /** Warning message */
  warning: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
  /** External references */
  references?: SSGDSRuleReference[];
  /** Rule identifiers */
  identifiers?: SSGDSRuleIdentifier[];
}

/**
 * Reference for an SSG data stream rule.
 */
export interface SSGDSRuleReference {
  /** Unique reference identifier */
  id: number;
  /** Rule identifier */
  ruleId: string;
  /** Reference URL */
  href: string;
  /** Reference text */
  text: string;
  /** ISO timestamp when created */
  createdAt: string;
}

/**
 * Identifier for an SSG data stream rule.
 */
export interface SSGDSRuleIdentifier {
  /** Unique identifier record */
  id: number;
  /** Rule identifier */
  ruleId: string;
  /** Identifier system (e.g., "CCE") */
  system: string;
  /** Identifier value */
  identifier: string;
  /** ISO timestamp when created */
  createdAt: string;
}

// SSG Data Stream RPC Types

/**
 * Request to list SSG data streams with filters.
 */
export interface SSGListDataStreamsRequest {
  /** Filter by product name */
  product?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing list of SSG data streams.
 */
export interface SSGListDataStreamsResponse {
  /** Array of data streams */
  dataStreams: SSGDataStream[];
  /** Number of data streams */
  count: number;
}

/**
 * Request to get an SSG data stream by ID.
 */
export interface SSGGetDataStreamRequest {
  /** Data stream identifier */
  dataStreamId: string;
}

/**
 * Response containing SSG data stream data.
 */
export interface SSGGetDataStreamResponse {
  /** The data stream */
  dataStream: SSGDataStream;
  /** Optional benchmark data */
  benchmark?: SSGBenchmark;
}

/**
 * Request to list data stream profiles.
 */
export interface SSGListDSProfilesRequest {
  /** Data stream identifier */
  dataStreamId: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing data stream profiles.
 */
export interface SSGListDSProfilesResponse {
  /** Array of profiles */
  profiles: SSGDSProfile[];
  /** Number of profiles */
  count: number;
}

/**
 * Request to get a data stream profile by ID.
 */
export interface SSGGetDSProfileRequest {
  /** Profile identifier */
  profileId: string;
}

/**
 * Response containing data stream profile data.
 */
export interface SSGGetDSProfileResponse {
  /** The profile data */
  profile: SSGDSProfile;
}

/**
 * Request to get rules for a data stream profile.
 */
export interface SSGGetDSProfileRulesRequest {
  /** Profile identifier */
  profileId: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing data stream profile rules.
 */
export interface SSGGetDSProfileRulesResponse {
  /** Array of profile rules */
  rules: SSGDSProfileRule[];
  /** Number of rules */
  count: number;
}

/**
 * Request to list data stream groups.
 */
export interface SSGListDSGroupsRequest {
  /** Data stream identifier */
  dataStreamId: string;
  /** Parent XCCDF group identifier */
  parentXccdfGroupId?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing data stream groups.
 */
export interface SSGListDSGroupsResponse {
  /** Array of groups */
  groups: SSGDSGroup[];
  /** Number of groups */
  count: number;
}

/**
 * Request to list data stream rules with filters.
 */
export interface SSGListDSRulesRequest {
  /** Data stream identifier */
  dataStreamId: string;
  /** Filter by group XCCDF ID */
  groupXccdfId?: string;
  /** Filter by severity */
  severity?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing data stream rules.
 */
export interface SSGListDSRulesResponse {
  /** Array of rules */
  rules: SSGDSRule[];
  /** Total number of rules */
  total: number;
}

/**
 * Request to get a data stream rule by ID.
 */
export interface SSGGetDSRuleRequest {
  /** Rule identifier */
  ruleId: string;
}

/**
 * Response containing data stream rule data.
 */
export interface SSGGetDSRuleResponse {
  /** The rule data */
  rule: SSGDSRule;
  /** Rule references */
  references: SSGDSRuleReference[];
  /** Rule identifiers */
  identifiers: SSGDSRuleIdentifier[];
}

/**
 * Request to import an SSG data stream from file.
 */
export interface SSGImportDataStreamRequest {
  /** File path to data stream */
  path: string;
}

/**
 * Response after importing an SSG data stream.
 */
export interface SSGImportDataStreamResponse {
  /** Whether import was successful */
  success: boolean;
  /** Imported data stream identifier */
  dataStreamId: string;
  /** Number of profiles imported */
  profileCount: number;
  /** Number of groups imported */
  groupCount: number;
  /** Number of rules imported */
  ruleCount: number;
}

// SSG Import Job RPC Types

/**
 * Request to start an SSG import job.
 */
export interface SSGStartImportJobRequest {
  /** Optional run identifier */
  runId?: string;
}

/**
 * Response after starting an SSG import job.
 */
export interface SSGStartImportJobResponse {
  /** Whether start was successful */
  success: boolean;
  /** Run identifier for the job */
  runId: string;
}

/**
 * Response after stopping an SSG import job.
 */
export interface SSGStopImportJobResponse {
  /** Whether stop was successful */
  success: boolean;
}

/**
 * Response after pausing an SSG import job.
 */
export interface SSGPauseImportJobResponse {
  /** Whether pause was successful */
  success: boolean;
}

/**
 * Request to resume a paused SSG import job.
 */
export interface SSGResumeImportJobRequest {
  /** Run identifier */
  runId: string;
}

/**
 * Response after resuming an SSG import job.
 */
export interface SSGResumeImportJobResponse {
  /** Whether resume was successful */
  success: boolean;
}

/**
 * Response containing SSG import job status.
 */
export interface SSGGetImportStatusResponse {
  /** Job identifier */
  id: string;
  /** Data type being imported */
  dataType: string;
  /** Job state */
  state: 'queued' | 'running' | 'paused' | 'completed' | 'failed' | 'stopped';
  /** ISO timestamp when job started */
  startedAt: string;
  /** ISO timestamp when job completed */
  completedAt?: string;
  /** Error message if failed */
  error?: string;
  /** Import progress details */
  progress: {
    /** Total number of guides to process */
    totalGuides: number;
    /** Number of guides processed */
    processedGuides: number;
    /** Number of guides that failed */
    failedGuides: number;
    /** Total number of tables to process */
    totalTables: number;
    /** Number of tables processed */
    processedTables: number;
    /** Number of tables that failed */
    failedTables: number;
    /** Total number of manifests to process */
    totalManifests: number;
    /** Number of manifests processed */
    processedManifests: number;
    /** Number of manifests that failed */
    failedManifests: number;
    /** Total number of data streams to process */
    totalDataStreams: number;
    /** Number of data streams processed */
    processedDataStreams: number;
    /** Number of data streams that failed */
    failedDataStreams: number;
    /** Current file being processed */
    currentFile: string;
    /** Current processing phase */
    currentPhase?: string;
  };
  /** Additional metadata */
  metadata?: Record<string, string>;
}

// SSG Remote Service RPC Types (Git operations)

/**
 * Response after cloning SSG repository.
 */
export interface SSGCloneRepoResponse {
  /** Whether clone was successful */
  success: boolean;
  /** Path to cloned repository */
  path: string;
}

/**
 * Response after pulling SSG repository updates.
 */
export interface SSGPullRepoResponse {
  /** Whether pull was successful */
  success: boolean;
}

/**
 * Response containing SSG repository status.
 */
export interface SSGGetRepoStatusResponse {
  /** Current commit hash */
  commitHash: string;
  /** Current branch name */
  branch: string;
  /** Whether working directory is clean */
  isClean: boolean;
}

/**
 * Response containing list of SSG guide files.
 */
export interface SSGListGuideFilesResponse {
  /** Array of file paths */
  files: string[];
  /** Number of files */
  count: number;
}

/**
 * Request to get file path from filename.
 */
export interface SSGGetFilePathRequest {
  /** Filename to look up */
  filename: string;
}

/**
 * Response containing file path.
 */
export interface SSGGetFilePathResponse {
  /** Full path to the file */
  path: string;
}

// SSG Cross-Reference Types

/**
 * Cross-reference between SSG objects.
 */
export interface SSGCrossReference {
  /** Unique cross-reference identifier */
  id: number;
  /** Source object type: guide, table, manifest, datastream */
  sourceType: string;
  /** Source object identifier */
  sourceId: string;
  /** Target object type: guide, table, manifest, datastream */
  targetType: string;
  /** Target object identifier */
  targetId: string;
  /** Link type: rule_id, cce, product, profile_id */
  linkType: string;
  /** JSON string with additional context */
  metadata: string;
  /** ISO timestamp when created */
  createdAt: string;
}

/**
 * Request to get SSG cross-references with filters.
 */
export interface SSGGetCrossReferencesRequest {
  /** Filter by source type */
  sourceType?: string;
  /** Filter by source ID */
  sourceId?: string;
  /** Filter by target type */
  targetType?: string;
  /** Filter by target ID */
  targetId?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing SSG cross-references.
 */
export interface SSGGetCrossReferencesResponse {
  /** Array of cross-references */
  crossReferences: SSGCrossReference[];
  /** Number of cross-references */
  count: number;
}

/**
 * Request to find related SSG objects.
 */
export interface SSGFindRelatedObjectsRequest {
  /** Object type */
  objectType: string;
  /** Object identifier */
  objectId: string;
  /** Filter by link type */
  linkType?: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing related SSG objects.
 */
export interface SSGFindRelatedObjectsResponse {
  /** Array of related objects */
  relatedObjects: SSGCrossReference[];
  /** Number of related objects */
  count: number;
}

// ============================================================================
// UEE (Unified ETL Engine) Types
// ============================================================================

/**
 * Macro FSM state for ETL orchestration.
 */
export type MacroFSMState =
  | "BOOTSTRAPPING"
  | "ORCHESTRATING"
  | "STABILIZING"
  | "DRAINING";

/**
 * Provider FSM state for individual ETL providers.
 */
export type ProviderFSMState =
  | "IDLE"
  | "ACQUIRING"
  | "RUNNING"
  | "WAITING_QUOTA"
  | "WAITING_BACKOFF"
  | "PAUSED"
  | "TERMINATED";

/**
 * ETL provider node in the orchestration tree.
 */
export interface ProviderNode {
  /** Provider identifier */
  id: string;
  /** Provider type (CVE, CWE, CAPEC, etc.) */
  providerType: string;
  /** Current provider FSM state */
  state: ProviderFSMState;
  /** Number of items processed */
  processedCount: number;
  /** Number of errors encountered */
  errorCount: number;
  /** Number of permits held */
  permitsHeld: number;
  /** Last checkpoint URN */
  lastCheckpoint?: string;
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * Macro node containing all providers.
 */
export interface MacroNode {
  /** Macro node identifier */
  id: string;
  /** Current macro FSM state */
  state: MacroFSMState;
  /** Array of provider nodes */
  providers: ProviderNode[];
  /** ISO timestamp when created */
  createdAt: string;
  /** ISO timestamp when last updated */
  updatedAt: string;
}

/**
 * Complete ETL tree structure.
 */
export interface ETLTree {
  /** Macro node containing all providers */
  macro: MacroNode;
  /** Total number of providers */
  totalProviders: number;
  /** Number of active providers */
  activeProviders: number;
}

/**
 * ETL kernel performance metrics.
 */
export interface KernelMetrics {
  /** P99 latency in milliseconds */
  p99Latency: number;
  /** Buffer saturation percentage (0-100) */
  bufferSaturation: number;
  /** Messages per second */
  messageRate: number;
  /** Errors per second */
  errorRate: number;
  /** ISO timestamp */
  timestamp: string;
}

/**
 * Checkpoint for provider progress tracking.
 */
export interface Checkpoint {
  /** URN key (v2e::provider::type::id) */
  urn: string;
  /** Provider identifier */
  providerID: string;
  /** Whether checkpoint was successful */
  success: boolean;
  /** Error message if failed */
  errorMessage?: string;
  /** ISO timestamp when processed */
  processedAt: string;
}

/**
 * Permit allocation for provider resource management.
 */
export interface PermitAllocation {
  /** Provider identifier */
  providerID: string;
  /** Number of permits held */
  permitsHeld: number;
  /** Number of permits requested */
  permitsRequested: number;
  /** ISO timestamp */
  timestamp: string;
}

// RPC Request/Response types

/**
 * Response containing ETL tree structure.
 */
export interface GetEtlTreeResponse {
  /** Complete ETL tree */
  tree: ETLTree;
}

/**
 * Response containing ETL kernel metrics.
 */
export interface GetKernelMetricsResponse {
  /** Kernel performance metrics */
  metrics: KernelMetrics;
}

/**
 * Request to get provider checkpoints.
 */
export interface GetProviderCheckpointsRequest {
  /** Provider identifier */
  providerID: string;
  /** Maximum results per page */
  limit?: number;
  /** Pagination offset */
  offset?: number;
}

/**
 * Response containing provider checkpoints.
 */
export interface GetProviderCheckpointsResponse {
  /** Array of checkpoints */
  checkpoints: Checkpoint[];
  /** Number of checkpoints */
  count: number;
}

// ============================================================================
// GLC (Graphized Learning Canvas) Types
// ============================================================================

// Graph Model
/**
 * GLC graph for knowledge visualization.
 * Part of the Unified Learning Portal framework.
 */
export interface GLCGraph {
  /** Unique database identifier */
  id: number;
  /** Unique graph identifier */
  graph_id: string;
  /** Graph name */
  name: string;
  /** Graph description */
  description: string;
  /** Associated preset identifier */
  preset_id: string;
  /** Comma-separated tags */
  tags: string;
  /** JSON array of nodes */
  nodes: string;
  /** JSON array of edges */
  edges: string;
  /** JSON viewport state */
  viewport: string;
  /** Base64 data URL for graph preview */
  thumbnail?: string;
  /** Version number for optimistic locking */
  version: number;
  /** Whether graph is archived */
  is_archived: boolean;
  /** ISO timestamp when created */
  created_at: string;
  /** ISO timestamp when last updated */
  updated_at: string;
}

/**
 * GLC graph version snapshot.
 */
export interface GLCGraphVersion {
  /** Unique version identifier */
  id: number;
  /** Associated graph database ID */
  graph_id: number;
  /** Version number */
  version: number;
  /** JSON array of nodes */
  nodes: string;
  /** JSON array of edges */
  edges: string;
  /** JSON viewport state */
  viewport: string;
  /** ISO timestamp when created */
  created_at: string;
}

// Graph Request/Response Types
/**
 * Request to create a new GLC graph.
 */
export interface CreateGLCGraphRequest {
  /** Graph name */
  name: string;
  /** Graph description */
  description?: string;
  /** Preset identifier to use */
  preset_id: string;
  /** Initial nodes JSON */
  nodes?: string;
  /** Initial edges JSON */
  edges?: string;
  /** Initial viewport state */
  viewport?: string;
  /** Comma-separated tags */
  tags?: string;
}

/**
 * Response after creating a GLC graph.
 */
export interface CreateGLCGraphResponse {
  /** Whether creation was successful */
  success: boolean;
  /** Created graph */
  graph: GLCGraph;
}

/**
 * Request to get a GLC graph by ID.
 */
export interface GetGLCGraphRequest {
  /** Graph identifier */
  graph_id: string;
}

/**
 * Response containing GLC graph data.
 */
export interface GetGLCGraphResponse {
  /** The graph data */
  graph: GLCGraph;
}

/**
 * Request to update a GLC graph.
 */
export interface UpdateGLCGraphRequest {
  /** Graph identifier */
  graph_id: string;
  /** New name */
  name?: string;
  /** New description */
  description?: string;
  /** New nodes JSON */
  nodes?: string;
  /** New edges JSON */
  edges?: string;
  /** New viewport state */
  viewport?: string;
  /** New tags */
  tags?: string;
  /** New archive status */
  is_archived?: boolean;
}

/**
 * Response after updating a GLC graph.
 */
export interface UpdateGLCGraphResponse {
  /** Whether update was successful */
  success: boolean;
  /** Updated graph */
  graph: GLCGraph;
}

/**
 * Request to delete a GLC graph.
 */
export interface DeleteGLCGraphRequest {
  /** Graph identifier */
  graph_id: string;
}

/**
 * Response after deleting a GLC graph.
 */
export interface DeleteGLCGraphResponse {
  /** Whether deletion was successful */
  success: boolean;
}

/**
 * Request to list GLC graphs with filters.
 */
export interface ListGLCGraphsRequest {
  /** Filter by preset ID */
  preset_id?: string;
  /** Pagination offset */
  offset?: number;
  /** Maximum results per page */
  limit?: number;
}

/**
 * Response containing paginated GLC graphs.
 */
export interface ListGLCGraphsResponse {
  /** Array of graphs */
  graphs: GLCGraph[];
  /** Total number of graphs */
  total: number;
  /** Current offset */
  offset: number;
  /** Current page limit */
  limit: number;
}

/**
 * Request to list recent GLC graphs.
 */
export interface ListRecentGLCGraphsRequest {
  /** Maximum number of results */
  limit?: number;
}

/**
 * Response containing recent GLC graphs.
 */
export interface ListRecentGLCGraphsResponse {
  /** Array of recent graphs */
  graphs: GLCGraph[];
}

// Version Request/Response Types
/**
 * Request to get a specific GLC graph version.
 */
export interface GetGLCVersionRequest {
  /** Graph identifier */
  graph_id: string;
  /** Version number to retrieve */
  version: number;
}

/**
 * Response containing GLC graph version.
 */
export interface GetGLCVersionResponse {
  /** The version data */
  version: GLCGraphVersion;
}

/**
 * Request to list GLC graph versions.
 */
export interface ListGLCVersionsRequest {
  /** Graph identifier */
  graph_id: string;
  /** Maximum number of versions */
  limit?: number;
}

/**
 * Response containing GLC graph versions.
 */
export interface ListGLCVersionsResponse {
  /** Array of versions */
  versions: GLCGraphVersion[];
}

/**
 * Request to restore a GLC graph to a previous version.
 */
export interface RestoreGLCVersionRequest {
  /** Graph identifier */
  graph_id: string;
  /** Version number to restore */
  version: number;
}

/**
 * Response after restoring GLC graph version.
 */
export interface RestoreGLCVersionResponse {
  /** Whether restoration was successful */
  success: boolean;
  /** Restored graph */
  graph: GLCGraph;
}

// User Preset Model
/**
 * User-defined GLC preset configuration.
 */
export interface GLCUserPreset {
  /** Unique database identifier */
  id: number;
  /** Unique preset identifier */
  preset_id: string;
  /** Preset name */
  name: string;
  /** Preset version */
  version: string;
  /** Preset description */
  description: string;
  /** Preset author */
  author: string;
  /** JSON theme configuration */
  theme: string;
  /** JSON behavior configuration */
  behavior: string;
  /** JSON array of node type definitions */
  node_types: string;
  /** JSON array of relationship definitions */
  relations: string;
  /** ISO timestamp when created */
  created_at: string;
  /** ISO timestamp when last updated */
  updated_at: string;
}

// Preset Request/Response Types
/**
 * Request to create a new GLC preset.
 */
export interface CreateGLCPresetRequest {
  /** Preset name */
  name: string;
  /** Preset version */
  version?: string;
  /** Preset description */
  description?: string;
  /** Preset author */
  author?: string;
  /** Theme configuration object */
  theme: object;
  /** Behavior configuration object */
  behavior: object;
  /** Node type definitions */
  node_types: object[];
  /** Relationship definitions */
  relations: object[];
}

/**
 * Response after creating a GLC preset.
 */
export interface CreateGLCPresetResponse {
  /** Whether creation was successful */
  success: boolean;
  /** Created preset */
  preset: GLCUserPreset;
}

/**
 * Request to get a GLC preset by ID.
 */
export interface GetGLCPresetRequest {
  /** Preset identifier */
  preset_id: string;
}

/**
 * Response containing GLC preset data.
 */
export interface GetGLCPresetResponse {
  /** The preset data */
  preset: GLCUserPreset;
}

/**
 * Request to update a GLC preset.
 */
export interface UpdateGLCPresetRequest {
  /** Preset identifier */
  preset_id: string;
  /** New name */
  name?: string;
  /** New version */
  version?: string;
  /** New description */
  description?: string;
  /** New author */
  author?: string;
  /** New theme configuration */
  theme?: object;
  /** New behavior configuration */
  behavior?: object;
  /** New node type definitions */
  node_types?: object[];
  /** New relationship definitions */
  relationships?: object[];
}

/**
 * Response after updating a GLC preset.
 */
export interface UpdateGLCPresetResponse {
  /** Whether update was successful */
  success: boolean;
  /** Updated preset */
  preset: GLCUserPreset;
}

/**
 * Request to delete a GLC preset.
 */
export interface DeleteGLCPresetRequest {
  /** Preset identifier */
  preset_id: string;
}

/**
 * Response after deleting a GLC preset.
 */
export interface DeleteGLCPresetResponse {
  /** Whether deletion was successful */
  success: boolean;
}

/**
 * Response containing all GLC presets.
 */
export interface ListGLCPresetsResponse {
  /** Array of user presets */
  presets: GLCUserPreset[];
}

// Share Link Model
/**
 * GLC share link for graph sharing.
 */
export interface GLCShareLink {
  /** Unique database identifier */
  id: number;
  /** Unique link identifier */
  link_id: string;
  /** Associated graph identifier */
  graph_id: string;
  /** Optional password protection */
  password?: string;
  /** Optional expiration timestamp */
  expires_at?: string;
  /** Number of times link was viewed */
  view_count: number;
  /** ISO timestamp when created */
  created_at: string;
}

// Share Link Request/Response Types
/**
 * Request to create a GLC share link.
 */
export interface CreateGLCShareLinkRequest {
  /** Graph identifier to share */
  graph_id: string;
  /** Optional password */
  password?: string;
  /** Optional expiration in hours */
  expires_in_hours?: number;
}

/**
 * Response after creating a GLC share link.
 */
export interface CreateGLCShareLinkResponse {
  /** Whether creation was successful */
  success: boolean;
  /** Created share link */
  share_link: GLCShareLink;
}

/**
 * Request to get a shared GLC graph.
 */
export interface GetGLCSharedGraphRequest {
  /** Share link identifier */
  link_id: string;
  /** Optional password for protected links */
  password?: string;
}

/**
 * Response containing shared GLC graph.
 */
export interface GetGLCSharedGraphResponse {
  /** The shared graph */
  graph: GLCGraph;
}

/**
 * Request to get embed data for a shared graph.
 */
export interface GetGLCShareEmbedDataRequest {
  /** Share link identifier */
  link_id: string;
}

/**
 * Response containing embed data for shared graph.
 */
export interface GetGLCShareEmbedDataResponse {
  /** Share link information */
  share_link: GLCShareLink;
  /** The shared graph */
  graph: GLCGraph;
}

// ============================================================================
// CVSS Calculator Types
// ============================================================================

/**
 * Supported CVSS versions
 */
export type CVSSVersion = '3.0' | '3.1' | '4.0';

/**
 * CVSS severity rating levels
 */
export type CVSSSeverity = 'NONE' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

/**
 * Validated metric value with display name
 */
export interface MetricValue<T = string> {
  /** Value identifier */
  value: T;
  /** Display name for UI */
  label: string;
  /** Detailed description */
  description: string;
}

// ============================================================================
// CVSS v3.0 / v3.1 Types
// ============================================================================

/**
 * CVSS v3 Attack Vector metric
 */
export type AV = 'N' | 'A' | 'L' | 'P';

/**
 * CVSS v3 Attack Complexity metric
 */
export type AC = 'L' | 'H';

/**
 * CVSS v3 Privileges Required metric
 */
export type PR = 'N' | 'L' | 'H';

/**
 * CVSS v3 User Interaction metric
 */
export type UI = 'N' | 'R';

/**
 * CVSS v3 Scope metric
 */
export type S = 'U' | 'C';

/**
 * CVSS v3 Confidentiality Impact metric
 */
export type C = 'H' | 'L' | 'N';

/**
 * CVSS v3 Integrity Impact metric
 */
export type I = 'H' | 'L' | 'N';

/**
 * CVSS v3 Availability Impact metric
 */
export type A = 'H' | 'L' | 'N';

/**
 * CVSS v3.0 / v3.1 base metrics
 */
export interface CVSS3BaseMetrics {
  /** Attack Vector (N/A/L/P) */
  AV: AV;
  /** Attack Complexity (L/H) */
  AC: AC;
  /** Privileges Required (N/L/H) */
  PR: PR;
  /** User Interaction (N/R) */
  UI: UI;
  /** Scope (U/C) */
  S: S;
  /** Confidentiality Impact (H/L/N) */
  C: C;
  /** Integrity Impact (H/L/N) */
  I: I;
  /** Availability Impact (H/L/N) */
  A: A;
}

/**
 * CVSS v3 temporal metrics
 */
export interface CVSS3TemporalMetrics {
  /** Exploit Code Maturity (X/U/F/P/H/R) */
  E: 'X' | 'U' | 'F' | 'P' | 'H' | 'R';
  /** Remediation Level (X/U/O/T/W) */
  RL: 'X' | 'U' | 'O' | 'T' | 'W';
  /** Report Confidence (X/U/C/R) */
  RC: 'X' | 'U' | 'C' | 'R';
}

/**
 * CVSS v3 environmental metrics
 */
export interface CVSS3EnvironmentalMetrics {
  /** Confidentiality Requirement (H/M/L/N) */
  CR: 'H' | 'M' | 'L' | 'N';
  /** Integrity Requirement (H/M/L/N) */
  IR: 'H' | 'M' | 'L' | 'N';
  /** Availability Requirement (H/M/L/N) */
  AR: 'H' | 'M' | 'L' | 'N';
  /** Modified Base Scope (X/U/C) */
  MS: 'X' | 'U' | 'C';
  /** Modified Confidentiality (H/L/N) */
  MC: 'H' | 'L' | 'N';
  /** Modified Integrity (H/L/N) */
  MI: 'H' | 'L' | 'N';
  /** Modified Availability (H/L/N) */
  MA: 'H' | 'L' | 'N';
}

/**
 * Complete CVSS v3 metrics
 */
export interface CVSS3Metrics extends CVSS3BaseMetrics {
  /** Temporal metrics */
  temporal?: CVSS3TemporalMetrics;
  /** Environmental metrics */
  environmental?: CVSS3EnvironmentalMetrics;
}

/**
 * CVSS v3 score breakdown
 */
export interface CVSS3ScoreBreakdown {
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Temporal score */
  temporalScore?: number;
  /** Environmental score */
  environmentalScore?: number;
  /** Exploitability sub-score */
  exploitabilityScore?: number;
  /** Impact sub-score */
  impactScore?: number;
  /** Base severity */
  baseSeverity: CVSSSeverity;
  /** Temporal severity */
  temporalSeverity?: CVSSSeverity;
  /** Environmental severity */
  environmentalSeverity?: CVSSSeverity;
  /** Overall score */
  score?: number;
  /** Final severity (computed from sub-scores) */
  finalSeverity?: CVSSSeverity;
}

// ============================================================================
// CVSS v4.0 Types
// ============================================================================

/**
 * CVSS v4.0 Attack Vector metrics
 */
export type AV4 = 'N' | 'A' | 'L' | 'P';

/**
 * CVSS v4.0 Attack Complexity metrics
 */
export type AC4 = 'L' | 'H';

/**
 * CVSS v4.0 Attack Requirements metrics
 */
export type AT4 = 'N' | 'P' | 'R';

/**
 * CVSS v4.0 Privileges Required metrics
 */
export type PR4 = 'N' | 'L' | 'H';

/**
 * CVSS v4.0 User Interaction metrics
 */
export type UI4 = 'N' | 'P' | 'A';

/**
 * CVSS v4.0 Vulnerable System Impact metrics
 */
export type VC4 = 'H' | 'L' | 'N';

/**
 * CVSS v4.0 Subsequent System Impact metrics
 */
export type VS4 = 'H' | 'L' | 'N';

/**
 * CVSS v4.0 Safety metrics
 */
export type S4 = 'X' | 'N' | 'P';

/**
 * CVSS v4.0 Automation Impact metrics
 */
export type AU4 = 'N' | 'P' | 'A';

/**
 * CVSS v4.0 Base metrics
 */
export interface CVSS4BaseMetrics {
  /** Attack Vector (N/A/L/P) */
  AV: AV4;
  /** Attack Complexity (L/H) */
  AC: AC4;
  /** Attack Requirements (N/P/R) */
  AT: AT4;
  /** Privileges Required (N/L/H) */
  PR: PR4;
  /** User Interaction (N/P/A) */
  UI: UI4;
  /** Vulnerable System Confidentiality (H/L/N) */
  VC: VC4;
  /** Vulnerable System Integrity (H/L/N) */
  VI: VS4;
  /** Vulnerable System Availability (H/L/N) */
  VA: VS4;
  /** Subsequent System Confidentiality (H/L/N) */
  SC: VS4;
  /** Subsequent System Integrity (H/L/N) */
  SI: VS4;
  /** Subsequent System Availability (H/L/N) */
  SA: VS4;
  /** Safety (X/N/P) */
  S: S4;
  /** Automation (N/P/A) */
  AU: AU4;
}

/**
 * CVSS v4.0 Threat metrics
 */
export interface CVSS4ThreatMetrics {
  /** Exploit Maturity (X/U/P/F/H/A/R) */
  E: 'X' | 'U' | 'P' | 'F' | 'H' | 'A' | 'R';
  /** Motivation (X/N/P/A/R/E) */
  M: 'X' | 'N' | 'P' | 'A' | 'R' | 'E';
  /** Value Density (X/N/L/M/H) */
  D: 'X' | 'N' | 'L' | 'M' | 'H';
  /** Provider (E for Impact, X for None) */
  I: 'E' | 'X';
}

/**
 * CVSS v4.0 base score ranges for I:E
 */
export interface CVSS4ImpactRanges {
  /** IS or VS range for N/H/L/I/S */
  N?: [number, number];
  /** IS or VS range for N/H/L/I/S */
  H?: [number, number];
  /** IS or VS range for N/H/L/I/S */
  L?: [number, number];
}

/**
 * CVSS v4.0 Environmental metrics
 */
export interface CVSS4EnvironmentalMetrics {
  /** Confidentiality Requirement (H/M/L/N) */
  CR: 'H' | 'M' | 'L' | 'N';
  /** Integrity Requirement (H/M/L/N) */
  IR: 'H' | 'M' | 'L' | 'N';
  /** Availability Requirement (H/M/L/N) */
  AR: 'H' | 'M' | 'L' | 'N';
  /** Modified Base Attack Vector */
  MAV?: AV4;
  /** Modified Base Attack Complexity */
  MAC?: AC4;
  /** Modified Base Attack Requirements */
  MAT?: AT4;
  /** Modified Base Privileges Required */
  MPR?: PR4;
  /** Modified Base User Interaction */
  MUI?: UI4;
  /** Modified Vulnerable System Confidentiality */
  MVC?: VC4;
  /** Modified Vulnerable System Integrity */
  MVI?: VS4;
  /** Modified Vulnerable System Availability */
  MVA?: VS4;
  /** Modified Subsequent System Confidentiality */
  MSC?: VS4;
  /** Modified Subsequent System Integrity */
  MSI?: VS4;
  /** Modified Subsequent System Availability */
  MSA?: VS4;
  /** Safety (X/N/P) */
  MS?: S4;
  /** Automation (N/P/A) */
  MAU?: AU4;
  /** Provider (E for Impact, X for None) */
  MI?: 'E' | 'X';
}

/**
 * CVSS v4.0 score ranges for environmental impact
 */
export interface CVSS4ModifiedRanges {
  /** Modified IS or VS range */
  MVS?: [number, number];
  /** Modified IS or VS range */
  MH?: [number, number];
  /** Modified IS or VS range */
  ML?: [number, number];
}

/**
 * Complete CVSS v4.0 metrics
 */
export interface CVSS4Metrics extends CVSS4BaseMetrics {
  /** Threat metrics */
  threat?: CVSS4ThreatMetrics;
  /** Environmental metrics */
  environmental?: CVSS4EnvironmentalMetrics;
}

/**
 * CVSS v4.0 score breakdown
 */
export interface CVSS4ScoreBreakdown {
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Threat score */
  threatScore?: number;
  /** Environmental score */
  environmentalScore?: number;
  /** Base severity */
  baseSeverity: CVSSSeverity;
  /** Threat severity */
  threatSeverity?: CVSSSeverity;
  /** Environmental severity */
  environmentalSeverity?: CVSSSeverity;
  /** Final severity (computed from sub-scores) */
  finalSeverity?: CVSSSeverity;
}

// ============================================================================
// CVSS Calculator State
// ============================================================================

/**
 * Current CVSS calculator state
 */
export interface CVSSCalculatorState {
  /** Selected CVSS version */
  version: CVSSVersion;
  /** Current metrics based on version */
  metrics: CVSS3Metrics | CVSS4Metrics;
  /** Calculated score breakdown */
  scores: CVSS3ScoreBreakdown | CVSS4ScoreBreakdown;
  /** Generated vector string */
  vectorString: string;
  /** Show/hide state for metric sections */
  showTemporal?: boolean;
  showExploitability?: boolean;
  showEnvironmental?: boolean;
  /** Show descriptions */
  showDescriptions?: boolean;
  /** View mode */
  viewMode?: 'compact' | 'full';
}

/**
 * CVSS calculator configuration
 */
export interface CVSSCalculatorConfig {
  /** Default version */
  defaultVersion?: CVSSVersion;
  /** Enable temporal metrics */
  enableTemporal?: boolean;
  /** Enable environmental metrics */
  enableEnvironmental?: boolean;
  /** Show metric descriptions */
  showDescriptions?: boolean;
}

// ============================================================================
// CVSS Export Types
// ============================================================================

/**
 * Export format for CVSS data
 */
export type CVSSExportFormat = 'json' | 'csv' | 'url';

/**
 * JSON export structure
 */
export interface CVSSExportJSON {
  /** CVSS version */
  version: CVSSVersion;
  /** Vector string */
  vectorString: string;
  /** Base score */
  baseScore: number;
  /** Severity rating */
  severity: CVSSSeverity;
  /** All metrics */
  metrics: CVSS3Metrics | CVSS4Metrics;
  /** Score breakdown */
  scoreBreakdown: CVSS3ScoreBreakdown | CVSS4ScoreBreakdown;
  /** Export timestamp */
  exportedAt: string;
}

/**
 * CSV export row
 */
export interface CVSSExportCSV {
  /** CVSS version */
  version: string;
  /** Vector string */
  vectorString: string;
  /** Base score */
  baseScore: string;
  /** Severity */
  severity: string;
  /** Metrics as JSON string */
  metrics: string;
  /** Export timestamp */
  exportedAt: string;
}

/**
 * URL sharing configuration
 */
export interface CVSSShareURL {
  /** Generated URL */
  url: string;
  /** Shortened URL (optional) */
  shortUrl?: string;
  /** QR code data URL (optional) */
  qrCodeUrl?: string;
  /** Expiration timestamp */
  expiresAt?: string;
}

/**
 * CVSS export result
 */
export interface CVSSExportResult {
  /** Export format */
  format: CVSSExportFormat;
  /** Exported data */
  data: CVSSExportJSON | CVSSExportCSV | CVSSShareURL;
  /** Success status */
  success: boolean;
}

// ============================================================================
// CVSS Metric Metadata
// ============================================================================

/**
 * Metric definition for UI selector
 */
export interface MetricDefinition<T = string> {
  /** Metric value */
  value: T;
  /** Display label */
  label: string;
  /** Short description */
  shortDesc: string;
  /** Full description */
  description: string;
  /** Weight in formula (if applicable) */
  weight?: number;
}

/**
 * Metric group for UI organization
 */
export interface MetricGroup {
  /** Group identifier */
  id: string;
  /** Group name */
  name: string;
  /** Group description */
  description: string;
  /** Metrics in this group */
  metrics: MetricDefinition<string>[];
}

/**
 * CVSS version metadata
 */
export interface CVSSVersionMetadata {
  /** Version identifier */
  version: CVSSVersion;
  /** Display name */
  name: string;
  /** Specification URL */
  specUrl: string;
  /** Release date */
  releaseDate?: string;
  /** Metric groups */
  metricGroups: MetricGroup[];
  /** Available metrics */
  availableMetrics: Record<string, MetricDefinition<string>[]>;
}
