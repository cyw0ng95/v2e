import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { providerFSMClient, ProviderState } from '../fsm/provider-client.js';

const TEST_PROVIDER_ID = 'test-cve-provider';
const POLL_INTERVAL = 1000;
const MAX_ATTEMPTS = 10;

describe('ProviderFSM State Transitions', () => {
  beforeAll(async () => {
    // Wait for services to be ready
    await new Promise(resolve => setTimeout(resolve, 2000));
  });

  describe('Start Provider', () => {
    it('should transition from IDLE to RUNNING', async () => {
      // Get initial state
      const initialState = await providerFSMClient.getProviderState(TEST_PROVIDER_ID);
      
      // Start the provider
      const result = await providerFSMClient.startProvider(TEST_PROVIDER_ID);
      
      // Should succeed
      expect(result.success).toBe(true);
      
      // Wait for running state
      const finalState = await providerFSMClient.waitForState(
        TEST_PROVIDER_ID,
        'RUNNING',
        MAX_ATTEMPTS,
        POLL_INTERVAL
      );
      
      expect(finalState.state).toBe('RUNNING');
    });
  });

  describe('Pause Provider', () => {
    it('should transition from RUNNING to PAUSED', async () => {
      // First ensure provider is running
      await providerFSMClient.startProvider(TEST_PROVIDER_ID);
      await providerFSMClient.waitForState(TEST_PROVIDER_ID, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
      
      // Pause the provider
      const result = await providerFSMClient.pauseProvider(TEST_PROVIDER_ID);
      expect(result.success).toBe(true);
      
      // Check state
      const state = await providerFSMClient.getProviderState(TEST_PROVIDER_ID);
      expect(state.state).toBe('PAUSED');
    });
  });

  describe('Resume Provider', () => {
    it('should transition from PAUSED to RUNNING', async () => {
      // First ensure provider is paused
      await providerFSMClient.pauseProvider(TEST_PROVIDER_ID);
      await providerFSMClient.waitForState(TEST_PROVIDER_ID, 'PAUSED', MAX_ATTEMPTS, POLL_INTERVAL);
      
      // Resume the provider
      const result = await providerFSMClient.resumeProvider(TEST_PROVIDER_ID);
      expect(result.success).toBe(true);
      
      // Wait for running state
      const state = await providerFSMClient.waitForState(
        TEST_PROVIDER_ID,
        'RUNNING',
        MAX_ATTEMPTS,
        POLL_INTERVAL
      );
      expect(state.state).toBe('RUNNING');
    });
  });

  describe('Stop Provider', () => {
    it('should transition from RUNNING to TERMINATED', async () => {
      // First ensure provider is running
      await providerFSMClient.startProvider(TEST_PROVIDER_ID);
      await providerFSMClient.waitForState(TEST_PROVIDER_ID, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
      
      // Stop the provider
      const result = await providerFSMClient.stopProvider(TEST_PROVIDER_ID);
      expect(result.success).toBe(true);
      
      // Check state
      const state = await providerFSMClient.getProviderState(TEST_PROVIDER_ID);
      expect(state.state).toBe('TERMINATED');
    });
  });
});

describe('ProviderFSM Error Conditions', () => {
  it('should fail when starting from RUNNING state', async () => {
    // Ensure provider is running
    await providerFSMClient.startProvider(TEST_PROVIDER_ID);
    await providerFSMClient.waitForState(TEST_PROVIDER_ID, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
    
    // Try to start again - should fail
    const result = await providerFSMClient.startProvider(TEST_PROVIDER_ID);
    expect(result.success).toBe(false);
    expect(result.error).toBeDefined();
  });

  it('should fail when pausing from IDLE state', async () => {
    // Ensure provider is stopped
    await providerFSMClient.stopProvider(TEST_PROVIDER_ID);
    await providerFSMClient.waitForState(TEST_PROVIDER_ID, 'TERMINATED', MAX_ATTEMPTS, POLL_INTERVAL);
    
    // Try to pause from TERMINATED - should fail
    const result = await providerFSMClient.pauseProvider(TEST_PROVIDER_ID);
    expect(result.success).toBe(false);
    expect(result.error).toBeDefined();
  });

  it('should fail when resuming from RUNNING state', async () => {
    // Ensure provider is running
    await providerFSMClient.startProvider(TEST_PROVIDER_ID);
    await providerFSMClient.waitForState(TEST_PROVIDER_ID, 'RUNNING', MAX_ATTEMPTS, POLL_INTERVAL);
    
    // Try to resume from RUNNING - should fail
    const result = await providerFSMClient.resumeProvider(TEST_PROVIDER_ID);
    expect(result.success).toBe(false);
    expect(result.error).toBeDefined();
  });
});
