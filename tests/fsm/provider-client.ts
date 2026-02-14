import { rpcClient } from '../src/rpc-client.js';

export interface ProviderState {
  state: string;
  providerId: string;
  error?: string;
}

export class ProviderFSMClient {
  async getProviderState(providerId: string): Promise<ProviderState> {
    const response = await rpcClient.call<{id: string}, {state: string}>(
      'RPCGetEtlTree',
      { id: providerId },
      'meta'
    );

    if (response.retcode !== 0) {
      return {
        state: 'UNKNOWN',
        providerId,
        error: response.message,
      };
    }

    return {
      state: response.payload?.state || 'UNKNOWN',
      providerId,
    };
  }

  async startProvider(providerId: string): Promise<{success: boolean; error?: string}> {
    const response = await rpcClient.call<{id: string}, {success: boolean}>(
      'RPCStartProvider',
      { id: providerId },
      'meta'
    );

    return {
      success: response.retcode === 0,
      error: response.retcode !== 0 ? response.message : undefined,
    };
  }

  async pauseProvider(providerId: string): Promise<{success: boolean; error?: string}> {
    const response = await rpcClient.call<{id: string}, {success: boolean}>(
      'RPCPauseProvider',
      { id: providerId },
      'meta'
    );

    return {
      success: response.retcode === 0,
      error: response.retcode !== 0 ? response.message : undefined,
    };
  }

  async resumeProvider(providerId: string): Promise<{success: boolean; error?: string}> {
    const response = await rpcClient.call<{id: string}, {success: boolean}>(
      'RPCResumeProvider',
      { id: providerId },
      'meta'
    );

    return {
      success: response.retcode === 0,
      error: response.retcode !== 0 ? response.message : undefined,
    };
  }

  async stopProvider(providerId: string): Promise<{success: boolean; error?: string}> {
    const response = await rpcClient.call<{id: string}, {success: boolean}>(
      'RPCStopProvider',
      { id: providerId },
      'meta'
    );

    return {
      success: response.retcode === 0,
      error: response.retcode !== 0 ? response.message : undefined,
    };
  }

  async waitForState(
    providerId: string,
    targetState: string,
    maxAttempts: number = 10,
    delayMs: number = 1000
  ): Promise<ProviderState> {
    for (let i = 0; i < maxAttempts; i++) {
      const state = await this.getProviderState(providerId);
      if (state.state === targetState) {
        return state;
      }
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    }
    return this.getProviderState(providerId);
  }

  async startAllProviders(): Promise<{success: boolean; started: string[]; failed: string[]; total: number}> {
    const response = await rpcClient.call<{}, {success: boolean; started: string[]; failed: string[]; total: number}>(
      'RPCFSMStartAllProviders',
      {},
      'meta'
    );

    return {
      success: response.retcode === 0 && (response.payload?.success ?? false),
      started: response.payload?.started || [],
      failed: response.payload?.failed || [],
      total: response.payload?.total || 0,
    };
  }

  async stopAllProviders(): Promise<{success: boolean; stopped: string[]; failed: string[]; total: number}> {
    const response = await rpcClient.call<{}, {success: boolean; stopped: string[]; failed: string[]; total: number}>(
      'RPCFSMStopAllProviders',
      {},
      'meta'
    );

    return {
      success: response.retcode === 0 && (response.payload?.success ?? false),
      stopped: response.payload?.stopped || [],
      failed: response.payload?.failed || [],
      total: response.payload?.total || 0,
    };
  }

  async pauseAllProviders(): Promise<{success: boolean; paused: string[]; failed: string[]; total: number}> {
    const response = await rpcClient.call<{}, {success: boolean; paused: string[]; failed: string[]; total: number}>(
      'RPCFSMPauseAllProviders',
      {},
      'meta'
    );

    return {
      success: response.retcode === 0 && (response.payload?.success ?? false),
      paused: response.payload?.paused || [],
      failed: response.payload?.failed || [],
      total: response.payload?.total || 0,
    };
  }

  async resumeAllProviders(): Promise<{success: boolean; resumed: string[]; failed: string[]; total: number}> {
    const response = await rpcClient.call<{}, {success: boolean; resumed: string[]; failed: string[]; total: number}>(
      'RPCFSMResumeAllProviders',
      {},
      'meta'
    );

    return {
      success: response.retcode === 0 && (response.payload?.success ?? false),
      resumed: response.payload?.resumed || [],
      failed: response.payload?.failed || [],
      total: response.payload?.total || 0,
    };
  }

  async getProviderList(): Promise<{providers: Array<{id: string; type: string; state: string}>; count: number}> {
    const response = await rpcClient.call<{}, {
      tree?: {
        providers?: Array<{id: string; type: string; state: string}>;
      };
      providers?: Array<{id: string; type: string; state: string}>;
    }>('RPCGetEtlTree', {}, 'meta');

    if (response.retcode !== 0) {
      return { providers: [], count: 0 };
    }

    // Handle wrapped format (v2access) vs direct format
    const providers = response.payload?.providers || response.payload?.tree?.providers || [];
    return {
      providers: providers.map(p => ({
        id: p.id,
        type: p.type,
        state: p.state,
      })),
      count: providers.length,
    };
  }

  async getEtlTree(): Promise<{
    macro_fsm: { state: string; total_providers: number; active_providers: number };
    providers: Array<{
      id: string;
      type: string;
      state: string;
      processed_count?: number;
      error_message?: string;
    }>;
  }> {
    // The API returns { tree: { macro, providers, ... } }
    // Need to handle both response formats
    const response = await rpcClient.call<{}, {
      tree?: {
        macro?: { state: string; total_providers?: number; active_providers?: number };
        providers?: Array<{
          id: string;
          type: string;
          state: string;
          processed_count?: number;
          error_message?: string;
        }>;
      };
      macro_fsm?: { state: string; total_providers?: number; active_providers?: number };
      providers?: Array<{
        id: string;
        type: string;
        state: string;
        processed_count?: number;
        error_message?: string;
      }>;
    }>('RPCGetEtlTree', {}, 'meta');

    if (response.retcode !== 0) {
      throw new Error('Failed to get ETL tree');
    }

    // Handle wrapped format (v2access) vs direct format (direct RPC)
    const payload = response.payload;
    let macro_fsm = payload?.macro_fsm || payload?.tree?.macro;
    let providers = payload?.providers || payload?.tree?.providers || [];

    // Convert snake_case to camelCase
    if (macro_fsm) {
      macro_fsm = {
        state: macro_fsm.state || 'UNKNOWN',
        total_providers: macro_fsm.total_providers || 0,
        active_providers: macro_fsm.active_providers || 0,
      };
    }

    return {
      macro_fsm: macro_fsm || { state: 'UNKNOWN', total_providers: 0, active_providers: 0 },
      providers: providers.map(p => ({
        id: p.id,
        type: p.type,
        state: p.state,
        processed_count: p.processed_count,
        error_message: p.error_message,
      })),
    };
  }
}

export const providerFSMClient = new ProviderFSMClient();
