/**
 * GLC Crash Recovery
 *
 * Handles recovery of unsaved changes from localStorage after a crash.
 */

import type { Graph } from '../types';
import { createLogger } from '../../logger';

const logger = createLogger('glc-crash-recovery');

const CRASH_RECOVERY_KEY_PREFIX = 'glc-crash-recovery-';
// Reserved for future use: batch recovery metadata tracking
// const RECOVERY_METADATA_KEY = 'glc-recovery-metadata';

/**
 * Stored crash recovery data
 */
export interface CrashRecoveryData {
  graph: Graph;
  savedAt: number;
  version: number;
}

/**
 * Recovery metadata for tracking multiple graphs
 */
export interface RecoveryMetadata {
  graphId: string;
  savedAt: number;
  version: number;
  graphName: string;
}

/**
 * Result of crash recovery check
 */
export interface CrashRecoveryResult {
  hasRecovery: boolean;
  data: CrashRecoveryData | null;
  metadata: RecoveryMetadata | null;
  age: number; // milliseconds since save
}

/**
 * Check if there's a crash recovery available for a graph
 */
export function checkCrashRecovery(graphId: string): CrashRecoveryResult {
  if (typeof window === 'undefined') {
    return { hasRecovery: false, data: null, metadata: null, age: 0 };
  }

  const key = `${CRASH_RECOVERY_KEY_PREFIX}${graphId}`;
  const stored = localStorage.getItem(key);

  if (!stored) {
    return { hasRecovery: false, data: null, metadata: null, age: 0 };
  }

  try {
    const data = JSON.parse(stored) as CrashRecoveryData;
    const age = Date.now() - data.savedAt;

    return {
      hasRecovery: true,
      data,
      metadata: {
        graphId: data.graph.metadata.id,
        savedAt: data.savedAt,
        version: data.version,
        graphName: data.graph.metadata.name,
      },
      age,
    };
  } catch (error) {
    logger.warn('Failed to parse crash recovery data', { graphId, error });
    return { hasRecovery: false, data: null, metadata: null, age: 0 };
  }
}

/**
 * Get all available crash recoveries
 */
export function getAllCrashRecoveries(): RecoveryMetadata[] {
  if (typeof window === 'undefined') {
    return [];
  }

  const recoveries: RecoveryMetadata[] = [];

  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith(CRASH_RECOVERY_KEY_PREFIX)) {
        const stored = localStorage.getItem(key);
        if (stored) {
          try {
            const data = JSON.parse(stored) as CrashRecoveryData;
            recoveries.push({
              graphId: data.graph.metadata.id,
              savedAt: data.savedAt,
              version: data.version,
              graphName: data.graph.metadata.name,
            });
          } catch {
            // Skip invalid entries
          }
        }
      }
    }
  } catch (error) {
    logger.error('Failed to get crash recoveries', error);
  }

  return recoveries.sort((a, b) => b.savedAt - a.savedAt);
}

/**
 * Recover graph from crash recovery data
 */
export function recoverFromCrash(graphId: string): Graph | null {
  const result = checkCrashRecovery(graphId);

  if (!result.hasRecovery || !result.data) {
    return null;
  }

  logger.info('Recovered graph from crash', {
    graphId,
    version: result.data.version,
    age: result.age,
  });

  return result.data.graph;
}

/**
 * Clear crash recovery data for a specific graph
 */
export function clearCrashRecovery(graphId: string): void {
  if (typeof window === 'undefined') return;

  const key = `${CRASH_RECOVERY_KEY_PREFIX}${graphId}`;
  localStorage.removeItem(key);

  logger.debug('Cleared crash recovery data', { graphId });
}

/**
 * Clear all crash recovery data
 */
export function clearAllCrashRecoveries(): void {
  if (typeof window === 'undefined') return;

  const keysToRemove: string[] = [];

  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key?.startsWith(CRASH_RECOVERY_KEY_PREFIX)) {
      keysToRemove.push(key);
    }
  }

  keysToRemove.forEach((key) => localStorage.removeItem(key));

  logger.info('Cleared all crash recovery data', { count: keysToRemove.length });
}

/**
 * Save crash recovery data
 */
export function saveCrashRecovery(graph: Graph): void {
  if (typeof window === 'undefined') return;

  const key = `${CRASH_RECOVERY_KEY_PREFIX}${graph.metadata.id}`;
  const data: CrashRecoveryData = {
    graph,
    savedAt: Date.now(),
    version: graph.metadata.version,
  };

  try {
    localStorage.setItem(key, JSON.stringify(data));
  } catch (error) {
    // Handle quota exceeded
    if (error instanceof DOMException && error.name === 'QuotaExceededError') {
      logger.warn('localStorage quota exceeded, cleaning old recoveries');
      pruneOldRecoveries();
      try {
        localStorage.setItem(key, JSON.stringify(data));
      } catch (retryError) {
        logger.error('Failed to save crash recovery after pruning', retryError);
      }
    } else {
      logger.error('Failed to save crash recovery', error);
    }
  }
}

/**
 * Prune old recovery data to free up space
 */
export function pruneOldRecoveries(maxAge: number = 7 * 24 * 60 * 60 * 1000): void {
  if (typeof window === 'undefined') return;

  const now = Date.now();
  const keysToRemove: string[] = [];

  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key?.startsWith(CRASH_RECOVERY_KEY_PREFIX)) {
      const stored = localStorage.getItem(key);
      if (stored) {
        try {
          const data = JSON.parse(stored) as CrashRecoveryData;
          if (now - data.savedAt > maxAge) {
            keysToRemove.push(key);
          }
        } catch {
          // Remove invalid entries
          keysToRemove.push(key);
        }
      }
    }
  }

  keysToRemove.forEach((key) => localStorage.removeItem(key));

  if (keysToRemove.length > 0) {
    logger.info('Pruned old crash recovery data', { count: keysToRemove.length });
  }
}

/**
 * Format age for display
 */
export function formatRecoveryAge(ageMs: number): string {
  const seconds = Math.floor(ageMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ${hours % 24}h ago`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ago`;
  }
  if (minutes > 0) {
    return `${minutes}m ago`;
  }
  return 'just now';
}

/**
 * Hook for crash recovery in React components
 */
export function useCrashRecovery(graphId: string | null) {
  if (!graphId) {
    return { hasRecovery: false, data: null, recover: () => null, clear: () => {} };
  }

  const result = checkCrashRecovery(graphId);

  return {
    hasRecovery: result.hasRecovery,
    data: result.data,
    metadata: result.metadata,
    age: result.age,
    formattedAge: formatRecoveryAge(result.age),
    recover: () => recoverFromCrash(graphId),
    clear: () => clearCrashRecovery(graphId),
  };
}
