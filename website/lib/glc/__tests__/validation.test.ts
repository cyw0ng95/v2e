import { describe, it, expect } from '@jest/globals';
import { validatePreset, validateGraph } from '../validation';
import { D3FEND_PRESET } from '../presets/d3fend-preset';
import { Graph } from '../types';

describe('Validation - Preset Validation', () => {
  it('should validate D3FEND preset successfully', () => {
    const result = validatePreset(D3FEND_PRESET);
    
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should fail validation for missing required fields', () => {
    const invalidPreset = {
      id: 'test',
    };
    
    const result = validatePreset(invalidPreset);
    
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThan(0);
  });

  it('should validate preset with warnings', () => {
    const presetWithWarnings = {
      ...D3FEND_PRESET,
      nodeTypes: [
        {
          ...D3FEND_PRESET.nodeTypes[0],
          id: '',
        },
      ],
    };
    
    const result = validatePreset(presetWithWarnings);
    
    expect(result.warnings.length).toBeGreaterThan(0);
  });
});

describe('Validation - Graph Validation', () => {
  it('should validate valid graph', () => {
    const validGraph: Graph = {
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
      nodes: [
        {
          id: 'node-1',
          type: 'event',
          position: { x: 100, y: 100 },
          data: { name: 'Test Event' },
        },
      ],
      edges: [],
    };
    
    const result = validateGraph(validGraph, D3FEND_PRESET);
    
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should fail validation for invalid node type', () => {
    const invalidGraph: Graph = {
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
      nodes: [
        {
          id: 'node-1',
          type: 'invalid-type',
          position: { x: 100, y: 100 },
          data: { name: 'Test' },
        },
      ],
      edges: [],
    };
    
    const result = validateGraph(invalidGraph, D3FEND_PRESET);
    
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThan(0);
  });

  it('should fail validation for duplicate node IDs', () => {
    const duplicateNodeGraph: Graph = {
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
      nodes: [
        {
          id: 'node-1',
          type: 'event',
          position: { x: 100, y: 100 },
          data: { name: 'Test 1' },
        },
        {
          id: 'node-1',
          type: 'countermeasure',
          position: { x: 200, y: 200 },
          data: { name: 'Test 2' },
        },
      ],
      edges: [],
    };
    
    const result = validateGraph(duplicateNodeGraph, D3FEND_PRESET);
    
    expect(result.valid).toBe(false);
    expect(result.errors.some(e => e.code === 'DUPLICATE_NODE_ID')).toBe(true);
  });

  it('should warn for self-loop edges', () => {
    const selfLoopGraph: Graph = {
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
      nodes: [
        {
          id: 'node-1',
          type: 'event',
          position: { x: 100, y: 100 },
          data: { name: 'Test' },
        },
      ],
      edges: [
        {
          id: 'edge-1',
          source: 'node-1',
          target: 'node-1',
          type: 'accesses',
          data: {},
        },
      ],
    };
    
    const result = validateGraph(selfLoopGraph, D3FEND_PRESET);
    
    expect(result.warnings.some(w => w.code === 'SELF_LOOP_EDGE')).toBe(true);
  });
});
