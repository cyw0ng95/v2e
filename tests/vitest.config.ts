import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    // Test file patterns
    include: ['basic/**/*.test.ts', 'etl/**/*.test.ts', 'mcards/**/*.test.ts', 'fsm/**/*.test.ts', 'steps/**/*.ts'],

    // Exclude config and node_modules
    exclude: ['**/node_modules/**', '**/dist/**'],

    // Default test timeout (30s for basic tests)
    testTimeout: 30_000,
    hookTimeout: 60_000,

    // Sequential execution - no parallel tests
    threads: false,
    singleThread: true,
    fileParallelism: false,

    // Output configuration
    reporters: ['verbose', 'json'],
    outputFile: {
      json: '../.build/package/reports/test-results.json'
    },

    // Global setup/teardown
    globalSetup: './src/global-setup.ts',
    globalTeardown: './src/global-teardown.ts',

    // Environment variables
    env: {
      NODE_ENV: 'test',
      V2E_API_BASE_URL: 'http://localhost:8080',
      V2E_TEST_TIMEOUT: '30000'
    },

    // Don't suppress console output (helpful for debugging)
    silence: false
  }
});
