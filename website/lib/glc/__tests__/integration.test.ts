import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import { useGLCStore } from '../store';
import { presetManager } from '../preset-manager';
import { D3FEND_PRESET, TOPO_PRESET } from '../presets';

describe('Integration - Preset Loading Flow', () => {
  beforeEach(() => {
    useGLCStore.getState().resetPreset();
    useGLCStore.getState().clearGraph();
    presetManager.clearBackups();
  });

  afterEach(() => {
    useGLCStore.getState().resetPreset();
    useGLCStore.getState().clearGraph();
    presetManager.clearBackups();
  });

  it('should load D3FEND preset successfully', () => {
    const store = useGLCStore.getState();
    
    store.setCurrentPreset(D3FEND_PRESET);
    
    expect(store.currentPreset).toBeDefined();
    expect(store.currentPreset?.id).toBe('d3fend');
    expect(store.currentPreset?.nodeTypes.length).toBe(9);
    expect(store.currentPreset?.relationshipTypes.length).toBe(8);
  });

  it('should load Topo-Graph preset successfully', () => {
    const store = useGLCStore.getState();
    
    store.setCurrentPreset(TOPO_PRESET);
    
    expect(store.currentPreset).toBeDefined();
    expect(store.currentPreset?.id).toBe('topo-graph');
    expect(store.currentPreset?.nodeTypes.length).toBe(8);
    expect(store.currentPreset?.relationshipTypes.length).toBe(8);
  });

  it('should maintain store state after preset load', () => {
    const store = useGLCStore.getState();
    
    store.setTheme('light');
    store.setSidebarOpen(false);
    store.setCurrentPreset(D3FEND_PRESET);
    
    expect(useGLCStore.getState().theme).toBe('light');
    expect(useGLCStore.getState().sidebarOpen).toBe(false);
    expect(useGLCStore.getState().currentPreset?.id).toBe('d3fend');
  });
});

describe('Integration - Preset Switching Flow', () => {
  beforeEach(() => {
    useGLCStore.getState().resetPreset();
    useGLCStore.getState().clearGraph();
  });

  afterEach(() => {
    useGLCStore.getState().resetPreset();
    useGLCStore.getState().clearGraph();
  });

  it('should switch from D3FEND to Topo-Graph', () => {
    const store = useGLCStore.getState();
    
    store.setCurrentPreset(D3FEND_PRESET);
    expect(store.currentPreset?.id).toBe('d3fend');
    
    store.setCurrentPreset(TOPO_PRESET);
    expect(store.currentPreset?.id).toBe('topo-graph');
  });

  it('should clear graph when switching presets', () => {
    const store = useGLCStore.getState();
    
    store.setCurrentPreset(D3FEND_PRESET);
    store.addNode({
      id: 'test-node',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test' },
    });
    
    expect(store.nodes).toHaveLength(1);
    
    store.setCurrentPreset(TOPO_PRESET);
    
    const newStore = useGLCStore.getState();
    expect(newStore.currentPreset?.id).toBe('topo-graph');
    expect(newStore.nodes).toHaveLength(0);
  });

  it('should switch to user preset', () => {
    const store = useGLCStore.getState();
    const userPreset = presetManager.createUserPreset(D3FEND_PRESET);
    
    store.setCurrentPreset(userPreset);
    
    expect(store.currentPreset).toBeDefined();
    expect(store.currentPreset?.id).toBe(userPreset.id);
    expect(store.currentPreset?.isBuiltIn).toBe(false);
  });
});

describe('Integration - Graph Operations', () => {
  beforeEach(() => {
    useGLCStore.getState().setCurrentPreset(D3FEND_PRESET);
  });

  afterEach(() => {
    useGLCStore.getState().clearGraph();
  });

  it('should add node with valid type', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test Event' },
    });
    
    expect(store.nodes).toHaveLength(1);
    expect(store.nodes[0].id).toBe('node-1');
    expect(store.nodes[0].type).toBe('event');
  });

  it('should add edge with valid relationship type', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Event 1' },
    });
    
    store.addNode({
      id: 'node-2',
      type: 'artifact',
      position: { x: 200, y: 100 },
      data: { name: 'Artifact 1' },
    });
    
    store.addEdge({
      id: 'edge-1',
      source: 'node-1',
      target: 'node-2',
      type: 'accesses',
      data: {},
    });
    
    expect(store.edges).toHaveLength(1);
    expect(store.edges[0].id).toBe('edge-1');
    expect(store.edges[0].type).toBe('accesses');
  });

  it('should delete node and associated edges', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Event' },
    });
    
    store.addNode({
      id: 'node-2',
      type: 'artifact',
      position: { x: 200, y: 100 },
      data: { name: 'Artifact' },
    });
    
    store.addEdge({
      id: 'edge-1',
      source: 'node-1',
      target: 'node-2',
      type: 'accesses',
      data: {},
    });
    
    expect(store.edges).toHaveLength(1);
    
    store.deleteNode('node-1');
    
    expect(store.nodes).toHaveLength(1);
    expect(store.edges).toHaveLength(0);
  });

  it('should update node data', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test' },
    });
    
    store.updateNode('node-1', { data: { name: 'Updated' } });
    
    expect(store.nodes[0].data.name).toBe('Updated');
  });

  it('should delete edge', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Event' },
    });
    
    store.addNode({
      id: 'node-2',
      type: 'artifact',
      position: { x: 200, y: 100 },
      data: { name: 'Artifact' },
    });
    
    store.addEdge({
      id: 'edge-1',
      source: 'node-1',
      target: 'node-2',
      type: 'accesses',
      data: {},
    });
    
    store.deleteEdge('edge-1');
    
    expect(store.edges).toHaveLength(0);
  });
});

describe('Integration - Undo/Redo with Graph Operations', () => {
  beforeEach(() => {
    useGLCStore.getState().setCurrentPreset(D3FEND_PRESET);
  });

  afterEach(() => {
    useGLCStore.getState().clearGraph();
    useGLCStore.getState().clearHistory();
  });

  it('should push history on node addition', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test' },
    });
    
    expect(store.history).toHaveLength(1);
    expect(store.historyIndex).toBe(0);
    expect(store.canUndo).toBe(true);
    expect(store.canRedo).toBe(false);
  });

  it('should undo node addition', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test' },
    });
    
    expect(store.nodes).toHaveLength(1);
    
    store.undo();
    
    expect(store.nodes).toHaveLength(0);
    expect(store.canUndo).toBe(false);
    expect(store.canRedo).toBe(true);
  });

  it('should redo node addition', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test' },
    });
    
    store.undo();
    expect(store.nodes).toHaveLength(0);
    
    store.redo();
    expect(store.nodes).toHaveLength(1);
  });

  it('should clear history on graph clear', () => {
    const store = useGLCStore.getState();
    
    store.addNode({
      id: 'node-1',
      type: 'event',
      position: { x: 100, y: 100 },
      data: { name: 'Test' },
    });
    
    expect(store.history).toHaveLength(1);
    
    store.clearHistory();
    
    expect(store.history).toHaveLength(0);
    expect(store.canUndo).toBe(false);
    expect(store.canRedo).toBe(false);
  });
});
