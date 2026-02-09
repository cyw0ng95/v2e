import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import { presetManager } from '../preset-manager';
import { D3FEND_PRESET } from '../presets/d3fend-preset';

describe('Preset Manager - CRUD Operations', () => {
  beforeEach(() => {
    presetManager.clearBackups();
  });

  afterEach(() => {
    presetManager.clearBackups();
  });

  it('should create new user preset', () => {
    const newPreset = presetManager.createUserPreset();
    
    expect(newPreset).toBeDefined();
    expect(newPreset.isBuiltIn).toBe(false);
    expect(newPreset.id).toContain('user-');
    expect(newPreset.name).toBe('New Preset');
  });

  it('should create preset from base preset', () => {
    const newPreset = presetManager.createUserPreset(D3FEND_PRESET);
    
    expect(newPreset).toBeDefined();
    expect(newPreset.isBuiltIn).toBe(false);
    expect(newPreset.name).toContain('(Copy)');
    expect(newPreset.nodeTypes.length).toBe(D3FEND_PRESET.nodeTypes.length);
  });

  it('should duplicate preset', () => {
    const newPreset = presetManager.duplicatePreset(D3FEND_PRESET);
    
    expect(newPreset).toBeDefined();
    expect(newPreset.isBuiltIn).toBe(false);
    expect(newPreset.id).not.toBe(D3FEND_PRESET.id);
    expect(newPreset.nodeTypes.length).toBe(D3FEND_PRESET.nodeTypes.length);
  });

  it('should update user preset', () => {
    const newPreset = presetManager.createUserPreset();
    const updatedPreset = {
      ...newPreset,
      name: 'Updated Preset',
    };
    
    const result = presetManager.updateUserPreset(updatedPreset);
    
    expect(result.name).toBe('Updated Preset');
    expect(result.updatedAt).not.toBe(newPreset.updatedAt);
  });

  it('should throw error when updating built-in preset', () => {
    expect(() => {
      presetManager.updateUserPreset(D3FEND_PRESET);
    }).toThrow('Cannot update built-in preset');
  });

  it('should delete user preset', () => {
    const newPreset = presetManager.createUserPreset();
    const presetId = newPreset.id;
    
    presetManager.deleteUserPreset(presetId);
    
    const backups = presetManager.getBackups(presetId);
    expect(backups).toHaveLength(0);
  });
});

describe('Preset Manager - Import/Export', () => {
  beforeEach(() => {
    presetManager.clearBackups();
  });

  afterEach(() => {
    presetManager.clearBackups();
  });

  it('should export preset to JSON string', async () => {
    const json = await presetManager.exportPreset(D3FEND_PRESET);
    
    expect(typeof json).toBe('string');
    const parsed = JSON.parse(json);
    expect(parsed.id).toBe(D3FEND_PRESET.id);
  });

  it('should import preset from JSON string', async () => {
    const json = JSON.stringify(D3FEND_PRESET);
    const imported = await presetManager.importPreset(json);
    
    expect(imported).toBeDefined();
    expect(imported.id).toBe(D3FEND_PRESET.id);
  });

  it('should throw error for invalid JSON', async () => {
    await expect(
      presetManager.importPreset('invalid json')
    ).rejects.toThrow();
  });

  it('should throw error for invalid preset structure', async () => {
    const invalidPreset = {
      id: 'test',
    };
    const json = JSON.stringify(invalidPreset);
    
    await expect(
      presetManager.importPreset(json)
    ).rejects.toThrow();
  });
});

describe('Preset Manager - Backup System', () => {
  beforeEach(() => {
    presetManager.clearBackups();
  });

  afterEach(() => {
    presetManager.clearBackups();
  });

  it('should create backup on create', () => {
    const newPreset = presetManager.createUserPreset();
    const backups = presetManager.getBackups(newPreset.id);
    
    expect(backups).toHaveLength(1);
    expect(backups[0].id).toBe(newPreset.id);
  });

  it('should create backup on update', () => {
    const newPreset = presetManager.createUserPreset();
    const updatedPreset = presetManager.updateUserPreset({
      ...newPreset,
      name: 'Updated',
    });
    
    const backups = presetManager.getBackups(newPreset.id);
    expect(backups.length).toBeGreaterThanOrEqual(1);
  });

  it('should restore from backup', () => {
    const newPreset = presetManager.createUserPreset();
    const backups = presetManager.getBackups(newPreset.id);
    const backupTimestamp = backups[0].timestamp;
    
    const restored = presetManager.restoreBackup(newPreset.id, backupTimestamp);
    
    expect(restored).toBeDefined();
    expect(restored?.id).toBe(newPreset.id);
  });

  it('should return null for non-existent backup', () => {
    const restored = presetManager.restoreBackup('non-existent', '2026-02-09');
    
    expect(restored).toBeNull();
  });

  it('should limit backups to max size', () => {
    const presetId = 'test-preset';
    for (let i = 0; i < 10; i++) {
      const preset = presetManager.createUserPreset();
    }
    
    const userPresets = presetManager.getAllBackups();
    const presetBackups = presetManager.getBackups(presetId);
    
    expect(presetBackups.length).toBeLessThanOrEqual(5);
  });

  it('should clear all backups', () => {
    presetManager.createUserPreset();
    presetManager.createUserPreset();
    
    presetManager.clearBackups();
    
    expect(presetManager.getAllBackups()).toHaveLength(0);
  });

  it('should clear backups for specific preset', () => {
    const preset1 = presetManager.createUserPreset();
    const preset2 = presetManager.createUserPreset();
    
    presetManager.clearBackups(preset1.id);
    
    expect(presetManager.getBackups(preset1.id)).toHaveLength(0);
    expect(presetManager.getBackups(preset2.id)).toHaveLength(1);
  });
});
