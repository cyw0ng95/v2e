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
