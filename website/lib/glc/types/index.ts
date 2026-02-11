/**
 * GLC (Graphized Learning Canvas) Type Definitions
 *
 * Core types for graph-based modeling with customizable presets.
 */

import type { Node, Edge, Viewport } from '@xyflow/react';

// ============================================================================
// Property & Reference Types
// ============================================================================

export interface Property {
  key: string;
  value: string;
  type: 'string' | 'number' | 'boolean' | 'date' | 'url';
  required?: boolean;
}

export interface Reference {
  type: 'cve' | 'cwe' | 'capec' | 'attack' | 'd3fend' | 'url' | 'stix';
  id: string;
  label?: string;
  url?: string;
}

// ============================================================================
// Node Type Definition (Preset Schema)
// ============================================================================

export interface NodeTypeDefinition {
  id: string;
  label: string;
  category: string;
  description?: string;
  icon?: string;
  color: string;
  borderColor?: string;
  backgroundColor?: string;
  defaultWidth?: number;
  defaultHeight?: number;
  properties?: Property[];
  d3fendClass?: string;
  allowedRelationships?: string[];
}

// ============================================================================
// Relationship Definition (Preset Schema)
// ============================================================================

export interface RelationshipDefinition {
  id: string;
  label: string;
  description?: string;
  sourceTypes: string[];
  targetTypes: string[];
  style?: {
    strokeColor?: string;
    strokeWidth?: number;
    strokeStyle?: 'solid' | 'dashed' | 'dotted';
    animated?: boolean;
    markerEnd?: boolean;
    markerStart?: boolean;
  };
}

// ============================================================================
// Canvas Preset
// ============================================================================

export interface CanvasPresetMeta {
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface CanvasPresetTheme {
  primary: string;
  background: string;
  surface: string;
  text: string;
  textMuted: string;
  border: string;
  accent: string;
  success: string;
  warning: string;
  error: string;
}

export interface CanvasPresetBehavior {
  snapToGrid: boolean;
  gridSize: number;
  autoLayout: boolean;
  historyLimit: number;
  autoSaveInterval: number;
  enableInference: boolean;
}

export interface CanvasPreset {
  meta: CanvasPresetMeta;
  theme: CanvasPresetTheme;
  behavior: CanvasPresetBehavior;
  nodeTypes: NodeTypeDefinition[];
  relationships: RelationshipDefinition[];
}

// ============================================================================
// Graph Data Types
// ============================================================================

export interface CADNodeData extends Record<string, unknown> {
  label: string;
  typeId: string;
  properties: Property[];
  references: Reference[];
  color?: string;
  icon?: string;
  d3fendClass?: string;
  notes?: string;
}

export type CADNode = Node<CADNodeData, 'glc'>;

export interface CADEdgeData extends Record<string, unknown> {
  relationshipId: string;
  label?: string;
  notes?: string;
}

export type CADEdge = Edge<CADEdgeData>;

export interface GraphMetadata {
  id: string;
  name: string;
  description?: string;
  presetId: string;
  tags: string[];
  createdAt: string;
  updatedAt: string;
  version: number;
}

export interface Graph {
  metadata: GraphMetadata;
  nodes: CADNode[];
  edges: CADEdge[];
  viewport?: Viewport;
}

// ============================================================================
// Store State Types
// ============================================================================

export interface PresetSlice {
  currentPreset: CanvasPreset | null;
  builtInPresets: CanvasPreset[];
  userPresets: CanvasPreset[];
  setCurrentPreset: (preset: CanvasPreset) => void;
  addUserPreset: (preset: CanvasPreset) => void;
  updateUserPreset: (id: string, preset: Partial<CanvasPreset>) => void;
  removeUserPreset: (id: string) => void;
}

export interface GraphSlice {
  graph: Graph | null;
  setGraph: (graph: Graph) => void;
  updateMetadata: (metadata: Partial<GraphMetadata>) => void;
  addNode: (node: CADNode) => void;
  updateNode: (id: string, data: Partial<CADNodeData>) => void;
  removeNode: (id: string) => void;
  addEdge: (edge: CADEdge) => void;
  updateEdge: (id: string, data: Partial<CADEdgeData>) => void;
  removeEdge: (id: string) => void;
  setViewport: (viewport: Viewport) => void;
  clearGraph: () => void;
}

export interface CanvasSlice {
  selectedNodes: string[];
  selectedEdges: string[];
  zoom: number;
  isPanning: boolean;
  setSelection: (nodes: string[], edges: string[]) => void;
  clearSelection: () => void;
  setZoom: (zoom: number) => void;
  setIsPanning: (isPanning: boolean) => void;
}

export interface UISlice {
  theme: 'light' | 'dark' | 'system';
  sidebarOpen: boolean;
  nodePaletteOpen: boolean;
  detailsPanelOpen: boolean;
  detailsPanelTab: 'properties' | 'references' | 'notes';
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  toggleSidebar: () => void;
  toggleNodePalette: () => void;
  setDetailsPanelOpen: (open: boolean) => void;
  setDetailsPanelTab: (tab: 'properties' | 'references' | 'notes') => void;
}

export interface UndoRedoAction {
  type: string;
  timestamp: number;
  before: unknown;
  after: unknown;
}

export interface UndoRedoSlice {
  canUndo: boolean;
  canRedo: boolean;
  history: UndoRedoAction[];
  currentIndex: number;
  pushAction: (action: Omit<UndoRedoAction, 'timestamp'>) => void;
  undo: () => UndoRedoAction | null;
  redo: () => UndoRedoAction | null;
  clearHistory: () => void;
}

// ============================================================================
// Utility Types
// ============================================================================

export type ViewMode = 'canvas' | 'dashboard' | 'study';

export type ExportFormat = 'json' | 'png' | 'svg' | 'pdf';

export interface ExportOptions {
  format: ExportFormat;
  includeBackground: boolean;
  includeGrid: boolean;
  quality: number;
}

export interface ShareLink {
  id: string;
  graphId: string;
  createdAt: string;
  expiresAt?: string;
  password?: string;
}
