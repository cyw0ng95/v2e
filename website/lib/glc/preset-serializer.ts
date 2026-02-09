import { CanvasPreset, Graph } from '../types';
import { migratePreset, applyAllMigrations } from '../validation';
import { SerializationError } from '../errors';

export const serializePreset = (preset: CanvasPreset): CanvasPreset => {
  try {
    const serialized = JSON.parse(JSON.stringify(preset)) as CanvasPreset;
    
    if (!serialized.id || !serialized.name || !serialized.version) {
      throw new SerializationError('Invalid preset structure for serialization');
    }

    return serialized;
  } catch (error) {
    throw new SerializationError(
      'Failed to serialize preset',
      { error: error instanceof Error ? error.message : String(error) }
    );
  }
};

export const deserializePreset = async (json: string | unknown): Promise<CanvasPreset> => {
  try {
    let parsed: unknown;

    if (typeof json === 'string') {
      parsed = JSON.parse(json);
    } else {
      parsed = json;
    }

    if (typeof parsed !== 'object' || parsed === null) {
      throw new SerializationError('Invalid preset format: not an object');
    }

    const preset = parsed as any;

    if (!preset.id || !preset.name) {
      throw new SerializationError('Invalid preset: missing required fields (id, name)');
    }

    if (!preset.version) {
      preset.version = '0.0.0';
    }

    const migratedPreset = applyAllMigrations(preset);

    return migratedPreset as CanvasPreset;
  } catch (error) {
    if (error instanceof SerializationError) {
      throw error;
    }
    throw new SerializationError(
      'Failed to deserialize preset',
      { error: error instanceof Error ? error.message : String(error) }
    );
  }
};

export const serializeGraph = (graph: Graph): Graph => {
  try {
    const serialized = JSON.parse(JSON.stringify(graph)) as Graph;
    
    if (!serialized.metadata || !serialized.metadata.id) {
      throw new SerializationError('Invalid graph structure for serialization');
    }

    return serialized;
  } catch (error) {
    throw new SerializationError(
      'Failed to serialize graph',
      { error: error instanceof Error ? error.message : String(error) }
    );
  }
};

export const deserializeGraph = async (json: string | unknown): Promise<Graph> => {
  try {
    let parsed: unknown;

    if (typeof json === 'string') {
      parsed = JSON.parse(json);
    } else {
      parsed = json;
    }

    if (typeof parsed !== 'object' || parsed === null) {
      throw new SerializationError('Invalid graph format: not an object');
    }

    const graph = parsed as any;

    if (!graph.metadata || !graph.metadata.id) {
      throw new SerializationError('Invalid graph: missing metadata or id');
    }

    if (!Array.isArray(graph.nodes)) {
      graph.nodes = [];
    }

    if (!Array.isArray(graph.edges)) {
      graph.edges = [];
    }

    return graph as Graph;
  } catch (error) {
    if (error instanceof SerializationError) {
      throw error;
    }
    throw new SerializationError(
      'Failed to deserialize graph',
      { error: error instanceof Error ? error.message : String(error) }
    );
  }
};

export const validateBeforeSave = (data: unknown): boolean => {
  try {
    const serialized = JSON.stringify(data);
    JSON.parse(serialized);
    return true;
  } catch (error) {
    return false;
  }
};

export default {
  serializePreset,
  deserializePreset,
  serializeGraph,
  deserializeGraph,
  validateBeforeSave,
};
