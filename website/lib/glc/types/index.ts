export interface PropertyDefinition {
  id: string;
  name: string;
  type: 'text' | 'number' | 'boolean' | 'enum' | 'multiselect' | 'date';
  required: boolean;
  defaultValue?: any;
  options?: string[];
  description?: string;
}

export interface Reference {
  id: string;
  label: string;
  type: string;
}

export interface NodeStyle {
  backgroundColor: string;
  borderColor: string;
  textColor: string;
  borderWidth?: number;
  borderRadius?: number;
  padding?: string;
  icon?: string;
}

export interface EdgeStyle {
  strokeColor: string;
  strokeWidth: number;
  strokeStyle?: 'solid' | 'dashed' | 'dotted';
  animated?: boolean;
  labelColor?: string;
  labelBackgroundColor?: string;
}

export interface ValidationRule {
  type: 'minLength' | 'maxLength' | 'min' | 'max' | 'pattern' | 'custom';
  value?: any;
  message: string;
  validator?: (value: any) => boolean;
}

export interface InferenceCapability {
  type: 'automatic' | 'suggested' | 'manual';
  sourceNodeTypes: string[];
  relationshipType: string;
  targetNodeType: string;
  properties: Record<string, any>;
}

export interface OntologyMapping {
  ontology: 'D3FEND' | 'CAPEC' | 'ATTACK' | 'CWE' | 'CVE' | 'custom';
  externalId?: string;
  externalType?: string;
  properties: Record<string, string>;
}

export interface NodeTypeDefinition {
  id: string;
  name: string;
  category: string;
  description: string;
  properties: PropertyDefinition[];
  style: NodeStyle;
  ontologyMappings?: OntologyMapping[];
}

export interface RelationshipDefinition {
  id: string;
  name: string;
  category: string;
  description: string;
  sourceNodeTypes: string[];
  targetNodeTypes: string[];
  style: EdgeStyle;
  directionality: 'directed' | 'bidirectional' | 'undirected';
  multiplicity: 'one-to-one' | 'one-to-many' | 'many-to-many';
  properties?: PropertyDefinition[];
}

export interface PresetStyling {
  theme: 'light' | 'dark';
  primaryColor: string;
  backgroundColor: string;
  gridColor: string;
  fontFamily: string;
  customCSS?: string;
}

export interface PresetBehavior {
  pan: boolean;
  zoom: boolean;
  snapToGrid: boolean;
  gridSize: number;
  undoRedo: boolean;
  autoSave: boolean;
  autoSaveInterval: number;
  maxNodes: number;
  maxEdges: number;
}

export interface CanvasPreset {
  id: string;
  name: string;
  version: string;
  category: string;
  description: string;
  author: string;
  createdAt: string;
  updatedAt: string;
  isBuiltIn: boolean;
  nodeTypes: NodeTypeDefinition[];
  relationshipTypes: RelationshipDefinition[];
  styling: PresetStyling;
  behavior: PresetBehavior;
  validationRules: ValidationRule[];
  inferenceCapabilities?: InferenceCapability[];
  metadata: {
    tags: string[];
    previewImage?: string;
    documentationUrl?: string;
  };
}

export interface CADNode {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: Record<string, any>;
  style?: Record<string, any>;
}

export interface CADEdge {
  id: string;
  source: string;
  target: string;
  type: string;
  data: Record<string, any>;
  style?: Record<string, any>;
  animated?: boolean;
}

export interface GraphMetadata {
  id: string;
  name: string;
  description: string;
  presetId: string;
  version: number;
  createdAt: string;
  updatedAt: string;
  author: string;
  tags: string[];
  isPublic: boolean;
}

export interface Graph {
  metadata: GraphMetadata;
  nodes: CADNode[];
  edges: CADEdge[];
  viewport?: {
    x: number;
    y: number;
    zoom: number;
  };
}

export type NodeId = string;
export type EdgeId = string;
export type PresetId = string;
export type GraphId = string;
