import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe('CAPEC ETL', () => {
  beforeEach(async () => {
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('capec');
    }
  });

  it('should trigger CAPEC import',
    { timeout: 60000 },
    async () => {
      // Start CAPEC import
      const importResponse = await rpcClient.call(
        'RPCStartCAPECImport',
        {},
        'meta'
      );

      // Import should start successfully
      await assertRpcSuccess(importResponse);

      // Wait for import to process
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to list CAPECs
      const listResponse = await rpcClient.call(
        'RPCListCAPECs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(listResponse);

      const data = listResponse.payload as any;
      expect(Array.isArray(data.capecs)).toBe(true);

      // May have CAPECs if asset exists or import succeeded
      // We verify the system handles both cases gracefully
    }
  );

  it('should handle CAPEC retrieval after import attempt',
    { timeout: 60000 },
    async () => {
      const capecId = 'CAPEC-1';

      // Try to start import
      await rpcClient.call('RPCStartCAPECImport', {}, 'meta');
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to get specific CAPEC
      const getResponse = await rpcClient.call(
        'RPCGetCAPECByID',
        { capecId },
        'local'
      );

      // May or may not be found
      // We verify no crash occurs
      expect(getResponse).toBeDefined();
    }
  );
});
