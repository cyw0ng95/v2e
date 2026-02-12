'use client';

/**
 * CVSS Desktop View Component
 * Desktop-optimized layout for the CVSS calculator
 */

import React from 'react';
import { Settings, Info, SlidersHorizontal, Maximize2, Minimize2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import type { CVSSVersion, CVSS3Metrics, CVSS4Metrics, CVSSScoreBreakdown } from '@/lib/types';
import { MetricCard, MetricGroup } from './metric-card';
import { ScoreBreakdown } from './score-display';
import { SeverityGauge, SemiCircularGauge } from './severity-gauge';
import { VectorString, VectorStringInput } from './vector-string';
import { ExportModal } from './export-modal';

// ============================================================================
// Types
// ============================================================================

export interface DesktopViewProps {
  version: CVSSVersion;
  metrics: CVSS3Metrics | CVSS4Metrics;
  scores: CVSSScoreBreakdown | null;
  onMetricChange: (metric: string, value: string) => void;
  onVersionChange: (version: CVSSVersion) => void;
  onReset: () => void;
  onImport?: (vector: string) => void;
  showTemporal?: boolean;
  showEnvironmental?: boolean;
  toggleTemporal?: () => void;
  toggleEnvironmental?: () => void;
  isCompact?: boolean;
  onToggleCompact?: () => void;
}

// ============================================================================
// Desktop Header Component
// ============================================================================

interface DesktopHeaderProps {
  version: CVSSVersion;
  onVersionChange: (version: CVSSVersion) => void;
  onReset: () => void;
  isCompact?: boolean;
  onToggleCompact?: () => void;
}

function DesktopHeader({
  version,
  onVersionChange,
  onReset,
  isCompact,
  onToggleCompact,
}: DesktopHeaderProps) {
  const versions: CVSSVersion[] = ['3.0', '3.1', '4.0'];

  return (
    <header className="sticky top-0 z-30 bg-white/95 backdrop-blur-sm border-b border-slate-200">
      <div className="max-w-7xl mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          {/* Title and Version Selector */}
          <div className="flex items-center gap-6">
            <div>
              <h1 className="text-xl font-bold text-slate-900">CVSS Calculator</h1>
              <p className="text-sm text-slate-500">Common Vulnerability Scoring System</p>
            </div>

            <Separator orientation="vertical" className="h-10" />

            {/* Version Tabs */}
            <div className="flex items-center bg-slate-100 rounded-lg p-1">
              {versions.map((v) => (
                <button
                  key={v}
                  onClick={() => onVersionChange(v)}
                  className={cn(
                    'px-4 py-2 rounded-md text-sm font-medium transition-all',
                    version === v
                      ? 'bg-white text-blue-600 shadow-sm'
                      : 'text-slate-600 hover:text-slate-900'
                  )}
                  aria-pressed={version === v}
                >
                  {v}
                </button>
              ))}
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2">
            {/* Toggle View Mode */}
            {onToggleCompact && (
              <Button
                size="sm"
                variant="ghost"
                onClick={onToggleCompact}
                leftIcon={isCompact ? <Maximize2 className="h-4 w-4" /> : <Minimize2 className="h-4 w-4" />}
                aria-label={isCompact ? 'Expand view' : 'Compact view'}
              >
                {isCompact ? 'Expand' : 'Compact'}
              </Button>
            )}

            {/* Reset */}
            <Button
              size="sm"
              variant="outline"
              onClick={onReset}
              leftIcon={<Settings className="h-4 w-4" />}
            >
              Reset
            </Button>

            {/* Help */}
            <Button
              size="sm"
              variant="ghost"
              leftIcon={<Info className="h-4 w-4" />}
              aria-label="Show help guide"
            >
              Guide
            </Button>
          </div>
        </div>
      </div>
    </header>
  );
}

// ============================================================================
// Desktop Sidebar Component
// ============================================================================

interface DesktopSidebarProps {
  scores: CVSSScoreBreakdown;
  version: CVSSVersion;
  vectorString: string;
  showEnvironmental?: boolean;
  toggleEnvironmental?: () => void;
  onImport?: (vector: string) => void;
}

function DesktopSidebar({
  scores,
  version,
  vectorString,
  showEnvironmental,
  toggleEnvironmental,
  onImport,
}: DesktopSidebarProps) {
  return (
    <aside className="w-80 space-y-6">
      {/* Score Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Score Results</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Semi-Circular Gauge */}
          <SemiCircularGauge
            score={scores.finalScore}
            severity={scores.finalSeverity}
            size="md"
            showScore={false}
          />

          {/* Numeric Score Display */}
          <div className="text-center">
            <div className="text-4xl font-bold text-slate-900">
              {scores.finalScore.toFixed(1)}
            </div>
            <div className={cn(
              'inline-block px-3 py-1 rounded-full text-sm font-medium mt-2',
              scores.finalSeverity === 'CRITICAL' && 'bg-purple-100 text-purple-700',
              scores.finalSeverity === 'HIGH' && 'bg-red-100 text-red-700',
              scores.finalSeverity === 'MEDIUM' && 'bg-orange-100 text-orange-700',
              scores.finalSeverity === 'LOW' && 'bg-yellow-100 text-yellow-700',
              scores.finalSeverity === 'NONE' && 'bg-gray-100 text-gray-700'
            )}>
              {scores.finalSeverity}
            </div>
          </div>

          {/* Score Breakdown */}
          <Separator />
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-slate-500">Base Score</span>
              <span className="font-medium">{scores.baseScore.toFixed(1)}</span>
            </div>
            {(scores as any).temporalScore !== undefined && (
              <div className="flex justify-between">
                <span className="text-slate-500">Temporal</span>
                <span className="font-medium">
                  {(scores as any).temporalScore.toFixed(1)}
                </span>
              </div>
            )}
            {(scores as any).environmentalScore !== undefined && (
              <div className="flex justify-between">
                <span className="text-slate-500">Environmental</span>
                <span className="font-medium">
                  {(scores as any).environmentalScore.toFixed(1)}
                </span>
              </div>
            )}
            {(scores as any).threatScore !== undefined && (
              <div className="flex justify-between">
                <span className="text-slate-500">Threat</span>
                <span className="font-medium">
                  {(scores as any).threatScore.toFixed(1)}
                </span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Vector String Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Vector String</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <VectorString
            vectorString={vectorString}
            displayMode="full"
            highlightSegments={true}
            showCopyButton={true}
            showExportButton={true}
            showShareButton={true}
          />

          {/* Import Vector */}
          {onImport && (
            <>
              <Separator />
              <div>
                <label className="text-sm font-medium text-slate-700 mb-2 block">
                  Import Vector String
                </label>
                <VectorStringInput
                  value=""
                  onChange={() => {}}
                  onImport={onImport}
                  placeholder="Paste CVSS vector string..."
                  showValidation={true}
                />
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Environmental Toggle */}
      {toggleEnvironmental && version !== '4.0' && (
        <Card>
          <CardContent className="p-4">
            <button
              onClick={toggleEnvironmental}
              className={cn(
                'w-full p-3 rounded-lg border-2 transition-all duration-200',
                'flex items-center justify-between',
                showEnvironmental
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-slate-200 hover:border-blue-300'
              )}
            >
              <div className="text-left">
                <div className="font-medium text-sm">Environmental Metrics</div>
                <div className="text-xs text-slate-500 mt-0.5">
                  Customize for your environment
                </div>
              </div>
              <SlidersHorizontal className={cn(
                'h-5 w-5 transition-colors',
                showEnvironmental ? 'text-blue-600' : 'text-slate-400'
              )} />
            </button>
          </CardContent>
        </Card>
      )}

      {/* Export Card */}
      <Card>
        <CardContent className="p-4">
          <ExportModal
            trigger={
              <Button variant="default" className="w-full" leftIcon={<SlidersHorizontal className="h-4 w-4" />}>
                Export CVSS Data
              </Button>
            }
            exportData={{
              version,
              vectorString,
              baseScore: scores.baseScore,
              severity: scores.finalSeverity,
              metrics: scores as any,
              scoreBreakdown: scores,
              exportedAt: new Date().toISOString(),
            }}
            formats={['json', 'csv', 'url']}
          />
        </CardContent>
      </Card>
    </aside>
  );
}

// ============================================================================
// Desktop Main Content Component
// ============================================================================

interface DesktopMainContentProps {
  version: CVSSVersion;
  metrics: CVSS3Metrics | CVSS4Metrics;
  onMetricChange: (metric: string, value: string) => void;
  showTemporal?: boolean;
  showEnvironmental?: boolean;
  toggleTemporal?: () => void;
  toggleEnvironmental?: () => void;
}

function DesktopMainContent({
  version,
  metrics,
  onMetricChange,
  showTemporal,
  showEnvironmental,
  toggleTemporal,
  toggleEnvironmental,
}: DesktopMainContentProps) {
  const isV3 = version === '3.0' || version === '3.1';

  return (
    <main className="flex-1 space-y-6">
      {/* Base Metrics Group */}
      <MetricGroup
        title="Base Score Metrics"
        description="Select the characteristics of this vulnerability"
      >
        <div className={cn(
          'grid gap-4',
          isV3 ? 'md:grid-cols-2 xl:grid-cols-4' : 'md:grid-cols-3'
        )}>
          {isV3 ? (
            <>
              <MetricCard
                name="Attack Vector"
                abbreviation="AV"
                value={(metrics as any).AV}
                onChange={(v) => onMetricChange('AV', v)}
                options={[
                  { value: 'N', label: 'Network', description: 'Network exploitable', abbreviation: 'N' },
                  { value: 'A', label: 'Adjacent', description: 'Adjacent network', abbreviation: 'A' },
                  { value: 'L', label: 'Local', description: 'Local access', abbreviation: 'L' },
                  { value: 'P', label: 'Physical', description: 'Physical access', abbreviation: 'P' },
                ]}
                description="How the vulnerability can be exploited"
                orientation="grid"
                gridCols={2}
                size="md"
              />
              <MetricCard
                name="Attack Complexity"
                abbreviation="AC"
                value={(metrics as any).AC}
                onChange={(v) => onMetricChange('AC', v)}
                options={[
                  { value: 'L', label: 'Low', description: 'Specialized access', abbreviation: 'L' },
                  { value: 'H', label: 'High', description: 'Specialized conditions', abbreviation: 'H' },
                ]}
                description="Complexity of the attack"
                orientation="vertical"
                size="md"
              />
              <MetricCard
                name="Privileges Required"
                abbreviation="PR"
                value={(metrics as any).PR}
                onChange={(v) => onMetricChange('PR', v)}
                options={[
                  { value: 'N', label: 'None', description: 'No privileges required', abbreviation: 'N', recommended: true },
                  { value: 'L', label: 'Low', description: 'Basic user privileges', abbreviation: 'L' },
                  { value: 'H', label: 'High', description: 'Admin privileges', abbreviation: 'H' },
                ]}
                description="Privileges needed before attack"
                orientation="grid"
                gridCols={3}
                size="md"
              />
              <MetricCard
                name="User Interaction"
                abbreviation="UI"
                value={(metrics as any).UI}
                onChange={(v) => onMetricChange('UI', v)}
                options={[
                  { value: 'N', label: 'None', description: 'No interaction required', abbreviation: 'N', recommended: true },
                  { value: 'R', label: 'Required', description: 'User participation needed', abbreviation: 'R' },
                ]}
                description="User interaction required"
                orientation="vertical"
                size="md"
              />
              <MetricCard
                name="Scope"
                abbreviation="S"
                value={(metrics as any).S}
                onChange={(v) => onMetricChange('S', v)}
                options={[
                  { value: 'U', label: 'Unchanged', description: 'Only vulnerable component affected', abbreviation: 'U', recommended: true },
                  { value: 'C', label: 'Changed', description: 'Affects other components', abbreviation: 'C' },
                ]}
                description="Scope of impact"
                orientation="vertical"
                size="md"
              />
              <MetricCard
                name="Confidentiality"
                abbreviation="C"
                value={(metrics as any).C}
                onChange={(v) => onMetricChange('C', v)}
                options={[
                  { value: 'H', label: 'High', description: 'Total compromise', abbreviation: 'H' },
                  { value: 'L', label: 'Low', description: 'Partial compromise', abbreviation: 'L' },
                  { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N', recommended: true },
                ]}
                description="Impact on data confidentiality"
                orientation="grid"
                gridCols={3}
                size="md"
              />
              <MetricCard
                name="Integrity"
                abbreviation="I"
                value={(metrics as any).I}
                onChange={(v) => onMetricChange('I', v)}
                options={[
                  { value: 'H', label: 'High', description: 'Total compromise', abbreviation: 'H' },
                  { value: 'L', label: 'Low', description: 'Partial compromise', abbreviation: 'L' },
                  { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N', recommended: true },
                ]}
                description="Impact on data integrity"
                orientation="grid"
                gridCols={3}
                size="md"
              />
              <MetricCard
                name="Availability"
                abbreviation="A"
                value={(metrics as any).A}
                onChange={(v) => onMetricChange('A', v)}
                options={[
                  { value: 'H', label: 'High', description: 'Total disruption', abbreviation: 'H' },
                  { value: 'L', label: 'Low', description: 'Partial disruption', abbreviation: 'L' },
                  { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N', recommended: true },
                ]}
                description="Impact on availability"
                orientation="grid"
                gridCols={3}
                size="md"
              />
            </>
          ) : (
            <>
              {/* CVSS v4.0 base metrics */}
              <MetricCard
                name="Attack Vector"
                abbreviation="AV"
                value={(metrics as any).AV}
                onChange={(v) => onMetricChange('AV', v)}
                options={[
                  { value: 'N', label: 'Network', description: 'Network exploitable', abbreviation: 'N' },
                  { value: 'A', label: 'Adjacent', description: 'Adjacent network', abbreviation: 'A' },
                  { value: 'L', label: 'Local', description: 'Local access', abbreviation: 'L' },
                  { value: 'P', label: 'Physical', description: 'Physical access', abbreviation: 'P' },
                ]}
                orientation="grid"
                gridCols={2}
                size="md"
              />
              <MetricCard
                name="Attack Complexity"
                abbreviation="AC"
                value={(metrics as any).AC}
                onChange={(v) => onMetricChange('AC', v)}
                options={[
                  { value: 'L', label: 'Low', description: 'Specialized access', abbreviation: 'L' },
                  { value: 'H', label: 'High', description: 'Specialized conditions', abbreviation: 'H' },
                ]}
                orientation="vertical"
                size="md"
              />
              <MetricCard
                name="Attack Requirements"
                abbreviation="AT"
                value={(metrics as any).AT}
                onChange={(v) => onMetricChange('AT', v)}
                options={[
                  { value: 'N', label: 'None', description: 'No requirements', abbreviation: 'N' },
                  { value: 'P', label: 'Present', description: 'Attack present', abbreviation: 'P' },
                  { value: 'R', label: 'Required', description: 'Attack required', abbreviation: 'R' },
                ]}
                orientation="grid"
                gridCols={3}
                size="md"
              />
            </>
          )}
        </div>
      </MetricGroup>

      {/* Temporal Metrics (CVSS v3.x only) */}
      {isV3 && (
        <MetricGroup
          title="Temporal Metrics"
          description="Adjust score based on current exploit status"
        >
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-slate-600">
              Temporal metrics adjust the base score based on factors that change over time
            </p>
            <Button
              size="sm"
              variant={showTemporal ? 'default' : 'outline'}
              onClick={toggleTemporal}
            >
              {showTemporal ? 'Hide' : 'Show'} Metrics
            </Button>
          </div>

          {showTemporal && (metrics as any).temporal && (
            <div className="grid md:grid-cols-3 gap-4">
              <MetricCard
                name="Exploit Maturity"
                abbreviation="E"
                value={(metrics as any).temporal?.E || 'X'}
                onChange={(v) => onMetricChange('temporal.E', v)}
                options={[
                  { value: 'X', label: 'Not Defined', description: 'Not assigned', abbreviation: 'X' },
                  { value: 'U', label: 'Unproven', description: 'No PoC exists', abbreviation: 'U' },
                  { value: 'P', label: 'POC', description: 'Proof of concept', abbreviation: 'P' },
                  { value: 'F', label: 'Functional', description: 'Exploitable', abbreviation: 'F' },
                  { value: 'H', label: 'High', description: 'Reliable exploit', abbreviation: 'H' },
                ]}
                orientation="vertical"
                size="sm"
              />
              <MetricCard
                name="Remediation Level"
                abbreviation="RL"
                value={(metrics as any).temporal?.RL || 'X'}
                onChange={(v) => onMetricChange('temporal.RL', v)}
                options={[
                  { value: 'X', label: 'Not Defined', description: 'Not assigned', abbreviation: 'X' },
                  { value: 'U', label: 'Unavailable', description: 'No fix available', abbreviation: 'U' },
                  { value: 'O', label: 'Workaround', description: 'Temporary fix', abbreviation: 'O' },
                  { value: 'T', label: 'Temporary', description: 'Partial fix', abbreviation: 'T' },
                  { value: 'W', label: 'Official', description: 'Complete fix', abbreviation: 'W' },
                ]}
                orientation="vertical"
                size="sm"
              />
              <MetricCard
                name="Report Confidence"
                abbreviation="RC"
                value={(metrics as any).temporal?.RC || 'X'}
                onChange={(v) => onMetricChange('temporal.RC', v)}
                options={[
                  { value: 'X', label: 'Not Defined', description: 'Not assigned', abbreviation: 'X' },
                  { value: 'U', label: 'Unknown', description: 'Unverified', abbreviation: 'U' },
                  { value: 'R', label: 'Confirmed', description: 'Verified report', abbreviation: 'R' },
                  { value: 'C', label: 'Reasonable', description: 'Some confidence', abbreviation: 'C' },
                ]}
                orientation="vertical"
                size="sm"
              />
            </div>
          )}
        </MetricGroup>
      )}
    </main>
  );
}

// ============================================================================
// Main Desktop View Component
// ============================================================================

export function DesktopView(props: DesktopViewProps) {
  if (!props.scores) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-pulse text-2xl text-slate-400 mb-2">Calculating...</div>
          <div className="text-sm text-slate-500">Computing CVSS score</div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <DesktopHeader
        version={props.version}
        onVersionChange={props.onVersionChange}
        onReset={props.onReset}
        isCompact={props.isCompact}
        onToggleCompact={props.onToggleCompact}
      />

      {/* Main Content Area */}
      <div className="max-w-7xl mx-auto px-6 py-8">
        <div className="flex gap-6">
          {/* Main Content */}
          <DesktopMainContent
            version={props.version}
            metrics={props.metrics}
            onMetricChange={props.onMetricChange}
            showTemporal={props.showTemporal}
            showEnvironmental={props.showEnvironmental}
            toggleTemporal={props.toggleTemporal}
            toggleEnvironmental={props.toggleEnvironmental}
          />

          {/* Sidebar */}
          <DesktopSidebar
            scores={props.scores}
            version={props.version}
            vectorString={props.scores.vectorString || ''}
            showEnvironmental={props.showEnvironmental}
            toggleEnvironmental={props.toggleEnvironmental}
            onImport={props.onImport}
          />
        </div>
      </div>
    </div>
  );
}
