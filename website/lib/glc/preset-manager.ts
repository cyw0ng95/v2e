import { CanvasPreset, CADNode, CADEdge, Graph } from '../types';
import { validatePreset, validatePresetFile, PresetValidationError } from '../validation';
import { serializePreset, deserializePreset } from './preset-serializer';
import { errorHandler, showError } from '../errors';

export interface PresetBackup {
  id: string;
  preset: CanvasPreset;
  timestamp: string;
  version: string;
}

class PresetManager {
  private static instance: PresetManager;
  private backups: Map<string, PresetBackup[]> = new Map();
  private maxBackupsPerPreset = 5;

  private constructor() {
    this.loadBackups();
  }

  public static getInstance(): PresetManager {
    if (!PresetManager.instance) {
      PresetManager.instance = new PresetManager();
    }
    return PresetManager.instance;
  }

  private loadBackups(): void {
    try {
      const stored = localStorage.getItem('glc-preset-backups');
      if (stored) {
        const backupsArray = JSON.parse(stored) as PresetBackup[];
        backupsArray.forEach(backup => {
          const existing = this.backups.get(backup.id) || [];
          existing.push(backup);
          this.backups.set(backup.id, existing);
        });
      }
    } catch (error) {
      errorHandler.handleError(error, { action: 'load-backups' });
    }
  }

  private saveBackups(): void {
    try {
      const backupsArray: PresetBackup[] = [];
      this.backups.forEach(backups => {
        backupsArray.push(...backups);
      });
      localStorage.setItem('glc-preset-backups', JSON.stringify(backupsArray));
    } catch (error) {
      errorHandler.handleError(error, { action: 'save-backups' });
    }
  }

  private createBackup(preset: CanvasPreset): void {
    const backup: PresetBackup = {
      id: preset.id,
      preset: preset,
      timestamp: new Date().toISOString(),
      version: preset.version,
    };

    const existing = this.backups.get(preset.id) || [];
    existing.push(backup);

    if (existing.length > this.maxBackupsPerPreset) {
      existing.shift();
    }

    this.backups.set(preset.id, existing);
    this.saveBackups();
  }

  private validateAndBackup(preset: CanvasPreset): void {
    const validation = validatePreset(preset);
    
    if (!validation.valid) {
      throw new PresetValidationError(
        `Preset validation failed with ${validation.errors.length} error(s)`,
        validation.errors,
        { presetId: preset.id }
      );
    }

    if (validation.warnings.length > 0) {
      console.warn('Preset validation warnings:', validation.warnings);
    }

    this.createBackup(preset);
  }

  public createUserPreset(basePreset?: CanvasPreset): CanvasPreset {
    const id = `user-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const now = new Date().toISOString();

    const newPreset: CanvasPreset = basePreset 
      ? { 
          ...basePreset, 
          id, 
          isBuiltIn: false, 
          createdAt: now, 
          updatedAt: now,
          author: 'User',
          name: `${basePreset.name} (Copy)`,
        }
      : {
          id,
          name: 'New Preset',
          version: '1.0.0',
          category: 'Custom',
          description: 'Custom preset',
          author: 'User',
          createdAt: now,
          updatedAt: now,
          isBuiltIn: false,
          nodeTypes: [],
          relationshipTypes: [],
          styling: {
            theme: 'light',
            primaryColor: '#3b82f6',
            backgroundColor: '#ffffff',
            gridColor: '#e5e7eb',
            fontFamily: 'Inter, sans-serif',
          },
          behavior: {
            pan: true,
            zoom: true,
            snapToGrid: false,
            gridSize: 10,
            undoRedo: true,
            autoSave: false,
            autoSaveInterval: 60000,
            maxNodes: 1000,
            maxEdges: 2000,
          },
          validationRules: [],
          metadata: {
            tags: [],
          },
        };

    try {
      this.validateAndBackup(newPreset);
      return newPreset;
    } catch (error) {
      errorHandler.handleError(error, { action: 'create-preset' });
      throw error;
    }
  }

  public updateUserPreset(preset: CanvasPreset): CanvasPreset {
    if (preset.isBuiltIn) {
      throw new Error('Cannot update built-in preset');
    }

    const updatedPreset: CanvasPreset = {
      ...preset,
      updatedAt: new Date().toISOString(),
    };

    try {
      this.validateAndBackup(updatedPreset);
      return updatedPreset;
    } catch (error) {
      errorHandler.handleError(error, { action: 'update-preset' });
      throw error;
    }
  }

  public deleteUserPreset(presetId: string): void {
    if (this.backups.has(presetId)) {
      this.backups.delete(presetId);
      this.saveBackups();
    }
  }

  public duplicatePreset(preset: CanvasPreset): CanvasPreset {
    const newPreset = this.createUserPreset(preset);
    return this.updateUserPreset(newPreset);
  }

  public async exportPreset(preset: CanvasPreset): Promise<string> {
    try {
      const json = serializePreset(preset);
      return JSON.stringify(json, null, 2);
    } catch (error) {
      errorHandler.handleError(error, { action: 'export-preset' });
      throw error;
    }
  }

  public async importPreset(json: string): Promise<CanvasPreset> {
    try {
      const preset = await deserializePreset(json);
      this.validateAndBackup(preset);
      return preset;
    } catch (error) {
      errorHandler.handleError(error, { action: 'import-preset' });
      throw error;
    }
  }

  public async importPresetFile(file: File): Promise<CanvasPreset> {
    try {
      const validation = await validatePresetFile(file);
      
      if (!validation.valid) {
        throw new PresetValidationError(
          'Invalid preset file',
          validation.errors,
          { filename: file.name }
        );
      }

      const content = await file.text();
      return this.importPreset(content);
    } catch (error) {
      errorHandler.handleError(error, { action: 'import-preset-file' });
      throw error;
    }
  }

  public restoreBackup(presetId: string, backupTimestamp: string): CanvasPreset | null {
    const backups = this.backups.get(presetId);
    
    if (!backups) {
      return null;
    }

    const backup = backups.find(b => b.timestamp === backupTimestamp);
    
    if (!backup) {
      return null;
    }

    return backup.preset;
  }

  public getBackups(presetId: string): PresetBackup[] {
    return this.backups.get(presetId) || [];
  }

  public getAllBackups(): PresetBackup[] {
    const allBackups: PresetBackup[] = [];
    this.backups.forEach(backups => {
      allBackups.push(...backups);
    });
    return allBackups.sort((a, b) => 
      new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    );
  }

  public clearBackups(presetId?: string): void {
    if (presetId) {
      this.backups.delete(presetId);
    } else {
      this.backups.clear();
    }
    this.saveBackups();
  }
}

export const presetManager = PresetManager.getInstance();

export default presetManager;
