'use client';

/**
 * CVSS Metric Card Component
 * Reusable metric selection card component for CVSS calculator
 */

import React, { useState, useCallback } from 'react';
import { Info, ChevronDown, ChevronUp } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

// ============================================================================
// Types
// ============================================================================

export interface MetricOption {
  /** Value identifier (e.g., "N", "L", "H") */
  value: string;
  /** Display label (e.g., "Network", "Low", "High") */
  label: string;
  /** Short description */
  description: string;
  /** Optional abbreviation */
  abbreviation?: string;
  /** Whether this is the recommended/default value */
  recommended?: boolean;
  /** Whether this option is disabled */
  disabled?: boolean;
}

export interface MetricCardProps {
  /** Metric name (e.g., "Attack Vector", "Confidentiality") */
  name: string;
  /** Metric abbreviation (e.g., "AV", "C") */
  abbreviation: string;
  /** Current selected value */
  value: string;
  /** Callback when value changes */
  onChange: (value: string) => void;
  /** Available options */
  options: MetricOption[];
  /** Detailed description of the metric */
  description?: string;
  /** Show descriptions on cards */
  showDescriptions?: boolean;
  /** Display orientation */
  orientation?: 'horizontal' | 'vertical' | 'grid';
  /** Grid columns (for grid orientation) */
  gridCols?: 2 | 3 | 4;
  /** Card size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Show tooltip with more info */
  showTooltip?: boolean;
  /** Tooltip content */
  tooltipContent?: React.ReactNode;
  /** Compact mode (for mobile) */
  compact?: boolean;
  /** Disabled state */
  disabled?: boolean;
  /** Custom className */
  className?: string;
}

// ============================================================================
// Metric Option Button Component
// ============================================================================

interface MetricOptionButtonProps {
  option: MetricOption;
  isSelected: boolean;
  onSelect: () => void;
  size: 'sm' | 'md' | 'lg';
  showDescription: boolean;
  disabled: boolean;
  orientation: 'horizontal' | 'vertical' | 'grid';
}

function MetricOptionButton({
  option,
  isSelected,
  onSelect,
  size,
  showDescription,
  disabled,
  orientation,
}: MetricOptionButtonProps) {
  const [isHovered, setIsHovered] = useState(false);

  const sizeClasses = {
    sm: 'p-2 text-xs',
    md: 'p-3 text-sm',
    lg: 'p-4 text-base',
  };

  const abbreviationSizeClasses = {
    sm: 'text-lg',
    md: 'text-2xl',
    lg: 'text-3xl',
  };

  const baseClasses = cn(
    'relative rounded-lg border-2 transition-all duration-200 cursor-pointer',
    'focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-offset-1',
    sizeClasses[size],
    isSelected
      ? 'border-blue-500 bg-blue-50 text-blue-700 shadow-sm scale-[1.02]'
      : 'border-slate-200 bg-white text-slate-600 hover:border-blue-300 hover:bg-slate-50 hover:scale-[1.01]',
    disabled && 'opacity-50 cursor-not-allowed hover:scale-100'
  );

  const handleClick = useCallback(() => {
    if (!disabled) {
      onSelect();
    }
  }, [disabled, onSelect]);

  if (orientation === 'horizontal' || orientation === 'vertical') {
    return (
      <button
        type="button"
        className={cn(
          baseClasses,
          'flex items-center gap-2 text-left',
          orientation === 'vertical' && 'flex-col text-center'
        )}
        onClick={handleClick}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
        disabled={disabled}
        aria-pressed={isSelected}
        aria-label={`${option.label} (${option.abbreviation || option.value})`}
      >
        <div className={cn('font-bold', abbreviationSizeClasses[size])}>
          {option.abbreviation || option.value}
        </div>
        <div className="flex-1 min-w-0">
          <div className="font-medium truncate">{option.label}</div>
          {showDescription && option.description && (
            <div className="text-xs text-slate-500 truncate mt-0.5">
              {option.description}
            </div>
          )}
        </div>
        {option.recommended && !isSelected && (
          <div className="absolute top-1 right-1">
            <span className="text-xs bg-blue-100 text-blue-600 px-1.5 py-0.5 rounded">
              Default
            </span>
          </div>
        )}
      </button>
    );
  }

  // Grid orientation
  return (
    <button
      type="button"
      className={cn(
        baseClasses,
        'flex flex-col items-center justify-center text-center min-h-[80px]'
      )}
      onClick={handleClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      disabled={disabled}
      aria-pressed={isSelected}
      aria-label={`${option.label} (${option.abbreviation || option.value})`}
    >
      <div className={cn('font-bold', abbreviationSizeClasses[size])}>
        {option.abbreviation || option.value}
      </div>
      <div className="text-xs font-medium mt-1">{option.label}</div>
      {showDescription && (isHovered || isSelected) && (
        <div className="absolute bottom-1 left-1 right-1">
          <div className="text-[10px] text-slate-500 bg-white/90 backdrop-blur rounded px-1 py-0.5">
            {option.description}
          </div>
        </div>
      )}
      {option.recommended && !isSelected && (
        <div className="absolute top-1 right-1">
          <div className="w-2 h-2 bg-blue-500 rounded-full" />
        </div>
      )}
    </button>
  );
}

// ============================================================================
// Main Metric Card Component
// ============================================================================

export function MetricCard({
  name,
  abbreviation,
  value,
  onChange,
  options,
  description,
  showDescriptions = true,
  orientation = 'grid',
  gridCols = 4,
  size = 'md',
  showTooltip = false,
  tooltipContent,
  compact = false,
  disabled = false,
  className,
}: MetricCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isInfoHovered, setIsInfoHovered] = useState(false);

  const gridColsClasses = {
    2: 'grid-cols-2',
    3: 'grid-cols-3',
    4: 'grid-cols-4',
  };

  const cardContent = (
    <Card className={cn(
      'transition-all duration-200',
      disabled && 'opacity-60',
      className
    )}>
      <CardContent className={cn('p-4', compact && 'p-3')}>
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h3 className={cn(
                'font-semibold',
                size === 'sm' ? 'text-sm' : size === 'lg' ? 'text-lg' : 'text-base'
              )}>
                {name}
              </h3>
              <span className={cn(
                'px-2 py-0.5 rounded bg-slate-100 text-slate-600 font-mono text-xs',
                compact && 'text-[10px]'
              )}>
                {abbreviation}
              </span>
            </div>
            {description && !compact && (
              <p className="text-xs text-slate-500 mt-1">{description}</p>
            )}
          </div>

          {/* Info button with tooltip */}
          {(showTooltip || tooltipContent) && (
            <TooltipProvider>
              <Tooltip open={isInfoHovered} onOpenChange={setIsInfoHovered}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    className="p-1 rounded hover:bg-slate-100 transition-colors"
                    aria-label={`More info about ${name}`}
                  >
                    <Info className="h-4 w-4 text-slate-400" />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="top" className="max-w-xs">
                  {tooltipContent || (
                    <div>
                      <p className="font-medium">{name} ({abbreviation})</p>
                      <p className="text-xs mt-1">{description}</p>
                    </div>
                  )}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
        </div>

        {/* Options */}
        <div
          className={cn(
            'gap-2',
            orientation === 'grid' && `grid ${gridColsClasses[gridCols]}`,
            orientation === 'horizontal' && 'flex flex-col',
            orientation === 'vertical' && 'flex flex-col'
          )}
        >
          {options.map((option) => (
            <MetricOptionButton
              key={option.value}
              option={option}
              isSelected={value === option.value}
              onSelect={() => onChange(option.value)}
              size={size}
              showDescription={showDescriptions}
              disabled={disabled || option.disabled}
              orientation={orientation}
            />
          ))}
        </div>

        {/* Expand/Collapse for descriptions */}
        {compact && showDescriptions && options.some(o => o.description) && (
          <button
            type="button"
            onClick={() => setIsExpanded(!isExpanded)}
            className="mt-3 flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700 transition-colors"
          >
            {isExpanded ? (
              <>
                <span>Hide descriptions</span>
                <ChevronUp className="h-3 w-3" />
              </>
            ) : (
              <>
                <span>Show descriptions</span>
                <ChevronDown className="h-3 w-3" />
              </>
            )}
          </button>
        )}
      </CardContent>
    </Card>
  );

  return cardContent;
}

// ============================================================================
// Compact Metric Card Component (for mobile/dense layouts)
// ============================================================================

export interface CompactMetricCardProps {
  name: string;
  abbreviation: string;
  value: string;
  onChange: (value: string) => void;
  options: MetricOption[];
  disabled?: boolean;
  className?: string;
}

export function CompactMetricCard({
  name,
  abbreviation,
  value,
  onChange,
  options,
  disabled = false,
  className,
}: CompactMetricCardProps) {
  const [isOpen, setIsOpen] = useState(false);

  const selectedOption = options.find(o => o.value === value);

  return (
    <div className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          'w-full flex items-center justify-between px-3 py-2 rounded-lg border-2',
          'transition-all duration-200',
          disabled && 'opacity-50 cursor-not-allowed',
          !disabled && 'hover:border-blue-300 bg-white'
        )}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs bg-slate-100 px-1.5 py-0.5 rounded">
            {abbreviation}
          </span>
          <span className="text-sm font-medium">{name}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="font-bold text-blue-600">{selectedOption?.abbreviation || selectedOption?.value || value}</span>
          <ChevronDown className={cn('h-4 w-4 transition-transform', isOpen && 'rotate-180')} />
        </div>
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-1 z-50 bg-white rounded-lg border shadow-lg overflow-hidden">
          <div role="listbox">
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => {
                  onChange(option.value);
                  setIsOpen(false);
                }}
                className={cn(
                  'w-full px-3 py-2 text-left hover:bg-slate-50 transition-colors',
                  'flex items-center justify-between',
                  value === option.value && 'bg-blue-50 text-blue-700'
                )}
                role="option"
                aria-selected={value === option.value}
              >
                <div>
                  <div className="font-medium">{option.label}</div>
                  {option.description && (
                    <div className="text-xs text-slate-500">{option.description}</div>
                  )}
                </div>
                <span className="font-mono font-bold">{option.abbreviation || option.value}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Metric Group Component
// ============================================================================

export interface MetricGroupProps {
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export function MetricGroup({ title, description, children, className }: MetricGroupProps) {
  return (
    <div className={cn('space-y-4', className)}>
      <div>
        <h2 className="text-lg font-bold text-slate-900">{title}</h2>
        {description && (
          <p className="text-sm text-slate-500 mt-1">{description}</p>
        )}
      </div>
      <div className="space-y-4">
        {children}
      </div>
    </div>
  );
}

// ============================================================================
// Preset Metrics Configuration
// ============================================================================

export const baseMetricOptions = {
  AV: [
    { value: 'N', label: 'Network', description: 'Exploitable over the network', abbreviation: 'N' },
    { value: 'A', label: 'Adjacent', description: 'Requires adjacent network', abbreviation: 'A' },
    { value: 'L', label: 'Local', description: 'Local access required', abbreviation: 'L' },
    { value: 'P', label: 'Physical', description: 'Physical access required', abbreviation: 'P' },
  ],
  AC: [
    { value: 'L', label: 'Low', description: 'Specialized access or conditions', abbreviation: 'L' },
    { value: 'H', label: 'High', description: 'Specialized access conditions exist', abbreviation: 'H' },
  ],
  PR: [
    { value: 'N', label: 'None', description: 'No privileges required', abbreviation: 'N', recommended: true },
    { value: 'L', label: 'Low', description: 'Basic user privileges', abbreviation: 'L' },
    { value: 'H', label: 'High', description: 'Admin/Superuser privileges', abbreviation: 'H' },
  ],
  UI: [
    { value: 'N', label: 'None', description: 'No user interaction required', abbreviation: 'N', recommended: true },
    { value: 'R', label: 'Required', description: 'User participation required', abbreviation: 'R' },
  ],
  S: [
    { value: 'U', label: 'Unchanged', description: 'Affects only vulnerable component', abbreviation: 'U', recommended: true },
    { value: 'C', label: 'Changed', description: 'Affects other components', abbreviation: 'C' },
  ],
  CIA: [
    { value: 'H', label: 'High', description: 'Total compromise', abbreviation: 'H' },
    { value: 'L', label: 'Low', description: 'Partial compromise', abbreviation: 'L' },
    { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N', recommended: true },
  ],
} as const;
