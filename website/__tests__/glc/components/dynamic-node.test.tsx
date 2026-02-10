/**
 * Dynamic Node Component Tests
 */

import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DynamicNode } from '@/components/glc/canvas/dynamic-node';
import type { Node } from '@xyflow/react';

describe('DynamicNode', () => {
  const mockNode: Node = {
    id: 'node-1',
    type: 'glc',
    position: { x: 100, y: 100 },
    data: {
      label: 'Test Node',
      typeId: 'test-type',
      color: '#3b82f6',
      icon: 'Circle',
      properties: [
        { key: 'property1', value: 'value1', type: 'string' },
      ],
    },
    selected: false,
  };

  const mockHandlers = {
    onNodeClick: vi.fn(),
    onNodeDoubleClick: vi.fn(),
  };

  it('should render node with label', () => {
    render(<DynamicNode data={mockNode.data} {...mockHandlers} />);
    expect(screen.getByText('Test Node')).toBeInTheDocument();
  });

  it('should handle node click', () => {
    render(<DynamicNode data={mockNode.data} {...mockHandlers} />);
    const node = screen.getByText('Test Node');
    fireEvent.click(node);
    expect(mockHandlers.onNodeClick).toHaveBeenCalledTimes(1);
  });

  it('should display properties when expanded', () => {
    const dataWithProperties = {
      ...mockNode.data,
      properties: [
        { key: 'property1', value: 'value1', type: 'string' },
        { key: 'property2', value: 'value2', type: 'string' },
      ],
    };

    render(<DynamicNode data={dataWithProperties} {...mockHandlers} />);
    expect(screen.getByText('property1')).toBeInTheDocument();
    expect(screen.getByText('property2')).toBeInTheDocument();
  });

  it('should apply custom color', () => {
    const dataWithColor = {
      ...mockNode.data,
      color: '#ef4444',
    };

    const { container } = render(<DynamicNode data={dataWithColor} {...mockHandlers} />);
    const nodeElement = container.querySelector('.react-flow__node');
    expect(nodeElement).toHaveStyle({ backgroundColor: '#ef4444' });
  });
});
