/**
 * GLC Graph Slice
 */

import type { StateCreator } from 'zustand';
import { nanoid } from 'nanoid';
import type { GraphSlice, Graph, GraphMetadata, CADNode, CADNodeData, CADEdge, CADEdgeData } from '../types';
import type { Viewport } from '@xyflow/react';

export const createGraphSlice: StateCreator<GraphSlice> = (set) => ({
  graph: null,

  setGraph: (graph: Graph) => {
    set({ graph });
  },

  updateMetadata: (metadata: Partial<GraphMetadata>) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          metadata: {
            ...state.graph.metadata,
            ...metadata,
            updatedAt: new Date().toISOString(),
            version: state.graph.metadata.version + 1,
          },
        },
      };
    });
  },

  addNode: (node: CADNode) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          nodes: [...state.graph.nodes, node],
        },
      };
    });
  },

  updateNode: (id: string, data: Partial<CADNodeData>) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          nodes: state.graph.nodes.map((node) =>
            node.id === id ? { ...node, data: { ...node.data, ...data } } : node
          ) as CADNode[],
        },
      };
    });
  },

  removeNode: (id: string) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          nodes: state.graph.nodes.filter((node) => node.id !== id),
          edges: state.graph.edges.filter(
            (edge) => edge.source !== id && edge.target !== id
          ),
        },
      };
    });
  },

  addEdge: (edge: CADEdge) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          edges: [...state.graph.edges, edge],
        },
      };
    });
  },

  updateEdge: (id: string, data: Partial<CADEdgeData>) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          edges: state.graph.edges.map((edge) =>
            edge.id === id ? { ...edge, data: { ...edge.data, ...data } as CADEdgeData } : edge
          ) as CADEdge[],
        },
      };
    });
  },

  removeEdge: (id: string) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          edges: state.graph.edges.filter((edge) => edge.id !== id),
        },
      };
    });
  },

  setViewport: (viewport: Viewport) => {
    set((state) => {
      if (!state.graph) return state;
      return {
        graph: {
          ...state.graph,
          viewport,
        },
      };
    });
  },

  clearGraph: () => {
    set({ graph: null });
  },
});

// Helper to create a new graph
export function createEmptyGraph(presetId: string, name = 'Untitled Graph'): Graph {
  const id = nanoid(12);
  const now = new Date().toISOString();

  return {
    metadata: {
      id,
      name,
      presetId,
      tags: [],
      createdAt: now,
      updatedAt: now,
      version: 1,
    },
    nodes: [],
    edges: [],
  };
}

// Helper to create a new node
export function createNode(
  typeId: string,
  position: { x: number; y: number },
  data: Partial<CADNodeData> = {}
): CADNode {
  return {
    id: nanoid(12),
    type: 'glc',
    position,
    data: {
      label: data.label || 'New Node',
      typeId,
      properties: data.properties || [],
      references: data.references || [],
      color: data.color,
      icon: data.icon,
      d3fendClass: data.d3fendClass,
      notes: data.notes,
    },
  };
}

// Helper to create a new edge
export function createEdge(
  source: string,
  target: string,
  relationshipId: string,
  data: Partial<CADEdgeData> = {}
): CADEdge {
  return {
    id: `edge-${source}-${target}`,
    source,
    target,
    type: 'glc',
    data: {
      relationshipId,
      label: data.label,
      notes: data.notes,
    },
  };
}
