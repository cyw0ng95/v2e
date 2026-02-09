import { StateCreator } from 'zustand';
import { CADNode, CADEdge, Graph, GraphMetadata } from '../types';

interface GraphState {
  graph: Graph | null;
  nodes: CADNode[];
  edges: CADEdge[];
  metadata: GraphMetadata | null;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  viewport: { x: number; y: number; zoom: number };
  setGraph: (graph: Graph) => void;
  addNode: (node: CADNode) => void;
  updateNode: (id: string, updates: Partial<CADNode>) => void;
  deleteNode: (id: string) => void;
  addEdge: (edge: CADEdge) => void;
  updateEdge: (id: string, updates: Partial<CADEdge>) => void;
  deleteEdge: (id: string) => void;
  setSelectedNodeId: (id: string | null) => void;
  setSelectedEdgeId: (id: string | null) => void;
  setViewport: (viewport: { x: number; y: number; zoom: number }) => void;
  clearGraph: () => void;
  setMetadata: (metadata: Partial<GraphMetadata>) => void;
  getGraph: () => Graph | null;
}

export const createGraphSlice: StateCreator<GraphState> = (set, get) => ({
  graph: null,
  nodes: [],
  edges: [],
  metadata: null,
  selectedNodeId: null,
  selectedEdgeId: null,
  viewport: { x: 0, y: 0, zoom: 1 },
  setGraph: (graph) => set({ 
    graph, 
    nodes: graph.nodes, 
    edges: graph.edges, 
    metadata: graph.metadata,
    viewport: graph.viewport || { x: 0, y: 0, zoom: 1 },
  }),
  addNode: (node) => set((state) => {
    const newNodes = [...state.nodes, node];
    const newGraph = state.graph ? { ...state.graph, nodes: newNodes } : null;
    return { nodes: newNodes, graph: newGraph };
  }),
  updateNode: (id, updates) => set((state) => {
    const newNodes = state.nodes.map(node => node.id === id ? { ...node, ...updates } : node);
    const newGraph = state.graph ? { ...state.graph, nodes: newNodes } : null;
    return { nodes: newNodes, graph: newGraph };
  }),
  deleteNode: (id) => set((state) => {
    const newNodes = state.nodes.filter(node => node.id !== id);
    const newEdges = state.edges.filter(edge => edge.source !== id && edge.target !== id);
    const newGraph = state.graph ? { ...state.graph, nodes: newNodes, edges: newEdges } : null;
    return { 
      nodes: newNodes, 
      edges: newEdges, 
      graph: newGraph,
      selectedNodeId: state.selectedNodeId === id ? null : state.selectedNodeId,
    };
  }),
  addEdge: (edge) => set((state) => {
    const newEdges = [...state.edges, edge];
    const newGraph = state.graph ? { ...state.graph, edges: newEdges } : null;
    return { edges: newEdges, graph: newGraph };
  }),
  updateEdge: (id, updates) => set((state) => {
    const newEdges = state.edges.map(edge => edge.id === id ? { ...edge, ...updates } : edge);
    const newGraph = state.graph ? { ...state.graph, edges: newEdges } : null;
    return { edges: newEdges, graph: newGraph };
  }),
  deleteEdge: (id) => set((state) => {
    const newEdges = state.edges.filter(edge => edge.id !== id);
    const newGraph = state.graph ? { ...state.graph, edges: newEdges } : null;
    return { 
      edges: newEdges, 
      graph: newGraph,
      selectedEdgeId: state.selectedEdgeId === id ? null : state.selectedEdgeId,
    };
  }),
  setSelectedNodeId: (id) => set({ selectedNodeId: id, selectedEdgeId: null }),
  setSelectedEdgeId: (id) => set({ selectedEdgeId: id, selectedNodeId: null }),
  setViewport: (viewport) => set((state) => {
    const newGraph = state.graph ? { ...state.graph, viewport } : null;
    return { viewport, graph: newGraph };
  }),
  clearGraph: () => set({
    graph: null,
    nodes: [],
    edges: [],
    metadata: null,
    selectedNodeId: null,
    selectedEdgeId: null,
    viewport: { x: 0, y: 0, zoom: 1 },
  }),
  setMetadata: (metadata) => set((state) => {
    const newMetadata = state.metadata ? { ...state.metadata, ...metadata } : metadata as GraphMetadata;
    const newGraph = state.graph ? { ...state.graph, metadata: newMetadata } : null;
    return { metadata: newMetadata, graph: newGraph };
  }),
  getGraph: () => get().graph,
});

export default createGraphSlice;
