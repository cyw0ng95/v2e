/**
 * GLC Canvas Slice
 */

import type { StateCreator } from 'zustand';
import type { CanvasSlice } from '../types';

export const createCanvasSlice: StateCreator<CanvasSlice, [], [], CanvasSlice> = (set) => ({
  selectedNodes: [],
  selectedEdges: [],
  zoom: 1,
  isPanning: false,

  setSelection: (nodes: string[], edges: string[]) => {
    set({ selectedNodes: nodes, selectedEdges: edges });
  },

  clearSelection: () => {
    set({ selectedNodes: [], selectedEdges: [] });
  },

  setZoom: (zoom: number) => {
    set({ zoom: Math.max(0.1, Math.min(4, zoom)) });
  },

  setIsPanning: (isPanning: boolean) => {
    set({ isPanning });
  },
});
