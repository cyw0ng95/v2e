import { RPCRequest, RPCResponse } from '../../website/lib/types.js';

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
   * Make an RPC call to the backend
   */
  async call<TRequest, TResponse>(
    method: string,
    params?: TRequest,
    target: string = 'meta'
  ): Promise<RPCResponse<TResponse>> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const request: RPCRequest<TRequest> = {
        method,
        params: params ? convertKeysToSnakeCase(params) : undefined,
        target
      };

      const response = await fetch(`${this.baseUrl}/restful/rpc`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
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
