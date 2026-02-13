/**
 * RPC Client for v2e access service
 * Implements the Service-Consumer pattern to bridge UI and backend
 */

import type {
  RPCRequest,
  RPCResponse,
  SysMetrics,
  GetCVERequest,
  GetCVEResponse,
  CreateCVERequest,
  CreateCVEResponse,
  UpdateCVERequest,
  UpdateCVEResponse,
  DeleteCVERequest,
  DeleteCVEResponse,
  ListCVEsRequest,
  ListCVEsResponse,
  CountCVEsResponse,
  StartSessionRequest,
  StartSessionResponse,
  StopSessionResponse,
  SessionStatus,
  PauseJobResponse,
  ResumeJobResponse,
  StartCWEViewJobRequest,
  StartCWEViewJobResponse,
  StopCWEViewJobResponse,
  HealthResponse,
  CVEItem,
  CWEItem,
  ListCWEsRequest,
  ListCWEsResponse,
  ListCWEViewsRequest,
  ListCWEViewsResponse,
  CWEView,
  GetCWEViewResponse,
  // ASVS Types
  ASVSItem,
  ListASVSRequest,
  ListASVSResponse,
  GetASVSByIDRequest,
  GetASVSByIDResponse,
  ImportASVSRequest,
  ImportASVSResponse,
  // Notes Framework Types
  Bookmark,
  CreateBookmarkRequest,
  CreateBookmarkResponse,
  GetBookmarkRequest,
  GetBookmarkResponse,
  ListBookmarksRequest,
  ListBookmarksResponse,
  UpdateBookmarkRequest,
  UpdateBookmarkResponse,
  DeleteBookmarkRequest,
  DeleteBookmarkResponse,
  NoteModel as Note,
  AddNoteRequest,
  AddNoteResponse,
  GetNoteRequest,
  GetNoteResponse,
  GetNotesByBookmarkRequest,
  GetNotesByBookmarkResponse,
  UpdateNoteRequest,
  UpdateNoteResponse,
  DeleteNoteRequest,
  DeleteNoteResponse,
  MemoryCard,
  CreateMemoryCardRequest,
  CreateMemoryCardResponse,
  GetMemoryCardRequest,
  GetMemoryCardResponse,
  ListMemoryCardsRequest,
  ListMemoryCardsResponse,
  UpdateMemoryCardRequest,
  UpdateMemoryCardResponse,
  DeleteMemoryCardRequest,
  DeleteMemoryCardResponse,
  RateMemoryCardRequest,
  RateMemoryCardResponse,
  CrossReference,
  CreateCrossReferenceRequest,
  CreateCrossReferenceResponse,
  GetCrossReferenceRequest,
  GetCrossReferenceResponse,
  ListCrossReferencesRequest,
  ListCrossReferencesResponse,
  UpdateCrossReferenceRequest,
  UpdateCrossReferenceResponse,
  DeleteCrossReferenceRequest,
  DeleteCrossReferenceResponse,
  HistoryEntry,
  AddHistoryRequest,
  AddHistoryResponse,
  GetHistoryRequest,
  GetHistoryResponse,
  GetHistoryByActionRequest,
  GetHistoryByActionResponse,
  RevertBookmarkStateRequest,
  RevertBookmarkStateResponse,
  // Graph Analysis Types
  GetGraphStatsRequest,
  GetGraphStatsResponse,
  AddNodeRequest,
  AddNodeResponse,
  AddEdgeRequest,
  AddEdgeResponse,
  GetNodeRequest,
  GetNodeResponse,
  GetNeighborsRequest,
  GetNeighborsResponse,
  FindPathRequest,
  FindPathResponse,
  GetNodesByTypeRequest,
  GetNodesByTypeResponse,
  BuildCVEGraphRequest,
  BuildCVEGraphResponse,
  ClearGraphRequest,
  ClearGraphResponse,
  GetFSMStateRequest,
  GetFSMStateResponse,
  PauseAnalysisRequest,
  PauseAnalysisResponse,
  ResumeAnalysisRequest,
  ResumeAnalysisResponse,
  SaveGraphRequest,
  SaveGraphResponse,
  LoadGraphRequest,
  LoadGraphResponse,
  // SSG Types
  CAPECItem,
  SSGGuide,
  TreeNode,
  SSGGetTreeNodeRequest,
  SSGGetTreeNodeResponse,
  SSGGetChildGroupsRequest,
  SSGGetChildGroupsResponse,
  SSGListGuidesRequest,
  SSGListGuidesResponse,
  SSGGetGuideRequest,
  SSGGetGuideResponse,
  // GLC Types
  GLCGraph,
  GLCGraphVersion,
  GLCUserPreset,
  GLCShareLink,
  CreateGLCGraphRequest,
  CreateGLCGraphResponse,
  GetGLCGraphRequest,
  GetGLCGraphResponse,
  UpdateGLCGraphRequest,
  UpdateGLCGraphResponse,
  DeleteGLCGraphRequest,
  DeleteGLCGraphResponse,
  ListGLCGraphsRequest,
  ListGLCGraphsResponse,
  ListRecentGLCGraphsRequest,
  ListRecentGLCGraphsResponse,
  GetGLCVersionRequest,
  GetGLCVersionResponse,
  ListGLCVersionsRequest,
  ListGLCVersionsResponse,
  RestoreGLCVersionRequest,
  RestoreGLCVersionResponse,
  CreateGLCPresetRequest,
  CreateGLCPresetResponse,
  GetGLCPresetRequest,
  GetGLCPresetResponse,
  UpdateGLCPresetRequest,
  UpdateGLCPresetResponse,
  DeleteGLCPresetRequest,
  DeleteGLCPresetResponse,
  ListGLCPresetsResponse,
  CreateGLCShareLinkRequest,
  CreateGLCShareLinkResponse,
  GetGLCSharedGraphRequest,
  GetGLCSharedGraphResponse,
  GetGLCShareEmbedDataRequest,
  GetGLCShareEmbedDataResponse,
} from './types';
import { logError, logWarn, logDebug, createLogger } from './logger';

// Create component-specific logger
const logger = createLogger('rpc-client');

// ============================================================================
// Path-Based RPC Routing
// ============================================================================

const methodToPathMap: Record<string, { path: string; target: string }> = {
  // CVE
  RPCGetCVE: { path: '/cve/get', target: 'local' },
  RPCCreateCVE: { path: '/cve/create', target: 'local' },
  RPCUpdateCVE: { path: '/cve/update', target: 'local' },
  RPCDeleteCVE: { path: '/cve/delete', target: 'local' },
  RPCListCVEs: { path: '/cve/list', target: 'local' },
  RPCCountCVEs: { path: '/cve/count', target: 'local' },

  // CWE
  RPCGetCWEByID: { path: '/cwe/get', target: 'local' },
  RPCListCWEs: { path: '/cwe/list', target: 'local' },
  RPCImportCWEs: { path: '/cwe/import', target: 'local' },

  // CWE View
  RPCSaveCWEView: { path: '/cwe-view/save', target: 'local' },
  RPCGetCWEViewByID: { path: '/cwe-view/get', target: 'local' },
  RPCListCWEViews: { path: '/cwe-view/list', target: 'local' },
  RPCDeleteCWEView: { path: '/cwe-view/delete', target: 'local' },
  RPCStartCWEViewJob: { path: '/cwe-view/start-job', target: 'meta' },
  RPCStopCWEViewJob: { path: '/cwe-view/stop-job', target: 'meta' },

  // CAPEC
  RPCGetCAPECByID: { path: '/capec/get', target: 'local' },
  RPCListCAPECs: { path: '/capec/list', target: 'local' },
  RPCImportCAPECs: { path: '/capec/import', target: 'local' },
  RPCForceImportCAPECs: { path: '/capec/force-import', target: 'local' },
  RPCGetCAPECCatalogMeta: { path: '/capec/metadata', target: 'local' },

  // ATT&CK
  RPCGetAttackTechnique: { path: '/attack/technique', target: 'local' },
  RPCGetAttackTactic: { path: '/attack/tactic', target: 'local' },
  RPCGetAttackMitigation: { path: '/attack/mitigation', target: 'local' },
  RPCGetAttackSoftware: { path: '/attack/software', target: 'local' },
  RPCGetAttackGroup: { path: '/attack/group', target: 'local' },
  RPCGetAttackTechniqueByID: { path: '/attack/technique-by-id', target: 'local' },
  RPCGetAttackTacticByID: { path: '/attack/tactic-by-id', target: 'local' },
  RPCGetAttackMitigationByID: { path: '/attack/mitigation-by-id', target: 'local' },
  RPCGetAttackSoftwareByID: { path: '/attack/software-by-id', target: 'local' },
  RPCGetAttackGroupByID: { path: '/attack/group-by-id', target: 'local' },
  RPCListAttackTechniques: { path: '/attack/techniques', target: 'local' },
  RPCListAttackTactics: { path: '/attack/tactics', target: 'local' },
  RPCListAttackMitigations: { path: '/attack/mitigations', target: 'local' },
  RPCListAttackSoftware: { path: '/attack/softwares', target: 'local' },
  RPCListAttackGroups: { path: '/attack/groups', target: 'local' },
  RPCImportATTACKs: { path: '/attack/import', target: 'local' },
  RPCGetAttackImportMetadata: { path: '/attack/import-metadata', target: 'local' },

  // ASVS
  RPCListASVS: { path: '/asvs/list', target: 'local' },
  RPCGetASVSByID: { path: '/asvs/get', target: 'local' },
  RPCImportASVS: { path: '/asvs/import', target: 'local' },

  // CCE
  RPCGetCCEByID: { path: '/cce/get', target: 'local' },
  RPCListCCEs: { path: '/cce/list', target: 'local' },
  RPCImportCCEs: { path: '/cce/import', target: 'local' },
  RPCImportCCE: { path: '/cce/import-one', target: 'local' },
  RPCCountCCEs: { path: '/cce/count', target: 'local' },
  RPCDeleteCCE: { path: '/cce/delete', target: 'local' },
  RPCUpdateCCE: { path: '/cce/update', target: 'local' },

  // Session/Job
  RPCStartSession: { path: '/session/start', target: 'meta' },
  RPCStartTypedSession: { path: '/session/start-typed', target: 'meta' },
  RPCStopSession: { path: '/session/stop', target: 'meta' },
  RPCGetSessionStatus: { path: '/session/status', target: 'meta' },
  RPCPauseJob: { path: '/job/pause', target: 'meta' },
  RPCResumeJob: { path: '/job/resume', target: 'meta' },

  // Bookmark
  RPCCreateBookmark: { path: '/bookmark/create', target: 'local' },
  RPCGetBookmark: { path: '/bookmark/get', target: 'local' },
  RPCUpdateBookmark: { path: '/bookmark/update', target: 'local' },
  RPCDeleteBookmark: { path: '/bookmark/delete', target: 'local' },
  RPCListBookmarks: { path: '/bookmark/list', target: 'local' },

  // Note
  RPCAddNote: { path: '/note/add', target: 'local' },
  RPCGetNote: { path: '/note/get', target: 'local' },
  RPCUpdateNote: { path: '/note/update', target: 'local' },
  RPCDeleteNote: { path: '/note/delete', target: 'local' },
  RPCGetNotesByBookmark: { path: '/note/by-bookmark', target: 'local' },

  // Memory Card
  RPCCreateMemoryCard: { path: '/memory-card/create', target: 'local' },
  RPCGetMemoryCard: { path: '/memory-card/get', target: 'local' },
  RPCUpdateMemoryCard: { path: '/memory-card/update', target: 'local' },
  RPCDeleteMemoryCard: { path: '/memory-card/delete', target: 'local' },
  RPCListMemoryCards: { path: '/memory-card/list', target: 'local' },
  RPCRateMemoryCard: { path: '/memory-card/rate', target: 'local' },

  // GLC
  RPCGLCGraphCreate: { path: '/glc/graph/create', target: 'local' },
  RPCGLCGraphGet: { path: '/glc/graph/get', target: 'local' },
  RPCGLCGraphUpdate: { path: '/glc/graph/update', target: 'local' },
  RPCGLCGraphDelete: { path: '/glc/graph/delete', target: 'local' },
  RPCGLCGraphList: { path: '/glc/graph/list', target: 'local' },
  RPCGLCGraphListRecent: { path: '/glc/graph/list-recent', target: 'local' },
  RPCGLCVersionGet: { path: '/glc/version/get', target: 'local' },
  RPCGLCVersionList: { path: '/glc/version/list', target: 'local' },
  RPCGLCVersionRestore: { path: '/glc/version/restore', target: 'local' },
  RPCGLCPresetCreate: { path: '/glc/preset/create', target: 'local' },
  RPCGLCPresetGet: { path: '/glc/preset/get', target: 'local' },
  RPCGLCPresetUpdate: { path: '/glc/preset/update', target: 'local' },
  RPCGLCPresetDelete: { path: '/glc/preset/delete', target: 'local' },
  RPCGLCPresetList: { path: '/glc/preset/list', target: 'local' },
  RPCGLCShareCreateLink: { path: '/glc/share/create', target: 'local' },
  RPCGLCShareGetShared: { path: '/glc/share/get', target: 'local' },
  RPCGLCShareGetEmbedData: { path: '/glc/share/embed', target: 'local' },

  // Analysis
  RPCGetGraphStats: { path: '/analysis/stats', target: 'analysis' },
  RPCAddNode: { path: '/analysis/node/add', target: 'analysis' },
  RPCAddEdge: { path: '/analysis/edge/add', target: 'analysis' },
  RPCGetNode: { path: '/analysis/node/get', target: 'analysis' },
  RPCGetNeighbors: { path: '/analysis/neighbors', target: 'analysis' },
  RPCFindPath: { path: '/analysis/path/find', target: 'analysis' },
  RPCGetNodesByType: { path: '/analysis/nodes/by-type', target: 'analysis' },
  RPCGetUEEStatus: { path: '/analysis/status', target: 'analysis' },
  RPCBuildCVEGraph: { path: '/analysis/graph/build', target: 'analysis' },
  RPCClearGraph: { path: '/analysis/graph/clear', target: 'analysis' },
  RPCGetFSMState: { path: '/analysis/fsm/state', target: 'analysis' },
  RPCPauseAnalysis: { path: '/analysis/fsm/pause', target: 'analysis' },
  RPCResumeAnalysis: { path: '/analysis/fsm/resume', target: 'analysis' },
  RPCSaveGraph: { path: '/analysis/graph/save', target: 'analysis' },
  RPCLoadGraph: { path: '/analysis/graph/load', target: 'analysis' },

  // System
  RPCGetSysMetrics: { path: '/system/metrics', target: 'sysmon' },

  // ETL
  RPCGetEtlTree: { path: '/etl/tree', target: 'meta' },
  RPCStartProvider: { path: '/etl/provider/start', target: 'meta' },
  RPCPauseProvider: { path: '/etl/provider/pause', target: 'meta' },
  RPCStopProvider: { path: '/etl/provider/stop', target: 'meta' },
  RPCUpdatePerformancePolicy: { path: '/etl/performance-policy', target: 'meta' },
  RPCGetKernelMetrics: { path: '/etl/kernel-metrics', target: 'meta' },

  // SSG
  RPCSSGImportGuide: { path: '/ssg/import-guide', target: 'local' },
  RPCSSGImportTable: { path: '/ssg/import-table', target: 'local' },
  RPCSSGGetGuide: { path: '/ssg/guide', target: 'local' },
  RPCSSGListGuides: { path: '/ssg/guides', target: 'local' },
  RPCSSGListTables: { path: '/ssg/tables', target: 'local' },
  RPCSSGGetTable: { path: '/ssg/table', target: 'local' },
  RPCSSGGetTableEntries: { path: '/ssg/table-entries', target: 'local' },
  RPCSSGGetTree: { path: '/ssg/tree', target: 'local' },
  RPCSSGGetTreeNode: { path: '/ssg/tree-node', target: 'local' },
  RPCSSGGetGroup: { path: '/ssg/group', target: 'local' },
  RPCSSGGetChildGroups: { path: '/ssg/child-groups', target: 'local' },
  RPCSSGGetRule: { path: '/ssg/rule', target: 'local' },
  RPCSSGListRules: { path: '/ssg/rules', target: 'local' },
  RPCSSGGetChildRules: { path: '/ssg/child-rules', target: 'local' },
  RPCSSGImportManifest: { path: '/ssg/import-manifest', target: 'local' },
  RPCSSGListManifests: { path: '/ssg/manifests', target: 'local' },
  RPCSSGGetManifest: { path: '/ssg/manifest', target: 'local' },
  RPCSSGListProfiles: { path: '/ssg/profiles', target: 'local' },
  RPCSSGGetProfile: { path: '/ssg/profile', target: 'local' },
  RPCSSGGetProfileRules: { path: '/ssg/profile-rules', target: 'local' },
  RPCSSGImportDataStream: { path: '/ssg/import-datastream', target: 'local' },
  RPCSSGListDataStreams: { path: '/ssg/datastreams', target: 'local' },
  RPCSSGGetDataStream: { path: '/ssg/datastream', target: 'local' },
  RPCSSGListDSProfiles: { path: '/ssg/ds-profiles', target: 'local' },
  RPCSSGGetDSProfile: { path: '/ssg/ds-profile', target: 'local' },
  RPCSSGGetDSProfileRules: { path: '/ssg/ds-profile-rules', target: 'local' },
  RPCSSGListDSGroups: { path: '/ssg/ds-groups', target: 'local' },
  RPCSSGListDSRules: { path: '/ssg/ds-rules', target: 'local' },
  RPCSSGGetDSRule: { path: '/ssg/ds-rule', target: 'local' },
  RPCSSGGetCrossReferences: { path: '/ssg/cross-references', target: 'local' },
  RPCSSGFindRelatedObjects: { path: '/ssg/find-related', target: 'local' },
  RPCSSGStartImportJob: { path: '/ssg/job/start', target: 'meta' },
  RPCSSGStopImportJob: { path: '/ssg/job/stop', target: 'meta' },
  RPCSSGPauseImportJob: { path: '/ssg/job/pause', target: 'meta' },
  RPCSSGResumeImportJob: { path: '/ssg/job/resume', target: 'meta' },
  RPCSSGGetImportStatus: { path: '/ssg/job/status', target: 'meta' },
};

function getPathForMethod(method: string): { path: string; target: string } | null {
  return methodToPathMap[method] || null;
}

// ============================================================================
// Default Configuration
// ============================================================================

// Detect if running in remote development environment
const isRemoteDev = typeof window !== 'undefined' && window.location.hostname !== 'localhost';

const DEFAULT_API_BASE_URL = isRemoteDev
  ? '/restful'  // Use relative path in remote dev (will be proxied)
  : process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';  // Local dev or fallback
const DEFAULT_TIMEOUT = 120000; // 120 seconds (2 minutes) - increased for SSG operations
const MOCK_DELAY_MS = 500; // 500ms delay for mock responses

// ============================================================================
// Mock Data for Development
// ============================================================================

const MOCK_CVE_DATA: CVEItem = {
  id: 'CVE-2021-44228',
  sourceIdentifier: 'cve@mitre.org',
  published: '2021-12-10T10:00:00.000',
  lastModified: '2024-11-21T09:23:00.000',
  vulnStatus: 'Modified',
  descriptions: [
    {
      lang: 'en',
      value: 'Apache Log4j2 2.0-beta9 through 2.15.0 (excluding security releases 2.12.2, 2.12.3, and 2.3.1) JNDI features used in configuration, log messages, and parameters do not protect against attacker controlled LDAP and other JNDI related endpoints. An attacker who can control log messages or log message parameters can execute arbitrary code loaded from LDAP servers when message lookup substitution is enabled.',
    },
  ],
  metrics: {
    cvssMetricV31: [
      {
        source: 'nvd@nist.gov',
        type: 'Primary',
        cvssData: {
          version: '3.1',
          vectorString: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H',
          baseScore: 10.0,
          baseSeverity: 'CRITICAL',
          attackVector: 'NETWORK',
          attackComplexity: 'LOW',
          privilegesRequired: 'NONE',
          userInteraction: 'NONE',
          scope: 'CHANGED',
          confidentialityImpact: 'HIGH',
          integrityImpact: 'HIGH',
          availabilityImpact: 'HIGH',
        },
        exploitabilityScore: 3.9,
        impactScore: 6.0,
      },
    ],
  },
  weaknesses: [
    {
      source: 'nvd@nist.gov',
      type: 'Primary',
      description: [
        {
          lang: 'en',
          value: 'CWE-502',
        },
      ],
    },
  ],
  references: [
    {
      url: 'https://logging.apache.org/log4j/2.x/security.html',
      source: 'cve@mitre.org',
      tags: ['Vendor Advisory'],
    },
  ],
};

const MOCK_ASVS_DATA: ASVSItem = {
  requirementID: '1.1.1',
  chapter: 'V1',
  section: 'Architecture, Design and Threat Modeling',
  description: 'Verify the use of a secure software development lifecycle that addresses security in all stages of development.',
  level1: true,
  level2: true,
  level3: true,
  cwe: 'CWE-1127',
};

// ============================================================================
// Case Conversion Utilities
// ============================================================================

// Pre-compiled regex patterns for case conversion (optimized for performance)
const SNAKE_CASE_REGEX = /_([a-zA-Z0-9])/g;
const CAMEL_CASE_REGEX = /[A-Z]/g;

/**
 * Convert PascalCase/snake_case to camelCase
 */
function toCamelCase(str: string): string {
  // Handle snake_case -> camelCase
  if (str.indexOf('_') >= 0) {
    return str.replace(SNAKE_CASE_REGEX, (_, letter) => letter.toUpperCase());
  }

  // If the key is ALL CAPS (e.g. "ID"), lower-case it entirely
  if (str === str.toUpperCase()) {
    return str.toLowerCase();
  }

  // PascalCase -> camelCase (lowercase first character)
  return str.charAt(0).toLowerCase() + str.slice(1);
}

/**
 * Recursively convert object keys from PascalCase/snake_case to camelCase
 */
function convertKeysToCamelCase<T>(obj: unknown): T {
  if (obj === null || obj === undefined) {
    return obj as T;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeysToCamelCase(item)) as T;
  }

  if (typeof obj === 'object') {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj)) {
      const camelKey = toCamelCase(key);
      result[camelKey] = convertKeysToCamelCase(value);
    }
    return result as T;
  }

  return obj as T;
}

/**
 * Convert camelCase to snake_case
 */
function toSnakeCase(str: string): string {
  return str.replace(CAMEL_CASE_REGEX, (letter) => `_${letter.toLowerCase()}`);
}

/**
 * Recursively convert object keys from camelCase to snake_case
 */
function convertKeysToSnakeCase<T>(obj: unknown): T {
  if (obj === null || obj === undefined) {
    return obj as T;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeysToSnakeCase(item)) as T;
  }

  if (typeof obj === 'object') {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj)) {
      const snakeKey = toSnakeCase(key);
      result[snakeKey] = convertKeysToSnakeCase(value);
    }
    return result as T;
  }

  return obj as T;
}

// ============================================================================
// RPC Client Class
// ============================================================================

// Track pending requests to prevent duplicates
const pendingRequests = new Map<string, Promise<RPCResponse<unknown>>>();

// Cleanup strategy for pending requests
// Prevents memory leaks from failed/aborted requests
const PENDING_REQUEST_CLEANUP_INTERVAL = 60000; // 1 minute
const MAX_PENDING_REQUEST_AGE = 300000; // 5 minutes - requests older than this are cleaned up

// Timestamps for tracking request age
const requestTimestamps = new Map<string, number>();

/**
 * Cleanup stale pending requests
 * Removes requests that have been pending longer than MAX_PENDING_REQUEST_AGE
 */
function cleanupStaleRequests(): void {
  const now = Date.now();
  const staleKeys: string[] = [];

  for (const [key, timestamp] of requestTimestamps.entries()) {
    if (now - timestamp > MAX_PENDING_REQUEST_AGE) {
      staleKeys.push(key);
    }
  }

  for (const key of staleKeys) {
    pendingRequests.delete(key);
    requestTimestamps.delete(key);
    logger.warn('Cleaned up stale pending request', { requestKey: key });
  }

  if (staleKeys.length > 0) {
    logger.debug('Cleaned up stale pending requests', { count: staleKeys.length });
  }
}

// Start periodic cleanup
if (typeof window !== 'undefined') {
  setInterval(cleanupStaleRequests, PENDING_REQUEST_CLEANUP_INTERVAL);
}

// Create a cache for RPC calls to deduplicate requests
async function cachedCall(
  baseUrl: string,
  method: string,
  params: any,
  target: string,
  timeout: number,
  useMock: boolean
): Promise<RPCResponse<unknown>> {
  // Create a unique key for this request
  const requestKey = `${method}:${JSON.stringify(params || {})}:${target}`;

  // Check if we already have a pending request for this key
  if (pendingRequests.has(requestKey)) {
    logger.debug('Deduplicating request', { requestKey });
    return pendingRequests.get(requestKey)!;
  }

  // Track request timestamp for cleanup
  requestTimestamps.set(requestKey, Date.now());

  if (useMock) {
    await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY_MS));
    const result = getMockResponseForCache(method, params);
    // Clean up both maps
    pendingRequests.delete(requestKey);
    requestTimestamps.delete(requestKey);
    return result;
  }

  const request: RPCRequest<unknown> = {
    method,
    params: params ? convertKeysToSnakeCase(params) : undefined,
    target,
  };

  // Check if we have a path-based route for this method
  const pathRoute = getPathForMethod(method);
  const usePathBased = pathRoute !== null;

  // Create promise for this request
  const requestPromise = (async () => {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      logger.debug('Making RPC request', { method, target, params, usePathBased });

      let url: string;
      let body: string;

      if (usePathBased && pathRoute) {
        // Use path-based endpoint
        const baseRpcPath = baseUrl.endsWith('/restful') ? '/rpc' : '/restful/rpc';
        url = `${baseUrl}${baseRpcPath}${pathRoute.path}`;
        // For path-based, params go directly in body
        body = JSON.stringify(params ? convertKeysToSnakeCase(params) : {});
      } else {
        // Use generic RPC endpoint
        const rpcPath = baseUrl.endsWith('/restful') ? '/rpc' : '/restful/rpc';
        url = `${baseUrl}${rpcPath}`;
        body = JSON.stringify(request);
      }

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      const raw = await response.text();

      if (!response.ok) {
        const rpcPath = baseUrl.endsWith('/restful') ? '/rpc' : '/restful/rpc';
        logger.error(`HTTP error: ${response.status} ${response.statusText}`, new Error(`HTTP ${response.status}`), {
          url,
          method,
          target,
          status: response.status,
          statusText: response.statusText,
          responseBody: raw.substring(0, 500),
        });
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      let rpcResponse: RPCResponse<unknown>;
      try {
        rpcResponse = JSON.parse(raw);
      } catch (err) {
        logger.error('Invalid JSON response from RPC endpoint', err, {
          responseBody: raw.substring(0, 500),
          method,
          target,
        });
        throw new Error('Invalid JSON response from RPC endpoint');
      }

      if (rpcResponse.payload) {
        rpcResponse.payload = convertKeysToCamelCase(rpcResponse.payload);
      }

      // Log failed RPC calls for debugging/bugfix purposes
      // Don't log certain expected "error" states as errors
      const acceptableErrors = [
        'no active import job',
        'not found',
        'does not exist',
      ];
      const isAcceptableError = rpcResponse.message && acceptableErrors.some(msg =>
        rpcResponse.message.toLowerCase().includes(msg)
      );

      if (rpcResponse.retcode !== 0) {
        if (isAcceptableError) {
          // Expected state, not an actual error - log at debug level
          logger.debug(`RPC call returned non-zero retcode=${rpcResponse.retcode}: ${rpcResponse.message}`, {
            request: { method, target },
          });
        } else {
          logger.warn(`RPC call failed with retcode=${rpcResponse.retcode}: ${rpcResponse.message}`, {
            request: { method, params, target },
            response: {
              retcode: rpcResponse.retcode,
              message: rpcResponse.message,
              payload: rpcResponse.payload,
            },
          });
        }
      }

      logger.debug('RPC request completed', { method, retcode: rpcResponse.retcode });
      return rpcResponse;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        logger.error('Request timeout', error, { method, target, timeout });
        return {
          retcode: 500,
          message: 'Request timeout',
          payload: null,
        };
      }
      logger.error('RPC request failed', error, { method, target });
      return {
        retcode: 500,
        message: error instanceof Error ? error.message : 'Unknown error',
        payload: null,
      };
    } finally {
      // Clean up pending request and timestamp
      pendingRequests.delete(requestKey);
      requestTimestamps.delete(requestKey);
    }
  })();
  
  // Store the promise
  pendingRequests.set(requestKey, requestPromise);
  return requestPromise;
}

function getMockResponseForCache<TResponse>(
  method: string,
  params?: unknown
): RPCResponse<TResponse> {
  switch (method) {
    case 'RPCGetCVE':
      return {
        retcode: 0,
        message: 'success',
        payload: {
          cve: MOCK_CVE_DATA,
          source: 'local',
        } as TResponse,
      };

    case 'RPCListCVEs':
      const listParams = params as ListCVEsRequest | undefined;
      return {
        retcode: 0,
        message: 'success',
        payload: {
          cves: [MOCK_CVE_DATA, { ...MOCK_CVE_DATA, id: 'CVE-2021-44229' }],
          total: 2,
          offset: listParams?.offset || 0,
          limit: listParams?.limit || 10,
        } as TResponse,
      };

    case 'RPCCountCVEs':
      return {
        retcode: 0,
        message: 'success',
        payload: {
          count: 150,
        } as TResponse,
      };

    case 'RPCGetSessionStatus':
      return {
        retcode: 0,
        message: 'success',
        payload: {
          hasSession: false,
        } as TResponse,
      };

    case 'RPCListCWEViews': {
      const lp = params as ListCWEViewsRequest | undefined;
      const sample: CWEView[] = [
        { id: 'V-1', name: 'View One', type: 'catalog', status: 'active', objective: 'Sample objective', audience: [], members: [], references: [], notes: [], contentHistory: [], raw: {} },
        { id: 'V-2', name: 'View Two', type: 'catalog', status: 'deprecated', objective: 'Second view', audience: [], members: [], references: [], notes: [], contentHistory: [], raw: {} },
      ];
      return {
        retcode: 0,
        message: 'success',
        payload: {
          views: sample.slice(0, lp?.limit || sample.length),
          offset: lp?.offset || 0,
          limit: lp?.limit || sample.length,
          total: sample.length,
        } as TResponse,
      };
    }

    case 'RPCGetCWEViewByID': {
      const req = params as { id?: string } | undefined;
      const id = req?.id || 'V-1';
      const view: CWEView = { id, name: `View ${id}`, type: 'catalog', status: 'active', objective: 'Mocked view detail', audience: [], members: [], references: [], notes: [], contentHistory: [], raw: {} };
      return {
        retcode: 0,
        message: 'success',
        payload: { view } as TResponse,
      };
    }

    case 'RPCListCAPECs': {
      const lp = params as { offset?: number; limit?: number } | undefined;
      const sample: CAPECItem[] = [
        { id: 'CAPEC-1', name: 'Example CAPEC One', summary: 'Example attack pattern one', description: 'Detailed description 1' },
        { id: 'CAPEC-2', name: 'Example CAPEC Two', summary: 'Example attack pattern two', description: 'Detailed description 2' },
      ];
      return {
        retcode: 0,
        message: 'success',
        payload: {
          capecs: sample.slice(0, lp?.limit || sample.length),
          offset: lp?.offset || 0,
          limit: lp?.limit || sample.length,
          total: sample.length,
        } as unknown as TResponse,
      };
    }

    case 'RPCGetCAPECByID': {
      const req = params as { capecId?: string } | undefined;
      const id = req?.capecId || 'CAPEC-1';
      const item = { id, name: `CAPEC ${id}`, summary: 'Mocked CAPEC summary', description: 'Mocked CAPEC details' };
      return {
        retcode: 0,
        message: 'success',
        payload: item as unknown as TResponse,
      };
    }

    case 'RPCStartCWEViewJob':
      return {
        retcode: 0,
        message: 'success',
        payload: { success: true, sessionId: `mock-session-${Date.now()}` } as TResponse,
      };

    case 'RPCStopCWEViewJob':
      return {
        retcode: 0,
        message: 'success',
        payload: { success: true, sessionId: undefined } as TResponse,
      };

    default:
      return {
        retcode: 0,
        message: 'success',
        payload: {} as TResponse,
      };
  }
}

export class RPCClient {
  private baseUrl: string;
  private timeout: number;
  private useMock: boolean;

  constructor(options?: {
    baseUrl?: string;
    timeout?: number;
    useMock?: boolean;
  }) {
    this.baseUrl = options?.baseUrl || DEFAULT_API_BASE_URL;
    this.timeout = options?.timeout || DEFAULT_TIMEOUT;
    this.useMock = options?.useMock || false;
  }

  /**
   * Make an RPC call to the backend
   */
  private async call<TRequest, TResponse>(
    method: string,
    params?: TRequest,
    target: string = 'meta'
  ): Promise<RPCResponse<TResponse>> {
    // Use the cached call function to deduplicate requests
    const result = await cachedCall(
      this.baseUrl,
      method,
      params ? convertKeysToSnakeCase(params) : undefined,
      target,
      this.timeout,
      this.useMock
    );
    
    // Log moved to cachedCall function for better timing
    
    return result as RPCResponse<TResponse>;
  }

  /**
   * Get mock response for development
   */
  private getMockResponse<TResponse>(
    method: string,
    params?: unknown
  ): RPCResponse<TResponse> {
    switch (method) {
      case 'RPCGetCVE':
        return {
          retcode: 0,
          message: 'success',
          payload: {
            cve: MOCK_CVE_DATA,
            source: 'local',
          } as TResponse,
        };

      case 'RPCListCVEs':
        const listParams = params as ListCVEsRequest | undefined;
        return {
          retcode: 0,
          message: 'success',
          payload: {
            cves: [MOCK_CVE_DATA, { ...MOCK_CVE_DATA, id: 'CVE-2021-44229' }],
            total: 2,
            offset: listParams?.offset || 0,
            limit: listParams?.limit || 10,
          } as TResponse,
        };

      case 'RPCCountCVEs':
        return {
          retcode: 0,
          message: 'success',
          payload: {
            count: 150,
          } as TResponse,
        };

      case 'RPCGetSessionStatus':
        return {
          retcode: 0,
          message: 'success',
          payload: {
            hasSession: false,
          } as TResponse,
        };

      case 'RPCListCWEViews': {
        const lp = params as ListCWEViewsRequest | undefined;
        const sample: CWEView[] = [
          { id: 'V-1', name: 'View One', type: 'catalog', status: 'active', objective: 'Sample objective', audience: [], members: [], references: [], notes: [], contentHistory: [], raw: {} },
          { id: 'V-2', name: 'View Two', type: 'catalog', status: 'deprecated', objective: 'Second view', audience: [], members: [], references: [], notes: [], contentHistory: [], raw: {} },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            views: sample.slice(0, lp?.limit || sample.length),
            offset: lp?.offset || 0,
            limit: lp?.limit || sample.length,
            total: sample.length,
          } as TResponse,
        };
      }

      case 'RPCGetCWEViewByID': {
        const req = params as { id?: string } | undefined;
        const id = req?.id || 'V-1';
        const view: CWEView = { id, name: `View ${id}`, type: 'catalog', status: 'active', objective: 'Mocked view detail', audience: [], members: [], references: [], notes: [], contentHistory: [], raw: {} };
        return {
          retcode: 0,
          message: 'success',
          payload: { view } as TResponse,
        };
      }

      case 'RPCListCAPECs': {
        const lp = params as { offset?: number; limit?: number } | undefined;
        const sample: CAPECItem[] = [
          { id: 'CAPEC-1', name: 'Example CAPEC One', summary: 'Example attack pattern one', description: 'Detailed description 1' },
          { id: 'CAPEC-2', name: 'Example CAPEC Two', summary: 'Example attack pattern two', description: 'Detailed description 2' },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            capecs: sample.slice(0, lp?.limit || sample.length),
            offset: lp?.offset || 0,
            limit: lp?.limit || sample.length,
            total: sample.length,
          } as unknown as TResponse,
        };
      }

      case 'RPCGetCAPECByID': {
        const req = params as { capecId?: string } | undefined;
        const id = req?.capecId || 'CAPEC-1';
        const item = { id, name: `CAPEC ${id}`, summary: 'Mocked CAPEC summary', description: 'Mocked CAPEC details' };
        return {
          retcode: 0,
          message: 'success',
          payload: item as unknown as TResponse,
        };
      }

      case 'RPCListASVS': {
        const lp = params as ListASVSRequest | undefined;
        const sample: ASVSItem[] = [
          { ...MOCK_ASVS_DATA },
          { ...MOCK_ASVS_DATA, requirementID: '1.1.2', description: 'Verify the use of threat modeling for every design change or sprint planning to identify threats, plan for countermeasures, facilitate appropriate risk responses, and guide security testing.', level1: false },
          { ...MOCK_ASVS_DATA, requirementID: '2.1.1', chapter: 'V2', section: 'Authentication', description: 'Verify that user set passwords are at least 12 characters in length.', cwe: 'CWE-521' },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            requirements: sample.slice(0, lp?.limit || sample.length),
            offset: lp?.offset || 0,
            limit: lp?.limit || sample.length,
            total: sample.length,
          } as unknown as TResponse,
        };
      }

      case 'RPCGetASVSByID': {
        const req = params as GetASVSByIDRequest | undefined;
        const requirementId = req?.requirementId || '1.1.1';
        const item: ASVSItem = { ...MOCK_ASVS_DATA, requirementID: requirementId };
        return {
          retcode: 0,
          message: 'success',
          payload: item as unknown as TResponse,
        };
      }

      case 'RPCImportASVS': {
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true } as unknown as TResponse,
        };
      }

      case 'RPCStartCWEViewJob':
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, sessionId: `mock-session-${Date.now()}` } as TResponse,
        };

      case 'RPCStopCWEViewJob':
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, sessionId: undefined } as TResponse,
        };

      // Notes Framework Mock Responses
      case 'RPCCreateBookmark': {
        const req = params as CreateBookmarkRequest | undefined;
        const mockBookmark: Bookmark = {
          urn: `v2e::${req?.item_type?.toLowerCase() || 'cve'}::${req?.item_id || 'CVE-2021-44228'}`,
          id: Math.floor(Math.random() * 1000),
          global_item_id: req?.global_item_id || 'CVE-2021-44228',
          item_type: req?.item_type || 'CVE',
          item_id: req?.item_id || 'CVE-2021-44228',
          title: req?.title || 'Mock Bookmark Title',
          description: req?.description || 'Mock bookmark description',
          author: req?.author,
          is_private: req?.is_private || false,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: req?.metadata || {},
        };
        const mockMemoryCard: MemoryCard = {
          urn: mockBookmark.urn,
          id: Math.floor(Math.random() * 10000),
          bookmark_id: mockBookmark.id,
          front_content: mockBookmark.title,
          back_content: mockBookmark.description,
          front: mockBookmark.title,
          back: mockBookmark.description,
          major_class: '',
          minor_class: '',
          status: 'new',
          content: '{}',
          card_type: 'basic',
          learning_state: 'to_review',
          author: 'test-user',
          is_private: false,
          interval: 1,
          ease_factor: 2.5,
          repetitions: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          next_review_at: new Date().toISOString(),
          metadata: {},
        };
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, bookmark: mockBookmark, memoryCard: mockMemoryCard } as unknown as TResponse,
        };
      }

      case 'RPCGetBookmark': {
        const req = params as GetBookmarkRequest | undefined;
        const mockBookmark: Bookmark = {
          urn: 'v2e::cve::CVE-2021-44228',
          id: req?.id || 1,
          global_item_id: 'CVE-2021-44228',
          item_type: 'CVE',
          item_id: 'CVE-2021-44228',
          title: 'Mock Bookmark Title',
          description: 'Mock bookmark description',
          author: 'mock-user',
          is_private: false,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: {},
        };
        return {
          retcode: 0,
          message: 'success',
          payload: { bookmark: mockBookmark } as unknown as TResponse,
        };
      }

      case 'RPCListBookmarks': {
        const req = params as ListBookmarksRequest | undefined;
        const mockBookmarks: Bookmark[] = [
          {
            urn: 'v2e::cve::CVE-2021-44228',
            id: 1,
            global_item_id: 'CVE-2021-44228',
            item_type: 'CVE',
            item_id: 'CVE-2021-44228',
            title: 'Log4Shell Vulnerability',
            description: 'Critical vulnerability in Log4j',
            author: 'test-user',
            is_private: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            metadata: {},
          },
          {
            id: 2,
            global_item_id: 'CVE-2020-1472',
            item_type: 'CVE',
            urn: 'v2e::cve::CVE-2020-1472',
            item_id: 'CVE-2020-1472',
            title: 'Zerologon Vulnerability',
            description: 'Privilege escalation in Windows Netlogon',
            author: 'test-user',
            is_private: true,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            metadata: {},
          },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            bookmarks: mockBookmarks.slice(0, req?.limit || mockBookmarks.length),
            offset: req?.offset || 0,
            limit: req?.limit || mockBookmarks.length,
            total: mockBookmarks.length,
          } as unknown as TResponse,
        };
      }

      case 'RPCAddNote': {
        const req = params as AddNoteRequest | undefined;
        const mockNote: Note = {
          id: Math.floor(Math.random() * 1000),
          bookmark_id: req?.bookmark_id || 1,
          content: req?.content || 'Mock note content',
          author: req?.author,
          is_private: req?.is_private || false,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: req?.metadata || {},
        };
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, note: mockNote } as unknown as TResponse,
        };
      }

      case 'RPCGetNote': {
        const req = params as GetNoteRequest | undefined;
        const mockNote: Note = {
          bookmark_id: 1,
          id: req?.id || 1,
          content: 'Mock note content',
          author: 'test-user',
          is_private: false,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: {},
        };
        return {
          retcode: 0,
          message: 'success',
          payload: { note: mockNote } as unknown as TResponse,
        };
      }

      case 'RPCGetNotesByBookmark': {
        const req = params as GetNotesByBookmarkRequest | undefined;
        const mockNotes: Note[] = [
          {
            id: 1,
            bookmark_id: req?.bookmark_id || 1,
            content: 'First note about this vulnerability',
            author: 'test-user',
            is_private: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            metadata: {},
          },
          {
            id: 2,
            bookmark_id: req?.bookmark_id || 1,
            content: 'Additional details and mitigation steps',
            author: 'test-user',
            is_private: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            metadata: {},
          },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            notes: mockNotes.slice(0, req?.limit || mockNotes.length),
            offset: req?.offset || 0,
            limit: req?.limit || mockNotes.length,
            total: mockNotes.length,
          } as unknown as TResponse,
        };
      }

      case 'RPCCreateMemoryCard': {
        const req = params as CreateMemoryCardRequest | undefined;
        const mockCard: MemoryCard = {
          urn: `v2e::card::${req?.bookmark_id || 1}::${Math.floor(Math.random() * 1000)}`,
          id: Math.floor(Math.random() * 1000),
          bookmark_id: req?.bookmark_id || 1,
          front_content: req?.front || 'What is Log4Shell?',
          back_content: req?.back || 'A critical vulnerability in Log4j allowing RCE',
          front: req?.front || 'What is Log4Shell?',
          back: req?.back || 'A critical vulnerability in Log4j allowing RCE',
          major_class: '',
          minor_class: '',
          status: 'new',
          content: '{}',
          card_type: req?.card_type || 'basic',
          learning_state: 'to_review',
          author: req?.author || 'test-user',
          is_private: req?.is_private || false,
          interval: 1,
          ease_factor: 2.5,
          repetitions: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          next_review_at: new Date(Date.now() + 86400000).toISOString(), // Tomorrow
          metadata: req?.metadata || {},
        };
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, memory_card: mockCard } as unknown as TResponse,
        };
      }

      case 'RPCListMemoryCards': {
        const req = params as ListMemoryCardsRequest | undefined;
        const mockCards: MemoryCard[] = [
          {
            urn: 'v2e::card::1::1',
            id: 1,
            bookmark_id: 1,
            front_content: 'What is Log4Shell?',
            back_content: 'A critical vulnerability in Log4j allowing RCE',
            front: 'What is Log4Shell?',
            back: 'A critical vulnerability in Log4j allowing RCE',
            major_class: '',
            minor_class: '',
            status: 'new',
            content: '{}',
            card_type: 'basic',
            learning_state: req?.learning_state || 'to_review',
            author: 'test-user',
            is_private: false,
            interval: 1,
            ease_factor: 2.5,
            repetitions: 0,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            next_review_at: new Date(Date.now() + 86400000).toISOString(), // Tomorrow
            metadata: {},
          },
          {
            id: 2,
            bookmark_id: 1,
            urn: 'v2e::card::1::2',
            front_content: 'How to mitigate Log4Shell?',
            back_content: 'Upgrade to Log4j 2.15.0 or apply JVM parameters',
            front: 'How to mitigate Log4Shell?',
            back: 'Upgrade to Log4j 2.15.0 or apply JVM parameters',
            major_class: '',
            minor_class: '',
            status: 'learning',
            content: '{}',
            card_type: 'basic',
            learning_state: 'learning',
            author: 'test-user',
            is_private: false,
            interval: 3,
            ease_factor: 2.0,
            repetitions: 2,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            next_review_at: new Date(Date.now() + 86400000 * 3).toISOString(), // In 3 days
            metadata: {},
          },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            memory_cards: mockCards.slice(0, req?.limit || mockCards.length),
            offset: req?.offset || 0,
            limit: req?.limit || mockCards.length,
            total: mockCards.length,
          } as unknown as TResponse,
        };
      }

      case 'RPCCreateCrossReference': {
        const req = params as CreateCrossReferenceRequest | undefined;
        const mockRef: CrossReference = {
          id: Math.floor(Math.random() * 1000),
          from_item_id: req?.from_item_id || 'CVE-2021-44228',
          from_item_type: req?.from_item_type || 'CVE',
          to_item_id: req?.to_item_id || 'CWE-502',
          to_item_type: req?.to_item_type || 'CWE',
          relationship_type: req?.relationship_type || 'related_to',
          description: req?.description || 'Cross-reference description',
          strength: req?.strength || 5,
          author: req?.author,
          is_private: req?.is_private || false,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: req?.metadata || {},
        };
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, cross_reference: mockRef } as unknown as TResponse,
        };
      }

      case 'RPCGetHistory': {
        const req = params as GetHistoryRequest | undefined;
        const mockHistory: HistoryEntry[] = [
          {
            id: 1,
            item_id: req?.item_id || 'CVE-2021-44228',
            item_type: req?.item_type || 'CVE',
            action: 'bookmarked',
            author: 'test-user',
            timestamp: new Date().toISOString(),
            metadata: {},
          },
          {
            id: 2,
            item_id: req?.item_id || 'CVE-2021-44228',
            item_type: req?.item_type || 'CVE',
            action: 'note_added',
            author: 'test-user',
            timestamp: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
            metadata: {},
          },
        ];
        return {
          retcode: 0,
          message: 'success',
          payload: {
            history_entries: mockHistory.slice(0, req?.limit || mockHistory.length),
            offset: req?.offset || 0,
            limit: req?.limit || mockHistory.length,
            total: mockHistory.length,
          } as unknown as TResponse,
        };
      }

      default:
        return {
          retcode: 0,
          message: 'success',
          payload: {} as TResponse,
        };
    }
  }

  // ============================================================================
  // CVE Data Methods
  // ============================================================================

  async getCVE(cveId: string): Promise<RPCResponse<GetCVEResponse>> {
    return this.call<GetCVERequest, GetCVEResponse>('RPCGetCVE', {
      cveId: cveId,
    });
  }

  async createCVE(cveId: string): Promise<RPCResponse<CreateCVEResponse>> {
    return this.call<CreateCVERequest, CreateCVEResponse>('RPCCreateCVE', {
      cveId: cveId,
    });
  }

  async updateCVE(cveId: string): Promise<RPCResponse<UpdateCVEResponse>> {
    return this.call<UpdateCVERequest, UpdateCVEResponse>('RPCUpdateCVE', {
      cveId: cveId,
    });
  }

  async deleteCVE(cveId: string): Promise<RPCResponse<DeleteCVEResponse>> {
    return this.call<DeleteCVERequest, DeleteCVEResponse>('RPCDeleteCVE', {
      cveId: cveId,
    });
  }

  async listCVEs(
    offset?: number,
    limit?: number
  ): Promise<RPCResponse<ListCVEsResponse>> {
    return this.call<ListCVEsRequest, ListCVEsResponse>('RPCListCVEs', {
      offset,
      limit,
    });
  }

  async countCVEs(): Promise<RPCResponse<CountCVEsResponse>> {
    return this.call<undefined, CountCVEsResponse>('RPCCountCVEs');
  }

  // ============================================================================
  // Job Session Methods
  // ============================================================================

  async startSession(
    sessionId: string,
    startIndex?: number,
    resultsPerBatch?: number
  ): Promise<RPCResponse<StartSessionResponse>> {
    return this.call<StartSessionRequest, StartSessionResponse>(
      'RPCStartSession',
      {
        sessionId: sessionId,
        startIndex: startIndex,
        resultsPerBatch: resultsPerBatch,
      }
    );
  }

  async startTypedSession(
    sessionId: string,
    dataType: string,
    startIndex?: number,
    resultsPerBatch?: number,
    params?: Record<string, unknown>
  ): Promise<RPCResponse<StartSessionResponse>> {
    return this.call<any, StartSessionResponse>(
      'RPCStartTypedSession',
      {
        session_id: sessionId,
        data_type: dataType,
        start_index: startIndex ?? 0,
        results_per_batch: resultsPerBatch ?? 100,
        params: params,
      }
    );
  }

  async stopSession(): Promise<RPCResponse<StopSessionResponse>> {
    return this.call<undefined, StopSessionResponse>('RPCStopSession');
  }

  async getSessionStatus(): Promise<RPCResponse<SessionStatus>> {
    return this.call<undefined, SessionStatus>('RPCGetSessionStatus');
  }

  async pauseJob(): Promise<RPCResponse<PauseJobResponse>> {
    return this.call<undefined, PauseJobResponse>('RPCPauseJob');
  }

  async resumeJob(): Promise<RPCResponse<ResumeJobResponse>> {
    return this.call<undefined, ResumeJobResponse>('RPCResumeJob');
  }

  // ==========================================================================
  // CWE View Job Methods
  // ==========================================================================

  async startCWEViewJob(
    sessionId?: string,
    startIndex?: number,
    resultsPerBatch?: number
  ): Promise<RPCResponse<StartCWEViewJobResponse>> {
    return this.call<StartCWEViewJobRequest, StartCWEViewJobResponse>(
      'RPCStartCWEViewJob',
      {
        sessionId: sessionId,
        startIndex: startIndex,
        resultsPerBatch: resultsPerBatch,
      }
    );
  }

  async stopCWEViewJob(sessionId?: string): Promise<RPCResponse<StopCWEViewJobResponse>> {
    return this.call<{ sessionId?: string }, StopCWEViewJobResponse>('RPCStopCWEViewJob', { sessionId });
  }

  // ==========================================================================
  // CWE View Data Methods
  // ==========================================================================

  async listCWEViews(offset?: number, limit?: number): Promise<RPCResponse<ListCWEViewsResponse>> {
    return this.call<ListCWEViewsRequest, ListCWEViewsResponse>('RPCListCWEViews', { offset: offset || 0, limit: limit || 100 }, 'local');
  }

  async getCWEViewByID(id: string): Promise<RPCResponse<GetCWEViewResponse>> {
    return this.call<{ id: string }, GetCWEViewResponse>('RPCGetCWEViewByID', { id }, 'local');
  }

  // ============================================================================
  // Health Check
  // ============================================================================

  async health(): Promise<RPCResponse<HealthResponse>> {
    if (this.useMock) {
      await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY_MS));
      return {
        retcode: 0,
        message: 'success',
        payload: { status: 'ok' },
      };
    }

    try {
      const response = await fetch(`${this.baseUrl}/restful/health`);
      const data = await response.json();
      return {
        retcode: 0,
        message: 'success',
        payload: data,
      };
    } catch (error) {
      return {
        retcode: 500,
        message: error instanceof Error ? error.message : 'Unknown error',
        payload: null,
      };
    }
  }

  // ==========================================================================
  // CWE Data Methods
  // ==========================================================================

  async listCWEs(params?: ListCWEsRequest): Promise<RPCResponse<ListCWEsResponse>> {
    return this.call<ListCWEsRequest, ListCWEsResponse>(
      'RPCListCWEs',
      params,
      'local'
    );
  }

  async listCAPECs(offset?: number, limit?: number): Promise<RPCResponse<{ capecs: CAPECItem[]; offset: number; limit: number; total: number }>> {
    return this.call<{ offset?: number; limit?: number }, { capecs: CAPECItem[]; offset: number; limit: number; total: number }>(
      'RPCListCAPECs',
      { offset: offset || 0, limit: limit || 50 },
      'local'
    );
  }

  async getCAPEC(capecId: string): Promise<RPCResponse<any>> {
    return this.call<{ capecId: string }, any>('RPCGetCAPECByID', { capecId }, 'local');
  }

  async getCWE(cweId: string): Promise<RPCResponse<{ cwe: CWEItem }>> {
    return this.call<{ cweId: string }, { cwe: CWEItem }>('RPCGetCWEByID', { cweId }, 'local');
  }

  // ASVS Methods
  async listASVS(params?: ListASVSRequest): Promise<RPCResponse<ListASVSResponse>> {
    return this.call<ListASVSRequest, ListASVSResponse>(
      'RPCListASVS',
      params,
      'local'
    );
  }

  async getASVS(requirementId: string): Promise<RPCResponse<ASVSItem>> {
    return this.call<GetASVSByIDRequest, ASVSItem>(
      'RPCGetASVSByID',
      { requirementId },
      'local'
    );
  }

  async importASVS(url: string): Promise<RPCResponse<ImportASVSResponse>> {
    return this.call<ImportASVSRequest, ImportASVSResponse>(
      'RPCImportASVS',
      { url },
      'local'
    );
  }

  // ATT&CK Methods
  async importATTACK(filePath: string, force: boolean = false): Promise<RPCResponse<any>> {
    return this.call<{ path: string, force: boolean }, any>('RPCImportATTACKs', { path: filePath, force }, 'local');
  }

  async getAttackTechnique(id: string): Promise<RPCResponse<any>> {
    return this.call<{ id: string }, any>('RPCGetAttackTechniqueByID', { id }, 'local');
  }

  async getAttackTactic(id: string): Promise<RPCResponse<any>> {
    return this.call<{ id: string }, any>('RPCGetAttackTacticByID', { id }, 'local');
  }

  async getAttackMitigation(id: string): Promise<RPCResponse<any>> {
    return this.call<{ id: string }, any>('RPCGetAttackMitigationByID', { id }, 'local');
  }

  async getAttackSoftware(id: string): Promise<RPCResponse<any>> {
    return this.call<{ id: string }, any>('RPCGetAttackSoftwareByID', { id }, 'local');
  }

  async getAttackGroup(id: string): Promise<RPCResponse<any>> {
    return this.call<{ id: string }, any>('RPCGetAttackGroupByID', { id }, 'local');
  }

  async listAttackTechniques(offset: number = 0, limit: number = 100): Promise<RPCResponse<any>> {
    return this.call<{ offset: number, limit: number }, any>('RPCListAttackTechniques', { offset, limit }, 'local');
  }

  async listAttackTactics(offset: number = 0, limit: number = 100): Promise<RPCResponse<any>> {
    return this.call<{ offset: number, limit: number }, any>('RPCListAttackTactics', { offset, limit }, 'local');
  }

  async listAttackMitigations(offset: number = 0, limit: number = 100): Promise<RPCResponse<any>> {
    return this.call<{ offset: number, limit: number }, any>('RPCListAttackMitigations', { offset, limit }, 'local');
  }

  async listAttackSoftware(offset: number = 0, limit: number = 100): Promise<RPCResponse<any>> {
    return this.call<{ offset: number, limit: number }, any>('RPCListAttackSoftware', { offset, limit }, 'local');
  }

  async listAttackGroups(offset: number = 0, limit: number = 100): Promise<RPCResponse<any>> {
    return this.call<{ offset: number, limit: number }, any>('RPCListAttackGroups', { offset, limit }, 'local');
  }

  async getAttackImportMetadata(): Promise<RPCResponse<any>> {
    return this.call<undefined, any>('RPCGetAttackImportMetadata', undefined, 'local');
  }

  // ==========================================================================
  // System Metrics
  // ==========================================================================

  async getSysMetrics(): Promise<RPCResponse<SysMetrics>> {
    return this.call<undefined, SysMetrics>("RPCGetSysMetrics", undefined, "sysmon");
  }

  // ==========================================================================
  // Notes Framework Methods
  // ==========================================================================

  // Bookmark Methods
  async createBookmark(params: CreateBookmarkRequest): Promise<RPCResponse<CreateBookmarkResponse>> {
    return this.call<CreateBookmarkRequest, CreateBookmarkResponse>('RPCCreateBookmark', params, 'local');
  }

  async getBookmark(params: GetBookmarkRequest): Promise<RPCResponse<GetBookmarkResponse>> {
    return this.call<GetBookmarkRequest, GetBookmarkResponse>('RPCGetBookmark', params, 'local');
  }

  async listBookmarks(params?: ListBookmarksRequest): Promise<RPCResponse<ListBookmarksResponse>> {
    return this.call<ListBookmarksRequest, ListBookmarksResponse>('RPCListBookmarks', params || {}, 'local');
  }

  async updateBookmark(params: UpdateBookmarkRequest): Promise<RPCResponse<UpdateBookmarkResponse>> {
    return this.call<UpdateBookmarkRequest, UpdateBookmarkResponse>('RPCUpdateBookmark', params, 'local');
  }

  async deleteBookmark(params: DeleteBookmarkRequest): Promise<RPCResponse<DeleteBookmarkResponse>> {
    return this.call<DeleteBookmarkRequest, DeleteBookmarkResponse>('RPCDeleteBookmark', params, 'local');
  }

  // Note Methods
  async addNote(params: AddNoteRequest): Promise<RPCResponse<AddNoteResponse>> {
    return this.call<AddNoteRequest, AddNoteResponse>('RPCAddNote', params, 'local');
  }

  async getNote(params: GetNoteRequest): Promise<RPCResponse<GetNoteResponse>> {
    return this.call<GetNoteRequest, GetNoteResponse>('RPCGetNote', params, 'local');
  }

  async getNotesByBookmark(params: GetNotesByBookmarkRequest): Promise<RPCResponse<GetNotesByBookmarkResponse>> {
    return this.call<GetNotesByBookmarkRequest, GetNotesByBookmarkResponse>('RPCGetNotesByBookmark', params, 'local');
  }

  async updateNote(params: UpdateNoteRequest): Promise<RPCResponse<UpdateNoteResponse>> {
    return this.call<UpdateNoteRequest, UpdateNoteResponse>('RPCUpdateNote', params, 'local');
  }

  async deleteNote(params: DeleteNoteRequest): Promise<RPCResponse<DeleteNoteResponse>> {
    return this.call<DeleteNoteRequest, DeleteNoteResponse>('RPCDeleteNote', params, 'local');
  }

  // Memory Card Methods
  async createMemoryCard(params: CreateMemoryCardRequest): Promise<RPCResponse<CreateMemoryCardResponse>> {
    return this.call<CreateMemoryCardRequest, CreateMemoryCardResponse>('RPCCreateMemoryCard', params, 'local');
  }

  async getMemoryCard(params: GetMemoryCardRequest): Promise<RPCResponse<GetMemoryCardResponse>> {
    return this.call<GetMemoryCardRequest, GetMemoryCardResponse>('RPCGetMemoryCard', params, 'local');
  }

  async listMemoryCards(params?: ListMemoryCardsRequest): Promise<RPCResponse<ListMemoryCardsResponse>> {
    return this.call<ListMemoryCardsRequest, ListMemoryCardsResponse>('RPCListMemoryCards', params || {}, 'local');
  }

  async updateMemoryCard(params: UpdateMemoryCardRequest): Promise<RPCResponse<UpdateMemoryCardResponse>> {
    return this.call<UpdateMemoryCardRequest, UpdateMemoryCardResponse>('RPCUpdateMemoryCard', params, 'local');
  }

  async deleteMemoryCard(params: DeleteMemoryCardRequest): Promise<RPCResponse<DeleteMemoryCardResponse>> {
    return this.call<DeleteMemoryCardRequest, DeleteMemoryCardResponse>('RPCDeleteMemoryCard', params, 'local');
  }

  async rateMemoryCard(params: RateMemoryCardRequest): Promise<RPCResponse<RateMemoryCardResponse>> {
    return this.call<RateMemoryCardRequest, RateMemoryCardResponse>('RPCRateMemoryCard', params, 'local');
  }

  // ============================================================================
  // GLC (Graphized Learning Canvas) Methods
  // ============================================================================

  // Graph Methods
  async createGLCGraph(params: CreateGLCGraphRequest): Promise<RPCResponse<CreateGLCGraphResponse>> {
    return this.call<CreateGLCGraphRequest, CreateGLCGraphResponse>('RPCGLCGraphCreate', params, 'local');
  }

  async getGLCGraph(params: GetGLCGraphRequest): Promise<RPCResponse<GetGLCGraphResponse>> {
    return this.call<GetGLCGraphRequest, GetGLCGraphResponse>('RPCGLCGraphGet', params, 'local');
  }

  async updateGLCGraph(params: UpdateGLCGraphRequest): Promise<RPCResponse<UpdateGLCGraphResponse>> {
    return this.call<UpdateGLCGraphRequest, UpdateGLCGraphResponse>('RPCGLCGraphUpdate', params, 'local');
  }

  async deleteGLCGraph(params: DeleteGLCGraphRequest): Promise<RPCResponse<DeleteGLCGraphResponse>> {
    return this.call<DeleteGLCGraphRequest, DeleteGLCGraphResponse>('RPCGLCGraphDelete', params, 'local');
  }

  async listGLCGraphs(params?: ListGLCGraphsRequest): Promise<RPCResponse<ListGLCGraphsResponse>> {
    return this.call<ListGLCGraphsRequest, ListGLCGraphsResponse>('RPCGLCGraphList', params || {}, 'local');
  }

  async listRecentGLCGraphs(params?: ListRecentGLCGraphsRequest): Promise<RPCResponse<ListRecentGLCGraphsResponse>> {
    return this.call<ListRecentGLCGraphsRequest, ListRecentGLCGraphsResponse>('RPCGLCGraphListRecent', params || {}, 'local');
  }

  // Version Methods
  async getGLCVersion(params: GetGLCVersionRequest): Promise<RPCResponse<GetGLCVersionResponse>> {
    return this.call<GetGLCVersionRequest, GetGLCVersionResponse>('RPCGLCVersionGet', params, 'local');
  }

  async listGLCVersions(params: ListGLCVersionsRequest): Promise<RPCResponse<ListGLCVersionsResponse>> {
    return this.call<ListGLCVersionsRequest, ListGLCVersionsResponse>('RPCGLCVersionList', params, 'local');
  }

  async restoreGLCVersion(params: RestoreGLCVersionRequest): Promise<RPCResponse<RestoreGLCVersionResponse>> {
    return this.call<RestoreGLCVersionRequest, RestoreGLCVersionResponse>('RPCGLCVersionRestore', params, 'local');
  }

  // Preset Methods
  async createGLCPreset(params: CreateGLCPresetRequest): Promise<RPCResponse<CreateGLCPresetResponse>> {
    return this.call<CreateGLCPresetRequest, CreateGLCPresetResponse>('RPCGLCPresetCreate', params, 'local');
  }

  async getGLCPreset(params: GetGLCPresetRequest): Promise<RPCResponse<GetGLCPresetResponse>> {
    return this.call<GetGLCPresetRequest, GetGLCPresetResponse>('RPCGLCPresetGet', params, 'local');
  }

  async updateGLCPreset(params: UpdateGLCPresetRequest): Promise<RPCResponse<UpdateGLCPresetResponse>> {
    return this.call<UpdateGLCPresetRequest, UpdateGLCPresetResponse>('RPCGLCPresetUpdate', params, 'local');
  }

  async deleteGLCPreset(params: DeleteGLCPresetRequest): Promise<RPCResponse<DeleteGLCPresetResponse>> {
    return this.call<DeleteGLCPresetRequest, DeleteGLCPresetResponse>('RPCGLCPresetDelete', params, 'local');
  }

  async listGLCPresets(): Promise<RPCResponse<ListGLCPresetsResponse>> {
    return this.call<Record<string, never>, ListGLCPresetsResponse>('RPCGLCPresetList', {}, 'local');
  }

  // Share Link Methods
  async createGLCShareLink(params: CreateGLCShareLinkRequest): Promise<RPCResponse<CreateGLCShareLinkResponse>> {
    return this.call<CreateGLCShareLinkRequest, CreateGLCShareLinkResponse>('RPCGLCShareCreateLink', params, 'local');
  }

  async getGLCSharedGraph(params: GetGLCSharedGraphRequest): Promise<RPCResponse<GetGLCSharedGraphResponse>> {
    return this.call<GetGLCSharedGraphRequest, GetGLCSharedGraphResponse>('RPCGLCShareGetShared', params, 'local');
  }

  async getGLCShareEmbedData(params: GetGLCShareEmbedDataRequest): Promise<RPCResponse<GetGLCShareEmbedDataResponse>> {
    return this.call<GetGLCShareEmbedDataRequest, GetGLCShareEmbedDataResponse>('RPCGLCShareGetEmbedData', params, 'local');
  }

  // Cross Reference Methods
  async createCrossReference(params: CreateCrossReferenceRequest): Promise<RPCResponse<CreateCrossReferenceResponse>> {
    return this.call<CreateCrossReferenceRequest, CreateCrossReferenceResponse>('RPCCreateCrossReference', params, 'local');
  }

  async getCrossReference(params: GetCrossReferenceRequest): Promise<RPCResponse<GetCrossReferenceResponse>> {
    return this.call<GetCrossReferenceRequest, GetCrossReferenceResponse>('RPCGetCrossReference', params, 'local');
  }

  async listCrossReferences(params?: ListCrossReferencesRequest): Promise<RPCResponse<ListCrossReferencesResponse>> {
    return this.call<ListCrossReferencesRequest, ListCrossReferencesResponse>('RPCListCrossReferences', params || {}, 'local');
  }

  async updateCrossReference(params: UpdateCrossReferenceRequest): Promise<RPCResponse<UpdateCrossReferenceResponse>> {
    return this.call<UpdateCrossReferenceRequest, UpdateCrossReferenceResponse>('RPCUpdateCrossReference', params, 'local');
  }

  async deleteCrossReference(params: DeleteCrossReferenceRequest): Promise<RPCResponse<DeleteCrossReferenceResponse>> {
    return this.call<DeleteCrossReferenceRequest, DeleteCrossReferenceResponse>('RPCDeleteCrossReference', params, 'local');
  }

  // History Methods
  async addHistory(params: AddHistoryRequest): Promise<RPCResponse<AddHistoryResponse>> {
    return this.call<AddHistoryRequest, AddHistoryResponse>('RPCAddHistory', params, 'local');
  }

  async getHistory(params: GetHistoryRequest): Promise<RPCResponse<GetHistoryResponse>> {
    return this.call<GetHistoryRequest, GetHistoryResponse>('RPCGetHistory', params, 'local');
  }

  async getHistoryByAction(params: GetHistoryByActionRequest): Promise<RPCResponse<GetHistoryByActionResponse>> {
    return this.call<GetHistoryByActionRequest, GetHistoryByActionResponse>('RPCGetHistoryByAction', params, 'local');
  }

  // Bookmark State Reversion
  async revertBookmarkState(params: RevertBookmarkStateRequest): Promise<RPCResponse<RevertBookmarkStateResponse>> {
    return this.call<RevertBookmarkStateRequest, RevertBookmarkStateResponse>('RPCRevertBookmarkState', params, 'local');
  }

  // ==========================================================================
  // SSG (SCAP Security Guide) Methods
  // ==========================================================================

  // SSG Import Job Methods
  async startSSGImportJob(runId?: string): Promise<RPCResponse<{ success: boolean; runId: string }>> {
    return this.call<{ runId?: string }, { success: boolean; runId: string }>('RPCSSGStartImportJob', { runId }, 'meta');
  }

  async stopSSGImportJob(): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<undefined, { success: boolean }>('RPCSSGStopImportJob', undefined, 'meta');
  }

  async pauseSSGImportJob(): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<undefined, { success: boolean }>('RPCSSGPauseImportJob', undefined, 'meta');
  }

  async resumeSSGImportJob(runId: string): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<{ runId: string }, { success: boolean }>('RPCSSGResumeImportJob', { runId }, 'meta');
  }

  async getSSGImportStatus(): Promise<RPCResponse<any>> {
    return this.call<undefined, any>('RPCSSGGetImportStatus', undefined, 'meta');
  }

  // SSG Guide Methods
  async listSSGGuides(product?: string, profileId?: string): Promise<RPCResponse<{ guides: SSGGuide[]; count: number }>> {
    return this.call<{ product?: string; profileId?: string }, { guides: SSGGuide[]; count: number }>('RPCSSGListGuides', { product, profileId }, 'local');
  }

  async getSSGGuide(id: string): Promise<RPCResponse<{ guide: SSGGuide }>> {
    return this.call<{ id: string }, { guide: SSGGuide }>('RPCSSGGetGuide', { id }, 'local');
  }

  async getSSGTree(guideId: string): Promise<RPCResponse<{ tree: any }>> {
    return this.call<{ guideId: string }, { tree: any }>('RPCSSGGetTree', { guideId }, 'local');
  }

  async getSSGTreeNodes(guideId: string): Promise<RPCResponse<{ nodes: TreeNode[]; count: number }>> {
    return this.call<{ guideId: string }, { nodes: TreeNode[]; count: number }>('RPCSSGGetTreeNode', { guideId }, 'local');
  }

  async getSSGGroup(id: string): Promise<RPCResponse<{ group: any }>> {
    return this.call<{ id: string }, { group: any }>('RPCSSGGetGroup', { id }, 'local');
  }

  async getSSGChildGroups(parentId?: string): Promise<RPCResponse<{ groups: any[]; count: number }>> {
    return this.call<{ parentId?: string }, { groups: any[]; count: number }>('RPCSSGGetChildGroups', { parentId }, 'local');
  }

  async getSSGRule(id: string): Promise<RPCResponse<{ rule: any }>> {
    return this.call<{ id: string }, { rule: any }>('RPCSSGGetRule', { id }, 'local');
  }

  async listSSGRules(groupId?: string, severity?: string, offset?: number, limit?: number): Promise<RPCResponse<{ rules: any[]; total: number }>> {
    return this.call<{ groupId?: string; severity?: string; offset?: number; limit?: number }, { rules: any[]; total: number }>('RPCSSGListRules', { groupId, severity, offset, limit }, 'local');
  }

  async getSSGChildRules(groupId: string): Promise<RPCResponse<{ rules: any[]; count: number }>> {
    return this.call<{ groupId: string }, { rules: any[]; count: number }>('RPCSSGGetChildRules', { groupId }, 'local');
  }

  // SSG Table Methods
  async listSSGTables(product?: string, tableType?: string): Promise<RPCResponse<{ tables: any[]; count: number }>> {
    return this.call<{ product?: string; tableType?: string }, { tables: any[]; count: number }>('RPCSSGListTables', { product, tableType }, 'local');
  }

  async getSSGTable(id: string): Promise<RPCResponse<{ table: any }>> {
    return this.call<{ id: string }, { table: any }>('RPCSSGGetTable', { id }, 'local');
  }

  async getSSGTableEntries(tableId: string, offset?: number, limit?: number): Promise<RPCResponse<{ entries: any[]; total: number }>> {
    return this.call<{ tableId: string; offset?: number; limit?: number }, { entries: any[]; total: number }>('RPCSSGGetTableEntries', { tableId, offset, limit }, 'local');
  }

  // SSG Manifest Methods
  async listSSGManifests(product?: string, limit?: number, offset?: number): Promise<RPCResponse<{ manifests: any[]; count: number }>> {
    return this.call<{ product?: string; limit?: number; offset?: number }, { manifests: any[]; count: number }>('RPCSSGListManifests', { product, limit, offset }, 'local');
  }

  async getSSGManifest(manifestId: string): Promise<RPCResponse<{ manifest: any }>> {
    return this.call<{ manifestId: string }, { manifest: any }>('RPCSSGGetManifest', { manifestId }, 'local');
  }

  async listSSGProfiles(product?: string, profileId?: string, limit?: number, offset?: number): Promise<RPCResponse<{ profiles: any[]; count: number }>> {
    return this.call<{ product?: string; profileId?: string; limit?: number; offset?: number }, { profiles: any[]; count: number }>('RPCSSGListProfiles', { product, profileId, limit, offset }, 'local');
  }

  async getSSGProfile(profileId: string): Promise<RPCResponse<{ profile: any }>> {
    return this.call<{ profileId: string }, { profile: any }>('RPCSSGGetProfile', { profileId }, 'local');
  }

  async getSSGProfileRules(profileId: string, limit?: number, offset?: number): Promise<RPCResponse<{ rules: any[]; count: number }>> {
    return this.call<{ profileId: string; limit?: number; offset?: number }, { rules: any[]; count: number }>('RPCSSGGetProfileRules', { profileId, limit, offset }, 'local');
  }

  // SSG Data Stream Methods
  async listSSGDataStreams(product?: string, limit?: number, offset?: number): Promise<RPCResponse<{ dataStreams: any[]; count: number }>> {
    return this.call<{ product?: string; limit?: number; offset?: number }, { dataStreams: any[]; count: number }>('RPCSSGListDataStreams', { product, limit, offset }, 'local');
  }

  async getSSGDataStream(dataStreamId: string): Promise<RPCResponse<{ dataStream: any; benchmark?: any }>> {
    return this.call<{ dataStreamId: string }, { dataStream: any; benchmark?: any }>('RPCSSGGetDataStream', { dataStreamId }, 'local');
  }

  async listDSProfiles(dataStreamId: string, limit?: number, offset?: number): Promise<RPCResponse<{ profiles: any[]; count: number }>> {
    return this.call<{ dataStreamId: string; limit?: number; offset?: number }, { profiles: any[]; count: number }>('RPCSSGListDSProfiles', { dataStreamId, limit, offset }, 'local');
  }

  async getDSProfile(profileId: string): Promise<RPCResponse<{ profile: any }>> {
    return this.call<{ profileId: string }, { profile: any }>('RPCSSGGetDSProfile', { profileId }, 'local');
  }

  async getDSProfileRules(profileId: string, limit?: number, offset?: number): Promise<RPCResponse<{ rules: any[]; count: number }>> {
    return this.call<{ profileId: string; limit?: number; offset?: number }, { rules: any[]; count: number }>('RPCSSGGetDSProfileRules', { profileId, limit, offset }, 'local');
  }

  async listDSGroups(dataStreamId: string, parentXccdfGroupId?: string, limit?: number, offset?: number): Promise<RPCResponse<{ groups: any[]; count: number }>> {
    return this.call<{ dataStreamId: string; parentXccdfGroupId?: string; limit?: number; offset?: number }, { groups: any[]; count: number }>('RPCSSGListDSGroups', { dataStreamId, parentXccdfGroupId, limit, offset }, 'local');
  }

  async listDSRules(dataStreamId: string, groupXccdfId?: string, severity?: string, limit?: number, offset?: number): Promise<RPCResponse<{ rules: any[]; total: number }>> {
    return this.call<{ dataStreamId: string; groupXccdfId?: string; severity?: string; limit?: number; offset?: number }, { rules: any[]; total: number }>('RPCSSGListDSRules', { dataStreamId, groupXccdfId, severity, limit, offset }, 'local');
  }

  async getDSRule(ruleId: string): Promise<RPCResponse<{ rule: any; references: any[]; identifiers: any[] }>> {
    return this.call<{ ruleId: string }, { rule: any; references: any[]; identifiers: any[] }>('RPCSSGGetDSRule', { ruleId }, 'local');
  }

  // SSG Cross-Reference Methods
  async getSSGCrossReferences(params: {
    sourceType?: string;
    sourceId?: string;
    targetType?: string;
    targetId?: string;
    limit?: number;
    offset?: number;
  }): Promise<RPCResponse<{ crossReferences: any[]; count: number }>> {
    return this.call<typeof params, { crossReferences: any[]; count: number }>('RPCSSGGetCrossReferences', params, 'local');
  }

  async findSSGRelatedObjects(params: {
    objectType: string;
    objectId: string;
    linkType?: string;
    limit?: number;
    offset?: number;
  }): Promise<RPCResponse<{ relatedObjects: any[]; count: number }>> {
    return this.call<typeof params, { relatedObjects: any[]; count: number }>('RPCSSGFindRelatedObjects', params, 'local');
  }

  // ============================================================================
  // UEE (Unified ETL Engine) Methods
  // ============================================================================

  async startProvider(providerId: string): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<{ providerId: string }, { success: boolean }>('RPCStartProvider', { providerId }, 'meta');
  }

  async pauseProvider(providerId: string): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<{ providerId: string }, { success: boolean }>('RPCPauseProvider', { providerId }, 'meta');
  }

  async stopProvider(providerId: string): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<{ providerId: string }, { success: boolean }>('RPCStopProvider', { providerId }, 'meta');
  }

  async updatePerformancePolicy(providerId: string, policy: any): Promise<RPCResponse<{ success: boolean }>> {
    return this.call<{ providerId: string; policy: any }, { success: boolean }>('RPCUpdatePerformancePolicy', { providerId, policy }, 'meta');
  }

  /**
   * Get the ETL tree showing macro FSM and all providers
   */
  async getEtlTree(): Promise<RPCResponse<{ tree: any }>> {
    if (this.useMock) {
      // Mock data for development
      return {
        retcode: 0,
        message: 'Success',
        payload: {
          tree: {
            macro: {
              id: 'main-orchestrator',
              state: 'ORCHESTRATING',
              providers: [
                {
                  id: 'cve-provider',
                  providerType: 'cve',
                  state: 'RUNNING',
                  processedCount: 245,
                  errorCount: 3,
                  permitsHeld: 5,
                  lastCheckpoint: 'v2e::nvd::cve::CVE-2024-00245',
                  createdAt: new Date(Date.now() - 3600000).toISOString(),
                  updatedAt: new Date().toISOString(),
                },
                {
                  id: 'cwe-provider',
                  providerType: 'cwe',
                  state: 'PAUSED',
                  processedCount: 128,
                  errorCount: 0,
                  permitsHeld: 0,
                  lastCheckpoint: 'v2e::mitre::cwe::CWE-128',
                  createdAt: new Date(Date.now() - 7200000).toISOString(),
                  updatedAt: new Date(Date.now() - 1800000).toISOString(),
                },
                {
                  id: 'capec-provider',
                  providerType: 'capec',
                  state: 'WAITING_QUOTA',
                  processedCount: 89,
                  errorCount: 1,
                  permitsHeld: 2,
                  lastCheckpoint: 'v2e::mitre::capec::CAPEC-89',
                  createdAt: new Date(Date.now() - 5400000).toISOString(),
                  updatedAt: new Date(Date.now() - 300000).toISOString(),
                },
              ],
              createdAt: new Date(Date.now() - 86400000).toISOString(),
              updatedAt: new Date().toISOString(),
            },
            totalProviders: 3,
            activeProviders: 1,
          },
        },
      };
    }
    return this.call<{}, { tree: any }>('RPCGetEtlTree', {}, 'meta');
  }

  /**
   * Get kernel performance metrics from the broker
   */
  async getKernelMetrics(): Promise<RPCResponse<{ metrics: any }>> {
    if (this.useMock) {
      // Mock data for development
      const now = new Date().toISOString();
      return {
        retcode: 0,
        message: 'Success',
        payload: {
          metrics: {
            p99Latency: 18.5 + Math.random() * 10, // 18.5-28.5ms
            bufferSaturation: 45 + Math.random() * 20, // 45-65%
            messageRate: 120 + Math.random() * 30, // 120-150 msgs/sec
            errorRate: Math.random() * 2, // 0-2 errors/sec
            timestamp: now,
          },
        },
      };
    }
    return this.call<{}, { metrics: any }>('RPCGetKernelMetrics', {}, 'broker');
  }

  /**
   * Get checkpoints for a specific provider
   */
  async getProviderCheckpoints(
    providerID: string,
    limit?: number,
    offset?: number
  ): Promise<RPCResponse<{ checkpoints: any[]; count: number }>> {
    if (this.useMock) {
      // Mock data for development
      const mockCheckpoints = Array.from({ length: limit || 10 }, (_, i) => ({
        urn: `v2e::nvd::cve::CVE-2024-${String((offset || 0) + i + 1).padStart(5, '0')}`,
        providerID,
        success: Math.random() > 0.1,
        errorMessage: Math.random() > 0.9 ? 'Network timeout' : undefined,
        processedAt: new Date(Date.now() - i * 60000).toISOString(),
      }));
      return {
        retcode: 0,
        message: 'Success',
        payload: {
          checkpoints: mockCheckpoints,
          count: 500,
        },
      };
    }
    return this.call<
      { providerID: string; limit?: number; offset?: number },
      { checkpoints: any[]; count: number }
    >('RPCGetProviderCheckpoints', { providerID, limit, offset }, 'meta');
  }

  // ============================================================================
  // Graph Analysis RPC Methods
  // ============================================================================

  /**
   * Get graph statistics
   */
  async getGraphStats(): Promise<RPCResponse<GetGraphStatsResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          node_count: 1500,
          edge_count: 3200,
        },
      };
    }
    return this.call<GetGraphStatsRequest, GetGraphStatsResponse>('RPCGetGraphStats', {}, 'analysis');
  }

  /**
   * Get neighbors of a node
   */
  async getNeighbors(urn: string): Promise<RPCResponse<GetNeighborsResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          neighbors: [
            'v2e::mitre::cwe::CWE-79',
            'v2e::mitre::cwe::CWE-89',
            'v2e::mitre::capec::CAPEC-66',
          ],
        },
      };
    }
    return this.call<GetNeighborsRequest, GetNeighborsResponse>('RPCGetNeighbors', { urn }, 'analysis');
  }

  /**
   * Find path between two nodes
   */
  async findPath(from: string, to: string): Promise<RPCResponse<FindPathResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          path: [
            'v2e::nvd::cve::CVE-2024-1234',
            'v2e::mitre::cwe::CWE-79',
            'v2e::mitre::capec::CAPEC-66',
            'v2e::mitre::attack::T1566',
          ],
          length: 4,
        },
      };
    }
    return this.call<FindPathRequest, FindPathResponse>('RPCFindPath', { from, to }, 'analysis');
  }

  /**
   * Get nodes by type
   */
  async getNodesByType(type: string): Promise<RPCResponse<GetNodesByTypeResponse>> {
    if (this.useMock) {
      const mockNodes = Array.from({ length: 10 }, (_, i) => ({
        urn: `v2e::nvd::cve::CVE-2024-${String(i + 1).padStart(5, '0')}`,
        properties: { severity: 'HIGH' },
      }));
      return {
        retcode: 0,
        message: 'success',
        payload: {
          nodes: mockNodes,
          count: 10,
        },
      };
    }
    return this.call<GetNodesByTypeRequest, GetNodesByTypeResponse>('RPCGetNodesByType', { type }, 'analysis');
  }

  /**
   * Build CVE graph
   */
  async buildCVEGraph(limit?: number): Promise<RPCResponse<BuildCVEGraphResponse>> {
    if (this.useMock) {
      await new Promise(resolve => setTimeout(resolve, 1000));
      return {
        retcode: 0,
        message: 'success',
        payload: {
          nodes_added: 250,
          edges_added: 180,
          total_nodes: 250,
          total_edges: 180,
        },
      };
    }
    return this.call<BuildCVEGraphRequest, BuildCVEGraphResponse>('RPCBuildCVEGraph', { limit }, 'analysis');
  }

  /**
   * Clear graph
   */
  async clearGraph(): Promise<RPCResponse<ClearGraphResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          status: 'cleared',
        },
      };
    }
    return this.call<ClearGraphRequest, ClearGraphResponse>('RPCClearGraph', {}, 'analysis');
  }

  /**
   * Get FSM state
   */
  async getFSMState(): Promise<RPCResponse<GetFSMStateResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          analyze_state: 'IDLE',
          graph_state: 'READY',
        },
      };
    }
    return this.call<GetFSMStateRequest, GetFSMStateResponse>('RPCGetFSMState', {}, 'analysis');
  }

  /**
   * Pause analysis
   */
  async pauseAnalysis(): Promise<RPCResponse<PauseAnalysisResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          status: 'paused',
        },
      };
    }
    return this.call<PauseAnalysisRequest, PauseAnalysisResponse>('RPCPauseAnalysis', {}, 'analysis');
  }

  /**
   * Resume analysis
   */
  async resumeAnalysis(): Promise<RPCResponse<ResumeAnalysisResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          status: 'resumed',
        },
      };
    }
    return this.call<ResumeAnalysisRequest, ResumeAnalysisResponse>('RPCResumeAnalysis', {}, 'analysis');
  }

  /**
   * Save graph to disk
   */
  async saveGraph(): Promise<RPCResponse<SaveGraphResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          status: 'saved',
          node_count: 250,
          edge_count: 180,
          last_saved: new Date().toISOString(),
        },
      };
    }
    return this.call<SaveGraphRequest, SaveGraphResponse>('RPCSaveGraph', {}, 'analysis');
  }

  /**
   * Load graph from disk
   */
  async loadGraph(): Promise<RPCResponse<LoadGraphResponse>> {
    if (this.useMock) {
      return {
        retcode: 0,
        message: 'success',
        payload: {
          status: 'loaded',
          node_count: 250,
          edge_count: 180,
        },
      };
    }
    return this.call<LoadGraphRequest, LoadGraphResponse>('RPCLoadGraph', {}, 'analysis');
  }
}

// ============================================================================
// Singleton Instance
// ============================================================================

// Check if we're in development mode and should use mock data
const useMockData = process.env.NEXT_PUBLIC_USE_MOCK_DATA === 'true';

export const rpcClient = new RPCClient({
  useMock: useMockData,
});
