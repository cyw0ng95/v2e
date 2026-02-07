import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertHasServices, assertProcessesHealthy } from '../helpers/assertions.js';
import { EXPECTED_SERVICES } from '../helpers/fixtures.js';

describe('Broker RPC', () => {
  it('should list all processes', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    await assertRpcSuccess(response);

    const processes = response.payload as Array<{ id: string; status: string }>;
    expect(Array.isArray(processes)).toBe(true);
    expect(processes.length).toBeGreaterThan(0);
  });

  it('should have all expected services', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    await assertRpcSuccess(response);

    const processes = response.payload as Array<{ id: string; status: string }>;
    assertHasServices(processes, [...EXPECTED_SERVICES]);
  });

  it('should get broker status', async () => {
    const response = await rpcClient.call('RPCGetStatus', {}, 'broker');

    await assertRpcSuccess(response);

    const status = response.payload as Record<string, unknown>;
    expect(status).toBeDefined();
    expect(status.state).toBeDefined();
  });

  it('should have all services in healthy state', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    await assertRpcSuccess(response);

    const processes = response.payload as Array<{ id: string; status: string }>;
    assertProcessesHealthy(processes);
  });
});
