import { describe, it, expect } from '@jest/globals';
import {
  generateId,
  findNodeTypeById,
  getValidRelationshipTypes,
  validateNodePosition,
  getDefaultNodePosition,
  isEmpty,
  cloneDeep,
} from '../utils';
import { D3FEND_PRESET } from '../presets/d3fend-preset';

describe('Utils - ID Generation', () => {
  it('should generate unique IDs', () => {
    const id1 = generateId();
    const id2 = generateId();
    
    expect(id1).not.toBe(id2);
  });

  it('should generate IDs with timestamp', () => {
    const id = generateId();
    const parts = id.split('-');
    
    expect(parts.length).toBe(2);
    expect(parts[0]).not.toBeNaN();
  });
});

describe('Utils - Node Type Lookup', () => {
  it('should find node type by ID', () => {
    const nodeType = findNodeTypeById(D3FEND_PRESET, 'event');
    
    expect(nodeType).toBeDefined();
    expect(nodeType?.id).toBe('event');
  });

  it('should return undefined for non-existent node type', () => {
    const nodeType = findNodeTypeById(D3FEND_PRESET, 'non-existent');
    
    expect(nodeType).toBeUndefined();
  });
});

describe('Utils - Relationship Type Lookup', () => {
  it('should find valid relationship types between node types', () => {
    const relTypes = getValidRelationshipTypes(D3FEND_PRESET, 'event', 'artifact');
    
    expect(relTypes.length).toBeGreaterThan(0);
    expect(relTypes.some(rt => rt.id === 'accesses')).toBe(true);
  });

  it('should find relationship types with wildcard source', () => {
    const relTypes = getValidRelationshipTypes(D3FEND_PRESET, 'event', 'artifact');
    
    expect(relTypes.length).toBeGreaterThan(0);
  });
});

describe('Utils - Position Validation', () => {
  it('should validate node position without overlap', () => {
    const nodes = [
      { id: '1', type: 'event', position: { x: 100, y: 100 }, data: {} },
    ];
    
    const isValid = validateNodePosition(nodes, { x: 200, y: 200 });
    
    expect(isValid).toBe(true);
  });

  it('should reject overlapping position', () => {
    const nodes = [
      { id: '1', type: 'event', position: { x: 100, y: 100 }, data: {} },
    ];
    
    const isValid = validateNodePosition(nodes, { x: 105, y: 105 });
    
    expect(isValid).toBe(false);
  });

  it('should validate empty node list', () => {
    const isValid = validateNodePosition([], { x: 100, y: 100 });
    
    expect(isValid).toBe(true);
  });
});

describe('Utils - Default Node Position', () => {
  it('should return default position for empty node list', () => {
    const position = getDefaultNodePosition([]);
    
    expect(position).toEqual({ x: 400, y: 300 });
  });

  it('should return position next to last node', () => {
    const nodes = [
      { id: '1', type: 'event', position: { x: 100, y: 100 }, data: {} },
    ];
    
    const position = getDefaultNodePosition(nodes);
    
    expect(position).toEqual({ x: 250, y: 100 });
  });
});

describe('Utils - Empty Check', () => {
  it('should detect empty string', () => {
    expect(isEmpty('')).toBe(true);
    expect(isEmpty('   ')).toBe(true);
    expect(isEmpty('test')).toBe(false);
  });

  it('should detect empty array', () => {
    expect(isEmpty([])).toBe(true);
    expect(isEmpty([1, 2, 3])).toBe(false);
  });

  it('should detect empty object', () => {
    expect(isEmpty({})).toBe(true);
    expect(isEmpty({ key: 'value' })).toBe(false);
  });

  it('should detect null and undefined', () => {
    expect(isEmpty(null)).toBe(true);
    expect(isEmpty(undefined)).toBe(true);
    expect(isEmpty(0)).toBe(false);
  });
});

describe('Utils - Deep Clone', () => {
  it('should clone object deeply', () => {
    const obj = {
      a: 1,
      b: {
        c: 2,
        d: [3, 4],
      },
    };
    
    const cloned = cloneDeep(obj);
    
    expect(cloned).toEqual(obj);
    expect(cloned).not.toBe(obj);
    expect(cloned.b).not.toBe(obj.b);
    expect(cloned.b.d).not.toBe(obj.b.d);
  });

  it('should clone array deeply', () => {
    const arr = [1, { a: 2 }, [3, 4]];
    
    const cloned = cloneDeep(arr);
    
    expect(cloned).toEqual(arr);
    expect(cloned).not.toBe(arr);
    expect(cloned[1]).not.toBe(arr[1]);
  });
});
