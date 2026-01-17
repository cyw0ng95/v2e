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
// Utility Types
// ============================================================================

export type JobState = 'idle' | 'running' | 'paused';

export interface HealthResponse {
  status: string;
}
