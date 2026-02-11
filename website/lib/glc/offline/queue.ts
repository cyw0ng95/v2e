/**
 * GLC Offline Operation Queue
 *
 * Manages a queue of operations that need to be synced when online.
 * Uses localStorage for persistence across page reloads.
 */

import type { CADNode, CADEdge, GraphMetadata } from '../types';
import type { UpdateGLCGraphRequest } from '@/lib/types';

// ============================================================================
// Types
// ============================================================================

export type OperationType =
  | 'add_node'
  | 'update_node'
  | 'remove_node'
  | 'add_edge'
  | 'update_edge'
  | 'remove_edge'
  | 'update_metadata'
  | 'full_sync';

export interface QueuedOperation {
  /** Unique ID for this operation */
  id: string;
  /** Graph ID this operation belongs to */
  graphId: string;
  /** Type of operation */
  type: OperationType;
  /** Operation payload */
  payload: unknown;
  /** Timestamp when operation was created */
  timestamp: number;
  /** Number of retry attempts */
  retryCount: number;
  /** Last error message if retry failed */
  lastError?: string;
}

export interface OperationQueueState {
  /** All queued operations */
  operations: QueuedOperation[];
  /** Graph IDs that have pending operations */
  pendingGraphIds: Set<string>;
  /** Total count of operations */
  count: number;
}

export interface OperationQueueOptions {
  /** Key for localStorage (default: 'glc-operation-queue') */
  storageKey?: string;
  /** Maximum number of operations to keep (default: 100) */
  maxOperations?: number;
  /** Maximum retry attempts per operation (default: 3) */
  maxRetries?: number;
}

export interface QueueListener {
  (state: OperationQueueState): void;
}

// ============================================================================
// Constants
// ============================================================================

const DEFAULT_STORAGE_KEY = 'glc-operation-queue';
const DEFAULT_MAX_OPERATIONS = 100;
const DEFAULT_MAX_RETRIES = 3;

// ============================================================================
// Payload Types
// ============================================================================

export interface AddNodePayload {
  node: CADNode;
}

export interface UpdateNodePayload {
  nodeId: string;
  data: Partial<CADNode>;
}

export interface RemoveNodePayload {
  nodeId: string;
}

export interface AddEdgePayload {
  edge: CADEdge;
}

export interface UpdateEdgePayload {
  edgeId: string;
  data: Partial<CADEdge>;
}

export interface RemoveEdgePayload {
  edgeId: string;
}

export interface UpdateMetadataPayload {
  metadata: Partial<GraphMetadata>;
}

export interface FullSyncPayload {
  nodes: CADNode[];
  edges: CADEdge[];
  metadata: GraphMetadata;
}

// ============================================================================
// Operation Queue Class
// ============================================================================

/**
 * Queue for managing offline operations
 *
 * Usage:
 * ```ts
 * const queue = new OperationQueue();
 *
 * // Queue operations
 * queue.enqueue({
 *   graphId: 'graph-123',
 *   type: 'add_node',
 *   payload: { node: newNode },
 * });
 *
 * // Subscribe to changes
 * queue.subscribe((state) => {
 *   console.log(`${state.count} operations pending`);
 * });
 *
 * // Process queue when online
 * for (const op of queue.getOperations('graph-123')) {
 *   await processOperation(op);
 *   queue.remove(op.id);
 * }
 * ```
 */
export class OperationQueue {
  private operations: QueuedOperation[] = [];
  private listeners: Set<QueueListener> = new Set();
  private options: Required<OperationQueueOptions>;

  constructor(options?: OperationQueueOptions) {
    this.options = {
      storageKey: options?.storageKey ?? DEFAULT_STORAGE_KEY,
      maxOperations: options?.maxOperations ?? DEFAULT_MAX_OPERATIONS,
      maxRetries: options?.maxRetries ?? DEFAULT_MAX_RETRIES,
    };

    // Load persisted operations
    this.loadFromStorage();
  }

  /**
   * Get current queue state
   */
  get state(): OperationQueueState {
    return {
      operations: [...this.operations],
      pendingGraphIds: new Set(this.operations.map((op) => op.graphId)),
      count: this.operations.length,
    };
  }

  /**
   * Get all operations
   */
  getOperations(graphId?: string): QueuedOperation[] {
    if (graphId) {
      return this.operations.filter((op) => op.graphId === graphId);
    }
    return [...this.operations];
  }

  /**
   * Get operations for a specific graph, grouped and optimized
   */
  getOptimizedOperations(graphId: string): QueuedOperation[] {
    const ops = this.operations.filter((op) => op.graphId === graphId);

    // If there are many operations, combine them into a single full_sync
    if (ops.length > 10) {
      // This will be handled by the sync manager
      return ops;
    }

    // Otherwise, return individual operations for granular sync
    return ops;
  }

  /**
   * Check if there are pending operations
   */
  hasPending(graphId?: string): boolean {
    if (graphId) {
      return this.operations.some((op) => op.graphId === graphId);
    }
    return this.operations.length > 0;
  }

  /**
   * Get count of pending operations
   */
  getPendingCount(graphId?: string): number {
    if (graphId) {
      return this.operations.filter((op) => op.graphId === graphId).length;
    }
    return this.operations.length;
  }

  /**
   * Add an operation to the queue
   */
  enqueue(operation: {
    graphId: string;
    type: OperationType;
    payload: unknown;
  }): QueuedOperation {
    const queuedOp: QueuedOperation = {
      id: this.generateId(),
      graphId: operation.graphId,
      type: operation.type,
      payload: operation.payload,
      timestamp: Date.now(),
      retryCount: 0,
    };

    // Add to queue
    this.operations.push(queuedOp);

    // Trim queue if over limit
    if (this.operations.length > this.options.maxOperations) {
      // Remove oldest operations first
      const excess = this.operations.length - this.options.maxOperations;
      this.operations = this.operations.slice(excess);
    }

    // Persist and notify
    this.saveToStorage();
    this.notifyListeners();

    return queuedOp;
  }

  /**
   * Remove an operation from the queue
   */
  remove(operationId: string): boolean {
    const index = this.operations.findIndex((op) => op.id === operationId);
    if (index === -1) {
      return false;
    }

    this.operations.splice(index, 1);
    this.saveToStorage();
    this.notifyListeners();

    return true;
  }

  /**
   * Remove all operations for a graph
   */
  clearForGraph(graphId: string): number {
    const initialLength = this.operations.length;
    this.operations = this.operations.filter((op) => op.graphId !== graphId);
    const removedCount = initialLength - this.operations.length;

    if (removedCount > 0) {
      this.saveToStorage();
      this.notifyListeners();
    }

    return removedCount;
  }

  /**
   * Clear all operations
   */
  clear(): void {
    this.operations = [];
    this.saveToStorage();
    this.notifyListeners();
  }

  /**
   * Mark an operation as failed (increment retry count)
   */
  markFailed(operationId: string, error: string): boolean {
    const op = this.operations.find((o) => o.id === operationId);
    if (!op) {
      return false;
    }

    op.retryCount++;
    op.lastError = error;

    // Remove if exceeded max retries
    if (op.retryCount >= this.options.maxRetries) {
      this.remove(operationId);
      return false;
    }

    this.saveToStorage();
    this.notifyListeners();

    return true;
  }

  /**
   * Subscribe to queue changes
   */
  subscribe(listener: QueueListener): () => void {
    this.listeners.add(listener);

    // Immediately notify of current state
    listener(this.state);

    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Get the combined payload for a full sync
   */
  getFullSyncPayload(graphId: string): FullSyncPayload | null {
    const ops = this.operations.filter((op) => op.graphId === graphId);
    if (ops.length === 0) {
      return null;
    }

    // If there's already a full_sync, use it
    const fullSyncOp = ops.find((op) => op.type === 'full_sync');
    if (fullSyncOp) {
      return fullSyncOp.payload as FullSyncPayload;
    }

    // Otherwise, we need to reconstruct from individual operations
    // This is a simplified version - in practice, you'd track the base state
    // and apply operations to it
    return null;
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private generateId(): string {
    return `op-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private saveToStorage(): void {
    if (typeof localStorage === 'undefined') {
      return;
    }

    try {
      localStorage.setItem(
        this.options.storageKey,
        JSON.stringify(this.operations)
      );
    } catch (error) {
      console.error('Failed to save operation queue to localStorage:', error);
    }
  }

  private loadFromStorage(): void {
    if (typeof localStorage === 'undefined') {
      return;
    }

    try {
      const data = localStorage.getItem(this.options.storageKey);
      if (data) {
        this.operations = JSON.parse(data);

        // Validate and filter operations
        const now = Date.now();
        const oneWeekAgo = now - 7 * 24 * 60 * 60 * 1000;

        this.operations = this.operations.filter((op) => {
          // Remove operations older than a week
          if (op.timestamp < oneWeekAgo) {
            return false;
          }

          // Validate required fields
          return (
            op.id &&
            op.graphId &&
            op.type &&
            op.payload !== undefined &&
            typeof op.timestamp === 'number'
          );
        });
      }
    } catch (error) {
      console.error('Failed to load operation queue from localStorage:', error);
      this.operations = [];
    }
  }

  private notifyListeners(): void {
    const state = this.state;
    this.listeners.forEach((listener) => {
      try {
        listener(state);
      } catch (error) {
        console.error('Error in operation queue listener:', error);
      }
    });
  }
}

// ============================================================================
// Singleton Instance
// ============================================================================

let queueInstance: OperationQueue | null = null;

/**
 * Get the singleton operation queue
 */
export function getOperationQueue(): OperationQueue {
  if (!queueInstance) {
    queueInstance = new OperationQueue();
  }
  return queueInstance;
}

/**
 * Clear the singleton queue (useful for testing)
 */
export function clearOperationQueue(): void {
  if (queueInstance) {
    queueInstance.clear();
  }
}

/**
 * Destroy the singleton queue
 */
export function destroyOperationQueue(): void {
  if (queueInstance) {
    queueInstance.clear();
    queueInstance = null;
  }
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Create a full sync operation from current graph state
 */
export function createFullSyncOperation(
  graphId: string,
  nodes: CADNode[],
  edges: CADEdge[],
  metadata: GraphMetadata
): {
  graphId: string;
  type: OperationType;
  payload: FullSyncPayload;
} {
  return {
    graphId,
    type: 'full_sync',
    payload: {
      nodes,
      edges,
      metadata,
    },
  };
}

/**
 * Convert queued operation to RPC request format
 */
export function operationToRPCRequest(
  operation: QueuedOperation
): UpdateGLCGraphRequest | null {
  // This is primarily used for full_sync operations
  if (operation.type !== 'full_sync') {
    return null;
  }

  const payload = operation.payload as FullSyncPayload;
  return {
    graph_id: payload.metadata.id,
    name: payload.metadata.name,
    description: payload.metadata.description,
    nodes: JSON.stringify(payload.nodes),
    edges: JSON.stringify(payload.edges),
    tags: payload.metadata.tags.join(','),
  };
}
