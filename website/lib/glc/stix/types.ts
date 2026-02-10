/**
 * GLC STIX Types
 *
 * Type definitions for STIX 2.1 objects.
 * See: https://docs.oasis-open.org/cti/stix/v2.1/
 */

// ============================================================================
// STIX 2.1 Core Types
// ============================================================================

/**
 * STIX 2.1 Common Properties
 */
export interface STIXCommonProperties {
  type: string;
  id: string;
  spec_version?: string;
  created: string;
  modified: string;
  object_marking_refs?: string[];
  granular_markings?: GranularMarking[];
  defanged?: boolean;
  extensions?: Record<string, unknown>;
}

/**
 * Granular Marking
 */
export interface GranularMarking {
  marking_ref: string;
  selectors: string[];
}

// ============================================================================
// Domain Objects
// ============================================================================

/**
 * STIX Attack Pattern (TTP)
 */
export interface STIXAttackPattern extends STIXCommonProperties {
  type: 'attack-pattern';
  name: string;
  description?: string;
  aliases?: string[];
  kill_chain_phases?: KillChainPhase[];
  external_references?: ExternalReference[];
  object_marking_refs?: string[];
}

/**
 * STIX Campaign
 */
export interface STIXCampaign extends STIXCommonProperties {
  type: 'campaign';
  name: string;
  description?: string;
  aliases?: string[];
  first_seen?: string;
  last_seen?: string;
}

/**
 * STIX Course of Action
 */
export interface STIXCourseOfAction extends STIXCommonProperties {
  type: 'course-of-action';
  name: string;
  description?: string;
}

/**
 * STIX Grouping
 */
export interface STIXGrouping extends STIXCommonProperties {
  type: 'grouping';
  name: string;
  description?: string;
  context?: string;
  object_refs: string[];
}

/**
 * STIX Identity
 */
export interface STIXIdentity extends STIXCommonProperties {
  type: 'identity';
  name: string;
  description?: string;
  identity_class?: string;
  roles?: string[];
  sectors?: string[];
}

/**
 * STIX Indicator
 */
export interface STIXIndicator extends STIXCommonProperties {
  type: 'indicator';
  name?: string;
  description?: string;
  pattern: string;
  pattern_type: string;
  valid_from?: string;
  valid_until?: string;
  kill_chain_phases?: KillChainPhase[];
}

/**
 * STIX Infrastructure
 */
export interface STIXInfrastructure extends STIXCommonProperties {
  type: 'infrastructure';
  name: string;
  description?: string;
  infrastructure_types?: string[];
  aliases?: string[];
  kill_chain_phases?: KillChainPhase[];
}

/**
 * STIX Intrusion Set
 */
export interface STIXIntrusionSet extends STIXCommonProperties {
  type: 'intrusion-set';
  name: string;
  description?: string;
  aliases?: string[];
  first_seen?: string;
  last_seen?: string;
  goals?: string[];
}

/**
 * STIX Location
 */
export interface STIXLocation extends STIXCommonProperties {
  type: 'location';
  name?: string;
  description?: string;
  latitude?: number;
  longitude?: number;
  region?: string;
  country?: string;
  administrative_area?: string;
  city?: string;
}

/**
 * STIX Malware
 */
export interface STIXMalware extends STIXCommonProperties {
  type: 'malware';
  name: string;
  description?: string;
  malware_types?: string[];
  kill_chain_phases?: KillChainPhase[];
  is_family?: boolean;
  aliases?: string[];
  first_seen?: string;
  last_seen?: string;
}

/**
 * STIX Threat Actor
 */
export interface STIXThreatActor extends STIXCommonProperties {
  type: 'threat-actor';
  name: string;
  description?: string;
  threat_actor_types?: string[];
  aliases?: string[];
  roles?: string[];
  goals?: string[];
  sophistication?: string;
  resource_level?: string;
  first_seen?: string;
  last_seen?: string;
}

/**
 * STIX Tool
 */
export interface STIXTool extends STIXCommonProperties {
  type: 'tool';
  name: string;
  description?: string;
  tool_types?: string[];
  aliases?: string[];
  kill_chain_phases?: KillChainPhase[];
}

/**
 * STIX Vulnerability
 */
export interface STIXVulnerability extends STIXCommonProperties {
  type: 'vulnerability';
  name: string;
  description?: string;
  external_references?: ExternalReference[];
}

// ============================================================================
// Relationship Objects
// ============================================================================

/**
 * STIX Relationship
 */
export interface STIXRelationship extends STIXCommonProperties {
  type: 'relationship';
  relationship_type: string;
  source_ref: string;
  target_ref: string;
  description?: string;
  start_time?: string;
  stop_time?: string;
}

/**
 * STIX Sighting
 */
export interface STIXSighting extends STIXCommonProperties {
  type: 'sighting';
  first_seen?: string;
  last_seen?: string;
  sighting_of_ref: string;
  where_sighted_refs?: string[];
  observed_data_refs?: string[];
}

// ============================================================================
// Supporting Types
// ============================================================================

/**
 * Kill Chain Phase
 */
export interface KillChainPhase {
  kill_chain_name: string;
  phase_name: string;
}

/**
 * External Reference
 */
export interface ExternalReference {
  source_name: string;
  description?: string;
  url?: string;
  external_id?: string;
}

/**
 * STIX Bundle
 */
export interface STIXBundle {
  type: 'bundle';
  id: string;
  spec_version?: string;
  objects: STIXObject[];
}

// ============================================================================
// Union Types
// ============================================================================

export type STIXObject =
  | STIXAttackPattern
  | STIXCampaign
  | STIXCourseOfAction
  | STIXGrouping
  | STIXIdentity
  | STIXIndicator
  | STIXInfrastructure
  | STIXIntrusionSet
  | STIXLocation
  | STIXMalware
  | STIXThreatActor
  | STIXTool
  | STIXVulnerability
  | STIXRelationship
  | STIXSighting;

// ============================================================================
// Validation & Mapping Types
// ============================================================================

/**
 * STIX Import Options
 */
export interface STIXImportOptions {
  /**
   * Map STIX objects to GLC node types
   */
  mapToGLCTypes?: boolean;

  /**
   * Map to D3FEND ontology for defensive context
   */
  mapToD3FEND?: boolean;

  /**
   * Filter objects by type
   */
  includeTypes?: string[];

  /**
   * Exclude objects by type
   */
  excludeTypes?: string[];

  /**
   * Resolve relationships to create edges
   */
  includeRelationships?: boolean;

  /**
   * Maximum depth for nested relationships
   */
  maxRelationshipDepth?: number;
}

/**
 * STIX Import Result
 */
export interface STIXImportResult {
  /**
   * Successfully imported nodes
   */
  nodes: GLCImportNode[];

  /**
   * Successfully imported edges
   */
  edges: GLCImportEdge[];

  /**
   * Validation errors
   */
  errors: STIXValidationError[];

  /**
   * Statistics
   */
  stats: STIXImportStats;
}

/**
 * GLC Import Node
 */
export interface GLCImportNode {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: {
    label: string;
    typeId: string;
    properties?: Property[];
    references?: Reference[];
    stixId?: string;
    stixType?: string;
  };
}

/**
 * GLC Import Edge
 */
export interface GLCImportEdge {
  id: string;
  source: string;
  target: string;
  type?: string;
  label?: string;
  data?: {
    relationshipType?: string;
    stixId?: string;
  };
}

/**
 * STIX Validation Error
 */
export interface STIXValidationError {
  type: 'missing-field' | 'invalid-type' | 'invalid-format' | 'duplicate-id';
  message: string;
  objectId?: string;
  field?: string;
}

/**
 * STIX Import Statistics
 */
export interface STIXImportStats {
  totalObjects: number;
  importedObjects: number;
  skippedObjects: number;
  errorObjects: number;
  relationshipCount: number;
  byType: Record<string, number>;
}

/**
 * Property
 */
export interface Property {
  key: string;
  value: string;
  type: 'string' | 'number' | 'boolean' | 'date' | 'url';
}

/**
 * Reference
 */
export interface Reference {
  type: 'cve' | 'cwe' | 'capec' | 'attack' | 'd3fend' | 'url' | 'stix';
  id: string;
  label?: string;
  url?: string;
}
