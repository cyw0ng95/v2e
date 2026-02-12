'use client';

/**
 * CVSS User Guide Component
 * Interactive guide explaining CVSS scoring methodology
 */

import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Calculator,
  Shield,
  Target,
  BookOpen,
  ExternalLink,
  AlertTriangle,
  CheckCircle
} from 'lucide-react';

interface GuideSection {
  id: string;
  title: string;
  icon: React.ReactNode;
  content: React.ReactNode;
}

export default function CVSSUserGuide() {
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['intro']));
  const [activeExample, setActiveExample] = useState<string>('example1');

  const toggleSection = (id: string) => {
    const newExpanded = new Set(expandedSections);
    if (newExpanded.has(id)) {
      newExpanded.delete(id);
    } else {
      newExpanded.add(id);
    }
    setExpandedSections(newExpanded);
  };

  const sections: GuideSection[] = [
    {
      id: 'intro',
      title: 'What is CVSS?',
      icon: <BookOpen className="h-5 w-5" />,
      content: (
        <div className="space-y-3 text-sm text-slate-700">
          <p>
            The <strong>Common Vulnerability Scoring System (CVSS)</strong> is an open industry
            standard for assessing the severity of security vulnerabilities. It provides a way to capture
            the principal characteristics of a vulnerability and produce a numerical score (0-10)
            reflecting its severity.
          </p>
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 className="font-semibold text-blue-900 mb-2">Key Concepts</h4>
            <ul className="space-y-2 text-xs">
              <li className="flex items-start gap-2">
                <CheckCircle className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
                <span><strong>Standardized:</strong> Provides consistent scoring across organizations</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
                <span><strong>Transparent:</strong> Open documentation ensures reproducible scoring</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
                <span><strong>Contextual:</strong> Base, Temporal, and Environmental scores for different use cases</span>
              </li>
            </ul>
          </div>
        </div>
      )
    },
    {
      id: 'scoring',
      title: 'Understanding CVSS Scores',
      icon: <Calculator className="h-5 w-5" />,
      content: (
        <div className="space-y-4 text-sm text-slate-700">
          <div>
            <h4 className="font-semibold text-slate-900 mb-2">Severity Ratings</h4>
            <div className="grid grid-cols-5 gap-2 text-xs">
              <div className="bg-gray-100 border border-gray-300 rounded-lg p-3 text-center">
                <div className="font-bold text-lg">0.0</div>
                <div className="text-gray-600">None</div>
              </div>
              <div className="bg-yellow-50 border border-yellow-300 rounded-lg p-3 text-center">
                <div className="font-bold text-lg">0.1-3.9</div>
                <div className="text-yellow-700">Low</div>
              </div>
              <div className="bg-orange-50 border border-orange-300 rounded-lg p-3 text-center">
                <div className="font-bold text-lg">4.0-6.9</div>
                <div className="text-orange-700">Medium</div>
              </div>
              <div className="bg-red-50 border border-red-300 rounded-lg p-3 text-center">
                <div className="font-bold text-lg">7.0-8.9</div>
                <div className="text-red-700">High</div>
              </div>
              <div className="bg-purple-50 border border-purple-300 rounded-lg p-3 text-center">
                <div className="font-bold text-lg">9.0-10.0</div>
                <div className="text-purple-700">Critical</div>
              </div>
            </div>
          </div>

          <div>
            <h4 className="font-semibold text-slate-900 mb-2">Score Types</h4>
            <div className="space-y-2 text-xs">
              <div className="bg-slate-50 rounded-lg p-3">
                <div className="font-semibold text-slate-900">Base Score</div>
                <p className="text-slate-600 mt-1">
                  The core severity based on intrinsic properties of the vulnerability.
                  Constant over time and across user environments.
                </p>
              </div>
              <div className="bg-slate-50 rounded-lg p-3">
                <div className="font-semibold text-slate-900">Temporal Score (v3.x)</div>
                <p className="text-slate-600 mt-1">
                  Adjusts the Base score based on factors that change over time:
                  exploit code maturity, available remediation, and report confidence.
                </p>
              </div>
              <div className="bg-slate-50 rounded-lg p-3">
                <div className="font-semibold text-slate-900">Environmental Score</div>
                <p className="text-slate-600 mt-1">
                  Customizes the score based on specific organizational requirements
                  and the characteristics of the user's environment.
                </p>
              </div>
            </div>
          </div>
        </div>
      )
    },
    {
      id: 'metrics',
      title: 'Base Metrics Explained',
      icon: <Target className="h-5 w-5" />,
      content: (
        <div className="space-y-3 text-sm text-slate-700">
          <p>
            Base metrics capture the characteristics of a vulnerability that are constant over time
            and across user environments. They are divided into two categories:
          </p>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <h4 className="font-semibold text-slate-900 mb-2 text-xs">Exploitability Metrics</h4>
              <div className="space-y-2 text-xs">
                <div className="border-l-2 border-blue-500 pl-3">
                  <div className="font-medium">Attack Vector (AV)</div>
                  <p className="text-slate-600 mt-1">How is the vulnerability accessed?</p>
                  <div className="mt-1 text-xs text-slate-500">Network &gt; Adjacent &gt; Local &gt; Physical</div>
                </div>
                <div className="border-l-2 border-blue-500 pl-3">
                  <div className="font-medium">Attack Complexity (AC)</div>
                  <p className="text-slate-600 mt-1">How difficult is the attack?</p>
                  <div className="mt-1 text-xs text-slate-500">Low &gt; High</div>
                </div>
                <div className="border-l-2 border-blue-500 pl-3">
                  <div className="font-medium">Privileges Required (PR)</div>
                  <p className="text-slate-600 mt-1">What privileges are needed?</p>
                  <div className="mt-1 text-xs text-slate-500">None &gt; Low &gt; High</div>
                </div>
                <div className="border-l-2 border-blue-500 pl-3">
                  <div className="font-medium">User Interaction (UI)</div>
                  <p className="text-slate-600 mt-1">Is user action required?</p>
                  <div className="mt-1 text-xs text-slate-500">None &gt; Required</div>
                </div>
              </div>
            </div>

            <div>
              <h4 className="font-semibold text-slate-900 mb-2 text-xs">Impact Metrics</h4>
              <div className="space-y-2 text-xs">
                <div className="border-l-2 border-red-500 pl-3">
                  <div className="font-medium">Scope (S)</div>
                  <p className="text-slate-600 mt-1">Does it affect other components?</p>
                  <div className="mt-1 text-xs text-slate-500">Changed vs Unchanged</div>
                </div>
                <div className="border-l-2 border-red-500 pl-3">
                  <div className="font-medium">Confidentiality (C)</div>
                  <p className="text-slate-600 mt-1">Is data exposed?</p>
                  <div className="mt-1 text-xs text-slate-500">High &gt; Low &gt; None</div>
                </div>
                <div className="border-l-2 border-red-500 pl-3">
                  <div className="font-medium">Integrity (I)</div>
                  <p className="text-slate-600 mt-1">Can data be modified?</p>
                  <div className="mt-1 text-xs text-slate-500">High &gt; Low &gt; None</div>
                </div>
                <div className="border-l-2 border-red-500 pl-3">
                  <div className="font-medium">Availability (A)</div>
                  <p className="text-slate-600 mt-1">Is service disrupted?</p>
                  <div className="mt-1 text-xs text-slate-500">High &gt; Low &gt; None</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )
    },
    {
      id: 'temporal',
      title: 'Temporal Metrics (v3.0/v3.1)',
      icon: <Shield className="h-5 w-5" />,
      content: (
        <div className="space-y-3 text-sm text-slate-700">
          <p>
            Temporal metrics adjust the Base score based on factors that <strong>change over time</strong>
            but are not specific to a particular customer environment.
          </p>

          <div className="space-y-2 text-xs">
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-mono font-bold text-amber-800">E - Exploit Maturity</span>
              </div>
              <p className="text-slate-600">
                The current state of exploit techniques or code availability.
                This can change rapidly as exploits are developed.
              </p>
              <div className="mt-2 text-xs text-slate-500">
                Values: Unproven → PoC → Functional → High → Official → Attacked
              </div>
            </div>

            <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-mono font-bold text-amber-800">RL - Remediation Level</span>
              </div>
              <p className="text-slate-600">
                The availability of a fix or workaround for the vulnerability.
              </p>
              <div className="mt-2 text-xs text-slate-500">
                Values: Unavailable → Workaround → Temporary Fix → Official Fix
              </div>
            </div>

            <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-mono font-bold text-amber-800">RC - Report Confidence</span>
              </div>
              <p className="text-slate-600">
                How reliable is the report of the vulnerability?
              </p>
              <div className="mt-2 text-xs text-slate-500">
                Values: Unknown → Reasonable → Confirmed
              </div>
            </div>
          </div>
        </div>
      )
    },
    {
      id: 'environmental',
      title: 'Environmental Metrics',
      icon: <AlertTriangle className="h-5 w-5" />,
      content: (
        <div className="space-y-3 text-sm text-slate-700">
          <p>
            Environmental metrics allow organizations to <strong>customize the CVSS score</strong>
            based on their specific security requirements and the characteristics of their environment.
          </p>

          <div className="bg-green-50 border border-green-200 rounded-lg p-4">
            <h4 className="font-semibold text-green-900 mb-2">When to Use Environmental Metrics</h4>
            <ul className="space-y-1 text-xs">
              <li>• You have specific Confidentiality/Integrity/Availability requirements</li>
              <li>• You have deployed mitigations that affect exploitability</li>
              <li>• You need to prioritize vulnerabilities for your specific environment</li>
              <li>• You have modified base metric values based on configuration</li>
            </ul>
          </div>

          <div>
            <h4 className="font-semibold text-slate-900 mb-2">Modified Metrics</h4>
            <p className="text-xs text-slate-600 mb-2">
              Environmental metrics include <strong>Modified (M)</strong> versions of all base metrics.
              If not specified, the base metric value is used.
            </p>
            <div className="bg-slate-100 rounded-lg p-3 text-xs">
              <div className="font-mono mb-1">CR, IR, AR</div>
              <p className="text-slate-600">
                Confidentiality/Integrity/Availability Requirements indicate the importance
                of the impacted asset to your organization. Multipliers are: None (1x), Low (0.5x),
                Medium (1x), High (1.5x).
              </p>
            </div>
          </div>
        </div>
      )
    },
    {
      id: 'examples',
      title: 'Example Vulnerabilities',
      icon: <Target className="h-5 w-5" />,
      content: (
        <div className="space-y-3">
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setActiveExample('example1')}
              className={`px-3 py-2 text-xs font-medium rounded-lg transition-colors ${
                activeExample === 'example1'
                  ? 'bg-blue-600 text-white'
                  : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
              }`}
            >
              Log4Shell (CVE-2021-44228)
            </button>
            <button
              type="button"
              onClick={() => setActiveExample('example2')}
              className={`px-3 py-2 text-xs font-medium rounded-lg transition-colors ${
                activeExample === 'example2'
                  ? 'bg-blue-600 text-white'
                  : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
              }`}
            >
              EternalBlue (CVE-2017-0144)
            </button>
            <button
              type="button"
              onClick={() => setActiveExample('example3')}
              className={`px-3 py-2 text-xs font-medium rounded-lg transition-colors ${
                activeExample === 'example3'
                  ? 'bg-blue-600 text-white'
                  : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
              }`}
            >
              XSS Vulnerability
            </button>
          </div>

          {activeExample === 'example1' && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm">
              <h4 className="font-semibold text-red-900 mb-2">Log4Shell - CVSS 3.1: CRITICAL (10.0)</h4>
              <div className="space-y-2 text-xs text-slate-700">
                <p className="font-medium">Vector: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H</p>
                <div>
                  <div className="font-semibold text-slate-900">Why Critical?</div>
                  <ul className="mt-1 space-y-1 text-slate-600">
                    <li>• <strong>AV:N</strong> - Exploitable remotely over network</li>
                    <li>• <strong>PR:N</strong> - No privileges required</li>
                    <li>• <strong>UI:N</strong> - No user interaction needed</li>
                    <li>• <strong>S:C</strong> - Scope changed (affects other systems)</li>
                    <li>• <strong>C:H/I:H/A:H</strong> - Complete data exposure and control</li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {activeExample === 'example2' && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm">
              <h4 className="font-semibold text-red-900 mb-2">EternalBlue - CVSS 3.0: CRITICAL (9.8)</h4>
              <div className="space-y-2 text-xs text-slate-700">
                <p className="font-medium">Vector: CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H</p>
                <div>
                  <div className="font-semibold text-slate-900">Why Critical?</div>
                  <ul className="mt-1 space-y-1 text-slate-600">
                    <li>• <strong>AV:N</strong> - Remote exploitation via SMB</li>
                    <li>• <strong>PR:N</strong> - No authentication required</li>
                    <li>• <strong>S:C</strong> - Wormable, spreads to other systems</li>
                    <li>• Impact: Complete system compromise possible</li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {activeExample === 'example3' && (
            <div className="bg-orange-50 border border-orange-200 rounded-lg p-4 text-sm">
              <h4 className="font-semibold text-orange-900 mb-2">Reflected XSS - CVSS 3.1: MEDIUM (6.1)</h4>
              <div className="space-y-2 text-xs text-slate-700">
                <p className="font-medium">Vector: CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N</p>
                <div>
                  <div className="font-semibold text-slate-900">Why Medium?</div>
                  <ul className="mt-1 space-y-1 text-slate-600">
                    <li>• <strong>UI:R</strong> - Requires user interaction (click link)</li>
                    <li>• <strong>C:L/I:L</strong> - Limited impact (browser context only)</li>
                    <li>• <strong>A:N</strong> - No availability impact</li>
                  </ul>
                </div>
              </div>
            </div>
          )}
        </div>
      )
    }
  ];

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
      <div className="flex items-center gap-2 mb-4">
        <BookOpen className="h-6 w-6 text-blue-600" />
        <h2 className="text-xl font-bold text-slate-900">CVSS User Guide</h2>
      </div>

      <div className="space-y-2">
        {sections.map((section) => (
          <div
            key={section.id}
            className="border border-slate-200 rounded-lg overflow-hidden"
          >
            <button
              type="button"
              onClick={() => toggleSection(section.id)}
              className="w-full flex items-center justify-between p-4 hover:bg-slate-50 transition-colors text-left"
            >
              <div className="flex items-center gap-3">
                <div className="text-blue-600">{section.icon}</div>
                <span className="font-semibold text-slate-900">{section.title}</span>
              </div>
              {expandedSections.has(section.id) ? (
                <ChevronDown className="h-5 w-5 text-slate-500" />
              ) : (
                <ChevronRight className="h-5 w-5 text-slate-500" />
              )}
            </button>

            {expandedSections.has(section.id) && (
              <div className="px-4 pb-4 border-t border-slate-100 pt-4">
                {section.content}
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="mt-6 pt-4 border-t border-slate-200">
        <h3 className="font-semibold text-slate-900 mb-3 text-sm">External Resources</h3>
        <div className="grid sm:grid-cols-2 gap-3 text-xs">
          <a
            href="https://www.first.org/cvss/calculator/4.0"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium"
          >
            <ExternalLink className="h-4 w-4" />
            <span>CVSS v4.0 Specification</span>
          </a>
          <a
            href="https://www.first.org/cvss/calculator/3.1"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium"
          >
            <ExternalLink className="h-4 w-4" />
            <span>CVSS v3.1 Specification</span>
          </a>
          <a
            href="https://www.first.org/cvss/calculator/3.0"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium"
          >
            <ExternalLink className="h-4 w-4" />
            <span>CVSS v3.0 Specification</span>
          </a>
          <a
            href="https://www.first.org/cvss/calculator-information"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium"
          >
            <ExternalLink className="h-4 w-4" />
            <span>CVSS Documentation</span>
          </a>
        </div>
      </div>
    </div>
  );
}
