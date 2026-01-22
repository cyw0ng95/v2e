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
  startIndex?: number;
  resultsPerBatch?: number;
  createdAt?: string;
  updatedAt?: string;
  fetchedCount?: number;
  storedCount?: number;
  errorCount?: number;
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
