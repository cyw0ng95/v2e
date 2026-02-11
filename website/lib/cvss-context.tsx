'use client';

/**
 * CVSS Calculator Context Provider
 * Manages CVSS calculator state across the application
 */

import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import type {
  CVSSVersion, CVSSCalculatorState, CVSS3Metrics, CVSS4Metrics,
  CVSS3ScoreBreakdown, CVSS4ScoreBreakdown, CVSSSeverity
} from './types';
import {
  calculateCVSS,
  getDefaultMetrics,
  getCVSSMetadata
} from './cvss-calculator';

// ============================================================================
// Context Types
// ============================================================================

interface CVSSContextType {
  state: CVSSCalculatorState;
  version: CVSSVersion;
  updateMetric: <K extends keyof CVSS3Metrics | keyof CVSS4Metrics>(
    metric: K,
    value: string
  ) => void;
  setVersion: (version: CVSSVersion) => void;
  resetMetrics: () => void;
  toggleTemporal: () => void;
  toggleEnvironmental: () => void;
  exportData: (format: 'json' | 'csv' | 'url') => Promise<void>;
  importFromVector: (vectorString: string) => boolean;
}

const CVSSContext = createContext<CVSSContextType | null>(null);

// ============================================================================
// Provider Props
// ============================================================================

interface CVSSProviderProps {
  children: ReactNode;
  initialVersion?: CVSSVersion;
  enableTemporal?: boolean;
  enableEnvironmental?: boolean;
}

// ============================================================================
// Provider Component
// ============================================================================

export function CVSSProvider({
  children,
  initialVersion = '3.1',
  enableTemporal = true,
  enableEnvironmental = true
}: CVSSProviderProps) {
  const [version, setVersion] = useState<CVSSVersion>(initialVersion);

  const [metrics, setMetrics] = useState<CVSS3Metrics | CVSS4Metrics>(() => {
    const defaults = getDefaultMetrics(initialVersion);
    if (enableTemporal && (initialVersion === '3.0' || initialVersion === '3.1')) {
      const temporal = { E: 'X', RL: 'X', RC: 'X' };
      return { ...defaults, temporal };
    }
    if (enableEnvironmental) {
      if (initialVersion === '3.0' || initialVersion === '3.1') {
        const environmental = { CR: 'N', IR: 'N', AR: 'N' };
        return {
          ...defaults,
          temporal: enableTemporal ? { E: 'X', RL: 'X', RC: 'X' } : undefined,
          environmental
        };
      }
      const threat = { E: 'X', M: 'X', D: 'N' };
      const environmental = { CR: 'N', IR: 'N', AR: 'N' };
      return { ...defaults, threat, environmental };
    }
    return defaults;
  });

  const [scores, setScores] = useState<CVSS3ScoreBreakdown | CVSS4ScoreBreakdown | null>(null);

  const [showTemporal, setShowTemporal] = useState(enableTemporal);
  const [showEnvironmental, setShowEnvironmental] = useState(enableEnvironmental);
  const [showDescriptions, setShowDescriptions] = useState(true);
  const [viewMode, setViewMode] = useState<'compact' | 'full'>('compact');

  const calculateScores = useCallback(() => {
    try {
      const result = calculateCVSS(version, metrics);
      setScores(result.breakdown as CVSS3ScoreBreakdown | CVSS4ScoreBreakdown);
    } catch (error) {
      console.error('CVSS calculation error:', error);
    }
  }, [version, metrics]);

  React.useEffect(() => {
    calculateScores();
  }, [calculateScores]);

  const updateMetric = useCallback(<K extends keyof CVSS3Metrics | keyof CVSS4Metrics>(
    metric: K,
    value: string
  ) => {
    setMetrics((prev: any) => ({ ...prev, [metric]: value }));
  }, []);

  const handleSetVersion = useCallback((newVersion: CVSSVersion) => {
    setVersion(newVersion);
    const defaults = getDefaultMetrics(newVersion);
    if (showTemporal && (newVersion === '3.0' || newVersion === '3.1')) {
      setMetrics({ ...defaults, temporal: { E: 'X', RL: 'X', RC: 'X' } });
    } else if (showEnvironmental) {
      if (newVersion === '3.0' || newVersion === '3.1') {
        const env = { CR: 'N', IR: 'N', AR: 'N' };
        setMetrics({
          ...defaults,
          temporal: showTemporal ? { E: 'X', RL: 'X', RC: 'X' } : undefined,
          environmental: env
        });
      } else {
        const threat = { E: 'X', M: 'X', D: 'N' };
        const env = { CR: 'N', IR: 'N', AR: 'N' };
        setMetrics({ ...defaults, threat, environmental: env });
      }
    } else {
      setMetrics(defaults);
    }
    setScores(null);
  }, [showTemporal, showEnvironmental]);

  const resetMetrics = useCallback(() => {
    handleSetVersion(version);
  }, [version, handleSetVersion]);

  const toggleTemporal = useCallback(() => {
    setShowEnvironmental(prev => {
      const newValue = !prev;
      if (newValue && (version === '3.0' || version === '3.1')) {
        setMetrics((prevMetrics: any) => ({
          ...prevMetrics,
          temporal: { E: 'X', RL: 'X', RC: 'X' }
        }));
      } else if (version === '3.0' || version === '3.1') {
        setMetrics((prevMetrics: any) => {
          const { temporal, ...rest } = prevMetrics;
          return rest;
        });
      }
      return newValue;
    });
  }, [version]);

  const toggleEnvironmental = useCallback(() => {
    setShowEnvironmental(prev => {
      const newValue = !prev;
      if (newValue) {
        if (version === '3.0' || version === '3.1') {
          setMetrics((prevMetrics: any) => ({
            ...prevMetrics,
            environmental: { CR: 'N', IR: 'N', AR: 'N' }
          }));
        } else {
          setMetrics((prevMetrics: any) => ({
            ...prevMetrics,
            environmental: { CR: 'N', IR: 'N', AR: 'N' }
          }));
        }
      } else {
        const { environmental, ...rest } = prevMetrics as any;
        setMetrics(rest);
      }
      return newValue;
    });
  }, [version]);

  const exportData = useCallback(async (format: 'json' | 'csv' | 'url') => {
    if (!scores) return;

    const metadata = getCVSSMetadata(version);
    const data = {
      version,
      vectorString: scores.vectorString,
      baseScore: scores.baseScore,
      severity: scores.finalSeverity,
      metrics,
      scoreBreakdown: scores,
      exportedAt: new Date().toISOString()
    };

    if (format === 'json') {
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `cvss-${version}-${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } else if (format === 'csv') {
      const metricsStr = JSON.stringify(metrics);
      const csv = [
        version,
        scores.vectorString,
        scores.baseScore.toFixed(1),
        scores.finalSeverity,
        `"${metricsStr}"`,
        data.exportedAt
      ].join(',');
      const blob = new Blob([csv], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `cvss-${version}-${Date.now()}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } else if (format === 'url') {
      const params = new URLSearchParams();
      params.set('v', scores.vectorString);
      const shareUrl = `${window.location.origin}/cvss/${version}?${params.toString()}`;
      await navigator.clipboard.writeText(shareUrl);
      alert('CVSS URL copied to clipboard!');
    }
  }, [version, metrics, scores]);

  const importFromVector = useCallback((vectorString: string): boolean => {
    try {
      if (vectorString.startsWith('CVSS:3.')) {
        const parts = vectorString.split('/');
        const newMetrics: any = {};
        for (const part of parts) {
          const [key, value] = part.split(':');
          const allowedKeys = [
            'AV', 'AC', 'PR', 'UI', 'S', 'C', 'I', 'A',
            'E', 'RL', 'RC', 'CR', 'IR', 'AR',
            'MAV', 'MAC', 'MPR', 'MUI', 'MS', 'MC', 'MI', 'MA'
          ];
          if (value && allowedKeys.includes(key)) {
            newMetrics[key] = value;
          }
        }

        const hasRequired = newMetrics.AV && newMetrics.AC &&
            newMetrics.PR && newMetrics.UI &&
            newMetrics.S && newMetrics.C &&
            newMetrics.I && newMetrics.A;

        if (hasRequired) {
          setMetrics(newMetrics);
          return true;
        }
      } else if (vectorString.startsWith('CVSS:4.0')) {
        const parts = vectorString.split('/');
        const newMetrics: any = {};
        for (const part of parts) {
          const [key, value] = part.split(':');
          const allowedKeys = [
            'AV', 'AC', 'AT', 'PR', 'UI',
            'VC', 'VI', 'VA', 'SC', 'SI', 'SA', 'S', 'AU',
            'E', 'M', 'D', 'CR', 'IR', 'AR',
            'MAV', 'MAC', 'MAT', 'MPR', 'MUI',
            'MVC', 'MVI', 'MVA', 'MSC', 'MSI', 'MSA', 'MS', 'MAU', 'MI'
          ];
          if (value && allowedKeys.includes(key)) {
            newMetrics[key] = value;
          }
        }

        const hasRequired = newMetrics.AV && newMetrics.AC &&
            newMetrics.AT && newMetrics.PR &&
            newMetrics.UI && newMetrics.VC &&
            newMetrics.VI && newMetrics.VA &&
            newMetrics.SC && newMetrics.SI &&
            newMetrics.SA;

        if (hasRequired) {
          setMetrics(newMetrics);
          return true;
        }
      }

      console.error('Invalid CVSS vector string:', vectorString);
      return false;
    } catch (error) {
      console.error('Error parsing CVSS vector:', error);
      return false;
    }
  }, [version]);

  const contextValue: CVSSContextType = {
    state: {
      version,
      metrics,
      scores,
      vectorString: scores?.vectorString ?? '',
      showTemporal,
      showEnvironmental,
      showDescriptions,
      viewMode
    },
    version,
    updateMetric,
    setVersion: handleSetVersion,
    resetMetrics,
    toggleTemporal,
    toggleEnvironmental,
    exportData,
    importFromVector
  };

  return (
    <CVSSContext.Provider value={contextValue}>
      {children}
    </CVSSContext.Provider>
  );
}

/**
 * Use CVSS calculator context
 */
export function useCVSS() {
  const context = useContext(CVSSContext);
  if (!context) {
    throw new Error('useCVSS must be used within CVSSProvider');
  }
  return context;
}
