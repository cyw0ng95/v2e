import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { providerFSMClient } from './provider-client.js';
import { 
  resetAllProviders, 
  waitForAllProvidersState, 
  getTestProviderIds,
  ensureProvidersIdle,
  EtlTreeResponse 
} from './test-utils.js';

const POLL_INTERVAL = 1000;
const MAX_ATTEMPTS = 15;

describe('MacroFSM Batch Operations', () => {
  beforeAll(async () => {
    await new Promise(resolve => setTimeout(resolve, 2000));
    await ensureProvidersIdle();
  });

  afterAll(async () => {
    await resetAllProviders();
  });

  describe('StartAllProviders', () => {
    it('should start all providers and transition them to RUNNING', async () => {
      const result = await providerFSMClient.startAllProviders();
      
      expect(result.success).toBe(true);
      expect(result.started.length).toBeGreaterThan(0);
      expect(result.total).toBeGreaterThan(0);
      
      const reachedRunning = await waitForAllProvidersState('RUNNING', 20000);
      expect(reachedRunning).toBe(true);
    });

    it('should return empty started array when providers already running', async () => {
      await waitForAllProvidersState('RUNNING', 15000);
      
      const result = await providerFSMClient.startAllProviders();
      
      expect(result.success).toBe(true);
      expect(result.started.length).toBe(0);
    });
  });

  describe('PauseAllProviders', () => {
    it('should pause all running providers', async () => {
      await providerFSMClient.startAllProviders();
      await waitForAllProvidersState('RUNNING', 15000);
      
      const result = await providerFSMClient.pauseAllProviders();
      
      expect(result.success).toBe(true);
      expect(result.paused.length).toBeGreaterThan(0);
      
      const reachedPaused = await waitForAllProvidersState('PAUSED', 15000);
      expect(reachedPaused).toBe(true);
    });

    it('should return empty paused array when no providers running', async () => {
      await providerFSMClient.stopAllProviders();
      await waitForAllProvidersState('TERMINATED', 15000);
      
      const result = await providerFSMClient.pauseAllProviders();
      
      expect(result.paused.length).toBe(0);
    });
  });

  describe('ResumeAllProviders', () => {
    it('should resume all paused providers', async () => {
      await providerFSMClient.pauseAllProviders();
      await waitForAllProvidersState('PAUSED', 15000);
      
      const result = await providerFSMClient.resumeAllProviders();
      
      expect(result.success).toBe(true);
      expect(result.resumed.length).toBeGreaterThan(0);
      
      const reachedRunning = await waitForAllProvidersState('RUNNING', 15000);
      expect(reachedRunning).toBe(true);
    });

    it('should return empty resumed array when no providers paused', async () => {
      await providerFSMClient.startAllProviders();
      await waitForAllProvidersState('RUNNING', 15000);
      
      const result = await providerFSMClient.resumeAllProviders();
      
      expect(result.resumed.length).toBe(0);
    });
  });

  describe('StopAllProviders', () => {
    it('should stop all providers and transition to TERMINATED', async () => {
      await providerFSMClient.startAllProviders();
      await waitForAllProvidersState('RUNNING', 15000);
      
      const result = await providerFSMClient.stopAllProviders();
      
      expect(result.success).toBe(true);
      expect(result.stopped.length).toBeGreaterThan(0);
      
      const reachedTerminated = await waitForAllProvidersState('TERMINATED', 15000);
      expect(reachedTerminated).toBe(true);
    });

    it('should return empty stopped array when providers already terminated', async () => {
      await waitForAllProvidersState('TERMINATED', 15000);
      
      const result = await providerFSMClient.stopAllProviders();
      
      expect(result.stopped.length).toBe(0);
    });
  });
});

describe('MacroFSM GetEtlTree', () => {
  beforeAll(async () => {
    await new Promise(resolve => setTimeout(resolve, 2000));
    await ensureProvidersIdle();
  });

  afterAll(async () => {
    await resetAllProviders();
  });

  it('should return ETL tree with macro FSM state', async () => {
    const tree = await providerFSMClient.getEtlTree();
    
    expect(tree.macro_fsm).toBeDefined();
    expect(tree.macro_fsm.state).toBeDefined();
    expect(['BOOTSTRAPPING', 'ORCHESTRATING', 'STABILIZING', 'DRAINING', 'IDLE']).toContain(tree.macro_fsm.state);
  });

  it('should return all provider states', async () => {
    const tree = await providerFSMClient.getEtlTree();
    
    expect(tree.providers).toBeDefined();
    expect(tree.providers.length).toBeGreaterThan(0);
    
    const providerIds = tree.providers.map(p => p.id);
    const expectedProviders = getTestProviderIds();
    
    expectedProviders.forEach(expectedId => {
      expect(providerIds).toContain(expectedId);
    });
  });

  it('should return correct provider counts in macro FSM', async () => {
    await providerFSMClient.startAllProviders();
    await waitForAllProvidersState('RUNNING', 15000);
    
    const tree = await providerFSMClient.getEtlTree();
    
    expect(tree.macro_fsm.total_providers).toBeGreaterThan(0);
    expect(tree.macro_fsm.active_providers).toBeGreaterThan(0);
    expect(tree.macro_fsm.active_providers).toBeLessThanOrEqual(tree.macro_fsm.total_providers);
  });

  it('should include provider type in each provider', async () => {
    const tree = await providerFSMClient.getEtlTree();
    
    tree.providers.forEach(provider => {
      expect(provider.type).toBeDefined();
      expect(provider.state).toBeDefined();
    });
  });
});

describe('MacroFSM GetProviderList', () => {
  beforeAll(async () => {
    await new Promise(resolve => setTimeout(resolve, 2000));
    await ensureProvidersIdle();
  });

  afterAll(async () => {
    await resetAllProviders();
  });

  it('should return list of all providers', async () => {
    const result = await providerFSMClient.getProviderList();
    
    expect(result.providers).toBeDefined();
    expect(result.count).toBeGreaterThan(0);
    
    const expectedProviders = getTestProviderIds();
    const providerIds = result.providers.map(p => p.id);
    
    expectedProviders.forEach(expectedId => {
      expect(providerIds).toContain(expectedId);
    });
  });

  it('should return provider with correct structure', async () => {
    const result = await providerFSMClient.getProviderList();
    
    if (result.providers.length > 0) {
      const provider = result.providers[0];
      expect(provider.id).toBeDefined();
      expect(provider.type).toBeDefined();
      expect(provider.state).toBeDefined();
    }
  });
});

describe('MacroFSM Error Conditions', () => {
  beforeAll(async () => {
    await new Promise(resolve => setTimeout(resolve, 2000));
    await ensureProvidersIdle();
  });

  afterAll(async () => {
    await resetAllProviders();
  });

  it('should handle FSM not initialized gracefully', async () => {
    const result = await providerFSMClient.getEtlTree();
    
    expect(result).toBeDefined();
  });

  it('should handle start when already running', async () => {
    await providerFSMClient.startAllProviders();
    await waitForAllProvidersState('RUNNING', 15000);
    
    const result = await providerFSMClient.startAllProviders();
    
    expect(result.started.length).toBe(0);
  });

  it('should handle pause when not running', async () => {
    await providerFSMClient.stopAllProviders();
    await waitForAllProvidersState('TERMINATED', 15000);
    
    const result = await providerFSMClient.pauseAllProviders();
    
    expect(result.paused.length).toBe(0);
  });

  it('should handle resume when not paused', async () => {
    await providerFSMClient.startAllProviders();
    await waitForAllProvidersState('RUNNING', 15000);
    
    const result = await providerFSMClient.resumeAllProviders();
    
    expect(result.resumed.length).toBe(0);
  });

  it('should handle stop when already terminated', async () => {
    await providerFSMClient.stopAllProviders();
    await waitForAllProvidersState('TERMINATED', 15000);
    
    const result = await providerFSMClient.stopAllProviders();
    
    expect(result.stopped.length).toBe(0);
  });
});

describe('ProviderFSM Individual Operations', () => {
  const TEST_PROVIDER = 'cve';
  
  beforeAll(async () => {
    await new Promise(resolve => setTimeout(resolve, 2000));
    await ensureProvidersIdle();
  });

  afterAll(async () => {
    await resetAllProviders();
  });

  it('should start individual provider', async () => {
    const result = await providerFSMClient.startProvider(TEST_PROVIDER);
    
    expect(result.success).toBe(true);
    
    const state = await providerFSMClient.waitForState(TEST_PROVIDER, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
    expect(state.state).toBe('RUNNING');
  });

  it('should pause individual provider', async () => {
    await providerFSMClient.startProvider(TEST_PROVIDER);
    await providerFSMClient.waitForState(TEST_PROVIDER, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
    
    const result = await providerFSMClient.pauseProvider(TEST_PROVIDER);
    expect(result.success).toBe(true);
    
    const state = await providerFSMClient.waitForState(TEST_PROVIDER, 'PAUSED', MAX_ATTEMPTS, POLL_INTERVAL);
    expect(state.state).toBe('PAUSED');
  });

  it('should resume individual provider', async () => {
    await providerFSMClient.pauseProvider(TEST_PROVIDER);
    await providerFSMClient.waitForState(TEST_PROVIDER, 'PAUSED', MAX_ATTEMPTS, POLL_INTERVAL);
    
    const result = await providerFSMClient.resumeProvider(TEST_PROVIDER);
    expect(result.success).toBe(true);
    
    const state = await providerFSMClient.waitForState(TEST_PROVIDER, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
    expect(state.state).toBe('RUNNING');
  });

  it('should stop individual provider', async () => {
    const result = await providerFSMClient.stopProvider(TEST_PROVIDER);
    expect(result.success).toBe(true);
    
    const state = await providerFSMClient.waitForState(TEST_PROVIDER, 'TERMINATED', MAX_ATTEMPTS, POLL_INTERVAL);
    expect(state.state).toBe('TERMINATED');
  });
});
