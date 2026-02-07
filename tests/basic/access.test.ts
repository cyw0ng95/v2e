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

  it('should reject missing method field', async () => {
    const response = await fetch(`${TEST_CONFIG.API_BASE_URL}/restful/rpc`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ params: {} })
    });

    // Access service validates input - should return 400
    expect([400, 200]).toContain(response.status);

    if (response.status === 200) {
      const data = await response.json();
      expect(data.retcode).not.toBe(0);
    }
  });

  it('should forward RPC calls to broker', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    expect(response.retcode).toBe(0);
    expect(response.payload).toBeDefined();
  });

  it('should return error for invalid RPC method', async () => {
    const response = await rpcClient.call('RPCInvalidMethodXYZ', {}, 'broker');

    expect(response.retcode).not.toBe(0);
  });
});
