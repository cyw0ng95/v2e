/**
 * GLC Preset Slice
 */

import type { StateCreator } from 'zustand';
import type { PresetSlice, CanvasPreset } from '../types';
import { d3fendPreset, topoPreset } from '../presets';

export const createPresetSlice: StateCreator<PresetSlice, [], [], PresetSlice> = (set) => ({
  currentPreset: null,
  builtInPresets: [d3fendPreset, topoPreset],
  userPresets: [],

  setCurrentPreset: (preset: CanvasPreset) => {
    set({ currentPreset: preset });
  },

  addUserPreset: (preset: CanvasPreset) => {
    set((state) => ({
      userPresets: [...state.userPresets, preset],
    }));
  },

  updateUserPreset: (id: string, updates: Partial<CanvasPreset>) => {
    set((state) => ({
      userPresets: state.userPresets.map((p) =>
        p.meta.id === id ? { ...p, ...updates } : p
      ),
    }));
  },

  removeUserPreset: (id: string) => {
    set((state) => ({
      userPresets: state.userPresets.filter((p) => p.meta.id !== id),
    }));
  },
});
