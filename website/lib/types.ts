// ============================================================================
// Sysmon Types (from sysmon RPC API)
// ============================================================================

export interface SysMetrics {
  cpuUsage: number;
  memoryUsage: number;
  // Optional expanded fields returned by sysmon
  loadAvg?: number[] | number;
  uptime?: number;
  // disk is an object keyed by mount path
  disk?: Record<string, { total: number; used: number }>;
  // compatibility totals
  diskTotal?: number;
  diskUsage?: number;
  // network totals and per-interface breakdown
  netRx?: number;
  netTx?: number;
  network?: Record<string, { rx: number; tx: number }>;
  swapUsage?: number;
}
/**
 * TypeScript types mirroring Go structs from the v2e backend
 * Generated from pkg/cve/types.go and RPC API specifications
 */

// ============================================================================
// CVE Data Types (from pkg/cve/types.go)
// ============================================================================

export interface Description {
  lang: string;
  value: string;
}

export interface CVETag {
  sourceIdentifier: string;
  tags?: string[];
}

export interface Weakness {
  source: string;
  type: string;
  description: Description[];
}

export interface CPEMatch {
  vulnerable: boolean;
  criteria: string;
  matchCriteriaId: string;
  versionStartExcluding?: string;
  versionStartIncluding?: string;
  versionEndExcluding?: string;
  versionEndIncluding?: string;
}

export interface Node {
  operator: string;
  negate?: boolean;
  cpeMatch: CPEMatch[];
}

export interface Config {
  operator?: string;
  negate?: boolean;
  nodes: Node[];
}

export interface Reference {
  url: string;
  source?: string;
  tags?: string[];
}

export interface VendorComment {
  organization: string;
  comment: string;
  lastModified: string;
}

export interface CVSSDataV3 {
  version: string;
  vectorString: string;
  baseScore: number;
  baseSeverity: string;
  attackVector?: string;
  attackComplexity?: string;
  privilegesRequired?: string;
  userInteraction?: string;
  scope?: string;
  confidentialityImpact?: string;
  integrityImpact?: string;
  availabilityImpact?: string;
  exploitabilityScore?: number;
  impactScore?: number;
}

export interface CVSSMetricV3 {
  source: string;
  type: string;
  cvssData: CVSSDataV3;
  exploitabilityScore?: number;
  impactScore?: number;
}

export interface CVSSDataV2 {
  version: string;
  vectorString: string;
  baseScore: number;
  accessVector?: string;
  accessComplexity?: string;
  authentication?: string;
  confidentialityImpact?: string;
  integrityImpact?: string;
  availabilityImpact?: string;
}

export interface CVSSMetricV2 {
  source: string;
  type: string;
  cvssData: CVSSDataV2;
  baseSeverity?: string;
  exploitabilityScore?: number;
  impactScore?: number;
}

export interface CVSSDataV40 {
  version: string;
  vectorString: string;
  baseScore: number;
  baseSeverity: string;
  attackVector?: string;
  attackComplexity?: string;
  attackRequirements?: string;
  privilegesRequired?: string;
  userInteraction?: string;
  vulnConfidentialityImpact?: string;
  vulnIntegrityImpact?: string;
  vulnAvailabilityImpact?: string;
  subConfidentialityImpact?: string;
  subIntegrityImpact?: string;
  subAvailabilityImpact?: string;
}

export interface CVSSMetricV40 {
  source: string;
  type: string;
  cvssData: CVSSDataV40;
}

export interface Metrics {
  cvssMetricV40?: CVSSMetricV40[];
  cvssMetricV31?: CVSSMetricV3[];
  cvssMetricV30?: CVSSMetricV3[];
  cvssMetricV2?: CVSSMetricV2[];
}

export interface CVEItem {
  id: string;
  sourceIdentifier: string;
  published: string;
  lastModified: string;
  vulnStatus: string;
  evaluatorComment?: string;
  evaluatorSolution?: string;
  evaluatorImpact?: string;
  cisaExploitAdd?: string;
  cisaActionDue?: string;
  cisaRequiredAction?: string;
  cisaVulnerabilityName?: string;
  cveTags?: CVETag[];
  descriptions: Description[];
  metrics?: Metrics;
  weaknesses?: Weakness[];
  configurations?: Config[];
  references?: Reference[];
  vendorComments?: VendorComment[];
}

export interface CVEResponse {
  resultsPerPage: number;
  startIndex: number;
  totalResults: number;
  format: string;
  version: string;
  timestamp: string;
  vulnerabilities: Array<{
    cve: CVEItem;
  }>;
}

// ============================================================================
// RPC Request/Response Types
// ============================================================================

export interface RPCRequest<T = unknown> {
  method: string;
  params?: T;
  target?: string;
}

export interface RPCResponse<T = unknown> {
  retcode: number;
  message: string;
  payload: T | null;
}

// ============================================================================
// CVE Meta Service RPC Types
// ============================================================================

export interface GetCVERequest {
  cveId: string;
}

export interface GetCVEResponse {
  cve: CVEItem;
  source: 'local' | 'remote';
}

export interface CreateCVERequest {
  cveId: string;
}

export interface CreateCVEResponse {
  success: boolean;
  cveId: string;
  cve: CVEItem;
}

export interface UpdateCVERequest {
  cveId: string;
}

export interface UpdateCVEResponse {
  success: boolean;
  cveId: string;
  cve: CVEItem;
}

export interface DeleteCVERequest {
  cveId: string;
}

export interface DeleteCVEResponse {
  success: boolean;
  cveId: string;
}

export interface ListCVEsRequest {
  offset?: number;
  limit?: number;
}

export interface ListCVEsResponse {
  cves: CVEItem[];
  total: number;
  offset: number;
  limit: number;
}

export interface CountCVEsResponse {
  count: number;
}

// ============================================================================
// Job Session Types
// ============================================================================

export interface StartSessionRequest {
  sessionId: string;
  startIndex?: number;
  resultsPerBatch?: number;
}

export interface StartSessionResponse {
  success: boolean;
  sessionId: string;
  state: string;
  createdAt: string;
}

export interface StopSessionResponse {
  success: boolean;
  sessionId: string;
  fetchedCount: number;
  storedCount: number;
  errorCount: number;
}

export interface SessionStatus {
  hasSession: boolean;
  sessionId?: string;
  state?: string;
  dataType?: string;
  startIndex?: number;
  resultsPerBatch?: number;
  createdAt?: string;
  updatedAt?: string;
  fetchedCount?: number;
  storedCount?: number;
  errorCount?: number;
  errorMessage?: string;
  progress?: Record<string, DataProgress>;
  params?: Record<string, unknown>;
}

export interface DataProgress {
  totalCount: number;
  processedCount: number;
  errorCount: number;
  startTime: string;
  lastUpdate: string;
  errorMessage?: string;
}

export interface PauseJobResponse {
  success: boolean;
  state: string;
}

export interface ResumeJobResponse {
  success: boolean;
  state: string;
}

// ============================================================================
// CWE View Job RPC Types
// ============================================================================

export interface StartCWEViewJobRequest {
  sessionId?: string;
  startIndex?: number;
  resultsPerBatch?: number;
}

export interface StartCWEViewJobResponse {
  success: boolean;
  sessionId: string;
}

export interface StopCWEViewJobResponse {
  success: boolean;
  sessionId?: string;
}

// ============================================================================
// Utility Types
// ============================================================================

export type JobState = 'idle' | 'running' | 'paused';

export interface HealthResponse {
  status: string;
}

// ============================================================================
// Graph Analysis RPC Types (from cmd/v2analysis/service.md)
// ============================================================================

export interface GraphStats {
  node_count: number;
  edge_count: number;
}

export interface GraphNode {
  urn: string;
  properties: Record<string, unknown>;
}

export interface GraphEdge {
  from: string;
  to: string;
  type: string;
  properties?: Record<string, unknown>;
}

export interface GraphPath {
  path: string[];
  length: number;
}

export interface GetGraphStatsRequest {}

export interface GetGraphStatsResponse {
  node_count: number;
  edge_count: number;
}

export interface AddNodeRequest {
  urn: string;
  properties?: Record<string, unknown>;
}

export interface AddNodeResponse {
  urn: string;
}

export interface AddEdgeRequest {
  from: string;
  to: string;
  type: string;
  properties?: Record<string, unknown>;
}

export interface AddEdgeResponse {
  from: string;
  to: string;
  type: string;
}

export interface GetNodeRequest {
  urn: string;
}

export interface GetNodeResponse {
  urn: string;
  properties: Record<string, unknown>;
}

export interface GetNeighborsRequest {
  urn: string;
}

export interface GetNeighborsResponse {
  neighbors: string[];
}

export interface FindPathRequest {
  from: string;
  to: string;
}

export interface FindPathResponse {
  path: string[];
  length: number;
}

export interface GetNodesByTypeRequest {
  type: string;
}

export interface GetNodesByTypeResponse {
  nodes: Array<{
    urn: string;
    properties: Record<string, unknown>;
  }>;
  count: number;
}

export interface BuildCVEGraphRequest {
  limit?: number;
}

export interface BuildCVEGraphResponse {
  nodes_added: number;
  edges_added: number;
  total_nodes: number;
  total_edges: number;
}

export interface ClearGraphRequest {}

export interface ClearGraphResponse {
  status: string;
}

export interface GetFSMStateRequest {}

export interface GetFSMStateResponse {
  analyze_state: string;
  graph_state: string;
}

export interface PauseAnalysisRequest {}

export interface PauseAnalysisResponse {
  status: string;
}

export interface ResumeAnalysisRequest {}

export interface ResumeAnalysisResponse {
  status: string;
}

export interface SaveGraphRequest {}

export interface SaveGraphResponse {
  status: string;
  node_count: number;
  edge_count: number;
  last_saved: string;
}

export interface LoadGraphRequest {}

export interface LoadGraphResponse {
  status: string;
  node_count: number;
  edge_count: number;
}

// ============================================================================
// CWE Data Types (from pkg/cwe/types.go)
// ============================================================================

export interface CWEItem {
  id: string;
  name: string;
  diagram?: string;
  abstraction: string;
  structure: string;
  status: string;
  description: string;
  extendedDescription?: string;
  likelihoodOfExploit?: string;
  relatedWeaknesses?: RelatedWeakness[];
  weaknessOrdinalities?: WeaknessOrdinality[];
  applicablePlatforms?: ApplicablePlatform[];
  backgroundDetails?: string[];
  alternateTerms?: AlternateTerm[];
  modesOfIntroduction?: ModeOfIntroduction[];
  commonConsequences?: Consequence[];
  detectionMethods?: DetectionMethod[];
  potentialMitigations?: Mitigation[];
  demonstrativeExamples?: DemonstrativeExample[];
  observedExamples?: ObservedExample[];
  functionalAreas?: string[];
  affectedResources?: string[];
  taxonomyMappings?: TaxonomyMapping[];
  relatedAttackPatterns?: string[];
  references?: Reference[];
  mappingNotes?: MappingNotes;
  notes?: Note[];
  contentHistory?: ContentHistory[];
}

// ============================================================================
// ASVS Data Types (from pkg/asvs/types.go)
// ============================================================================

export interface ASVSItem {
  requirementID: string;
  chapter: string;
  section: string;
  description: string;
  level1: boolean;
  level2: boolean;
  level3: boolean;
  cwe?: string;
}

// ============================================================================
// CAPEC Data Types
// ============================================================================

export interface CAPECRelatedWeakness {
  cweId?: string;
}

export interface CAPECItem {
  id: string; // e.g. "CAPEC-123"
  name: string;
  summary?: string;
  description?: string;
  status?: string;
  likelihood?: string;
  typicalSeverity?: string;
  relatedWeaknesses?: CAPECRelatedWeakness[];
  references?: Reference[];
}

// ATT&CK Types
export interface AttackTechnique {
  id: string; // e.g. "T1001"
  name: string;
  description?: string;
  domain?: string;
  platform?: string;
  created?: string;
  modified?: string;
  revoked?: boolean;
  deprecated?: boolean;
}

export interface AttackTactic {
  id: string; // e.g. "TA0001"
  name: string;
  description?: string;
  domain?: string;
  created?: string;
  modified?: string;
}

export interface AttackMitigation {
  id: string; // e.g. "M1001"
  name: string;
  description?: string;
  domain?: string;
  created?: string;
  modified?: string;
}

export interface AttackSoftware {
  id: string; // e.g. "S0001"
  name: string;
  description?: string;
  type?: string; // e.g. "malware", "tool"
  domain?: string;
  created?: string;
  modified?: string;
}

export interface AttackGroup {
  id: string; // e.g. "G0001"
  name: string;
  description?: string;
  domain?: string;
  created?: string;
  modified?: string;
}

export interface AttackListResponse {
  techniques?: AttackTechnique[];
  tactics?: AttackTactic[];
  mitigations?: AttackMitigation[];
  offset: number;
  limit: number;
  total: number;
}

export interface RelatedWeakness {
  nature: string;
  cweId: string;
  viewId: string;
  ordinal?: string;
}

export interface WeaknessOrdinality {
  ordinality: string;
  description?: string;
}

export interface ApplicablePlatform {
  type: string;
  name?: string;
  class?: string;
  prevalence: string;
}

export interface AlternateTerm {
  term: string;
  description?: string;
}

export interface ModeOfIntroduction {
  phase: string;
  note?: string;
}

export interface Consequence {
  scope: string[];
  impact?: string[];
  likelihood?: string[];
  note?: string;
}

export interface DetectionMethod {
  detectionMethodId?: string;
  method: string;
  description: string;
  effectiveness?: string;
  effectivenessNotes?: string;
}

export interface Mitigation {
  mitigationId?: string;
  phase?: string[];
  strategy: string;
  description: string;
  effectiveness?: string;
  effectivenessNotes?: string;
}

export interface DemonstrativeExample {
  id?: string;
  entries: DemonstrativeEntry[];
}

export interface DemonstrativeEntry {
  introText?: string;
  bodyText?: string;
  nature?: string;
  language?: string;
  exampleCode?: string;
  reference?: string;
}

export interface ObservedExample {
  reference: string;
  description: string;
  link: string;
}

export interface TaxonomyMapping {
  taxonomyName: string;
  entryName?: string;
  entryId?: string;
  mappingFit?: string;
}

export interface MappingNotes {
  usage: string;
  rationale: string;
  comments: string;
  reasons: string[];
  suggestions?: SuggestionComment[];
}

export interface SuggestionComment {
  comment: string;
  cweId: string;
}

export interface Note {
  type: string;
  note: string;
}

export interface ContentHistory {
  type: string;
  submissionName?: string;
  submissionOrganization?: string;
  submissionDate?: string;
  submissionVersion?: string;
  submissionReleaseDate?: string;
  submissionComment?: string;
  modificationName?: string;
  modificationOrganization?: string;
  modificationDate?: string;
  modificationVersion?: string;
  modificationReleaseDate?: string;
  modificationComment?: string;
  contributionName?: string;
  contributionOrganization?: string;
  contributionDate?: string;
  contributionVersion?: string;
  contributionReleaseDate?: string;
  contributionComment?: string;
  contributionType?: string;
  previousEntryName?: string;
  date?: string;
  version?: string;
}

export interface ListCWEsRequest {
  offset?: number;
  limit?: number;
  search?: string;
}

export interface ListCWEsResponse {
  cwes: CWEItem[];
  offset: number;
  limit: number;
  total: number;
}

// ============================================================================
// ASVS RPC Types
// ============================================================================

export interface ListASVSRequest {
  offset?: number;
  limit?: number;
  chapter?: string;
  level?: number;
}

export interface ListASVSResponse {
  requirements: ASVSItem[];
  offset: number;
  limit: number;
  total: number;
}

export interface GetASVSByIDRequest {
  requirementId: string;
}

export interface GetASVSByIDResponse extends ASVSItem {}

export interface ImportASVSRequest {
  url: string;
}

export interface ImportASVSResponse {
  success: boolean;
}

// CWE View Types (from pkg/cwe/views.go)
export interface CWEViewMember {
  cweId: string;
  role?: string;
}

export interface CWEViewStakeholder {
  type: string;
  description?: string;
}

export interface CWEView {
  id: string;
  name?: string;
  type?: string;
  status?: string;
  objective?: string;
  audience?: CWEViewStakeholder[];
  members?: CWEViewMember[];
  references?: Reference[];
  notes?: Note[];
  contentHistory?: ContentHistory[];
  raw?: unknown;
}

export interface ListCWEViewsRequest {
  offset?: number;
  limit?: number;
}

export interface ListCWEViewsResponse {
  views: CWEView[];
  offset: number;
  limit: number;
  total: number;
}

export interface GetCWEViewResponse {
  view: CWEView;
}

// ============================================================================
// Notes Framework Types
// ============================================================================

// Bookmark Types
export interface Bookmark {
  id: number;
  global_item_id: string;
  item_type: string;
  item_id: string;
  urn: string; // URN reference (e.g., v2e::nvd::cve::CVE-2021-1234)
  title: string;
  description: string;
  author?: string;
  is_private: boolean;
  created_at: string;
  updated_at: string;
  metadata: Record<string, unknown>;
}

export interface CreateBookmarkRequest {
  global_item_id: string;
  item_type: string;
  item_id: string;
  title: string;
  description: string;
  author?: string;
  is_private?: boolean;
  metadata?: Record<string, unknown>;
}

export interface CreateBookmarkResponse {
  success: boolean;
  bookmark: Bookmark;
  memoryCard?: MemoryCard;
}

export interface GetBookmarkRequest {
  id: number;
}

export interface GetBookmarkResponse {
  bookmark: Bookmark;
}

export interface ListBookmarksRequest {
  offset?: number;
  limit?: number;
  item_type?: string;
  item_id?: string;
  author?: string;
  is_private?: boolean;
}

export interface ListBookmarksResponse {
  bookmarks: Bookmark[];
  offset: number;
  limit: number;
  total: number;
}

export interface UpdateBookmarkRequest {
  id: number;
  title?: string;
  description?: string;
  author?: string;
  is_private?: boolean;
  metadata?: Record<string, unknown>;
}

export interface UpdateBookmarkResponse {
  success: boolean;
  bookmark: Bookmark;
}

export interface DeleteBookmarkRequest {
  id: number;
}

export interface DeleteBookmarkResponse {
  success: boolean;
}

// Note Types
export interface NoteModel {
  id: number;
  bookmark_id: number;
  content: string;
  author?: string;
  is_private: boolean;
  created_at: string;
  updated_at: string;
  metadata: Record<string, unknown>;
}

export interface AddNoteRequest {
  bookmark_id: number;
  content: string;
  author?: string;
  is_private?: boolean;
  metadata?: Record<string, unknown>;
}

export interface AddNoteResponse {
  success: boolean;
  note: NoteModel;
}

export interface GetNoteRequest {
  id: number;
}

export interface GetNoteResponse {
  note: NoteModel;
}

export interface GetNotesByBookmarkRequest {
  bookmark_id: number;
  offset?: number;
  limit?: number;
}

export interface GetNotesByBookmarkResponse {
  notes: NoteModel[];
  offset: number;
  limit: number;
  total: number;
}

export interface UpdateNoteRequest {
  id: number;
  content?: string;
  author?: string;
  is_private?: boolean;
  metadata?: Record<string, unknown>;
}

export interface UpdateNoteResponse {
  success: boolean;
  note: NoteModel;
}

export interface DeleteNoteRequest {
  id: number;
}

export interface DeleteNoteResponse {
  success: boolean;
}

// Memory Card Types
export interface MemoryCard {
  id: number;
  bookmark_id: number;
  urn: string;
  front: string; // Front content (question)
  back: string; // Back content (answer)
  front_content: string; // Front content (question) - alias for frontend compatibility
  back_content: string; // Back content (answer) - alias for frontend compatibility
  major_class: string;
  minor_class: string;
  status: string;
  content: any; // TipTap JSON
  card_type: string; // Card type: basic, cloze, reverse
  learning_state: string; // Derived from bookmark.learning_state, not stored on card
  author: string; // Card creator/author
  is_private: boolean; // Whether card is private
  interval: number; // Days until next review
  ease_factor: number; // SM-2 algorithm factor
  repetitions: number; // Number of times reviewed
  created_at: string;
  updated_at: string;
  next_review_at: string;
  metadata: Record<string, unknown>;
}

export interface CreateMemoryCardRequest {
  bookmark_id: number;
  front: string; // Note: backend expects 'front', stored as front_content in model
  back: string; // Note: backend expects 'back', stored as back_content in model
  major_class?: string;
  minor_class?: string;
  status?: string;
  content?: any; // TipTap JSON
  card_type?: string; // Card type: basic, cloze, reverse
  author?: string; // Card creator/author
  is_private?: boolean; // Whether card is private
  metadata?: Record<string, unknown>;
}

export interface CreateMemoryCardResponse {
  success: boolean;
  memory_card: MemoryCard;
}

export interface GetMemoryCardRequest {
  id: number;
}

export interface GetMemoryCardResponse {
  memory_card: MemoryCard;
}

export interface ListMemoryCardsRequest {
  bookmark_id?: number;
  learning_state?: string;
  author?: string;
  is_private?: boolean;
  offset?: number;
  limit?: number;
}

export interface ListMemoryCardsResponse {
  memory_cards: MemoryCard[];
  offset: number;
  limit: number;
  total: number;
}

export interface UpdateMemoryCardRequest {
  card_id: number; // Note: backend expects 'card_id' not 'id'
  front?: string; // Note: backend expects 'front', stored as front_content in model
  back?: string; // Note: backend expects 'back', stored as back_content in model
  major_class?: string;
  minor_class?: string;
  status?: string;
  content?: any; // TipTap JSON
  learning_state?: string; // Derived from bookmark, use bookmark RPC to update
  author?: string; // Card creator/author
  is_private?: boolean; // Whether card is private
  interval?: number;
  ease_factor?: number;
  repetitions?: number;
  next_review_at?: string;
  metadata?: Record<string, unknown>;
}

export interface UpdateMemoryCardResponse {
  success: boolean;
  memory_card: MemoryCard;
}

export interface DeleteMemoryCardRequest {
  card_id: number; // Note: backend expects 'card_id' not 'id'
}

export interface DeleteMemoryCardResponse {
  success: boolean;
}

export interface RateMemoryCardRequest {
  card_id: number; // Note: backend expects 'card_id' not 'id'
  rating: string; // 'again', 'hard', 'good', 'easy'
}

export interface RateMemoryCardResponse {
  success: boolean;
  memory_card: MemoryCard;
}

// Cross Reference Types
export interface CrossReference {
  id: number;
  from_item_id: string;
  from_item_type: string;
  to_item_id: string;
  to_item_type: string;
  relationship_type: string; // 'related_to', 'depends_on', 'similar_to', 'opposite_of', etc.
  description?: string;
  strength: number; // 1-5 scale
  author?: string;
  is_private: boolean;
  created_at: string;
  updated_at: string;
  metadata: Record<string, unknown>;
}

export interface CreateCrossReferenceRequest {
  from_item_id: string;
  from_item_type: string;
  to_item_id: string;
  to_item_type: string;
  relationship_type: string;
  description?: string;
  strength?: number;
  author?: string;
  is_private?: boolean;
  metadata?: Record<string, unknown>;
}

export interface CreateCrossReferenceResponse {
  success: boolean;
  cross_reference: CrossReference;
}

export interface GetCrossReferenceRequest {
  id: number;
}

export interface GetCrossReferenceResponse {
  cross_reference: CrossReference;
}

export interface ListCrossReferencesRequest {
  from_item_id?: string;
  from_item_type?: string;
  to_item_id?: string;
  to_item_type?: string;
  relationship_type?: string;
  author?: string;
  is_private?: boolean;
  offset?: number;
  limit?: number;
}

export interface ListCrossReferencesResponse {
  cross_references: CrossReference[];
  offset: number;
  limit: number;
  total: number;
}

export interface UpdateCrossReferenceRequest {
  id: number;
  relationship_type?: string;
  description?: string;
  strength?: number;
  author?: string;
  is_private?: boolean;
  metadata?: Record<string, unknown>;
}

export interface UpdateCrossReferenceResponse {
  success: boolean;
  cross_reference: CrossReference;
}

export interface DeleteCrossReferenceRequest {
  id: number;
}

export interface DeleteCrossReferenceResponse {
  success: boolean;
}

// History Types
export interface HistoryEntry {
  id: number;
  item_id: string;
  item_type: string;
  action: string; // 'created', 'updated', 'deleted', 'bookmarked', 'rated', etc.
  old_values?: Record<string, unknown>;
  new_values?: Record<string, unknown>;
  author?: string;
  timestamp: string;
  metadata: Record<string, unknown>;
}

export interface AddHistoryRequest {
  item_id: string;
  item_type: string;
  action: string;
  old_values?: Record<string, unknown>;
  new_values?: Record<string, unknown>;
  author?: string;
  metadata?: Record<string, unknown>;
}

export interface AddHistoryResponse {
  success: boolean;
  history_entry: HistoryEntry;
}

export interface GetHistoryRequest {
  item_id: string;
  item_type: string;
  offset?: number;
  limit?: number;
}

export interface GetHistoryResponse {
  history_entries: HistoryEntry[];
  offset: number;
  limit: number;
  total: number;
}

export interface GetHistoryByActionRequest {
  action: string;
  author?: string;
  offset?: number;
  limit?: number;
}

export interface GetHistoryByActionResponse {
  history_entries: HistoryEntry[];
  offset: number;
  limit: number;
  total: number;
}

// Bookmark State Reversion
export interface RevertBookmarkStateRequest {
  item_id: string;
  item_type: string;
  to_timestamp: string;
  author?: string;
}

export interface RevertBookmarkStateResponse {
  success: boolean;
  message: string;
}

// ============================================================================
// SSG (SCAP Security Guide) Data Types
// ============================================================================

export interface SSGGuide {
  id: string;
  product: string;
  profileId: string;
  shortId: string;
  title: string;
  htmlContent: string;
  createdAt: string;
  updatedAt: string;
}

export interface SSGGroup {
  id: string;
  guideId: string;
  parentId: string;
  title: string;
  description: string;
  level: number;
  groupCount: number;
  ruleCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface SSGReference {
  href: string;
  label: string;
  value: string;
}

export interface SSGRule {
  id: string;
  guideId: string;
  groupId: string;
  shortId: string;
  title: string;
  description: string;
  rationale: string;
  severity: 'low' | 'medium' | 'high';
  references: SSGReference[];
  level: number;
  createdAt: string;
  updatedAt: string;
}

export interface SSGTable {
  id: string;
  product: string;
  tableType: string;
  title: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface SSGTableEntry {
  id: number;
  tableId: string;
  mapping: string;
  ruleTitle: string;
  description: string;
  rationale: string;
  createdAt: string;
  updatedAt: string;
}

export interface SSGManifest {
  id: string;
  product: string;
  createdAt: string;
  updatedAt: string;
}

export interface SSGProfile {
  id: string;
  manifestId: string;
  profileId: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface SSGProfileRule {
  id: number;
  profileId: string;
  ruleShortId: string;
  createdAt: string;
}

export interface SSGTree {
  guide: SSGGuide;
  groups: SSGGroup[];
  rules: SSGRule[];
}

export interface TreeNode {
  id: string;
  parentId: string;
  level: number;
  type: 'group' | 'rule';
  group?: SSGGroup;
  rule?: SSGRule;
  children: TreeNode[];
}

// SSG RPC Request/Response Types

export interface SSGImportGuideRequest {
  path: string;
}

export interface SSGImportGuideResponse {
  success: boolean;
  guideId: string;
  groupCount: number;
  ruleCount: number;
}

export interface SSGGetGuideRequest {
  id: string;
}

export interface SSGGetGuideResponse {
  guide: SSGGuide;
}

export interface SSGListGuidesRequest {
  product?: string;
  profileId?: string;
}

export interface SSGListGuidesResponse {
  guides: SSGGuide[];
  count: number;
}

export interface SSGGetTreeRequest {
  guideId: string;
}

export interface SSGGetTreeResponse {
  tree: SSGTree;
}

export interface SSGGetTreeNodeRequest {
  guideId: string;
}

export interface SSGGetTreeNodeResponse {
  nodes: TreeNode[];
  count: number;
}

export interface SSGGetGroupRequest {
  id: string;
}

export interface SSGGetGroupResponse {
  group: SSGGroup;
}

export interface SSGGetChildGroupsRequest {
  parentId?: string;
}

export interface SSGGetChildGroupsResponse {
  groups: SSGGroup[];
  count: number;
}

export interface SSGGetRuleRequest {
  id: string;
}

export interface SSGGetRuleResponse {
  rule: SSGRule;
}

export interface SSGListRulesRequest {
  groupId?: string;
  severity?: string;
  offset?: number;
  limit?: number;
}

export interface SSGListRulesResponse {
  rules: SSGRule[];
  total: number;
}

export interface SSGGetChildRulesRequest {
  groupId: string;
}

export interface SSGGetChildRulesResponse {
  rules: SSGRule[];
  count: number;
}

// SSG Table RPC Types

export interface SSGListTablesRequest {
  product?: string;
  tableType?: string;
}

export interface SSGListTablesResponse {
  tables: SSGTable[];
  count: number;
}

export interface SSGGetTableRequest {
  id: string;
}

export interface SSGGetTableResponse {
  table: SSGTable;
}

export interface SSGGetTableEntriesRequest {
  tableId: string;
  offset?: number;
  limit?: number;
}

export interface SSGGetTableEntriesResponse {
  entries: SSGTableEntry[];
  total: number;
}

export interface SSGImportTableRequest {
  path: string;
}

export interface SSGImportTableResponse {
  success: boolean;
  tableId: string;
  entryCount: number;
}

// SSG Manifest RPC Types

export interface SSGListManifestsRequest {
  product?: string;
  limit?: number;
  offset?: number;
}

export interface SSGListManifestsResponse {
  manifests: SSGManifest[];
  count: number;
}

export interface SSGGetManifestRequest {
  manifestId: string;
}

export interface SSGGetManifestResponse {
  manifest: SSGManifest;
}

export interface SSGListProfilesRequest {
  product?: string;
  profileId?: string;
  limit?: number;
  offset?: number;
}

export interface SSGListProfilesResponse {
  profiles: SSGProfile[];
  count: number;
}

export interface SSGGetProfileRequest {
  profileId: string;
}

export interface SSGGetProfileResponse {
  profile: SSGProfile;
}

export interface SSGGetProfileRulesRequest {
  profileId: string;
  limit?: number;
  offset?: number;
}

export interface SSGGetProfileRulesResponse {
  rules: SSGProfileRule[];
  count: number;
}

// SSG Data Stream Types

export interface SSGDataStream {
  id: string;
  product: string;
  scapVersion: string;
  generated: string;
  xccdfBenchmarkId: string;
  ovalChecksId: string;
  ocilQuestionnairesId: string;
  cpeDictId: string;
  createdAt: string;
  updatedAt: string;
}

export interface SSGBenchmark {
  id: string;
  dataStreamId: string;
  xccdfId: string;
  title: string;
  version: string;
  description: string;
  totalProfiles: number;
  totalGroups: number;
  totalRules: number;
  maxGroupLevel: number;
  createdAt: string;
  updatedAt: string;
}

export interface SSGDSProfile {
  id: string;
  dataStreamId: string;
  xccdfProfileId: string;
  title: string;
  description: string;
  totalRules: number;
  createdAt: string;
  updatedAt: string;
}

export interface SSGDSProfileRule {
  id: number;
  profileId: string;
  ruleShortId: string;
  selected: boolean;
  createdAt: string;
}

export interface SSGDSGroup {
  id: string;
  dataStreamId: string;
  xccdfGroupId: string;
  parentXccdfGroupId: string;
  title: string;
  description: string;
  level: number;
  createdAt: string;
  updatedAt: string;
}

export interface SSGDSRule {
  id: string;
  dataStreamId: string;
  xccdfRuleId: string;
  groupXccdfId: string;
  shortId: string;
  title: string;
  description: string;
  rationale: string;
  severity: string;
  warning: string;
  createdAt: string;
  updatedAt: string;
  references?: SSGDSRuleReference[];
  identifiers?: SSGDSRuleIdentifier[];
}

export interface SSGDSRuleReference {
  id: number;
  ruleId: string;
  href: string;
  text: string;
  createdAt: string;
}

export interface SSGDSRuleIdentifier {
  id: number;
  ruleId: string;
  system: string;
  identifier: string;
  createdAt: string;
}

// SSG Data Stream RPC Types

export interface SSGListDataStreamsRequest {
  product?: string;
  limit?: number;
  offset?: number;
}

export interface SSGListDataStreamsResponse {
  dataStreams: SSGDataStream[];
  count: number;
}

export interface SSGGetDataStreamRequest {
  dataStreamId: string;
}

export interface SSGGetDataStreamResponse {
  dataStream: SSGDataStream;
  benchmark?: SSGBenchmark;
}

export interface SSGListDSProfilesRequest {
  dataStreamId: string;
  limit?: number;
  offset?: number;
}

export interface SSGListDSProfilesResponse {
  profiles: SSGDSProfile[];
  count: number;
}

export interface SSGGetDSProfileRequest {
  profileId: string;
}

export interface SSGGetDSProfileResponse {
  profile: SSGDSProfile;
}

export interface SSGGetDSProfileRulesRequest {
  profileId: string;
  limit?: number;
  offset?: number;
}

export interface SSGGetDSProfileRulesResponse {
  rules: SSGDSProfileRule[];
  count: number;
}

export interface SSGListDSGroupsRequest {
  dataStreamId: string;
  parentXccdfGroupId?: string;
  limit?: number;
  offset?: number;
}

export interface SSGListDSGroupsResponse {
  groups: SSGDSGroup[];
  count: number;
}

export interface SSGListDSRulesRequest {
  dataStreamId: string;
  groupXccdfId?: string;
  severity?: string;
  limit?: number;
  offset?: number;
}

export interface SSGListDSRulesResponse {
  rules: SSGDSRule[];
  total: number;
}

export interface SSGGetDSRuleRequest {
  ruleId: string;
}

export interface SSGGetDSRuleResponse {
  rule: SSGDSRule;
  references: SSGDSRuleReference[];
  identifiers: SSGDSRuleIdentifier[];
}

export interface SSGImportDataStreamRequest {
  path: string;
}

export interface SSGImportDataStreamResponse {
  success: boolean;
  dataStreamId: string;
  profileCount: number;
  groupCount: number;
  ruleCount: number;
}

// SSG Import Job RPC Types

export interface SSGStartImportJobRequest {
  runId?: string;
}

export interface SSGStartImportJobResponse {
  success: boolean;
  runId: string;
}

export interface SSGStopImportJobResponse {
  success: boolean;
}

export interface SSGPauseImportJobResponse {
  success: boolean;
}

export interface SSGResumeImportJobRequest {
  runId: string;
}

export interface SSGResumeImportJobResponse {
  success: boolean;
}

export interface SSGGetImportStatusResponse {
  id: string;
  dataType: string;
  state: 'queued' | 'running' | 'paused' | 'completed' | 'failed' | 'stopped';
  startedAt: string;
  completedAt?: string;
  error?: string;
  progress: {
    totalGuides: number;
    processedGuides: number;
    failedGuides: number;
    totalTables: number;
    processedTables: number;
    failedTables: number;
    totalManifests: number;
    processedManifests: number;
    failedManifests: number;
    totalDataStreams: number;
    processedDataStreams: number;
    failedDataStreams: number;
    currentFile: string;
    currentPhase?: string;
  };
  metadata?: Record<string, string>;
}

// SSG Remote Service RPC Types (Git operations)

export interface SSGCloneRepoResponse {
  success: boolean;
  path: string;
}

export interface SSGPullRepoResponse {
  success: boolean;
}

export interface SSGGetRepoStatusResponse {
  commitHash: string;
  branch: string;
  isClean: boolean;
}

export interface SSGListGuideFilesResponse {
  files: string[];
  count: number;
}

export interface SSGGetFilePathRequest {
  filename: string;
}

export interface SSGGetFilePathResponse {
  path: string;
}

// SSG Cross-Reference Types

export interface SSGCrossReference {
  id: number;
  sourceType: string;  // "guide", "table", "manifest", "datastream"
  sourceId: string;
  targetType: string;  // "guide", "table", "manifest", "datastream"
  targetId: string;
  linkType: string;    // "rule_id", "cce", "product", "profile_id"
  metadata: string;    // JSON string with additional context
  createdAt: string;
}

export interface SSGGetCrossReferencesRequest {
  sourceType?: string;
  sourceId?: string;
  targetType?: string;
  targetId?: string;
  limit?: number;
  offset?: number;
}

export interface SSGGetCrossReferencesResponse {
  crossReferences: SSGCrossReference[];
  count: number;
}

export interface SSGFindRelatedObjectsRequest {
  objectType: string;
  objectId: string;
  linkType?: string;
  limit?: number;
  offset?: number;
}

export interface SSGFindRelatedObjectsResponse {
  relatedObjects: SSGCrossReference[];
  count: number;
}

// ============================================================================
// UEE (Unified ETL Engine) Types
// ============================================================================

export type MacroFSMState = 
  | "BOOTSTRAPPING"
  | "ORCHESTRATING"
  | "STABILIZING"
  | "DRAINING";

export type ProviderFSMState = 
  | "IDLE"
  | "ACQUIRING"
  | "RUNNING"
  | "WAITING_QUOTA"
  | "WAITING_BACKOFF"
  | "PAUSED"
  | "TERMINATED";

export interface ProviderNode {
  id: string;
  providerType: string;
  state: ProviderFSMState;
  processedCount: number;
  errorCount: number;
  permitsHeld: number;
  lastCheckpoint?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MacroNode {
  id: string;
  state: MacroFSMState;
  providers: ProviderNode[];
  createdAt: string;
  updatedAt: string;
}

export interface ETLTree {
  macro: MacroNode;
  totalProviders: number;
  activeProviders: number;
}

export interface KernelMetrics {
  p99Latency: number;           // P99 latency in milliseconds
  bufferSaturation: number;     // Buffer saturation percentage (0-100)
  messageRate: number;          // Messages per second
  errorRate: number;            // Errors per second
  timestamp: string;            // ISO timestamp
}

export interface Checkpoint {
  urn: string;                  // URN key (v2e::provider::type::id)
  providerID: string;
  success: boolean;
  errorMessage?: string;
  processedAt: string;          // ISO timestamp
}

export interface PermitAllocation {
  providerID: string;
  permitsHeld: number;
  permitsRequested: number;
  timestamp: string;
}

// RPC Request/Response types

export interface GetEtlTreeResponse {
  tree: ETLTree;
}

export interface GetKernelMetricsResponse {
  metrics: KernelMetrics;
}

export interface GetProviderCheckpointsRequest {
  providerID: string;
  limit?: number;
  offset?: number;
}

export interface GetProviderCheckpointsResponse {
  checkpoints: Checkpoint[];
  count: number;
}

// ============================================================================
// GLC (Graphized Learning Canvas) Types
// ============================================================================

// Graph Model
export interface GLCGraph {
  id: number;
  graph_id: string;
  name: string;
  description: string;
  preset_id: string;
  tags: string;
  nodes: string; // JSON array of CADNode
  edges: string; // JSON array of CADEdge
  viewport: string; // JSON viewport state
  thumbnail?: string; // Base64 data URL for graph preview
  version: number;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
}

export interface GLCGraphVersion {
  id: number;
  graph_id: number;
  version: number;
  nodes: string;
  edges: string;
  viewport: string;
  created_at: string;
}

// Graph Request/Response Types
export interface CreateGLCGraphRequest {
  name: string;
  description?: string;
  preset_id: string;
  nodes?: string;
  edges?: string;
  viewport?: string;
  tags?: string;
}

export interface CreateGLCGraphResponse {
  success: boolean;
  graph: GLCGraph;
}

export interface GetGLCGraphRequest {
  graph_id: string;
}

export interface GetGLCGraphResponse {
  graph: GLCGraph;
}

export interface UpdateGLCGraphRequest {
  graph_id: string;
  name?: string;
  description?: string;
  nodes?: string;
  edges?: string;
  viewport?: string;
  tags?: string;
  is_archived?: boolean;
}

export interface UpdateGLCGraphResponse {
  success: boolean;
  graph: GLCGraph;
}

export interface DeleteGLCGraphRequest {
  graph_id: string;
}

export interface DeleteGLCGraphResponse {
  success: boolean;
}

export interface ListGLCGraphsRequest {
  preset_id?: string;
  offset?: number;
  limit?: number;
}

export interface ListGLCGraphsResponse {
  graphs: GLCGraph[];
  total: number;
  offset: number;
  limit: number;
}

export interface ListRecentGLCGraphsRequest {
  limit?: number;
}

export interface ListRecentGLCGraphsResponse {
  graphs: GLCGraph[];
}

// Version Request/Response Types
export interface GetGLCVersionRequest {
  graph_id: string;
  version: number;
}

export interface GetGLCVersionResponse {
  version: GLCGraphVersion;
}

export interface ListGLCVersionsRequest {
  graph_id: string;
  limit?: number;
}

export interface ListGLCVersionsResponse {
  versions: GLCGraphVersion[];
}

export interface RestoreGLCVersionRequest {
  graph_id: string;
  version: number;
}

export interface RestoreGLCVersionResponse {
  success: boolean;
  graph: GLCGraph;
}

// User Preset Model
export interface GLCUserPreset {
  id: number;
  preset_id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  theme: string; // JSON CanvasPresetTheme
  behavior: string; // JSON CanvasPresetBehavior
  node_types: string; // JSON array of NodeTypeDefinition
  relations: string; // JSON array of RelationshipDefinition
  created_at: string;
  updated_at: string;
}

// Preset Request/Response Types
export interface CreateGLCPresetRequest {
  name: string;
  version?: string;
  description?: string;
  author?: string;
  theme: object;
  behavior: object;
  node_types: object[];
  relations: object[];
}

export interface CreateGLCPresetResponse {
  success: boolean;
  preset: GLCUserPreset;
}

export interface GetGLCPresetRequest {
  preset_id: string;
}

export interface GetGLCPresetResponse {
  preset: GLCUserPreset;
}

export interface UpdateGLCPresetRequest {
  preset_id: string;
  name?: string;
  version?: string;
  description?: string;
  author?: string;
  theme?: object;
  behavior?: object;
  node_types?: object[];
  relationships?: object[];
}

export interface UpdateGLCPresetResponse {
  success: boolean;
  preset: GLCUserPreset;
}

export interface DeleteGLCPresetRequest {
  preset_id: string;
}

export interface DeleteGLCPresetResponse {
  success: boolean;
}

export interface ListGLCPresetsResponse {
  presets: GLCUserPreset[];
}

// Share Link Model
export interface GLCShareLink {
  id: number;
  link_id: string;
  graph_id: string;
  password?: string;
  expires_at?: string;
  view_count: number;
  created_at: string;
}

// Share Link Request/Response Types
export interface CreateGLCShareLinkRequest {
  graph_id: string;
  password?: string;
  expires_in_hours?: number;
}

export interface CreateGLCShareLinkResponse {
  success: boolean;
  share_link: GLCShareLink;
}

export interface GetGLCSharedGraphRequest {
  link_id: string;
  password?: string;
}

export interface GetGLCSharedGraphResponse {
  graph: GLCGraph;
}

export interface GetGLCShareEmbedDataRequest {
  link_id: string;
}

export interface GetGLCShareEmbedDataResponse {
  share_link: GLCShareLink;
  graph: GLCGraph;
}
