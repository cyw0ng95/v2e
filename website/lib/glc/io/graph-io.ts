/**
 * GLC Graph I/O Utilities
 *
 * Handles saving, loading, and exporting graphs.
 */

import type { Graph, GraphMetadata } from '../types';

const STORAGE_KEY_PREFIX = 'glc-graph-';
const GRAPH_LIST_KEY = 'glc-graph-list';
const AUTO_SAVE_KEY = 'glc-autosave-';

/**
 * Save graph to localStorage
 */
export function saveGraphToStorage(graph: Graph): void {
  const key = `${STORAGE_KEY_PREFIX}${graph.metadata.id}`;
  localStorage.setItem(key, JSON.stringify(graph));

  // Update graph list
  const list = getGraphList();
  const existingIndex = list.findIndex((g) => g.id === graph.metadata.id);
  const metadata: GraphMetadata = {
    ...graph.metadata,
    updatedAt: new Date().toISOString(),
  };

  if (existingIndex >= 0) {
    list[existingIndex] = metadata;
  } else {
    list.push(metadata);
  }

  localStorage.setItem(GRAPH_LIST_KEY, JSON.stringify(list));
}

/**
 * Load graph from localStorage
 */
export function loadGraphFromStorage(graphId: string): Graph | null {
  const key = `${STORAGE_KEY_PREFIX}${graphId}`;
  const data = localStorage.getItem(key);
  if (!data) return null;

  try {
    return JSON.parse(data) as Graph;
  } catch {
    return null;
  }
}

/**
 * Delete graph from localStorage
 */
export function deleteGraphFromStorage(graphId: string): void {
  localStorage.removeItem(`${STORAGE_KEY_PREFIX}${graphId}`);

  // Update graph list
  const list = getGraphList();
  const filtered = list.filter((g) => g.id !== graphId);
  localStorage.setItem(GRAPH_LIST_KEY, JSON.stringify(filtered));
}

/**
 * Get list of saved graph metadata
 */
export function getGraphList(): GraphMetadata[] {
  const data = localStorage.getItem(GRAPH_LIST_KEY);
  if (!data) return [];

  try {
    return JSON.parse(data) as GraphMetadata[];
  } catch {
    return [];
  }
}

/**
 * Export graph to JSON file
 */
export function exportGraphToFile(graph: Graph): void {
  const json = JSON.stringify(graph, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);

  const a = document.createElement('a');
  a.href = url;
  a.download = `${graph.metadata.name || 'graph'}.json`;
  a.click();

  URL.revokeObjectURL(url);
}

/**
 * Import graph from JSON file
 */
export function importGraphFromFile(file: File): Promise<Graph> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = (e) => {
      try {
        const content = e.target?.result as string;
        const graph = JSON.parse(content) as Graph;

        // Validate basic structure
        if (!graph.metadata || !graph.nodes || !graph.edges) {
          throw new Error('Invalid graph structure');
        }

        // Generate new ID to avoid conflicts
        graph.metadata.id = crypto.randomUUID();
        graph.metadata.createdAt = new Date().toISOString();
        graph.metadata.updatedAt = new Date().toISOString();

        resolve(graph);
      } catch (err) {
        reject(err);
      }
    };

    reader.onerror = () => reject(new Error('Failed to read file'));
    reader.readAsText(file);
  });
}

/**
 * Auto-save graph (debounced save to separate key)
 */
export function autoSaveGraph(graph: Graph, presetId: string): void {
  const key = `${AUTO_SAVE_KEY}${presetId}`;
  localStorage.setItem(key, JSON.stringify({
    graph,
    timestamp: Date.now(),
  }));
}

/**
 * Load auto-saved graph
 */
export function loadAutoSavedGraph(presetId: string): { graph: Graph; timestamp: number } | null {
  const key = `${AUTO_SAVE_KEY}${presetId}`;
  const data = localStorage.getItem(key);
  if (!data) return null;

  try {
    return JSON.parse(data);
  } catch {
    return null;
  }
}

/**
 * Clear auto-saved graph
 */
export function clearAutoSavedGraph(presetId: string): void {
  localStorage.removeItem(`${AUTO_SAVE_KEY}${presetId}`);
}

/**
 * Create debounced auto-save function
 */
export function createAutoSaver(
  presetId: string,
  delay: number = 2000
): (graph: Graph) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  return (graph: Graph) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }

    timeoutId = setTimeout(() => {
      autoSaveGraph(graph, presetId);
      timeoutId = null;
    }, delay);
  };
}
