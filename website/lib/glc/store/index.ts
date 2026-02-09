import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { createPresetSlice, PresetState } from './slices/preset';
import { createGraphSlice, GraphState } from './slices/graph';
import { createCanvasSlice, CanvasState } from './slices/canvas';
import { createUISlice, UIState } from './slices/ui';
import { createUndoRedoSlice, UndoRedoState } from './slices/undo-redo';

export interface GLCStore extends PresetState, GraphState, CanvasState, UIState, UndoRedoState {}

export const useGLCStore = create<GLCStore>()(
  devtools(
    persist(
      (...a) => ({
        ...createPresetSlice(...a),
        ...createGraphSlice(...a),
        ...createCanvasSlice(...a),
        ...createUISlice(...a),
        ...createUndoRedoSlice(...a),
      }),
      {
        name: 'glc-storage',
        partialize: (state) => ({
          theme: state.theme,
          sidebarOpen: state.sidebarOpen,
          showMiniMap: state.showMiniMap,
          showControls: state.showControls,
          userPresets: state.userPresets,
        }),
      }
    ),
    { name: 'GLCStore' }
  )
);

export default useGLCStore;
