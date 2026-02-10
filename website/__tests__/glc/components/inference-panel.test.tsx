/**
 * Inference Panel Component Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { InferencePanel } from '@/components/glc/d3fend/inference-panel';
import type { Node, Edge } from '@xyflow/react';

describe('InferencePanel', () => {
  const mockNodes: Node[] = [
    {
      id: 'node-1',
      type: 'd3f:NetworkTrafficAnalysis',
      position: { x: 0, y: 0 },
      data: {
        label: 'Network Traffic Analysis',
        typeId: 'd3f:NetworkTrafficAnalysis',
        d3fendClass: 'd3f:NetworkTrafficAnalysis',
      },
    },
  ];

  const mockEdges: Edge[] = [
    {
      id: 'edge-1',
      source: 'node-1',
      target: 'node-2',
      type: 'glc',
      data: { relationshipType: 'connects' },
    },
  ];

  const mockOnClose = vi.fn();

  it('should render sensor coverage gauge', () => {
    render(
      <InferencePanel
        nodes={mockNodes}
        edges={mockEdges}
        isOpen={true}
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText(/sensor coverage/i)).toBeInTheDocument();
  });

  it('should display coverage score', () => {
    render(
      <InferencePanel
        nodes={mockNodes}
        edges={mockEdges}
        isOpen={true}
        onClose={mockOnClose}
      />
    );

    // Score should be visible
    expect(screen.queryByText(/\d+/)).toBeInTheDocument();
  });

  it('should render active sensors', () => {
    render(
      <InferencePanel
        nodes={mockNodes}
        edges={mockEdges}
        isOpen={true}
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText(/active sensors/i)).toBeInTheDocument();
  });

  it('should not render when closed', () => {
    const { container } = render(
      <InferencePanel
        nodes={mockNodes}
        edges={mockEdges}
        isOpen={false}
        onClose={mockOnClose}
      />
    );

    expect(container.firstChild).toBeNull();
  });

  it('should call onClose when close button clicked', () => {
    render(
      <InferencePanel
        nodes={mockNodes}
        edges={mockEdges}
        isOpen={true}
        onClose={mockOnClose}
      />
    );

    const closeButton = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('should display "No Critical Issues" message when no issues', () => {
    render(
      <InferencePanel
        nodes={[]}
        edges={[]}
        isOpen={true}
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText(/no critical issues/i)).toBeInTheDocument();
  });
});
