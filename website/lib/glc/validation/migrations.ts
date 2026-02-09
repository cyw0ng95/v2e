import { CanvasPreset } from '../types';

export interface PresetMigration {
  fromVersion: string;
  toVersion: string;
  migrate: (preset: any) => CanvasPreset;
}

export const migrateFrom090To100: PresetMigration = {
  fromVersion: '0.9.0',
  toVersion: '1.0.0',
  migrate: (preset: any): CanvasPreset => {
    const migrated = { ...preset };

    migrated.version = '1.0.0';

    if (!migrated.behavior) {
      migrated.behavior = {
        pan: true,
        zoom: true,
        snapToGrid: false,
        gridSize: 10,
        undoRedo: true,
        autoSave: false,
        autoSaveInterval: 60000,
        maxNodes: 1000,
        maxEdges: 2000,
      };
    } else {
      if (migrated.behavior.autoSave === undefined) {
        migrated.behavior.autoSave = false;
      }
      if (migrated.behavior.autoSaveInterval === undefined) {
        migrated.behavior.autoSaveInterval = 60000;
      }
      if (migrated.behavior.maxNodes === undefined) {
        migrated.behavior.maxNodes = 1000;
      }
      if (migrated.behavior.maxEdges === undefined) {
        migrated.behavior.maxEdges = 2000;
      }
    }

    if (!migrated.styling) {
      migrated.styling = {
        theme: 'light',
        primaryColor: '#3b82f6',
        backgroundColor: '#ffffff',
        gridColor: '#e5e7eb',
        fontFamily: 'Inter, sans-serif',
      };
    }

    if (!migrated.metadata) {
      migrated.metadata = {
        tags: [],
        previewImage: undefined,
        documentationUrl: undefined,
      };
    }

    if (!migrated.validationRules) {
      migrated.validationRules = [];
    }

    if (migrated.nodeTypes) {
      migrated.nodeTypes = migrated.nodeTypes.map((nodeType: any) => {
        const migratedNodeType = { ...nodeType };

        if (!migratedNodeType.style) {
          migratedNodeType.style = {
            backgroundColor: '#3b82f6',
            borderColor: '#2563eb',
            textColor: '#ffffff',
            borderRadius: 8,
          };
        }

        if (!migratedNodeType.properties) {
          migratedNodeType.properties = [];
        }

        if (!migratedNodeType.ontologyMappings) {
          migratedNodeType.ontologyMappings = [];
        }

        return migratedNodeType;
      });
    }

    if (migrated.relationshipTypes) {
      migrated.relationshipTypes = migrated.relationshipTypes.map((relType: any) => {
        const migratedRelType = { ...relType };

        if (!migratedRelType.style) {
          migratedRelType.style = {
            strokeColor: '#3b82f6',
            strokeWidth: 2,
          };
        }

        if (!migratedRelType.directionality) {
          migratedRelType.directionality = 'directed';
        }

        if (!migratedRelType.multiplicity) {
          migratedRelType.multiplicity = 'one-to-many';
        }

        if (!migratedRelType.properties) {
          migratedRelType.properties = [];
        }

        return migratedRelType;
      });
    }

    return migrated as CanvasPreset;
  },
};

export const MIGRATION_REGISTRY: PresetMigration[] = [
  migrateFrom090To100,
];

export const getCurrentVersion = (): string => {
  return '1.0.0';
};

export const getPresetVersion = (preset: any): string => {
  return preset?.version || '0.0.0';
};

export const needsMigration = (preset: any): boolean => {
  const presetVersion = getPresetVersion(preset);
  const currentVersion = getCurrentVersion();
  
  if (presetVersion === currentVersion) {
    return false;
  }

  const hasMigration = MIGRATION_REGISTRY.some(
    migration => migration.fromVersion === presetVersion
  );

  return hasMigration;
};

export const migratePreset = (preset: any): CanvasPreset => {
  const presetVersion = getPresetVersion(preset);
  
  if (presetVersion === getCurrentVersion()) {
    return preset as CanvasPreset;
  }

  let migratedPreset = preset;
  
  for (const migration of MIGRATION_REGISTRY) {
    if (getPresetVersion(migratedPreset) === migration.fromVersion) {
      migratedPreset = migration.migrate(migratedPreset);
    }
  }

  if (getPresetVersion(migratedPreset) !== getCurrentVersion()) {
    throw new Error(
      `Cannot migrate preset from version ${presetVersion} to ${getCurrentVersion()}. ` +
      `No migration path available.`
    );
  }

  return migratedPreset as CanvasPreset;
};

export const applyAllMigrations = (preset: any): CanvasPreset => {
  return migratePreset(preset);
};

export const getMigrationPath = (preset: any): string[] => {
  const presetVersion = getPresetVersion(preset);
  const currentVersion = getCurrentVersion();
  const path: string[] = [];

  if (presetVersion === currentVersion) {
    return path;
  }

  let workingVersion = presetVersion;
  
  for (const migration of MIGRATION_REGISTRY) {
    if (workingVersion === migration.fromVersion) {
      path.push(`${workingVersion} → ${migration.toVersion}`);
      workingVersion = migration.toVersion;
      
      if (workingVersion === currentVersion) {
        break;
      }
    }
  }

  return path;
};

export const validateMigrationPath = (preset: any): { valid: boolean; path: string[]; error?: string } => {
  try {
    const path = getMigrationPath(preset);
    const finalVersion = getPresetVersion(preset);
    
    if (path.length > 0 && finalVersion !== getCurrentVersion()) {
      return {
        valid: false,
        path,
        error: `Cannot migrate from ${finalVersion} to ${getCurrentVersion()} - incomplete migration path`,
      };
    }

    return {
      valid: true,
      path,
    };
  } catch (error) {
    return {
      valid: false,
      path: [],
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
};

export default {
  migratePreset,
  applyAllMigrations,
  needsMigration,
  getCurrentVersion,
  getPresetVersion,
  getMigrationPath,
  validateMigrationPath,
  MIGRATION_REGISTRY,
};
