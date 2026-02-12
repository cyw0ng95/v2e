'use client';

/**
 * CVSS Vector String Component
 * Displays and manages CVSS vector strings with copy functionality
 */

import React, { useState, useCallback, useEffect } from 'react';
import { Copy, Check, Link2, Download, AlertCircle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';

// ============================================================================
// Types
// ============================================================================

export interface VectorStringProps {
  /** CVSS vector string */
  vectorString: string;
  /** Copy button label */
  copyLabel?: string;
  /** Show copy button */
  showCopyButton?: boolean;
  /** Show export button */
  showExportButton?: boolean;
  /** Show share button */
  showShareButton?: boolean;
  /** Custom className */
  className?: string;
  /** Callback when vector is copied */
  onCopy?: (vector: string) => void;
  /** Display mode */
  displayMode?: 'full' | 'compact' | 'inline';
  /** Format for highlighting segments */
  highlightSegments?: boolean;
}

export interface VectorSegment {
  /** Metric name (e.g., "AV") */
  name: string;
  /** Metric value (e.g., "N") */
  value: string;
  /** Full segment string (e.g., "AV:N") */
  segment: string;
  /** Whether this segment is modified/overridden */
  isModified?: boolean;
}

export interface VectorParseResult {
  /** Parsed segments */
  segments: VectorSegment[];
  /** CVSS version */
  version: '3.0' | '3.1' | '4.0';
  /** Prefix (e.g., "CVSS:3.1") */
  prefix: string;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Parse a CVSS vector string into segments
 */
export function parseVectorString(vectorString: string): VectorParseResult | null {
  if (!vectorString || !vectorString.startsWith('CVSS:')) {
    return null;
  }

  const parts = vectorString.split('/');

  // Extract version
  const prefix = parts[0];
  const versionMatch = prefix.match(/CVSS:(3\.0|3\.1|4\.0)/);

  if (!versionMatch) {
    return null;
  }

  const version = versionMatch[1] as '3.0' | '3.1' | '4.0';

  // Parse segments
  const segments: VectorSegment[] = [];

  for (let i = 1; i < parts.length; i++) {
    const part = parts[i];
    const colonIndex = part.indexOf(':');

    if (colonIndex === -1) {
      continue;
    }

    const name = part.substring(0, colonIndex);
    const value = part.substring(colonIndex + 1);

    segments.push({
      name,
      value,
      segment: part,
      isModified: name.startsWith('M') && name !== 'MAV',
    });
  }

  return {
    segments,
    version,
    prefix,
  };
}

/**
 * Format vector string for display
 */
function formatVectorString(
  vectorString: string,
  displayMode: 'full' | 'compact' | 'inline'
): string {
  if (displayMode === 'compact' && vectorString.length > 60) {
    return vectorString.substring(0, 57) + '...';
  }
  return vectorString;
}

// ============================================================================
// Vector Segment Display Component
// ============================================================================

interface VectorSegmentDisplayProps {
  segment: VectorSegment;
  highlight?: boolean;
  onClick?: () => void;
}

function VectorSegmentDisplay({ segment, highlight, onClick }: VectorSegmentDisplayProps) {
  const getSegmentColor = (name: string): string => {
    const baseColors = {
      AV: 'bg-blue-100 text-blue-700 border-blue-300',
      AC: 'bg-indigo-100 text-indigo-700 border-indigo-300',
      PR: 'bg-purple-100 text-purple-700 border-purple-300',
      UI: 'bg-pink-100 text-pink-700 border-pink-300',
      S: 'bg-slate-100 text-slate-700 border-slate-300',
      C: 'bg-green-100 text-green-700 border-green-300',
      I: 'bg-yellow-100 text-yellow-700 border-yellow-300',
      A: 'bg-orange-100 text-orange-700 border-orange-300',
    };

    const temporalColors = {
      E: 'bg-red-100 text-red-700 border-red-300',
      RL: 'bg-amber-100 text-amber-700 border-amber-300',
      RC: 'bg-lime-100 text-lime-700 border-lime-300',
    };

    const envPrefix = 'bg-teal-100 text-teal-700 border-teal-300';

    if (segment.name.startsWith('M') || segment.name === 'CR' || segment.name === 'IR' || segment.name === 'AR') {
      return envPrefix;
    }
    if (segment.name in temporalColors) {
      return temporalColors[segment.name as keyof typeof temporalColors];
    }
    if (segment.name in baseColors) {
      return baseColors[segment.name as keyof typeof baseColors];
    }
    return 'bg-gray-100 text-gray-700 border-gray-300';
  };

  return (
    <button
      onClick={onClick}
      className={cn(
        'px-2 py-1 rounded-md border font-mono text-sm transition-all duration-200',
        'hover:shadow-sm hover:scale-105',
        getSegmentColor(segment.name),
        segment.isModified && 'ring-2 ring-offset-1 ring-teal-400',
        onClick && 'cursor-pointer',
        highlight && 'ring-2 ring-offset-1 ring-blue-400'
      )}
      title={`${segment.name}: ${segment.value}`}
      type="button"
    >
      {segment.segment}
    </button>
  );
}

// ============================================================================
// Vector String Display Component
// ============================================================================

export function VectorStringDisplay({
  vectorString,
  displayMode = 'full',
  highlightSegments = true,
  className,
}: {
  vectorString: string;
  displayMode?: 'full' | 'compact' | 'inline';
  highlightSegments?: boolean;
  className?: string;
}) {
  const parsed = parseVectorString(vectorString);

  if (!parsed || !highlightSegments) {
    return (
      <code
        className={cn(
          'bg-slate-800 text-green-400 p-4 rounded-lg text-sm break-all font-mono',
          className
        )}
      >
        {formatVectorString(vectorString, displayMode)}
      </code>
    );
  }

  return (
    <div
      className={cn(
        'bg-slate-100 p-4 rounded-lg border border-slate-200 flex flex-wrap gap-2',
        className
      )}
    >
      <span className="font-mono text-sm text-slate-700 font-semibold">
        {parsed.prefix}
      </span>
      {parsed.segments.map((segment, index) => (
        <VectorSegmentDisplay
          key={`${segment.name}-${index}`}
          segment={segment}
          highlight={highlightSegments}
        />
      ))}
    </div>
  );
}

// ============================================================================
// Main Vector String Component
// ============================================================================

export function VectorString({
  vectorString,
  copyLabel = 'Copy',
  showCopyButton = true,
  showExportButton = false,
  showShareButton = false,
  className,
  onCopy,
  displayMode = 'full',
  highlightSegments = false,
}: VectorStringProps) {
  const [copied, setCopied] = useState(false);
  const [copiedVector, setCopiedVector] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(vectorString);
      setCopied(true);
      setCopiedVector(vectorString);
      setError(null);
      onCopy?.(vectorString);

      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      setError('Failed to copy to clipboard');
      console.error('Copy failed:', err);
      setTimeout(() => setError(null), 3000);
    }
  }, [vectorString, onCopy]);

  useEffect(() => {
    if (copiedVector !== vectorString) {
      setCopied(false);
    }
  }, [vectorString, copiedVector]);

  const parsed = parseVectorString(vectorString);

  return (
    <div className={cn('space-y-4', className)}>
      {/* Vector String Display */}
      <div className="relative group">
        {displayMode === 'inline' ? (
          <code className="bg-slate-100 px-2 py-1 rounded text-sm font-mono text-slate-700">
            {vectorString}
          </code>
        ) : (
          <VectorStringDisplay
            vectorString={vectorString}
            displayMode={displayMode}
            highlightSegments={highlightSegments}
          />
        )}

        {/* Copy button overlay on hover */}
        {showCopyButton && (
          <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              size="icon-sm"
              variant="outline"
              onClick={handleCopy}
              className="bg-white shadow-sm"
              aria-label="Copy vector string"
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-600" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
          </div>
        )}
      </div>

      {/* Error display */}
      {error && (
        <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 p-2 rounded">
          <AlertCircle className="h-4 w-4" />
          {error}
        </div>
      )}

      {/* Action buttons */}
      <div className="flex items-center gap-2">
        {showCopyButton && displayMode !== 'inline' && (
          <Button
            onClick={handleCopy}
            variant={copied ? 'success' : 'outline'}
            size="sm"
            leftIcon={copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          >
            {copied ? 'Copied!' : copyLabel}
          </Button>
        )}

        {showShareButton && (
          <ShareVectorDialog vectorString={vectorString} />
        )}

        {showExportButton && (
          <ExportVectorDialog vectorString={vectorString} parsed={parsed} />
        )}
      </div>
    </div>
  );
}

// ============================================================================
// Vector String Input Component
// ============================================================================

export interface VectorStringInputProps {
  /** Current vector string value */
  value: string;
  /** Callback when value changes */
  onChange: (value: string) => void;
  /** Callback when import is confirmed */
  onImport?: (vector: string) => void;
  /** Placeholder text */
  placeholder?: string;
  /** Error message */
  error?: string;
  /** Show validation status */
  showValidation?: boolean;
  /** Custom className */
  className?: string;
}

export function VectorStringInput({
  value,
  onChange,
  onImport,
  placeholder = 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
  error,
  showValidation = true,
  className,
}: VectorStringInputProps) {
  const [isValid, setIsValid] = useState<boolean | null>(null);
  const [validationMessage, setValidationMessage] = useState<string>('');

  useEffect(() => {
    if (!showValidation || !value) {
      setIsValid(null);
      setValidationMessage('');
      return;
    }

    const parsed = parseVectorString(value);

    if (!parsed) {
      setIsValid(false);
      setValidationMessage('Invalid CVSS vector string format');
      return;
    }

    // Check for required base metrics
    const requiredMetrics = ['AV', 'AC', 'PR', 'UI', 'S', 'C', 'I', 'A'];
    const hasAllRequired = requiredMetrics.every(m =>
      parsed.segments.some(s => s.name === m)
    );

    if (!hasAllRequired) {
      setIsValid(false);
      setValidationMessage('Missing required base metrics');
      return;
    }

    setIsValid(true);
    setValidationMessage(`Valid CVSS ${parsed.version} vector string`);
  }, [value, showValidation]);

  const handleImport = useCallback(() => {
    if (isValid && onImport) {
      onImport(value);
    }
  }, [isValid, value, onImport]);

  return (
    <div className={cn('space-y-3', className)}>
      <div className="relative">
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className={cn(
            'w-full px-4 py-3 pr-24 rounded-lg border font-mono text-sm',
            'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent',
            'transition-all duration-200',
            isValid === false && 'border-red-300 bg-red-50 focus:ring-red-500',
            isValid === true && 'border-green-300 bg-green-50 focus:ring-green-500',
            isValid === null && 'border-slate-300 bg-white'
          )}
          aria-invalid={isValid === false}
          aria-describedby={error || validationMessage ? 'vector-validation' : undefined}
        />
        {onImport && isValid && (
          <Button
            onClick={handleImport}
            size="sm"
            className="absolute right-2 top-1/2 -translate-y-1/2"
          >
            Import
          </Button>
        )}
      </div>

      {/* Validation message */}
      {(validationMessage || error) && (
        <div className={cn('flex items-center gap-2 text-sm', isValid === false || error ? 'text-red-600' : 'text-green-600')}>
          {isValid === false || error ? (
            <AlertCircle className="h-4 w-4" />
          ) : (
            <Check className="h-4 w-4" />
          )}
          {error || validationMessage}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Share Vector Dialog Component
// ============================================================================

interface ShareVectorDialogProps {
  vectorString: string;
}

function ShareVectorDialog({ vectorString }: ShareVectorDialogProps) {
  const [copied, setCopied] = useState(false);

  const handleCopyShareUrl = async () => {
    const url = `${window.location.origin}${window.location.pathname}?v=${encodeURIComponent(vectorString)}`;
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" leftIcon={<Link2 className="h-4 w-4" />}>
          Share
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Share CVSS Vector</DialogTitle>
          <DialogDescription>
            Copy this URL to share your CVSS vector with others
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="bg-slate-100 p-3 rounded-lg break-all text-sm font-mono">
            {`${window.location.origin}${window.location.pathname}?v=${encodeURIComponent(vectorString)}`}
          </div>
          <Button
            onClick={handleCopyShareUrl}
            variant={copied ? 'success' : 'default'}
            className="w-full"
            leftIcon={copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          >
            {copied ? 'URL Copied!' : 'Copy Share URL'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// ============================================================================
// Export Vector Dialog Component
// ============================================================================

interface ExportVectorDialogProps {
  vectorString: string;
  parsed: VectorParseResult | null;
}

function ExportVectorDialog({ vectorString, parsed }: ExportVectorDialogProps) {
  const handleExport = (format: 'json' | 'csv') => {
    if (format === 'json') {
      const data = {
        vectorString,
        version: parsed?.version,
        segments: parsed?.segments,
        exportedAt: new Date().toISOString(),
      };
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `cvss-vector-${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } else if (format === 'csv') {
      const csv = `Vector String,Version,Exported At\n"${vectorString}",${parsed?.version || ''},${new Date().toISOString()}`;
      const blob = new Blob([csv], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `cvss-vector-${Date.now()}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" leftIcon={<Download className="h-4 w-4" />}>
          Export
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Export CVSS Vector</DialogTitle>
          <DialogDescription>
            Choose a format to export your CVSS vector string
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <Button
            onClick={() => handleExport('json')}
            variant="outline"
            className="w-full justify-start"
            leftIcon={<Copy className="h-4 w-4 text-blue-600" />}
          >
            <div className="text-left">
              <div className="font-medium">JSON Format</div>
              <div className="text-xs text-slate-500">Export with all metadata</div>
            </div>
          </Button>
          <Button
            onClick={() => handleExport('csv')}
            variant="outline"
            className="w-full justify-start"
            leftIcon={<Copy className="h-4 w-4 text-green-600" />}
          >
            <div className="text-left">
              <div className="font-medium">CSV Format</div>
              <div className="text-xs text-slate-500">Export as spreadsheet data</div>
            </div>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// ============================================================================
// Vector Info Tooltip Component
// ============================================================================

interface VectorSegmentTooltipProps {
  segment: VectorSegment;
  metricInfo?: {
    name: string;
    description: string;
    values: Record<string, { label: string; description: string }>;
  };
}

export function VectorSegmentInfo({ segment, metricInfo }: VectorSegmentTooltipProps) {
  return (
    <div className="px-3 py-2 max-w-xs">
      <div className="font-semibold text-sm mb-1">{segment.segment}</div>
      {metricInfo && (
        <>
          <div className="text-xs text-slate-600 mb-2">{metricInfo.description}</div>
          <div className="text-xs text-slate-700">
            <span className="font-medium">Current:</span>{' '}
            {metricInfo.values[segment.value]?.label || segment.value}
          </div>
          {metricInfo.values[segment.value]?.description && (
            <div className="text-xs text-slate-500 mt-1">
              {metricInfo.values[segment.value].description}
            </div>
          )}
        </>
      )}
    </div>
  );
}
