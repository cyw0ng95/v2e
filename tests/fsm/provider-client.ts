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
}

export const providerFSMClient = new ProviderFSMClient();
