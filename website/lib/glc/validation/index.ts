/**
 * GLC Validation Schemas using Zod
 */

import { z } from 'zod';

// ============================================================================
// Property & Reference Schemas
// ============================================================================

export const PropertySchema = z.object({
  key: z.string().min(1, 'Property key is required'),
  value: z.string(),
  type: z.enum(['string', 'number', 'boolean', 'date', 'url']),
  required: z.boolean().optional(),
});

export const ReferenceSchema = z.object({
  type: z.enum(['cve', 'cwe', 'capec', 'attack', 'd3fend', 'url', 'stix']),
  id: z.string().min(1, 'Reference ID is required'),
  label: z.string().optional(),
  url: z.string().url().optional(),
});

// ============================================================================
// Node Type & Relationship Schemas
// ============================================================================

export const NodeTypeDefinitionSchema = z.object({
  id: z.string().min(1, 'Node type ID is required'),
  label: z.string().min(1, 'Node type label is required'),
  category: z.string().min(1, 'Category is required'),
  description: z.string().optional(),
  icon: z.string().optional(),
  color: z.string().regex(/^#[0-9a-fA-F]{6}$/, 'Invalid color format'),
  borderColor: z.string().regex(/^#[0-9a-fA-F]{6}$/, 'Invalid color format').optional(),
  backgroundColor: z.string().regex(/^#[0-9a-fA-F]{6}$/, 'Invalid color format').optional(),
  defaultWidth: z.number().positive().optional(),
  defaultHeight: z.number().positive().optional(),
  properties: z.array(PropertySchema).optional(),
  d3fendClass: z.string().optional(),
  allowedRelationships: z.array(z.string()).optional(),
});

export const RelationshipStyleSchema = z.object({
  strokeColor: z.string().regex(/^#[0-9a-fA-F]{6}$/, 'Invalid color format').optional(),
  strokeWidth: z.number().positive().optional(),
  strokeStyle: z.enum(['solid', 'dashed', 'dotted']).optional(),
  animated: z.boolean().optional(),
  markerEnd: z.boolean().optional(),
  markerStart: z.boolean().optional(),
});

export const RelationshipDefinitionSchema = z.object({
  id: z.string().min(1, 'Relationship ID is required'),
  label: z.string().min(1, 'Relationship label is required'),
  description: z.string().optional(),
  sourceTypes: z.array(z.string()).min(1, 'At least one source type is required'),
  targetTypes: z.array(z.string()).min(1, 'At least one target type is required'),
  style: RelationshipStyleSchema.optional(),
});

// ============================================================================
// Preset Schemas
// ============================================================================

export const CanvasPresetMetaSchema = z.object({
  id: z.string().min(1, 'Preset ID is required'),
  name: z.string().min(1, 'Preset name is required'),
  version: z.string().regex(/^\d+\.\d+\.\d+$/, 'Version must be semver (e.g., 1.0.0)'),
  description: z.string().optional(),
  author: z.string().optional(),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
});

export const CanvasPresetThemeSchema = z.object({
  primary: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  background: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  surface: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  text: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  textMuted: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  border: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  accent: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  success: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  warning: z.string().regex(/^#[0-9a-fA-F]{6}$/),
  error: z.string().regex(/^#[0-9a-fA-F]{6}$/),
});

export const CanvasPresetBehaviorSchema = z.object({
  snapToGrid: z.boolean(),
  gridSize: z.number().positive(),
  autoLayout: z.boolean(),
  historyLimit: z.number().int().min(10).max(1000),
  autoSaveInterval: z.number().int().min(1000).max(300000),
  enableInference: z.boolean(),
});

export const CanvasPresetSchema = z.object({
  meta: CanvasPresetMetaSchema,
  theme: CanvasPresetThemeSchema,
  behavior: CanvasPresetBehaviorSchema,
  nodeTypes: z.array(NodeTypeDefinitionSchema).min(1, 'At least one node type is required'),
  relationships: z.array(RelationshipDefinitionSchema),
});

// ============================================================================
// Graph Schemas
// ============================================================================

export const CADNodeDataSchema = z.object({
  label: z.string().min(1, 'Node label is required'),
  typeId: z.string().min(1, 'Node type is required'),
  properties: z.array(PropertySchema),
  references: z.array(ReferenceSchema),
  color: z.string().optional(),
  icon: z.string().optional(),
  d3fendClass: z.string().optional(),
  notes: z.string().optional(),
});

export const CADNodeSchema = z.object({
  id: z.string().min(1),
  type: z.string().optional(),
  position: z.object({
    x: z.number(),
    y: z.number(),
  }),
  data: CADNodeDataSchema,
  width: z.number().optional(),
  height: z.number().optional(),
  selected: z.boolean().optional(),
});

export const CADEdgeDataSchema = z.object({
  relationshipId: z.string().min(1, 'Relationship type is required'),
  label: z.string().optional(),
  notes: z.string().optional(),
});

export const CADEdgeSchema = z.object({
  id: z.string().min(1),
  source: z.string().min(1),
  target: z.string().min(1),
  type: z.string().optional(),
  data: CADEdgeDataSchema,
  selected: z.boolean().optional(),
});

export const GraphMetadataSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1, 'Graph name is required'),
  description: z.string().optional(),
  presetId: z.string().min(1, 'Preset ID is required'),
  tags: z.array(z.string()),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  version: z.number().int().nonnegative(),
});

export const ViewportSchema = z.object({
  x: z.number(),
  y: z.number(),
  zoom: z.number().positive(),
});

export const GraphSchema = z.object({
  metadata: GraphMetadataSchema,
  nodes: z.array(CADNodeSchema),
  edges: z.array(CADEdgeSchema),
  viewport: ViewportSchema.optional(),
});

// ============================================================================
// Validation Functions
// ============================================================================

export function validatePreset(data: unknown) {
  return CanvasPresetSchema.safeParse(data);
}

export function validateGraph(data: unknown) {
  return GraphSchema.safeParse(data);
}

export function validateNodeType(data: unknown) {
  return NodeTypeDefinitionSchema.safeParse(data);
}

export function validateRelationship(data: unknown) {
  return RelationshipDefinitionSchema.safeParse(data);
}

// ============================================================================
// Preset Migration
// ============================================================================

interface MigrationContext {
  fromVersion: string;
  toVersion: string;
  preset: Record<string, unknown>;
}

const migrations: Record<string, (ctx: MigrationContext) => Record<string, unknown>> = {};

export function registerMigration(
  fromVersion: string,
  toVersion: string,
  migrate: (ctx: MigrationContext) => Record<string, unknown>
) {
  const key = `${fromVersion}->${toVersion}`;
  migrations[key] = migrate;
}

export function migratePreset(
  preset: Record<string, unknown>,
  targetVersion: string
): Record<string, unknown> {
  const currentVersion = (preset.meta as Record<string, string>)?.version || '1.0.0';

  if (currentVersion === targetVersion) {
    return preset;
  }

  const key = `${currentVersion}->${targetVersion}`;
  const migration = migrations[key];

  if (!migration) {
    console.warn(`No migration path from ${currentVersion} to ${targetVersion}`);
    return preset;
  }

  return migration({ fromVersion: currentVersion, toVersion: targetVersion, preset });
}
