import { CanvasPreset } from '../types';
import D3FEND_PRESET from './d3fend-preset';
import TOPO_PRESET from './topo-preset';

export const BUILT_IN_PRESETS: CanvasPreset[] = [
  D3FEND_PRESET,
  TOPO_PRESET,
];

export const getPresetById = (id: string): CanvasPreset | undefined => {
  return BUILT_IN_PRESETS.find(preset => preset.id === id);
};

export const getAllPresets = (): CanvasPreset[] => {
  return BUILT_IN_PRESETS;
};

export const getPresetsByCategory = (category: string): CanvasPreset[] => {
  return BUILT_IN_PRESETS.filter(preset => preset.category === category);
};

export const searchPresets = (query: string): CanvasPreset[] => {
  const lowerQuery = query.toLowerCase();
  return BUILT_IN_PRESETS.filter(preset =>
    preset.name.toLowerCase().includes(lowerQuery) ||
    preset.description.toLowerCase().includes(lowerQuery) ||
    preset.metadata.tags.some(tag => tag.toLowerCase().includes(lowerQuery))
  );
};

export default {
  BUILT_IN_PRESETS,
  getPresetById,
  getAllPresets,
  getPresetsByCategory,
  searchPresets,
};
