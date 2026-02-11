/**
 * GLC Optimistic UI Utilities
 *
 * Provides immediate UI updates before server confirmation with:
 * - Version conflict detection (compare client/server versions)
 * - Automatic rollback on failure
 * - Toast notifications for status updates
 */

import { toast } from 'sonner';
import type { Graph, GraphMetadata, CADNode, CADEdge } from '../types';
import type { GLCGraph, UpdateGLCGraphRequest, UpdateGLCGraphResponse, RPCResponse } from '@/lib/types';
import { useGLCStore } from '../store';
import { rpcClient } from '@/lib/rpc-client';

// Re-export types for convenience
export type { CADNode, CADEdge, Graph, GraphMetadata } from '../types';

// ============================================================================
// Types
// ============================================================================

/**
 * Result of a version conflict check
 */
export interface VersionConflictResult {
  hasConflict: boolean;
  clientVersion: number;
  serverVersion: number;
  serverGraph?: GLCGraph;
}

/**
 * Snapshot of the current graph state for rollback
 */
export interface GraphSnapshot {
  nodes: CADNode[];
  edges: CADEdge[];
  metadata: GraphMetadata;
  viewport?: { x: number; y: number; zoom: number } | undefined;
}

/**
 * Options for optimistic updates
 */
export interface OptimisticUpdateOptions {
  /** Skip conflict detection (force update) */
  force?: boolean;
  /** Custom error message on failure */
  errorMessage?: string;
  /** Custom success message */
  successMessage?: string;
  /** Callback before optimistic update (for validation) */
  onBeforeUpdate?: () => boolean | Promise<boolean>;
  /** Callback after successful server confirmation */
  onSuccess?: (response: RPCResponse<UpdateGLCGraphResponse>) => void;
  /** Callback on conflict detected */
  onConflict?: (conflict: VersionConflictResult) => void;
  /** Callback on rollback */
  onRollback?: (snapshot: GraphSnapshot) => void;
}

/**
 * Result of an optimistic update operation
 */
export interface OptimisticUpdateResult {
  success: boolean;
  conflicted?: boolean;
  conflict?: VersionConflictResult;
  error?: string;
}

// ============================================================================
// Version Conflict Detection
// ============================================================================

/**
 * Check for version conflicts between client and server
 */
export async function detectVersionConflict(
  graphId: string,
  clientVersion: number
): Promise<VersionConflictResult> {
  try {
    const response = await rpcClient.getGLCGraph({ graph_id: graphId });

    if (response.retcode !== 0 || !response.payload?.graph) {
      // Graph not found or error - no conflict, will fail on update
      return {
        hasConflict: false,
        clientVersion,
        serverVersion: clientVersion,
      };
    }

    const serverGraph = response.payload.graph;
    const serverVersion = serverGraph.version;

    return {
      hasConflict: serverVersion > clientVersion,
      clientVersion,
      serverVersion,
      serverGraph,
    };
  } catch (error) {
    console.error('Failed to check version conflict:', error);
    // On error, assume no conflict and let the update proceed
    return {
      hasConflict: false,
      clientVersion,
      serverVersion: clientVersion,
    };
  }
}

// ============================================================================
// Snapshot Management
// ============================================================================

/**
 * Create a snapshot of the current graph state for potential rollback
 */
export function createGraphSnapshot(): GraphSnapshot | null {
  const store = useGLCStore.getState();
  const graph = store.graph;

  if (!graph) {
    return null;
  }

  return {
    nodes: [...graph.nodes],
    edges: [...graph.edges],
    metadata: { ...graph.metadata },
    viewport: graph.viewport ? { ...graph.viewport } : undefined,
  };
}

/**
 * Restore graph state from a snapshot (rollback)
 */
export function restoreFromSnapshot(snapshot: GraphSnapshot): void {
  const store = useGLCStore.getState();

  store.setGraph({
    metadata: snapshot.metadata,
    nodes: snapshot.nodes,
    edges: snapshot.edges,
    viewport: snapshot.viewport,
  });
}

// ============================================================================
// Optimistic Update Operations
// ============================================================================

/**
 * Convert local Graph to GLCGraph update request
 */
function graphToUpdateRequest(graph: Graph): UpdateGLCGraphRequest {
  return {
    graph_id: graph.metadata.id,
    name: graph.metadata.name,
    description: graph.metadata.description,
    nodes: JSON.stringify(graph.nodes),
    edges: JSON.stringify(graph.edges),
    viewport: graph.viewport ? JSON.stringify(graph.viewport) : undefined,
    tags: graph.metadata.tags.join(','),
  };
}

/**
 * Parse GLCGraph nodes/edges from JSON strings
 */
export function parseGLCGraphData(serverGraph: GLCGraph): {
  nodes: CADNode[];
  edges: CADEdge[];
} {
  let nodes: CADNode[] = [];
  let edges: CADEdge[] = [];

  try {
    if (serverGraph.nodes) {
      nodes = JSON.parse(serverGraph.nodes);
    }
  } catch (e) {
    console.error('Failed to parse server nodes:', e);
  }

  try {
    if (serverGraph.edges) {
      edges = JSON.parse(serverGraph.edges);
    }
  } catch (e) {
    console.error('Failed to parse server edges:', e);
  }

  return { nodes, edges };
}

/**
 * Perform an optimistic update with automatic rollback on failure
 *
 * This function:
 * 1. Creates a snapshot for potential rollback
 * 2. Applies optimistic update immediately to the store
 * 3. Sends the update to the server
 * 4. On success, updates with server response
 * 5. On failure, rolls back and shows error toast
 */
export async function performOptimisticUpdate(
  graphId: string,
  updateFn: () => void,
  options: OptimisticUpdateOptions = {}
): Promise<OptimisticUpdateResult> {
  const {
    force = false,
    errorMessage = 'Failed to save changes',
    successMessage,
    onBeforeUpdate,
    onSuccess,
    onConflict,
    onRollback,
  } = options;

  // Get current state
  const store = useGLCStore.getState();
  const currentGraph = store.graph;

  if (!currentGraph) {
    return { success: false, error: 'No graph loaded' };
  }

  const clientVersion = currentGraph.metadata.version;

  // Run pre-update validation if provided
  if (onBeforeUpdate) {
    const canProceed = await onBeforeUpdate();
    if (!canProceed) {
      return { success: false, error: 'Update cancelled by validation' };
    }
  }

  // Check for conflicts unless forced
  if (!force) {
    const conflict = await detectVersionConflict(graphId, clientVersion);
    if (conflict.hasConflict) {
      if (onConflict) {
        onConflict(conflict);
      }
      return {
        success: false,
        conflicted: true,
        conflict,
      };
    }
  }

  // Create snapshot for rollback
  const snapshot = createGraphSnapshot();
  if (!snapshot) {
    return { success: false, error: 'Failed to create snapshot' };
  }

  // Apply optimistic update immediately
  try {
    updateFn();
  } catch (error) {
    console.error('Failed to apply optimistic update:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to apply update',
    };
  }

  // Get the updated graph state
  const updatedGraph = useGLCStore.getState().graph;
  if (!updatedGraph) {
    // This shouldn't happen, but handle it
    restoreFromSnapshot(snapshot);
    return { success: false, error: 'Graph state lost after update' };
  }

  // Send update to server
  try {
    const request = graphToUpdateRequest(updatedGraph);
    const response = await rpcClient.updateGLCGraph(request);

    if (response.retcode !== 0) {
      // Server rejected the update - rollback
      restoreFromSnapshot(snapshot);

      if (onRollback) {
        onRollback(snapshot);
      }

      toast.error(errorMessage, {
        description: response.message || 'Server rejected the update',
      });

      return {
        success: false,
        error: response.message || errorMessage,
      };
    }

    // Success - update version from server response
    if (response.payload?.graph) {
      const serverGraph = response.payload.graph;
      const { nodes, edges } = parseGLCGraphData(serverGraph);

      // Update store with server response to sync versions
      store.updateMetadata({
        version: serverGraph.version,
        updatedAt: serverGraph.updated_at,
      });

      // If server modified the data (unlikely but possible), sync it
      if (JSON.stringify(nodes) !== JSON.stringify(updatedGraph.nodes)) {
        store.setGraph({
          metadata: {
            ...updatedGraph.metadata,
            version: serverGraph.version,
            updatedAt: serverGraph.updated_at,
          },
          nodes,
          edges,
          viewport: updatedGraph.viewport,
        });
      }
    }

    if (successMessage) {
      toast.success(successMessage);
    }

    if (onSuccess) {
      onSuccess(response);
    }

    return { success: true };
  } catch (error) {
    // Network error or unexpected failure - rollback
    restoreFromSnapshot(snapshot);

    if (onRollback) {
      onRollback(snapshot);
    }

    const errorMsg = error instanceof Error ? error.message : 'Network error';
    toast.error(errorMessage, {
      description: errorMsg,
    });

    return {
      success: false,
      error: errorMsg,
    };
  }
}

// ============================================================================
// Specialized Optimistic Operations
// ============================================================================

/**
 * Optimistically add a node to the graph
 */
export async function optimisticAddNode(
  node: CADNode,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.addNode(node),
    {
      successMessage: 'Node added',
      errorMessage: 'Failed to add node',
      ...options,
    }
  );
}

/**
 * Optimistically update a node in the graph
 */
export async function optimisticUpdateNode(
  nodeId: string,
  data: Partial<CADNode>,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.updateNode(nodeId, data),
    {
      successMessage: 'Node updated',
      errorMessage: 'Failed to update node',
      ...options,
    }
  );
}

/**
 * Optimistically remove a node from the graph
 */
export async function optimisticRemoveNode(
  nodeId: string,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.removeNode(nodeId),
    {
      successMessage: 'Node removed',
      errorMessage: 'Failed to remove node',
      ...options,
    }
  );
}

/**
 * Optimistically add an edge to the graph
 */
export async function optimisticAddEdge(
  edge: CADEdge,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.addEdge(edge),
    {
      successMessage: 'Connection added',
      errorMessage: 'Failed to add connection',
      ...options,
    }
  );
}

/**
 * Optimistically update an edge in the graph
 */
export async function optimisticUpdateEdge(
  edgeId: string,
  data: Partial<CADNode>,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.updateEdge(edgeId, data),
    {
      successMessage: 'Connection updated',
      errorMessage: 'Failed to update connection',
      ...options,
    }
  );
}

/**
 * Optimistically remove an edge from the graph
 */
export async function optimisticRemoveEdge(
  edgeId: string,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.removeEdge(edgeId),
    {
      successMessage: 'Connection removed',
      errorMessage: 'Failed to remove connection',
      ...options,
    }
  );
}

/**
 * Optimistically update graph metadata
 */
export async function optimisticUpdateMetadata(
  metadata: Partial<GraphMetadata>,
  options?: OptimisticUpdateOptions
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => store.updateMetadata(metadata),
    {
      successMessage: 'Graph updated',
      errorMessage: 'Failed to update graph',
      ...options,
    }
  );
}

// ============================================================================
// Batch Operations
// ============================================================================

/**
 * Batch update options
 */
export interface BatchUpdateOptions extends OptimisticUpdateOptions {
  /** Whether to apply all updates in a single server request */
  atomic?: boolean;
}

/**
 * Perform multiple optimistic updates in a batch
 */
export async function optimisticBatchUpdate(
  updates: (() => void)[],
  options: BatchUpdateOptions = {}
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graphId = store.graph?.metadata.id;

  if (!graphId) {
    return { success: false, error: 'No graph loaded' };
  }

  return performOptimisticUpdate(
    graphId,
    () => {
      updates.forEach((update) => update());
    },
    {
      successMessage: 'Changes saved',
      errorMessage: 'Failed to save changes',
      ...options,
    }
  );
}

// ============================================================================
// Conflict Resolution Helpers
// ============================================================================

/**
 * Merge strategy for conflict resolution
 */
export type MergeStrategy = 'client' | 'server' | 'manual';

/**
 * Result of a merge operation
 */
export interface MergeResult {
  nodes: CADNode[];
  edges: CADEdge[];
  conflicts: Array<{
    type: 'node' | 'edge';
    id: string;
    clientData: unknown;
    serverData: unknown;
  }>;
}

/**
 * Compute a simple diff between two arrays of nodes
 */
export function diffNodes(
  clientNodes: CADNode[],
  serverNodes: CADNode[]
): { added: CADNode[]; removed: CADNode[]; modified: CADNode[] } {
  const clientMap = new Map(clientNodes.map((n) => [n.id, n]));
  const serverMap = new Map(serverNodes.map((n) => [n.id, n]));

  const added: CADNode[] = [];
  const removed: CADNode[] = [];
  const modified: CADNode[] = [];

  // Find added and modified
  for (const node of clientNodes) {
    const serverNode = serverMap.get(node.id);
    if (!serverNode) {
      added.push(node);
    } else if (JSON.stringify(node) !== JSON.stringify(serverNode)) {
      modified.push(node);
    }
  }

  // Find removed
  for (const node of serverNodes) {
    if (!clientMap.has(node.id)) {
      removed.push(node);
    }
  }

  return { added, removed, modified };
}

/**
 * Compute a simple diff between two arrays of edges
 */
export function diffEdges(
  clientEdges: CADEdge[],
  serverEdges: CADEdge[]
): { added: CADEdge[]; removed: CADEdge[]; modified: CADEdge[] } {
  const clientMap = new Map(clientEdges.map((e) => [e.id, e]));
  const serverMap = new Map(serverEdges.map((e) => [e.id, e]));

  const added: CADEdge[] = [];
  const removed: CADEdge[] = [];
  const modified: CADEdge[] = [];

  // Find added and modified
  for (const edge of clientEdges) {
    const serverEdge = serverMap.get(edge.id);
    if (!serverEdge) {
      added.push(edge);
    } else if (JSON.stringify(edge) !== JSON.stringify(serverEdge)) {
      modified.push(edge);
    }
  }

  // Find removed
  for (const edge of serverEdges) {
    if (!clientMap.has(edge.id)) {
      removed.push(edge);
    }
  }

  return { added, removed, modified };
}

/**
 * Resolve conflict by choosing client version (overwrite server)
 */
export async function resolveWithClientVersion(
  conflict: VersionConflictResult
): Promise<OptimisticUpdateResult> {
  const store = useGLCStore.getState();
  const graph = store.graph;

  if (!graph) {
    return { success: false, error: 'No graph loaded' };
  }

  // Force update with current client state
  return performOptimisticUpdate(
    graph.metadata.id,
    () => {
      // No local change needed, just force the update
    },
    {
      force: true,
      successMessage: 'Changes saved (overwrote server version)',
      errorMessage: 'Failed to overwrite server version',
    }
  );
}

/**
 * Resolve conflict by choosing server version (discard client changes)
 */
export function resolveWithServerVersion(
  conflict: VersionConflictResult
): OptimisticUpdateResult {
  if (!conflict.serverGraph) {
    return { success: false, error: 'No server version available' };
  }

  const store = useGLCStore.getState();
  const { nodes, edges } = parseGLCGraphData(conflict.serverGraph);

  store.setGraph({
    metadata: {
      id: conflict.serverGraph.graph_id,
      name: conflict.serverGraph.name,
      description: conflict.serverGraph.description,
      presetId: conflict.serverGraph.preset_id,
      tags: conflict.serverGraph.tags ? conflict.serverGraph.tags.split(',').filter(Boolean) : [],
      createdAt: conflict.serverGraph.created_at,
      updatedAt: conflict.serverGraph.updated_at,
      version: conflict.serverGraph.version,
    },
    nodes,
    edges,
    viewport: conflict.serverGraph.viewport
      ? JSON.parse(conflict.serverGraph.viewport)
      : undefined,
  });

  toast.success('Discarded local changes, loaded server version');

  return { success: true };
}

// ============================================================================
// Export all
// ============================================================================

export {
  toast,
};
