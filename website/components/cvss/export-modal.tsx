'use client';

/**
 * CVSS Export Modal Component
 * Modal dialog for export format selection (JSON, CSV, URL)
 */

import React, { useState, useCallback } from 'react';
import {
  Download,
  FileJson,
  FileSpreadsheet,
  Link as Link2,
  Check,
  Copy,
  Loader2,
  Info,
  X,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Card, CardContent } from '@/components/ui/card';
import type { CVSSVersion, CVSSExportFormat, CVSSScoreBreakdown } from '@/lib/types';

// ============================================================================
// Types
// ============================================================================

export interface CVSSExportData {
  version: CVSSVersion;
  vectorString: string;
  baseScore: number;
  severity: string;
  metrics: Record<string, string>;
  scoreBreakdown: CVSSScoreBreakdown;
  exportedAt: string;
}

export interface ExportModalProps {
  /** Whether the modal is open */
  open?: boolean;
  /** Callback when modal open state changes */
  onOpenChange?: (open: boolean) => void;
  /** CVSS data to export */
  exportData: CVSSExportData;
  /** Available export formats */
  formats?: CVSSExportFormat[];
  /** Custom trigger element */
  trigger?: React.ReactNode;
  /** Callback after successful export */
  onExport?: (format: CVSSExportFormat, data: any) => void;
}

export interface ExportFormatOption {
  value: CVSSExportFormat;
  label: string;
  description: string;
  icon: React.ReactNode;
  mimeType: string;
  fileExtension: string;
}

// ============================================================================
// Export Format Options
// ============================================================================

const exportFormats: Record<CVSSExportFormat, ExportFormatOption> = {
  json: {
    value: 'json',
    label: 'JSON',
    description: 'Structured data format with full metric breakdown and metadata',
    icon: <FileJson className="h-6 w-6" />,
    mimeType: 'application/json',
    fileExtension: 'json',
  },
  csv: {
    value: 'csv',
    label: 'CSV',
    description: 'Comma-separated values for spreadsheet applications',
    icon: <FileSpreadsheet className="h-6 w-6" />,
    mimeType: 'text/csv',
    fileExtension: 'csv',
  },
  url: {
    value: 'url',
    label: 'Share URL',
    description: 'Copy a shareable URL to clipboard',
    icon: <Link2 className="h-6 w-6" />,
    mimeType: '',
    fileExtension: '',
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

function generateShareUrl(vectorString: string): string {
  const baseUrl = window.location.origin + window.location.pathname;
  return `${baseUrl}?v=${encodeURIComponent(vectorString)}`;
}

function flattenMetrics(metrics: Record<string, any>): Record<string, string> {
  const flattened: Record<string, string> = {};

  for (const [key, value] of Object.entries(metrics)) {
    if (value && typeof value === 'object') {
      const nested = flattenMetrics(value);
      for (const [nestedKey, nestedValue] of Object.entries(nested)) {
        flattened[`${key}.${nestedKey}`] = nestedValue;
      }
    } else if (value !== undefined && value !== null && value !== 'X') {
      flattened[key] = String(value);
    }
  }

  return flattened;
}

// ============================================================================
// Export Format Card Component
// ============================================================================

interface FormatCardProps {
  option: ExportFormatOption;
  selected: boolean;
  onSelect: () => void;
  disabled?: boolean;
}

function FormatCard({ option, selected, onSelect, disabled }: FormatCardProps) {
  return (
    <Card
      className={cn(
        'cursor-pointer transition-all duration-200 hover:shadow-md',
        selected
          ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-500 ring-offset-2'
          : 'border-slate-200 hover:border-blue-300',
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      onClick={!disabled ? onSelect : undefined}
    >
      <CardContent className="p-4">
        <div className="flex items-start gap-3">
          <div className={cn(
            'p-2 rounded-lg',
            selected ? 'bg-blue-100 text-blue-600' : 'bg-slate-100 text-slate-600'
          )}>
            {option.icon}
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-sm">{option.label}</h3>
              {selected && (
                <Check className="h-4 w-4 text-blue-600" />
              )}
            </div>
            <p className="text-xs text-slate-500 mt-1">{option.description}</p>
            {option.fileExtension && (
              <div className="mt-2 inline-flex items-center gap-1 text-xs text-slate-400">
                <span>.{option.fileExtension}</span>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Export Preview Component
// ============================================================================

interface ExportPreviewProps {
  data: CVSSExportData;
  format: CVSSExportFormat;
}

function ExportPreview({ data, format }: ExportPreviewProps) {
  const [copied, setCopied] = useState(false);

  const getPreviewContent = (): string => {
    switch (format) {
      case 'json':
        return JSON.stringify(data, null, 2);

      case 'csv':
        const flatMetrics = flattenMetrics(data.metrics);
        const headers = ['Version', 'Vector String', 'Base Score', 'Severity', ...Object.keys(flatMetrics), 'Exported At'];
        const values = [
          data.version,
          `"${data.vectorString}"`,
          data.baseScore.toFixed(1),
          data.severity,
          ...Object.values(flatMetrics),
          data.exportedAt,
        ];
        return [headers.join(','), values.join(',')].join('\n');

      case 'url':
        return generateShareUrl(data.vectorString);

      default:
        return '';
    }
  };

  const handleCopy = useCallback(async () => {
    const content = getPreviewContent();
    await navigator.clipboard.writeText(content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [format]);

  const previewContent = getPreviewContent();

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium">Preview</h4>
        <Button
          size="sm"
          variant="ghost"
          onClick={handleCopy}
          leftIcon={copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
        >
          {copied ? 'Copied!' : 'Copy'}
        </Button>
      </div>
      <div className="bg-slate-900 rounded-lg p-3 max-h-48 overflow-auto">
        <pre className="text-xs text-green-400 font-mono whitespace-pre-wrap break-all">
          {previewContent}
        </pre>
      </div>
    </div>
  );
}

// ============================================================================
// Main Export Modal Component
// ============================================================================

export function ExportModal({
  open,
  onOpenChange,
  exportData,
  formats = ['json', 'csv', 'url'],
  trigger,
  onExport,
}: ExportModalProps) {
  const [selectedFormat, setSelectedFormat] = useState<CVSSExportFormat>('json');
  const [isExporting, setIsExporting] = useState(false);
  const [exportSuccess, setExportSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleExport = useCallback(async () => {
    setIsExporting(true);
    setError(null);
    setExportSuccess(false);

    try {
      const option = exportFormats[selectedFormat];

      switch (selectedFormat) {
        case 'json': {
          const blob = new Blob([JSON.stringify(exportData, null, 2)], {
            type: option.mimeType,
          });
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = `cvss-${exportData.version}-${Date.now()}.${option.fileExtension}`;
          a.click();
          URL.revokeObjectURL(url);
          break;
        }

        case 'csv': {
          const flatMetrics = flattenMetrics(exportData.metrics);
          const headers = ['Version', 'Vector String', 'Base Score', 'Severity', 'Export Date'];
          const metricHeaders = Object.keys(flatMetrics);
          const allHeaders = [...headers, ...metricHeaders];

          const values = [
            exportData.version,
            `"${exportData.vectorString}"`,
            exportData.baseScore.toFixed(1),
            exportData.severity,
            exportData.exportedAt,
            ...metricHeaders.map(h => flatMetrics[h] || ''),
          ];

          const csv = [
            allHeaders.join(','),
            values.join(','),
          ].join('\n');

          const blob = new Blob([csv], { type: option.mimeType });
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = `cvss-${exportData.version}-${Date.now()}.${option.fileExtension}`;
          a.click();
          URL.revokeObjectURL(url);
          break;
        }

        case 'url': {
          const shareUrl = generateShareUrl(exportData.vectorString);
          await navigator.clipboard.writeText(shareUrl);
          break;
        }
      }

      setExportSuccess(true);
      onExport?.(selectedFormat, exportData);

      setTimeout(() => {
        setExportSuccess(false);
        if (selectedFormat === 'url') {
          onOpenChange?.(false);
        }
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed');
      console.error('Export error:', err);
    } finally {
      setIsExporting(false);
    }
  }, [selectedFormat, exportData, onExport, onOpenChange]);

  const availableFormats = formats.map(f => exportFormats[f]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {trigger && (
        <DialogTrigger asChild>
          {trigger}
        </DialogTrigger>
      )}
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Export CVSS Data</DialogTitle>
          <DialogDescription>
            Choose an export format for your CVSS {exportData.version} calculation
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Format Selection */}
          <div>
            <h3 className="text-sm font-semibold mb-3">Select Format</h3>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {availableFormats.map((format) => (
                <FormatCard
                  key={format.value}
                  option={format}
                  selected={selectedFormat === format.value}
                  onSelect={() => setSelectedFormat(format.value)}
                  disabled={isExporting}
                />
              ))}
            </div>
          </div>

          {/* Preview */}
          <ExportPreview data={exportData} format={selectedFormat} />

          {/* Info Notice */}
          <div className="flex items-start gap-2 p-3 bg-blue-50 rounded-lg border border-blue-200">
            <Info className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
            <p className="text-xs text-blue-700">
              {selectedFormat === 'url'
                ? 'The URL contains the vector string and can be shared to recreate this CVSS calculation.'
                : `Exported files include all metric values and can be imported later for analysis.`}
            </p>
          </div>

          {/* Error Display */}
          {error && (
            <div className="flex items-center gap-2 p-3 bg-red-50 rounded-lg border border-red-200">
              <X className="h-4 w-4 text-red-600" />
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange?.(false)}
            disabled={isExporting}
          >
            Cancel
          </Button>
          <Button
            onClick={handleExport}
            disabled={isExporting}
            leftIcon={isExporting ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
            rightIcon={!isExporting ? <Download className="h-4 w-4" /> : null}
          >
            {isExporting
              ? 'Exporting...'
              : exportSuccess
              ? 'Exported!'
              : `Export as ${exportFormats[selectedFormat].label}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ============================================================================
// Quick Export Button Component
// ============================================================================

export interface QuickExportButtonProps {
  exportData: CVSSExportData;
  format: CVSSExportFormat;
  icon?: React.ReactNode;
  label?: string;
  onExportSuccess?: () => void;
}

export function QuickExportButton({
  exportData,
  format,
  icon,
  label,
  onExportSuccess,
}: QuickExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);

  const handleQuickExport = useCallback(async () => {
    setIsExporting(true);
    try {
      const option = exportFormats[format];

      if (format === 'url') {
        const url = generateShareUrl(exportData.vectorString);
        await navigator.clipboard.writeText(url);
      } else {
        // For json and csv, trigger the full modal or direct download
        const modal = document.querySelector('[data-export-modal]') as HTMLElement;
        if (modal) {
          modal.click();
        }
      }

      setShowSuccess(true);
      onExportSuccess?.();

      setTimeout(() => setShowSuccess(false), 2000);
    } catch (err) {
      console.error('Quick export failed:', err);
    } finally {
      setIsExporting(false);
    }
  }, [exportData, format, onExportSuccess]);

  const option = exportFormats[format];

  return (
    <Button
      onClick={handleQuickExport}
      disabled={isExporting}
      variant={showSuccess ? 'success' : 'outline'}
      size="sm"
      leftIcon={isExporting ? <Loader2 className="h-4 w-4 animate-spin" /> : showSuccess ? <Check className="h-4 w-4" /> : icon || option.icon}
    >
      {label || (showSuccess ? 'Done!' : option.label)}
    </Button>
  );
}

// ============================================================================
// Export All Button Component
// ============================================================================

export interface ExportAllButtonProps {
  exportData: CVSSExportData;
  onExport?: (format: CVSSExportFormat, data: any) => void;
}

export function ExportAllButton({ exportData, onExport }: ExportAllButtonProps) {
  return (
    <ExportModal
      trigger={
        <Button variant="default" leftIcon={<Download className="h-4 w-4" />}>
          Export CVSS Data
        </Button>
      }
      exportData={exportData}
      formats={['json', 'csv', 'url']}
      onExport={onExport}
    />
  );
}
