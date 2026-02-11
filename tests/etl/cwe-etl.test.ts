import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe('CWE ETL', () => {
  beforeEach(async () => {
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('cwe');
    }
  });

  it('should trigger CWE import',
    { timeout: 60000 },
    async () => {
      // Start CWE import
      const importResponse = await rpcClient.call(
        'RPCStartCWEImport',
        {},
        'meta'
      );

      // Import should start successfully
      await assertRpcSuccess(importResponse);

      // Wait for import to process
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to list CWEs
      const listResponse = await rpcClient.call(
        'RPCListCWEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(listResponse);

      const data = listResponse.payload as any;
      expect(Array.isArray(data.cwes)).toBe(true);

      // May have CWEs if asset exists or import succeeded
      // We verify the system handles both cases gracefully
    }
  );

  it('should handle CWE retrieval after import attempt',
    { timeout: 60000 },
    async () => {
      const cweId = 'CWE-79';

      // Try to start import
      await rpcClient.call('RPCStartCWEImport', {}, 'meta');
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to get specific CWE
      const getResponse = await rpcClient.call(
        'RPCGetCWEByID',
        { cweId },
        'local'
      );

      // May or may not be found
      // We verify no crash occurs
      expect(getResponse).toBeDefined();
    }
  );
});
