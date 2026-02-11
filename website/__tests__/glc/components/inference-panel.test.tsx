/**
 * Inference Panel Component Tests
 *
 * Note: These tests verify the inference engine functionality.
 * Full integration testing with React components is done via browser testing.
 */

import { describe, it, expect } from 'vitest';
import {
  getSensorCoverage,
  getGraphInferences,
  type Node,
  type Edge,
} from '@/lib/glc/d3fend';

describe('D3FEND Inference', () => {
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
    {
      id: 'node-2',
      type: 'd3f:FileAnalysis',
      position: { x: 100, y: 100 },
      data: {
        label: 'File Analysis',
        typeId: 'd3f:FileAnalysis',
        d3fendClass: 'd3f:FileAnalysis',
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

  it('should calculate sensor coverage score', () => {
    const result = getSensorCoverage(mockNodes, mockEdges);

    expect(result).toBeDefined();
    expect(result.score).toBeGreaterThanOrEqual(0);
    expect(result.score).toBeLessThanOrEqual(100);
    expect(result.detections).toBeDefined();
    expect(Array.isArray(result.detections)).toBe(true);
  });

  it('should detect sensor nodes', () => {
    const result = getSensorCoverage(mockNodes, mockEdges);

    expect(result.detections.length).toBeGreaterThan(0);
    expect(result.detections[0].nodeId).toBeDefined();
    expect(result.detections[0].nodeType).toBeDefined();
    expect(result.detections[0].sensors).toBeDefined();
    expect(Array.isArray(result.detections[0].sensors)).toBe(true);
  });

  it('should generate graph inferences', () => {
    const inferences = getGraphInferences(mockNodes, mockEdges);

    expect(inferences).toBeDefined();
    expect(Array.isArray(inferences)).toBe(true);
    expect(inferences.length).toBeGreaterThan(0);
  });

  it('should return zero score for empty graph', () => {
    const result = getSensorCoverage([], []);

    expect(result.score).toBe(0);
    expect(result.detections).toHaveLength(0);
  });

  it('should have sensor coverage score between sensor scores', () => {
    // NetworkTrafficAnalysis has coverage 85, FileAnalysis has 80
    const result = getSensorCoverage(mockNodes, mockEdges);
    // Average should be (85 + 80) / 2 = 82.5, rounded to 83
    expect(result.score).toBe(83);
  });
});
