'use client';

/**
 * CVSS Screen Reader Components
 * Enhanced ARIA announcements and screen reader support for CVSS calculator
 */

import React, { useEffect, useRef } from 'react';
import { cn } from '@/lib/utils';
import type { CVSSScoreBreakdown, CVSSSeverity } from '@/lib/types';

// ============================================================================
// Types
// ============================================================================

export interface ScreenReaderAnnouncerProps {
  message: string;
  priority?: 'polite' | 'assertive';
  timeout?: number;
}

export interface ScoreAnnouncerProps {
  scores: CVSSScoreBreakdown;
  previousScore?: number;
  previousSeverity?: CVSSSeverity;
}

export interface MetricChangeAnnouncerProps {
  metricName: string;
  metricValue: string;
  metricLabel: string;
  newScore?: number;
  newSeverity?: CVSSSeverity;
}

export interface VectorStringAnnouncerProps {
  vectorString: string;
  isValid: boolean;
  error?: string;
}

// ============================================================================
// Screen Reader Announcer
// ============================================================================

export function ScreenReaderAnnouncer({
  message,
  priority = 'polite',
  timeout = 0,
}: ScreenReaderAnnouncerProps) {
  const [announcement, setAnnouncement] = React.useState('');

  useEffect(() => {
    const timer = setTimeout(() => {
      setAnnouncement(message);
    }, timeout);

    return () => clearTimeout(timer);
  }, [message, timeout]);

  return (
    <div
      role="status"
      aria-live={priority}
      aria-atomic="true"
      className="sr-only"
    >
      {announcement}
    </div>
  );
}

// ============================================================================
// Score Change Announcer
// ============================================================================

export function ScoreAnnouncer({
  scores,
  previousScore,
  previousSeverity,
}: ScoreAnnouncerProps) {
  const announcedRef = useRef<string>('');

  const getMessage = (): string => {
    const messages: string[] = [];

    // Score change
    if (previousScore !== undefined && scores.finalScore !== previousScore) {
      const difference = scores.finalScore - previousScore;
      if (difference > 0) {
        messages.push(`Score increased by ${difference.toFixed(1)} points`);
      } else if (difference < 0) {
        messages.push(`Score decreased by ${Math.abs(difference).toFixed(1)} points`);
      }
      messages.push(`New score is ${scores.finalScore.toFixed(1)}`);
    }

    // Severity change
    if (previousSeverity !== undefined && scores.finalSeverity !== previousSeverity) {
      messages.push(`Severity changed from ${previousSeverity} to ${scores.finalSeverity}`);
    }

    // Full announcement for new scores
    if (previousScore === undefined) {
      messages.push(
        `CVSS score calculated: ${scores.finalScore.toFixed(1)}, Severity: ${scores.finalSeverity}`
      );
    }

    return messages.join('. ');
  };

  const message = getMessage();
  const messageId = `${scores.finalScore}-${scores.finalSeverity}`;

  // Only announce if the state actually changed
  if (messageId === announcedRef.current) {
    return null;
  }

  useEffect(() => {
    announcedRef.current = messageId;
  }, [messageId]);

  if (!message) {
    return null;
  }

  return (
    <ScreenReaderAnnouncer message={message} priority="polite" />
  );
}

// ============================================================================
// Metric Change Announcer
// ============================================================================

export function MetricChangeAnnouncer({
  metricName,
  metricValue,
  metricLabel,
  newScore,
  newSeverity,
}: MetricChangeAnnouncerProps) {
  const getMessage = (): string => {
    const parts = [
      `${metricName} changed to ${metricLabel} (${metricValue})`,
    ];

    if (newScore !== undefined) {
      parts.push(`New score is ${newScore.toFixed(1)}`);
    }

    if (newSeverity !== undefined) {
      parts.push(`Severity is ${newSeverity}`);
    }

    return parts.join('. ');
  };

  return (
    <ScreenReaderAnnouncer
      message={getMessage()}
      priority="polite"
      timeout={100}
    />
  );
}

// ============================================================================
// Vector String Announcer
// ============================================================================

export function VectorStringAnnouncer({
  vectorString,
  isValid,
  error,
}: VectorStringAnnouncerProps) {
  const getMessage = (): string => {
    if (!vectorString) {
      return 'Vector string is empty';
    }

    if (!isValid) {
      return error || 'Invalid CVSS vector string';
    }

    // Parse and announce the vector
    const parts = vectorString.split('/');
    const version = parts[0];
    const metrics = parts.slice(1).join(', ').replace(/:/g, ' set to ');

    return `${version}. ${metrics}`;
  };

  return (
    <ScreenReaderAnnouncer
      message={getMessage()}
      priority="assertive"
      timeout={200}
    />
  );
}

// ============================================================================
// Live Region Component
// ============================================================================

export interface LiveRegionProps {
  children: React.ReactNode;
  ariaLabel?: string;
  role?: 'status' | 'alert' | 'log';
}

export function LiveRegion({
  children,
  ariaLabel,
  role = 'status',
}: LiveRegionProps) {
  return (
    <div
      role={role}
      aria-live="polite"
      aria-atomic="true"
      aria-label={ariaLabel}
      className="sr-only"
    >
      {children}
    </div>
  );
}

// ============================================================================
// Visually Hidden Component
// ============================================================================

export interface VisuallyHiddenProps {
  children: React.ReactNode;
  focusable?: boolean;
}

export function VisuallyHidden({
  children,
  focusable = false,
}: VisuallyHiddenProps) {
  return (
    <span
      className={cn(
        'sr-only',
        focusable && 'focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-blue-600 focus:text-white focus:rounded-lg'
      )}
    >
      {children}
    </span>
  );
}

// ============================================================================
// Screen Reader Only Text Component
// ============================================================================

export interface SrOnlyProps {
  children: React.ReactNode;
  as?: 'span' | 'div' | 'p';
}

export function SrOnly({ children, as = 'span' }: SrOnlyProps) {
  const Tag = as;

  return (
    <Tag className="sr-only">
      {children}
    </Tag>
  );
}

// ============================================================================
// Focus Trap Component
// ============================================================================

export interface FocusTrapProps {
  isActive: boolean;
  onEscape?: () => void;
  children: React.ReactNode;
}

export function FocusTrap({ isActive, onEscape, children }: FocusTrapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const previousActiveElement = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isActive || !containerRef.current) return;

    // Store the previously focused element
    previousActiveElement.current = document.activeElement as HTMLElement;

    // Get all focusable elements within the container
    const focusableElements = containerRef.current.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
    );

    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    // Focus the first element
    firstElement?.focus();

    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      if (e.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstElement) {
          e.preventDefault();
          lastElement?.focus();
        }
      } else {
        // Tab
        if (document.activeElement === lastElement) {
          e.preventDefault();
          firstElement?.focus();
        }
      }
    };

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onEscape?.();
      }
    };

    document.addEventListener('keydown', handleTab);
    document.addEventListener('keydown', handleEscape);

    return () => {
      document.removeEventListener('keydown', handleTab);
      document.removeEventListener('keydown', handleEscape);
      // Restore focus when trap is deactivated
      previousActiveElement.current?.focus();
    };
  }, [isActive, onEscape]);

  return <div ref={containerRef}>{children}</div>;
}

// ============================================================================
// Accessible Description Component
// ============================================================================

export interface AccessibleDescriptionProps {
  id: string;
  children: React.ReactNode;
}

export function AccessibleDescription({ id, children }: AccessibleDescriptionProps) {
  return (
    <span id={id} className="sr-only">
      {children}
    </span>
  );
}

// ============================================================================
// ARIA Live Region Manager Hook
// ============================================================================

export function useAnnouncer() {
  const [announcement, setAnnouncement] = React.useState('');
  const announceRef = useRef('');

  const announce = React.useCallback((message: string, priority: 'polite' | 'assertive' = 'polite') => {
    if (message !== announceRef.current) {
      announceRef.current = message;
      setAnnouncement(message);
    }
  }, []);

  return {
    announce,
    Announcer: ({ message = announcement, priority = 'polite' }: {
      message?: string;
      priority?: 'polite' | 'assertive';
    }) => (
      <div
        role="status"
        aria-live={priority}
        aria-atomic="true"
        className="sr-only"
      >
        {message}
      </div>
    ),
  };
}

// ============================================================================
// Screen Reader Score Table Component
// ============================================================================

export interface ScreenReaderScoreTableProps {
  scores: CVSSScoreBreakdown;
}

export function ScreenReaderScoreTable({ scores }: ScreenReaderScoreTableProps) {
  return (
    <div className="sr-only" role="region" aria-label="Detailed score breakdown">
      <table>
        <thead>
          <tr>
            <th>Score Type</th>
            <th>Value</th>
            <th>Severity</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Base Score</td>
            <td>{scores.baseScore.toFixed(1)}</td>
            <td>{scores.baseSeverity}</td>
          </tr>
          {(scores as any).temporalScore !== undefined && (
            <tr>
              <td>Temporal Score</td>
              <td>{(scores as any).temporalScore.toFixed(1)}</td>
              <td>{(scores as any).temporalSeverity}</td>
            </tr>
          )}
          {(scores as any).environmentalScore !== undefined && (
            <tr>
              <td>Environmental Score</td>
              <td>{(scores as any).environmentalScore.toFixed(1)}</td>
              <td>{(scores as any).environmentalSeverity}</td>
            </tr>
          )}
          {(scores as any).threatScore !== undefined && (
            <tr>
              <td>Threat Score</td>
              <td>{(scores as any).threatScore.toFixed(1)}</td>
              <td>-</td>
            </tr>
          )}
          <tr>
            <td><strong>Final Score</strong></td>
            <td><strong>{scores.finalScore.toFixed(1)}</strong></td>
            <td><strong>{scores.finalSeverity}</strong></td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

// ============================================================================
// Context Announcer for Calculator State
// ============================================================================

export interface CalculatorContextAnnouncerProps {
  version: string;
  metricCount: number;
  hasTemporalMetrics: boolean;
  hasEnvironmentalMetrics: boolean;
}

export function CalculatorContextAnnouncer({
  version,
  metricCount,
  hasTemporalMetrics,
  hasEnvironmentalMetrics,
}: CalculatorContextAnnouncerProps) {
  const getMessage = (): string => {
    const parts = [
      `Using CVSS version ${version}`,
      `${metricCount} base metrics available`,
    ];

    if (hasTemporalMetrics) {
      parts.push('Temporal metrics are enabled');
    }

    if (hasEnvironmentalMetrics) {
      parts.push('Environmental metrics are enabled');
    }

    return parts.join('. ');
  };

  return (
    <ScreenReaderAnnouncer
      message={getMessage()}
      priority="polite"
    />
  );
}

// ============================================================================
// Helper to generate accessible descriptions
// ============================================================================

export function generateMetricAriaDescription(
  metricName: string,
  metricValue: string,
  options: Array<{ value: string; label: string; description?: string }>
): string {
  const selectedOption = options.find(o => o.value === metricValue);

  if (!selectedOption) {
    return `${metricName} set to ${metricValue}`;
  }

  let description = `${metricName} set to ${selectedOption.label}`;

  if (selectedOption.description) {
    description += `. ${selectedOption.description}`;
  }

  return description;
}

// ============================================================================
// Export a comprehensive screen reader helper
// ============================================================================

export const ScreenReaderHelpers = {
  Announcer: ScreenReaderAnnouncer,
  ScoreAnnouncer,
  MetricChangeAnnouncer,
  VectorStringAnnouncer,
  LiveRegion,
  VisuallyHidden,
  SrOnly,
  FocusTrap,
  AccessibleDescription,
  useAnnouncer,
  ScreenReaderScoreTable,
  CalculatorContextAnnouncer,
  generateMetricAriaDescription,
};
