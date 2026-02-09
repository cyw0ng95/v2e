import { useGLCStore } from '../store';
import { Graph, CADNode, CADEdge, GraphMetadata } from '../types';

export interface LoadResult {
  graph: Graph | null;
  loading: boolean;
  error: string | null;
  progress: number;
}

const loadGraph = async (graphId: string): Promise<LoadResult> => {
  const state = useGLCStore.getState();
  
  try {
    return {
      graph: null,
      loading: true,
      error: null,
      progress: 0,
    };
  } catch (error) {
    return {
      graph: null,
      loading: false,
      error: 'Failed to load graph',
      progress: 0,
    };
  }
};

export const saveGraph = async (graph: Graph): Promise<void> => {
  const state = useGLCStore.getState();

  try {
    const validGraph = validateGraph(graph);

    if (!validGraph.valid) {
      throw new Error(validGraph.errors.join(', '));
    }

    state.addGraph(validGraph.graph);
    state.pushGraphToHistory();

    // Auto-save to localStorage
    const graphJSON = JSON.stringify(validGraph.graph, null, 2);

    localStorage.setItem(`glc-graph-${graph.metadata.id}`, graphJSON);

    console.log('Graph saved:', graph.metadata.name);
  } catch (error) {
    console.error('Failed to save graph:', error);
    throw error;
  }
};

export const getGraph = async (graphId: string): Promise<Graph | null> => {
  const graphJSON = localStorage.getItem(`glc-graph-${graphId}`);
  
  if (!graphJSON) {
    return null;
  }

  const graph = JSON.parse(graphJSON) as Graph;

  const validation = validateGraph(graph);

  if (!validation.valid) {
    console.warn('Invalid graph structure:', validation.errors.join(', '));
    return null;
  }

  return graph;
};

export const getAllGraphs = async (): Promise<Graph[]> => {
  const graphs: Graph[] = [];
  
  const graphKeys = Object.keys(localStorage)
    .filter(key => key.startsWith('glc-graph-'));

  for (const key of graphKeys) {
    const graphJSON = localStorage.getItem(key);
    
    if (!graphJSON) continue;

    try {
      const graph = JSON.parse(graphJSON) as Graph;
      graphs.push(graph);
    } catch (error) {
      console.warn(`Failed to load graph from ${key}:`, error);
    }
  }

  return graphs;
};

export const deleteGraph = async (graphId: string): Promise<void> => {
  localStorage.removeItem(`glc-graph-${graphId}`);
  
  useGLCStore.getState().deleteGraph(graphId);
};

export const getGraphSize = (graph: Graph): number => {
  const graphJSON = JSON.stringify(graph, null, 2);
  const blob = new Blob([graphJSON], { type: 'application/json' });
  
  return Math.round(blob.size / 1024);
};

export const validateGraph = (graph: unknown): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (typeof graph !== 'object' || graph === null) {
    errors.push('Graph must be an object');
  }

  const { metadata, nodes, edges, viewport } = graph as Graph;
  const state = useGLCStore.getState();
  const currentPreset = state.currentPreset;

  if (!metadata || typeof metadata.id !== 'string') {
    errors.push('Graph must have valid metadata with string id');
  }

  if (!nodes || !Array.isArray(nodes)) {
    errors.push('Graph must have nodes array');
  }

  if (!edges || !Array.isArray(edges)) {
    errors.push('Graph must have edges array');
  }

  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i];

    if (!node.id || typeof node.id !== 'string') {
      errors.push(`Node at index ${i} missing or has invalid id`);
    }

    if (!node.type || typeof node.type !== 'string') {
      errors.push(`Node ${node.id} missing type`);
    }

    if (!node.position || typeof node.position !== 'object') {
      errors.push(`Node ${node.id} has invalid position`);
    }

    if (typeof node.position.x !== 'number' || typeof node.position.y !== 'number') {
      errors.push(`Node ${node.id} has invalid coordinates`);
    }

    if (!node.data || typeof node.data !== 'object') {
      errors.push(`Node ${node.id} has invalid data`);
    }

    const nodeType = currentPreset?.nodeTypes.find(nt => nt.id === node.type);

    if (!nodeType) {
      errors.push(`Node ${node.id} references non-existent node type: ${node.type}`);
    }
  }

  for (let i = 0; i < edges.length; i++) {
    const edge = edges[i];

    if (!edge.id || typeof edge.id !== 'string') {
      errors.push(`Edge at index ${i} missing or has invalid id`);
    }

    if (!edge.source || typeof edge.source !== 'string') {
      errors.push(`Edge ${edge.id} has missing or invalid source`);
    }

    if (!edge.target || typeof edge.target !== 'string') {
      errors.push(`Edge ${edge.id} has missing or invalid target`);
    }

    if (edges.some(e => e.source === edge.target) {
      const cyclicNodes = new Set<string>();
      const findCycle = (nodeId: string, visited: Set<string>): boolean => {
        if (visited.has(nodeId)) return true;
        visited.add(nodeId);
        return edges
          .filter(e => e.source === nodeId)
          .some(e => e.target === nodeId);
      };

      if (findCycle(nodeId, new Set())) {
        errors.push(`Detected cycle at node ${nodeId}`);
      }
    }

    const edgeType = currentPreset?.relationshipTypes.find(rt => rt.id === edge.type);

    if (!edgeType) {
      errors.push(`Edge ${edge.id} references non-existent relationship type: ${edge.type}`);
    }

    const sourceNode = nodes.find(n => n.id === edge.source);
    const targetNode = nodes.find(n => n.id === edge.target);

    if (!sourceNode || !targetNode) {
      errors.push(`Edge ${edge.id} references non-existent source or target node`);
    }

    const validRelationships = currentPreset.relationshipTypes.filter(rel =>
      (rel.sourceNodeTypes.includes('*') || rel.sourceNodeTypes.includes(sourceNode?.type)) &&
      (rel.targetNodeTypes.includes('*') || rel.targetNodeTypes.includes(targetNode?.type))
    );

    const isValid = validRelationships.some(rel => rel.id === edge.type);

    if (!isValid) {
      errors.push(`Edge ${edge.id} has invalid relationship type for ${edge.source} -> ${edge.target}`);
    }
  }

  if (nodes.length > (currentPreset?.behavior.maxNodes || 1000)) {
    errors.push(`Graph has ${nodes.length} nodes, exceeds preset limit of ${currentPreset?.behavior.maxNodes || 1000}`);
  }

  if (edges.length > (currentPreset?.behavior.maxEdges || 2000)) {
    errors.push(`Graph has ${edges.length} edges, exceeds preset limit of ${currentPreset?.behavior.maxEdges || 2000}`);
  }

  return {
    valid: errors.length === 0,
    errors,
  };
};

export const getInvalidNodes = (graph: Graph): CADNode[] => {
  const errors = validateGraph(graph).errors;
  const invalidNodeIds = errors
    .filter(e => e.code.startsWith('INVALID_NODE_'))
    .map(e => e.code.replace('INVALID_NODE_', ''));

  const invalidNodes = graph.nodes.filter(node =>
    invalidNodeIds.includes(node.id)
  );

  return invalidNodes;
};

export const getInvalidEdges = (graph: Graph): CADEdge[] => {
  const errors = validateGraph(graph).errors;
  const invalidEdgeIds = errors
    .filter(e => e.code.startsWith('INVALID_EDGE_'))
    .map(e => e.code.replace('INVALID_EDGE_', ''));

  const invalidEdges = graph.edges.filter(edge =>
    invalidEdgeIds.includes(edge.id)
  );

  return invalidEdges;
};

export const hasErrors = (graph: Graph): boolean => {
  const validation = validateGraph(graph);
  return !validation.valid;
};

export const getGraphStats = (graph: Graph): {
  return {
    nodes: graph.nodes.length,
    edges: graph.edges.length,
    density: graph.edges.length / Math.max(1, graph.nodes.length),
    maxConnections: graph.edges.length,
    isolatedNodes: graph.nodes.filter(n => !graph.edges.some(e => e.source === n.id || e.target === n.id)).length,
  selfLoops: graph.edges.filter(e => e.source === e.target).length,
    graphSize: getGraphSize(graph),
  };
};

export default {
  loadGraph,
  saveGraph,
  getGraph,
  getAllGraphs,
  deleteGraph,
  getGraphSize,
  validateGraph,
  getInvalidNodes,
  getInvalidEdges,
  hasErrors,
  getGraphStats,
};
};
