/**
 * GLC Store Slice Unit Tests
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { create } from 'zustand';
import { createGraphSlice, createEmptyGraph, createNode, createEdge } from '../../lib/glc/store/graph-slice';
import { createCanvasSlice } from '../../lib/glc/store/canvas-slice';
import { createUISlice } from '../../lib/glc/store/ui-slice';
import { createUndoRedoSlice } from '../../lib/glc/store/undo-redo-slice';
import type { GraphSlice, CanvasSlice, UISlice, UndoRedoSlice, Graph, CADNode, CADEdge } from '../../lib/glc/types';

describe('Graph Slice', () => {
  describe('createEmptyGraph', () => {
    it('should create a graph with default values', () => {
      const graph = createEmptyGraph('d3fend');

      expect(graph.metadata.id).toBeDefined();
      expect(graph.metadata.id.length).toBe(12);
      expect(graph.metadata.name).toBe('Untitled Graph');
      expect(graph.metadata.presetId).toBe('d3fend');
      expect(graph.metadata.tags).toEqual([]);
      expect(graph.metadata.version).toBe(1);
      expect(graph.nodes).toEqual([]);
      expect(graph.edges).toEqual([]);
    });

    it('should create a graph with custom name', () => {
      const graph = createEmptyGraph('topo', 'My Custom Graph');
      expect(graph.metadata.name).toBe('My Custom Graph');
      expect(graph.metadata.presetId).toBe('topo');
    });
  });

  describe('createNode', () => {
    it('should create a node with required properties', () => {
      const node = createNode('d3fend:Technique', { x: 100, y: 200 });

      expect(node.id).toBeDefined();
      expect(node.id.length).toBe(12);
      expect(node.type).toBe('glc');
      expect(node.position).toEqual({ x: 100, y: 200 });
      expect(node.data.typeId).toBe('d3fend:Technique');
      expect(node.data.label).toBe('New Node');
      expect(node.data.properties).toEqual([]);
      expect(node.data.references).toEqual([]);
    });

    it('should create a node with custom data', () => {
      const node = createNode('custom-type', { x: 50, y: 75 }, {
        label: 'Custom Label',
        notes: 'Some notes',
        d3fendClass: 'D3FENDClass',
      });

      expect(node.data.label).toBe('Custom Label');
      expect(node.data.notes).toBe('Some notes');
      expect(node.data.d3fendClass).toBe('D3FENDClass');
    });
  });

  describe('createEdge', () => {
    it('should create an edge with required properties', () => {
      const edge = createEdge('node-1', 'node-2', 'relates-to');

      expect(edge.id).toBe('edge-node-1-node-2');
      expect(edge.source).toBe('node-1');
      expect(edge.target).toBe('node-2');
      expect(edge.type).toBe('glc');
      expect(edge.data.relationshipId).toBe('relates-to');
    });

    it('should create an edge with custom data', () => {
      const edge = createEdge('node-a', 'node-b', 'submits', {
        label: 'Submits Data',
        notes: 'Edge notes',
      });

      expect(edge.data.label).toBe('Submits Data');
      expect(edge.data.notes).toBe('Edge notes');
    });
  });

  describe('Store actions', () => {
    type TestStore = GraphSlice;
    let store: ReturnType<typeof create<TestStore>>;
    let getState: () => TestStore;

    beforeEach(() => {
      store = create<TestStore>((...args) => ({
        ...createGraphSlice(...args),
      }));
      getState = store.getState;
    });

    it('should set a graph', () => {
      const graph = createEmptyGraph('d3fend');
      getState().setGraph(graph);

      expect(getState().graph).toEqual(graph);
    });

    it('should update metadata', () => {
      const graph = createEmptyGraph('d3fend');
      getState().setGraph(graph);

      getState().updateMetadata({ name: 'Updated Name', tags: ['tag1', 'tag2'] });

      const updated = getState().graph;
      expect(updated?.metadata.name).toBe('Updated Name');
      expect(updated?.metadata.tags).toEqual(['tag1', 'tag2']);
      expect(updated?.metadata.version).toBe(2);
    });

    it('should add a node', () => {
      const graph = createEmptyGraph('d3fend');
      getState().setGraph(graph);

      const node = createNode('test-type', { x: 0, y: 0 });
      getState().addNode(node as CADNode);

      expect(getState().graph?.nodes.length).toBe(1);
      expect(getState().graph?.nodes[0].id).toBe(node.id);
    });

    it('should update a node', () => {
      const graph = createEmptyGraph('d3fend');
      const node = createNode('test-type', { x: 0, y: 0 });
      graph.nodes = [node as CADNode];
      getState().setGraph(graph);

      getState().updateNode(node.id, { label: 'Updated Label', notes: 'Added notes' });

      const updatedNode = getState().graph?.nodes[0];
      expect(updatedNode?.data.label).toBe('Updated Label');
      expect(updatedNode?.data.notes).toBe('Added notes');
    });

    it('should remove a node and its connected edges', () => {
      const graph = createEmptyGraph('d3fend');
      const node1 = createNode('type-a', { x: 0, y: 0 });
      const node2 = createNode('type-b', { x: 100, y: 0 });
      const edge = createEdge(node1.id, node2.id, 'connects');
      graph.nodes = [node1 as CADNode, node2 as CADNode];
      graph.edges = [edge as CADEdge];
      getState().setGraph(graph);

      getState().removeNode(node1.id);

      expect(getState().graph?.nodes.length).toBe(1);
      expect(getState().graph?.edges.length).toBe(0);
    });

    it('should add an edge', () => {
      const graph = createEmptyGraph('d3fend');
      getState().setGraph(graph);

      const edge = createEdge('node-1', 'node-2', 'connects');
      getState().addEdge(edge as CADEdge);

      expect(getState().graph?.edges.length).toBe(1);
      expect(getState().graph?.edges[0].id).toBe(edge.id);
    });

    it('should update an edge', () => {
      const graph = createEmptyGraph('d3fend');
      const edge = createEdge('node-1', 'node-2', 'connects');
      graph.edges = [edge as CADEdge];
      getState().setGraph(graph);

      getState().updateEdge(edge.id, { label: 'Updated Label', notes: 'Edge notes' });

      const updatedEdge = getState().graph?.edges[0];
      expect(updatedEdge?.data.label).toBe('Updated Label');
      expect(updatedEdge?.data.notes).toBe('Edge notes');
    });

    it('should remove an edge', () => {
      const graph = createEmptyGraph('d3fend');
      const edge = createEdge('node-1', 'node-2', 'connects');
      graph.edges = [edge as CADEdge];
      getState().setGraph(graph);

      getState().removeEdge(edge.id);

      expect(getState().graph?.edges.length).toBe(0);
    });

    it('should set viewport', () => {
      const graph = createEmptyGraph('d3fend');
      getState().setGraph(graph);

      const viewport = { x: 100, y: 200, zoom: 1.5 };
      getState().setViewport(viewport);

      expect(getState().graph?.viewport).toEqual(viewport);
    });

    it('should clear graph', () => {
      const graph = createEmptyGraph('d3fend');
      getState().setGraph(graph);
      getState().clearGraph();

      expect(getState().graph).toBeNull();
    });

    it('should handle actions when graph is null', () => {
      // All actions should be no-ops when graph is null
      getState().updateMetadata({ name: 'Test' });
      getState().addNode(createNode('type', { x: 0, y: 0 }) as CADNode);
      getState().updateNode('non-existent', { label: 'Test' });
      getState().removeNode('non-existent');
      getState().addEdge(createEdge('a', 'b', 'rel') as CADEdge);
      getState().updateEdge('non-existent', { label: 'Test' });
      getState().removeEdge('non-existent');
      getState().setViewport({ x: 0, y: 0, zoom: 1 });

      expect(getState().graph).toBeNull();
    });
  });
});

describe('Canvas Slice', () => {
  type TestStore = CanvasSlice;
  let store: ReturnType<typeof create<TestStore>>;
  let getState: () => TestStore;

  beforeEach(() => {
    store = create<TestStore>((...args) => ({
      ...createCanvasSlice(...args),
    }));
    getState = store.getState;
  });

  it('should have default values', () => {
    const state = getState();
    expect(state.selectedNodes).toEqual([]);
    expect(state.selectedEdges).toEqual([]);
    expect(state.zoom).toBe(1);
    expect(state.isPanning).toBe(false);
  });

  it('should set selection', () => {
    getState().setSelection(['node-1', 'node-2'], ['edge-1']);

    expect(getState().selectedNodes).toEqual(['node-1', 'node-2']);
    expect(getState().selectedEdges).toEqual(['edge-1']);
  });

  it('should clear selection', () => {
    getState().setSelection(['node-1'], ['edge-1']);
    getState().clearSelection();

    expect(getState().selectedNodes).toEqual([]);
    expect(getState().selectedEdges).toEqual([]);
  });

  it('should set zoom with bounds', () => {
    getState().setZoom(2);
    expect(getState().zoom).toBe(2);

    // Min zoom
    getState().setZoom(0);
    expect(getState().zoom).toBe(0.1);

    // Max zoom
    getState().setZoom(10);
    expect(getState().zoom).toBe(4);
  });

  it('should set panning state', () => {
    getState().setIsPanning(true);
    expect(getState().isPanning).toBe(true);

    getState().setIsPanning(false);
    expect(getState().isPanning).toBe(false);
  });
});

describe('UI Slice', () => {
  type TestStore = UISlice;
  let store: ReturnType<typeof create<TestStore>>;
  let getState: () => TestStore;

  beforeEach(() => {
    store = create<TestStore>((...args) => ({
      ...createUISlice(...args),
    }));
    getState = store.getState;
  });

  it('should have default values', () => {
    const state = getState();
    expect(state.theme).toBe('dark');
    expect(state.sidebarOpen).toBe(true);
    expect(state.nodePaletteOpen).toBe(true);
    expect(state.detailsPanelOpen).toBe(false);
    expect(state.detailsPanelTab).toBe('properties');
  });

  it('should set theme', () => {
    getState().setTheme('light');
    expect(getState().theme).toBe('light');

    getState().setTheme('system');
    expect(getState().theme).toBe('system');
  });

  it('should toggle sidebar', () => {
    const initial = getState().sidebarOpen;
    getState().toggleSidebar();
    expect(getState().sidebarOpen).toBe(!initial);
    getState().toggleSidebar();
    expect(getState().sidebarOpen).toBe(initial);
  });

  it('should toggle node palette', () => {
    const initial = getState().nodePaletteOpen;
    getState().toggleNodePalette();
    expect(getState().nodePaletteOpen).toBe(!initial);
  });

  it('should set details panel open state', () => {
    getState().setDetailsPanelOpen(true);
    expect(getState().detailsPanelOpen).toBe(true);

    getState().setDetailsPanelOpen(false);
    expect(getState().detailsPanelOpen).toBe(false);
  });

  it('should set details panel tab', () => {
    getState().setDetailsPanelTab('references');
    expect(getState().detailsPanelTab).toBe('references');

    getState().setDetailsPanelTab('notes');
    expect(getState().detailsPanelTab).toBe('notes');
  });
});

describe('Undo/Redo Slice', () => {
  type TestStore = UndoRedoSlice;
  let store: ReturnType<typeof create<TestStore>>;
  let getState: () => TestStore;

  beforeEach(() => {
    store = create<TestStore>((...args) => ({
      ...createUndoRedoSlice(...args),
    }));
    getState = store.getState;
  });

  it('should have default values', () => {
    const state = getState();
    expect(state.canUndo).toBe(false);
    expect(state.canRedo).toBe(false);
    expect(state.history).toEqual([]);
    expect(state.currentIndex).toBe(-1);
  });

  it('should push action and update canUndo', () => {
    getState().pushAction({ type: 'test', before: {}, after: {} });

    expect(getState().history.length).toBe(1);
    expect(getState().currentIndex).toBe(0);
    expect(getState().canUndo).toBe(true);
    expect(getState().canRedo).toBe(false);
  });

  it('should undo and redo', () => {
    getState().pushAction({ type: 'action1', before: { a: 1 }, after: { a: 2 } });
    getState().pushAction({ type: 'action2', before: { b: 1 }, after: { b: 2 } });

    // Undo first action
    const undo1 = getState().undo();
    expect(undo1?.type).toBe('action2');
    expect(getState().currentIndex).toBe(0);
    expect(getState().canUndo).toBe(true);
    expect(getState().canRedo).toBe(true);

    // Undo second action
    const undo2 = getState().undo();
    expect(undo2?.type).toBe('action1');
    expect(getState().currentIndex).toBe(-1);
    expect(getState().canUndo).toBe(false);
    expect(getState().canRedo).toBe(true);

    // Redo
    const redo1 = getState().redo();
    expect(redo1?.type).toBe('action1');
    expect(getState().canUndo).toBe(true);
    expect(getState().canRedo).toBe(true);

    // Redo again
    const redo2 = getState().redo();
    expect(redo2?.type).toBe('action2');
    expect(getState().canRedo).toBe(false);
  });

  it('should return null when cannot undo/redo', () => {
    // Cannot undo empty history
    expect(getState().undo()).toBeNull();

    // Cannot redo at latest
    getState().pushAction({ type: 'test', before: {}, after: {} });
    expect(getState().redo()).toBeNull();
  });

  it('should clear history', () => {
    getState().pushAction({ type: 'test1', before: {}, after: {} });
    getState().pushAction({ type: 'test2', before: {}, after: {} });
    getState().undo();

    getState().clearHistory();

    expect(getState().history).toEqual([]);
    expect(getState().currentIndex).toBe(-1);
    expect(getState().canUndo).toBe(false);
    expect(getState().canRedo).toBe(false);
  });

  it('should truncate redo history on new action', () => {
    getState().pushAction({ type: 'action1', before: {}, after: {} });
    getState().pushAction({ type: 'action2', before: {}, after: {} });
    getState().undo(); // Now at action1

    // Push new action should remove action2
    getState().pushAction({ type: 'action3', before: {}, after: {} });

    expect(getState().history.length).toBe(2);
    expect(getState().history[0].type).toBe('action1');
    expect(getState().history[1].type).toBe('action3');
    expect(getState().canRedo).toBe(false);
  });

  it('should limit history size', () => {
    // Push more than MAX_HISTORY (100) actions
    for (let i = 0; i < 105; i++) {
      getState().pushAction({ type: `action${i}`, before: {}, after: {} });
    }

    expect(getState().history.length).toBe(100);
    // First action should be removed
    expect(getState().history[0].type).toBe('action5');
  });

  it('should add timestamp to actions', () => {
    const beforeTime = Date.now();
    getState().pushAction({ type: 'test', before: {}, after: {} });
    const afterTime = Date.now();

    const action = getState().history[0];
    expect(action.timestamp).toBeGreaterThanOrEqual(beforeTime);
    expect(action.timestamp).toBeLessThanOrEqual(afterTime);
  });
});
