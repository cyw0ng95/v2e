/**
 * Canvas Toolbar Component Tests
 *
 * Note: These tests verify the toolbar can be rendered with proper props.
 * Full integration testing with Zustand store is done via browser testing.
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { CanvasToolbar } from '@/components/glc/toolbar/canvas-toolbar';
import { d3fendPreset } from '@/lib/glc/presets';

// Mock the Zustand store
vi.mock('@/lib/glc/store', () => ({
  useGLCStore: vi.fn(() => ({
    canUndo: true,
    canRedo: false,
    undo: vi.fn(),
    redo: vi.fn(),
    graph: null,
  })),
}));

// Mock useResponsive hook
vi.mock('@/lib/glc/responsive', () => ({
  useResponsive: vi.fn(() => ({
    isMobile: false,
    isTablet: false,
  })),
  TOUCH_TARGET_SIZE: 44,
}));

describe('CanvasToolbar', () => {
  const mockHandlers = {
    onShowShortcuts: vi.fn(),
    onShowExport: vi.fn(),
    onShowShare: vi.fn(),
    onShowInferences: vi.fn(),
    onShowSTIXImport: vi.fn(),
  };

  it('should have access to d3fend preset theme', () => {
    expect(d3fendPreset.theme).toBeDefined();
    expect(d3fendPreset.theme.primary).toBe('#6366f1');
    expect(d3fendPreset.theme.background).toBe('#0f172a');
  });

  it('should have correct number of node types in preset', () => {
    expect(d3fendPreset.nodeTypes.length).toBe(9);
  });

  it('should have correct number of relationships in preset', () => {
    expect(d3fendPreset.relationships.length).toBeGreaterThan(0);
  });

  it('should render without crashing when provided with preset', () => {
    // Basic smoke test - verifies the component can be imported and props are valid
    expect(() => {
      render(
        <CanvasToolbar
          preset={d3fendPreset}
          graphName="Test Graph"
          {...mockHandlers}
        />
      );
    }).not.toThrow();
  });

  it('should accept all required props', () => {
    // Verify that preset has all required properties
    expect(d3fendPreset.meta).toBeDefined();
    expect(d3fendPreset.theme).toBeDefined();
    expect(d3fendPreset.behavior).toBeDefined();
    expect(d3fendPreset.nodeTypes).toBeDefined();
    expect(d3fendPreset.relationships).toBeDefined();
  });
});
