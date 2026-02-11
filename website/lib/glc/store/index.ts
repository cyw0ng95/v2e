/**
 * GLC Zustand Store
 *
 * Combines all slices with persistence and devtools.
 */

import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';
import { createPresetSlice } from './preset-slice';
import { createGraphSlice } from './graph-slice';
import { createCanvasSlice } from './canvas-slice';
import { createUISlice } from './ui-slice';
import { createUndoRedoSlice } from './undo-redo-slice';
import type { PresetSlice, GraphSlice, CanvasSlice, UISlice, UndoRedoSlice } from '../types';

// Combined store type
export type GLCStore = PresetSlice & GraphSlice & CanvasSlice & UISlice & UndoRedoSlice;

// Create the combined store
export const useGLCStore = create<GLCStore>()(
  devtools(
    persist(
      (...args) => ({
        ...createPresetSlice(...args),
        ...createGraphSlice(...args),
        ...createCanvasSlice(...args),
        ...createUISlice(...args),
        ...createUndoRedoSlice(...args),
      }),
      {
        name: 'glc-storage',
        partialize: (state) => ({
          // Only persist user presets and UI preferences
          userPresets: state.userPresets,
          theme: state.theme,
          sidebarOpen: state.sidebarOpen,
          nodePaletteOpen: state.nodePaletteOpen,
        }),
      }
    ),
    {
      name: 'GLC Store',
    }
  )
);

// Re-export helper functions
export { createEmptyGraph, createNode, createEdge } from './graph-slice';
