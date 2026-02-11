'use client';

/**
 * CVSS Severity Gauge Component
 * Visual gauge showing severity levels (None, Low, Medium, High, Critical)
 */

import React, { useMemo } from 'react';
import { cn } from '@/lib/utils';
import type { CVSSSeverity } from '@/lib/types';

// ============================================================================
// Types
// ============================================================================

export interface SeverityGaugeProps {
  /** Current score (0.0-10.0) */
  score: number;
  /** Current severity level */
  severity: CVSSSeverity;
  /** Gauge orientation */
  orientation?: 'horizontal' | 'vertical';
  /** Gauge size */
  size?: 'sm' | 'md' | 'lg';
  /** Show numeric labels */
  showLabels?: boolean;
  /** Show score value */
  showScore?: boolean;
  /** Custom className */
  className?: string;
  /** Animation duration in ms */
  animationDuration?: number;
}

// ============================================================================
// Severity Configuration
// ============================================================================

interface SeverityLevel {
  name: CVSSSeverity;
  label: string;
  range: [number, number];
  color: string;
  bgColor: string;
}

const severityLevels: SeverityLevel[] = [
  {
    name: 'NONE',
    label: 'None',
    range: [0.0, 0.0],
    color: '#6b7280',
    bgColor: 'bg-gray-400',
  },
  {
    name: 'LOW',
    label: 'Low',
    range: [0.1, 3.9],
    color: '#eab308',
    bgColor: 'bg-yellow-400',
  },
  {
    name: 'MEDIUM',
    label: 'Medium',
    range: [4.0, 6.9],
    color: '#f97316',
    bgColor: 'bg-orange-500',
  },
  {
    name: 'HIGH',
    label: 'High',
    range: [7.0, 8.9],
    color: '#ef4444',
    bgColor: 'bg-red-500',
  },
  {
    name: 'CRITICAL',
    label: 'Critical',
    range: [9.0, 10.0],
    color: '#9333ea',
    bgColor: 'bg-purple-600',
  },
];

// ============================================================================
// Helper Functions
// ============================================================================

function getSeverityRangeForScore(score: number): number {
  if (score >= 9.0) return 4; // CRITICAL
  if (score >= 7.0) return 3; // HIGH
  if (score >= 4.0) return 2; // MEDIUM
  if (score > 0.0) return 1; // LOW
  return 0; // NONE
}

function getScorePosition(score: number, maxScore: number = 10): number {
  return Math.min(Math.max(score / maxScore, 0), 1) * 100;
}

// ============================================================================
// Horizontal Gauge Component
// ============================================================================

interface HorizontalGaugeProps {
  score: number;
  severity: CVSSSeverity;
  size: 'sm' | 'md' | 'lg';
  showLabels: boolean;
  showScore: boolean;
  animationDuration: number;
}

const sizeClasses = {
  sm: 'h-2',
  md: 'h-3',
  lg: 'h-4',
};

const markerSizeClasses = {
  sm: 'w-3 h-3',
  md: 'w-4 h-4',
  lg: 'w-5 h-5',
};

function HorizontalGauge({
  score,
  severity,
  size,
  showLabels,
  showScore,
  animationDuration,
}: HorizontalGaugeProps) {
  const [animatedScore, setAnimatedScore] = React.useState(0);

  React.useEffect(() => {
    const startTime = performance.now();
    const startValue = animatedScore;

    const animate = (currentTime: number) => {
      const elapsed = currentTime - startTime;
      const progress = Math.min(elapsed / animationDuration, 1);

      // Ease-out cubic
      const easeOut = 1 - Math.pow(1 - progress, 3);
      setAnimatedScore(startValue + (score - startValue) * easeOut);

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [score, animationDuration]);

  const scorePosition = getScorePosition(animatedScore);
  const severityIndex = getSeverityRangeForScore(animatedScore);
  const currentSeverity = severityLevels[severityIndex];

  return (
    <div className="w-full">
      {/* Gauge Bar */}
      <div className="relative w-full">
        {/* Background segments */}
        <div className={cn('flex w-full rounded-full overflow-hidden', sizeClasses[size])}>
          {severityLevels.map((level) => {
            const width = ((level.range[1] - level.range[0] + (level.name === 'NONE' ? 0.1 : 0)) / 10) * 100;
            return (
              <div
                key={level.name}
                className={cn(level.bgColor, 'transition-all duration-300')}
                style={{ width: `${width}%` }}
                aria-label={`${level.label} (${level.range[0]}-${level.range[1]})`}
              />
            );
          })}
        </div>

        {/* Score Marker */}
        <div
          className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 transition-all duration-300 ease-out"
          style={{ left: `${scorePosition}%` }}
        >
          <div
            className={cn(
              'rounded-full border-2 border-white shadow-lg',
              markerSizeClasses[size]
            )}
            style={{ backgroundColor: currentSeverity.color }}
            role="presentation"
          />
        </div>
      </div>

      {/* Labels */}
      {showLabels && (
        <div className="flex justify-between mt-2 px-1">
          {severityLevels.map((level) => (
            <div
              key={level.name}
              className="text-xs text-slate-600 text-center"
              style={{
                color: level.name === severity ? level.color : undefined,
                fontWeight: level.name === severity ? 'bold' : 'normal',
              }}
            >
              {level.label}
            </div>
          ))}
        </div>
      )}

      {/* Score Display */}
      {showScore && (
        <div className="text-center mt-3">
          <div
            className="text-2xl font-bold"
            style={{ color: currentSeverity.color }}
          >
            {animatedScore.toFixed(1)}
          </div>
          <div className="text-sm text-slate-600">{currentSeverity.label}</div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Vertical Gauge Component
// ============================================================================

interface VerticalGaugeProps {
  score: number;
  severity: CVSSSeverity;
  size: 'sm' | 'md' | 'lg';
  showLabels: boolean;
  showScore: boolean;
  animationDuration: number;
}

const verticalSizeClasses = {
  sm: 'w-2',
  md: 'w-3',
  lg: 'w-4',
};

const verticalMarkerSizeClasses = {
  sm: 'w-3 h-3',
  md: 'w-4 h-4',
  lg: 'w-5 h-5',
};

function VerticalGauge({
  score,
  severity,
  size,
  showLabels,
  showScore,
  animationDuration,
}: VerticalGaugeProps) {
  const [animatedScore, setAnimatedScore] = React.useState(0);

  React.useEffect(() => {
    const startTime = performance.now();
    const startValue = animatedScore;

    const animate = (currentTime: number) => {
      const elapsed = currentTime - startTime;
      const progress = Math.min(elapsed / animationDuration, 1);

      // Ease-out cubic
      const easeOut = 1 - Math.pow(1 - progress, 3);
      setAnimatedScore(startValue + (score - startValue) * easeOut);

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [score, animationDuration]);

  const scorePosition = getScorePosition(animatedScore);
  const severityIndex = getSeverityRangeForScore(animatedScore);
  const currentSeverity = severityLevels[severityIndex];

  return (
    <div className="flex flex-col items-center">
      {/* Score Display */}
      {showScore && (
        <div className="text-center mb-3">
          <div
            className="text-2xl font-bold"
            style={{ color: currentSeverity.color }}
          >
            {animatedScore.toFixed(1)}
          </div>
          <div className="text-sm text-slate-600">{currentSeverity.label}</div>
        </div>
      )}

      {/* Gauge Bar */}
      <div className="relative h-48">
        {/* Background segments */}
        <div
          className={cn(
            'flex flex-col-reverse rounded-full overflow-hidden',
            verticalSizeClasses[size]
          )}
        >
          {severityLevels.map((level) => {
            const height = ((level.range[1] - level.range[0] + (level.name === 'NONE' ? 0.1 : 0)) / 10) * 100;
            return (
              <div
                key={level.name}
                className={cn(level.bgColor, 'transition-all duration-300')}
                style={{ height: `${height}%` }}
                aria-label={`${level.label} (${level.range[0]}-${level.range[1]})`}
              />
            );
          })}
        </div>

        {/* Score Marker */}
        <div
          className="absolute left-1/2 -translate-x-1/2 transition-all duration-300 ease-out"
          style={{ bottom: `${scorePosition}%` }}
        >
          <div
            className={cn(
              'rounded-full border-2 border-white shadow-lg',
              verticalMarkerSizeClasses[size]
            )}
            style={{ backgroundColor: currentSeverity.color }}
            role="presentation"
          />
        </div>
      </div>

      {/* Labels */}
      {showLabels && (
        <div className="flex flex-col justify-around mt-2 h-full py-1">
          {severityLevels.slice().reverse().map((level) => (
            <div
              key={level.name}
              className="text-xs text-slate-600"
              style={{
                color: level.name === severity ? level.color : undefined,
                fontWeight: level.name === severity ? 'bold' : 'normal',
              }}
            >
              {level.label}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Semi-Circular Gauge Component
// ============================================================================

export interface SemiCircularGaugeProps {
  score: number;
  severity: CVSSSeverity;
  size?: 'sm' | 'md' | 'lg';
  showScore?: boolean;
  animationDuration?: number;
  className?: string;
}

const semiCircularSizeConfig = {
  sm: { width: 120, height: 60, strokeWidth: 8 },
  md: { width: 160, height: 80, strokeWidth: 12 },
  lg: { width: 200, height: 100, strokeWidth: 16 },
};

export function SemiCircularGauge({
  score,
  severity,
  size = 'md',
  showScore = true,
  animationDuration = 500,
  className,
}: SemiCircularGaugeProps) {
  const [animatedScore, setAnimatedScore] = React.useState(0);

  React.useEffect(() => {
    const startTime = performance.now();
    const startValue = animatedScore;

    const animate = (currentTime: number) => {
      const elapsed = currentTime - startTime;
      const progress = Math.min(elapsed / animationDuration, 1);

      // Ease-out cubic
      const easeOut = 1 - Math.pow(1 - progress, 3);
      setAnimatedScore(startValue + (score - startValue) * easeOut);

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [score, animationDuration]);

  const severityIndex = getSeverityRangeForScore(animatedScore);
  const currentSeverity = severityLevels[severityIndex];

  const config = semiCircularSizeConfig[size];
  const radius = (config.width - config.strokeWidth) / 2;
  const circumference = Math.PI * radius;
  const scorePercent = animatedScore / 10;
  const strokeDashoffset = circumference * (1 - scorePercent);

  return (
    <div className={cn('flex flex-col items-center', className)}>
      <svg
        width={config.width}
        height={config.height + (showScore ? 40 : 0)}
        viewBox={`0 0 ${config.width} ${config.height + (showScore ? 40 : 0)}`}
        className="overflow-visible"
      >
        {/* Background arc */}
        <path
          d={`M ${config.strokeWidth / 2} ${config.height} A ${radius} ${radius} 0 0 1 ${config.width - config.strokeWidth / 2} ${config.height}`}
          fill="none"
          stroke="#e5e7eb"
          strokeWidth={config.strokeWidth}
          strokeLinecap="round"
        />

        {/* Colored segments */}
        {severityLevels.map((level, index) => {
          const segmentStart = level.range[0] / 10;
          const segmentEnd = Math.min(level.range[1] / 10, 1);
          const segmentLength = segmentEnd - segmentStart;

          const startAngle = Math.PI * (1 - segmentStart);
          const endAngle = Math.PI * (1 - segmentEnd);

          const startX = config.strokeWidth / 2 + radius * (1 + Math.cos(startAngle));
          const startY = config.height - radius * Math.sin(startAngle);
          const endX = config.strokeWidth / 2 + radius * (1 + Math.cos(endAngle));
          const endY = config.height - radius * Math.sin(endAngle);

          const largeArc = segmentLength > 0.5 ? 1 : 0;

          return (
            <path
              key={level.name}
              d={`M ${startX} ${startY} A ${radius} ${radius} 0 ${largeArc} 1 ${endX} ${endY}`}
              fill="none"
              stroke={level.color}
              strokeWidth={config.strokeWidth}
              strokeLinecap="round"
              className="transition-opacity duration-300"
              style={{
                opacity: animatedScore >= level.range[0] ? 1 : 0.3,
              }}
            />
          );
        })}

        {/* Needle */}
        <g
          transform={`translate(${config.width / 2}, ${config.height})`}
          style={{
            transformOrigin: 'center',
            transform: `translate(${config.width / 2}px, ${config.height}px) rotate(${180 - animatedScore * 18}deg)`,
          }}
          transition="transform 0.3s ease-out"
        >
          <polygon
            points={`0,${-radius + 5} -4,0 4,0`}
            fill={currentSeverity.color}
            className="transition-all duration-300"
          />
          <circle
            cx="0"
            cy="0"
            r="4"
            fill={currentSeverity.color}
            className="transition-all duration-300"
          />
        </g>

        {/* Score text */}
        {showScore && (
          <text
            x={config.width / 2}
            y={config.height + 30}
            textAnchor="middle"
            className="text-2xl font-bold"
            fill={currentSeverity.color}
          >
            {animatedScore.toFixed(1)}
          </text>
        )}
      </svg>
    </div>
  );
}

// ============================================================================
// Main Severity Gauge Component
// ============================================================================

export function SeverityGauge({
  score,
  severity,
  orientation = 'horizontal',
  size = 'md',
  showLabels = true,
  showScore = true,
  className,
  animationDuration = 500,
}: SeverityGaugeProps) {
  const gaugeProps = {
    score,
    severity,
    size,
    showLabels,
    showScore,
    animationDuration,
  };

  return (
    <div className={className} role="region" aria-label={`Severity gauge: ${severity} (${score.toFixed(1)})`}>
      {orientation === 'vertical' ? (
        <VerticalGauge {...gaugeProps} />
      ) : (
        <HorizontalGauge {...gaugeProps} />
      )}
    </div>
  );
}

// ============================================================================
// Severity Legend Component
// ============================================================================

export interface SeverityLegendProps {
  /** Custom className */
  className?: string;
  /** Display direction */
  direction?: 'row' | 'column';
  /** Show score ranges */
  showRanges?: boolean;
}

export function SeverityLegend({
  className,
  direction = 'row',
  showRanges = true,
}: SeverityLegendProps) {
  return (
    <div
      className={cn(
        'flex gap-4',
        direction === 'row' ? 'flex-row flex-wrap' : 'flex-col',
        className
      )}
      role="list"
      aria-label="Severity levels legend"
    >
      {severityLevels.map((level) => (
        <div
          key={level.name}
          className="flex items-center gap-2"
          role="listitem"
        >
          <div
            className={cn('rounded-full', level.bgColor)}
            style={{ width: '12px', height: '12px' }}
          />
          <div className="text-sm">
            <span className="font-medium">{level.label}</span>
            {showRanges && (
              <span className="text-slate-500 ml-1">
                ({level.range[0]}-{level.range[1]})
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
