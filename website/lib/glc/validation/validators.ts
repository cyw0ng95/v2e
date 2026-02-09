import { z } from 'zod';
import {
  canvasPresetSchema,
  nodeTypeDefinitionSchema,
  relationshipDefinitionSchema,
  presetStylingSchema,
  presetBehaviorSchema,
  ValidationRule,
  CanvasPreset,
  NodeTypeDefinition,
  RelationshipDefinition,
} from '../types';

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
}

export interface ValidationError {
  path: string;
  message: string;
  code: string;
}

export interface ValidationWarning {
  path: string;
  message: string;
  code: string;
}

export class PresetValidationError extends Error {
  public validationErrors: ValidationError[];
  public validationWarnings: ValidationWarning[];

  constructor(
    message: string,
    errors: ValidationError[],
    warnings: ValidationWarning[] = []
  ) {
    super(message);
    this.name = 'PresetValidationError';
    this.validationErrors = errors;
    this.validationWarnings = warnings;
  }
}

export const validatePreset = (preset: unknown): ValidationResult => {
  const errors: ValidationError[] = [];
  const warnings: ValidationWarning[] = [];

  try {
    const result = canvasPresetSchema.safeParse(preset);
    
    if (!result.success) {
      result.error.errors.forEach((err) => {
        errors.push({
          path: err.path.join('.'),
          message: err.message,
          code: err.code,
        });
      });
    }

    if (errors.length > 0) {
      return { valid: false, errors, warnings };
    }

    const validPreset = result.data as CanvasPreset;
    
    warnings.push(...validateNodeTypes(validPreset));
    warnings.push(...validateRelationshipTypes(validPreset));
    warnings.push(...validateBehavior(validPreset));

    return { valid: true, errors, warnings };
  } catch (error) {
    errors.push({
      path: 'root',
      message: 'Unexpected validation error',
      code: 'UNEXPECTED_ERROR',
    });
    return { valid: false, errors, warnings };
  }
};

const validateNodeTypes = (preset: CanvasPreset): ValidationWarning[] => {
  const warnings: ValidationWarning[] = [];
  const nodeTypeIds = new Set<string>();

  preset.nodeTypes.forEach((nodeType) => {
    if (nodeTypeIds.has(nodeType.id)) {
      warnings.push({
        path: `nodeTypes.${nodeType.id}`,
        message: `Duplicate node type ID: ${nodeType.id}`,
        code: 'DUPLICATE_NODE_TYPE_ID',
      });
    }
    nodeTypeIds.add(nodeType.id);

    if (!nodeType.name || nodeType.name.trim() === '') {
      warnings.push({
        path: `nodeTypes.${nodeType.id}`,
        message: `Node type has empty name: ${nodeType.id}`,
        code: 'EMPTY_NODE_TYPE_NAME',
      });
    }

    if (!nodeType.category || nodeType.category.trim() === '') {
      warnings.push({
        path: `nodeTypes.${nodeType.id}`,
        message: `Node type has empty category: ${nodeType.id}`,
        code: 'EMPTY_NODE_TYPE_CATEGORY',
      });
    }

    if (!nodeType.description || nodeType.description.trim() === '') {
      warnings.push({
        path: `nodeTypes.${nodeType.id}`,
        message: `Node type has empty description: ${nodeType.id}`,
        code: 'EMPTY_NODE_TYPE_DESCRIPTION',
      });
    }

    nodeType.properties.forEach((property) => {
      if (!property.name || property.name.trim() === '') {
        warnings.push({
          path: `nodeTypes.${nodeType.id}.properties.${property.id}`,
          message: `Property has empty name: ${property.id}`,
          code: 'EMPTY_PROPERTY_NAME',
        });
      }

      if (property.type === 'enum' || property.type === 'multiselect') {
        if (!property.options || property.options.length === 0) {
          warnings.push({
            path: `nodeTypes.${nodeType.id}.properties.${property.id}`,
            message: `Enum/multiselect property has no options: ${property.id}`,
            code: 'EMPTY_ENUM_OPTIONS',
          });
        }
      }
    });
  });

  if (preset.nodeTypes.length === 0) {
    warnings.push({
      path: 'nodeTypes',
      message: 'No node types defined',
      code: 'NO_NODE_TYPES',
    });
  }

  return warnings;
};

const validateRelationshipTypes = (preset: CanvasPreset): ValidationWarning[] => {
  const warnings: ValidationWarning[] = [];
  const relationshipTypeIds = new Set<string>();
  const nodeTypeIds = new Set(preset.nodeTypes.map(nt => nt.id));

  preset.relationshipTypes.forEach((relType) => {
    if (relationshipTypeIds.has(relType.id)) {
      warnings.push({
        path: `relationshipTypes.${relType.id}`,
        message: `Duplicate relationship type ID: ${relType.id}`,
        code: 'DUPLICATE_RELATIONSHIP_TYPE_ID',
      });
    }
    relationshipTypeIds.add(relType.id);

    if (!relType.name || relType.name.trim() === '') {
      warnings.push({
        path: `relationshipTypes.${relType.id}`,
        message: `Relationship type has empty name: ${relType.id}`,
        code: 'EMPTY_RELATIONSHIP_TYPE_NAME',
      });
    }

    if (!relType.category || relType.category.trim() === '') {
      warnings.push({
        path: `relationshipTypes.${relType.id}`,
        message: `Relationship type has empty category: ${relType.id}`,
        code: 'EMPTY_RELATIONSHIP_TYPE_CATEGORY',
      });
    }

    const validSourceTypes = relType.sourceNodeTypes.filter(t => t === '*' || nodeTypeIds.has(t));
    const validTargetTypes = relType.targetNodeTypes.filter(t => t === '*' || nodeTypeIds.has(t));

    if (validSourceTypes.length === 0) {
      warnings.push({
        path: `relationshipTypes.${relType.id}.sourceNodeTypes`,
        message: `No valid source node types found`,
        code: 'INVALID_SOURCE_NODE_TYPES',
      });
    }

    if (validTargetTypes.length === 0) {
      warnings.push({
        path: `relationshipTypes.${relType.id}.targetNodeTypes`,
        message: `No valid target node types found`,
        code: 'INVALID_TARGET_NODE_TYPES',
      });
    }
  });

  if (preset.relationshipTypes.length === 0) {
    warnings.push({
      path: 'relationshipTypes',
      message: 'No relationship types defined',
      code: 'NO_RELATIONSHIP_TYPES',
    });
  }

  return warnings;
};

const validateBehavior = (preset: CanvasPreset): ValidationWarning[] => {
  const warnings: ValidationWarning[] = [];

  if (preset.behavior.autoSave && preset.behavior.autoSaveInterval <= 0) {
    warnings.push({
      path: 'behavior.autoSaveInterval',
      message: 'Auto-save interval must be greater than 0',
      code: 'INVALID_AUTO_SAVE_INTERVAL',
    });
  }

  if (preset.behavior.snapToGrid && preset.behavior.gridSize <= 0) {
    warnings.push({
      path: 'behavior.gridSize',
      message: 'Grid size must be greater than 0',
      code: 'INVALID_GRID_SIZE',
    });
  }

  if (preset.behavior.maxNodes <= 0) {
    warnings.push({
      path: 'behavior.maxNodes',
      message: 'Max nodes must be greater than 0',
      code: 'INVALID_MAX_NODES',
    });
  }

  if (preset.behavior.maxEdges <= 0) {
    warnings.push({
      path: 'behavior.maxEdges',
      message: 'Max edges must be greater than 0',
      code: 'INVALID_MAX_EDGES',
    });
  }

  return warnings;
};

export const validatePresetFile = async (file: File): Promise<ValidationResult> => {
  try {
    const content = await file.text();
    const json = JSON.parse(content);
    return validatePreset(json);
  } catch (error) {
    return {
      valid: false,
      errors: [
        {
          path: 'root',
          message: error instanceof Error ? error.message : 'Failed to parse JSON',
          code: 'INVALID_JSON',
        },
      ],
      warnings: [],
    };
  }
};

export const validateGraph = (graph: unknown, preset: CanvasPreset): ValidationResult => {
  const errors: ValidationError[] = [];
  const warnings: ValidationWarning[] = [];

  try {
    if (typeof graph !== 'object' || graph === null) {
      errors.push({
        path: 'root',
        message: 'Graph must be an object',
        code: 'INVALID_GRAPH_TYPE',
      });
      return { valid: false, errors, warnings };
    }

    const graphObj = graph as any;
    const nodes = graphObj.nodes || [];
    const edges = graphObj.edges || [];

    if (!Array.isArray(nodes)) {
      errors.push({
        path: 'nodes',
        message: 'Nodes must be an array',
        code: 'INVALID_NODES_TYPE',
      });
    }

    if (!Array.isArray(edges)) {
      errors.push({
        path: 'edges',
        message: 'Edges must be an array',
        code: 'INVALID_EDGES_TYPE',
      });
    }

    const nodeIds = new Set<string>();
    const nodeTypeIds = new Set(preset.nodeTypes.map(nt => nt.id));

    nodes.forEach((node: any) => {
      if (!node.id) {
        errors.push({
          path: `nodes[${nodes.indexOf(node)}]`,
          message: 'Node missing ID',
          code: 'MISSING_NODE_ID',
        });
      } else if (nodeIds.has(node.id)) {
        errors.push({
          path: `nodes.${node.id}`,
          message: `Duplicate node ID: ${node.id}`,
          code: 'DUPLICATE_NODE_ID',
        });
      } else {
        nodeIds.add(node.id);
      }

      if (!node.type || !nodeTypeIds.has(node.type)) {
        errors.push({
          path: `nodes.${node.id}.type`,
          message: `Invalid or missing node type: ${node.type}`,
          code: 'INVALID_NODE_TYPE',
        });
      }

      if (!node.position || typeof node.position.x !== 'number' || typeof node.position.y !== 'number') {
        errors.push({
          path: `nodes.${node.id}.position`,
          message: 'Node has invalid position',
          code: 'INVALID_NODE_POSITION',
        });
      }
    });

    if (nodes.length > preset.behavior.maxNodes) {
      warnings.push({
        path: 'nodes',
        message: `Graph exceeds max nodes limit: ${nodes.length}/${preset.behavior.maxNodes}`,
        code: 'MAX_NODES_EXCEEDED',
      });
    }

    edges.forEach((edge: any) => {
      if (!edge.id) {
        errors.push({
          path: `edges[${edges.indexOf(edge)}]`,
          message: 'Edge missing ID',
          code: 'MISSING_EDGE_ID',
        });
      }

      if (!edge.source || !nodeIds.has(edge.source)) {
        errors.push({
          path: `edges.${edge.id}.source`,
          message: `Invalid or missing source node: ${edge.source}`,
          code: 'INVALID_EDGE_SOURCE',
        });
      }

      if (!edge.target || !nodeIds.has(edge.target)) {
        errors.push({
          path: `edges.${edge.id}.target`,
          message: `Invalid or missing target node: ${edge.target}`,
          code: 'INVALID_EDGE_TARGET',
        });
      }

      if (edge.source === edge.target) {
        warnings.push({
          path: `edges.${edge.id}`,
          message: 'Edge connects node to itself (self-loop)',
          code: 'SELF_LOOP_EDGE',
        });
      }
    });

    if (edges.length > preset.behavior.maxEdges) {
      warnings.push({
        path: 'edges',
        message: `Graph exceeds max edges limit: ${edges.length}/${preset.behavior.maxEdges}`,
        code: 'MAX_EDGES_EXCEEDED',
      });
    }

    return { valid: errors.length === 0, errors, warnings };
  } catch (error) {
    errors.push({
      path: 'root',
      message: 'Unexpected validation error',
      code: 'UNEXPECTED_ERROR',
    });
    return { valid: false, errors, warnings };
  }
};

export default {
  validatePreset,
  validatePresetFile,
  validateGraph,
  PresetValidationError,
};
