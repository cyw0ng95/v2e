import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertEmpty, assertNotFound } from '../helpers/assertions.js';

describe('Local Service - Empty Database', () => {
  beforeEach(async () => {
    // Reset databases via service manager
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('cve');
      await manager.resetDatabase('cwe');
      await manager.resetDatabase('capec');
    }
  });

  describe('CVE Operations', () => {
    it('should return empty CVE list', async () => {
      const response = await rpcClient.call(
        'RPCListCVEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.cves)).toBe(true);
      assertEmpty(data.cves);
    });

    it('should return not found for unknown CVE', async () => {
      const response = await rpcClient.call(
        'RPCGetCVE',
        { cveId: 'CVE-2024-9999' },
        'local'
      );

      assertNotFound(response);
    });
  });

  describe('CWE Operations', () => {
    it('should return empty CWE list', async () => {
      const response = await rpcClient.call(
        'RPCListCWEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.cwes)).toBe(true);
      assertEmpty(data.cwes);
    });

    it('should return not found for unknown CWE', async () => {
      const response = await rpcClient.call(
        'RPCGetCWEByID',
        { cweId: 'CWE-999999' },
        'local'
      );

      assertNotFound(response);
    });
  });

  describe('CAPEC Operations', () => {
    it('should return empty CAPEC list', async () => {
      const response = await rpcClient.call(
        'RPCListCAPECs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.capecs)).toBe(true);
      assertEmpty(data.capecs);
    });
  });
});
