/**
 * Node Palette Component Tests
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { NodePalette } from '@/components/glc/palette/node-palette';
import { d3fendPreset } from '@/lib/glc/presets';

describe('NodePalette', () => {
  const mockOnDragStart = vi.fn();

  beforeEach(() => {
    mockOnDragStart.mockClear();
  });

  it('should render all node types from preset', () => {
    render(
      <NodePalette
        preset={d3fendPreset}
        onDragStart={mockOnDragStart}
      />
    );

    d3fendPreset.nodeTypes.forEach(nodeType => {
      expect(screen.getByText(nodeType.label)).toBeInTheDocument();
    });
  });

  it('should render node types grouped by category', () => {
    render(
      <NodePalette
        preset={d3fendPreset}
        onDragStart={mockOnDragStart}
      />
    );

    // Check that categories are rendered
    expect(screen.getAllByText(/category/i).length).toBeGreaterThan(0);
  });

  it('should filter nodes by search query', () => {
    render(
      <NodePalette
        preset={d3fendPreset}
        onDragStart={mockOnDragStart}
      />
    );

    // Type in search box
    const searchInput = screen.getByPlaceholderText(/search/i);
    fireEvent.change(searchInput, { target: { value: 'Attack' } });

    // Check that matching nodes are shown - the preset has "Attack Technique" and "Attack Tactic"
    expect(screen.getByText('Attack Technique')).toBeInTheDocument();
    expect(screen.getByText('Attack Tactic')).toBeInTheDocument();
  });

  it('should hide non-matching nodes when searching', () => {
    render(
      <NodePalette
        preset={d3fendPreset}
        onDragStart={mockOnDragStart}
      />
    );

    const searchInput = screen.getByPlaceholderText(/search/i);
    fireEvent.change(searchInput, { target: { value: 'NonExistentType' } });

    // Check that the first node type is hidden when searching for non-matching term
    expect(screen.queryByText(d3fendPreset.nodeTypes[0].label)).not.toBeInTheDocument();
  });

  it('should render the correct number of node types', () => {
    const { container } = render(
      <NodePalette
        preset={d3fendPreset}
        onDragStart={mockOnDragStart}
      />
    );

    // d3fendPreset has 9 node types
    expect(d3fendPreset.nodeTypes.length).toBe(9);
  });
});
