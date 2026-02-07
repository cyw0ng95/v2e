import { spawn, ChildProcess } from 'node:child_process';
import { realpath, mkdir, rm } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { join } from 'node:path';
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
    let readyServices: string[] = [];

    while (Date.now() - startTime < timeout) {
      // Check UDS sockets
      readyServices = [];
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
