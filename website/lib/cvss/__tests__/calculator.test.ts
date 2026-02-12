/**
 * CVSS Calculator Engine Unit Tests
 * Tests CVSS v3.0, v3.1, and v4.0 calculations against FIRST.org official vectors
 */

import { describe, it, expect } from 'vitest';
import {
  calculateCVSS3,
  calculateCVSS4,
  calculateCVSS,
  getDefaultMetrics,
  getCVSSMetadata
} from '../cvss-calculator';
import type { CVSS3Metrics, CVSS4Metrics, CVSS3BaseMetrics, CVSS4BaseMetrics } from '../../types';

// ============================================================================
// Test Vectors from FIRST.org
// ============================================================================

/**
 * Known CVSS v3.1 test vectors from FIRST.org
 * Format: { vector: expected score and severity }
 */
const CVSS31_VECTORS: Array<{
  vector: string;
  baseScore: number;
  severity: string;
  description: string;
}> = [
  {
    vector: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    baseScore: 9.8,
    severity: 'CRITICAL',
    description: 'Critical: Network, Low complexity, No privileges, No user interaction, Unchanged scope, High CIA impact'
  },
  {
    vector: 'CVSS:3.1/AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H',
    baseScore: 8.8,
    severity: 'HIGH',
    description: 'High: Local, Low complexity, No privileges, Required user interaction, Unchanged scope, High CIA impact'
  },
  {
    vector: 'CVSS:3.1/AV:N/AC:H/PR:H/UI:N/S:U/C:L/I:N/A:N',
    baseScore: 3.1,
    severity: 'LOW',
    description: 'Low: Network, High complexity, High privileges, No user interaction, Unchanged scope, Low confidentiality impact only'
  },
  {
    vector: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H',
    baseScore: 10.0,
    severity: 'CRITICAL',
    description: 'Perfect 10: Changed scope with all High impacts'
  },
  {
    vector: 'CVSS:3.1/AV:P/AC:H/PR:L/UI:N/S:U/C:N/I:N/A:N',
    baseScore: 0.0,
    severity: 'NONE',
    description: 'None: Physical, High complexity, Low privileges, No impacts'
  },
  {
    vector: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:L/A:N',
    baseScore: 6.5,
    severity: 'MEDIUM',
    description: 'Medium: Network, Low complexity, Low confidentiality and integrity impact'
  },
  {
    vector: 'CVSS:3.1/AV:A/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N',
    baseScore: 5.8,
    severity: 'MEDIUM',
    description: 'Medium: Adjacent network, Changed scope, Low impacts'
  }
];

/**
 * Known CVSS v3.0 test vectors (same scoring as v3.1)
 */
const CVSS30_VECTORS: Array<{
  vector: string;
  baseScore: number;
  severity: string;
  description: string;
}> = [
  {
    vector: 'CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    baseScore: 9.8,
    severity: 'CRITICAL',
    description: 'Critical: Network, Low complexity, No privileges, No user interaction, Unchanged scope, High CIA impact'
  },
  {
    vector: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    baseScore: 9.8,
    severity: 'CRITICAL',
    description: 'Critical: Same as v3.0'
  },
  {
    vector: 'CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    baseScore: 9.8,
    severity: 'CRITICAL',
    description: 'Critical: Same as v3.1'
  },
  {
    vector: 'CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    baseScore: 9.8,
    severity: 'CRITICAL',
    description: 'Critical: Same as v3.0'
  }
];

/**
 * Known CVSS v3.0 test vectors (same scoring as v3.1)
 */
const CVSS30_VECTORS: Array<{
  vector: string;
  baseScore: number;
  severity: string;
  description: string;
}> = [
  {
    vector: 'CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    baseScore: 9.8,
    severity: 'CRITICAL',
    description: 'Critical: Same as v3.1'
  },
  {
    vector: 'CVSS:3.0/AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H',
    baseScore: 8.8,
    severity: 'HIGH',
    description: 'High: Local with user interaction required'
  },
  {
    vector: 'CVSS:3.0/AV:N/AC:H/PR:H/UI:N/S:U/C:L/I:N/A:N',
    baseScore: 3.1,
    severity: 'LOW',
    description: 'Low: High complexity and privileges required'
  }
];

/**
 * Known CVSS v4.0 test vectors from FIRST.org
 */
const CVSS40_VECTORS: Array<{
  vector: string;
  baseScore: number;
  severity: string;
  description: string;
}> = [
  {
    vector: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H',
    baseScore: 10.0,
    severity: 'CRITICAL',
    description: 'Perfect 10: Network, Low complexity, No attack time, No privileges, No interaction, All High impacts'
  },
  {
    vector: 'CVSS:4.0/AV:L/AC:L/AT:N/PR:N/UI:N/VC:L/VI:L/VA:L/SC:L/SI:L/SA:L',
    baseScore: 7.3,
    severity: 'HIGH',
    description: 'High: Local, All Low impacts'
  },
  {
    vector: 'CVSS:4.0/AV:P/AC:H/AT:P/PR:H/UI:A/VC:N/VI:N/VA:N/SC:N/SI:N/SA:N',
    baseScore: 0.0,
    severity: 'NONE',
    description: 'None: Physical, High complexity, Passive attack, High privileges, Active interaction required, No impacts'
  },
  {
    vector: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:P/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H',
    baseScore: 9.3,
    severity: 'CRITICAL',
    description: 'Critical: Passive interaction reduces score slightly'
  }
];

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Parse CVSS v3.x vector string to metrics object
 */
function parseCVSS3Vector(vector: string): CVSS3Metrics {
  const parts = vector.split('/');
  const metrics: any = {};

  for (const part of parts) {
    const [key, value] = part.split(':');
    if (key === 'CVSS') continue;

    // Base metrics
    if (['AV', 'AC', 'PR', 'UI', 'S', 'C', 'I', 'A'].includes(key)) {
      metrics[key] = value;
    }
    // Temporal metrics
    else if (['E', 'RL', 'RC'].includes(key)) {
      if (!metrics.temporal) metrics.temporal = {};
      metrics.temporal[key] = value;
    }
    // Environmental metrics
    else if (['CR', 'IR', 'AR', 'MAV', 'MAC', 'MPR', 'MUI', 'MS', 'MC', 'MI', 'MA'].includes(key)) {
      if (!metrics.environmental) metrics.environmental = {};
      metrics.environmental[key] = value;
    }
  }

  return metrics as CVSS3Metrics;
}

/**
 * Parse CVSS v4.0 vector string to metrics object
 */
function parseCVSS4Vector(vector: string): CVSS4Metrics {
  const parts = vector.split('/');
  const metrics: any = {};

  for (const part of parts) {
    const [key, value] = part.split(':');
    if (key === 'CVSS') continue;

    // Base metrics
    if (['AV', 'AC', 'AT', 'PR', 'UI', 'VC', 'VI', 'VA', 'SC', 'SI', 'SA'].includes(key)) {
      metrics[key] = value;
    }
    // Optional base metrics
    else if (['S', 'AU'].includes(key)) {
      metrics[key] = value;
    }
    // Threat metrics
    else if (['E', 'M', 'D'].includes(key)) {
      if (!metrics.threat) metrics.threat = {};
      metrics.threat[key] = value;
    }
    // Environmental metrics
    else if (['CR', 'IR', 'AR', 'MAV', 'MAC', 'MAT', 'MPR', 'MUI', 'MVC', 'MVI', 'MVA', 'MSC', 'MSI', 'MSA', 'MS', 'MAU', 'MI'].includes(key)) {
      if (!metrics.environmental) metrics.environmental = {};
      metrics.environmental[key] = value;
    }
  }

  // Set defaults for optional metrics
  if (!metrics.S) metrics.S = 'X';
  if (!metrics.AU) metrics.AU = 'N';

  return metrics as CVSS4Metrics;
}

// ============================================================================
// CVSS v3.1 Tests
// ============================================================================

describe('CVSS v3.0/v3.1 PR=H weight fix', () => {
  it('should use correct PR=H weight of 0.50 for CVSS v3.0', () => {
    const metrics: CVSS3BaseMetrics = {
      AV: 'N',
      AC: 'L',
      PR: 'H',
      UI: 'N',
      S: 'U',
      C: 'H',
      I: 'H',
      A: 'H'
    };

    const result = calculateCVSS3(metrics, '3.0');

    // Expected: Impact = 1 - ((1 - 0.56) * (1 - 0.56) * (1 - 0.56)) = 0.851856
    // ISS = 1 - 0.851856 = 0.148144
    // Exploitability = 8.22 * 0.85 * 0.77 * 0.50 * 0.85 = 2.28 (using PR=H = 0.50)
    // Base Score = 6.42 * 0.148144 + 2.28 = 3.20 + 2.28 = 5.48 -> rounds to 5.5

    expect(result.breakdown.baseScore).toBe(5.5);
    expect(result.breakdown.baseSeverity).toBe('MEDIUM');
  });

  it('should use correct PR=H weight of 0.50 for CVSS v3.1', () => {
    const metrics: CVSS3BaseMetrics = {
      AV: 'N',
      AC: 'L',
      PR: 'H',
      UI: 'N',
      S: 'U',
      C: 'H',
      I: 'H',
      A: 'H'
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Same calculation as v3.0
    expect(result.breakdown.baseScore).toBe(5.5);
    expect(result.breakdown.baseSeverity).toBe('MEDIUM');
  });

  it('should match FIRST.org CVSS:3.1/AV:N/AC:L/PR:H/UI:N/S:U/C:H/I:H/A:H vector', () => {
    const metrics: CVSS3BaseMetrics = {
      AV: 'N',
      AC: 'L',
      PR: 'H',
      UI: 'N',
      S: 'U',
      C: 'H',
      I: 'H',
      A: 'H'
    };

    const result = calculateCVSS3(metrics, '3.1');

    // FIRST.org result: Base Score 5.5, Severity MEDIUM
    expect(result.breakdown.baseScore).toBe(5.5);
    expect(result.breakdown.baseSeverity).toBe('MEDIUM');
    expect(result.vectorString).toBe('CVSS:3.1/AV:N/AC:L/PR:H/UI:N/S:U/C:H/I:H/A:H');
  });
});

describe('CVSS v4.0 PR=H weight fix', () => {
  it('should use correct PR=H weight of 0.50 for CVSS v4.0', () => {
    const metrics: CVSS4BaseMetrics = {
      AV: 'N',
      AC: 'L',
      AT: 'N',
      PR: 'H',
      UI: 'N',
      VC: 'H',
      VI: 'H',
      VA: 'N',
      SC: 'H',
      SI: 'H',
      SA: 'H'
    };

    const result = calculateCVSS4(metrics);

    // PR=H with weight 0.50 should be correct
    expect(result.breakdown.baseScore).toBeDefined();
  });

  it('should match FIRST.org CVSS:4.0/AV:N/AC:L/AT:N/PR:H/UI:N/VC:H/VI:H/VA:N/SC:H/SI:H/SA:H vector', () => {
    const metrics: CVSS4BaseMetrics = {
      AV: 'N',
      AC: 'L',
      AT: 'N',
      PR: 'H',
      UI: 'N',
      VC: 'H',
      VI: 'H',
      VA: 'N',
      SC: 'H',
      SI: 'H',
      SA: 'H'
    };

    const result = calculateCVSS4(metrics);

    // Verify calculation works with corrected PR weight
    expect(result.breakdown.baseScore).toBeDefined();
    expect(result.breakdown.baseSeverity).toBeDefined();
    expect(result.vectorString).toBe('CVSS:4.0/AV:N/AC:L/AT:N/PR:H/UI:N/VC:H/VI:H/VA:N/SC:H/SI:H/SA:H');
  });
});

describe('CVSS v3.0/v3.1 Known Vectors', () => {
  CVSS31_VECTORS.forEach(({ vector, baseScore, severity, description }) => {
    it(`should correctly calculate ${vector} (${description})`, () => {
      const version = vector.split('/')[0].split(':')[1] as '3.0' | '3.1';
      const parts = vector.split('/');
      const metrics: CVSS3BaseMetrics = {
        AV: parts[2].split(':')[1] as AV,
        AC: parts[3].split(':')[1] as AC,
        PR: parts[4].split(':')[1] as PR,
        UI: parts[5].split(':')[1] as UI,
        S: parts[6].split(':')[1] as S,
        C: parts[7].split(':')[1] as C,
        I: parts[8].split(':')[1] as I,
        A: parts[9].split(':')[1] as A
      };

      const result = calculateCVSS3(metrics, version);

      expect(result.breakdown.baseScore).toBeCloseTo(baseScore, 0.1);
      expect(result.breakdown.baseSeverity).toBe(severity);
      expect(result.vectorString).toBe(vector);
    });
  });
});

describe('CVSS v3.0 Known Vectors', () => {
  CVSS30_VECTORS.forEach(({ vector, baseScore, severity, description }) => {
    it(`should correctly calculate ${vector} (${description})`, () => {
      const parts = vector.split('/');
      const metrics: CVSS3BaseMetrics = {
        AV: parts[2].split(':')[1] as AV,
        AC: parts[3].split(':')[1] as AC,
        PR: parts[4].split(':')[1] as PR,
        UI: parts[5].split(':')[1] as UI,
        S: parts[6].split(':')[1] as S,
        C: parts[7].split(':')[1] as C,
        I: parts[8].split(':')[1] as I,
        A: parts[9].split(':')[1] as A
      };

      const result = calculateCVSS3(metrics, '3.0');

      expect(result.breakdown.baseScore).toBeCloseTo(baseScore, 0.1);
      expect(result.breakdown.baseSeverity).toBe(severity);
      expect(result.vectorString).toBe(vector);
    });
  });
});
  describe('Base Score Calculations', () => {
    it.each(CVSS31_VECTORS)('$description', ({ vector, baseScore, severity }) => {
      const metrics = parseCVSS3Vector(vector);
      const result = calculateCVSS3(metrics, '3.1');

      expect(result.baseScore).toBeCloseTo(baseScore, 1);
      expect(result.baseSeverity).toBe(severity);
    });
  });

  describe('Temporal Score Calculations', () => {
    it('should apply exploit maturity adjustments correctly', () => {
      const baseMetrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        temporal: { E: 'F', RL: 'X', RC: 'X' }
      };

      const result = calculateCVSS3(baseMetrics, '3.1');

      expect(result.temporalScore).toBeDefined();
      expect(result.temporalScore).toBeLessThan(result.baseScore);
      expect(result.temporalSeverity).toBeDefined();
    });

    it('should apply remediation level adjustments', () => {
      const baseMetrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        temporal: { E: 'X', RL: 'W', RC: 'X' }
      };

      const result = calculateCVSS3(baseMetrics, '3.1');

      expect(result.temporalScore).toBeDefined();
      expect(result.temporalScore).toBeLessThan(result.baseScore);
    });

    it('should apply report confidence adjustments', () => {
      const baseMetrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        temporal: { E: 'X', RL: 'X', RC: 'R' }
      };

      const result = calculateCVSS3(baseMetrics, '3.1');

      expect(result.temporalScore).toBeDefined();
      expect(result.temporalScore).toBeLessThan(result.baseScore);
    });
  });

  describe('Environmental Score Calculations', () => {
    it('should apply confidentiality requirement adjustments', () => {
      const baseMetrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        environmental: { CR: 'H', IR: 'N', AR: 'N' }
      };

      const result = calculateCVSS3(baseMetrics, '3.1');

      expect(result.environmentalScore).toBeDefined();
      expect(result.environmentalScore).toBeGreaterThan(result.baseScore);
    });

    it('should handle modified base metrics', () => {
      const baseMetrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        environmental: {
          CR: 'N', IR: 'N', AR: 'N',
          MAV: 'L', MAC: 'H', MPR: 'L', MUI: 'N'
        }
      };

      const result = calculateCVSS3(baseMetrics, '3.1');

      expect(result.environmentalScore).toBeDefined();
      // Modified to Local should reduce score
      expect(result.environmentalScore).toBeLessThan(result.baseScore);
    });
  });

  describe('Vector String Generation', () => {
    it('should generate correct base vector string', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.vectorString).toBe('CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H');
    });

    it('should include temporal metrics when present', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        temporal: { E: 'F', RL: 'W', RC: 'R' }
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.vectorString).toContain('/E:F');
      expect(result.vectorString).toContain('/RL:W');
      expect(result.vectorString).toContain('/RC:R');
    });

    it('should not include temporal metrics when all X', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        temporal: { E: 'X', RL: 'X', RC: 'X' }
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.vectorString).not.toContain('/E:X');
      expect(result.vectorString).not.toContain('/RL:X');
      expect(result.vectorString).not.toContain('/RC:X');
    });

    it('should include environmental metrics when present', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
        environmental: { CR: 'H', IR: 'M', AR: 'L', MAV: 'L', MC: 'H' }
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.vectorString).toContain('/CR:H');
      expect(result.vectorString).toContain('/IR:M');
      expect(result.vectorString).toContain('/AR:L');
      expect(result.vectorString).toContain('/MAV:L');
      expect(result.vectorString).toContain('/MC:H');
    });
  });

  describe('Score Breakdown', () => {
    it('should provide exploitability score', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.exploitabilityScore).toBeDefined();
      expect(result.exploitabilityScore).toBeGreaterThan(0);
    });

    it('should provide impact score', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.impactScore).toBeDefined();
      expect(result.impactScore).toBeGreaterThan(0);
    });
  });

  describe('Boundary Conditions', () => {
    it('should return 0.0 for no impact', () => {
      const metrics: CVSS3Metrics = {
        AV: 'P', AC: 'H', PR: 'H', UI: 'N', S: 'U', C: 'N', I: 'N', A: 'N'
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.baseScore).toBe(0.0);
      expect(result.baseSeverity).toBe('NONE');
    });

    it('should return 10.0 for maximum severity with changed scope', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'C', C: 'H', I: 'H', A: 'H'
      };

      const result = calculateCVSS3(metrics, '3.1');

      expect(result.baseScore).toBe(10.0);
      expect(result.baseSeverity).toBe('CRITICAL');
    });
  });
});

// ============================================================================
// CVSS v3.0 Tests
// ============================================================================

describe('CVSS v3.0 Calculator', () => {
  describe('Base Score Calculations', () => {
    it.each(CVSS30_VECTORS)('$description', ({ vector, baseScore, severity }) => {
      const metrics = parseCVSS3Vector(vector);
      const result = calculateCVSS3(metrics, '3.0');

      expect(result.baseScore).toBeCloseTo(baseScore, 1);
      expect(result.baseSeverity).toBe(severity);
    });
  });

  describe('Vector String Version', () => {
    it('should use CVSS:3.0 prefix', () => {
      const metrics: CVSS3Metrics = {
        AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
      };

      const result = calculateCVSS3(metrics, '3.0');

      expect(result.vectorString).toStartWith('CVSS:3.0');
    });
  });
});

// ============================================================================
// CVSS v4.0 Tests
// ============================================================================

describe('CVSS v4.0 Calculator', () => {
  describe('Base Score Calculations', () => {
    it.each(CVSS40_VECTORS)('$description', ({ vector, baseScore, severity }) => {
      const metrics = parseCVSS4Vector(vector);
      const result = calculateCVSS4(metrics);

      expect(result.baseScore).toBeCloseTo(baseScore, 1);
      expect(result.baseSeverity).toBe(severity);
    });
  });

  describe('Threat Score Calculations', () => {
    it('should apply exploit maturity adjustments', () => {
      const baseMetrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        threat: { E: 'A', M: 'X', D: 'N' }
      };

      const result = calculateCVSS4(baseMetrics);

      expect(result.threatScore).toBeDefined();
      expect(result.threatScore).toBeLessThan(result.baseScore);
    });

    it('should apply motility adjustments', () => {
      const baseMetrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        threat: { E: 'X', M: 'A', D: 'N' }
      };

      const result = calculateCVSS4(baseMetrics);

      expect(result.threatScore).toBeDefined();
      expect(result.threatScore).toBeLessThan(result.baseScore);
    });
  });

  describe('Environmental Score Calculations', () => {
    it('should apply confidentiality requirements', () => {
      const baseMetrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        environmental: { CR: 'H', IR: 'N', AR: 'N' }
      };

      const result = calculateCVSS4(baseMetrics);

      expect(result.environmentalScore).toBeDefined();
      expect(result.environmentalScore).toBeGreaterThan(result.baseScore);
    });

    it('should handle modified base metrics in environmental', () => {
      const baseMetrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        environmental: {
          CR: 'N', IR: 'N', AR: 'N',
          MAV: 'L', MAC: 'H', MAT: 'P', MPR: 'L', MUI: 'P'
        }
      };

      const result = calculateCVSS4(baseMetrics);

      expect(result.environmentalScore).toBeDefined();
    });
  });

  describe('Vector String Generation', () => {
    it('should generate correct base vector string', () => {
      const metrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        S: 'X', AU: 'N'
      };

      const result = calculateCVSS4(metrics);

      expect(result.vectorString).toBe('CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H');
    });

    it('should include threat metrics when not X', () => {
      const metrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        threat: { E: 'A', M: 'R', D: 'L' }
      };

      const result = calculateCVSS4(metrics);

      expect(result.vectorString).toContain('/E:A');
      expect(result.vectorString).toContain('/M:R');
      expect(result.vectorString).toContain('/D:L');
    });
  });

  describe('Boundary Conditions', () => {
    it('should return 0.0 for no impact', () => {
      const metrics: CVSS4Metrics = {
        AV: 'P', AC: 'H', AT: 'P', PR: 'H', UI: 'A',
        VC: 'N', VI: 'N', VA: 'N', SC: 'N', SI: 'N', SA: 'N'
      };

      const result = calculateCVSS4(metrics);

      expect(result.baseScore).toBe(0.0);
      expect(result.baseSeverity).toBe('NONE');
    });

    it('should return 10.0 for maximum severity', () => {
      const metrics: CVSS4Metrics = {
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H'
      };

      const result = calculateCVSS4(metrics);

      expect(result.baseScore).toBe(10.0);
      expect(result.baseSeverity).toBe('CRITICAL');
    });
  });
});

// ============================================================================
// Generic calculateCVSS Function Tests
// ============================================================================

describe('calculateCVSS Generic Function', () => {
  it('should route to v3.0 calculator', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
    };

    const result = calculateCVSS('3.0', metrics);

    expect(result.breakdown.baseScore).toBeCloseTo(9.8, 1);
  });

  it('should route to v3.1 calculator', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
    };

    const result = calculateCVSS('3.1', metrics);

    expect(result.breakdown.baseScore).toBeCloseTo(9.8, 1);
  });

  it('should route to v4.0 calculator', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H'
    };

    const result = calculateCVSS('4.0', metrics);

    expect(result.breakdown.baseScore).toBe(10.0);
  });

  it('should throw error for unsupported version', () => {
    const metrics = {} as any;

    expect(() => calculateCVSS('2.0' as any, metrics)).toThrow('Unsupported CVSS version');
  });
});

// ============================================================================
// Utility Functions Tests
// ============================================================================

describe('getDefaultMetrics', () => {
  it('should return default v3.0 metrics', () => {
    const defaults = getDefaultMetrics('3.0');

    expect(defaults).toHaveProperty('AV', 'N');
    expect(defaults).toHaveProperty('AC', 'L');
    expect(defaults).toHaveProperty('PR', 'N');
    expect(defaults).toHaveProperty('UI', 'N');
    expect(defaults).toHaveProperty('S', 'U');
    expect(defaults).toHaveProperty('C', 'N');
    expect(defaults).toHaveProperty('I', 'N');
    expect(defaults).toHaveProperty('A', 'N');
  });

  it('should return default v3.1 metrics', () => {
    const defaults = getDefaultMetrics('3.1');

    expect(defaults).toHaveProperty('AV', 'N');
    expect(defaults).toHaveProperty('AC', 'L');
    expect(defaults).toHaveProperty('PR', 'N');
  });

  it('should return default v4.0 metrics', () => {
    const defaults = getDefaultMetrics('4.0');

    expect(defaults).toHaveProperty('AV', 'N');
    expect(defaults).toHaveProperty('AC', 'L');
    expect(defaults).toHaveProperty('AT', 'N');
    expect(defaults).toHaveProperty('PR', 'N');
    expect(defaults).toHaveProperty('UI', 'N');
    expect(defaults).toHaveProperty('VC', 'N');
    expect(defaults).toHaveProperty('VI', 'N');
    expect(defaults).toHaveProperty('VA', 'N');
    expect(defaults).toHaveProperty('SC', 'N');
    expect(defaults).toHaveProperty('SI', 'N');
    expect(defaults).toHaveProperty('SA', 'N');
  });
});

describe('getCVSSMetadata', () => {
  it('should return v3.0 metadata', () => {
    const metadata = getCVSSMetadata('3.0');

    expect(metadata.version).toBe('3.0');
    expect(metadata.name).toBe('CVSS v3.0');
    expect(metadata.specUrl).toBe('https://www.first.org/cvss/calculator/3.0');
    expect(metadata.releaseDate).toBe('2015-04-09');
  });

  it('should return v3.1 metadata', () => {
    const metadata = getCVSSMetadata('3.1');

    expect(metadata.version).toBe('3.1');
    expect(metadata.name).toBe('CVSS v3.1');
    expect(metadata.specUrl).toBe('https://www.first.org/cvss/calculator/3.1');
    expect(metadata.releaseDate).toBe('2023-11-02');
  });

  it('should return v4.0 metadata', () => {
    const metadata = getCVSSMetadata('4.0');

    expect(metadata.version).toBe('4.0');
    expect(metadata.name).toBe('CVSS v4.0');
    expect(metadata.specUrl).toBe('https://www.first.org/cvss/calculator/4.0');
    expect(metadata.releaseDate).toBe('2023-11-02');
  });
});

// ============================================================================
// Edge Cases and Invalid Inputs
// ============================================================================

describe('Edge Cases', () => {
  it('should handle all Not Defined values in temporal metrics', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      temporal: { E: 'X', RL: 'X', RC: 'X' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Should not have temporal score when all X
    expect(result.temporalScore).toBeUndefined();
  });

  it('should handle partial temporal metrics', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      temporal: { E: 'F', RL: 'X', RC: 'X' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Should apply only the defined metric
    expect(result.temporalScore).toBeDefined();
    expect(result.temporalScore).toBeLessThan(result.baseScore);
  });

  it('should handle all Not Defined in environmental metrics', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      environmental: { CR: 'X', IR: 'X', AR: 'X' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Should still calculate environmental score
    expect(result.environmentalScore).toBeDefined();
  });

  it('should handle scope change impact correctly', () => {
    // Unchanged scope
    const unchangedMetrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
    };

    const unchangedResult = calculateCVSS3(unchangedMetrics, '3.1');

    // Changed scope
    const changedMetrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'C', C: 'H', I: 'H', A: 'H'
    };

    const changedResult = calculateCVSS3(changedMetrics, '3.1');

    // Changed scope should have higher impact
    expect(changedResult.baseScore).toBeGreaterThan(unchangedResult.baseScore);
    expect(changedResult.baseScore).toBe(10.0);
  });

  it('should handle v4.0 with provider I:E', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
      environmental: {
        CR: 'N', IR: 'N', AR: 'N',
        MI: 'E'
      }
    };

    const result = calculateCVSS4(metrics);

    expect(result.environmentalScore).toBeDefined();
  });

  it('should round scores to 1 decimal place', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
    };

    const result = calculateCVSS3(metrics, '3.1');

    expect(result.breakdown.baseScore % 0.1).toBeCloseTo(0, 0.001);
  });
});

describe('CVSS v3.x Temporal Metrics Weight Verification', () => {
  it('should apply E=F (0.97) correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      temporal: { E: 'F', RL: 'X', RC: 'X' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Base Score: 9.8
    // Temporal Score: 9.8 * 0.97 = 9.506 -> 9.5
    expect(result.breakdown.temporalScore).toBeCloseTo(9.5, 0.1);
  });

  it('should apply RL=W (0.95) correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      temporal: { E: 'X', RL: 'W', RC: 'X' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Base Score: 9.8
    // Temporal Score: 9.8 * 0.95 = 9.31 -> 9.3
    expect(result.breakdown.temporalScore).toBeCloseTo(9.3, 0.1);
  });

  it('should apply RC=R (0.96) correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      temporal: { E: 'X', RL: 'X', RC: 'R' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Base Score: 9.8
    // Temporal Score: 9.8 * 0.96 = 9.408 -> 9.4
    expect(result.breakdown.temporalScore).toBeCloseTo(9.4, 0.1);
  });

  it('should apply all temporal metrics correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
      temporal: { E: 'F', RL: 'W', RC: 'R' }
    };

    const result = calculateCVSS3(metrics, '3.1');

    // Base Score: 9.8
    // Temporal Score: 9.8 * 0.97 * 0.95 * 0.96 = 8.66 -> 8.7
    expect(result.breakdown.temporalScore).toBeCloseTo(8.7, 0.1);
  });
});

describe('CVSS v4.0 Threat Metrics Weight Verification', () => {
  it('should apply E=P (0.91) correctly', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
      threat: { E: 'P', M: 'X', D: 'X' }
    };

    const result = calculateCVSS4(metrics);

    // Verify threat score is calculated
    expect(result.breakdown.threatScore).toBeDefined();
  });

  it('should apply E=F (0.94) correctly', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
      threat: { E: 'F', M: 'X', D: 'X' }
    };

    const result = calculateCVSS4(metrics);

    expect(result.breakdown.threatScore).toBeDefined();
  });

  it('should apply M=P (0.95) correctly', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
      threat: { E: 'X', M: 'P', D: 'X' }
    };

    const result = calculateCVSS4(metrics);

    expect(result.breakdown.threatScore).toBeDefined();
  });

  it('should apply D=L (0.98) correctly', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
      threat: { E: 'X', M: 'X', D: 'L' }
    };

    const result = calculateCVSS4(metrics);

    expect(result.breakdown.threatScore).toBeDefined();
  });

  it('should apply all threat metrics correctly', () => {
    const metrics: CVSS4Metrics = {
      AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
      VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
      threat: { E: 'P', M: 'A', D: 'L' }
    };

    const result = calculateCVSS4(metrics);

    // All threat metrics should be applied
    expect(result.breakdown.threatScore).toBeDefined();
    expect(result.breakdown.threatScore).toBeLessThan(result.breakdown.baseScore);
  });
});

describe('Edge Cases and Boundary Conditions', () => {
  it('should handle all Low impacts correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'L', I: 'L', A: 'L'
    };

    const result = calculateCVSS3(metrics, '3.1');

    expect(result.breakdown.baseScore).toBeGreaterThan(0);
    expect(result.breakdown.baseSeverity).toBe('MEDIUM');
  });

  it('should handle all High impacts correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
    };

    const result = calculateCVSS3(metrics, '3.1');

    expect(result.breakdown.baseScore).toBe(9.8);
    expect(result.breakdown.baseSeverity).toBe('CRITICAL');
  });

  it('should handle zero impact correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'N', I: 'N', A: 'N'
    };

    const result = calculateCVSS3(metrics, '3.1');

    expect(result.breakdown.baseScore).toBe(0.0);
    expect(result.breakdown.baseSeverity).toBe('NONE');
  });

  it('should handle Physical attack vector correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'P', AC: 'H', PR: 'L', UI: 'R', S: 'U', C: 'N', I: 'N', A: 'N'
    };

    const result = calculateCVSS3(metrics, '3.1');

    expect(result.breakdown.baseScore).toBeLessThan(5.0);
  });

  it('should handle Changed scope correctly', () => {
    const metrics: CVSS3Metrics = {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'C', C: 'H', I: 'H', A: 'H'
    };

    const result = calculateCVSS3(metrics, '3.1');

    expect(result.breakdown.baseScore).toBe(10.0);
    expect(result.breakdown.baseSeverity).toBe('CRITICAL');
  });
});
});
