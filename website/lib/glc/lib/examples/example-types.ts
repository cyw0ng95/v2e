import { z } from 'zod';
import { Node, Edge } from '@xyflow/react';

export const ExampleGraphNodeSchema = z.object({
  id: z.string(),
  type: z.string(),
  position: z.object({
    x: z.number(),
    y: z.number(),
  }),
  data: z.object({
    label: z.string(),
    description: z.string().optional(),
    properties: z.record(z.any()).optional(),
  }),
});

export const ExampleGraphEdgeSchema = z.object({
  id: z.string(),
  source: z.string(),
  target: z.string(),
  type: z.string(),
  data: z.object({
    label: z.string(),
  }),
});

export const ExampleGraphMetadataSchema = z.object({
  nodeCount: z.number(),
  edgeCount: z.number(),
  complexity: z.enum(['beginner', 'intermediate', 'advanced']),
  created: z.string(),
});

export const ExampleGraphSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  preset: z.enum(['d3fend', 'topo']),
  category: z.string(),
  nodes: z.array(ExampleGraphNodeSchema),
  edges: z.array(ExampleGraphEdgeSchema),
  metadata: ExampleGraphMetadataSchema,
});

export const ExampleGraphsDataSchema = z.object({
  version: z.string(),
  examples: z.array(ExampleGraphSchema),
});

export type ExampleGraphNode = z.infer<typeof ExampleGraphNodeSchema>;
export type ExampleGraphEdge = z.infer<typeof ExampleGraphEdgeSchema>;
export type ExampleGraphMetadata = z.infer<typeof ExampleGraphMetadataSchema>;
export type ExampleGraph = z.infer<typeof ExampleGraphSchema>;
export type ExampleGraphsData = z.infer<typeof ExampleGraphsDataSchema>;
