import { ServiceManager } from './service-manager.js';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

export default async function setup() {
  console.log('\n[v2e-integration] ================================================');
  console.log('[v2e-integration] Starting services...');
  console.log('[v2e-integration] ================================================');

  // Resolve package directory path
  const packageDir = resolve(dirname(fileURLToPath(import.meta.url)), '../../.build/package');
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
