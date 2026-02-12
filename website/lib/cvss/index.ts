/**
 * CVSS Metrics Library
 *
 * Exports all CVSS metric types and utility functions
 * Phase 5.1 Performance Optimization - Code Splitting
 */

// ============================================================================
// Types
// ============================================================================

/**
 * CVSS v3.x Base Metrics
 */
export interface CVSS3BaseMetrics {
  AV: AV;
  AC: AC;
  PR: PR;
  UI: UI;
  S: S;
  /** Attack Vector (N/A/L/P) */
  E: 'X' | 'U' | 'R' | 'C' | 'N' | 'A';
  /** Privileges Required (N/L/H) */
  PR: PR;
  /** User Interaction (N/R) */
  UI: UI;
  /** Scope (U/C) */
  S: S;
  /** Confidentiality Impact (H/L/N) */
  C: C;
  /** Integrity Impact (H/L/N) */
  I: I;
  /** Availability Impact (H/L/N) */
  A: A;
  /** Provider (E for Exploitability, M for Impact) */
  E: 'X' | 'U' | 'R' | 'C' | 'N';
  /** Modified Base Scope (X/U/C) - v3.1 only */
  MS?: 'X' | 'U' | 'C';
  /** Confidentiality Requirement (H/M/L/N) - v3.1 only */
  CR?: 'H' | 'M' | 'L' | 'N';
  /** Remediation Level (X/U/O/T/W) */
  RL: 'X' | 'U' | 'O' | 'T' | 'W';
  /** Report Confidence (X/U/C/R) */
  RC: 'X' | 'U' | 'C' | 'R';
  /** Modified Base Scope (X/U/C) - v3.1 only */
  MS?: 'X' | 'U' | 'C' | 'N';
}

/**
 * CVSS v3.x Temporal Metrics
 */
export interface CVSS3TemporalMetrics extends CVSS3BaseMetrics {
  /** Exploit Code Maturity (X/U/F/P/H/R) */
  E: 'X' | 'U' | 'F' | 'P' | 'H' | 'R';
}

/**
 * CVSS v3.x Environmental Metrics
 */
export interface CVSS3EnvironmentalMetrics extends CVSS3BaseMetrics {
  /** Confidentiality Requirement (H/M/L/N) */
  CR?: 'H' | 'M' | 'L' | 'N';
}

/**
 * CVSS v3.x Score Breakdown
 */
export interface CVSS3ScoreBreakdown extends CVSS3BaseMetrics {
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Temporal score */
  temporalScore?: number;
  /** Environmental score */
  environmentalScore?: number;
  /** Exploitability sub-score */
  impactScore?: number;
  /** Base severity */
  baseSeverity: CVSSSeverity;
  /** Temporal severity */
  temporalSeverity?: CVSSSeverity;
  /** Environmental severity */
  /** Final severity (computed from sub-scores) */
  finalSeverity?: CVSSSeverity;
}

/**
 * CVSS v4.0 Score Breakdown
 */
export interface CVSS4ScoreBreakdown {
  /** Base score (0.0-10.0) */
  baseScore: number;
  /** Threat score */
  environmentalScore?: number;
  /** Exploitability sub-score */
  impactScore?: number;
  /** Base severity */
  baseSeverity: CVSSSeverity;
  /** Threat severity */
  threatSeverity?: CVSSSeverity;
  /** Environmental severity */
  /** Final severity (computed from sub-scores) */
  finalSeverity?: CVSSSeverity;
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

/**
 * Update a single CVSS metric in the state
 *
 * @param metric - The metric to update (e.g., 'AV', 'AC', 'temporal.E')
 * @param value - The value to set
 * @param label - The metric label (for enums, e.g., 'High', 'Medium')
 *
 * Note: Uses discriminated union to ensure type safety
 */
export function updateMetric(
  metric: 'AV' | 'AC' | 'PR' | 'UI' | 'S' | 'C' | 'I' | 'A' | 'temporal.E' | 'temporal.RL' | 'temporal.RC' | 'environmental.C' | 'threat',
  value: string | number,
  label?: string
): void {
  const setCVSSState = useCVSSStore();

  // Validate parameters
  if (!value || typeof value !== 'string') {
    console.warn(`updateMetric: Invalid value type for metric ${metric}. Expected string, got ${typeof value}`);
    return;
  }

  setCVSSState(prevState => {
    const metrics = { ...prevState.scores, [metric]: { value, label } };
    return { ...prevState, scores: metrics };
  });
}

/**
 * Type Guards
 *
 * Check if value is valid for the given metric type
 */
function isValidMetricValue(metric: string, value: any): boolean {
  switch (metric) {
    case 'AV':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'AC':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'PR':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'UI':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'S':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'C':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'I':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'A':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'temporal.E':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'temporal.RL':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'temporal.RC':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'environmental.C':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
    case 'threat':
      return typeof value === 'string' && ['N', 'A', 'L', 'P'].includes(value);
  default:
      console.warn(`updateMetric: Unknown metric ${metric} - accepting value of type ${typeof value}`);
      return;
  }
  }
}

/**
 * Re-export updateMetric from components
 *
 * Makes components use the new centralized updateMetric function
 */
// Re-export from current locations
export { updateMetric } from '@/lib/cvss/metrics';
export type { CVSS3Metrics, CVSS3TemporalMetrics, CVSS3EnvironmentalMetrics, CVSS3ScoreBreakdown, CVSS4ScoreBreakdown } from '@/lib/cvss/types';
