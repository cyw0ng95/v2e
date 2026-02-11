import { RPCResponse } from '../../website/lib/types.js';

/**
 * Assert RPC call succeeded (retcode === 0)
 *
 * Test case: Verify RPC calls complete successfully
 */
export async function assertRpcSuccess(response: RPCResponse<unknown>): Promise<void> {
  if (response.retcode !== 0) {
    throw new Error(`RPC failed: ${response.message} (retcode=${response.retcode})`);
  }
}

/**
 * Assert all expected services are present in process list
 *
 * Test case: Verify broker detects all subprocesses
 */
export function assertHasServices(
  processes: Array<{ id: string }>,
  expected: string[]
): void {
  const ids = new Set(processes.map(p => p.id));
  const missing = expected.filter(id => !ids.has(id));

  if (missing.length > 0) {
    throw new Error(`Missing services: ${missing.join(', ')}`);
  }
}

/**
 * Assert array has expected count
 *
 * Test case: Verify pagination returns correct number of items
 */
export function assertDataCount(data: unknown, expected: number): void {
  if (!Array.isArray(data)) {
    throw new Error(`Expected array, got ${typeof data}`);
  }

  if (data.length !== expected) {
    throw new Error(`Expected ${expected} items, got ${data.length}`);
  }
}

/**
 * Assert array is empty
 *
 * Test case: Verify empty database state
 */
export function assertEmpty(data: unknown): void {
  if (!Array.isArray(data)) {
    throw new Error(`Expected array, got ${typeof data}`);
  }

  if (data.length !== 0) {
    throw new Error(`Expected empty array, got ${data.length} items`);
  }
}

/**
 * Assert response indicates not found
 *
 * Test case: Verify unknown IDs return not found
 * Accepts either non-zero retcode OR null payload
 */
export function assertNotFound(response: RPCResponse<unknown>): void {
  const isNotFound = response.retcode !== 0 || response.payload === null;

  if (!isNotFound) {
    throw new Error('Expected not found, but got success response');
  }
}

/**
 * Assert all processes are in running/idle state
 *
 * Test case: Verify broker health
 */
export function assertProcessesHealthy(processes: Array<{ id: string; status: string }>): void {
  for (const proc of processes) {
    if (!proc.status.match(/running|idle/)) {
      throw new Error(`Process ${proc.id} has unhealthy status: ${proc.status}`);
    }
  }
}
