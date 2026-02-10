/**
 * D3FEND Inference Engine
 *
 * Provides intelligent suggestions for D3FEND-based graphs:
 * 1. Sensor detection - identifies potential attack indicators in the graph
 * 2. Defensive technique suggestion - recommends mitigations for detected threats
 * 3. Weakness mapping - connects D3FEND techniques to CWE vulnerabilities
 */

import type { Node, Edge } from '@xyflow/react';
import { D3FEND_CLASSES, type D3FENDClass } from './ontology';

// ============================================================================
// Inference Types
// ============================================================================

/**
 * Severity level for recommendations
 */
export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info';

/**
 * Type of inference result
 */
export type InferenceType = 'sensor' | 'mitigation' | 'detection' | 'weakness';

/**
 * Single inference result
 */
export interface InferenceResult {
  id: string;
  type: InferenceType;
  severity: Severity;
  title: string;
  description: string;
  sourceNodeId: string;
  targetNodeIds: string[];
  recommendedEdgeType: string;
  confidence: number; // 0-100
  metadata?: {
    d3fendClass?: string;
    techniqueIds?: string[];
    cweIds?: string[];
    relatedAttacks?: string[];
  };
}

/**
 * Sensor detection result
 */
export interface SensorDetection {
  nodeId: string;
  nodeType: string;
  sensors: string[];
  coverageScore: number; // 0-100
}

/**
 * Mitigation suggestion
 */
export interface MitigationSuggestion {
  nodeId: string;
  nodeType: string;
  mitigations: {
    d3fendClass: string;
    d3fendLabel: string;
    techniqueIds: string[];
    description: string;
  }[];
}

/**
 * Weakness mapping
 */
export interface WeaknessMapping {
  nodeId: string;
  nodeType: string;
  weaknesses: {
    cweId: string;
    cweName: string;
    severity: Severity;
    mitigatedBy: string[]; // D3FEND technique IDs
  }[];
}

// ============================================================================
// D3FEND to CWE Mappings
// ============================================================================

/**
 * Mappings from D3FEND techniques to CWE weaknesses they mitigate
 */
const D3FEND_TO_CWE_MAPPINGS: Record<string, string[]> = {
  'd3f:Hardening': [
    'CWE-20', 'CWE-89', 'CWE-79', 'CWE-787', 'CWE-22', 'CWE-78', 'CWE-94',
    'CWE-502', 'CWE-295', 'CWE-400', 'CWE-770', 'CWE-732'
  ],
  'd3f:ApplicationHardening': [
    'CWE-20', 'CWE-89', 'CWE-79', 'CWE-787', 'CWE-502'
  ],
  'd3f:PlatformHardening': [
    'CWE-400', 'CWE-770', 'CWE-732', 'CWE-200', 'CWE-269', 'CWE-863'
  ],
  'd3f:Detection': [
    'CWE-94', 'CWE-416', 'CWE-190', 'CWE-119', 'CWE-125'
  ],
  'd3f:NetworkTrafficAnalysis': [
    'CWE-94', 'CWE-311', 'CWE-327', 'CWE-759', 'CWE-352'
  ],
  'd3f:FileAnalysis': [
    'CWE-502', 'CWE-434', 'CWE-94', 'CWE-20'
  ],
  'd3f:ProcessAnalysis': [
    'CWE-94', 'CWE-416', 'CWE-787', 'CWE-119'
  ],
  'd3f:Isolation': [
    'CWE-400', 'CWE-770', 'CWE-732', 'CWE-502'
  ],
  'd3f:Restoration': [
    'CWE-400', 'CWE-770', 'CWE-200'
  ],
};

/**
 * CWE metadata
 */
const CWE_METADATA: Record<string, { name: string; severity: Severity }> = {
  'CWE-20': { name: 'Improper Input Validation', severity: 'high' },
  'CWE-89': { name: 'SQL Injection', severity: 'critical' },
  'CWE-79': { name: 'Cross-site Scripting (XSS)', severity: 'high' },
  'CWE-787': { name: 'Out-of-bounds Write', severity: 'critical' },
  'CWE-22': { name: 'Path Traversal', severity: 'high' },
  'CWE-78': { name: 'OS Command Injection', severity: 'critical' },
  'CWE-94': { name: 'Code Injection', severity: 'critical' },
  'CWE-502': { name: 'Deserialization of Untrusted Data', severity: 'high' },
  'CWE-295': { name: 'Certificate Validation', severity: 'high' },
  'CWE-400': { name: 'Resource Exhaustion', severity: 'medium' },
  'CWE-770': { name: 'Allocation of Resources Without Limits', severity: 'high' },
  'CWE-732': { name: 'Incorrect Permission Assignment', severity: 'high' },
  'CWE-200': { name: 'Exposure of Sensitive Information', severity: 'medium' },
  'CWE-269': { name: 'Improper Privilege Management', severity: 'high' },
  'CWE-863': { name: 'Incorrect Authorization', severity: 'high' },
  'CWE-311': { name: 'Missing Encryption of Sensitive Data', severity: 'medium' },
  'CWE-327': { name: 'Use of Broken or Risky Cryptographic Algorithm', severity: 'high' },
  'CWE-759': { name: 'Use of Hard-coded Cryptographic Key', severity: 'high' },
  'CWE-352': { name: 'CSRF', severity: 'high' },
  'CWE-434': { name: 'Unrestricted Upload of File with Dangerous Type', severity: 'high' },
  'CWE-416': { name: 'Use After Free', severity: 'high' },
  'CWE-190': { name: 'Integer Overflow', severity: 'medium' },
  'CWE-119': { name: 'Buffer Overflow', severity: 'high' },
  'CWE-125': { name: 'Out-of-bounds Read', severity: 'medium' },
};

/**
 * Sensor capabilities by D3FEND node type
 */
const D3FEND_SENSOR_CAPABILITIES: Record<string, { sensors: string[]; coverageScore: number }> = {
  'd3f:NetworkTrafficAnalysis': {
    sensors: ['network-traffic', 'dns-queries', 'http-requests', 'tls-connections'],
    coverageScore: 85
  },
  'd3f:FileAnalysis': {
    sensors: ['file-creation', 'file-modification', 'file-deletion', 'file-access'],
    coverageScore: 80
  },
  'd3f:ProcessAnalysis': {
    sensors: ['process-creation', 'process-termination', 'process-injection', 'process-chain'],
    coverageScore: 90
  },
  'd3f:Hardening': {
    sensors: ['configuration-changes', 'permission-changes', 'authentication-events'],
    coverageScore: 70
  },
  'd3f:Isolation': {
    sensors: ['network-segmentation', 'container-escape', 'privilege-escalation'],
    coverageScore: 65
  },
};

/**
 * Attack indicators that suggest need for specific D3FEND defenses
 */
const ATTACK_INDICATORS: Record<string, { indicator: string; recommended: string[]; severity: Severity }> = {
  'external-connection': {
    indicator: 'External network connection detected',
    recommended: ['d3f:NetworkTrafficAnalysis', 'd3f:Isolation'],
    severity: 'high'
  },
  'file-execution': {
    indicator: 'File execution in suspicious location',
    recommended: ['d3f:ProcessAnalysis', 'd3f:FileAnalysis'],
    severity: 'high'
  },
  'privilege-change': {
    indicator: 'Privilege elevation attempt',
    recommended: ['d3f:Detection', 'd3f:Isolation'],
    severity: 'critical'
  },
  'data-exfiltration': {
    indicator: 'Large data transfer to external endpoint',
    recommended: ['d3f:NetworkTrafficAnalysis', 'd3f:Detection'],
    severity: 'critical'
  },
  'unusual-login': {
    indicator: 'Unusual login pattern detected',
    recommended: ['d3f:Detection', 'd3f:Hardening'],
    severity: 'high'
  },
  'registry-change': {
    indicator: 'Registry or configuration modification',
    recommended: ['d3f:FileAnalysis', 'd3f:Hardening'],
    severity: 'medium'
  },
};

// ============================================================================
// Inference Engine Class
// ============================================================================

export class D3FENDInferenceEngine {
  private nodes: Map<string, Node>;
  private edges: Map<string, Edge>;

  constructor(nodes: Node[], edges: Edge[]) {
    this.nodes = new Map(nodes.map(n => [n.id, n]));
    this.edges = new Map(edges.map(e => [e.id, e]));
  }

  /**
   * Detect sensors in the current graph
   */
  detectSensors(): SensorDetection[] {
    const results: SensorDetection[] = [];

    for (const [nodeId, node] of this.nodes) {
      const nodeType = node.type;
      const capabilities = D3FEND_SENSOR_CAPABILITIES[nodeType!];

      if (capabilities) {
        results.push({
          nodeId,
          nodeType: nodeType!,
          sensors: capabilities.sensors,
          coverageScore: capabilities.coverageScore
        });
      }
    }

    return results;
  }

  /**
   * Get overall sensor coverage score
   */
  getSensorCoverageScore(): number {
    const detections = this.detectSensors();
    if (detections.length === 0) return 0;

    const totalScore = detections.reduce((sum, d) => sum + d.coverageScore, 0);
    return Math.round(totalScore / detections.length);
  }

  /**
   * Suggest mitigations for attack indicators
   */
  suggestMitigations(): MitigationSuggestion[] {
    const results: MitigationSuggestion[] = [];

    // Find nodes that might represent attack indicators
    for (const [nodeId, node] of this.nodes) {
      const nodeType = node.type;
      const indicatorMatch = this.findAttackIndicator(node);

      if (indicatorMatch) {
        const mitigations = indicatorMatch.recommended.map(d3fendClassId => {
          const d3fendClass = D3FEND_CLASSES.find(c => c.id === d3fendClassId);
          return {
            d3fendClass: d3fendClassId,
            d3fendLabel: d3fendClass?.label || d3fendClassId,
            techniqueIds: d3fendClass?.techniques || [],
            description: this.getMitigationDescription(d3fendClassId)
          };
        });

        results.push({
          nodeId,
          nodeType: nodeType!,
          mitigations
        });
      }
    }

    return results;
  }

  /**
   * Map nodes to weaknesses (CWE) and suggest mitigations
   */
  mapWeaknesses(): WeaknessMapping[] {
    const results: WeaknessMapping[] = [];

    for (const [nodeId, node] of this.nodes) {
      const nodeType = node.type;

      // Check if this node type is associated with any D3FEND class
      const nodeData = node.data as { d3fendClass?: string };
      const d3fendClassId = nodeData?.d3fendClass || this.mapNodeTypeToD3FENDClass(nodeType!);

      if (d3fendClassId && D3FEND_TO_CWE_MAPPINGS[d3fendClassId]) {
        const cweIds = D3FEND_TO_CWE_MAPPINGS[d3fendClassId];
        const weaknesses = cweIds.map(cweId => {
          const cwe = CWE_METADATA[cweId];
          return {
            cweId,
            cweName: cwe?.name || cweId,
            severity: cwe?.severity || 'medium',
            mitigatedBy: [d3fendClassId]
          };
        });

        results.push({
          nodeId,
          nodeType: nodeType!,
          weaknesses
        });
      }
    }

    return results;
  }

  /**
   * Generate all inferences for the graph
   */
  generateInferences(selectedNodeId?: string): InferenceResult[] {
    const inferences: InferenceResult[] = [];
    let idCounter = 0;

    // Generate sensor inferences
    const sensors = this.detectSensors();
    sensors.forEach(sensor => {
      if (!selectedNodeId || sensor.nodeId === selectedNodeId) {
        inferences.push({
          id: `sensor-${idCounter++}`,
          type: 'sensor',
          severity: 'info',
          title: `${sensor.nodeType} Active`,
          description: `Monitoring: ${sensor.sensors.join(', ')}`,
          sourceNodeId: sensor.nodeId,
          targetNodeIds: [],
          recommendedEdgeType: '',
          confidence: sensor.coverageScore,
          metadata: {
            coverageScore: sensor.coverageScore
          }
        });
      }
    });

    // Generate mitigation inferences
    const mitigations = this.suggestMitigations();
    mitigations.forEach(suggestion => {
      if (!selectedNodeId || suggestion.nodeId === selectedNodeId) {
        suggestion.mitigations.forEach(mitigation => {
          inferences.push({
            id: `mitigation-${idCounter++}`,
            type: 'mitigation',
            severity: 'high',
            title: mitigation.d3fendLabel,
            description: mitigation.description,
            sourceNodeId: suggestion.nodeId,
            targetNodeIds: [],
            recommendedEdgeType: 'mitigates',
            confidence: 75,
            metadata: {
              d3fendClass: mitigation.d3fendClass,
              techniqueIds: mitigation.techniqueIds
            }
          });
        });
      }
    });

    // Generate weakness inferences
    const weaknesses = this.mapWeaknesses();
    weaknesses.forEach(mapping => {
      if (!selectedNodeId || mapping.nodeId === selectedNodeId) {
        mapping.weaknesses.forEach(weakness => {
          inferences.push({
            id: `weakness-${idCounter++}`,
            type: 'weakness',
            severity: weakness.severity,
            title: `${weakness.cweId}: ${weakness.cweName}`,
            description: `Potential vulnerability mitigated by: ${weakness.mitigatedBy.join(', ')}`,
            sourceNodeId: mapping.nodeId,
            targetNodeIds: [],
            recommendedEdgeType: 'mitigates',
            confidence: 70,
            metadata: {
              cweIds: [weakness.cweId]
            }
          });
        });
      }
    });

    return inferences.sort((a, b) => {
      const severityOrder: Record<Severity, number> = { critical: 0, high: 1, medium: 2, low: 3, info: 4 };
      return severityOrder[a.severity] - severityOrder[b.severity];
    });
  }

  /**
   * Find attack indicators in a node
   */
  private findAttackIndicator(node: Node): { indicator: string; recommended: string[]; severity: Severity } | null {
    const nodeData = node.data as any;
    const label = node.data?.label?.toLowerCase() || '';
    const properties = node.data?.properties || [];

    // Check label for indicator keywords
    for (const [key, value] of Object.entries(ATTACK_INDICATORS)) {
      if (label.includes(key)) {
        return value;
      }
    }

    // Check properties for indicators
    for (const prop of properties) {
      const propValue = (prop.value as string)?.toLowerCase() || '';
      for (const [key, value] of Object.entries(ATTACK_INDICATORS)) {
        if (propValue.includes(key)) {
          return value;
        }
      }
    }

    return null;
  }

  /**
   * Map node type to D3FEND class
   */
  private mapNodeTypeToD3FENDClass(nodeType: string): string | null {
    // Direct mapping for known D3FEND node types
    if (D3FEND_CLASSES.find(c => c.id === nodeType)) {
      return nodeType;
    }

    // Map common node types to D3FEND classes
    const mappings: Record<string, string> = {
      'firewall': 'd3f:Isolation',
      'ids': 'd3f:Detection',
      'ips': 'd3f:Detection',
      'siem': 'd3f:Detection',
      'waf': 'd3f:Hardening',
      'endpoint-protection': 'd3f:ApplicationHardening',
      'server-hardening': 'd3f:PlatformHardening',
      'backup': 'd3f:Restoration',
    };

    return mappings[nodeType] || null;
  }

  /**
   * Get mitigation description for a D3FEND class
   */
  private getMitigationDescription(d3fendClassId: string): string {
    const d3fendClass = D3FEND_CLASSES.find(c => c.id === d3fendClassId);
    return d3fendClass?.description || `Implement ${d3fendClassId} to mitigate this threat`;
  }
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Create inference engine from graph data
 */
export function createInferenceEngine(nodes: Node[], edges: Edge[]): D3FENDInferenceEngine {
  return new D3FENDInferenceEngine(nodes, edges);
}

/**
 * Get inferences for a specific node
 */
export function getNodeInferences(
  nodes: Node[],
  edges: Edge[],
  nodeId: string
): InferenceResult[] {
  const engine = new D3FENDInferenceEngine(nodes, edges);
  return engine.generateInferences(nodeId);
}

/**
 * Get inferences for the entire graph
 */
export function getGraphInferences(nodes: Node[], edges: Edge[]): InferenceResult[] {
  const engine = new D3FENDInferenceEngine(nodes, edges);
  return engine.generateInferences();
}

/**
 * Get sensor coverage summary
 */
export function getSensorCoverage(nodes: Node[], edges: Edge[]): {
  score: number;
  detections: SensorDetection[];
} {
  const engine = new D3FENDInferenceEngine(nodes, edges);
  const detections = engine.detectSensors();
  return {
    score: engine.getSensorCoverageScore(),
    detections
  };
}
