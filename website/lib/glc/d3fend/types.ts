/**
 * D3FEND Full Ontology Types
 *
 * Types for the full D3FEND ontology data structure.
 * These types represent the complete D3FEND knowledge graph.
 */

/**
 * D3FEND Class definition from the full ontology
 */
export interface D3FENDOntologyClass {
  /** Unique D3FEND identifier (e.g., 'd3f:DefensiveTechnique') */
  id: string;
  /** Human-readable label */
  label: string;
  /** Detailed description */
  description?: string;
  /** Parent class ID in the hierarchy */
  parent?: string;
  /** Direct children class IDs */
  children: string[];
  /** Associated technique IDs */
  techniques?: string[];
  /** Additional metadata from ontology */
  metadata?: Record<string, unknown>;
}

/**
 * D3FEND Technique definition
 */
export interface D3FENDTechnique {
  /** Unique technique ID (e.g., 'D3-T101') */
  id: string;
  /** Human-readable name */
  name: string;
  /** Detailed description */
  description?: string;
  /** Associated D3FEND class */
  classId: string;
  /** ATT&CK techniques this mitigates */
  mitigates?: string[];
  /** ATT&CK techniques this detects */
  detects?: string[];
  /** Related techniques */
  related?: string[];
}

/**
 * D3FEND Relationship definition
 */
export interface D3FENDRelationship {
  /** Relationship ID */
  id: string;
  /** Source class/technique ID */
  source: string;
  /** Target class/technique ID */
  target: string;
  /** Relationship type (e.g., 'subClassOf', 'mitigates', 'detects') */
  type: string;
}

/**
 * Full D3FEND Ontology data structure
 */
export interface D3FENDOntologyData {
  /** Ontology version */
  version: string;
  /** Last updated timestamp */
  updatedAt: string;
  /** All classes indexed by ID */
  classes: Record<string, D3FENDOntologyClass>;
  /** All techniques indexed by ID */
  techniques: Record<string, D3FENDTechnique>;
  /** All relationships */
  relationships: D3FENDRelationship[];
  /** Root class IDs (classes without parents) */
  rootClassIds: string[];
}

/**
 * Tree node for the class browser
 */
export interface D3FENDClassTreeNode {
  /** Class ID */
  id: string;
  /** Display label */
  label: string;
  /** Description for tooltip */
  description?: string;
  /** Depth in the tree (0 = root) */
  depth: number;
  /** Whether this node has children */
  hasChildren: boolean;
  /** Whether the node is expanded */
  isExpanded: boolean;
  /** Child node IDs */
  childIds: string[];
  /** Parent node ID */
  parentId?: string;
}

/**
 * Options for the class browser
 */
export interface ClassBrowserOptions {
  /** Filter by class type */
  classType?: string;
  /** Search query filter */
  searchQuery?: string;
  /** Initially expanded node IDs */
  initiallyExpanded?: string[];
  /** Maximum depth to show (-1 for unlimited) */
  maxDepth?: number;
  /** Show techniques under classes */
  showTechniques?: boolean;
  /** Callback when a class is selected */
  onSelect?: (classId: string, classData: D3FENDOntologyClass) => void;
}

/**
 * Options for the class picker
 */
export interface ClassPickerOptions {
  /** Currently selected class ID */
  value?: string;
  /** Placeholder text */
  placeholder?: string;
  /** Whether the picker is disabled */
  disabled?: boolean;
  /** Filter to only show certain class hierarchies */
  rootClassId?: string;
  /** Callback when selection changes */
  onChange?: (classId: string | null, classData: D3FENDOntologyClass | null) => void;
}

/**
 * Loading state for the ontology
 */
export type OntologyLoadState = 'idle' | 'loading' | 'loaded' | 'error';
