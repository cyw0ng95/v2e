import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe('Meta Service', () => {
  it('should get service status', async () => {
    const response = await rpcClient.call('RPCGetStatus', {}, 'meta');

    await assertRpcSuccess(response);

    const status = response.payload as Record<string, unknown>;
    expect(status).toBeDefined();
    expect(status.state).toBeDefined();
  });

  it('should list jobs (possibly empty)', async () => {
    const response = await rpcClient.call('RPCListJobs', {}, 'meta');

    await assertRpcSuccess(response);

    const jobs = response.payload as unknown[];
    expect(Array.isArray(jobs)).toBe(true);
  });

  it('should get ETL tree', async () => {
    const response = await rpcClient.call('RPCGetEtlTree', {}, 'meta');

    await assertRpcSuccess(response);

    const tree = response.payload as Record<string, unknown>;
    expect(tree).toBeDefined();
    expect(tree.macro).toBeDefined();
  });
});
