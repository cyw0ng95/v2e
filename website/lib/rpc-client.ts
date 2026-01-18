/**
 * RPC Client for v2e access service
 * Implements the Service-Consumer pattern to bridge UI and backend
 */

import type {
  RPCRequest,
  RPCResponse,
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
  HealthResponse,
  CVEItem,
  CWEItem,
  ListCWEsRequest,
  ListCWEsResponse,
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
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
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
      console.error('[rpc-client] request failed', { method, target, error });
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

  async getCWE(cweId: string): Promise<RPCResponse<{ cwe: CWEItem }>> {
    return this.call<{ cweId: string }, { cwe: CWEItem }>('RPCGetCWEByID', { cweId }, 'local');
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
