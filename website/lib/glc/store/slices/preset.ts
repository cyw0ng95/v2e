import { StateCreator } from 'zustand';
import { CanvasPreset } from '../../types';
import { BUILT_IN_PRESETS } from '../../presets';

interface PresetState {
  currentPreset: CanvasPreset | null;
  builtInPresets: CanvasPreset[];
  userPresets: CanvasPreset[];
  setCurrentPreset: (preset: CanvasPreset) => void;
  addUserPreset: (preset: CanvasPreset) => void;
  updateUserPreset: (id: string, preset: Partial<CanvasPreset>) => void;
  deleteUserPreset: (id: string) => void;
  getAllPresets: () => CanvasPreset[];
  getPresetById: (id: string) => CanvasPreset | undefined;
  resetPreset: () => void;
}

export const createPresetSlice: StateCreator<PresetState> = (set, get) => ({
  currentPreset: null,
  builtInPresets: BUILT_IN_PRESETS,
  userPresets: [],
  setCurrentPreset: (preset) => set({ currentPreset: preset }),
  addUserPreset: (preset) => set((state) => ({ userPresets: [...state.userPresets, preset] })),
  updateUserPreset: (id, preset) => set((state) => ({
    userPresets: state.userPresets.map(p => p.id === id ? { ...p, ...preset } : p),
    currentPreset: state.currentPreset?.id === id ? { ...state.currentPreset, ...preset } : state.currentPreset,
  })),
  deleteUserPreset: (id) => set((state) => ({
    userPresets: state.userPresets.filter(p => p.id !== id),
    currentPreset: state.currentPreset?.id === id ? null : state.currentPreset,
  })),
  getAllPresets: () => {
    const { builtInPresets, userPresets } = get();
    return [...builtInPresets, ...userPresets];
  },
  getPresetById: (id) => {
    const { builtInPresets, userPresets, currentPreset } = get();
    return currentPreset?.id === id ? currentPreset : 
           builtInPresets.find(p => p.id === id) || 
           userPresets.find(p => p.id === id);
  },
  resetPreset: () => set({ currentPreset: null }),
});

export default createPresetSlice;
