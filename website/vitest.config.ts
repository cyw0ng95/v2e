import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  test: {
    // Test file patterns
    include: ['__tests__/**/*.test.ts'],

    // Exclude node_modules
    exclude: ['**/node_modules/**', '**/dist/**', '**/out/**'],

    // Default test timeout
    testTimeout: 10_000,

    // Environment
    environment: 'node',

    // reporters
    reporters: ['verbose'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './'),
    },
  },
});
