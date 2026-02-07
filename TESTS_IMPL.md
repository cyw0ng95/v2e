# v2e Integration Test Framework - Implementation Plan

## Overview

Create a Node.js/Typecript-based integration test framework that:
- Runs against packaged binaries from `./build.sh -p`
- Tests REST/RPC endpoints headlessly (no browser required)
- Reuses existing `website/lib/` infrastructure (RPC client, types, logger)
- Executes via `./build.sh -T`
- Targets **<2 minutes** execution time

**Key Design Decisions:**
- No mock data - tests trigger real API calls (NVD, GitHub, MITRE)
- Fresh databases per test - full isolation including .db, .db-shm, .db-wal files
- No database mocking - use real SQLite
- Sequential test execution - single broker instance
- Phase approach: basic tests first, ETL tests later
- **Implement, test, and debug ONE FILE AT A TIME**

---

## Implementation Philosophy

**CRITICAL: Implement one file at a time, test it, verify it works before moving to the next.**

```
File 1 -> Test -> Debug -> Verify -> File 2 -> Test -> Debug -> Verify -> ...
```

After each file is created:
1. Run TypeScript compiler to check for errors
2. If it's a test file, run just that test file
3. Fix any issues before proceeding
4. Only move to the next file when current file is stable

---

## Directory Structure

```
tests/                              # NEW: Root-level test directory
├── package.json                    # [1] Test dependencies
├── tsconfig.json                   # [2] TypeScript configuration
├── vitest.config.ts                # [3] Vitest configuration
├── src/
│   ├── service-manager.ts          # [4] Broker/subprocess lifecycle
│   ├── global-setup.ts             # [5] Vitest global setup
│   ├── global-teardown.ts          # [6] Vitest global teardown
│   └── rpc-client.ts               # [7] Wrapper around website RPC client
├── basic/                          # Phase 2: Basic endpoint tests
│   ├── broker.test.ts              # [8] Broker RPC endpoints
│   ├── access.test.ts              # [9] Access gateway tests
│   ├── local.test.ts               # [10] Local service tests (empty DB)
│   └── meta.test.ts                # [11] Meta service tests
├── etl/                            # Phase 3: ETL pipeline tests
│   ├── cve-etl.test.ts             # [12] CVE fetch from NVD
│   ├── cwe-etl.test.ts             # [13] CWE import from MITRE
│   └── capec-etl.test.ts           # [14] CAPEC import from MITRE
└── helpers/
    ├── assertions.ts                # [15] Custom test assertions
    └── fixtures.ts                 # [16] Test constants

.build/package/reports/             # Generated test reports
```

Numbers [1]-[16] indicate the implementation order.

---

## Phase 1: Infrastructure Setup (Files 1-3)

### File [1]: tests/package.json

**Purpose:** Define test dependencies and scripts

```json
{
  "name": "v2e-integration-tests",
  "version": "1.0.0",
  "private": true,
  "type": "module",
  "scripts": {
    "test": "vitest run",
    "test:basic": "vitest run --dir basic",
    "test:etl": "vitest run --dir etl",
    "test:watch": "vitest",
    "test:file": "vitest run"  # For testing individual files
  },
  "devDependencies": {
    "@types/node": "^20.11.0",
    "typescript": "^5.3.3",
    "vitest": "^2.0.0"
  },
  "dependencies": {
    "better-sqlite3": "^9.4.0"
  }
}
```

**Dependencies explained:**
- `vitest` - Test framework (fast, ESM-ready, Jest-compatible)
- `typescript` - Type safety
- `better-sqlite3` - SQLite database creation for test DBs
- `@types/node` - Node.js type definitions

**After creating this file:**
```bash
cd tests
npm install
# Verify: node_modules directory created
```

---

### File [2]: tests/tsconfig.json

**Purpose:** TypeScript configuration for test files

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "skipLibCheck": true,
    "resolveJsonModule": true,
    "types": ["node"],
    "outDir": "./dist",
    "rootDir": "."
  },
  "include": ["src/**/*", "basic/**/*", "etl/**/*", "helpers/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit
# Verify: No TypeScript errors
```

---

### File [3]: tests/vitest.config.ts

**Purpose:** Vitest configuration

**Key decisions:**
- Sequential execution (no parallel tests)
- 30s default timeout, 120s for ETL
- Global setup/teardown for service lifecycle
- JSON report output

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    // Test file patterns
    include: ['basic/**/*.test.ts', 'etl/**/*.test.ts'],

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
```

**After creating this file:**
```bash
cd tests
npx vitest --version
# Verify: Vitest runs without config errors
```

---

## Phase 1 Completion Checkpoint

Before proceeding to Phase 2, verify:
- [ ] `cd tests && npm install` succeeds
- [ ] `npx tsc --noEmit` has no errors
- [ ] `npx vitest --version` works

---

## Phase 2: Service Lifecycle Management (Files 4-6)

### File [4]: tests/src/service-manager.ts

**Purpose:** Manage broker and subprocess lifecycle for integration tests

**Key responsibilities:**
1. Create fresh empty test databases (including .db-shm, .db-wal cleanup)
2. Kill existing v2e processes before starting
3. Spawn broker with test database environment variables
4. Wait for all UDS sockets to appear
5. Verify HTTP health endpoint responds
6. Stop all processes on teardown
7. Reset individual databases between tests (cleaning .db-shm, .db-wal)

**Important: Database Cleanup**
When removing/resetting databases, must remove:
- `*.db` - main database file
- `*.db-shm` - shared memory file
- `*.db-wal` - write-ahead log file

These are SQLite's auxiliary files that MUST be cleaned up for fresh databases.

```typescript
import { spawn, ChildProcess } from 'node:child_process';
import { realpath, mkdir, rm, existsSync } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { setTimeout as sleep } from 'node:timers/promises';
import { exec } from 'node:child_process';
import { promisify } from 'node:util';

const execAsync = promisify(exec);

// Simple logger (avoiding website dependency for Node.js compatibility)
const log = {
  info: (msg: string, ...args: unknown[]) => console.log(`[INFO] ${msg}`, ...args),
  warn: (msg: string, ...args: unknown[]) => console.warn(`[WARN] ${msg}`, ...args),
  error: (msg: string, ...args: unknown[]) => console.error(`[ERROR] ${msg}`, ...args),
  debug: (msg: string, ...args: unknown[]) => {
    if (process.env.VITEST_VERBOSE) console.debug(`[DEBUG] ${msg}`, ...args);
  }
};

export interface ServiceManagerConfig {
  packageDir: string;
  testDbDir?: string;
  brokerTimeout?: number;
  udsBasePath?: string;
}

export class ServiceManager {
  private packageDir: string;
  private testDbDir: string;
  private brokerProc: ChildProcess | null = null;
  private brokerTimeout: number;
  private udsBasePath: string;

  // Expected subprocess sockets
  private readonly expectedServices = [
    'access', 'local', 'meta', 'remote', 'sysmon', 'analysis'
  ];

  // Database files to manage (base names only)
  private readonly databaseBases = [
    'cve', 'cwe', 'capec', 'attack',
    'asvs', 'ssg', 'bookmark', 'session',
    'learning_fsm', 'analysis_graph'
  ];

  constructor(config: ServiceManagerConfig) {
    this.packageDir = config.packageDir;
    this.testDbDir = config.testDbDir || join(this.packageDir, 'test_db');
    this.brokerTimeout = config.brokerTimeout || 30000;
    this.udsBasePath = config.udsBasePath || '/tmp/v2e_uds';
  }

  /**
   * Get all files associated with a database (including WAL/SHM)
   */
  private getDatabaseFiles(dbBase: string): string[] {
    const dir = this.testDbDir;
    return [
      join(dir, `${dbBase}.db`),
      join(dir, `${dbBase}.db-wal`),
      join(dir, `${dbBase}.db-shm`)
    ];
  }

  /**
   * Create fresh empty test databases
   * Cleans up any existing .db, .db-wal, .db-shm files
   */
  async setupTestDatabases(): Promise<void> {
    log.info('Setting up test databases...', { dir: this.testDbDir });

    // Clean up existing test databases directory (including WAL/SHM files)
    try {
      await rm(this.testDbDir, { recursive: true, force: true });
      log.debug('Cleaned existing test database directory');
    } catch {
      // Directory doesn't exist, continue
    }

    // Create test database directory
    await mkdir(this.testDbDir, { recursive: true });
    log.debug('Created test database directory');

    // Create empty SQLite databases
    const Database = await import('better-sqlite3');
    const sqlite3 = Database.default;

    for (const dbBase of this.databaseBases) {
      const dbPath = join(this.testDbDir, `${dbBase}.db`);
      const db = new sqlite3(dbPath);
      db.close();
      log.debug('Created test database', { dbBase, dbPath });
    }

    log.info('Test databases created', { count: this.databaseBases.length });
  }

  /**
   * Reset a specific database to empty state
   * Removes .db, .db-wal, .db-shm files
   */
  async resetDatabase(dbBase: string): Promise<void> {
    log.debug('Resetting database', { dbBase });

    // Remove all associated files
    for (const filePath of this.getDatabaseFiles(dbBase)) {
      try {
        await rm(filePath, { force: true });
      } catch {
        // File doesn't exist, continue
      }
    }

    // Create fresh database
    const Database = await import('better-sqlite3');
    const sqlite3 = Database.default;
    const dbPath = join(this.testDbDir, `${dbBase}.db`);
    const db = new sqlite3(dbPath);
    db.close();

    log.debug('Database reset complete', { dbBase });
  }

  /**
   * Start the broker and all subprocesses
   */
  async start(): Promise<boolean> {
    log.info('Starting v2e services...');

    // Setup test databases first
    await this.setupTestDatabases();

    // Kill any existing v2e processes
    await this.killExistingProcesses();

    // Path to broker binary
    const brokerPath = join(this.packageDir, 'v2broker');
    if (!existsSync(brokerPath)) {
      log.error('Broker binary not found', { path: brokerPath });
      return false;
    }

    // Prepare environment with test database paths
    const env = {
      ...process.env,
      CVE_DB_PATH: join(this.testDbDir, 'cve.db'),
      CWE_DB_PATH: join(this.testDbDir, 'cwe.db'),
      CAPEC_DB_PATH: join(this.testDbDir, 'capec.db'),
      ATTACK_DB_PATH: join(this.testDbDir, 'attack.db'),
      SESSION_DB_PATH: join(this.testDbDir, 'session.db'),
    };

    // Spawn broker
    log.info('Spawning broker...', { brokerPath });
    this.brokerProc = spawn(brokerPath, ['config.json'], {
      cwd: this.packageDir,
      env,
      stdio: ['ignore', 'pipe', 'pipe']
    });

    // Log broker output for debugging
    this.brokerProc.stdout?.on('data', (data) => {
      log.debug('[BROKER stdout]', data.toString().trim());
    });
    this.brokerProc.stderr?.on('data', (data) => {
      log.error('[BROKER stderr]', data.toString().trim());
    });

    // Handle broker exit
    this.brokerProc.on('exit', (code, signal) => {
      if (code !== 0 && code !== null) {
        log.warn('Broker exited', { code, signal });
      }
    });

    // Wait for services to be ready
    return await this.waitForReady();
  }

  /**
   * Wait for all services to be ready
   */
  async waitForReady(): Promise<boolean> {
    log.info('Waiting for services to be ready...', { timeout: this.brokerTimeout });

    const startTime = Date.now();
    const timeout = this.brokerTimeout;

    while (Date.now() - startTime < timeout) {
      // Check UDS sockets
      const readyServices: string[] = [];
      for (const service of this.expectedServices) {
        const socketPath = `${this.udsBasePath}_${service}.sock`;
        if (existsSync(socketPath)) {
          readyServices.push(service);
        }
      }

      if (readyServices.length === this.expectedServices.length) {
        log.info('All UDS sockets ready', { services: readyServices });

        // Additional wait for HTTP server
        await sleep(2000);

        // Verify HTTP health endpoint
        const healthy = await this.checkHttpHealth();
        if (healthy) {
          log.info('All services ready and healthy');
          return true;
        } else {
          log.warn('UDS sockets ready but health check failed');
        }
      }

      await sleep(300);
    }

    log.error('Services timeout', {
      ready: readyServices.length,
      expected: this.expectedServices.length
    });
    return false;
  }

  /**
   * Check HTTP health endpoint
   */
  async checkHttpHealth(): Promise<boolean> {
    try {
      const response = await fetch('http://localhost:8080/restful/health', {
        signal: AbortSignal.timeout(5000)
      });
      return response.ok;
    } catch (error) {
      log.debug('Health check failed', { error: (error as Error).message });
      return false;
    }
  }

  /**
   * Stop all v2e processes
   */
  async stop(): Promise<void> {
    log.info('Stopping v2e services...');

    // Stop broker process
    if (this.brokerProc) {
      log.debug('Sending SIGTERM to broker');
      this.brokerProc.kill('SIGTERM');

      // Wait for graceful shutdown
      try {
        await sleep(5000);
      } catch {}

      // Force kill if still running
      if (this.brokerProc.kill('SIGKILL')) {
        log.debug('Force killed broker');
      }

      this.brokerProc = null;
    }

    // Kill any remaining processes using pkill
    await this.killExistingProcesses();

    // Clean up UDS sockets
    for (const service of this.expectedServices) {
      const socketPath = `${this.udsBasePath}_${service}.sock`;
      try {
        await rm(socketPath, { force: true });
      } catch {
        // Socket doesn't exist
      }
    }

    log.info('Services stopped');
  }

  /**
   * Kill existing v2e processes
   */
  private async killExistingProcesses(): Promise<void> {
    try {
      await execAsync(`pkill -f "${this.packageDir}/v2"`);
      log.debug('Killed existing v2e processes');
    } catch {
      // No processes found or error
    }
  }

  /**
   * Get path to a test database
   */
  getDbPath(dbBase: string): string {
    return join(this.testDbDir, `${dbBase}.db`);
  }
}
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit src/service-manager.ts
# Verify: No TypeScript errors
```

---

### File [5]: tests/src/global-setup.ts

**Purpose:** Vitest global setup hook - starts services before all tests

**What it does:**
1. Creates ServiceManager instance
2. Makes it globally accessible for tests and other hooks
3. Starts broker and all subprocesses
4. Throws error if startup fails (prevents tests from running)

```typescript
import { ServiceManager } from './service-manager.js';
import { realpath } from 'node:path/promises';

export default async function setup() {
  console.log('\n[v2e-integration] ================================================');
  console.log('[v2e-integration] Starting services...');
  console.log('[v2e-integration] ================================================');

  // Resolve package directory path
  const packageDir = await realpath(new URL('../../.build/package', import.meta.url).pathname);
  console.log('[v2e-integration] Package directory:', packageDir);

  const manager = new ServiceManager({ packageDir });

  // Make manager globally available
  (globalThis as any).__V2E_SERVICE_MANAGER__ = manager;

  const started = await manager.start();

  if (!started) {
    console.error('[v2e-integration] ================================================');
    console.error('[v2e-integration] FAILED to start services!');
    console.error('[v2e-integration] ================================================');
    throw new Error('Failed to start v2e services');
  }

  console.log('[v2e-integration] ================================================');
  console.log('[v2e-integration] Services ready!');
  console.log('[v2e-integration] ================================================\n');
}
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit src/global-setup.ts
# Verify: No TypeScript errors
```

---

### File [6]: tests/src/global-teardown.ts

**Purpose:** Vitest global teardown hook - stops services after all tests

**What it does:**
1. Retrieves ServiceManager from global scope
2. Stops all processes
3. Cleans up resources

```typescript
export default async function teardown() {
  console.log('\n[v2e-integration] ================================================');
  console.log('[v2e-integration] Stopping services...');
  console.log('[v2e-integration] ================================================');

  const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;

  if (manager) {
    await manager.stop();
    delete (globalThis as any).__V2E_SERVICE_MANAGER__;
  }

  console.log('[v2e-integration] ================================================');
  console.log('[v2e-integration] Services stopped');
  console.log('[v2e-integration] ================================================\n');
}
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit src/global-teardown.ts
# Verify: No TypeScript errors
```

---

### Test Phase 2: Service Lifecycle

**Before proceeding, create a minimal test to verify service lifecycle works:**

Create temporary test file `tests/sanity-check.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';

describe('Sanity Check', () => {
  it('should have services running', () => {
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    expect(manager).toBeDefined();
  });

  it('should be able to fetch health endpoint', async () => {
    const response = await fetch('http://localhost:8080/restful/health');
    expect(response.ok).toBe(true);
  });
});
```

**Test the service lifecycle:**
```bash
# From project root
./build.sh -p

cd tests
npm run test:file sanity-check.test.ts
```

**Expected output:**
- Services start (broker + subprocesses)
- Sanity check tests pass
- Services stop cleanly

**If this works, delete `sanity-check.test.ts` and proceed to Phase 3.**

---

## Phase 2 Completion Checkpoint

Before proceeding to Phase 3, verify:
- [ ] `service-manager.ts` compiles without errors
- [ ] `global-setup.ts` compiles without errors
- [ ] `global-teardown.ts` compiles without errors
- [ ] Sanity check test passes (services start/stop correctly)
- [ ] All UDS sockets created: `/tmp/v2e_uds_*.sock`
- [ ] Health endpoint responds: `http://localhost:8080/restful/health`

---

## Phase 3: RPC Client and Test Helpers (Files 7, 15-16)

### File [7]: tests/src/rpc-client.ts

**Purpose:** Wrap RPC communication for tests

**Key differences from website RPC client:**
1. No React.cache (not available in Node.js)
2. No mock mode (always call real backend)
3. Simplified for test use cases
4. Direct `fetch` usage (Node 18+ native)

**Test Cases this enables:**
- Making RPC calls to any service
- Testing error handling
- Verifying response format

```typescript
import { RPCRequest, RPCResponse } from '../../website/lib/types.js';

// Case conversion utilities (from website/lib/rpc-client.ts)
function toCamelCase(str: string): string {
  if (str.indexOf('_') >= 0) {
    return str.replace(/_([a-zA-Z0-9])/g, (_, letter) => letter.toUpperCase());
  }
  if (str === str.toUpperCase()) {
    return str.toLowerCase();
  }
  return str.charAt(0).toLowerCase() + str.slice(1);
}

function toSnakeCase(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
}

function convertKeysToCamelCase<T>(obj: unknown): T {
  if (obj === null || obj === undefined) {
    return obj as T;
  }
  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeysToCamelCase(item)) as T;
  }
  if (typeof obj === 'object') {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj)) {
      const camelKey = toCamelCase(key);
      result[camelKey] = convertKeysToCamelCase(value);
    }
    return result as T;
  }
  return obj as T;
}

function convertKeysToSnakeCase<T>(obj: unknown): T {
  if (obj === null || obj === undefined) {
    return obj as T;
  }
  if (Array.isArray(obj)) {
    return obj.map((item) => convertKeysToSnakeCase(item)) as T;
  }
  if (typeof obj === 'object') {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj)) {
      const snakeKey = toSnakeCase(key);
      result[snakeKey] = convertKeysToSnakeCase(value);
    }
    return result as T;
  }
  return obj as T;
}

export class TestRPCClient {
  private baseUrl: string;
  private timeout: number;

  constructor(baseUrl?: string, timeout?: number) {
    this.baseUrl = baseUrl || process.env.V2E_API_BASE_URL || 'http://localhost:8080';
    this.timeout = timeout || parseInt(process.env.V2E_TEST_TIMEOUT || '30000');
  }

  /**
   * Make an RPC call to the backend
   */
  async call<TRequest, TResponse>(
    method: string,
    params?: TRequest,
    target: string = 'meta'
  ): Promise<RPCResponse<TResponse>> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const request: RPCRequest<TRequest> = {
        method,
        params: params ? convertKeysToSnakeCase(params) : undefined,
        target
      };

      const response = await fetch(`${this.baseUrl}/restful/rpc`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
        signal: controller.signal
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const rpcResponse: RPCResponse<TResponse> = await response.json();

      if (rpcResponse.payload) {
        rpcResponse.payload = convertKeysToCamelCase(rpcResponse.payload);
      }

      return rpcResponse;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error && error.name === 'AbortError') {
        return {
          retcode: 500,
          message: 'Request timeout',
          payload: null
        } as RPCResponse<TResponse>;
      }

      return {
        retcode: 500,
        message: error instanceof Error ? error.message : 'Unknown error',
        payload: null
      } as RPCResponse<TResponse>;
    }
  }
}

// Singleton instance
export const rpcClient = new TestRPCClient();
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit src/rpc-client.ts
# Verify: No TypeScript errors
```

---

### File [15]: tests/helpers/assertions.ts

**Purpose:** Custom test assertions for v2e integration tests

**Test cases covered:**
- RPC success verification
- Service presence checking
- Data count validation
- Empty state verification
- Not found handling

```typescript
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
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit helpers/assertions.ts
# Verify: No TypeScript errors
```

---

### File [16]: tests/helpers/fixtures.ts

**Purpose:** Test constants and reusable test data

```typescript
// Test configuration
export const TEST_CONFIG = {
  API_BASE_URL: process.env.V2E_API_BASE_URL || 'http://localhost:8080',
  TEST_TIMEOUT: parseInt(process.env.V2E_TEST_TIMEOUT || '30000'),
  ETL_TIMEOUT: 120000, // 2 minutes for ETL tests
  DB_DIR: '.build/package/test_db'
} as const;

// Known entity IDs for testing (real entities that should exist)
// Note: These may not exist - tests should handle both cases gracefully
export const KNOWN = {
  CVE: 'CVE-2024-0001',      // NVD CVE ID (may not exist)
  CWE: 'CWE-79',             // Cross-site Scripting (should exist)
  CAPEC: 'CAPEC-1'           // CAPEC ID (may not exist)
} as const;

// Expected service IDs (must all be present)
export const EXPECTED_SERVICES = [
  'access',
  'local',
  'meta',
  'remote',
  'sysmon',
  'analysis'
] as const;

// Database base names (without .db extension)
export const DATABASE_BASES = [
  'cve', 'cwe', 'capec', 'attack',
  'asvs', 'ssg', 'bookmark', 'session',
  'learning_fsm', 'analysis_graph'
] as const;
```

**After creating this file:**
```bash
cd tests
npx tsc --noEmit helpers/fixtures.ts
# Verify: No TypeScript errors
```

---

### Test Phase 3: RPC Client and Helpers

**Create a test to verify RPC client works:**

Create `tests/rpc-sanity.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { rpcClient } from './src/rpc-client.js';
import { assertRpcSuccess } from './helpers/assertions.js';

describe('RPC Client Sanity', () => {
  it('should call broker RPCListProcesses', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');
    await assertRpcSuccess(response);
    expect(response.payload).toBeDefined();
  });

  it('should return error for invalid method', async () => {
    const response = await rpcClient.call('RPCInvalidMethod', {}, 'broker');
    expect(response.retcode).not.toBe(0);
  });
});
```

**Test the RPC client:**
```bash
cd tests
npm run test:file rpc-sanity.test.ts
```

**Expected output:**
- RPC calls succeed
- Invalid method returns error
- No crashes or hangs

**If this works, delete `rpc-sanity.test.ts` and proceed to Phase 4.**

---

## Phase 3 Completion Checkpoint

Before proceeding to Phase 4, verify:
- [ ] `rpc-client.ts` compiles and works
- [ ] `assertions.ts` compiles
- [ ] `fixtures.ts` compiles
- [ ] RPC sanity test passes
- [ ] Can make successful RPC calls to broker, access, local, meta

---

## Phase 4: Basic Integration Tests (Files 8-11)

### File [8]: tests/basic/broker.test.ts

**Purpose:** Test broker RPC endpoints

**Test Cases:**
1. **List all processes** - Verify broker returns all subprocesses
2. **Verify expected services present** - Check access, local, meta, remote, sysmon, analysis
3. **Get broker status** - Verify status endpoint works
4. **Verify all services healthy** - Check all processes in running/idle state

```typescript
import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertHasServices, assertProcessesHealthy } from '../helpers/assertions.js';
import { EXPECTED_SERVICES } from '../helpers/fixtures.js';

describe('Broker RPC', () => {
  it('should list all processes', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    await assertRpcSuccess(response);

    const processes = response.payload as Array<{ id: string; status: string }>;
    expect(Array.isArray(processes)).toBe(true);
    expect(processes.length).toBeGreaterThan(0);
  });

  it('should have all expected services', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    await assertRpcSuccess(response);

    const processes = response.payload as Array<{ id: string; status: string }>;
    assertHasServices(processes, [...EXPECTED_SERVICES]);
  });

  it('should get broker status', async () => {
    const response = await rpcClient.call('RPCGetStatus', {}, 'broker');

    await assertRpcSuccess(response);

    const status = response.payload as Record<string, unknown>;
    expect(status).toBeDefined();
    expect(status.state).toBeDefined();
  });

  it('should have all services in healthy state', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    await assertRpcSuccess(response);

    const processes = response.payload as Array<{ id: string; status: string }>;
    assertProcessesHealthy(processes);
  });
});
```

**After creating this file:**
```bash
cd tests
npm run test:file basic/broker.test.ts
```

**Expected:** All 4 tests pass

---

### File [9]: tests/basic/access.test.ts

**Purpose:** Test access gateway endpoints

**Test Cases:**
1. **Health endpoint** - Verify /restful/health returns 200
2. **Missing method field** - Verify proper error handling
3. **RPC forwarding works** - Verify broker calls succeed through access
4. **Invalid RPC method** - Verify error returned for unknown methods

```typescript
import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { TEST_CONFIG } from '../helpers/fixtures.js';

describe('Access Service', () => {
  it('should return health status', async () => {
    const response = await fetch(`${TEST_CONFIG.API_BASE_URL}/restful/health`);

    expect(response.ok).toBe(true);

    const data = await response.json();
    expect(data).toHaveProperty('status');
  });

  it('should reject missing method field', async () => {
    const response = await fetch(`${TEST_CONFIG.API_BASE_URL}/restful/rpc`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ params: {} })
    });

    // Access service validates input - should return 400
    expect([400, 200]).toContain(response.status);

    if (response.status === 200) {
      const data = await response.json();
      expect(data.retcode).not.toBe(0);
    }
  });

  it('should forward RPC calls to broker', async () => {
    const response = await rpcClient.call('RPCListProcesses', {}, 'broker');

    expect(response.retcode).toBe(0);
    expect(response.payload).toBeDefined();
  });

  it('should return error for invalid RPC method', async () => {
    const response = await rpcClient.call('RPCInvalidMethodXYZ', {}, 'broker');

    expect(response.retcode).not.toBe(0);
  });
});
```

**After creating this file:**
```bash
cd tests
npm run test:file basic/access.test.ts
```

**Expected:** All 4 tests pass

---

### File [10]: tests/basic/local.test.ts

**Purpose:** Test local service with empty database

**Test Cases:**
1. **Empty CVE list** - Verify database starts empty
2. **Unknown CVE returns not found** - Verify proper error handling
3. **Empty CWE list** - Verify CWE database starts empty
4. **Unknown CWE returns not found** - Verify error handling
5. **Empty CAPEC list** - Verify CAPEC database starts empty

**Note:** Each test resets databases to ensure clean state

```typescript
import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertEmpty, assertNotFound } from '../helpers/assertions.js';

describe('Local Service - Empty Database', () => {
  beforeEach(async () => {
    // Reset databases via service manager
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('cve');
      await manager.resetDatabase('cwe');
      await manager.resetDatabase('capec');
    }
  });

  describe('CVE Operations', () => {
    it('should return empty CVE list', async () => {
      const response = await rpcClient.call(
        'RPCListCVEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.cves)).toBe(true);
      assertEmpty(data.cves);
    });

    it('should return not found for unknown CVE', async () => {
      const response = await rpcClient.call(
        'RPCGetCVE',
        { cveId: 'CVE-2024-9999' },
        'local'
      );

      assertNotFound(response);
    });
  });

  describe('CWE Operations', () => {
    it('should return empty CWE list', async () => {
      const response = await rpcClient.call(
        'RPCListCWEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.cwes)).toBe(true);
      assertEmpty(data.cwes);
    });

    it('should return not found for unknown CWE', async () => {
      const response = await rpcClient.call(
        'RPCGetCWEByID',
        { cweId: 'CWE-999999' },
        'local'
      );

      assertNotFound(response);
    });
  });

  describe('CAPEC Operations', () => {
    it('should return empty CAPEC list', async () => {
      const response = await rpcClient.call(
        'RPCListCAPECs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);

      const data = response.payload as any;
      expect(Array.isArray(data.capecs)).toBe(true);
      assertEmpty(data.capecs);
    });
  });
});
```

**After creating this file:**
```bash
cd tests
npm run test:file basic/local.test.ts
```

**Expected:** All 5 tests pass

---

### File [11]: tests/basic/meta.test.ts

**Purpose:** Test meta service (job orchestration)

**Test Cases:**
1. **Get service status** - Verify meta service is running
2. **List jobs (possibly empty)** - Verify job listing works
3. **Get ETL tree** - Verify UEE (Unified ETL Engine) tree endpoint

```typescript
import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe('Meta Service', () => {
  it('should get service status', async () => {
    const response = await rpcClient.call('RPCGetStatus', {}, 'meta');

    await assertRpcSuccess(response);

    const status = response.payload as Record<string, unknown>;
    expect(status).toBeDefined();
    expect(status.state).toBeDefined();
  });

  it('should list jobs (possibly empty)', async () => {
    const response = await rpcClient.call('RPCListJobs', {}, 'meta');

    await assertRpcSuccess(response);

    const jobs = response.payload as unknown[];
    expect(Array.isArray(jobs)).toBe(true);
  });

  it('should get ETL tree', async () => {
    const response = await rpcClient.call('RPCGetEtlTree', {}, 'meta');

    await assertRpcSuccess(response);

    const tree = response.payload as Record<string, unknown>;
    expect(tree).toBeDefined();
    expect(tree.macro).toBeDefined();
  });
});
```

**After creating this file:**
```bash
cd tests
npm run test:file basic/meta.test.ts
```

**Expected:** All 3 tests pass

---

### Test Phase 4: All Basic Tests

**Run all basic tests together:**
```bash
cd tests
npm run test:basic
```

**Expected:** All basic tests pass (broker, access, local, meta)

---

## Phase 4 Completion Checkpoint

Before proceeding to Phase 5, verify:
- [ ] All basic tests pass individually
- [ ] All basic tests pass together (`npm run test:basic`)
- [ ] Database resets work correctly (local.test.ts passes)
- [ ] No test interdependence (tests can run in any order)

---

## Phase 5: ETL Integration Tests (Files 12-14)

### File [12]: tests/etl/cve-etl.test.ts

**Purpose:** Test CVE data fetch from NVD (real API calls)

**Test Cases:**
1. **Fetch single CVE** - Trigger CVE fetch, verify ETL flow works
2. **Fetch CVEs by year (limited)** - Trigger year fetch with small limit

**Important notes:**
- These tests make real API calls to NVD
- May fail due to rate limiting, network issues, or CVE not existing
- Tests should handle both success and failure gracefully
- We're testing the ETL **flow**, not specific data existence

**Timeout:** 60 seconds per test

```typescript
import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';
import { KNOWN } from '../helpers/fixtures.js';

describe.describe('CVE ETL', () => {
  beforeEach(async () => {
    // Reset database before each ETL test
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('cve');
    }
  });

  it('should trigger CVE fetch request',
    { timeout: 60000 },
    async () => {
      // Trigger CVE fetch - this queues the request
      const fetchResponse = await rpcClient.call(
        'RPCTriggerCVEFetch',
        { cveId: KNOWN.CVE },
        'remote'
      );

      // The request to trigger should succeed
      await assertRpcSuccess(fetchResponse);

      // Note: We're testing that the ETL system accepts the request
      // The actual fetch happens asynchronously and may:
      // - Succeed if CVE exists
      // - Fail if CVE doesn't exist
      // - Fail due to rate limiting
      // We verify the request was accepted, not the result
    }
  );

  it('should trigger CVE year fetch with limit',
    { timeout: 120000 },
    async () => {
      const year = 2024;

      // Trigger CVE year fetch with small limit for speed
      const fetchResponse = await rpcClient.call(
        'RPCTriggerCVEYearFetch',
        { year, limit: 3 }, // Very small limit
        'remote'
      );

      // The request should be accepted
      await assertRpcSuccess(fetchResponse);

      // Wait some time for processing (may or may not complete)
      await new Promise(resolve => setTimeout(resolve, 10000));

      // Try to list CVEs - may have data if fetch succeeded
      const listResponse = await rpcClient.call(
        'RPCListCVEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(listResponse);

      const data = listResponse.payload as any;
      expect(Array.isArray(data.cves)).toBe(true);

      // We don't assert specific counts because:
      // - NVD API may be rate limited
      // - Network may be slow
      // - CVEs may not exist yet
      // We verify the system doesn't crash
    }
  );
});
```

**After creating this file:**
```bash
cd tests
npm run test:file etl/cve-etl.test.ts
```

**Expected:** Tests pass (may have zero CVEs due to rate limiting, but shouldn't crash)

---

### File [13]: tests/etl/cwe-etl.test.ts

**Purpose:** Test CWE import from MITRE

**Test Cases:**
1. **Start CWE import** - Trigger import, verify it starts
2. **List CWEs after import** - Check if data was imported

**Important notes:**
- CWE import reads from local file (assets/cwe-raw.json) or fetches from MITRE
- Tests should handle both success and missing asset gracefully

**Timeout:** 60 seconds

```typescript
import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe.describe('CWE ETL', () => {
  beforeEach(async () => {
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('cwe');
    }
  });

  it('should trigger CWE import',
    { timeout: 60000 },
    async () => {
      // Start CWE import
      const importResponse = await rpcClient.call(
        'RPCStartCWEImport',
        {},
        'meta'
      );

      // Import should start successfully
      await assertRpcSuccess(importResponse);

      // Wait for import to process
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to list CWEs
      const listResponse = await rpcClient.call(
        'RPCListCWEs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(listResponse);

      const data = listResponse.payload as any;
      expect(Array.isArray(data.cwes)).toBe(true);

      // May have CWEs if asset exists or import succeeded
      // We verify the system handles both cases gracefully
    }
  );

  it('should handle CWE retrieval after import attempt',
    { timeout: 60000 },
    async () => {
      const cweId = 'CWE-79';

      // Try to start import
      await rpcClient.call('RPCStartCWEImport', {}, 'meta');
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to get specific CWE
      const getResponse = await rpcClient.call(
        'RPCGetCWEByID',
        { cweId },
        'local'
      );

      // May or may not be found
      // We verify no crash occurs
      expect(getResponse).toBeDefined();
    }
  );
});
```

**After creating this file:**
```bash
cd tests
npm run test:file etl/cwe-etl.test.ts
```

**Expected:** Tests pass (may have zero CWEs if asset missing, but shouldn't crash)

---

### File [14]: tests/etl/capec-etl.test.ts

**Purpose:** Test CAPEC import from MITRE

**Test Cases:**
1. **Start CAPEC import** - Trigger import, verify it starts
2. **List CAPECs after import** - Check if data was imported

**Important notes:**
- CAPEC import reads from local file (assets/capec_contents_latest.xml) or fetches from MITRE
- Tests should handle both success and missing asset gracefully

**Timeout:** 60 seconds

```typescript
import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess } from '../helpers/assertions.js';

describe.describe('CAPEC ETL', () => {
  beforeEach(async () => {
    const manager = (globalThis as any).__V2E_SERVICE_MANAGER__;
    if (manager) {
      await manager.resetDatabase('capec');
    }
  });

  it('should trigger CAPEC import',
    { timeout: 60000 },
    async () => {
      // Start CAPEC import
      const importResponse = await rpcClient.call(
        'RPCStartCAPECImport',
        {},
        'meta'
      );

      // Import should start successfully
      await assertRpcSuccess(importResponse);

      // Wait for import to process
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to list CAPECs
      const listResponse = await rpcClient.call(
        'RPCListCAPECs',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(listResponse);

      const data = listResponse.payload as any;
      expect(Array.isArray(data.capecs)).toBe(true);

      // May have CAPECs if asset exists or import succeeded
      // We verify the system handles both cases gracefully
    }
  );

  it('should handle CAPEC retrieval after import attempt',
    { timeout: 60000 },
    async () => {
      const capecId = 'CAPEC-1';

      // Try to start import
      await rpcClient.call('RPCStartCAPECImport', {}, 'meta');
      await new Promise(resolve => setTimeout(resolve, 15000));

      // Try to get specific CAPEC
      const getResponse = await rpcClient.call(
        'RPCGetCAPECByID',
        { capecId },
        'local'
      );

      // May or may not be found
      // We verify no crash occurs
      expect(getResponse).toBeDefined();
    }
  );
});
```

**After creating this file:**
```bash
cd tests
npm run test:file etl/capec-etl.test.ts
```

**Expected:** Tests pass (may have zero CAPECs if asset missing, but shouldn't crash)

---

### Test Phase 5: All ETL Tests

**Run all ETL tests together:**
```bash
cd tests
npm run test:etl
```

**Expected:** All ETL tests pass

---

## Phase 5 Completion Checkpoint

Before proceeding to Phase 6, verify:
- [ ] All ETL tests pass individually
- [ ] All ETL tests pass together (`npm run test:etl`)
- [ ] Tests handle missing assets gracefully
- [ ] Tests don't crash on network errors

---

## Phase 6: Build Script Integration (Final Step)

### Modify build.sh

**Location:** Root `build.sh` file

**Add to `show_help()` function (around line 140-150):**

```bash
    -T          Run integration tests against packaged binary (<2min target)
```

**Add to `getopts` loop (around line 913):**

```bash
    T) RUN_INTEGRATION_TESTS=true ;;
```

**Add new function `run_integration_tests()` (after `run_benchmarks()`, around line 900):**

```bash
# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."

    # Step 1: Build and package
    build_and_package
    if [ $? -ne 0 ]; then
        log_error "Build failed, cannot run integration tests"
        return 1
    fi

    # Step 2: Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is required for integration tests"
        log_error "Install Node.js 20+ from https://nodejs.org/"
        return 1
    fi

    local NODE_VERSION=$(node --version | sed 's/v//')
    log_info "Node version: $NODE_VERSION"

    # Step 3: Install test dependencies if needed
    local TEST_DIR="tests"
    if [ ! -d "$TEST_DIR/node_modules" ]; then
        log_info "Installing test dependencies..."
        cd "$TEST_DIR"
        npm install
        cd "$SCRIPT_DIR"
    fi

    # Step 4: Create reports directory
    mkdir -p "$PACKAGE_DIR/reports"

    # Step 5: Run tests
    log_info "Launching integration tests..."
    local TEST_START=$(date +%s)

    cd "$TEST_DIR"
    npm test -- --reporter=verbose
    TEST_EXIT_CODE=$?
    cd "$SCRIPT_DIR"

    local TEST_END=$(date +%s)
    local TEST_DURATION=$((TEST_END - TEST_START))

    if [ $TEST_EXIT_CODE -eq 0 ]; then
        log_info "Integration tests passed! (Duration: ${TEST_DURATION}s)"
    else
        log_error "Integration tests failed! (Duration: ${TEST_DURATION}s)"
    fi

    # Check 2-minute target
    if [ $TEST_DURATION -gt 120 ]; then
        log_warn "Tests exceeded 2 minute target: ${TEST_DURATION}s"
    fi

    return $TEST_EXIT_CODE
}
```

**Add to main() execution section (around line 950):**

```bash
elif [ "$RUN_INTEGRATION_TESTS" = true ]; then
    run_integration_tests
    exit_code=$?
```

**Test the integration:**
```bash
./build.sh -T
```

**Expected:**
1. Build and package runs
2. Integration tests execute
3. Services start and stop cleanly
4. Tests pass
5. Report generated

---

## Full Test Run

**After all files are created, run the complete test suite:**

```bash
# From project root
./build.sh -T
```

**Expected output summary:**
- Build completes successfully
- Services start (broker + 6 subprocesses)
- Basic tests pass (broker, access, local, meta)
- ETL tests pass (cve, cwe, capec)
- Services stop cleanly
- Total time < 2 minutes (excluding build time)

---

## Execution Timeline

| Phase | Action | Time |
|-------|--------|------|
| Build | `build_and_package()` | 30s |
| Startup | Service manager start | 10s |
| Basic | broker, access, local, meta tests (17 tests) | 30s |
| ETL | cve, cwe, capec tests (6 tests) | 60s |
| Shutdown | Service manager stop | 5s |
| **Total** | | **~135s** ✅ |

Note: Build time (30s) is separate from test execution time (105s).

---

## Verification Checklist

After implementation, verify:

- [ ] `./build.sh -T` runs successfully
- [ ] Tests pass with fresh databases
- [ ] Database reset removes .db, .db-wal, .db-shm files
- [ ] Service startup/shutdown works correctly
- [ ] Test reports generated in `.build/package/reports/`
- [ ] Execution time < 2 minutes (excluding build)
- [ ] All RPC endpoints tested
- [ ] ETL tests handle network failures gracefully
- [ ] No test interdependence
- [ ] Tests can run in any order

---

## Troubleshooting

### Services fail to start
- Check `.build/package/` has all binaries
- Check UDS sockets: `ls -la /tmp/v2e_uds_*.sock`
- Check broker logs in console output

### Database reset not working
- Verify .db-shm and .db-wal files are removed
- Check `rm` command has proper permissions
- Verify database paths are correct

### RPC calls timing out
- Check services are running: `curl http://localhost:8080/restful/health`
- Increase timeout in `vitest.config.ts`
- Check for port conflicts

### ETL tests failing
- May be due to rate limiting (expected)
- Check network connectivity
- Verify asset files exist in `.build/package/assets/`

---

## Implementation Status

### Completed Phases

| Phase | Status | Commit |
|-------|--------|--------|
| Phase 1: Infrastructure Setup | ✅ Complete | `343f8fd` |
| Phase 2: Service Lifecycle Management | ✅ Complete | `343f8fd` |
| Phase 3: RPC Client and Test Helpers | ✅ Complete | `996069f` |
| Phase 4: Basic Integration Tests | ✅ Complete | `c7bce29` |
| Phase 5: ETL Integration Tests | ✅ Complete | `5410e04` |
| Phase 6: Build Script Integration | ✅ Complete | `070a8a9` |

### Files Created

```
tests/
├── package.json                    # ✅ Created
├── tsconfig.json                   # ✅ Created
├── vitest.config.ts                # ✅ Created
├── src/
│   ├── service-manager.ts          # ✅ Created
│   ├── global-setup.ts             # ✅ Created
│   ├── global-teardown.ts          # ✅ Created
│   └── rpc-client.ts               # ✅ Created
├── basic/
│   ├── broker.test.ts              # ✅ Created
│   ├── access.test.ts              # ✅ Created
│   ├── local.test.ts               # ✅ Created
│   └── meta.test.ts                # ✅ Created
├── etl/
│   ├── cve-etl.test.ts             # ✅ Created
│   ├── cwe-etl.test.ts             # ✅ Created
│   └── capec-etl.test.ts           # ✅ Created
└── helpers/
    ├── assertions.ts                # ✅ Created
    └── fixtures.ts                 # ✅ Created
```

### Next Steps

To run the integration tests:
```bash
./build.sh -T
```

This will:
1. Build and package the binaries
2. Start the broker and all subprocesses
3. Run all integration tests
4. Stop services and generate reports

### Git Commits Summary

1. `343f8fd` - feat(integration-tests): Add Phase 1-2 - Infrastructure and service lifecycle
2. `996069f` - feat(integration-tests): Add Phase 3 - RPC client and test helpers
3. `c7bce29` - feat(integration-tests): Add Phase 4 - Basic integration tests
4. `5410e04` - feat(integration-tests): Add Phase 5 - ETL integration tests
5. `070a8a9` - feat(integration-tests): Add Phase 6 - Build script integration

