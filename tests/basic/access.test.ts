import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { TEST_CONFIG } from '../helpers/fixtures.js';

describe('Access Service', () => {
  it('should return health status', async () => {
    const response = await fetch(`${TEST_CONFIG.API_BASE_URL}/restful/health`);

    expect(response.ok).toBe(true);

    const data = await response.json();
    expect(data).toHaveProperty('status');
  });

  it('should reject empty body on path-based endpoint', async () => {
    const response = await fetch(`${TEST_CONFIG.API_BASE_URL}/restful/rpc/cve/list`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({})
    });

    expect(response.ok).toBe(true);
    const data = await response.json();
    expect(data.retcode).not.toBe(0);
  });

  it('should forward RPC calls to local service', async () => {
    const response = await rpcClient.call('RPCCountCVEs', {}, 'local');

    expect(response.retcode).toBe(0);
    expect(response.payload).toBeDefined();
  });

  it('should return error for invalid RPC method', async () => {
    const response = await rpcClient.call('RPCInvalidMethodXYZ', {}, 'local');

    expect(response.retcode).not.toBe(0);
  });
});
