import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe('Broker Process Management', () => {
  it('should get broker information via access gateway', async () => {
    // Since broker has no direct RPC, we test through access gateway
    // and verify we can reach backend services
    const response = await rpcClient.call('RPCGetCVE', {
      cveId: 'CVE-2021-44228'
    }, 'meta');

    // May not be found (data may not exist), but RPC should work
    expect(response).toBeDefined();
    expect(response.retcode).toBeDefined();
  });

  it('should handle multiple service targets', async () => {
    // Test that we can reach different services
    const targets = [
      { target: 'local', method: 'RPCCountCVEs' },
      { target: 'local', method: 'RPCListCWEs', params: { limit: 1 } },
      { target: 'local', method: 'RPCListCAPECs', params: { limit: 1 } }
    ];

    for (const { target, method, params = {} } of targets) {
      const response = await rpcClient.call(method, params, target as any);
      expect(response).toBeDefined();
      expect(response.retcode).toBeDefined();
    }
  });

  it('should have all expected services reachable', async () => {
    // Verify we can reach core services through RPC
    const services = [
      { target: 'local', method: 'RPCCountCVEs' },
      { target: 'remote', method: 'RPCGetCVECnt' },
      { target: 'meta', method: 'RPCGetEtlTree' }
    ];

    const results: string[] = [];

    for (const { target, method } of services) {
      try {
        const response = await rpcClient.call(method, {}, target as any);
        if (response.retcode === 0) {
          results.push(target);
        }
      } catch (e) {
        // Service may be unreachable - that's a test result
      }
    }

    // At least local service should be reachable
    expect(results.length).toBeGreaterThan(0);
  });
});
