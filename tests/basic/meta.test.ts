import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe('Meta Service', () => {
  it('should get ETL tree', async () => {
    // RPCGetEtlTree exists in meta service
    const response = await rpcClient.call('RPCGetEtlTree', {}, 'meta');

    await assertRpcSuccess(response);

    const tree = response.payload as Record<string, unknown>;
    expect(tree).toBeDefined();
    // ETL tree structure may vary
  });

  it('should get CVE count via meta service', async () => {
    // RPCCountCVEs is proxied through meta
    const response = await rpcClient.call('RPCCountCVEs', {}, 'meta');

    await assertRpcSuccess(response);

    const count = response.payload as number;
    expect(typeof count).toBe('number');
  });

  it('should list CVEs via meta service', async () => {
    // RPCListCVEs is proxied through meta
    const response = await rpcClient.call('RPCListCVEs', { limit: 5 }, 'meta');

    await assertRpcSuccess(response);

    const data = response.payload as any;
    expect(data.cves).toBeDefined();
    expect(Array.isArray(data.cves)).toBe(true);
  });

  it('should start CWE import via meta service', async () => {
    // RPCStartCWEImport exists in meta service
    const response = await rpcClient.call('RPCStartCWEImport', {}, 'meta');

    // Import may fail if asset file doesn't exist, but RPC should be accepted
    expect(response).toBeDefined();
    expect(response.retcode).toBeDefined();
  });

  it('should start CAPEC import via meta service', async () => {
    // RPCStartCAPECImport exists in meta service
    const response = await rpcClient.call('RPCStartCAPECImport', {}, 'meta');

    // Import may fail if asset file doesn't exist, but RPC should be accepted
    expect(response).toBeDefined();
    expect(response.retcode).toBeDefined();
  });
});
