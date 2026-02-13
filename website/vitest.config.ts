import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  test: {
    // Test file patterns - include both .ts and .tsx files
    include: ['__tests__/**/*.test.ts', '__tests__/**/*.test.tsx'],

    // Exclude node_modules
    exclude: ['**/node_modules/**', '**/dist/**', '**/out/**'],

    // Default test timeout
    testTimeout: 10_000,

    // Setup files
    setupFiles: ['./vitest.setup.ts'],

    // Globals - make vi, describe, it, expect globally available
    globals: true,

    // Environment - use jsdom for component tests
    environment: 'jsdom',

    // reporters
    reporters: ['verbose'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './'),
    },
  },
});
