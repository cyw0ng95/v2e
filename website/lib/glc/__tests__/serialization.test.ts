import { describe, it, expect } from '@jest/globals';
import {
  serializePreset,
  deserializePreset,
  serializeGraph,
  deserializeGraph,
  validateBeforeSave,
} from '../preset-serializer';
import { D3FEND_PRESET } from '../presets/d3fend-preset';
import { Graph } from '../types';

describe('Serializer - Preset Serialization', () => {
  it('should serialize preset to JSON-compatible object', () => {
    const serialized = serializePreset(D3FEND_PRESET);
    
    expect(serialized).toBeDefined();
    expect(serialized.id).toBe(D3FEND_PRESET.id);
    expect(serialized.name).toBe(D3FEND_PRESET.name);
    expect(serialized.version).toBe(D3FEND_PRESET.version);
  });

  it('should handle circular references safely', () => {
    const preset = { ...D3FEND_PRESET };
    (preset as any).self = preset;
    
    const serialized = serializePreset(preset);
    
    expect(serialized).toBeDefined();
    expect(serialized.id).toBe(preset.id);
  });

  it('should throw error for invalid preset structure', () => {
    const invalidPreset = { id: null };
    
    expect(() => serializePreset(invalidPreset as any)).toThrow('SerializationError');
  });
});

describe('Serializer - Preset Deserialization', () => {
  it('should deserialize preset from JSON string', async () => {
    const json = JSON.stringify(D3FEND_PRESET);
    const deserialized = await deserializePreset(json);
    
    expect(deserialized).toBeDefined();
    expect(deserialized.id).toBe(D3FEND_PRESET.id);
    expect(deserialized.name).toBe(D3FEND_PRESET.name);
  });

  it('should deserialize preset from object', async () => {
    const deserialized = await deserializePreset(D3FEND_PRESET);
    
    expect(deserialized).toBeDefined();
    expect(deserialized.id).toBe(D3FEND_PRESET.id);
  });

  it('should throw error for invalid JSON string', async () => {
    await expect(deserializePreset('invalid json')).rejects.toThrow('SerializationError');
  });

  it('should throw error for non-object input', async () => {
    await expect(deserializePreset(null)).rejects.toThrow('SerializationError');
  });

  it('should throw error for missing required fields', async () => {
    const invalidPreset = { name: 'Test' };
    
    await expect(deserializePreset(invalidPreset)).rejects.toThrow('SerializationError');
  });

  it('should add default version if missing', async () => {
    const presetWithoutVersion = { ...D3FEND_PRESET };
    delete (presetWithoutVersion as any).version;
    
    const deserialized = await deserializePreset(presetWithoutVersion);
    
    expect(deserialized.version).toBeDefined();
  });
});

describe('Serializer - Graph Serialization', () => {
  it('should serialize graph to JSON-compatible object', () => {
    const graph: Graph = {
      metadata: {
        id: 'test-graph',
        name: 'Test Graph',
        description: 'Test',
        presetId: 'd3fend',
        version: 1,
        createdAt: '2026-02-09',
        updatedAt: '2026-02-09',
        author: 'Test',
        tags: [],
        isPublic: false,
      },
      nodes: [],
      edges: [],
    };
    
    const serialized = serializeGraph(graph);
    
    expect(serialized).toBeDefined();
    expect(serialized.metadata.id).toBe('test-graph');
  });

  it('should throw error for invalid graph structure', () => {
    const invalidGraph = { metadata: { id: null } };
    
    expect(() => serializeGraph(invalidGraph as any)).toThrow('SerializationError');
  });
});

describe('Serializer - Graph Deserialization', () => {
  it('should deserialize graph from JSON string', async () => {
    const graph: Graph = {
      metadata: {
        id: 'test-graph',
        name: 'Test',
        description: 'Test',
        presetId: 'd3fend',
        version: 1,
        createdAt: '2026-02-09',
        updatedAt: '2026-02-09',
        author: 'Test',
        tags: [],
        isPublic: false,
      },
      nodes: [],
      edges: [],
    };
    const json = JSON.stringify(graph);
    
    const deserialized = await deserializeGraph(json);
    
    expect(deserialized).toBeDefined();
    expect(deserialized.metadata.id).toBe('test-graph');
  });

  it('should deserialize graph from object', async () => {
    const graph: Graph = {
      metadata: {
        id: 'test-graph',
        name: 'Test',
        description: 'Test',
        presetId: 'd3fend',
        version: 1,
        createdAt: '2026-02-09',
        updatedAt: '2026-02-09',
        author: 'Test',
        tags: [],
        isPublic: false,
      },
      nodes: [],
      edges: [],
    };
    
    const deserialized = await deserializeGraph(graph);
    
    expect(deserialized).toBeDefined();
    expect(deserialized.metadata.id).toBe('test-graph');
  });

  it('should throw error for invalid JSON string', async () => {
    await expect(deserializeGraph('invalid json')).rejects.toThrow('SerializationError');
  });

  it('should throw error for non-object input', async () => {
    await expect(deserializeGraph(null)).rejects.toThrow('SerializationError');
  });

  it('should throw error for missing metadata', async () => {
    const invalidGraph = { nodes: [], edges: [] };
    
    await expect(deserializeGraph(invalidGraph)).rejects.toThrow('SerializationError');
  });

  it('should initialize empty arrays if missing', async () => {
    const graph: any = {
      metadata: {
        id: 'test',
        name: 'Test',
        description: 'Test',
        presetId: 'd3fend',
        version: 1,
        createdAt: '2026-02-09',
        updatedAt: '2026-02-09',
        author: 'Test',
        tags: [],
        isPublic: false,
      },
    };
    
    const deserialized = await deserializeGraph(graph);
    
    expect(Array.isArray(deserialized.nodes)).toBe(true);
    expect(Array.isArray(deserialized.edges)).toBe(true);
  });
});

describe('Serializer - Validation', () => {
  it('should validate valid object', () => {
    const valid = { test: 'value' };
    const result = validateBeforeSave(valid);
    
    expect(result).toBe(true);
  });

  it('should validate primitive values', () => {
    expect(validateBeforeSave('string')).toBe(true);
    expect(validateBeforeSave(123)).toBe(true);
    expect(validateBeforeSave(true)).toBe(true);
    expect(validateBeforeSave(null)).toBe(true);
  });

  it('should validate arrays', () => {
    const array = [1, 2, 3];
    const result = validateBeforeSave(array);
    
    expect(result).toBe(true);
  });

  it('should reject circular references', () => {
    const circular: any = { name: 'test' };
    circular.self = circular;
    
    const result = validateBeforeSave(circular);
    
    expect(result).toBe(false);
  });

  it('should reject undefined', () => {
    const result = validateBeforeSave(undefined);
    
    expect(result).toBe(false);
  });

  it('should reject functions', () => {
    const func = () => {};
    const result = validateBeforeSave(func);
    
    expect(result).toBe(false);
  });

  it('should reject symbols', () => {
    const symbol = Symbol('test');
    const result = validateBeforeSave(symbol);
    
    expect(result).toBe(false);
  });
});
