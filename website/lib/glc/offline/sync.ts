/**
 * GLC Offline Sync Manager
 *
 * Handles synchronization of queued operations when coming back online.
 * Implements conflict resolution and retry logic.
 */

import { toast } from 'sonner';
import { getOnlineStatusDetector, type OnlineStatus } from './detector';
import {
  getOperationQueue,
  type QueuedOperation,
  type OperationQueueState,
  createFullSyncOperation,
  operationToRPCRequest,
} from './queue';
import { useGLCStore } from '../store';
import { rpcClient } from '@/lib/rpc-client';
import type { UpdateGLCGraphRequest, UpdateGLCGraphResponse, RPCResponse } from '@/lib/types';
import {
  detectVersionConflict,
  resolveWithServerVersion,
  type VersionConflictResult,
} from '../optimistic';

// ============================================================================
// Types
// ============================================================================

export type SyncStatus =
  | 'idle'
  | 'syncing'
  | 'synced'
  | 'error'
  | 'conflict';

export interface SyncState {
  status: SyncStatus;
  operationsSynced: number;
  operationsFailed: number;
  operationsTotal: number;
  currentOperation?: string;
  error?: string;
  conflict?: VersionConflictResult;
}

export interface SyncManagerOptions {
  /** Auto-sync when coming online (default: true) */
  autoSync?: boolean;
  /** Show toast notifications (default: true) */
  showNotifications?: boolean;
  /** Delay before syncing after coming online (default: 1000ms) */
  syncDelay?: number;
}

export interface SyncListener {
  (state: SyncState): void;
}

// ============================================================================
// Constants
// ============================================================================

const DEFAULT_AUTO_SYNC = true;
const DEFAULT_SHOW_NOTIFICATIONS = true;
const DEFAULT_SYNC_DELAY = 1000;

// ============================================================================
// Sync Manager Class
// ============================================================================

/**
 * Manages synchronization of offline operations
 *
 * Usage:
 * ```ts
 * const syncManager = new SyncManager();
 *
 * // Subscribe to sync state changes
 * syncManager.subscribe((state) => {
 *   console.log('Sync status:', state.status);
 * });
 *
 * // Manually trigger sync
 * await syncManager.sync();
 *
 * // Cleanup
 * syncManager.destroy();
 * ```
 */
export class SyncManager {
  private listeners: Set<SyncListener> = new Set();
  private _state: SyncState;
  private options: Required<SyncManagerOptions>;
  private unsubscribeOnline: (() => void) | null = null;
  private unsubscribeQueue: (() => void) | null = null;
  private syncTimeout: ReturnType<typeof setTimeout> | null = null;
  private isSyncing = false;

  constructor(options?: SyncManagerOptions) {
    this.options = {
      autoSync: options?.autoSync ?? DEFAULT_AUTO_SYNC,
      showNotifications: options?.showNotifications ?? DEFAULT_SHOW_NOTIFICATIONS,
      syncDelay: options?.syncDelay ?? DEFAULT_SYNC_DELAY,
    };

    this._state = {
      status: 'idle',
      operationsSynced: 0,
      operationsFailed: 0,
      operationsTotal: 0,
    };

    // Subscribe to online status changes
    const detector = getOnlineStatusDetector();
    this.unsubscribeOnline = detector.subscribe(this.handleOnlineStatusChange.bind(this));

    // Subscribe to queue changes
    const queue = getOperationQueue();
    this.unsubscribeQueue = queue.subscribe(this.handleQueueChange.bind(this));
  }

  /**
   * Get current sync state
   */
  get state(): SyncState {
    return { ...this._state };
  }

  /**
   * Subscribe to sync state changes
   */
  subscribe(listener: SyncListener): () => void {
    this.listeners.add(listener);
    listener(this._state); // Immediately notify of current state

    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Manually trigger synchronization
   */
  async sync(graphId?: string): Promise<SyncState> {
    if (this.isSyncing) {
      return this._state;
    }

    const queue = getOperationQueue();
    const operations = graphId
      ? queue.getOperations(graphId)
      : queue.getOperations();

    if (operations.length === 0) {
      this.updateState({
        status: 'synced',
        operationsSynced: 0,
        operationsFailed: 0,
        operationsTotal: 0,
      });
      return this._state;
    }

    this.isSyncing = true;
    this.updateState({
      status: 'syncing',
      operationsSynced: 0,
      operationsFailed: 0,
      operationsTotal: operations.length,
    });

    if (this.options.showNotifications) {
      toast.info(`Syncing ${operations.length} offline changes...`);
    }

    try {
      // Group operations by graph
      const operationsByGraph = this.groupOperationsByGraph(operations);

      let synced = 0;
      let failed = 0;

      for (const [gid, ops] of operationsByGraph) {
        const result = await this.syncGraphOperations(gid, ops);
        synced += result.synced;
        failed += result.failed;

        if (result.conflict) {
          // Stop syncing - need user intervention
          this.updateState({
            status: 'conflict',
            conflict: result.conflict,
            operationsSynced: synced,
            operationsFailed: failed,
            operationsTotal: operations.length,
          });

          if (this.options.showNotifications) {
            toast.warning('Sync conflict detected', {
              description: 'The graph has been modified on the server.',
              action: {
                label: 'Resolve',
                onClick: () => {
                  // This would open the conflict dialog
                  // The actual integration is done via the UI component
                },
              },
            });
          }

          this.isSyncing = false;
          return this._state;
        }
      }

      const finalStatus: SyncStatus = failed > 0 ? 'error' : 'synced';
      this.updateState({
        status: finalStatus,
        operationsSynced: synced,
        operationsFailed: failed,
        operationsTotal: operations.length,
        error: failed > 0 ? `${failed} operations failed` : undefined,
      });

      if (this.options.showNotifications) {
        if (failed === 0) {
          toast.success(`Synced ${synced} changes`);
        } else {
          toast.error(`Synced ${synced} changes, ${failed} failed`);
        }
      }

      return this._state;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Sync failed';
      this.updateState({
        status: 'error',
        error: errorMsg,
      });

      if (this.options.showNotifications) {
        toast.error('Sync failed', { description: errorMsg });
      }

      return this._state;
    } finally {
      this.isSyncing = false;
    }
  }

  /**
   * Resolve a sync conflict
   */
  async resolveConflict(
    conflict: VersionConflictResult,
    resolution: 'client' | 'server'
  ): Promise<boolean> {
    if (resolution === 'server') {
      // Accept server version
      const result = resolveWithServerVersion(conflict);
      if (result.success) {
        // Clear the queue for this graph
        const queue = getOperationQueue();
        if (conflict.serverGraph) {
          queue.clearForGraph(conflict.serverGraph.graph_id);
        }
        this.updateState({ status: 'synced', conflict: undefined });
        return true;
      }
      return false;
    } else {
      // Force client version
      const queue = getOperationQueue();
      if (conflict.serverGraph) {
        // Create a full sync operation with current state
        const store = useGLCStore.getState();
        const graph = store.graph;
        if (graph) {
          queue.clearForGraph(conflict.serverGraph.graph_id);
          queue.enqueue(
            createFullSyncOperation(
              graph.metadata.id,
              graph.nodes,
              graph.edges,
              graph.metadata
            )
          );
        }
      }

      // Try syncing again
      const state = await this.sync();
      return state.status === 'synced';
    }
  }

  /**
   * Destroy the sync manager
   */
  destroy(): void {
    if (this.unsubscribeOnline) {
      this.unsubscribeOnline();
      this.unsubscribeOnline = null;
    }
    if (this.unsubscribeQueue) {
      this.unsubscribeQueue();
      this.unsubscribeQueue = null;
    }
    if (this.syncTimeout) {
      clearTimeout(this.syncTimeout);
      this.syncTimeout = null;
    }
    this.listeners.clear();
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private updateState(partial: Partial<SyncState>): void {
    this._state = { ...this._state, ...partial };
    this.notifyListeners();
  }

  private notifyListeners(): void {
    const state = this._state;
    this.listeners.forEach((listener) => {
      try {
        listener(state);
      } catch (error) {
        console.error('Error in sync listener:', error);
      }
    });
  }

  private handleOnlineStatusChange(status: OnlineStatus): void {
    if (status === 'online' && this.options.autoSync) {
      // Delay sync slightly to allow network to stabilize
      if (this.syncTimeout) {
        clearTimeout(this.syncTimeout);
      }
      this.syncTimeout = setTimeout(() => {
        this.sync();
      }, this.options.syncDelay);
    }
  }

  private handleQueueChange(queueState: OperationQueueState): void {
    if (queueState.count === 0 && this._state.status !== 'syncing') {
      this.updateState({ status: 'synced' });
    }
  }

  private groupOperationsByGraph(
    operations: QueuedOperation[]
  ): Map<string, QueuedOperation[]> {
    const grouped = new Map<string, QueuedOperation[]>();

    for (const op of operations) {
      const existing = grouped.get(op.graphId) || [];
      existing.push(op);
      grouped.set(op.graphId, existing);
    }

    return grouped;
  }

  private async syncGraphOperations(
    graphId: string,
    operations: QueuedOperation[]
  ): Promise<{
    synced: number;
    failed: number;
    conflict?: VersionConflictResult;
  }> {
    const queue = getOperationQueue();
    let synced = 0;
    let failed = 0;

    // If there are many operations, use full sync instead
    if (operations.length > 5) {
      // Check for conflicts first
      const store = useGLCStore.getState();
      const graph = store.graph;

      if (graph && graph.metadata.id === graphId) {
        const conflict = await detectVersionConflict(graphId, graph.metadata.version);
        if (conflict.hasConflict) {
          return { synced: 0, failed: 0, conflict };
        }

        // Perform full sync
        const request: UpdateGLCGraphRequest = {
          graph_id: graphId,
          name: graph.metadata.name,
          description: graph.metadata.description,
          nodes: JSON.stringify(graph.nodes),
          edges: JSON.stringify(graph.edges),
          viewport: graph.viewport ? JSON.stringify(graph.viewport) : undefined,
          tags: graph.metadata.tags.join(','),
        };

        const response = await this.performSync(request);

        if (response.success) {
          // Clear all operations for this graph
          queue.clearForGraph(graphId);
          synced = operations.length;

          // Update version from server
          if (response.data?.graph) {
            store.updateMetadata({
              version: response.data.graph.version,
              updatedAt: response.data.graph.updated_at,
            });
          }
        } else {
          failed = operations.length;
        }

        return { synced, failed };
      }
    }

    // Process individual operations
    for (const op of operations) {
      this.updateState({ currentOperation: op.type });

      const result = await this.syncOperation(op);

      if (result.success) {
        queue.remove(op.id);
        synced++;
      } else if (result.conflict) {
        return { synced, failed, conflict: result.conflict };
      } else {
        queue.markFailed(op.id, result.error || 'Unknown error');
        failed++;
      }
    }

    return { synced, failed };
  }

  private async syncOperation(
    operation: QueuedOperation
  ): Promise<{
    success: boolean;
    error?: string;
    conflict?: VersionConflictResult;
  }> {
    // For individual operations, we need to get current state and apply changes
    // This is simplified - in production, you'd want more sophisticated conflict detection

    const store = useGLCStore.getState();
    const graph = store.graph;

    if (!graph || graph.metadata.id !== operation.graphId) {
      // Graph not loaded - can't sync individual operations
      // Create a full sync request instead
      const request = operationToRPCRequest(operation);
      if (request) {
        const response = await this.performSync(request);
        return {
          success: response.success,
          error: response.error,
        };
      }
      return { success: false, error: 'Graph not loaded' };
    }

    // Check for version conflicts
    const conflict = await detectVersionConflict(
      operation.graphId,
      graph.metadata.version
    );
    if (conflict.hasConflict) {
      return { success: false, conflict };
    }

    // Build update request from current state
    const request: UpdateGLCGraphRequest = {
      graph_id: graph.metadata.id,
      name: graph.metadata.name,
      description: graph.metadata.description,
      nodes: JSON.stringify(graph.nodes),
      edges: JSON.stringify(graph.edges),
      viewport: graph.viewport ? JSON.stringify(graph.viewport) : undefined,
      tags: graph.metadata.tags.join(','),
    };

    const response = await this.performSync(request);

    if (response.success && response.data?.graph) {
      // Update version from server
      store.updateMetadata({
        version: response.data.graph.version,
        updatedAt: response.data.graph.updated_at,
      });
    }

    return {
      success: response.success,
      error: response.error,
    };
  }

  private async performSync(
    request: UpdateGLCGraphRequest
  ): Promise<{
    success: boolean;
    error?: string;
    data?: UpdateGLCGraphResponse;
  }> {
    try {
      const response: RPCResponse<UpdateGLCGraphResponse> =
        await rpcClient.updateGLCGraph(request);

      if (response.retcode !== 0) {
        return {
          success: false,
          error: response.message || 'Server error',
        };
      }

      return {
        success: true,
        data: response.payload || undefined,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }
}

// ============================================================================
// Singleton Instance
// ============================================================================

let syncManagerInstance: SyncManager | null = null;

/**
 * Get the singleton sync manager
 */
export function getSyncManager(): SyncManager {
  if (!syncManagerInstance) {
    syncManagerInstance = new SyncManager();
  }
  return syncManagerInstance;
}

/**
 * Destroy the singleton sync manager
 */
export function destroySyncManager(): void {
  if (syncManagerInstance) {
    syncManagerInstance.destroy();
    syncManagerInstance = null;
  }
}

// ============================================================================
// Offline-Aware RPC Wrapper
// ============================================================================

/**
 * Wrapper for RPC calls that queues operations when offline
 *
 * Usage:
 * ```ts
 * const offlineRPC = new OfflineAwareRPC();
 *
 * // This will queue the operation if offline
 * await offlineRPC.updateGraph(graphId, request);
 * ```
 */
export class OfflineAwareRPC {
  private queue = getOperationQueue();
  private detector = getOnlineStatusDetector();

  /**
   * Update graph with offline support
   */
  async updateGraph(
    graphId: string,
    request: UpdateGLCGraphRequest
  ): Promise<RPCResponse<UpdateGLCGraphResponse>> {
    if (this.detector.isOffline) {
      // Queue the operation
      const store = useGLCStore.getState();
      const graph = store.graph;

      if (graph && graph.metadata.id === graphId) {
        this.queue.enqueue(
          createFullSyncOperation(
            graphId,
            graph.nodes,
            graph.edges,
            graph.metadata
          )
        );
      }

      // Return a mock success response
      return {
        retcode: 0,
        message: 'Queued for sync when online',
        payload: null,
      };
    }

    // Online - make the actual RPC call
    return rpcClient.updateGLCGraph(request);
  }

  /**
   * Check if currently offline
   */
  get isOffline(): boolean {
    return this.detector.isOffline;
  }

  /**
   * Get pending operation count
   */
  get pendingCount(): number {
    return this.queue.getPendingCount();
  }
}

/**
 * Get the singleton offline-aware RPC instance
 */
let offlineRPCInstance: OfflineAwareRPC | null = null;

export function getOfflineAwareRPC(): OfflineAwareRPC {
  if (!offlineRPCInstance) {
    offlineRPCInstance = new OfflineAwareRPC();
  }
  return offlineRPCInstance;
}
