/**
 * STIX Import Engine Tests
 */

import { describe, it, expect } from 'vitest';
import {
  STIXImportEngine,
  createSTIXImportEngine,
  importSTIX,
  validateSTIX,
} from '@/lib/glc/stix/import-engine';

// ============================================================================
// Test Data
// ============================================================================

const VALID_STIX_BUNDLE = {
  type: 'bundle',
  id: 'bundle--12345678-1234-1234-1234-123456789012',
  spec_version: '2.1',
  objects: [
    {
      type: 'attack-pattern',
      id: 'attack-pattern--abc123',
      created: '2024-01-01T00:00:00.000Z',
      modified: '2024-01-01T00:00:00.000Z',
      name: 'Spear Phishing',
      description: 'Targeted phishing attack',
      kill_chain_phases: [
        { kill_chain_name: 'mitre-attack', phase_name: 'initial-access' },
      ],
    },
    {
      type: 'identity',
      id: 'identity--def456',
      created: '2024-01-01T00:00:00.000Z',
      modified: '2024-01-01T00:00:00.000Z',
      name: 'Acme Corp',
      identity_class: 'organization',
    },
    {
      type: 'relationship',
      id: 'relationship--xyz789',
      created: '2024-01-01T00:00:00.000Z',
      modified: '2024-01-01T00:00:00.000Z',
      relationship_type: 'targets',
      source_ref: 'attack-pattern--abc123',
      target_ref: 'identity--def456',
      description: 'Targets this organization',
    },
  ],
} as const;

const INVALID_STIX_BUNDLE = {
  type: 'not-a-bundle',
  objects: [],
} as const;

const STIX_WITH_ERRORS = {
  type: 'bundle',
  id: 'bundle--test',
  objects: [
    {
      type: 'invalid-id-format',
      id: 'bad-id',
      created: '2024-01-01T00:00:00.000Z',
      modified: '2024-01-01T00:00:00.000Z',
      name: 'Invalid Object',
    },
  ],
} as const;

// ============================================================================
// Tests
// ============================================================================

describe('STIXImportEngine', () => {
  describe('parse', () => {
    it('should parse valid STIX bundle', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes).toHaveLength(2);
      expect(result.edges).toHaveLength(1);
      expect(result.errors).toHaveLength(0);
      expect(result.stats.importedObjects).toBe(2);
      expect(result.stats.totalObjects).toBe(3);
    });

    it('should handle invalid bundle type', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(INVALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes).toHaveLength(0);
      expect(result.edges).toHaveLength(0);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0].type).toBe('invalid-type');
    });

    it('should handle invalid JSON', async () => {
      const engine = createSTIXImportEngine();
      const json = '{ invalid json }';
      const result = await engine.parse(json);

      expect(result.nodes).toHaveLength(0);
      expect(result.errors[0].type).toBe('invalid-format');
    });

    it('should validate STIX object IDs', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(STIX_WITH_ERRORS);
      const result = await engine.parse(json);

      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0].type).toBe('invalid-format');
      expect(result.stats.errorObjects).toBe(1);
    });

    it('should filter by includeTypes', async () => {
      const engine = createSTIXImportEngine({
        includeTypes: ['attack-pattern'],
      });
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes).toHaveLength(1);
      expect(result.nodes[0].data.stixType).toBe('attack-pattern');
      expect(result.stats.skippedObjects).toBeGreaterThan(0);
    });

    it('should filter by excludeTypes', async () => {
      const engine = createSTIXImportEngine({
        excludeTypes: ['identity'],
      });
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes).toHaveLength(1);
      expect(result.nodes[0].data.stixType).toBe('attack-pattern');
    });

    it('should skip relationships when includeRelationships is false', async () => {
      const engine = createSTIXImportEngine({
        includeRelationships: false,
      });
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes).toHaveLength(2);
      expect(result.edges).toHaveLength(0);
      expect(result.stats.relationshipCount).toBe(0);
    });
  });

  describe('Type Mapping', () => {
    it('should map STIX types to GLC types', async () => {
      const engine = createSTIXImportEngine({
        mapToGLCTypes: true,
      });
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes[0].data.typeId).toBe('attack-technique');
      expect(result.nodes[1].data.typeId).toBe('asset');
    });

    it('should use original STIX types when mapToGLCTypes is false', async () => {
      const engine = createSTIXImportEngine({
        mapToGLCTypes: false,
      });
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes[0].data.typeId).toBe('attack-pattern');
      expect(result.nodes[1].data.typeId).toBe('identity');
    });

    it('should map to D3FEND when enabled', async () => {
      const engine = createSTIXImportEngine({
        mapToD3FEND: true,
      });
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes[0].data.d3fendClass).toBe('d3f:Detection');
    });
  });

  describe('Node Properties', () => {
    it('should extract name as label', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes[0].data.label).toBe('Spear Phishing');
    });

    it('should extract properties from STIX object', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      const node = result.nodes[0];
      expect(node.data.properties).toBeDefined();
      expect(node.data.properties?.some(p => p.key === 'description')).toBe(true);
      expect(node.data.properties?.some(p => p.key === 'created')).toBe(true);
    });

    it('should store STIX metadata', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.nodes[0].data.stixId).toBe('attack-pattern--abc123');
      expect(result.nodes[0].data.stixType).toBe('attack-pattern');
    });
  });

  describe('Relationships', () => {
    it('should create edges from STIX relationships', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.edges).toHaveLength(1);
      expect(result.edges[0].source).toBe('attack-pattern--abc123');
      expect(result.edges[0].target).toBe('identity--def456');
      expect(result.edges[0].label).toBe('targets');
    });

    it('should map relationship types', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.edges[0].data?.relationshipType).toBe('targets');
    });

    it('should skip relationships with invalid references', async () => {
      const bundleWithInvalidRef = {
        type: 'bundle',
        id: 'bundle--test',
        objects: [
          {
            type: 'attack-pattern',
            id: 'attack-pattern--test',
            created: '2024-01-01T00:00:00.000Z',
            modified: '2024-01-01T00:00:00.000Z',
            name: 'Test',
          },
          {
            type: 'relationship',
            id: 'relationship--invalid',
            created: '2024-01-01T00:00:00.000Z',
            modified: '2024-01-01T00:00:00.000Z',
            relationship_type: 'related-to',
            source_ref: 'attack-pattern--test',
            target_ref: 'non-existent--id',
          },
        ],
      };

      const engine = createSTIXImportEngine();
      const json = JSON.stringify(bundleWithInvalidRef);
      const result = await engine.parse(json);

      expect(result.edges).toHaveLength(0);
      expect(result.errors.length).toBeGreaterThan(0);
    });
  });

  describe('Statistics', () => {
    it('should track statistics correctly', async () => {
      const engine = createSTIXImportEngine();
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await engine.parse(json);

      expect(result.stats.totalObjects).toBe(3);
      expect(result.stats.importedObjects).toBe(2);
      expect(result.stats.skippedObjects).toBe(0);
      expect(result.stats.errorObjects).toBe(0);
      expect(result.stats.relationshipCount).toBe(1);
      expect(result.stats.byType['attack-pattern']).toBe(1);
      expect(result.stats.byType['identity']).toBe(1);
      expect(result.stats.byType['relationship']).toBe(1);
    });
  });
});

describe('Helper Functions', () => {
  describe('createSTIXImportEngine', () => {
    it('should create engine instance', () => {
      const engine = createSTIXImportEngine();
      expect(engine).toBeInstanceOf(STIXImportEngine);
    });

    it('should accept default options', () => {
      const engine = createSTIXImportEngine({
        mapToD3FEND: true,
      });
      expect(engine).toBeDefined();
    });
  });

  describe('importSTIX', () => {
    it('should import STIX and return result', async () => {
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await importSTIX(json);

      expect(result).toBeDefined();
      expect(result.nodes.length).toBeGreaterThan(0);
      expect(result.edges.length).toBeGreaterThan(0);
    });

    it('should accept import options', async () => {
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await importSTIX(json, { mapToGLCTypes: false });

      expect(result).toBeDefined();
      expect(result.nodes[0].data.typeId).toBe('attack-pattern');
    });
  });

  describe('validateSTIX', () => {
    it('should return valid for correct STIX', async () => {
      const json = JSON.stringify(VALID_STIX_BUNDLE);
      const result = await validateSTIX(json);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should return invalid for incorrect STIX', async () => {
      const json = JSON.stringify(INVALID_STIX_BUNDLE);
      const result = await validateSTIX(json);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should return validation errors', async () => {
      const json = JSON.stringify(STIX_WITH_ERRORS);
      const result = await validateSTIX(json);

      expect(result.valid).toBe(false);
      expect(result.errors[0].type).toBeDefined();
      expect(result.errors[0].message).toBeDefined();
    });
  });
});
