import { StateCreator } from 'zustand';

interface CanvasState {
  isCanvasReady: boolean;
  isDragging: boolean;
  isConnecting: boolean;
  connectSource: string | null;
  canvasMode: 'select' | 'drag' | 'connect' | 'pan';
  setCanvasReady: (ready: boolean) => void;
  setDragging: (dragging: boolean) => void;
  setConnecting: (connecting: boolean) => void;
  setConnectSource: (source: string | null) => void;
  setCanvasMode: (mode: 'select' | 'drag' | 'connect' | 'pan') => void;
  resetCanvas: () => void;
}

export const createCanvasSlice: StateCreator<CanvasState> = (set) => ({
  isCanvasReady: false,
  isDragging: false,
  isConnecting: false,
  connectSource: null,
  canvasMode: 'select',
  setCanvasReady: (ready) => set({ isCanvasReady: ready }),
  setDragging: (dragging) => set({ isDragging: dragging }),
  setConnecting: (connecting) => set((state) => ({ 
    isConnecting: connecting,
    connectSource: connecting ? state.connectSource : null,
  })),
  setConnectSource: (source) => set({ connectSource: source }),
  setCanvasMode: (mode) => set({ canvasMode: mode }),
  resetCanvas: () => set({
    isCanvasReady: false,
    isDragging: false,
    isConnecting: false,
    connectSource: null,
    canvasMode: 'select',
  }),
});

export default createCanvasSlice;
