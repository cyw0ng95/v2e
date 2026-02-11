/**
 * Tests for D3FEND Ontology Loader
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  loadD3FENDOntology,
  getD3FENDOntology,
  getOntologyLoadState,
  getClassById,
  getClassChildren,
  getClassAncestors,
  buildTreeNodes,
  searchClasses,
  resetLoader,
} from '@/lib/glc/d3fend/loader';

// Mock fetch for full ontology loading
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('D3FEND Loader', () => {
  beforeEach(() => {
    resetLoader();
    mockFetch.mockReset();
  });

  describe('loadD3FENDOntology', () => {
    it('should return simplified data when full ontology is not available', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));

      const data = await loadD3FENDOntology();

      expect(data).toBeDefined();
      expect(data.version).toBe('1.0.0-simplified');
      expect(data.classes).toBeDefined();
      expect(data.rootClassIds).toContain('d3f:DefensiveTechnique');
    });

    it('should load full ontology when available', async () => {
      const mockData = {
        version: '2.0.0',
        updatedAt: '2024-01-01T00:00:00Z',
        classes: {
          'd3f:TestClass': {
            id: 'd3f:TestClass',
            label: 'Test Class',
            children: [],
          },
        },
        techniques: {},
        relationships: [],
        rootClassIds: ['d3f:TestClass'],
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockData),
      });

      const data = await loadD3FENDOntology();

      expect(data.version).toBe('2.0.0');
      expect(data.classes['d3f:TestClass']).toBeDefined();
    });

    it('should cache loaded data', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));

      await loadD3FENDOntology();
      await loadD3FENDOntology();

      // Fetch should only be called once due to caching
      expect(mockFetch).toHaveBeenCalledTimes(1);
    });
  });

  describe('getOntologyLoadState', () => {
    it('should return idle initially', () => {
      expect(getOntologyLoadState()).toBe('idle');
    });

    it('should return loaded after loading', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      expect(getOntologyLoadState()).toBe('loaded');
    });
  });

  describe('getD3FENDOntology', () => {
    it('should return null initially', () => {
      expect(getD3FENDOntology()).toBeNull();
    });

    it('should return data after loading', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const data = getD3FENDOntology();
      expect(data).not.toBeNull();
      expect(data?.classes).toBeDefined();
    });
  });

  describe('getClassById', () => {
    it('should return class data after loading', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const cls = getClassById('d3f:Hardening');
      expect(cls).not.toBeNull();
      expect(cls?.label).toBe('Hardening');
    });

    it('should return null for unknown class', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const cls = getClassById('d3f:Unknown');
      expect(cls).toBeNull();
    });
  });

  describe('getClassChildren', () => {
    it('should return children of a class', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const children = getClassChildren('d3f:DefensiveTechnique');
      expect(children.length).toBeGreaterThan(0);
      expect(children.some((c) => c.id === 'd3f:Hardening')).toBe(true);
    });

    it('should return empty array for leaf class', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const children = getClassChildren('d3f:Isolation');
      expect(children).toEqual([]);
    });
  });

  describe('getClassAncestors', () => {
    it('should return ancestors of a class', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const ancestors = getClassAncestors('d3f:ApplicationHardening');
      expect(ancestors.length).toBeGreaterThan(0);
      expect(ancestors.some((c) => c.id === 'd3f:Hardening')).toBe(true);
    });

    it('should return empty array for root class', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const ancestors = getClassAncestors('d3f:DefensiveTechnique');
      expect(ancestors).toEqual([]);
    });
  });

  describe('buildTreeNodes', () => {
    it('should build tree nodes from simplified data', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const nodes = buildTreeNodes();
      expect(nodes.length).toBeGreaterThan(0);
      expect(nodes[0].depth).toBe(0);
    });

    it('should respect expanded nodes', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const expanded = new Set(['d3f:DefensiveTechnique']);
      const nodes = buildTreeNodes(undefined, expanded);

      // Should include children when expanded
      expect(nodes.some((n) => n.parentId === 'd3f:DefensiveTechnique')).toBe(true);
    });

    it('should filter by search query', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const nodes = buildTreeNodes(undefined, new Set(), 'Hardening');
      expect(nodes.length).toBeGreaterThan(0);
      expect(nodes.every((n) => n.label.toLowerCase().includes('hardening'))).toBe(true);
    });
  });

  describe('searchClasses', () => {
    it('should find classes by label', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const results = searchClasses('Hardening');
      expect(results.length).toBeGreaterThan(0);
      expect(results.some((c) => c.label.includes('Hardening'))).toBe(true);
    });

    it('should find classes by ID', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const results = searchClasses('d3f:Detection');
      expect(results.length).toBeGreaterThan(0);
    });

    it('should respect limit parameter', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Not found'));
      await loadD3FENDOntology();

      const results = searchClasses('', 2);
      expect(results.length).toBeLessThanOrEqual(2);
    });
  });
});
