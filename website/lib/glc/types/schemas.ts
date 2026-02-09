import { z } from 'zod';

export const propertyDefinitionSchema = z.object({
  id: z.string(),
  name: z.string(),
  type: z.enum(['text', 'number', 'boolean', 'enum', 'multiselect', 'date']),
  required: z.boolean(),
  defaultValue: z.any().optional(),
  options: z.array(z.string()).optional(),
  description: z.string().optional(),
});

export const referenceSchema = z.object({
  id: z.string(),
  label: z.string(),
  type: z.string(),
});

export const nodeStyleSchema = z.object({
  backgroundColor: z.string(),
  borderColor: z.string(),
  textColor: z.string(),
  borderWidth: z.number().optional(),
  borderRadius: z.number().optional(),
  padding: z.string().optional(),
  icon: z.string().optional(),
});

export const edgeStyleSchema = z.object({
  strokeColor: z.string(),
  strokeWidth: z.number(),
  strokeStyle: z.enum(['solid', 'dashed', 'dotted']).optional(),
  animated: z.boolean().optional(),
  labelColor: z.string().optional(),
  labelBackgroundColor: z.string().optional(),
});

export const validationRuleSchema = z.object({
  type: z.enum(['minLength', 'maxLength', 'min', 'max', 'pattern', 'custom']),
  value: z.any().optional(),
  message: z.string(),
  validator: z.function().optional(),
});

export const inferenceCapabilitySchema = z.object({
  type: z.enum(['automatic', 'suggested', 'manual']),
  sourceNodeTypes: z.array(z.string()),
  relationshipType: z.string(),
  targetNodeType: z.string(),
  properties: z.record(z.any()),
});

export const ontologyMappingSchema = z.object({
  ontology: z.enum(['D3FEND', 'CAPEC', 'ATTACK', 'CWE', 'CVE', 'custom']),
  externalId: z.string().optional(),
  externalType: z.string().optional(),
  properties: z.record(z.string()),
});

export const nodeTypeDefinitionSchema = z.object({
  id: z.string(),
  name: z.string(),
  category: z.string(),
  description: z.string(),
  properties: z.array(propertyDefinitionSchema),
  style: nodeStyleSchema,
  ontologyMappings: z.array(ontologyMappingSchema).optional(),
});

export const relationshipDefinitionSchema = z.object({
  id: z.string(),
  name: z.string(),
  category: z.string(),
  description: z.string(),
  sourceNodeTypes: z.array(z.string()),
  targetNodeTypes: z.array(z.string()),
  style: edgeStyleSchema,
  directionality: z.enum(['directed', 'bidirectional', 'undirected']),
  multiplicity: z.enum(['one-to-one', 'one-to-many', 'many-to-many']),
  properties: z.array(propertyDefinitionSchema).optional(),
});

export const presetStylingSchema = z.object({
  theme: z.enum(['light', 'dark']),
  primaryColor: z.string(),
  backgroundColor: z.string(),
  gridColor: z.string(),
  fontFamily: z.string(),
  customCSS: z.string().optional(),
});

export const presetBehaviorSchema = z.object({
  pan: z.boolean(),
  zoom: z.boolean(),
  snapToGrid: z.boolean(),
  gridSize: z.number(),
  undoRedo: z.boolean(),
  autoSave: z.boolean(),
  autoSaveInterval: z.number(),
  maxNodes: z.number(),
  maxEdges: z.number(),
});

export const graphMetadataSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  presetId: z.string(),
  version: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
  author: z.string(),
  tags: z.array(z.string()),
  isPublic: z.boolean(),
});

export const canvasPresetSchema = z.object({
  id: z.string(),
  name: z.string(),
  version: z.string(),
  category: z.string(),
  description: z.string(),
  author: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  isBuiltIn: z.boolean(),
  nodeTypes: z.array(nodeTypeDefinitionSchema),
  relationshipTypes: z.array(relationshipDefinitionSchema),
  styling: presetStylingSchema,
  behavior: presetBehaviorSchema,
  validationRules: z.array(validationRuleSchema),
  inferenceCapabilities: z.array(inferenceCapabilitySchema).optional(),
  metadata: z.object({
    tags: z.array(z.string()),
    previewImage: z.string().optional(),
    documentationUrl: z.string().optional(),
  }),
});

export const cadNodeSchema = z.object({
  id: z.string(),
  type: z.string(),
  position: z.object({ x: z.number(), y: z.number() }),
  data: z.record(z.any()),
  style: z.record(z.any()).optional(),
});

export const cadEdgeSchema = z.object({
  id: z.string(),
  source: z.string(),
  target: z.string(),
  type: z.string(),
  data: z.record(z.any()),
  style: z.record(z.any()).optional(),
  animated: z.boolean().optional(),
});

export const graphSchema = z.object({
  metadata: graphMetadataSchema,
  nodes: z.array(cadNodeSchema),
  edges: z.array(cadEdgeSchema),
  viewport: z.object({
    x: z.number(),
    y: z.number(),
    zoom: z.number(),
  }).optional(),
});

export type PropertyDefinitionZod = z.infer<typeof propertyDefinitionSchema>;
export type NodeStyleZod = z.infer<typeof nodeStyleSchema>;
export type EdgeStyleZod = z.infer<typeof edgeStyleSchema>;
export type ValidationRuleZod = z.infer<typeof validationRuleSchema>;
export type InferenceCapabilityZod = z.infer<typeof inferenceCapabilitySchema>;
export type OntologyMappingZod = z.infer<typeof ontologyMappingSchema>;
export type NodeTypeDefinitionZod = z.infer<typeof nodeTypeDefinitionSchema>;
export type RelationshipDefinitionZod = z.infer<typeof relationshipDefinitionSchema>;
export type PresetStylingZod = z.infer<typeof presetStylingSchema>;
export type PresetBehaviorZod = z.infer<typeof presetBehaviorSchema>;
export type GraphMetadataZod = z.infer<typeof graphMetadataSchema>;
export type CanvasPresetZod = z.infer<typeof canvasPresetSchema>;
export type CADNodeZod = z.infer<typeof cadNodeSchema>;
export type CADEdgeZod = z.infer<typeof cadEdgeSchema>;
export type GraphZod = z.infer<typeof graphSchema>;
