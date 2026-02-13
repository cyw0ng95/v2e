/**
 * GLC Diff Utilities
 *
 * Diff visualization helpers for comparing graph versions.
 */

import type { Graph, CADNode, CADEdge, CADNodeData, CADEdgeData } from '../types';
import type { GLCGraphVersion } from '../../types';

/**
 * Change types for diff visualization
 */
export type ChangeType = 'added' | 'removed' | 'modified' | 'unchanged';

/**
 * Base change record
 */
export interface ChangeRecord {
  type: ChangeType;
  timestamp?: string;
}

/**
 * Node change record
 */
export interface NodeChange extends ChangeRecord {
  nodeId: string;
  nodeType?: string;
  label?: string;
  oldData?: Partial<CADNodeData>;
  newData?: Partial<CADNodeData>;
  oldPosition?: { x: number; y: number };
  newPosition?: { x: number; y: number };
}

/**
 * Edge change record
 */
export interface EdgeChange extends ChangeRecord {
  edgeId: string;
  source?: string;
  target?: string;
  oldData?: Partial<CADEdgeData>;
  newData?: Partial<CADEdgeData>;
}

/**
 * Viewport change record
 */
export interface ViewportChange extends ChangeRecord {
  oldViewport?: { x: number; y: number; zoom: number };
  newViewport?: { x: number; y: number; zoom: number };
}

/**
 * Complete diff result between two graph versions
 */
export interface GraphDiff {
  /** Node changes */
  nodes: NodeChange[];
  /** Edge changes */
  edges: EdgeChange[];
  /** Viewport changes */
  viewport: ViewportChange | null;
  /** Summary statistics */
  summary: {
    nodesAdded: number;
    nodesRemoved: number;
    nodesModified: number;
    edgesAdded: number;
    edgesRemoved: number;
    edgesModified: number;
    viewportChanged: boolean;
    totalChanges: number;
  };
  /** Version info */
  fromVersion?: number;
  toVersion?: number;
  fromTimestamp?: string;
  toTimestamp?: string;
}

/**
 * Parse nodes from JSON string
 */
function parseNodes(nodesJson: string): CADNode[] {
  try {
    return JSON.parse(nodesJson) as CADNode[];
  } catch {
    return [];
  }
}

/**
 * Parse edges from JSON string
 */
function parseEdges(edgesJson: string): CADEdge[] {
  try {
    return JSON.parse(edgesJson) as CADEdge[];
  }
  catch {
    return [];
  }
}

/**
 * Parse viewport from JSON string
 */
function parseViewport(viewportJson?: string): { x: number; y: number; zoom: number } | null {
  if (!viewportJson) return null;
  try {
    return JSON.parse(viewportJson) as { x: number; y: number; zoom: number };
  } catch {
    return null;
  }
}

/**
 * Check if node data has changed
 */
function hasNodeDataChanged(
  oldNode: CADNode,
  newNode: CADNode
): { changed: boolean; oldData?: Partial<CADNodeData>; newData?: Partial<CADNodeData> } {
  const oldData = oldNode.data;
  const newData = newNode.data;

  const keys: (keyof CADNodeData)[] = ['label', 'typeId', 'properties', 'references', 'color', 'icon', 'd3fendClass', 'notes'];
  const changes: { oldData: Partial<CADNodeData>; newData: Partial<CADNodeData> } = {
    oldData: {},
    newData: {},
  };

  let hasChanges = false;

  for (const key of keys) {
    const oldValue = JSON.stringify(oldData[key]);
    const newValue = JSON.stringify(newData[key]);

    if (oldValue !== newValue) {
      changes.oldData[key] = oldData[key];
      changes.newData[key] = newData[key];
      hasChanges = true;
    }
  }

  return hasChanges ? { changed: true, ...changes } : { changed: false };
}

/**
 * Check if edge data has changed
 */
function hasEdgeDataChanged(
  oldEdge: CADEdge,
  newEdge: CADEdge
): { changed: boolean; oldData?: Partial<CADEdgeData>; newData?: Partial<CADEdgeData> } {
  const oldData = oldEdge.data ?? {};
  const newData = newEdge.data ?? {};

  const keys: (keyof CADEdgeData)[] = ['relationshipId', 'label', 'notes'];
  const oldDataTyped = oldData as Record<string, unknown>;
  const newDataTyped = newData as Record<string, unknown>;

  const changes: { oldData: Record<string, unknown>; newData: Record<string, unknown> } = {
    oldData: {},
    newData: {},
  };

  let hasChanges = false;

  for (const key of keys) {
    if (JSON.stringify(oldDataTyped[key]) !== JSON.stringify(newDataTyped[key])) {
      changes.oldData[key] = oldDataTyped[key];
      changes.newData[key] = newDataTyped[key];
      hasChanges = true;
    }
  }

  return hasChanges ? { changed: true, ...changes } : { changed: false };
}

/**
 * Compare two graphs and produce a diff
 */
export function diffGraphs(
  oldGraph: Graph | GLCGraphVersion | null,
  newGraph: Graph | GLCGraphVersion
): GraphDiff {
  // Parse nodes and edges based on input type
  const oldNodes: CADNode[] = oldGraph
    ? ('nodes' in oldGraph && typeof oldGraph.nodes === 'string'
      ? parseNodes(oldGraph.nodes)
      : (oldGraph as Graph).nodes || [])
    : [];

  const newNodes: CADNode[] = 'nodes' in newGraph && typeof newGraph.nodes === 'string'
    ? parseNodes(newGraph.nodes)
    : (newGraph as Graph).nodes || [];

  const oldEdges: CADEdge[] = oldGraph
    ? ('edges' in oldGraph && typeof oldGraph.edges === 'string'
      ? parseEdges(oldGraph.edges)
      : (oldGraph as Graph).edges || [])
    : [];

  const newEdges: CADEdge[] = 'edges' in newGraph && typeof newGraph.edges === 'string'
    ? parseEdges(newGraph.edges)
    : (newGraph as Graph).edges || [];

  // Parse viewports
  const oldViewport = oldGraph
    ? ('viewport' in oldGraph
      ? (typeof oldGraph.viewport === 'string'
        ? parseViewport(oldGraph.viewport)
        : (oldGraph as Graph).viewport)
      : null)
    : null;

  const newViewport = 'viewport' in newGraph
    ? (typeof newGraph.viewport === 'string'
      ? parseViewport(newGraph.viewport)
      : (newGraph as Graph).viewport)
    : null;

  // Build node maps for quick lookup
  const oldNodeMap = new Map<string, CADNode>();
  const newNodeMap = new Map<string, CADNode>();

  oldNodes.forEach((node) => oldNodeMap.set(node.id, node));
  newNodes.forEach((node) => newNodeMap.set(node.id, node));

  // Build edge maps for quick lookup
  const oldEdgeMap = new Map<string, CADEdge>();
  const newEdgeMap = new Map<string, CADEdge>();

  oldEdges.forEach((edge) => oldEdgeMap.set(edge.id, edge));
  newEdges.forEach((edge) => newEdgeMap.set(edge.id, edge));

  const nodeChanges: NodeChange[] = [];
  const edgeChanges: EdgeChange[] = [];

  // Find added and modified nodes
  for (const [id, newNode] of newNodeMap) {
    if (!oldNodeMap.has(id)) {
      nodeChanges.push({
        type: 'added',
        nodeId: id,
        nodeType: newNode.data.typeId,
        label: newNode.data.label,
        newPosition: newNode.position,
      });
    } else {
      const oldNode = oldNodeMap.get(id)!;
      const dataChange = hasNodeDataChanged(oldNode, newNode);
      const positionChanged =
        oldNode.position.x !== newNode.position.x ||
        oldNode.position.y !== newNode.position.y;

      if (dataChange.changed || positionChanged) {
        nodeChanges.push({
          type: 'modified',
          nodeId: id,
          nodeType: newNode.data.typeId,
          label: newNode.data.label,
          oldData: dataChange.oldData,
          newData: dataChange.newData,
          oldPosition: positionChanged ? oldNode.position : undefined,
          newPosition: positionChanged ? newNode.position : undefined,
        });
      }
    }
  }

  // Find removed nodes
  for (const [id, oldNode] of oldNodeMap) {
    if (!newNodeMap.has(id)) {
      nodeChanges.push({
        type: 'removed',
        nodeId: id,
        nodeType: oldNode.data.typeId,
        label: oldNode.data.label,
        oldPosition: oldNode.position,
      });
    }
  }

  // Find added and modified edges
  for (const [id, newEdge] of newEdgeMap) {
    if (!oldEdgeMap.has(id)) {
      edgeChanges.push({
        type: 'added',
        edgeId: id,
        source: newEdge.source,
        target: newEdge.target,
      });
    } else {
      const oldEdge = oldEdgeMap.get(id)!;
      const dataChange = hasEdgeDataChanged(oldEdge, newEdge);

      if (dataChange.changed) {
        edgeChanges.push({
          type: 'modified',
          edgeId: id,
          source: newEdge.source,
          target: newEdge.target,
          oldData: dataChange.oldData,
          newData: dataChange.newData,
        });
      }
    }
  }

  // Find removed edges
  for (const [id, oldEdge] of oldEdgeMap) {
    if (!newEdgeMap.has(id)) {
      edgeChanges.push({
        type: 'removed',
        edgeId: id,
        source: oldEdge.source,
        target: oldEdge.target,
      });
    }
  }

  // Check viewport changes
  let viewportChange: ViewportChange | null = null;
  let viewportChanged = false;

  if (oldViewport && newViewport) {
    if (
      oldViewport.x !== newViewport.x ||
      oldViewport.y !== newViewport.y ||
      oldViewport.zoom !== newViewport.zoom
    ) {
      viewportChanged = true;
      viewportChange = {
        type: 'modified',
        oldViewport,
        newViewport,
      };
    }
  } else if (!oldViewport && newViewport) {
    viewportChanged = true;
    viewportChange = {
      type: 'added',
      newViewport,
    };
  } else if (oldViewport && !newViewport) {
    viewportChanged = true;
    viewportChange = {
      type: 'removed',
      oldViewport,
    };
  }

  // Calculate summary
  const nodesAdded = nodeChanges.filter((c) => c.type === 'added').length;
  const nodesRemoved = nodeChanges.filter((c) => c.type === 'removed').length;
  const nodesModified = nodeChanges.filter((c) => c.type === 'modified').length;
  const edgesAdded = edgeChanges.filter((c) => c.type === 'added').length;
  const edgesRemoved = edgeChanges.filter((c) => c.type === 'removed').length;
  const edgesModified = edgeChanges.filter((c) => c.type === 'modified').length;
  const totalChanges = nodeChanges.length + edgeChanges.length + (viewportChanged ? 1 : 0);

  // Extract version info
  const fromVersion = oldGraph && 'version' in oldGraph ? oldGraph.version : undefined;
  const toVersion = 'version' in newGraph ? newGraph.version : undefined;
  const fromTimestamp = oldGraph && 'created_at' in oldGraph ? oldGraph.created_at : undefined;
  const toTimestamp = 'created_at' in newGraph ? newGraph.created_at : undefined;

  return {
    nodes: nodeChanges,
    edges: edgeChanges,
    viewport: viewportChange,
    summary: {
      nodesAdded,
      nodesRemoved,
      nodesModified,
      edgesAdded,
      edgesRemoved,
      edgesModified,
      viewportChanged,
      totalChanges,
    },
    fromVersion,
    toVersion,
    fromTimestamp,
    toTimestamp,
  };
}

/**
 * Format a diff summary for display
 */
export function formatDiffSummary(diff: GraphDiff): string {
  const parts: string[] = [];

  if (diff.summary.nodesAdded > 0) {
    parts.push(`+${diff.summary.nodesAdded} nodes`);
  }
  if (diff.summary.nodesRemoved > 0) {
    parts.push(`-${diff.summary.nodesRemoved} nodes`);
  }
  if (diff.summary.nodesModified > 0) {
    parts.push(`~${diff.summary.nodesModified} nodes`);
  }
  if (diff.summary.edgesAdded > 0) {
    parts.push(`+${diff.summary.edgesAdded} edges`);
  }
  if (diff.summary.edgesRemoved > 0) {
    parts.push(`-${diff.summary.edgesRemoved} edges`);
  }
  if (diff.summary.edgesModified > 0) {
    parts.push(`~${diff.summary.edgesModified} edges`);
  }
  if (diff.summary.viewportChanged) {
    parts.push('viewport');
  }

  return parts.length > 0 ? parts.join(', ') : 'No changes';
}

/**
 * Get change type color
 */
export function getChangeTypeColor(type: ChangeType): string {
  switch (type) {
    case 'added':
      return 'text-green-600 dark:text-green-400';
    case 'removed':
      return 'text-red-600 dark:text-red-400';
    case 'modified':
      return 'text-amber-600 dark:text-amber-400';
    default:
      return 'text-gray-600 dark:text-gray-400';
  }
}

/**
 * Get change type background color
 */
export function getChangeTypeBgColor(type: ChangeType): string {
  switch (type) {
    case 'added':
      return 'bg-green-100 dark:bg-green-900/30';
    case 'removed':
      return 'bg-red-100 dark:bg-red-900/30';
    case 'modified':
      return 'bg-amber-100 dark:bg-amber-900/30';
    default:
      return 'bg-gray-100 dark:bg-gray-900/30';
  }
}

/**
 * Get change type icon name
 */
export function getChangeTypeIcon(type: ChangeType): string {
  switch (type) {
    case 'added':
      return 'plus';
    case 'removed':
      return 'minus';
    case 'modified':
      return 'pencil';
    default:
      return 'minus';
  }
}
