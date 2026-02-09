import { StateCreator } from 'zustand';
import { CADNode, CADEdge, Graph } from '../../types';

interface HistoryItem {
  nodes: CADNode[];
  edges: CADEdge[];
  metadata?: Graph['metadata'];
}

interface UndoRedoState {
  history: HistoryItem[];
  historyIndex: number;
  maxHistorySize: number;
  canUndo: boolean;
  canRedo: boolean;
  pushHistory: (nodes: CADNode[], edges: CADEdge[], metadata?: Graph['metadata']) => void;
  undo: () => void;
  redo: () => void;
  clearHistory: () => void;
  getCurrentState: () => HistoryItem | null;
}

export const createUndoRedoSlice: StateCreator<UndoRedoState> = (set, get) => ({
  history: [],
  historyIndex: -1,
  maxHistorySize: 50,
  canUndo: false,
  canRedo: false,
  pushHistory: (nodes, edges, metadata) => set((state) => {
    const newItem: HistoryItem = { nodes, edges, metadata };
    let newHistory;
    
    if (state.historyIndex < state.history.length - 1) {
      newHistory = [...state.history.slice(0, state.historyIndex + 1), newItem];
    } else {
      newHistory = [...state.history, newItem];
    }
    
    if (newHistory.length > state.maxHistorySize) {
      newHistory = newHistory.slice(-state.maxHistorySize);
    }
    
    return {
      history: newHistory,
      historyIndex: newHistory.length - 1,
      canUndo: newHistory.length > 0,
      canRedo: false,
    };
  }),
  undo: () => set((state) => {
    if (state.historyIndex <= 0) return state;
    
    return {
      historyIndex: state.historyIndex - 1,
      canUndo: state.historyIndex > 0,
      canRedo: true,
    };
  }),
  redo: () => set((state) => {
    if (state.historyIndex >= state.history.length - 1) return state;
    
    return {
      historyIndex: state.historyIndex + 1,
      canUndo: true,
      canRedo: state.historyIndex < state.history.length - 2,
    };
  }),
  clearHistory: () => set({
    history: [],
    historyIndex: -1,
    canUndo: false,
    canRedo: false,
  }),
  getCurrentState: () => {
    const { history, historyIndex } = get();
    if (historyIndex < 0 || historyIndex >= history.length) {
      return null;
    }
    return history[historyIndex];
  },
});

export default createUndoRedoSlice;
