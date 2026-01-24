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
} from './types';

// ============================================================================
// Configuration
// ============================================================================

const DEFAULT_API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
const DEFAULT_TIMEOUT = 30000; // 30 seconds
const MOCK_DELAY_MS = 500; // Simulate network delay in mock mode

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

// ============================================================================
// Case Conversion Utilities
// ============================================================================

/**
 * Convert PascalCase/snake_case to camelCase
 */
function toCamelCase(str: string): string {
  // Handle snake_case -> camelCase
  if (str.indexOf('_') >= 0) {
    return str.replace(/_([a-zA-Z0-9])/g, (_, letter) => letter.toUpperCase());
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
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
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
    // Mock mode: return simulated data
    if (this.useMock) {
      await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY_MS));
      return this.getMockResponse<TResponse>(method, params);
    }

    // Real mode: make HTTP request
    const request: RPCRequest<unknown> = {
      method,
      params: params ? convertKeysToSnakeCase(params) : undefined,
      target,
    };

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      // Debug: log request details
      console.debug('[rpc-client] RPC request', { url: `${this.baseUrl}/restful/rpc`, method, target, params: request.params });

      const response = await fetch(`${this.baseUrl}/restful/rpc`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      // Read raw response text so we can log invalid JSON bodies as well
      const raw = await response.text();

      if (!response.ok) {
        console.error('[rpc-client] HTTP error response', { status: response.status, body: raw });
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      let rpcResponse: RPCResponse<unknown>;
      try {
        rpcResponse = JSON.parse(raw);
      } catch (err) {
        console.error('[rpc-client] Failed to parse JSON response', { raw, err });
        throw new Error('Invalid JSON response from RPC endpoint');
      }

      // Debug: log parsed RPC response
      console.debug('[rpc-client] RPC response', rpcResponse);

      // Convert response payload keys to camelCase
      if (rpcResponse.payload) {
        rpcResponse.payload = convertKeysToCamelCase(rpcResponse.payload);
      }

      return rpcResponse as RPCResponse<TResponse>;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        console.error('[rpc-client] request aborted (timeout)', { method, target });
        return {
          retcode: 500,
          message: 'Request timeout',
          payload: null,
        };
      }
      console.error('[rpc-client] request failed', {
        method,
        target,
        error: error instanceof Error ? error.message : error,
      });
      return {
        retcode: 500,
        message: error instanceof Error ? error.message : 'Unknown error',
        payload: null,
      };
    }
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
        const sample: any[] = [
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

  async startCWEImport(params?: Record<string, unknown>): Promise<RPCResponse<any>> {
    return this.call<any, any>('RPCStartCWEImport', params);
  }

  async startCAPECImport(params?: Record<string, unknown>): Promise<RPCResponse<any>> {
    return this.call<any, any>('RPCStartCAPECImport', params);
  }

  async startATTACKImport(params?: Record<string, unknown>): Promise<RPCResponse<any>> {
    return this.call<any, any>('RPCStartATTACKImport', params);
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

  async listCAPECs(offset?: number, limit?: number): Promise<RPCResponse<{ capecs: any[]; offset: number; limit: number; total: number }>> {
    return this.call<{ offset?: number; limit?: number }, { capecs: any[]; offset: number; limit: number; total: number }>(
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
}

// ============================================================================
// Singleton Instance
// ============================================================================

// Check if we're in development mode and should use mock data
const useMockData = process.env.NEXT_PUBLIC_USE_MOCK_DATA === 'true';

export const rpcClient = new RPCClient({
  useMock: useMockData,
});
