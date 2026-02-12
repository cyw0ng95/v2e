'use client';

/**
 * CVSS Calculator Component
 * Main calculator interface with metric selectors and score display
 */

import { useState, useCallback } from 'react';
import {
  ArrowUpDown,
  Copy,
  Download,
  Eye,
  EyeOff,
  FileText,
  Info,
  Link as Link2,
  Settings,
  X,
  Check,
  BookOpen
} from 'lucide-react';
import { useCVSS } from '@/lib/cvss-context';
import { getCVSSMetadata } from '@/lib/cvss-calculator';
import type { CVSSVersion, CVSSSeverity, CVSS3ScoreBreakdown, CVSS4ScoreBreakdown } from '@/lib/types';
import CVSSUserGuide from './user-guide';

const severityColors: Record<CVSSSeverity, string> = {
  NONE: 'from-gray-400 to-gray-500 bg-gray-50',
  LOW: 'from-yellow-400 to-yellow-600 bg-yellow-50',
  MEDIUM: 'from-orange-400 to-orange-600 bg-orange-50',
  HIGH: 'from-red-400 to-red-600 bg-red-50',
  CRITICAL: 'from-purple-500 to-purple-700 bg-purple-50'
};

const severityLabels: Record<CVSSSeverity, string> = {
  NONE: 'None (0.0)',
  LOW: 'Low (0.1-3.9)',
  MEDIUM: 'Medium (4.0-6.9)',
  HIGH: 'High (7.0-8.9)',
  CRITICAL: 'Critical (9.0-10.0)'
};

const avOptions = [
  { value: 'N', label: 'Network (N)', desc: 'Network exploitable' },
  { value: 'A', label: 'Adjacent (A)', desc: 'Adjacent network' },
  { value: 'L', label: 'Local (L)', desc: 'Local access' },
  { value: 'P', label: 'Physical (P)', desc: 'Physical access' }
];

const acOptions = [
  { value: 'L', label: 'Low (L)', desc: 'Specialized access' },
  { value: 'H', label: 'High (H)', desc: 'Specialized conditions' }
];

const prOptions = [
  { value: 'N', label: 'None (N)', desc: 'No privilege required' },
  { value: 'L', label: 'Low (L)', desc: 'Basic user' },
  { value: 'H', label: 'High (H)', desc: 'Admin/Superuser' }
];

const uiOptions = [
  { value: 'N', label: 'None (N)', desc: 'No interaction' },
  { value: 'R', label: 'Required (R)', desc: 'User participation' }
];

const sOptions = [
  { value: 'U', label: 'Unchanged (U)', desc: 'Only vulnerable component' },
  { value: 'C', label: 'Changed (C)', desc: 'Affects other components' }
];

const cOptions = [
  { value: 'H', label: 'High (H)', desc: 'Total compromise' },
  { value: 'L', label: 'Low (L)', desc: 'Partial compromise' },
  { value: 'N', label: 'None (N)', desc: 'No impact' }
];

interface MetricSelectorProps<T extends string> {
  value: T;
  onChange: (val: T) => void;
  options: Array<{ value: string; label: string; desc: string }>;
  label: string;
  description?: string;
}

function MetricSelector<T extends string>({
  value,
  onChange,
  options,
  label,
  description
}: MetricSelectorProps<T>) {
  return (
    <div className="space-y-2">
      {description && (
        <p className="text-xs text-slate-500">{description}</p>
      )}
      <label className="block text-sm font-medium text-slate-700 mb-2">{label}</label>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
        {options.map((opt) => (
          <button
            key={opt.value}
            type="button"
            onClick={() => onChange(opt.value as T)}
            className={`p-3 rounded-lg border-2 text-left transition-all duration-200 ${
              value === opt.value
                ? 'border-blue-500 bg-blue-50 text-blue-700 font-medium shadow-sm'
                : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-50'
            }`}
          >
            <div className="font-semibold">{opt.value}</div>
            <div className="text-xs">{opt.label}</div>
          </button>
        ))}
      </div>
    </div>
  );
}

export default function CVSSCalculator() {
  const {
    state,
    version,
    updateMetric,
    setVersion,
    resetMetrics,
    toggleTemporal,
    toggleEnvironmental,
    exportData
  } = useCVSS();

  const [showVectorHelp, setShowVectorHelp] = useState(false);
  const [showExportMenu, setShowExportMenu] = useState(false);
  const [copiedVector, setCopiedVector] = useState(false);
  const [showUserGuide, setShowUserGuide] = useState(false);

  const metadata = getCVSSMetadata(state.version);

  const handleCopyVector = async () => {
    // Check if vectorString exists (for CVSS 4.x which has vector strings)
    const vectorString = (state.scores as any)?.vectorString;
    if (vectorString) {
      await navigator.clipboard.writeText(vectorString);
      setCopiedVector(true);
      setTimeout(() => setCopiedVector(false), 2000);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <header className="border-b border-slate-200 bg-white/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-4">
              <a href="/cvss" className="text-blue-600 hover:text-blue-700">
                <ArrowUpDown className="h-5 w-5" />
              </a>
              <div>
                <h1 className="text-xl font-bold text-slate-900">CVSS Calculator {metadata.name}</h1>
                <a
                  href={metadata.specUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:text-blue-700 ml-2"
                >
                  <Info className="h-4 w-4 inline" />
                </a>
              </div>
            </div>

            <div className="flex items-center gap-2">
              {(['3.0', '3.1', '4.0'] as CVSSVersion[]).map((v) => (
                <button
                  key={v}
                  onClick={() => setVersion(v)}
                  className={`px-3 py-2 rounded-lg text-sm font-medium transition-all ${
                    version === v
                      ? 'bg-blue-600 text-white shadow-sm'
                      : 'bg-white text-slate-700 border border-slate-200 hover:border-blue-300'
                  }`}
                >
                  {v}
                </button>
              ))}
            </div>

            <button
              onClick={resetMetrics}
              className="text-sm text-slate-500 hover:text-slate-700 flex items-center gap-2"
              title="Reset all metrics to defaults"
            >
              <Settings className="h-4 w-4" />
              Reset
            </button>

            <button
              onClick={() => setShowUserGuide(!showUserGuide)}
              className="text-sm text-blue-600 hover:text-blue-700 flex items-center gap-2"
              title="Show CVSS user guide"
            >
              <BookOpen className="h-4 w-4" />
              Guide
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
        {showUserGuide && (
          <div className="mb-6">
            <CVSSUserGuide />
          </div>
        )}

        <div className="grid lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            <section className="bg-white rounded-xl shadow-sm p-6 border border-slate-200">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold text-slate-900">Base Metrics</h2>
              </div>

              <div className="space-y-5">
                <MetricSelector
                  label="Attack Vector (AV)"
                  description="How vulnerable is component?"
                  value={(state.metrics as any).AV}
                  onChange={(val) => updateMetric('AV', val)}
                  options={avOptions}
                />
                <MetricSelector
                  label="Attack Complexity (AC)"
                  description="How complex is attack?"
                  value={(state.metrics as any).AC}
                  onChange={(val) => updateMetric('AC', val)}
                  options={acOptions}
                />
                <MetricSelector
                  label="Privileges Required (PR)"
                  description="What privileges does attacker need?"
                  value={(state.metrics as any).PR}
                  onChange={(val) => updateMetric('PR', val)}
                  options={prOptions}
                />
                <MetricSelector
                  label="User Interaction (UI)"
                  description="Is user interaction required?"
                  value={(state.metrics as any).UI}
                  onChange={(val) => updateMetric('UI', val)}
                  options={uiOptions}
                />
                <MetricSelector
                  label="Scope (S)"
                  description="Does exploit affect other components?"
                  value={(state.metrics as any).S}
                  onChange={(val) => updateMetric('S', val)}
                  options={sOptions}
                />
                <MetricSelector
                  label="Confidentiality (C)"
                  description="Impact on data confidentiality?"
                  value={(state.metrics as any).C}
                  onChange={(val) => updateMetric('C', val)}
                  options={cOptions}
                />
                <MetricSelector
                  label="Integrity (I)"
                  description="Impact on data integrity?"
                  value={(state.metrics as any).I}
                  onChange={(val) => updateMetric('I', val)}
                  options={cOptions}
                />
                <MetricSelector
                  label="Availability (A)"
                  description="Impact on system availability?"
                  value={(state.metrics as any).A}
                  onChange={(val) => updateMetric('A', val)}
                  options={cOptions}
                />
              </div>
            </section>

            {(state.version === '3.0' || state.version === '3.1') && (
              <section className="bg-white rounded-xl shadow-sm p-6 border border-slate-200">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-bold text-slate-900">Temporal Metrics</h2>
                  <button
                    onClick={toggleTemporal}
                    className={`text-sm px-3 py-2 rounded-lg flex items-center gap-1 ${
                      state.showTemporal
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                    }`}
                  >
                    {state.showTemporal ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                    {state.showTemporal ? 'Hide' : 'Show'}
                  </button>
                </div>

                {state.showTemporal && (
                  <div className="space-y-5">
                    <MetricSelector
                      label="Exploit Maturity (E)"
                      description="Current state of exploit code?"
                      value={(state.metrics as any).temporal?.E || 'X'}
                      onChange={(val) => updateMetric('temporal.E' as any, val)}
                      options={[
                        { value: 'X', label: 'Not Defined (X)', desc: 'Not assigned' },
                        { value: 'U', label: 'Unproven (U)', desc: 'No proof of concept' },
                        { value: 'P', label: 'Proof-of-Concept (P)', desc: 'PoC exists' },
                        { value: 'F', label: 'Functional (F)', desc: 'Exploitable' },
                        { value: 'H', label: 'High (H)', desc: 'Reliable exploit' },
                        { value: 'R', label: 'Official (R)', desc: 'Vendor confirmed' }
                      ]}
                    />
                    <MetricSelector
                      label="Remediation Level (RL)"
                      description="Is a fix available?"
                      value={(state.metrics as any).temporal?.RL || 'X'}
                      onChange={(val) => updateMetric('temporal.RL' as any, val)}
                      options={[
                        { value: 'X', label: 'Not Defined (X)', desc: 'Not assigned' },
                        { value: 'U', label: 'Unavailable (U)', desc: 'No fix available' },
                        { value: 'O', label: 'Workaround (O)', desc: 'Temporary workaround' },
                        { value: 'T', label: 'Temporary (T)', desc: 'Partial fix' },
                        { value: 'W', label: 'Official (W)', desc: 'Complete fix' }
                      ]}
                    />
                    <MetricSelector
                      label="Report Confidence (RC)"
                      description="How reliable is report?"
                      value={(state.metrics as any).temporal?.RC || 'X'}
                      onChange={(val) => updateMetric('temporal.RC' as any, val)}
                      options={[
                        { value: 'X', label: 'Not Defined (X)', desc: 'Not assigned' },
                        { value: 'U', label: 'Unknown (U)', desc: 'Unverified' },
                        { value: 'C', label: 'Reasonable (C)', desc: 'Some confidence' },
                        { value: 'R', label: 'Confirmed (R)', desc: 'Verified report' }
                      ]}
                    />
                  </div>
                )}
              </section>
            )}
          </div>

          <div className="lg:col-span-1 space-y-6">
            <section className={`rounded-2xl shadow-lg p-8 border-2 ${
              (state.scores && 'finalSeverity' in state.scores && state.scores.finalSeverity)
                ? severityColors[state.scores.finalSeverity]
                : 'bg-slate-100 border-slate-200'
            }`}>
              {state.scores ? (
                <>
                  <div className="text-center mb-6">
                    <div className="text-sm text-slate-600 mb-2">Base Score</div>
                    <div className="text-5xl font-bold">{state.scores.baseScore.toFixed(1)}</div>
                  </div>
                  {(state.version === '3.0' || state.version === '3.1') && (state.scores as CVSS3ScoreBreakdown).temporalScore !== undefined && (
                    <div className="text-center py-4 border-t border-slate-200">
                      <div className="text-sm text-slate-600">Temporal Score</div>
                      <div className="text-2xl font-semibold">{(state.scores as CVSS3ScoreBreakdown).temporalScore?.toFixed(1)}</div>
                    </div>
                  )}
                  {state.scores.environmentalScore !== undefined && (
                    <div className="text-center py-4 border-t border-slate-200">
                      <div className="text-sm text-slate-600">Environmental Score</div>
                      <div className="text-2xl font-semibold">{state.scores.environmentalScore.toFixed(1)}</div>
                    </div>
                  )}
                  {state.version === '4.0' && (state.scores as CVSS4ScoreBreakdown).threatScore !== undefined && (
                    <div className="text-center py-4 border-t border-slate-200">
                      <div className="text-sm text-slate-600">Threat Score</div>
                      <div className="text-2xl font-semibold">{(state.scores as CVSS4ScoreBreakdown).threatScore?.toFixed(1)}</div>
                    </div>
                  )}
                  <div className="text-center pt-4 border-t-2 border-slate-300">
                    <div className="text-sm text-slate-600 mb-2">Final Score</div>
                    <div className="text-5xl font-bold">{state.scores.baseScore.toFixed(1)}</div>
                    <div className="text-sm font-semibold mt-2">{severityLabels[state.scores.finalSeverity || state.scores.baseSeverity]}</div>
                  </div>
                </>
              ) : (
                <div className="text-center py-12">
                  <div className="animate-pulse">
                    <div className="text-6xl font-bold text-slate-300 mb-4">Calculating...</div>
                  </div>
                </div>
              )}
            </section>

            <section className="bg-white rounded-xl shadow-sm p-6 border border-slate-200">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold text-slate-900">Vector String</h2>
                <button
                  onClick={handleCopyVector}
                  className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors flex items-center gap-2"
                  title="Copy vector string to clipboard"
                >
                  {copiedVector ? (
                    <>
                      <Check className="h-4 w-4 text-green-400" />
                      <span className="text-sm text-green-400">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4 text-slate-100" />
                      <span className="text-sm text-slate-100">Copy</span>
                    </>
                  )}
                </button>
              </div>

              <code className="block bg-slate-800 text-green-400 p-4 rounded-lg text-sm break-all font-mono">
                {state.vectorString || ''}
              </code>

              <div className="mt-6 relative">
                <button
                  onClick={() => setShowExportMenu(!showExportMenu)}
                  className="w-full py-3 px-4 rounded-lg bg-slate-800 hover:bg-slate-700 text-white font-medium flex items-center justify-center gap-2 transition-colors"
                >
                  <Download className="h-5 w-5" />
                  Export CVSS Data
                </button>

                {showExportMenu && (
                  <div className="absolute top-full left-0 right-0 mt-2 bg-white rounded-lg shadow-xl border border-slate-200 z-50">
                    <div className="p-2">
                      <button
                        onClick={() => {
                          exportData('json');
                          setShowExportMenu(false);
                        }}
                        className="w-full text-left px-4 py-3 rounded-lg hover:bg-slate-100 transition-colors flex items-center gap-3"
                      >
                        <FileText className="h-5 w-5 text-blue-600" />
                        <div>
                          <div className="font-medium">JSON</div>
                          <div className="text-xs text-slate-500">Export as JSON file</div>
                        </div>
                      </button>
                      <button
                        onClick={() => {
                          exportData('csv');
                          setShowExportMenu(false);
                        }}
                        className="w-full text-left px-4 py-3 rounded-lg hover:bg-slate-100 transition-colors flex items-center gap-3"
                      >
                        <FileText className="h-5 w-5 text-green-600" />
                        <div>
                          <div className="font-medium">CSV</div>
                          <div className="text-xs text-slate-500">Export as CSV file</div>
                        </div>
                      </button>
                      <button
                        onClick={() => {
                          exportData('url');
                          setShowExportMenu(false);
                        }}
                        className="w-full text-left px-4 py-3 rounded-lg hover:bg-slate-100 transition-colors flex items-center gap-3"
                      >
                        <Link2 className="h-5 w-5 text-purple-600" />
                        <div>
                          <div className="font-medium">Share URL</div>
                          <div className="text-xs text-slate-500">Copy URL to clipboard</div>
                        </div>
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </section>

            <section className="bg-blue-50 rounded-xl p-6 border border-blue-200">
              <h3 className="font-semibold text-blue-900 mb-2">CVSS Scoring Guide</h3>
              <p className="text-sm text-slate-700 mb-2">Base Score (0-10): The core severity based on exploitability and impact metrics.</p>
              <p className="text-sm text-slate-700 mb-2">Temporal Score: Adjustments for exploit maturity, remediation availability, and report confidence.</p>
              <p className="text-sm text-slate-700 mb-2">Environmental Score: Customizes for your specific security environment and requirements.</p>
              <p className="text-sm text-slate-700">
                <strong>Severity</strong>:
                <span className="ml-2 text-xs">None (0.0), Low (0.1-3.9), Medium (4.0-6.9), High (7.0-8.9), Critical (9.0-10.0)</span>
              </p>
            </section>
          </div>
        </div>
      </main>

      <footer className="border-t border-slate-200 bg-white/80 backdrop-blur-sm mt-12">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-sm text-slate-600 text-center">
            CVSS Calculator following FIRST.org specifications. Scores are calculated client-side.
          </p>
        </div>
      </footer>
    </div>
  );
}
