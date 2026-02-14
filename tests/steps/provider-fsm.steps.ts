import { Given, When, Then, Before } from '@cucumber/cucumber';
import { providerFSMClient, ProviderState } from '../fsm/provider-client.js';

const providerStates = new Map<string, string>();

Before(function () {
  providerStates.clear();
});

Given('the provider {string} is in {string} state', function (providerId: string, state: string) {
  providerStates.set(providerId, state);
});

When('StartProvider is called with provider ID {string}', async function (providerId: string) {
  const result = await providerFSMClient.startProvider(providerId);
  this.lastResult = result;
  if (result.success) {
    providerStates.set(providerId, 'ACQUIRING');
  }
});

When('PauseProvider is called with provider ID {string}', async function (providerId: string) {
  const result = await providerFSMClient.pauseProvider(providerId);
  this.lastResult = result;
  if (result.success) {
    providerStates.set(providerId, 'PAUSED');
  }
});

When('ResumeProvider is called with provider ID {string}', async function (providerId: string) {
  const result = await providerFSMClient.resumeProvider(providerId);
  this.lastResult = result;
  if (result.success) {
    providerStates.set(providerId, 'RUNNING');
  }
});

When('StopProvider is called with provider ID {string}', async function (providerId: string) {
  const result = await providerFSMClient.stopProvider(providerId);
  this.lastResult = result;
  if (result.success) {
    providerStates.set(providerId, 'TERMINATED');
  }
});

Then('the provider {string} should transition to {string} state', function (providerId: string, state: string) {
  const currentState = providerStates.get(providerId);
  if (currentState !== state) {
    throw new Error(`Expected state ${state}, but got ${currentState}`);
  }
});

Then('eventually the provider {string} should be in {string} state', async function (providerId: string, state: string) {
  const finalState = await providerFSMClient.waitForState(providerId, state);
  if (finalState.state !== state) {
    throw new Error(`Expected state ${state}, but got ${finalState.state}`);
  }
});

Then('the provider {string} should be in {string} state', async function (providerId: string, state: string) {
  const currentState = providerFSMClient.getProviderState(providerId);
  const actualState = (await currentState).state;
  if (actualState !== state) {
    throw new Error(`Expected state ${state}, but got ${actualState}`);
  }
});

Then('an error {string} should be returned', function (expectedError: string) {
  if (!this.lastResult) {
    throw new Error('No result available');
  }
  if (!this.lastResult.error) {
    throw new Error(`Expected error "${expectedError}", but no error was returned`);
  }
  if (!this.lastResult.error.toLowerCase().includes(expectedError.toLowerCase())) {
    throw new Error(`Expected error containing "${expectedError}", but got "${this.lastResult.error}"`);
  }
});

Then('the provider {string} state should remain {string}', function (providerId: string, state: string) {
  const currentState = providerStates.get(providerId);
  if (currentState !== state) {
    throw new Error(`Expected state to remain ${state}, but got ${currentState}`);
  }
});

Given('the provider FSM service is available', function () {
  // This is a placeholder - in a real test, we'd verify the service is running
});
