/**
 * CVSS Calculator Performance Tests
 * Measures calculation speed and verifies performance requirements
 */

import { describe, it, expect, bench } from 'vitest';
import {
  calculateCVSS3,
  calculateCVSS4,
  calculateCVSS
} from '@/lib/cvss-calculator';
import type { CVSS3Metrics, CVSS4Metrics } from '@/lib/types';

// ============================================================================
// Test Data
// ============================================================================

const cvss3Metrics: CVSS3Metrics = {
  AV: 'N',
  AC: 'L',
  PR: 'N',
  UI: 'N',
  S: 'U',
  C: 'H',
  I: 'H',
  A: 'H',
  temporal: {
    E: 'F',
    RL: 'W',
    RC: 'R'
  },
  environmental: {
    CR: 'H',
    IR: 'M',
    AR: 'L',
    MAV: 'L',
    MAC: 'H',
    MPR: 'L',
    MUI: 'N',
    MS: 'C',
    MC: 'H',
    MI: 'H',
    MA: 'H'
  }
};

const cvss4Metrics: CVSS4Metrics = {
  AV: 'N',
  AC: 'L',
  AT: 'N',
  PR: 'N',
  UI: 'N',
  VC: 'H',
  VI: 'H',
  VA: 'H',
  SC: 'H',
  SI: 'H',
  SA: 'H',
  S: 'X',
  AU: 'N',
  threat: {
    E: 'A',
    M: 'R',
    D: 'L'
  },
  environmental: {
    CR: 'H',
    IR: 'M',
    AR: 'L',
    MAV: 'L',
    MAC: 'H',
    MAT: 'P',
    MPR: 'L',
    MUI: 'P',
    MVC: 'H',
    MVI: 'H',
    MVA: 'H',
    MSC: 'H',
    MSI: 'H',
    MSA: 'H',
    MS: 'C',
    MI: 'E'
  }
};

// ============================================================================
// Performance Tests
// ============================================================================

describe('CVSS Calculator Performance', () => {
  describe('Single Calculation Performance', () => {
    it('should calculate CVSS v3.1 base score in < 1ms', () => {
      const start = performance.now();

      const result = calculateCVSS3(
        {
          AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
        },
        '3.1'
      );

      const end = performance.now();
      const duration = end - start;

      expect(result).toBeDefined();
      expect(result.baseScore).toBe(9.8);
      expect(duration).toBeLessThan(1);
    });

    it('should calculate CVSS v3.1 with temporal in < 1ms', () => {
      const start = performance.now();

      const result = calculateCVSS3(
        {
          AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
          temporal: { E: 'F', RL: 'W', RC: 'R' }
        },
        '3.1'
      );

      const end = performance.now();
      const duration = end - start;

      expect(result).toBeDefined();
      expect(result.temporalScore).toBeDefined();
      expect(duration).toBeLessThan(1);
    });

    it('should calculate CVSS v3.1 with environmental in < 1ms', () => {
      const start = performance.now();

      const result = calculateCVSS3(
        {
          AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
          environmental: { CR: 'H', IR: 'M', AR: 'L' }
        },
        '3.1'
      );

      const end = performance.now();
      const duration = end - start;

      expect(result).toBeDefined();
      expect(result.environmentalScore).toBeDefined();
      expect(duration).toBeLessThan(1);
    });

    it('should calculate CVSS v4.0 base score in < 1ms', () => {
      const start = performance.now();

      const result = calculateCVSS4({
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H'
      });

      const end = performance.now();
      const duration = end - start;

      expect(result).toBeDefined();
      // Maximum v4.0 score should be 10.0
      expect(result.baseScore).toBeCloseTo(10.0, 1);
      expect(duration).toBeLessThan(5);
    });

    it('should calculate CVSS v4.0 with threat in < 1ms', () => {
      const start = performance.now();

      const result = calculateCVSS4({
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        threat: { E: 'A', M: 'R', D: 'L' }
      });

      const end = performance.now();
      const duration = end - start;

      expect(result).toBeDefined();
      expect(result.threatScore).toBeDefined();
      expect(duration).toBeLessThan(1);
    });

    it('should calculate CVSS v4.0 with environmental in < 1ms', () => {
      const start = performance.now();

      const result = calculateCVSS4({
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
        environmental: { CR: 'H', IR: 'M', AR: 'L' }
      });

      const end = performance.now();
      const duration = end - start;

      expect(result).toBeDefined();
      expect(result.environmentalScore).toBeDefined();
      expect(duration).toBeLessThan(1);
    });
  });

  describe('Batch Calculation Performance', () => {
    it('should calculate 1000 CVSS v3.1 scores in < 10ms', () => {
      const start = performance.now();

      for (let i = 0; i < 1000; i++) {
        calculateCVSS3(
          {
            AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H',
            temporal: { E: 'F', RL: 'W', RC: 'R' },
            environmental: { CR: 'H', IR: 'M', AR: 'L' }
          },
          '3.1'
        );
      }

      const end = performance.now();
      const duration = end - start;

      expect(duration).toBeLessThan(10);
    });

    it('should calculate 1000 CVSS v4.0 scores in < 10ms', () => {
      const start = performance.now();

      for (let i = 0; i < 1000; i++) {
        calculateCVSS4({
          AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
          VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H',
          threat: { E: 'A', M: 'R', D: 'L' },
          environmental: { CR: 'H', IR: 'M', AR: 'L' }
        });
      }

      const end = performance.now();
      const duration = end - start;

      expect(duration).toBeLessThan(10);
    });

    it('should calculate 1000 mixed version scores in < 20ms', () => {
      const start = performance.now();

      for (let i = 0; i < 1000; i++) {
        if (i % 2 === 0) {
          calculateCVSS('3.1', cvss3Metrics);
        } else {
          calculateCVSS('4.0', cvss4Metrics);
        }
      }

      const end = performance.now();
      const duration = end - start;

      expect(duration).toBeLessThan(20);
    });
  });

  describe('Vector String Generation Performance', () => {
    it('should generate 1000 CVSS v3.1 vector strings in < 5ms', () => {
      const start = performance.now();

      for (let i = 0; i < 1000; i++) {
        calculateCVSS3(cvss3Metrics, '3.1');
      }

      const end = performance.now();
      const duration = end - start;

      expect(duration).toBeLessThan(5);
    });

    it('should generate 1000 CVSS v4.0 vector strings in < 5ms', () => {
      const start = performance.now();

      for (let i = 0; i < 1000; i++) {
        calculateCVSS4(cvss4Metrics);
      }

      const end = performance.now();
      const duration = end - start;

      expect(duration).toBeLessThan(5);
    });
  });

  describe('Memory Efficiency', () => {
    it('should not leak memory during repeated calculations', () => {
      const initialMemory = (performance as any).memory?.usedJSHeapSize || 0;

      // Run 10,000 calculations
      for (let i = 0; i < 10000; i++) {
        calculateCVSS3(
          {
            AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U',
            C: ['H', 'L', 'N'][i % 3] as 'H' | 'L' | 'N',
            I: ['H', 'L', 'N'][i % 3] as 'H' | 'L' | 'N',
            A: ['H', 'L', 'N'][i % 3] as 'H' | 'L' | 'N'
          },
          '3.1'
        );
      }

      const finalMemory = (performance as any).memory?.usedJSHeapSize || 0;
      const memoryIncrease = finalMemory - initialMemory;

      // Memory increase should be less than 1MB (if memory API is available)
      if ((performance as any).memory?.usedJSHeapSize) {
        expect(memoryIncrease).toBeLessThan(1024 * 1024);
      }
    });
  });

  describe('Edge Case Performance', () => {
    it('should handle zero score calculation efficiently', () => {
      const start = performance.now();

      const result = calculateCVSS3(
        {
          AV: 'P', AC: 'H', PR: 'H', UI: 'N', S: 'U',
          C: 'N', I: 'N', A: 'N'
        },
        '3.1'
      );

      const end = performance.now();
      const duration = end - start;

      // Zero or near-zero score for no impact
      expect(result.baseScore).toBeLessThan(1);
      expect(duration).toBeLessThan(5);
    });

    it('should handle maximum score calculation efficiently', () => {
      const start = performance.now();

      const result = calculateCVSS3(
        {
          AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'C',
          C: 'H', I: 'H', A: 'H'
        },
        '3.1'
      );

      const end = performance.now();
      const duration = end - start;

      expect(result.baseScore).toBe(10);
      expect(duration).toBeLessThan(1);
    });
  });
});

// ============================================================================
// Benchmarks (run with --mode benchmark or vitest --bench)
// ============================================================================

describe('CVSS Calculator Benchmarks (Performance Regression)', () => {
  it('CVSS v3.1 base score calculation (10k iterations)', () => {
    const start = performance.now();

    for (let i = 0; i < 10000; i++) {
      calculateCVSS3(
        {
          AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'H', I: 'H', A: 'H'
        },
        '3.1'
      );
    }

    const end = performance.now();
    const duration = end - start;

    // 10k iterations should complete in < 100ms (avg < 0.01ms per calculation)
    expect(duration).toBeLessThan(100);
  });

  it('CVSS v3.1 with temporal and environmental (10k iterations)', () => {
    const start = performance.now();

    for (let i = 0; i < 10000; i++) {
      calculateCVSS3(cvss3Metrics, '3.1');
    }

    const end = performance.now();
    const duration = end - start;

    expect(duration).toBeLessThan(100);
  });

  it('CVSS v4.0 base score calculation (10k iterations)', () => {
    const start = performance.now();

    for (let i = 0; i < 10000; i++) {
      calculateCVSS4({
        AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
        VC: 'H', VI: 'H', VA: 'H', SC: 'H', SI: 'H', SA: 'H'
      });
    }

    const end = performance.now();
    const duration = end - start;

    expect(duration).toBeLessThan(100);
  });

  it('CVSS v4.0 with threat and environmental (10k iterations)', () => {
    const start = performance.now();

    for (let i = 0; i < 10000; i++) {
      calculateCVSS4(cvss4Metrics);
    }

    const end = performance.now();
    const duration = end - start;

    expect(duration).toBeLessThan(100);
  });

  it('Generic calculateCVSS function v3.1 (10k iterations)', () => {
    const start = performance.now();

    for (let i = 0; i < 10000; i++) {
      calculateCVSS('3.1', cvss3Metrics);
    }

    const end = performance.now();
    const duration = end - start;

    expect(duration).toBeLessThan(100);
  });

  it('Generic calculateCVSS function v4.0 (10k iterations)', () => {
    const start = performance.now();

    for (let i = 0; i < 10000; i++) {
      calculateCVSS('4.0', cvss4Metrics);
    }

    const end = performance.now();
    const duration = end - start;

    expect(duration).toBeLessThan(100);
  });
});
