import { RPCRequest, RPCResponse } from '../../website/lib/types.js';

const methodToPathMap: Record<string, { path: string; target: string }> = {
  RPCGetCVE: { path: '/cve/get', target: 'local' },
  RPCCreateCVE: { path: '/cve/create', target: 'local' },
  RPCUpdateCVE: { path: '/cve/update', target: 'local' },
  RPCDeleteCVE: { path: '/cve/delete', target: 'local' },
  RPCListCVEs: { path: '/cve/list', target: 'local' },
  RPCCountCVEs: { path: '/cve/count', target: 'local' },
  RPCGetCWEByID: { path: '/cwe/get', target: 'local' },
  RPCListCWEs: { path: '/cwe/list', target: 'local' },
  RPCImportCWEs: { path: '/cwe/import', target: 'local' },
  RPCSaveCWEView: { path: '/cwe-view/save', target: 'local' },
  RPCGetCWEViewByID: { path: '/cwe-view/get', target: 'local' },
  RPCListCWEViews: { path: '/cwe-view/list', target: 'local' },
  RPCDeleteCWEView: { path: '/cwe-view/delete', target: 'local' },
  RPCStartCWEViewJob: { path: '/cwe-view/start-job', target: 'meta' },
  RPCStopCWEViewJob: { path: '/cwe-view/stop-job', target: 'meta' },
  RPCGetCAPECByID: { path: '/capec/get', target: 'local' },
  RPCListCAPECs: { path: '/capec/list', target: 'local' },
  RPCImportCAPECs: { path: '/capec/import', target: 'local' },
  RPCForceImportCAPECs: { path: '/capec/force-import', target: 'local' },
  RPCGetCAPECCatalogMeta: { path: '/capec/metadata', target: 'local' },
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
  RPCListASVS: { path: '/asvs/list', target: 'local' },
  RPCGetASVSByID: { path: '/asvs/get', target: 'local' },
  RPCImportASVS: { path: '/asvs/import', target: 'local' },
  RPCGetCCEByID: { path: '/cce/get', target: 'local' },
  RPCListCCEs: { path: '/cce/list', target: 'local' },
  RPCImportCCEs: { path: '/cce/import', target: 'local' },
  RPCImportCCE: { path: '/cce/import-one', target: 'local' },
  RPCCountCCEs: { path: '/cce/count', target: 'local' },
  RPCDeleteCCE: { path: '/cce/delete', target: 'local' },
  RPCUpdateCCE: { path: '/cce/update', target: 'local' },
  RPCStartSession: { path: '/session/start', target: 'meta' },
  RPCStartTypedSession: { path: '/session/start-typed', target: 'meta' },
  RPCStopSession: { path: '/session/stop', target: 'meta' },
  RPCGetSessionStatus: { path: '/session/status', target: 'meta' },
  RPCPauseJob: { path: '/job/pause', target: 'meta' },
  RPCResumeJob: { path: '/job/resume', target: 'meta' },
  RPCCreateBookmark: { path: '/bookmark/create', target: 'local' },
  RPCGetBookmark: { path: '/bookmark/get', target: 'local' },
  RPCUpdateBookmark: { path: '/bookmark/update', target: 'local' },
  RPCDeleteBookmark: { path: '/bookmark/delete', target: 'local' },
  RPCListBookmarks: { path: '/bookmark/list', target: 'local' },
  RPCAddNote: { path: '/note/add', target: 'local' },
  RPCGetNote: { path: '/note/get', target: 'local' },
  RPCUpdateNote: { path: '/note/update', target: 'local' },
  RPCDeleteNote: { path: '/note/delete', target: 'local' },
  RPCGetNotesByBookmark: { path: '/note/by-bookmark', target: 'local' },
  RPCCreateMemoryCard: { path: '/memory-card/create', target: 'local' },
  RPCGetMemoryCard: { path: '/memory-card/get', target: 'local' },
  RPCUpdateMemoryCard: { path: '/memory-card/update', target: 'local' },
  RPCDeleteMemoryCard: { path: '/memory-card/delete', target: 'local' },
  RPCListMemoryCards: { path: '/memory-card/list', target: 'local' },
  RPCRateMemoryCard: { path: '/memory-card/rate', target: 'local' },
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
  RPCGetSysMetrics: { path: '/system/metrics', target: 'sysmon' },
  RPCGetEtlTree: { path: '/etl/tree', target: 'meta' },
  RPCStartProvider: { path: '/etl/provider/start', target: 'meta' },
  RPCPauseProvider: { path: '/etl/provider/pause', target: 'meta' },
  RPCStopProvider: { path: '/etl/provider/stop', target: 'meta' },
  RPCUpdatePerformancePolicy: { path: '/etl/performance-policy', target: 'meta' },
  RPCGetKernelMetrics: { path: '/etl/kernel-metrics', target: 'meta' },
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

// Case conversion utilities (from website/lib/rpc-client.ts)
function toCamelCase(str: string): string {
  if (str.indexOf('_') >= 0) {
    return str.replace(/_([a-zA-Z0-9])/g, (_, letter) => letter.toUpperCase());
  }
  if (str === str.toUpperCase()) {
    return str.toLowerCase();
  }
  return str.charAt(0).toLowerCase() + str.slice(1);
}

function toSnakeCase(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
}

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

export class TestRPCClient {
  private baseUrl: string;
  private timeout: number;

  constructor(baseUrl?: string, timeout?: number) {
    this.baseUrl = baseUrl || process.env.V2E_API_BASE_URL || 'http://localhost:8080';
    this.timeout = timeout || parseInt(process.env.V2E_TEST_TIMEOUT || '30000');
  }

  /**
   * Make an RPC call to the backend using path-based routing
   */
  async call<TRequest, TResponse>(
    method: string,
    params?: TRequest,
    target: string = 'meta'
  ): Promise<RPCResponse<TResponse>> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const pathRoute = methodToPathMap[method];
      let url: string;
      let body: string;

      if (pathRoute) {
        url = `${this.baseUrl}/restful/rpc${pathRoute.path}`;
        body = JSON.stringify(params ? convertKeysToSnakeCase(params) : {});
      } else {
        const request: RPCRequest<TRequest> = {
          method,
          params: params ? convertKeysToSnakeCase(params) : undefined,
          target
        };
        url = `${this.baseUrl}/restful/rpc`;
        body = JSON.stringify(request);
      }

      const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body,
        signal: controller.signal
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const rpcResponse: RPCResponse<TResponse> = await response.json();

      if (rpcResponse.payload) {
        rpcResponse.payload = convertKeysToCamelCase(rpcResponse.payload);
      }

      return rpcResponse;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error && error.name === 'AbortError') {
        return {
          retcode: 500,
          message: 'Request timeout',
          payload: null
        } as RPCResponse<TResponse>;
      }

      return {
        retcode: 500,
        message: error instanceof Error ? error.message : 'Unknown error',
        payload: null
      } as RPCResponse<TResponse>;
    }
  }
}

// Singleton instance
export const rpcClient = new TestRPCClient();
