/**
 * GLC Offline Support Module
 *
 * Provides offline detection, operation queueing, and sync capabilities.
 *
 * Usage:
 * ```ts
 * import {
 *   getOnlineStatusDetector,
 *   getOperationQueue,
 *   getSyncManager,
 *   useOfflineStatus,
 * } from '@/lib/glc/offline';
 *
 * // Check online status
 * const detector = getOnlineStatusDetector();
 * console.log(detector.isOnline);
 *
 * // Queue operations
 * const queue = getOperationQueue();
 * queue.enqueue({ graphId, type: 'add_node', payload: { node } });
 *
 * // Sync when online
 * const syncManager = getSyncManager();
 * await syncManager.sync();
 * ```
 */

// Re-export all public APIs
export {
  // Detector
  OnlineStatusDetector,
  getOnlineStatusDetector,
  destroyOnlineStatusDetector,
  type OnlineStatus,
  type OnlineStatusListener,
  type ConnectivityCheckOptions,
} from './detector';

export {
  // Queue
  OperationQueue,
  getOperationQueue,
  clearOperationQueue,
  destroyOperationQueue,
  createFullSyncOperation,
  operationToRPCRequest,
  type OperationType,
  type QueuedOperation,
  type OperationQueueState,
  type OperationQueueOptions,
  type QueueListener,
  type AddNodePayload,
  type UpdateNodePayload,
  type RemoveNodePayload,
  type AddEdgePayload,
  type UpdateEdgePayload,
  type RemoveEdgePayload,
  type UpdateMetadataPayload,
  type FullSyncPayload,
} from './queue';

export {
  // Sync
  SyncManager,
  getSyncManager,
  destroySyncManager,
  OfflineAwareRPC,
  getOfflineAwareRPC,
  type SyncStatus,
  type SyncState,
  type SyncManagerOptions,
  type SyncListener,
} from './sync';

export {
  // React Hooks
  useOnlineStatus,
  useOperationQueue,
  useSyncStatus,
  useOfflineIndicator,
  type UseOnlineStatusResult,
  type UseOperationQueueResult,
  type UseSyncStatusResult,
  type UseOfflineIndicatorResult,
} from './hooks';
