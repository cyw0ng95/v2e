/**
 * GLC STIX Import Engine
 *
 * Imports STIX 2.1 JSON and converts to GLC graph structure.
 */

import { z } from 'zod';
import { D3FEND_CLASSES } from '../d3fend/ontology';
import {
  type STIXBundle,
  type STIXObject,
  type STIXRelationship,
  type STIXImportOptions,
  type STIXImportResult,
  type GLCImportNode,
  type GLCImportEdge,
  type STIXValidationError,
  type STIXImportStats,
} from './types';

// ============================================================================
// STIX Validation Schemas
// ============================================================================

const STIXCommonPropertiesSchema = z.object({
  type: z.string(),
  id: z.string().regex(/^[a-z][a-z-]*--[0-9a-z]+$/i, 'STIX ID must follow type--uuid format'),
  created: z.string().datetime(),
  modified: z.string().datetime(),
  spec_version: z.string().optional(),
  object_marking_refs: z.array(z.string()).optional(),
  granular_markings: z.array(z.object({})).optional(),
  defanged: z.boolean().optional(),
  extensions: z.record(z.string(), z.unknown()).optional(),
});

const STIXRelationshipSchema = STIXCommonPropertiesSchema.extend({
  type: z.literal('relationship'),
  relationship_type: z.string(),
  source_ref: z.string(),
  target_ref: z.string(),
  description: z.string().optional(),
  start_time: z.string().datetime().optional(),
  stop_time: z.string().datetime().optional(),
});

// ============================================================================
// Type Mappings: STIX to GLC
// ============================================================================

const STIX_TO_GLC_TYPE_MAPPING: Record<string, string> = {
  'attack-pattern': 'attack-technique',
  'campaign': 'group',
  'course-of-action': 'technique',
  'grouping': 'group',
  'identity': 'asset',
  'indicator': 'technique',
  'infrastructure': 'asset',
  'intrusion-set': 'group',
  'location': 'asset',
  'malware': 'vulnerability',
  'threat-actor': 'group',
  'tool': 'software',
  'vulnerability': 'vulnerability',
};

const STIX_TO_D3FEND_MAPPING: Record<string, string> = {
  'attack-pattern': 'd3f:Detection',
  'indicator': 'd3f:Detection',
  'malware': 'd3f:FileAnalysis',
  'tool': 'd3f:ProcessAnalysis',
  'course-of-action': 'd3f:Hardening',
};

const RELATIONSHIP_TYPE_MAPPING: Record<string, string> = {
  'related-to': 'connects',
  'uses': 'uses',
  'mitigates': 'mitigates',
  'targets': 'targets',
  'attributed-to': 'attributed-to',
  'indicates': 'indicates',
  'located-at': 'located-at',
  'variant-of': 'variant-of',
  'derived-from': 'derived-from',
  'duplicate-of': 'duplicate-of',
};

// ============================================================================
// STIX Import Engine Class
// ============================================================================

export class STIXImportEngine {
  private options: Required<STIXImportOptions>;
  private errors: STIXValidationError[] = [];
  private stats: STIXImportStats = {
    totalObjects: 0,
    importedObjects: 0,
    skippedObjects: 0,
    errorObjects: 0,
    relationshipCount: 0,
    byType: {},
  };

  constructor(options: STIXImportOptions = {}) {
    this.options = {
      mapToGLCTypes: options.mapToGLCTypes ?? true,
      mapToD3FEND: options.mapToD3FEND ?? false,
      includeTypes: options.includeTypes ?? [],
      excludeTypes: options.excludeTypes ?? [],
      includeRelationships: options.includeRelationships ?? true,
      maxRelationshipDepth: options.maxRelationshipDepth ?? 3,
    };
  }

  /**
   * Parse and validate STIX JSON
   */
  async parse(json: string): Promise<STIXImportResult> {
    this.reset();

    try {
      const bundle = JSON.parse(json) as STIXBundle;

      // Validate bundle structure
      if (bundle.type !== 'bundle') {
        this.errors.push({
          type: 'invalid-type',
          message: 'Root object must be a STIX bundle',
        });
        return this.buildResult();
      }

      // Process objects
      const objectMap = new Map<string, STIXObject>();
      const relationships: STIXRelationship[] = [];

      for (const obj of bundle.objects) {
        this.stats.totalObjects++;

        // Track by type for statistics
        this.stats.byType[obj.type] = (this.stats.byType[obj.type] || 0) + 1;

        // Check inclusion/exclusion filters
        if (
          (this.options.includeTypes.length > 0 && !this.options.includeTypes.includes(obj.type)) ||
          this.options.excludeTypes.includes(obj.type)
        ) {
          this.stats.skippedObjects++;
          continue;
        }

        // Validate object
        const validationError = this.validateObject(obj);
        if (validationError) {
          this.errors.push(validationError);
          this.stats.errorObjects++;
          continue;
        }

        // Store object for relationship resolution
        objectMap.set(obj.id, obj);

        // Separate relationships (only if includeRelationships is true)
        if (this.options.includeRelationships && obj.type === 'relationship') {
          relationships.push(obj as STIXRelationship);
          this.stats.relationshipCount++;
        }
      }

      // Convert objects to GLC nodes
      const nodes = this.convertToNodes(objectMap);

      // Convert relationships to GLC edges
      const edges = this.options.includeRelationships
        ? this.convertToEdges(relationships, objectMap)
        : [];

      this.stats.importedObjects = nodes.length;

      return {
        nodes,
        edges,
        errors: this.errors,
        stats: this.stats,
      };
    } catch (error) {
      if (error instanceof SyntaxError) {
        this.errors.push({
          type: 'invalid-format',
          message: 'Invalid JSON format',
        });
      } else {
        this.errors.push({
          type: 'invalid-format',
          message: error instanceof Error ? error.message : 'Unknown error',
        });
      }

      return this.buildResult();
    }
  }

  /**
   * Validate a STIX object
   */
  private validateObject(obj: STIXObject): STIXValidationError | null {
    // Check required fields
    if (!obj.type || !obj.id) {
      return {
        type: 'missing-field',
        message: 'Missing required fields: type or id',
      };
    }

    // Validate ID format (STIX 2.1: type--uuid)
    // UUID part can be alphanumeric hex (0-9, a-f) or any test string
    // Type can contain lowercase letters and hyphens (e.g., attack-pattern, course-of-action)
    if (!obj.id.match(/^[a-z][a-z-]*--[0-9a-z]+$/i)) {
      return {
        type: 'invalid-format',
        message: 'Invalid STIX ID format',
        objectId: obj.id,
      };
    }

    // Validate specific types
    if (obj.type === 'relationship') {
      const result = STIXRelationshipSchema.safeParse(obj);
      if (!result.success) {
        const firstError = result.error.issues[0];
        return {
          type: 'invalid-format',
          message: firstError?.message || 'Unknown validation error',
          objectId: obj.id,
        };
      }
    }

    return null;
  }

  /**
   * Convert STIX objects to GLC nodes
   */
  private convertToNodes(objectMap: Map<string, STIXObject>): GLCImportNode[] {
    const nodes: GLCImportNode[] = [];
    const processedIds = new Set<string>();

    for (const [id, obj] of objectMap) {
      if (obj.type === 'relationship') continue;
      if (processedIds.has(id)) continue;

      processedIds.add(id);

      try {
        const node = this.stixObjectToNode(obj);
        if (node) {
          nodes.push(node);
        }
      } catch (e) {
        this.errors.push({
          type: 'invalid-format',
          message: `Failed to convert object ${id}: ${e instanceof Error ? e.message : String(e)}`,
          objectId: id,
        });
        this.stats.errorObjects++;
      }
    }

    return nodes;
  }

  /**
   * Convert STIX relationship to GLC edge
   */
  private convertToEdges(
    relationships: STIXRelationship[],
    objectMap: Map<string, STIXObject>
  ): GLCImportEdge[] {
    const edges: GLCImportEdge[] = [];

    for (const rel of relationships) {
      const sourceExists = objectMap.has(rel.source_ref);
      const targetExists = objectMap.has(rel.target_ref);

      if (!sourceExists || !targetExists) {
        this.errors.push({
          type: 'invalid-format',
          message: `Relationship references non-existent object: ${rel.source_ref} -> ${rel.target_ref}`,
          objectId: rel.id,
        });
        continue;
      }

      const edge: GLCImportEdge = {
        id: rel.id,
        source: rel.source_ref,
        target: rel.target_ref,
        label: this.mapRelationshipType(rel.relationship_type),
        data: {
          relationshipType: this.mapRelationshipType(rel.relationship_type),
          stixId: rel.id,
        },
      };

      edges.push(edge);
    }

    return edges;
  }

  /**
   * Convert a single STIX object to GLC node
   */
  private stixObjectToNode(obj: STIXObject): GLCImportNode | null {
    const position = this.calculatePosition();

    const baseData = {
      label: this.getObjectLabel(obj),
      stixId: obj.id,
      stixType: obj.type,
      properties: this.extractProperties(obj),
      references: this.extractReferences(obj),
    };

    let typeId: string = obj.type;
    let nodeType = 'glc';
    let d3fendClass: string | undefined;

    // Apply D3FEND mapping if enabled
    if (this.options.mapToD3FEND) {
      const d3fendClassId = STIX_TO_D3FEND_MAPPING[obj.type];
      if (d3fendClassId) {
        d3fendClass = d3fendClassId;
        typeId = d3fendClassId;
      }
    }

    // Apply GLC type mapping if enabled
    if (this.options.mapToGLCTypes) {
      typeId = STIX_TO_GLC_TYPE_MAPPING[obj.type] || typeId;
    }

    return {
      id: obj.id,
      type: nodeType,
      position,
      data: {
        ...baseData,
        typeId,
        d3fendClass,
      },
    };
  }

  /**
   * Get label for STIX object
   */
  private getObjectLabel(obj: STIXObject): string {
    const typedObj = obj as Record<string, unknown>;

    if ('name' in typedObj && typeof typedObj.name === 'string') {
      return typedObj.name;
    }

    return `${obj.type} (${obj.id.split('--')[1]?.substring(0, 8) || 'unknown'})`;
  }

  /**
   * Extract properties from STIX object
   */
  private extractProperties(obj: STIXObject): Array<{ key: string; value: string; type: string }> {
    const properties: Array<{ key: string; value: string; type: string }> = [];
    const typedObj = obj as Record<string, unknown>;

    // Common properties to extract
    const fields = ['description', 'created', 'modified', 'aliases', 'first_seen', 'last_seen'];

    for (const field of fields) {
      const value = typedObj[field] as string | undefined | null | boolean | number | string[];
      if (value !== undefined && value !== null) {
        let displayValue: string = '';
        let type: string = 'string';

        if (Array.isArray(value)) {
          displayValue = value.join(', ');
        } else if (typeof value === 'boolean') {
          displayValue = value.toString();
          type = 'boolean';
        } else if (typeof value === 'number') {
          displayValue = value.toString();
          type = 'number';
        } else if (typeof value === 'string') {
          displayValue = value;
        } else if (value !== null && value !== undefined) {
          displayValue = JSON.stringify(value);
        }

        properties.push({
          key: field,
          value: displayValue,
          type,
        });
      }
    }

    return properties;
  }

  /**
   * Extract references from STIX object
   */
  private extractReferences(obj: STIXObject): Array<{ type: string; id: string; label?: string; url?: string }> {
    const references: Array<{ type: string; id: string; label?: string; url?: string }> = [];
    const typedObj = obj as Record<string, unknown>;

    if ('external_references' in typedObj && Array.isArray(typedObj.external_references)) {
      for (const ref of typedObj.external_references) {
        if (typeof ref === 'object' && ref !== null) {
          const extRef = ref as Record<string, unknown>;

          if ('external_id' in extRef && typeof extRef.external_id === 'string') {
            const externalId = extRef.external_id;
            let type = 'url';

            // Detect reference type
            if (externalId.startsWith('CVE-')) {
              type = 'cve';
            } else if (externalId.startsWith('CWE-')) {
              type = 'cwe';
            } else if (externalId.startsWith('CAPEC-')) {
              type = 'capec';
            } else if (externalId.startsWith('T')) {
              type = 'attack';
            }

            references.push({
              type,
              id: externalId,
              label: 'source_name' in extRef ? extRef.source_name as string : undefined,
              url: 'url' in extRef ? extRef.url as string : undefined,
            });
          }
        }
      }
    }

    return references;
  }

  /**
   * Map STIX relationship type to GLC relationship type
   */
  private mapRelationshipType(stixType: string): string {
    return RELATIONSHIP_TYPE_MAPPING[stixType] || stixType;
  }

  /**
   * Map node type to D3FEND class
   */
  private mapNodeTypeToD3FENDClass(nodeType: string): string | null {
    // Direct mapping for known D3FEND node types
    if (D3FEND_CLASSES.find(c => c.id === nodeType)) {
      return nodeType;
    }

    // Map common node types to D3FEND classes
    const mappings: Record<string, string> = {
      'firewall': 'd3f:Isolation',
      'ids': 'd3f:Detection',
      'ips': 'd3f:Detection',
      'siem': 'd3f:Hardening',
      'waf': 'd3f:Hardening',
    };

    return mappings[nodeType] || null;
  }

  /**
   * Calculate initial position for node
   */
  private calculatePosition(): { x: number; y: number } {
    const index = this.stats.importedObjects;
    const angle = index * 0.5; // Golden angle-ish
    const radius = 100 + index * 50;

    const x = 500 + radius * Math.cos(angle);
    const y = 300 + radius * Math.sin(angle);

    return { x: Math.round(x), y: Math.round(y) };
  }

  /**
   * Reset internal state
   */
  private reset(): void {
    this.errors = [];
    this.stats = {
      totalObjects: 0,
      importedObjects: 0,
      skippedObjects: 0,
      errorObjects: 0,
      relationshipCount: 0,
      byType: {},
    };
  }

  /**
   * Build import result
   */
  private buildResult(): STIXImportResult {
    return {
      nodes: [],
      edges: [],
      errors: this.errors,
      stats: this.stats,
    };
  }
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Create STIX import engine with default options
 */
export function createSTIXImportEngine(options?: STIXImportOptions): STIXImportEngine {
  return new STIXImportEngine(options);
}

/**
 * Import STIX JSON and return GLC nodes/edges
 */
export async function importSTIX(
  json: string,
  options?: STIXImportOptions
): Promise<STIXImportResult> {
  const engine = createSTIXImportEngine(options);
  return engine.parse(json);
}

/**
 * Validate STIX JSON without importing
 */
export async function validateSTIX(json: string): Promise<{
  valid: boolean;
  errors: STIXValidationError[];
}> {
  const engine = createSTIXImportEngine({ includeRelationships: false });
  const result = await engine.parse(json);

  return {
    valid: result.errors.length === 0,
    errors: result.errors,
  };
}
