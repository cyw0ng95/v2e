export interface D3FENDClass {
  id: string;
  name: string;
  d3fend_type: string;
  type: string;
  description: string;
  tactics: string[];
}

export interface D3FENDProperty {
  id: id: string;
  name: string;
  type: string;
  description: string;
  required: boolean;
  options?: string[];
}

export interface D3FENDInference {
  id: string;
  type: 'automatic' | 'suggested' | 'manual';
  source_node_types: string[];
  relationship_type: string;
  target_node_types: string[];
  properties: Record<string, any>;
  description: string;
}

export interface D3FENDData {
  version: string;
  classes: D3FENDClass[];
  properties: Record<string, D3FENDProperty>;
  relationships: Record<string, any>;
  inferences: D3FENDInference[];
}

export const DEFAULT_D3FEND_DATA: D3FENDData = {
  version: '1.0.0',
  classes: [],
  properties: {},
  relationships: {},
  inferences: [],
};

export const validateD3FENDData = (data: unknown): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (typeof data !== 'object' || data === null) {
    return { valid: false, errors: ['Data must be an object'] };
  }

  if (!data.version || typeof data.version !== 'string') {
    errors.push('Missing or invalid version field');
  }

  if (!Array.isArray(data.classes)) {
    errors.push('Classes must be an array');
  }

  if (typeof data.properties !== 'object' || data.properties === null) {
    errors.push('Properties must be an object');
  }

  if (!Array.isArray(data.inferences)) {
    errors.push('Inferences must be an array');
  }

  if (data.classes.length === 0) {
    errors.push('Classes array cannot be empty');
  }

  if (Object.keys(data.properties).length === 0) {
    errors.push('Properties object cannot be empty');
  }

  return {
    valid: errors.length === 0,
    errors,
  };
};

export const getD3FENDClassById = (data: D3FENDData, classId: string): D3FENDClass | null => {
  return data.classes.find(c => c.id === classId) || null;
};

export const getD3FENDProperties = (data: D3FENDData, classId: string): D3FENDProperty[] => {
  const classDef = getD3FENDClassById(data, classId);
  return classDef?.properties.map(prop => ({
    id: prop.id,
    name: prop.name,
    type: prop.type,
    description: prop.description,
    required: prop.required,
    options: prop.options,
  })) || [];
};

export const getD3FENDRelationships = (data: D3FENDData): Record<string, any> => {
  return data.relationships || {};
};

export const getD3FENDInferences = (data: D3FENDData, relationshipType: string): D3FENDInference[] => {
  return data.inferences.filter(inf => 
    inf.type === 'automatic' && 
    inf.relationship_type === relationshipType
  );
};

export const isD3FENDNode = (nodeType: string): boolean => {
  const d3fendClasses = [
    'event',
    'remote-command',
    'countermeasure',
    'artifact',
    'agent',
    'vulnerability',
    'condition',
    'thing',
  ];
  return d3fendClasses.includes(nodeType);
};

export const getNodeD3FENDClass = (
  nodeType: string,
  data: D3FENDData
): D3FENDClass | null => {
  if (!isD3FENDNode(nodeType)) {
    return null;
  }

  return getD3FENDClassById(data, nodeType);
};

export const getNodeD3FENDProperties = (
  nodeType: string,
  data: D3FENDData
): D3FENDProperty[] => {
  const classDef = getNodeD3ENDClass(nodeType, data);
  return classDef ? getD3FENDProperties(data, classDef.id) : [];
};

export const getNodeD3FENDTactics = (
  nodeType: string,
  data: D3FENDData
): string[] => {
  const classDef = getNodeD3ENDClass(nodeType, data);
  return classDef?.tactics || [];
};

export default {
  DEFAULT_D3FEND_DATA,
  validateD3FENDData,
  getD3FENDClassById,
  getD3FENDProperties,
  getD3FENDRelationships,
  getD3FENDInferences,
  isD3FENDNode,
  getNodeD3ENDClass,
  getNodeD3ENDProperties,
  getNodeD3ENDTactics,
};
