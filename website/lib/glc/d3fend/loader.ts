/**
 * D3FEND Ontology Loader
 *
 * Lazy loading for D3FEND ontology data.
 * Attempts to load full ontology from assets, falls back to simplified data.
 */

import { D3FEND_CLASSES, getD3FENDClass, getD3FENDChildren, getD3FENDAncestors } from './ontology';
import type {
  D3FENDOntologyData,
  D3FENDOntologyClass,
  D3FENDClassTreeNode,
  OntologyLoadState,
} from './types';

// Singleton for loaded ontology data
let ontologyData: D3FENDOntologyData | null = null;
let loadState: OntologyLoadState = 'idle';
let loadPromise: Promise<D3FENDOntologyData> | null = null;

/**
 * Convert simplified D3FEND class to full ontology class format
 */
function convertToFullClass(cls: (typeof D3FEND_CLASSES)[0]): D3FENDOntologyClass {
  return {
    id: cls.id,
    label: cls.label,
    description: cls.description,
    parent: cls.parent,
    children: cls.children || [],
    techniques: cls.techniques,
  };
}

/**
 * Build ontology data from simplified D3FEND classes
 */
function buildFromSimplified(): D3FENDOntologyData {
  const classes: Record<string, D3FENDOntologyClass> = {};
  const rootClassIds: string[] = [];

  // Convert all classes
  for (const cls of D3FEND_CLASSES) {
    classes[cls.id] = convertToFullClass(cls);
  }

  // Find root classes (those without parents)
  for (const cls of D3FEND_CLASSES) {
    if (!cls.parent) {
      rootClassIds.push(cls.id);
    }
  }

  return {
    version: '1.0.0-simplified',
    updatedAt: new Date().toISOString(),
    classes,
    techniques: {},
    relationships: [],
    rootClassIds,
  };
}

/**
 * Attempt to load full D3FEND ontology from assets
 * Falls back to simplified data if not available
 */
async function loadFullOntology(): Promise<D3FENDOntologyData> {
  try {
    // Try to load from assets (would be populated by backend)
    const response = await fetch('/assets/d3fend/d3fend.json');

    if (!response.ok) {
      throw new Error(`Failed to load D3FEND ontology: ${response.status}`);
    }

    const data = await response.json();

    // Validate the data structure
    if (!data.classes || !data.rootClassIds) {
      throw new Error('Invalid D3FEND ontology data structure');
    }

    return data as D3FENDOntologyData;
  } catch (error) {
    console.warn('D3FEND full ontology not available, using simplified data:', error);
    return buildFromSimplified();
  }
}

/**
 * Load D3FEND ontology data (lazy, singleton)
 */
export async function loadD3FENDOntology(): Promise<D3FENDOntologyData> {
  // Return cached data if already loaded
  if (ontologyData) {
    return ontologyData;
  }

  // Return existing promise if loading in progress
  if (loadPromise) {
    return loadPromise;
  }

  // Start loading
  loadState = 'loading';
  loadPromise = loadFullOntology()
    .then((data) => {
      ontologyData = data;
      loadState = 'loaded';
      return data;
    })
    .catch((error) => {
      loadState = 'error';
      throw error;
    });

  return loadPromise;
}

/**
 * Get current ontology data (sync, returns null if not loaded)
 */
export function getD3FENDOntology(): D3FENDOntologyData | null {
  return ontologyData;
}

/**
 * Get current load state
 */
export function getOntologyLoadState(): OntologyLoadState {
  return loadState;
}

/**
 * Get a class by ID (works with loaded or simplified data)
 */
export function getClassById(id: string): D3FENDOntologyClass | null {
  // Try loaded data first
  if (ontologyData?.classes[id]) {
    return ontologyData.classes[id];
  }

  // Fall back to simplified data
  const simplified = getD3FENDClass(id);
  if (simplified) {
    return convertToFullClass(simplified);
  }

  return null;
}

/**
 * Get children of a class
 */
export function getClassChildren(classId: string): D3FENDOntologyClass[] {
  // Try loaded data first
  if (ontologyData) {
    const cls = ontologyData.classes[classId];
    if (cls?.children) {
      return cls.children
        .map((id) => ontologyData!.classes[id])
        .filter((c): c is D3FENDOntologyClass => c !== undefined);
    }
    return [];
  }

  // Fall back to simplified data
  return getD3FENDChildren(classId).map(convertToFullClass);
}

/**
 * Get ancestors of a class
 */
export function getClassAncestors(classId: string): D3FENDOntologyClass[] {
  // Try loaded data first
  if (ontologyData) {
    const ancestors: D3FENDOntologyClass[] = [];
    let current = ontologyData.classes[classId];

    while (current?.parent) {
      const parent = ontologyData.classes[current.parent];
      if (parent) {
        ancestors.push(parent);
        current = parent;
      } else {
        break;
      }
    }

    return ancestors.reverse();
  }

  // Fall back to simplified data
  return getD3FENDAncestors(classId).map(convertToFullClass);
}

/**
 * Build tree nodes for the class browser
 */
export function buildTreeNodes(
  rootIds?: string[],
  expandedIds?: Set<string>,
  searchQuery?: string,
  maxDepth = -1
): D3FENDClassTreeNode[] {
  const data = ontologyData || buildFromSimplified();
  const nodes: D3FENDClassTreeNode[] = [];
  const processedIds = new Set<string>();
  const expanded = expandedIds || new Set<string>();

  // Build search filter
  const searchLower = searchQuery?.toLowerCase();
  const matchesSearch = (cls: D3FENDOntologyClass): boolean => {
    if (!searchLower) return true;
    return (
      cls.label.toLowerCase().includes(searchLower) ||
      cls.id.toLowerCase().includes(searchLower) ||
      cls.description?.toLowerCase().includes(searchLower) ||
      false
    );
  };

  // Recursively build tree
  function processNode(classId: string, depth: number): void {
    if (processedIds.has(classId)) return;
    if (maxDepth >= 0 && depth > maxDepth) return;

    const cls = data.classes[classId];
    if (!cls) return;

    processedIds.add(classId);

    const hasChildren = cls.children.length > 0;
    const isExpanded = expanded.has(classId);

    // When searching, show all matching nodes and their ancestors
    const showNode = !searchLower || matchesSearch(cls);

    if (showNode) {
      nodes.push({
        id: cls.id,
        label: cls.label,
        description: cls.description,
        depth,
        hasChildren,
        isExpanded,
        childIds: cls.children,
        parentId: cls.parent,
      });
    }

    // Process children if expanded or if searching
    if ((isExpanded || searchLower) && hasChildren) {
      for (const childId of cls.children) {
        processNode(childId, depth + 1);
      }
    }
  }

  // Start from root classes or specified roots
  const roots = rootIds || data.rootClassIds;
  for (const rootId of roots) {
    processNode(rootId, 0);
  }

  return nodes;
}

/**
 * Search classes by query
 */
export function searchClasses(query: string, limit = 50): D3FENDOntologyClass[] {
  const data = ontologyData || buildFromSimplified();
  const lower = query.toLowerCase();
  const results: D3FENDOntologyClass[] = [];

  for (const cls of Object.values(data.classes)) {
    if (
      cls.label.toLowerCase().includes(lower) ||
      cls.id.toLowerCase().includes(lower) ||
      cls.description?.toLowerCase().includes(lower)
    ) {
      results.push(cls);
      if (results.length >= limit) break;
    }
  }

  return results;
}

/**
 * Reset loader state (for testing)
 */
export function resetLoader(): void {
  ontologyData = null;
  loadState = 'idle';
  loadPromise = null;
}
