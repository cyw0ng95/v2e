'use client';

/**
 * CVSS Mobile View Component
 * Mobile-optimized layout for the CVSS calculator
 */

import React, { useState } from 'react';
import { ChevronDown, ChevronUp, Info, SlidersHorizontal } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Card, CardContent } from '@/components/ui/card';
import type { CVSSVersion, CVSS3Metrics, CVSS4Metrics, CVSS3ScoreBreakdown, CVSS4ScoreBreakdown } from '@/lib/types';
import { MetricCard, CompactMetricCard } from './metric-card';
import { ScoreBadge, ScoreDisplay } from './score-display';
import { SeverityGauge } from './severity-gauge';
import { VectorString } from './vector-string';

// ============================================================================
// Types
// ============================================================================

export interface MobileViewProps {
  version: CVSSVersion;
  metrics: CVSS3Metrics | CVSS4Metrics;
  scores: CVSS3ScoreBreakdown | CVSS4ScoreBreakdown | null;
  onMetricChange: (metric: string, value: string) => void;
  onVersionChange: (version: CVSSVersion) => void;
  onReset: () => void;
  showTemporal?: boolean;
  showEnvironmental?: boolean;
  toggleTemporal?: () => void;
  toggleEnvironmental?: () => void;
}

// ============================================================================
// Mobile Header Component
// ============================================================================

interface MobileHeaderProps {
  version: CVSSVersion;
  scores: CVSS3ScoreBreakdown | CVSS4ScoreBreakdown | null;
  onVersionChange: (version: CVSSVersion) => void;
  onReset: () => void;
}

function MobileHeader({ version, scores, onVersionChange, onReset }: MobileHeaderProps) {
  const [showVersionMenu, setShowVersionMenu] = useState(false);

  const versions: CVSSVersion[] = ['3.0', '3.1', '4.0'];

  return (
    <header className="sticky top-0 z-30 bg-white border-b border-slate-200 px-4 py-3">
      <div className="flex items-center justify-between">
        {/* Score Badge */}
        <div className="flex items-center gap-2">
          {scores && (
            <ScoreBadge
              severity={scores.finalSeverity || scores.baseSeverity}
              score={scores.baseScore}
              size="sm"
            />
          )}
        </div>

        {/* Version Selector */}
        <div className="relative">
          <Button
            size="sm"
            variant="outline"
            onClick={() => setShowVersionMenu(!showVersionMenu)}
          >
            CVSS {version}
            <ChevronDown className={cn('h-4 w-4 ml-1 transition-transform', showVersionMenu && 'rotate-180')} />
          </Button>

          {showVersionMenu && (
            <div className="absolute top-full right-0 mt-1 bg-white rounded-lg border shadow-lg overflow-hidden z-50">
              {versions.map((v) => (
                <button
                  key={v}
                  onClick={() => {
                    onVersionChange(v);
                    setShowVersionMenu(false);
                  }}
                  className={cn(
                    'px-4 py-2 text-left hover:bg-slate-50 transition-colors',
                    'min-w-[80px]',
                    version === v && 'bg-blue-50 text-blue-600'
                  )}
                >
                  CVSS {v}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Reset Button */}
        <button
          onClick={onReset}
          className="p-2 text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors"
          aria-label="Reset metrics"
        >
          <SlidersHorizontal className="h-5 w-5" />
        </button>
      </div>
    </header>
  );
}

// ============================================================================
// Mobile Score Panel Component
// ============================================================================

interface MobileScorePanelProps {
  scores: CVSS3ScoreBreakdown | CVSS4ScoreBreakdown;
  isExpanded?: boolean;
}

function MobileScorePanel({ scores, isExpanded = false }: MobileScorePanelProps) {
  const [expanded, setExpanded] = useState(isExpanded);

  return (
    <Card className="mx-4 mt-4">
      <CardContent className="p-4">
        <button
          onClick={() => setExpanded(!expanded)}
          className="w-full flex items-center justify-between"
        >
          <span className="font-semibold text-sm">Score Details</span>
          {expanded ? (
            <ChevronUp className="h-4 w-4 text-slate-600" />
          ) : (
            <ChevronDown className="h-4 w-4 text-slate-600" />
          )}
        </button>

        {expanded && (
          <div className="mt-4 space-y-3">
            {/* Severity Gauge */}
            <SeverityGauge
              score={scores.baseScore}
              severity={scores.finalSeverity || scores.baseSeverity}
              size="sm"
              showLabels={true}
              showScore={true}
            />

            {/* Score Breakdown */}
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div className="bg-slate-50 rounded-lg p-2 text-center">
                <div className="text-slate-500 text-xs">Base</div>
                <div className="font-bold text-slate-900">{scores.baseScore.toFixed(1)}</div>
              </div>
              {(scores as any).temporalScore !== undefined && (
                <div className="bg-slate-50 rounded-lg p-2 text-center">
                  <div className="text-slate-500 text-xs">Temporal</div>
                  <div className="font-bold text-slate-900">
                    {(scores as any).temporalScore.toFixed(1)}
                  </div>
                </div>
              )}
              {(scores as any).environmentalScore !== undefined && (
                <div className="bg-slate-50 rounded-lg p-2 text-center">
                  <div className="text-slate-500 text-xs">Environmental</div>
                  <div className="font-bold text-slate-900">
                    {(scores as any).environmentalScore.toFixed(1)}
                  </div>
                </div>
              )}
              {(scores as any).threatScore !== undefined && (
                <div className="bg-slate-50 rounded-lg p-2 text-center">
                  <div className="text-slate-500 text-xs">Threat</div>
                  <div className="font-bold text-slate-900">
                    {(scores as any).threatScore.toFixed(1)}
                  </div>
                </div>
              )}
              <div className="col-span-2 bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg p-2 text-center">
                <div className="text-slate-500 text-xs">Final Score</div>
                <div className="font-bold text-lg text-blue-700">{scores.baseScore.toFixed(1)}</div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// ============================================================================
// Mobile Metrics Accordion Component
// ============================================================================

interface MobileMetricsAccordionProps {
  version: CVSSVersion;
  metrics: CVSS3Metrics | CVSS4Metrics;
  onMetricChange: (metric: string, value: string) => void;
  showTemporal?: boolean;
  showEnvironmental?: boolean;
  toggleTemporal?: () => void;
  toggleEnvironmental?: () => void;
}

function MobileMetricsAccordion({
  version,
  metrics,
  onMetricChange,
  showTemporal,
  showEnvironmental,
  toggleTemporal,
  toggleEnvironmental,
}: MobileMetricsAccordionProps) {
  const isV3 = version === '3.0' || version === '3.1';

  return (
    <div className="px-4 py-4">
      <Accordion type="multiple" defaultValue={['base']}>
        {/* Base Metrics */}
        <AccordionItem value="base">
          <AccordionTrigger className="px-4 py-3 hover:no-underline">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-blue-500" />
              <span className="font-semibold">Base Metrics</span>
            </div>
          </AccordionTrigger>
          <AccordionContent className="px-4 pb-4">
            <div className="space-y-4">
              {isV3 ? (
                <>
                  <CompactMetricCard
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
                  />
                  <CompactMetricCard
                    name="Attack Complexity"
                    abbreviation="AC"
                    value={(metrics as any).AC}
                    onChange={(v) => onMetricChange('AC', v)}
                    options={[
                      { value: 'L', label: 'Low', description: 'Specialized access', abbreviation: 'L' },
                      { value: 'H', label: 'High', description: 'Specialized conditions', abbreviation: 'H' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Privileges Required"
                    abbreviation="PR"
                    value={(metrics as any).PR}
                    onChange={(v) => onMetricChange('PR', v)}
                    options={[
                      { value: 'N', label: 'None', description: 'No privileges', abbreviation: 'N' },
                      { value: 'L', label: 'Low', description: 'Basic user', abbreviation: 'L' },
                      { value: 'H', label: 'High', description: 'Admin/Superuser', abbreviation: 'H' },
                    ]}
                  />
                  <CompactMetricCard
                    name="User Interaction"
                    abbreviation="UI"
                    value={(metrics as any).UI}
                    onChange={(v) => onMetricChange('UI', v)}
                    options={[
                      { value: 'N', label: 'None', description: 'No interaction', abbreviation: 'N' },
                      { value: 'R', label: 'Required', description: 'User participation', abbreviation: 'R' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Scope"
                    abbreviation="S"
                    value={(metrics as any).S}
                    onChange={(v) => onMetricChange('S', v)}
                    options={[
                      { value: 'U', label: 'Unchanged', description: 'Only vulnerable component', abbreviation: 'U' },
                      { value: 'C', label: 'Changed', description: 'Affects other components', abbreviation: 'C' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Confidentiality Impact"
                    abbreviation="C"
                    value={(metrics as any).C}
                    onChange={(v) => onMetricChange('C', v)}
                    options={[
                      { value: 'H', label: 'High', description: 'Total compromise', abbreviation: 'H' },
                      { value: 'L', label: 'Low', description: 'Partial compromise', abbreviation: 'L' },
                      { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Integrity Impact"
                    abbreviation="I"
                    value={(metrics as any).I}
                    onChange={(v) => onMetricChange('I', v)}
                    options={[
                      { value: 'H', label: 'High', description: 'Total compromise', abbreviation: 'H' },
                      { value: 'L', label: 'Low', description: 'Partial compromise', abbreviation: 'L' },
                      { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Availability Impact"
                    abbreviation="A"
                    value={(metrics as any).A}
                    onChange={(v) => onMetricChange('A', v)}
                    options={[
                      { value: 'H', label: 'High', description: 'Total compromise', abbreviation: 'H' },
                      { value: 'L', label: 'Low', description: 'Partial compromise', abbreviation: 'L' },
                      { value: 'N', label: 'None', description: 'No impact', abbreviation: 'N' },
                    ]}
                  />
                </>
              ) : (
                <>
                  <CompactMetricCard
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
                  />
                  <CompactMetricCard
                    name="Attack Complexity"
                    abbreviation="AC"
                    value={(metrics as any).AC}
                    onChange={(v) => onMetricChange('AC', v)}
                    options={[
                      { value: 'L', label: 'Low', description: 'Specialized access', abbreviation: 'L' },
                      { value: 'H', label: 'High', description: 'Specialized conditions', abbreviation: 'H' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Attack Requirements"
                    abbreviation="AT"
                    value={(metrics as any).AT}
                    onChange={(v) => onMetricChange('AT', v)}
                    options={[
                      { value: 'N', label: 'None', description: 'No attack requirements', abbreviation: 'N' },
                      { value: 'P', label: 'Present', description: 'Attack requirements present', abbreviation: 'P' },
                      { value: 'R', label: 'Required', description: 'Attack required', abbreviation: 'R' },
                    ]}
                  />
                </>
              )}
            </div>
          </AccordionContent>
        </AccordionItem>

        {/* Temporal Metrics (CVSS v3.x only) */}
        {isV3 && (
          <AccordionItem value="temporal">
            <AccordionTrigger className="px-4 py-3 hover:no-underline">
              <div className="flex items-center gap-2">
                <div className={cn(
                  'w-3 h-3 rounded-full',
                  showTemporal ? 'bg-green-500' : 'bg-slate-300'
                )} />
                <span className="font-semibold">Temporal Metrics</span>
                {!showTemporal && <span className="text-xs text-slate-400">(Hidden)</span>}
              </div>
            </AccordionTrigger>
            <AccordionContent className="px-4 pb-4">
              <div className="flex items-center justify-between mb-4">
                <p className="text-sm text-slate-600">
                  Adjust score based on exploit maturity and remediation status
                </p>
                <Button
                  size="sm"
                  variant={showTemporal ? 'default' : 'outline'}
                  onClick={toggleTemporal}
                >
                  {showTemporal ? 'Hide' : 'Show'}
                </Button>
              </div>

              {showTemporal && (metrics as any).temporal && (
                <div className="space-y-4">
                  <CompactMetricCard
                    name="Exploit Maturity"
                    abbreviation="E"
                    value={(metrics as any).temporal?.E || 'X'}
                    onChange={(v) => onMetricChange('temporal.E', v)}
                    options={[
                      { value: 'X', label: 'Not Defined', description: 'Not assigned', abbreviation: 'X' },
                      { value: 'U', label: 'Unproven', description: 'No proof of concept', abbreviation: 'U' },
                      { value: 'P', label: 'POC', description: 'Proof of concept exists', abbreviation: 'P' },
                      { value: 'F', label: 'Functional', description: 'Exploitable', abbreviation: 'F' },
                    ]}
                  />
                  <CompactMetricCard
                    name="Remediation Level"
                    abbreviation="RL"
                    value={(metrics as any).temporal?.RL || 'X'}
                    onChange={(v) => onMetricChange('temporal.RL', v)}
                    options={[
                      { value: 'X', label: 'Not Defined', description: 'Not assigned', abbreviation: 'X' },
                      { value: 'U', label: 'Unavailable', description: 'No fix available', abbreviation: 'U' },
                      { value: 'O', label: 'Workaround', description: 'Temporary workaround', abbreviation: 'O' },
                      { value: 'W', label: 'Official', description: 'Complete fix', abbreviation: 'W' },
                    ]}
                  />
                </div>
              )}
            </AccordionContent>
          </AccordionItem>
        )}

        {/* Vector String Sheet */}
        <AccordionItem value="vector">
          <AccordionTrigger className="px-4 py-3 hover:no-underline">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-purple-500" />
              <span className="font-semibold">Vector String</span>
            </div>
          </AccordionTrigger>
          <AccordionContent className="px-4 pb-4">
            <VectorString
              vectorString={(metrics as any).vectorString || ''}
              displayMode="compact"
              showCopyButton={true}
              showExportButton={true}
              showShareButton={true}
            />
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  );
}

// ============================================================================
// Main Mobile View Component
// ============================================================================

export function MobileView(props: MobileViewProps) {
  return (
    <div className="min-h-screen bg-slate-50 pb-20">
      {/* Fixed Header */}
      <MobileHeader
        version={props.version}
        scores={props.scores}
        onVersionChange={props.onVersionChange}
        onReset={props.onReset}
      />

      {/* Score Panel (collapsible) */}
      {props.scores && (
        <MobileScorePanel scores={props.scores} />
      )}

      {/* Metrics Accordion */}
      <MobileMetricsAccordion
        version={props.version}
        metrics={props.metrics}
        onMetricChange={props.onMetricChange}
        showTemporal={props.showTemporal}
        showEnvironmental={props.showEnvironmental}
        toggleTemporal={props.toggleTemporal}
        toggleEnvironmental={props.toggleEnvironmental}
      />

      {/* Bottom Action Bar */}
      <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-slate-200 p-4 z-40">
        <div className="flex gap-2">
          <Button
            variant="outline"
            className="flex-1"
            leftIcon={<Info className="h-4 w-4" />}
          >
            Guide
          </Button>
          <Button
            className="flex-1"
            leftIcon={<SlidersHorizontal className="h-4 w-4" />}
          >
            Export
          </Button>
        </div>
      </div>
    </div>
  );
}
