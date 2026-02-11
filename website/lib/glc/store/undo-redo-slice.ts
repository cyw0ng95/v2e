/**
 * GLC Undo/Redo Slice
 */

import type { StateCreator } from 'zustand';
import type { UndoRedoSlice } from '../types';

const MAX_HISTORY = 100;

export const createUndoRedoSlice: StateCreator<UndoRedoSlice> = (set, get) => ({
  canUndo: false,
  canRedo: false,
  history: [],
  currentIndex: -1,

  pushAction: (action) => {
    set((state) => {
      // Truncate any redo history
      const newHistory = state.history.slice(0, state.currentIndex + 1);

      // Add new action
      newHistory.push({
        ...action,
        timestamp: Date.now(),
      });

      // Limit history size
      if (newHistory.length > MAX_HISTORY) {
        newHistory.shift();
      }

      const newIndex = newHistory.length - 1;

      return {
        history: newHistory,
        currentIndex: newIndex,
        canUndo: newIndex >= 0,
        canRedo: false,
      };
    });
  },

  undo: () => {
    const { history, currentIndex } = get();
    if (currentIndex < 0) return null;

    const action = history[currentIndex];

    set((state) => {
      const newIndex = state.currentIndex - 1;
      return {
        currentIndex: newIndex,
        canUndo: newIndex >= 0,
        canRedo: true,
      };
    });

    return action;
  },

  redo: () => {
    const { history, currentIndex } = get();
    if (currentIndex >= history.length - 1) return null;

    const newIndex = currentIndex + 1;
    const action = history[newIndex];

    set({
      currentIndex: newIndex,
      canUndo: true,
      canRedo: newIndex < history.length - 1,
    });

    return action;
  },

  clearHistory: () => {
    set({
      history: [],
      currentIndex: -1,
      canUndo: false,
      canRedo: false,
    });
  },
});
