import { describe, it, expect } from '@jest/globals';
import { useGLCStore } from '../store';

describe('GLC Store - Preset Slice', () => {
  it('should initialize with default preset state', () => {
    const store = useGLCStore.getState();
    
    expect(store.currentPreset).toBeNull();
    expect(store.builtInPresets).toHaveLength(2);
    expect(store.userPresets).toHaveLength(0);
  });

  it('should set current preset', () => {
    const store = useGLCStore.getState();
    const preset = store.builtInPresets[0];
    
    store.setCurrentPreset(preset);
    
    expect(useGLCStore.getState().currentPreset).toEqual(preset);
  });

  it('should add user preset', () => {
    const store = useGLCStore.getState();
    const newPreset = {
      id: 'test-preset',
      name: 'Test Preset',
      version: '1.0.0',
      category: 'Test',
      description: 'Test preset',
      author: 'Test',
      createdAt: '2026-02-09',
      updatedAt: '2026-02-09',
      isBuiltIn: false,
      nodeTypes: [],
      relationshipTypes: [],
      styling: {
        theme: 'light' as const,
        primaryColor: '#000000',
        backgroundColor: '#ffffff',
        gridColor: '#e5e7eb',
        fontFamily: 'Inter',
      },
      behavior: {
        pan: true,
        zoom: true,
        snapToGrid: false,
        gridSize: 10,
        undoRedo: true,
        autoSave: false,
        autoSaveInterval: 60000,
        maxNodes: 1000,
        maxEdges: 2000,
      },
      validationRules: [],
      metadata: {
        tags: [],
      },
    };
    
    store.addUserPreset(newPreset);
    
    expect(useGLCStore.getState().userPresets).toHaveLength(1);
    expect(useGLCStore.getState().userPresets[0].id).toBe('test-preset');
  });

  it('should get preset by ID', () => {
    const store = useGLCStore.getState();
    const d3fendPreset = store.getPresetById('d3fend');
    
    expect(d3fendPreset).toBeDefined();
    expect(d3fendPreset?.name).toBe('D3FEND Cyber Defense Modeling');
  });

  it('should get all presets', () => {
    const store = useGLCStore.getState();
    const allPresets = store.getAllPresets();
    
    expect(allPresets.length).toBeGreaterThanOrEqual(2);
  });
});
