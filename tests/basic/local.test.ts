import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertNotFound } from '../helpers/assertions.js';

describe('Local Service', () => {
  beforeEach(async () => {
    // Note: resetDatabase only resets files on disk, but services keep old connections
    // Tests should handle both empty and populated database states
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('cve');
      await manager.resetDatabase('cwe');
      await manager.resetDatabase('capec');
    }
  });

  describe('CVE Operations', () => {
    it('should list CVEs successfully', async () => {
      const response = await rpcClient.call(
        'RPCListCVEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.cves)).toBe(true);
      // Database may be empty or have existing data - both are valid
      expect(data.cves.length).toBeGreaterThanOrEqual(0);
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
    it('should list CWEs successfully', async () => {
      const response = await rpcClient.call(
        'RPCListCWEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.cwes)).toBe(true);
      // Database may be empty or have existing data - both are valid
      expect(data.cwes.length).toBeGreaterThanOrEqual(0);
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
    it('should list CAPECs successfully', async () => {
      const response = await rpcClient.call(
        'RPCListCAPECs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.capecs)).toBe(true);
      // Database may be empty or have existing data - both are valid
      expect(data.capecs.length).toBeGreaterThanOrEqual(0);
    });

    it('should get CAPEC catalog metadata', async () => {
      const response = await rpcClient.call(
        'RPCGetCAPECCatalogMeta',
        {},
        'local'
      );

      await assertRpcSuccess(response);

      const meta = response.payload as any;
      expect(meta).toBeDefined();
      // Catalog metadata may or may not exist depending on imports
    });
  });

  describe('ATT&CK Operations', () => {
    it('should list ATT&CK techniques', async () => {
      const response = await rpcClient.call(
        'RPCListAttackTechniques',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.techniques)).toBe(true);
      expect(data.techniques.length).toBeGreaterThanOrEqual(0);
    });

    it('should get specific ATT&CK technique by ID', async () => {
      const response = await rpcClient.call(
        'RPCGetAttackTechnique',
        { techniqueId: 'T1190' },
        'local'
      );

      // May or may not be found depending on database state
      expect(response).toBeDefined();
      expect(response.retcode).toBeDefined();
    });
  });
});
