/**
 * RPC Client for v2e access service
 * Implements the Service-Consumer pattern to bridge UI and backend
 */

import React from 'react';
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

// Track pending requests to prevent duplicates
const pendingRequests = new Map<string, Promise<RPCResponse<unknown>>>();

// Create a cache for RPC calls to deduplicate requests
const cachedCall = React.cache(async function (
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
    console.debug('[rpc-client] Deduplicating request:', requestKey);
    return pendingRequests.get(requestKey)!;
  }
  
  if (useMock) {
    await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY_MS));
    const result = getMockResponseForCache(method, params);
    pendingRequests.delete(requestKey); // Clean up
    return result;
  }

  const request: RPCRequest<unknown> = {
    method,
    params: params ? convertKeysToSnakeCase(params) : undefined,
    target,
  };

  // Create promise for this request
  const requestPromise = (async () => {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      console.debug('[rpc-client] Making request:', { method, target, params });
      
      const response = await fetch(`${baseUrl}/restful/rpc`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      const raw = await response.text();

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      let rpcResponse: RPCResponse<unknown>;
      try {
        rpcResponse = JSON.parse(raw);
      } catch (err) {
        throw new Error('Invalid JSON response from RPC endpoint');
      }

      if (rpcResponse.payload) {
        rpcResponse.payload = convertKeysToCamelCase(rpcResponse.payload);
      }

      // Log failed RPC calls for debugging/bugfix purposes
      if (rpcResponse.retcode !== 0) {
        console.log('[rpc-error] Failed RPC call:\n' +
          'REQUEST: ' + JSON.stringify({ method, params, target }) + '\n' +
          'RESPONSE: ' + JSON.stringify({ retcode: rpcResponse.retcode, message: rpcResponse.message, payload: rpcResponse.payload }));
      }

      console.debug('[rpc-client] Request completed:', { method, retcode: rpcResponse.retcode });
      return rpcResponse;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        return {
          retcode: 500,
          message: 'Request timeout',
          payload: null,
        };
      }
      return {
        retcode: 500,
        message: error instanceof Error ? error.message : 'Unknown error',
        payload: null,
      };
    } finally {
      // Clean up pending request
      pendingRequests.delete(requestKey);
    }
  })();
  
  // Store the promise
  pendingRequests.set(requestKey, requestPromise);
  return requestPromise;
});

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

      // Notes Framework Mock Responses
      case 'RPCCreateBookmark': {
        const req = params as CreateBookmarkRequest | undefined;
        const mockBookmark: Bookmark = {
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
        return {
          retcode: 0,
          message: 'success',
          payload: { success: true, bookmark: mockBookmark } as unknown as TResponse,
        };
      }

      case 'RPCGetBookmark': {
        const req = params as GetBookmarkRequest | undefined;
        const mockBookmark: Bookmark = {
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
          id: req?.id || 1,
          bookmark_id: 1,
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
          id: Math.floor(Math.random() * 1000),
          bookmark_id: req?.bookmark_id || 1,
          front_content: req?.front_content || 'What is Log4Shell?',
          back_content: req?.back_content || 'A critical vulnerability in Log4j allowing RCE',
          card_type: req?.card_type || 'basic',
          learning_state: 'to_review',
          author: req?.author,
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
            id: 1,
            bookmark_id: 1,
            front_content: 'What is Log4Shell?',
            back_content: 'A critical vulnerability in Log4j allowing RCE',
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
            front_content: 'How to mitigate Log4Shell?',
            back_content: 'Upgrade to Log4j 2.15.0 or apply JVM parameters',
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
}

// ============================================================================
// Singleton Instance
// ============================================================================

// Check if we're in development mode and should use mock data
const useMockData = process.env.NEXT_PUBLIC_USE_MOCK_DATA === 'true';

export const rpcClient = new RPCClient({
  useMock: useMockData,
});
