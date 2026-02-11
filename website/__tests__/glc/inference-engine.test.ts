/**
 * D3FEND Inference Engine Tests
 */

import { describe, it, expect } from 'vitest';
import {
  D3FENDInferenceEngine,
  createInferenceEngine,
  getNodeInferences,
  getGraphInferences,
  getSensorCoverage,
  type Node,
  type Edge,
} from '@/lib/glc/d3fend/inference-engine';

describe('D3FENDInferenceEngine', () => {
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
    {
      id: 'node-3',
      type: 'firewall',
      position: { x: 200, y: 200 },
      data: {
        label: 'External Connection Detected',
        typeId: 'firewall',
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

  describe('detectSensors', () => {
    it('should detect sensor nodes correctly', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const sensors = engine.detectSensors();

      expect(sensors).toHaveLength(2);
      expect(sensors[0].nodeId).toBe('node-1');
      expect(sensors[0].nodeType).toBe('d3f:NetworkTrafficAnalysis');
      expect(sensors[1].nodeId).toBe('node-2');
      expect(sensors[1].nodeType).toBe('d3f:FileAnalysis');
    });

    it('should return empty array when no sensor nodes exist', () => {
      const noSensorNodes: Node[] = [
        {
          id: 'node-1',
          type: 'custom',
          position: { x: 0, y: 0 },
          data: { label: 'Custom Node' },
        },
      ];

      const engine = createInferenceEngine(noSensorNodes, []);
      const sensors = engine.detectSensors();

      expect(sensors).toHaveLength(0);
    });
  });

  describe('getSensorCoverageScore', () => {
    it('should calculate correct coverage score', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const score = engine.getSensorCoverageScore();

      // (85 + 80) / 2 = 82.5, Math.round(82.5) = 83
      expect(score).toBe(83);
    });

    it('should return 0 when no sensors detected', () => {
      const noSensorNodes: Node[] = [
        {
          id: 'node-1',
          type: 'custom',
          position: { x: 0, y: 0 },
          data: { label: 'Custom Node' },
        },
      ];

      const engine = createInferenceEngine(noSensorNodes, []);
      const score = engine.getSensorCoverageScore();

      expect(score).toBe(0);
    });
  });

  describe('suggestMitigations', () => {
    it('should suggest mitigations for nodes mapped to D3FEND', () => {
      // Node with d3fendClass should generate mitigations
      const mockNodesWithClass: Node[] = [
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

      const engine = createInferenceEngine(mockNodesWithClass, []);
      const mitigations = engine.mapWeaknesses();

      expect(mitigations.length).toBeGreaterThan(0);
      expect(mitigations[0].nodeId).toBe('node-1');
      expect(mitigations[0].weaknesses.length).toBeGreaterThan(0);
    });

    it('should map weaknesses for nodes with D3FEND mapping', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const weaknesses = engine.mapWeaknesses();

      // All three nodes have D3FEND mappings:
      // - node-1: d3f:NetworkTrafficAnalysis
      // - node-2: d3f:FileAnalysis
      // - node-3: firewall -> d3f:Isolation (via mapNodeTypeToD3FENDClass)
      expect(weaknesses.length).toBe(3);

      // Verify node-3 (firewall) maps to d3f:Isolation with CWE weaknesses
      const firewallWeaknesses = weaknesses.find(w => w.nodeId === 'node-3');
      expect(firewallWeaknesses).toBeDefined();
      expect(firewallWeaknesses?.weaknesses.length).toBeGreaterThan(0);
    });

    it('should suggest mitigations for nodes with attack indicators', () => {
      // Create a node with label that contains the exact attack indicator key
      const nodesWithIndicator: Node[] = [
        {
          id: 'node-test',
          type: 'firewall',
          position: { x: 0, y: 0 },
          data: {
            label: 'Detected external-connection to suspicious server',
            typeId: 'firewall',
          },
        },
      ];

      const engine = createInferenceEngine(nodesWithIndicator, []);
      const mitigations = engine.suggestMitigations();

      expect(mitigations.length).toBeGreaterThan(0);
      const testMitigations = mitigations.find(m => m.nodeId === 'node-test');
      expect(testMitigations).toBeDefined();
      expect(testMitigations?.mitigations.length).toBeGreaterThan(0);
      expect(testMitigations?.mitigations.some(m => m.d3fendClass === 'd3f:NetworkTrafficAnalysis')).toBe(true);
    });
  });

  describe('mapWeaknesses', () => {
    it('should map D3FEND nodes to CWE weaknesses', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const weaknesses = engine.mapWeaknesses();

      expect(weaknesses.length).toBeGreaterThan(0);

      // Check that weaknesses have correct structure
      const firstWeakness = weaknesses[0];
      expect(firstWeakness.nodeId).toBeDefined();
      expect(firstWeakness.weaknesses).toBeDefined();
      expect(Array.isArray(firstWeakness.weaknesses)).toBe(true);
    });

    it('should include CWE metadata', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const weaknesses = engine.mapWeaknesses();

      // Find a weakness with CWE data
      const weaknessWithCwe = weaknesses.find(w =>
        w.weaknesses.some(weak => weak.cweId.startsWith('CWE-'))
      );

      expect(weaknessWithCwe).toBeDefined();
      if (weaknessWithCwe) {
        const firstCwe = weaknessWithCwe.weaknesses.find(w => w.cweId === 'CWE-94');
        expect(firstCwe?.cweName).toBe('Code Injection');
      }
    });
  });

  describe('generateInferences', () => {
    it('should generate all inference types', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const inferences = engine.generateInferences();

      expect(inferences.length).toBeGreaterThan(0);

      const types = new Set(inferences.map(i => i.type));
      expect(types.has('sensor')).toBe(true);
      expect(types.has('mitigation') || types.has('weakness')).toBe(true);
    });

    it('should sort inferences by severity', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const inferences = engine.generateInferences();

      // Check that critical items come before high priority items
      const criticalIndex = inferences.findIndex(i => i.severity === 'critical');
      const highIndex = inferences.findIndex(i => i.severity === 'high');

      if (criticalIndex !== -1 && highIndex !== -1) {
        expect(criticalIndex).toBeLessThan(highIndex);
      }
    });

    it('should filter by selected node ID', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      const allInferences = engine.generateInferences();
      const nodeInferences = engine.generateInferences('node-1');

      expect(nodeInferences.length).toBeLessThanOrEqual(allInferences.length);
      nodeInferences.forEach(inf => {
        expect(inf.sourceNodeId).toBe('node-1');
      });
    });
  });
});

describe('Helper Functions', () => {
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

  describe('createInferenceEngine', () => {
    it('should create an instance of D3FENDInferenceEngine', () => {
      const engine = createInferenceEngine(mockNodes, mockEdges);
      expect(engine).toBeInstanceOf(D3FENDInferenceEngine);
    });
  });

  describe('getNodeInferences', () => {
    it('should return inferences for specific node', () => {
      const inferences = getNodeInferences(mockNodes, mockEdges, 'node-1');

      expect(inferences).toBeDefined();
      expect(Array.isArray(inferences)).toBe(true);
      inferences.forEach(inf => {
        expect(inf.sourceNodeId).toBe('node-1');
      });
    });
  });

  describe('getGraphInferences', () => {
    it('should return inferences for entire graph', () => {
      const inferences = getGraphInferences(mockNodes, mockEdges);

      expect(inferences).toBeDefined();
      expect(Array.isArray(inferences)).toBe(true);
      expect(inferences.length).toBeGreaterThan(0);
    });
  });

  describe('getSensorCoverage', () => {
    it('should return score and detections', () => {
      const result = getSensorCoverage(mockNodes, mockEdges);

      expect(result).toBeDefined();
      expect(typeof result.score).toBe('number');
      expect(Array.isArray(result.detections)).toBe(true);
      expect(result.score).toBeGreaterThanOrEqual(0);
      expect(result.score).toBeLessThanOrEqual(100);
    });

    it('should return 0 score when no sensors', () => {
      const noSensorNodes: Node[] = [
        {
          id: 'node-1',
          type: 'custom',
          position: { x: 0, y: 0 },
          data: { label: 'Custom Node' },
        },
      ];

      const result = getSensorCoverage(noSensorNodes, []);

      expect(result.score).toBe(0);
      expect(result.detections).toHaveLength(0);
    });
  });
});
