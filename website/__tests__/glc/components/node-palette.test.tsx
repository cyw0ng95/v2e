/**
 * Node Palette Component Tests
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { NodePalette } from '@/components/glc/palette/node-palette';
import { D3FEND_PRESET } from '@/lib/glc/presets';

describe('NodePalette', () => {
  const mockOnDragStart = vi.fn();

  beforeEach(() => {
    mockOnDragStart.mockClear();
  });

  it('should render all node types from preset', () => {
    render(
      <NodePalette
        preset={D3FEND_PRESET}
        onDragStart={mockOnDragStart}
      />
    );

    D3FEND_PRESET.nodeTypes.forEach(nodeType => {
      expect(screen.getByText(nodeType.label)).toBeInTheDocument();
    });
  });

  it('should render node types grouped by category', () => {
    render(
      <NodePalette
        preset={D3FEND_PRESET}
        onDragStart={mockOnDragStart}
      />
    );

    // Check that categories are rendered
    expect(screen.getAllByText(/category/i).length).toBeGreaterThan(0);
  });

  it('should filter nodes by search query', () => {
    render(
      <NodePalette
        preset={D3FEND_PRESET}
        onDragStart={mockOnDragStart}
      />
    );

    // Type in search box
    const searchInput = screen.getByPlaceholderText(/search/i);
    fireEvent.change(searchInput, { target: { value: 'Network' } });

    // Check that only matching nodes are shown
    expect(screen.getByText('Network Traffic Analysis')).toBeInTheDocument();
  });

  it('should hide non-matching nodes when searching', () => {
    render(
      <NodePalette
        preset={D3FEND_PRESET}
        onDragStart={mockOnDragStart}
      />
    );

    const searchInput = screen.getByPlaceholderText(/search/i);
    fireEvent.change(searchInput, { target: { value: 'NonExistentType' } });

    expect(screen.queryByText('Network Traffic Analysis')).not.toBeInTheDocument();
  });

  it('should call onDragStart when dragging node', () => {
    render(
      <NodePalette
        preset={D3FEND_PRESET}
        onDragStart={mockOnDragStart}
      />
    );

    const firstNode = screen.getByText(D3FEND_PRESET.nodeTypes[0].label);
    fireEvent.dragStart(firstNode);

    expect(mockOnDragStart).toHaveBeenCalledTimes(1);
    expect(mockOnDragStart).toHaveBeenCalledWith(
      expect.objectContaining({
        id: D3FEND_PRESET.nodeTypes[0].id,
        label: D3FEND_PRESET.nodeTypes[0].label,
      })
    );
  });
});
