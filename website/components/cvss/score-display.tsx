'use client';

/**
 * CVSS Score Display Component
 * Displays CVSS scores with severity color coding and animations
 */

import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';
import type { CVSSSeverity } from '@/lib/types';

// ============================================================================
// Severity Color Variants
// ============================================================================

const severityVariants = cva(
  'rounded-xl border-2 p-6 transition-all duration-300',
  {
    variants: {
      severity: {
        NONE: 'from-gray-100 to-gray-200 bg-gradient-to-br border-gray-300 text-gray-700',
        LOW: 'from-yellow-50 to-yellow-100 bg-gradient-to-br border-yellow-300 text-yellow-800',
        MEDIUM: 'from-orange-50 to-orange-100 bg-gradient-to-br border-orange-300 text-orange-800',
        HIGH: 'from-red-50 to-red-100 bg-gradient-to-br border-red-300 text-red-800',
        CRITICAL: 'from-purple-50 to-purple-100 bg-gradient-to-br border-purple-400 text-purple-900',
      },
      size: {
        default: 'p-6',
        compact: 'p-4',
        large: 'p-8',
      },
    },
    defaultVariants: {
      severity: 'NONE',
      size: 'default',
    },
  }
);

const scoreSizeVariants = cva(
  'font-bold tracking-tight',
  {
    variants: {
      size: {
        default: 'text-5xl',
        compact: 'text-3xl',
        large: 'text-6xl',
      },
    },
    defaultVariants: {
      size: 'default',
    },
  }
);

const labelSizeVariants = cva(
  'font-medium',
  {
    variants: {
      size: {
        default: 'text-sm',
        compact: 'text-xs',
        large: 'text-base',
      },
    },
    defaultVariants: {
      size: 'default',
    },
  }
);

// ============================================================================
// Types
// ============================================================================

export interface ScoreDisplayProps extends VariantProps<typeof severityVariants> {
  /** Score value (0.0-10.0) */
  score: number;
  /** Severity level */
  severity: CVSSSeverity;
  /** Score label (e.g., "Base Score", "Temporal Score") */
  label?: string;
  /** Additional description text */
  description?: string;
  /** Show border glow effect */
  showGlow?: boolean;
  /** Custom className */
  className?: string;
  /** Children for additional content */
  children?: React.ReactNode;
}

export interface ScoreBreakdownProps {
  /** Base score */
  baseScore: number;
  /** Temporal score (optional) */
  temporalScore?: number;
  /** Environmental score (optional) */
  environmentalScore?: number;
  /** Threat score (CVSS v4.0, optional) */
  threatScore?: number;
  /** Final severity */
  finalSeverity: CVSSSeverity;
  /** Size variant */
  size?: 'compact' | 'default' | 'large';
  /** Show individual score breakdowns */
  showBreakdown?: boolean;
  /** Custom className */
  className?: string;
}

// ============================================================================
// Score Badge Component
// ============================================================================

export interface ScoreBadgeProps {
  severity: CVSSSeverity;
  score: number;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

const badgeSizeVariants = cva(
  'rounded-full font-bold',
  {
    variants: {
      size: {
        sm: 'px-2 py-0.5 text-xs',
        md: 'px-3 py-1 text-sm',
        lg: 'px-4 py-1.5 text-base',
      },
      severity: {
        NONE: 'bg-gray-200 text-gray-700',
        LOW: 'bg-yellow-200 text-yellow-800',
        MEDIUM: 'bg-orange-200 text-orange-800',
        HIGH: 'bg-red-200 text-red-800',
        CRITICAL: 'bg-purple-200 text-purple-900',
      },
    },
    defaultVariants: {
      size: 'md',
      severity: 'NONE',
    },
  }
);

export function ScoreBadge({ severity, score, size = 'md', showLabel = true }: ScoreBadgeProps) {
  const severityLabels: Record<CVSSSeverity, string> = {
    NONE: 'None',
    LOW: 'Low',
    MEDIUM: 'Medium',
    HIGH: 'High',
    CRITICAL: 'Critical',
  };

  return (
    <span
      className={badgeSizeVariants({ size, severity })}
      role="status"
      aria-label={`Severity: ${severityLabels[severity]} (Score: ${score.toFixed(1)})`}
    >
      {showLabel && <span className="mr-1">{severityLabels[severity]}</span>}
      <span>{score.toFixed(1)}</span>
    </span>
  );
}

// ============================================================================
// Score Display Component
// ============================================================================

export function ScoreDisplay({
  score,
  severity,
  label,
  description,
  size = 'default',
  showGlow = false,
  className,
  children,
}: ScoreDisplayProps) {
  return (
    <div
      className={cn(
        severityVariants({ severity, size }),
        showGlow && 'shadow-lg',
        className
      )}
      role="region"
      aria-label={`${label || 'Score'}: ${score.toFixed(1)} - ${severity}`}
    >
      {label && (
        <div className={cn(labelSizeVariants({ size }), 'text-center mb-2 opacity-75')}>
          {label}
        </div>
      )}
      <div className={cn('text-center', scoreSizeVariants({ size }))}>
        {score.toFixed(1)}
      </div>
      {description && (
        <div className={cn(labelSizeVariants({ size }), 'text-center mt-2 opacity-70')}>
          {description}
        </div>
      )}
      {children && (
        <div className="mt-4">
          {children}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Score Breakdown Component
// ============================================================================

export function ScoreBreakdown({
  baseScore,
  temporalScore,
  environmentalScore,
  threatScore,
  finalSeverity,
  size = 'default',
  showBreakdown = true,
  className,
}: ScoreBreakdownProps) {
  const getSeverityForScore = (score: number): CVSSSeverity => {
    if (score >= 9.0) return 'CRITICAL';
    if (score >= 7.0) return 'HIGH';
    if (score >= 4.0) return 'MEDIUM';
    if (score > 0.0) return 'LOW';
    return 'NONE';
  };

  const getSeverityColor = (severity: CVSSSeverity): string => {
    const colors = {
      NONE: 'text-gray-600',
      LOW: 'text-yellow-600',
      MEDIUM: 'text-orange-600',
      HIGH: 'text-red-600',
      CRITICAL: 'text-purple-600',
    };
    return colors[severity];
  };

  return (
    <div
      className={cn(
        'rounded-xl border-2 bg-gradient-to-br from-slate-50 to-slate-100 p-6 shadow-sm',
        className
      )}
      role="region"
      aria-label="CVSS Score Breakdown"
    >
      {/* Base Score */}
      <div className="text-center mb-4">
        <div className="text-sm text-slate-600 mb-2">Base Score</div>
        <div className="text-4xl font-bold text-slate-900">
          {baseScore.toFixed(1)}
        </div>
        <ScoreBadge
          severity={getSeverityForScore(baseScore)}
          score={baseScore}
          size="sm"
          className="mt-2"
        />
      </div>

      {showBreakdown && (
        <div className="space-y-3 border-t border-slate-200 pt-4">
          {/* Temporal Score (CVSS v3.x) */}
          {temporalScore !== undefined && (
            <div className="flex justify-between items-center py-2 border-b border-slate-100">
              <span className="text-sm text-slate-600">Temporal Score</span>
              <div className="flex items-center gap-2">
                <span className="text-lg font-semibold text-slate-900">
                  {temporalScore.toFixed(1)}
                </span>
                <ScoreBadge
                  severity={getSeverityForScore(temporalScore)}
                  score={temporalScore}
                  size="sm"
                />
              </div>
            </div>
          )}

          {/* Environmental Score */}
          {environmentalScore !== undefined && (
            <div className="flex justify-between items-center py-2 border-b border-slate-100">
              <span className="text-sm text-slate-600">Environmental Score</span>
              <div className="flex items-center gap-2">
                <span className="text-lg font-semibold text-slate-900">
                  {environmentalScore.toFixed(1)}
                </span>
                <ScoreBadge
                  severity={getSeverityForScore(environmentalScore)}
                  score={environmentalScore}
                  size="sm"
                />
              </div>
            </div>
          )}

          {/* Threat Score (CVSS v4.0) */}
          {threatScore !== undefined && (
            <div className="flex justify-between items-center py-2 border-b border-slate-100">
              <span className="text-sm text-slate-600">Threat Score</span>
              <div className="flex items-center gap-2">
                <span className="text-lg font-semibold text-slate-900">
                  {threatScore.toFixed(1)}
                </span>
                <ScoreBadge
                  severity={getSeverityForScore(threatScore)}
                  score={threatScore}
                  size="sm"
                />
              </div>
            </div>
          )}

          {/* Final Score */}
          <div className="flex justify-between items-center py-3 mt-2">
            <span className="text-sm font-medium text-slate-700">Final Score</span>
            <div className="flex items-center gap-2">
              <span className={cn('text-2xl font-bold', getSeverityColor(finalSeverity))}>
                {Math.max(baseScore, temporalScore ?? 0, environmentalScore ?? 0, threatScore ?? 0).toFixed(1)}
              </span>
              <ScoreBadge
                severity={finalSeverity}
                score={Math.max(baseScore, temporalScore ?? 0, environmentalScore ?? 0, threatScore ?? 0)}
                size="md"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Animated Score Display
// ============================================================================

export interface AnimatedScoreDisplayProps {
  score: number;
  severity: CVSSSeverity;
  label?: string;
  animate?: boolean;
}

export function AnimatedScoreDisplay({
  score,
  severity,
  label = 'Score',
  animate = true,
}: AnimatedScoreDisplayProps) {
  const [displayScore, setDisplayScore] = React.useState(0);
  const [prevScore, setPrevScore] = React.useState(0);

  React.useEffect(() => {
    if (!animate) {
      setDisplayScore(score);
      return;
    }

    setPrevScore(displayScore);
    const duration = 500;
    const startTime = performance.now();
    const startValue = displayScore;

    const animateValue = (currentTime: number) => {
      const elapsed = currentTime - startTime;
      const progress = Math.min(elapsed / duration, 1);

      // Ease-out cubic function
      const easeOut = 1 - Math.pow(1 - progress, 3);
      setDisplayScore(startValue + (score - startValue) * easeOut);

      if (progress < 1) {
        requestAnimationFrame(animateValue);
      }
    };

    requestAnimationFrame(animateValue);
  }, [score, animate]);

  return (
    <div className="relative">
      <ScoreDisplay
        score={displayScore}
        severity={severity}
        label={label}
        showGlow={true}
      />
      {animate && displayScore !== score && (
        <div
          className="absolute inset-0 rounded-xl transition-opacity duration-300 pointer-events-none"
          style={{
            background: `linear-gradient(135deg, ${
              severity === 'CRITICAL'
                ? 'rgba(147, 51, 234, 0.1)'
                : severity === 'HIGH'
                ? 'rgba(239, 68, 68, 0.1)'
                : severity === 'MEDIUM'
                ? 'rgba(249, 115, 22, 0.1)'
                : severity === 'LOW'
                ? 'rgba(234, 179, 8, 0.1)'
                : 'rgba(156, 163, 175, 0.1)'
            })`,
          }}
        />
      )}
    </div>
  );
}

// Re-export React import for component usage
import React from 'react';
