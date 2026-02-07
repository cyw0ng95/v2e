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
