/**
 * GLC Offline React Hooks
 *
 * React hooks for offline detection and sync status.
 */

'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import {
  getOnlineStatusDetector,
  getOperationQueue,
  getSyncManager,
  type OnlineStatus,
  type OperationQueueState,
  type SyncState,
} from './index';

// ============================================================================
// useOnlineStatus Hook
// ============================================================================

export interface UseOnlineStatusResult {
  /** Current online status */
  status: OnlineStatus;
  /** Whether currently online */
  isOnline: boolean;
  /** Whether currently offline */
  isOffline: boolean;
  /** Whether currently checking connectivity */
  isChecking: boolean;
  /** Manually check connectivity */
  checkConnectivity: () => Promise<boolean>;
}

/**
 * Hook to track online/offline status
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const { isOnline, isOffline, status } = useOnlineStatus();
 *
 *   return (
 *     <div>
 *       {isOffline && <OfflineBanner />}
 *       <p>Status: {status}</p>
 *     </div>
 *   );
 * }
 * ```
 */
export function useOnlineStatus(): UseOnlineStatusResult {
  const [status, setStatus] = useState<OnlineStatus>(() => {
    // Initialize with current status
    if (typeof navigator !== 'undefined') {
      return navigator.onLine ? 'online' : 'offline';
    }
    return 'online'; // Default to online for SSR
  });

  useEffect(() => {
    const detector = getOnlineStatusDetector();

    // Subscribe to status changes
    const unsubscribe = detector.subscribe((newStatus) => {
      setStatus(newStatus);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  const checkConnectivity = useCallback(async () => {
    const detector = getOnlineStatusDetector();
    return detector.checkConnectivity();
  }, []);

  return useMemo(
    () => ({
      status,
      isOnline: status === 'online',
      isOffline: status === 'offline',
      isChecking: status === 'checking',
      checkConnectivity,
    }),
    [status, checkConnectivity]
  );
}

// ============================================================================
// useOperationQueue Hook
// ============================================================================

export interface UseOperationQueueResult {
  /** Current queue state */
  state: OperationQueueState;
  /** Number of pending operations */
  pendingCount: number;
  /** Whether there are pending operations */
  hasPending: boolean;
  /** Pending graph IDs */
  pendingGraphIds: string[];
}

/**
 * Hook to track operation queue state
 *
 * @example
 * ```tsx
 * function QueueIndicator() {
 *   const { pendingCount, hasPending } = useOperationQueue();
 *
 *   if (!hasPending) return null;
 *
 *   return <Badge>{pendingCount} pending</Badge>;
 * }
 * ```
 */
export function useOperationQueue(graphId?: string): UseOperationQueueResult {
  const [state, setState] = useState<OperationQueueState>(() => {
    const queue = getOperationQueue();
    return queue.state;
  });

  useEffect(() => {
    const queue = getOperationQueue();

    const unsubscribe = queue.subscribe((newState) => {
      setState(newState);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  return useMemo(() => {
    const pendingCount = graphId
      ? state.operations.filter((op) => op.graphId === graphId).length
      : state.count;

    const hasPending = pendingCount > 0;

    const pendingGraphIds = Array.from(state.pendingGraphIds);

    return {
      state,
      pendingCount,
      hasPending,
      pendingGraphIds,
    };
  }, [state, graphId]);
}

// ============================================================================
// useSyncStatus Hook
// ============================================================================

export interface UseSyncStatusResult {
  /** Current sync state */
  state: SyncState;
  /** Whether currently syncing */
  isSyncing: boolean;
  /** Whether synced successfully */
  isSynced: boolean;
  /** Whether there's an error */
  hasError: boolean;
  /** Whether there's a conflict */
  hasConflict: boolean;
  /** Trigger manual sync */
  sync: (graphId?: string) => Promise<SyncState>;
  /** Resolve conflict */
  resolveConflict: (
    resolution: 'client' | 'server'
  ) => Promise<boolean>;
}

/**
 * Hook to track sync status and trigger sync
 *
 * @example
 * ```tsx
 * function SyncButton() {
 *   const { isSyncing, hasPending, sync } = useSyncStatus();
 *
 *   return (
 *     <Button onClick={() => sync()} disabled={isSyncing}>
 *       {isSyncing ? 'Syncing...' : 'Sync Now'}
 *     </Button>
 *   );
 * }
 * ```
 */
export function useSyncStatus(): UseSyncStatusResult {
  const [state, setState] = useState<SyncState>(() => {
    const syncManager = getSyncManager();
    return syncManager.state;
  });

  useEffect(() => {
    const syncManager = getSyncManager();

    const unsubscribe = syncManager.subscribe((newState) => {
      setState(newState);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  const sync = useCallback(async (graphId?: string) => {
    const syncManager = getSyncManager();
    return syncManager.sync(graphId);
  }, []);

  const resolveConflict = useCallback(
    async (resolution: 'client' | 'server') => {
      const syncManager = getSyncManager();
      if (state.conflict) {
        return syncManager.resolveConflict(state.conflict, resolution);
      }
      return false;
    },
    [state.conflict]
  );

  return useMemo(
    () => ({
      state,
      isSyncing: state.status === 'syncing',
      isSynced: state.status === 'synced',
      hasError: state.status === 'error',
      hasConflict: state.status === 'conflict',
      sync,
      resolveConflict,
    }),
    [state, sync, resolveConflict]
  );
}

// ============================================================================
// useOfflineIndicator Hook
// ============================================================================

export interface UseOfflineIndicatorResult {
  /** Whether currently offline */
  isOffline: boolean;
  /** Whether currently online */
  isOnline: boolean;
  /** Number of pending operations */
  pendingCount: number;
  /** Whether there are pending operations */
  hasPending: boolean;
  /** Whether currently syncing */
  isSyncing: boolean;
  /** Whether synced successfully */
  isSynced: boolean;
  /** Whether there's a sync error */
  hasSyncError: boolean;
  /** Sync progress (0-100) */
  syncProgress: number;
  /** Current sync status */
  syncStatus: SyncState['status'];
  /** Trigger manual sync */
  sync: () => Promise<void>;
}

/**
 * Combined hook for offline indicator component
 *
 * @example
 * ```tsx
 * function OfflineIndicator() {
 *   const {
 *     isOffline,
 *     pendingCount,
 *     isSyncing,
 *     syncProgress,
 *     sync
 *   } = useOfflineIndicator();
 *
 *   if (!isOffline && pendingCount === 0) return null;
 *
 *   return (
 *     <div>
 *       {isOffline && <span>Offline</span>}
 *       {pendingCount > 0 && (
 *         <span>{pendingCount} changes pending</span>
 *       )}
 *       {isSyncing && <Progress value={syncProgress} />}
 *     </div>
 *   );
 * }
 * ```
 */
export function useOfflineIndicator(): UseOfflineIndicatorResult {
  const { isOffline, isOnline } = useOnlineStatus();
  const { pendingCount, hasPending } = useOperationQueue();
  const {
    state: syncState,
    isSyncing,
    isSynced,
    hasError: hasSyncError,
    sync: doSync,
  } = useSyncStatus();

  const syncProgress = useMemo(() => {
    if (syncState.operationsTotal === 0) return 0;
    return Math.round(
      (syncState.operationsSynced / syncState.operationsTotal) * 100
    );
  }, [syncState.operationsSynced, syncState.operationsTotal]);

  const sync = useCallback(async () => {
    await doSync();
  }, [doSync]);

  return useMemo(
    () => ({
      isOffline,
      isOnline,
      pendingCount,
      hasPending,
      isSyncing,
      isSynced,
      hasSyncError,
      syncProgress,
      syncStatus: syncState.status,
      sync,
    }),
    [
      isOffline,
      isOnline,
      pendingCount,
      hasPending,
      isSyncing,
      isSynced,
      hasSyncError,
      syncProgress,
      syncState.status,
      sync,
    ]
  );
}
