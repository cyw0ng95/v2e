'use client';

/**
 * CVSS Keyboard Accessibility Components
 * Enhanced keyboard navigation and ARIA support for CVSS calculator
 */

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';

// ============================================================================
// Types
// ============================================================================

export interface KeyboardNavigableProps {
  children: React.ReactNode;
  className?: string;
  orientation?: 'horizontal' | 'vertical' | 'grid';
  loop?: boolean;
  onEscape?: () => void;
  ariaLabel?: string;
}

export interface KeyboardOptionProps {
  value: string;
  label: string;
  description?: string;
  disabled?: boolean;
  selected?: boolean;
  onSelect: () => void;
  onFocus?: () => void;
  hotkey?: string;
}

export interface KeyboardRadioGroupProps {
  name: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: Array<{
    value: string;
    label: string;
    description?: string;
    disabled?: boolean;
  }>;
  orientation?: 'horizontal' | 'vertical';
  description?: string;
  className?: string;
}

// ============================================================================
// Keyboard Navigable Container
// ============================================================================

export function KeyboardNavigable({
  children,
  className,
  orientation = 'horizontal',
  loop = true,
  onEscape,
  ariaLabel,
}: KeyboardNavigableProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [focusedIndex, setFocusedIndex] = useState<number>(-1);
  const focusableItems = useRef<HTMLElement[]>([]);

  // Update focusable items on mount and children change
  useEffect(() => {
    if (containerRef.current) {
      focusableItems.current = Array.from(
        containerRef.current.querySelectorAll<HTMLElement>(
          '[role="button"], [role="option"], button, [tabindex="0"]'
        )
      );
    }
  }, [children]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    const items = focusableItems.current;
    if (items.length === 0) return;

    const currentIndex = items.indexOf(document.activeElement as HTMLElement);
    let nextIndex = currentIndex;

    switch (e.key) {
      case 'ArrowRight':
        if (orientation === 'horizontal' || orientation === 'grid') {
          e.preventDefault();
          nextIndex = currentIndex + 1;
          if (nextIndex >= items.length && loop) nextIndex = 0;
        }
        break;

      case 'ArrowLeft':
        if (orientation === 'horizontal' || orientation === 'grid') {
          e.preventDefault();
          nextIndex = currentIndex - 1;
          if (nextIndex < 0 && loop) nextIndex = items.length - 1;
        }
        break;

      case 'ArrowDown':
        if (orientation === 'vertical' || orientation === 'grid') {
          e.preventDefault();
          nextIndex = currentIndex + (orientation === 'grid' && Math.floor(currentIndex / 4) < Math.floor((currentIndex + 4) / 4)
            ? 4  // Move to next row in grid
            : 1); // Move down in vertical
          if (nextIndex >= items.length && loop) nextIndex = 0;
        }
        break;

      case 'ArrowUp':
        if (orientation === 'vertical' || orientation === 'grid') {
          e.preventDefault();
          nextIndex = currentIndex - (orientation === 'grid' && currentIndex >= 4
            ? 4  // Move to previous row in grid
            : 1); // Move up in vertical
          if (nextIndex < 0 && loop) nextIndex = items.length - 1;
        }
        break;

      case 'Home':
        e.preventDefault();
        nextIndex = 0;
        break;

      case 'End':
        e.preventDefault();
        nextIndex = items.length - 1;
        break;

      case 'Escape':
        e.preventDefault();
        onEscape?.();
        return;

      default:
        return;
    }

    if (nextIndex >= 0 && nextIndex < items.length && nextIndex !== currentIndex) {
      items[nextIndex]?.focus();
      setFocusedIndex(nextIndex);
    }
  }, [orientation, loop, onEscape]);

  return (
    <div
      ref={containerRef}
      className={cn('outline-none', className)}
      role={orientation === 'grid' ? 'grid' : orientation === 'vertical' ? 'listbox' : 'group'}
      aria-label={ariaLabel}
      aria-orientation={orientation === 'grid' ? undefined : orientation}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
    >
      {children}
    </div>
  );
}

// ============================================================================
// Keyboard Accessible Radio Group
// ============================================================================

export function KeyboardRadioGroup({
  name,
  label,
  value,
  onChange,
  options,
  orientation = 'horizontal',
  description,
  className,
}: KeyboardRadioGroupProps) {
  const [focusedValue, setFocusedValue] = useState<string | null>(null);

  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLButtonElement>, optionValue: string) => {
    const items = options.map(o => o.value);
    const currentIndex = items.indexOf(optionValue);
    let nextIndex = currentIndex;

    switch (e.key) {
      case 'ArrowRight':
      case 'ArrowDown':
        e.preventDefault();
        nextIndex = currentIndex + 1;
        if (nextIndex >= items.length) nextIndex = 0;
        break;

      case 'ArrowLeft':
      case 'ArrowUp':
        e.preventDefault();
        nextIndex = currentIndex - 1;
        if (nextIndex < 0) nextIndex = items.length - 1;
        break;

      case 'Home':
        e.preventDefault();
        nextIndex = 0;
        break;

      case 'End':
        e.preventDefault();
        nextIndex = items.length - 1;
        break;

      case 'Enter':
      case ' ':
        e.preventDefault();
        onChange(optionValue);
        return;

      default:
        // Check for number keys 1-9
        const numKey = parseInt(e.key);
        if (!isNaN(numKey) && numKey > 0 && numKey <= items.length) {
          e.preventDefault();
          onChange(items[numKey - 1]);
        }
        return;
    }

    if (nextIndex >= 0 && nextIndex < items.length && nextIndex !== currentIndex) {
      setFocusedValue(items[nextIndex]);
      // Focus the next element
      e.currentTarget.parentNode?.querySelectorAll<HTMLButtonElement>(
        `[role="radio"][data-value="${items[nextIndex]}"]`
      )?.[0]?.focus();
    }
  }, [options, onChange]);

  return (
    <div
      className={cn('space-y-2', className)}
      role="radiogroup"
      aria-label={label}
      aria-orientation={orientation}
    >
      {/* Group Label */}
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-slate-700">{label}</span>
        {description && (
          <span className="text-xs text-slate-500">({description})</span>
        )}
        <span className="text-xs text-slate-400">
          Use arrow keys or 1-{options.length} to select
        </span>
      </div>

      {/* Radio Options */}
      <div
        className={cn(
          'flex gap-2',
          orientation === 'vertical' ? 'flex-col' : 'flex-wrap'
        )}
      >
        {options.map((option, index) => (
          <button
            key={option.value}
            type="button"
            role="radio"
            data-value={option.value}
            aria-checked={value === option.value}
            aria-label={`${option.label}${option.description ? `: ${option.description}` : ''}`}
            disabled={option.disabled}
            className={cn(
              'px-3 py-2 rounded-lg border-2 transition-all duration-150',
              'focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-offset-1',
              'text-left',
              value === option.value
                ? 'border-blue-500 bg-blue-50 text-blue-700'
                : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300',
              option.disabled && 'opacity-50 cursor-not-allowed'
            )}
            onClick={() => !option.disabled && onChange(option.value)}
            onFocus={() => setFocusedValue(option.value)}
            onBlur={() => setFocusedValue(null)}
            onKeyDown={(e) => handleKeyDown(e, option.value)}
          >
            <div className="flex items-center gap-2">
              {/* Keyboard hint */}
              <span className={cn(
                'w-5 h-5 rounded bg-slate-100 text-slate-500',
                'flex items-center justify-center text-xs font-mono',
                focusedValue === option.value && 'bg-blue-100 text-blue-600'
              )}>
                {index + 1}
              </span>
              <div>
                <div className="font-medium text-sm">{option.label}</div>
                {option.description && (
                  <div className="text-xs text-slate-500">{option.description}</div>
                )}
              </div>
            </div>
          </button>
        ))}
      </div>

      {/* Announcer for screen readers */}
      <div
        role="status"
        aria-live="polite"
        className="sr-only"
      >
        {value && `${label} changed to ${options.find(o => o.value === value)?.label}`}
      </div>
    </div>
  );
}

// ============================================================================
// Keyboard Accessible Metric Selector
// ============================================================================

export interface KeyboardMetricSelectorProps {
  label: string;
  abbreviation: string;
  value: string;
  onChange: (value: string) => void;
  options: Array<{
    value: string;
    label: string;
    abbreviation: string;
    description?: string;
  }>;
  description?: string;
  className?: string;
}

export function KeyboardMetricSelector({
  label,
  abbreviation,
  value,
  onChange,
  options,
  description,
  className,
}: KeyboardMetricSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);

  // Handle keyboard navigation within dropdown
  const handleDropdownKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (!isOpen) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setFocusedIndex(prev => Math.min(prev + 1, options.length - 1));
        break;

      case 'ArrowUp':
        e.preventDefault();
        setFocusedIndex(prev => Math.max(prev - 1, 0));
        break;

      case 'Enter':
      case ' ':
        e.preventDefault();
        if (focusedIndex >= 0 && focusedIndex < options.length) {
          onChange(options[focusedIndex].value);
          setIsOpen(false);
        }
        break;

      case 'Escape':
        e.preventDefault();
        setIsOpen(false);
        triggerRef.current?.focus();
        break;

      case 'Home':
        e.preventDefault();
        setFocusedIndex(0);
        break;

      case 'End':
        e.preventDefault();
        setFocusedIndex(options.length - 1);
        break;

      case 'Tab':
        setIsOpen(false);
        break;
    }
  }, [isOpen, focusedIndex, options, onChange]);

  // Scroll focused option into view
  useEffect(() => {
    if (isOpen && focusedIndex >= 0) {
      const focusedElement = containerRef.current?.querySelector(
        `[data-index="${focusedIndex}"]`
      );
      focusedElement?.scrollIntoView({ block: 'nearest' });
    }
  }, [isOpen, focusedIndex]);

  const selectedOption = options.find(o => o.value === value);

  return (
    <div className={cn('relative', className)}>
      {/* Label */}
      <label className="block text-sm font-medium text-slate-700 mb-1">
        {label}
        <span className="ml-2 px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-xs">
          {abbreviation}
        </span>
      </label>
      {description && (
        <p className="text-xs text-slate-500 mb-2">{description}</p>
      )}

      {/* Trigger Button */}
      <button
        ref={triggerRef}
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
            e.preventDefault();
            setIsOpen(true);
          }
        }}
        className={cn(
          'w-full flex items-center justify-between px-4 py-3 rounded-lg border-2',
          'focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-offset-1',
          'transition-all duration-200',
          isOpen
            ? 'border-blue-500 ring-2 ring-blue-200'
            : 'border-slate-200 hover:border-blue-300'
        )}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-labelledby={`${label}-label`}
      >
        <div className="flex items-center gap-3">
          <span className="text-2xl font-bold text-blue-600">
            {selectedOption?.abbreviation || value}
          </span>
          <div className="text-left">
            <div className="font-medium text-slate-900">{selectedOption?.label}</div>
            {selectedOption?.description && (
              <div className="text-xs text-slate-500">{selectedOption.description}</div>
            )}
          </div>
        </div>
        <svg
          className={cn('h-4 w-4 text-slate-400 transition-transform', isOpen && 'rotate-180')}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Dropdown */}
      {isOpen && (
        <div
          ref={containerRef}
          role="listbox"
          aria-activedescendant={focusedIndex >= 0 ? `option-${focusedIndex}` : undefined}
          className="absolute top-full left-0 right-0 mt-1 z-50 bg-white rounded-lg border shadow-lg max-h-64 overflow-auto"
          onKeyDown={handleDropdownKeyDown}
        >
          {options.map((option, index) => (
            <div
              key={option.value}
              role="option"
              id={`option-${index}`}
              data-index={index}
              className={cn(
                'px-4 py-3 cursor-pointer transition-colors',
                'flex items-center gap-3',
                index === focusedIndex && 'bg-blue-50',
                value === option.value && 'bg-blue-100 text-blue-700',
                index !== focusedIndex && value !== option.value && 'hover:bg-slate-50'
              )}
              onClick={() => {
                onChange(option.value);
                setIsOpen(false);
              }}
              onMouseEnter={() => setFocusedIndex(index)}
              aria-selected={value === option.value}
            >
              <span className={cn(
                'text-xl font-bold w-8 h-8 flex items-center justify-center rounded',
                index === focusedIndex ? 'bg-blue-200 text-blue-700' : 'bg-slate-100 text-slate-600'
              )}>
                {option.abbreviation}
              </span>
              <div>
                <div className="font-medium">{option.label}</div>
                {option.description && (
                  <div className="text-xs text-slate-500">{option.description}</div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Screen reader announcer */}
      <div
        role="status"
        aria-live="polite"
        className="sr-only"
      >
        {selectedOption && `${label} set to ${selectedOption.label}`}
      </div>
    </div>
  );
}

// ============================================================================
// Skip Links Component
// ============================================================================

export interface SkipLinksProps {
  onSkipToMain?: () => void;
  onSkipToMetrics?: () => void;
  onSkipToScore?: () => void;
}

export function SkipLinks({
  onSkipToMain,
  onSkipToMetrics,
  onSkipToScore,
}: SkipLinksProps) {
  return (
    <div className="sr-only focus-within:not-sr-only">
      <a
        href="#"
        onClick={(e) => {
          e.preventDefault();
          onSkipToMain?.();
        }}
        className="absolute top-4 left-4 px-4 py-2 bg-blue-600 text-white rounded-lg z-50"
      >
        Skip to main content
      </a>
      <a
        href="#"
        onClick={(e) => {
          e.preventDefault();
          onSkipToMetrics?.();
        }}
        className="absolute top-14 left-4 px-4 py-2 bg-blue-600 text-white rounded-lg z-50"
      >
        Skip to metrics
      </a>
      <a
        href="#"
        onClick={(e) => {
          e.preventDefault();
          onSkipToScore?.();
        }}
        className="absolute top-24 left-4 px-4 py-2 bg-blue-600 text-white rounded-lg z-50"
      >
        Skip to score
      </a>
    </div>
  );
}
