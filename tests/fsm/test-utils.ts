import { rpcClient } from '../src/rpc-client.js';

export const TEST_PROVIDER_IDS = ['cve', 'cwe', 'capec', 'attack', 'cce'];

export interface EtlTreeResponse {
  tree?: {
    macro_fsm?: {
      state: string;
      total_providers: number;
      active_providers: number;
    };
    providers?: Array<{
      id: string;
      type: string;
      state: string;
      processed_count?: number;
      last_checkpoint?: string;
    }>;
  };
  macro_fsm?: {
    state: string;
    total_providers: number;
    active_providers: number;
  };
  providers?: Array<{
    id: string;
    type: string;
    state: string;
    processed_count?: number;
    last_checkpoint?: string;
  }>;
}

export async function resetAllProviders(): Promise<void> {
  await rpcClient.call('RPCFSMStopAllProviders', {}, 'meta');
  await waitForAllProvidersState('TERMINATED', 15000);
}

function extractProviders(response: EtlTreeResponse): Array<{id: string; state: string}> {
  const providers = response.providers || response.tree?.providers || [];
  return providers.map(p => ({ id: p.id, state: p.state }));
}

export async function waitForAllProvidersState(
  targetState: string,
  timeout: number = 15000
): Promise<boolean> {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    const response = await rpcClient.call<{}, EtlTreeResponse>(
      'RPCGetEtlTree',
      {},
      'meta'
    );

    if (response.retcode !== 0) {
      await new Promise(resolve => setTimeout(resolve, 1000));
      continue;
    }

    const providers = extractProviders(response);
    const allInState = providers.every(p => p.state === targetState);
    
    if (allInState) {
      return true;
    }
    
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  return false;
}

export async function waitForAnyProviderState(
  targetState: string,
  timeout: number = 15000
): Promise<boolean> {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    const response = await rpcClient.call<{}, EtlTreeResponse>(
      'RPCGetEtlTree',
      {},
      'meta'
    );

    if (response.retcode !== 0) {
      await new Promise(resolve => setTimeout(resolve, 1000));
      continue;
    }

    const providers = extractProviders(response);
    const anyInState = providers.some(p => p.state === targetState);
    
    if (anyInState) {
      return true;
    }
    
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  return false;
}

export function getTestProviderIds(): string[] {
  return [...TEST_PROVIDER_IDS];
}

export async function ensureProvidersIdle(): Promise<void> {
  await resetAllProviders();
}
