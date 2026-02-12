'use client';

/**
 * CVSS Calculator - Version Selector Landing Page
 * /cvss route with version selection
 */

import Link from 'next/link';
import { ArrowRight, Calculator, FileText, Share2, ExternalLink, BookOpen } from 'lucide-react';

const versions = [
  {
    id: '4.0',
    name: 'CVSS v4.0',
    description: 'Latest version with enhanced granularity and threat modeling',
    specUrl: 'https://www.first.org/cvss/calculator/4.0',
    status: 'current',
    color: 'from-blue-500 to-blue-600'
  },
  {
    id: '3.1',
    name: 'CVSS v3.1',
    description: 'Refined v3.0 with additional environmental metrics',
    specUrl: 'https://www.first.org/cvss/calculator/3.1',
    status: 'stable',
    color: 'from-emerald-500 to-emerald-600'
  },
  {
    id: '3.0',
    name: 'CVSS v3.0',
    description: 'Current standard with temporal and environmental metrics',
    specUrl: 'https://www.first.org/cvss/calculator/3.0',
    status: 'stable',
    color: 'from-emerald-500 to-emerald-600'
  }
];

const features = [
  { icon: Calculator, title: 'Real-time Calculation', description: 'Instant score updates as you adjust metrics' },
  { icon: FileText, title: 'Vector Generation', description: 'Auto-generate CVSS vector strings' },
  { icon: Share2, title: 'Export Options', description: 'JSON, CSV, and URL sharing' },
  { icon: Calculator, title: 'Version Support', description: 'CVSS v3.0, v3.1, and v4.0' }
];

export default function CVSSPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <header className="border-b border-slate-200 bg-white/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <Calculator className="h-8 w-8 text-blue-600" />
              <h1 className="text-2xl font-bold text-slate-900">
                CVSS Calculator
              </h1>
            </div>
            <nav className="hidden md:flex items-center gap-6">
              <Link
                href="/"
                className="text-sm font-medium text-slate-600 hover:text-blue-600 transition-colors"
              >
                Back to v2e
              </Link>
            </nav>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
        {/* Page Header */}
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold text-slate-900 mb-4">
            Common Vulnerability Scoring System
          </h2>
          <p className="text-lg text-slate-600 max-w-2xl mx-auto">
            Calculate CVSS scores for vulnerabilities using FIRST.org standards.
            Select a CVSS version below to get started.
          </p>
        </div>

        {/* Version Cards */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
          {versions.map((version) => (
            <Link
              key={version.id}
              href={`/cvss/${version.id}`}
              className="group relative bg-white rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 overflow-hidden border border-slate-200 hover:border-blue-300"
            >
              {/* Status Badge */}
              {version.status !== 'current' && (
                <div className="absolute top-4 right-4">
                  <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold text-white ${version.color.replace('to-', ' bg-gradient-to-r from-')}`}>
                    {version.status}
                  </span>
                </div>
              )}

              {/* Content */}
              <div className="p-8">
                <div className="flex items-start gap-4 mb-4">
                  <div className={`p-3 rounded-xl ${version.color.replace('to-', ' bg-gradient-to-br from-')}`}>
                    <Calculator className="h-8 w-8 text-white" />
                  </div>
                  <div className="flex-1">
                    <h3 className="text-xl font-bold text-slate-900 mb-2 group-hover:text-blue-600 transition-colors">
                      {version.name}
                    </h3>
                    <p className="text-sm text-slate-600 mb-4">
                      {version.description}
                    </p>
                    <a
                      href={version.specUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-600 hover:text-blue-700 font-medium inline-flex items-center gap-1"
                    >
                      View Specification
                      <ArrowRight className="h-4 w-4" />
                    </a>
                  </div>
                </div>

                {/* Quick Start Button */}
                <div className="pt-4 border-t border-slate-100">
                  <button className={`w-full py-3 px-4 rounded-lg font-semibold text-white ${version.color.replace('to-', ' bg-gradient-to-r from-')} hover:opacity-90 transition-opacity flex items-center justify-center gap-2`}>
                    {version.id === '4.0' ? 'Explore v4.0' : `Open ${version.name} Calculator`}
                    <ArrowRight className="h-5 w-5" />
                  </button>
                </div>
              </div>

              {/* Hover Gradient Effect */}
              <div className={`absolute inset-0 bg-gradient-to-br ${version.color.replace('to-', ' to-')} opacity-0 group-hover:opacity-5 transition-opacity duration-300 -z-10`} />
            </Link>
          ))}
        </div>

        {/* Features Section */}
        <div className="bg-white rounded-2xl shadow-lg p-8 mb-12 border border-slate-200">
          <h3 className="text-xl font-bold text-slate-900 mb-6">
            Calculator Features
          </h3>
          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {features.map((feature) => (
              <div key={feature.title} className="flex gap-4">
                <div className="flex-shrink-0">
                  <div className="p-3 bg-slate-100 rounded-lg">
                    <feature.icon className="h-6 w-6 text-slate-600" />
                  </div>
                </div>
                <div>
                  <h4 className="font-semibold text-slate-900 mb-1">
                    {feature.title}
                  </h4>
                  <p className="text-sm text-slate-600">
                    {feature.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* About CVSS Section */}
        <div className="bg-slate-50 rounded-2xl p-8 border border-slate-200">
          <h3 className="text-xl font-bold text-slate-900 mb-4">
            About CVSS
          </h3>
          <p className="text-slate-700 mb-4">
            The Common Vulnerability Scoring System (CVSS) provides a way to capture the
            principal characteristics of a vulnerability and produce a numerical score
            reflecting its severity. CVSS is owned and managed by FIRST.org
            (Forum of Incident Response and Security Teams).
          </p>

          <div className="bg-blue-50 border border-blue-200 rounded-xl p-5 mb-4">
            <h4 className="font-semibold text-blue-900 mb-2">Understanding CVSS Scores</h4>
            <div className="grid grid-cols-5 gap-2 text-xs mb-4">
              <div className="bg-gray-100 border border-gray-300 rounded-lg p-2 text-center">
                <div className="font-bold">0.0</div>
                <div className="text-gray-600">None</div>
              </div>
              <div className="bg-yellow-100 border border-yellow-300 rounded-lg p-2 text-center">
                <div className="font-bold">0.1-3.9</div>
                <div className="text-yellow-700">Low</div>
              </div>
              <div className="bg-orange-100 border border-orange-300 rounded-lg p-2 text-center">
                <div className="font-bold">4.0-6.9</div>
                <div className="text-orange-700">Medium</div>
              </div>
              <div className="bg-red-100 border border-red-300 rounded-lg p-2 text-center">
                <div className="font-bold">7.0-8.9</div>
                <div className="text-red-700">High</div>
              </div>
              <div className="bg-purple-100 border border-purple-300 rounded-lg p-2 text-center">
                <div className="font-bold">9.0-10.0</div>
                <div className="text-purple-700">Critical</div>
              </div>
            </div>
            <div className="space-y-1 text-xs text-slate-700">
              <p><strong>Base Score:</strong> Core severity based on intrinsic properties (constant over time)</p>
              <p><strong>Temporal Score:</strong> Adjusted for exploit maturity, remediation, and report confidence</p>
              <p><strong>Environmental Score:</strong> Customized for your organization's requirements</p>
            </div>
          </div>

          <h4 className="font-semibold text-slate-900 mb-3">Official Resources</h4>
          <div className="grid sm:grid-cols-2 gap-3">
            <a
              href="https://www.first.org/cvss/calculator/4.0"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 p-3 bg-white border border-slate-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all"
            >
              <Calculator className="h-5 w-5 text-blue-600" />
              <div className="text-left">
                <div className="font-medium text-slate-900">CVSS v4.0 Spec</div>
                <div className="text-xs text-slate-500">Latest specification (2023)</div>
              </div>
              <ExternalLink className="h-4 w-4 text-slate-400 ml-auto" />
            </a>
            <a
              href="https://www.first.org/cvss/calculator/3.1"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 p-3 bg-white border border-slate-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all"
            >
              <FileText className="h-5 w-5 text-emerald-600" />
              <div className="text-left">
                <div className="font-medium text-slate-900">CVSS v3.1 Spec</div>
                <div className="text-xs text-slate-500">Refined v3.0 standard</div>
              </div>
              <ExternalLink className="h-4 w-4 text-slate-400 ml-auto" />
            </a>
            <a
              href="https://www.first.org/cvss/calculator/3.0"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 p-3 bg-white border border-slate-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all"
            >
              <FileText className="h-5 w-5 text-emerald-600" />
              <div className="text-left">
                <div className="font-medium text-slate-900">CVSS v3.0 Spec</div>
                <div className="text-xs text-slate-500">Original modern standard</div>
              </div>
              <ExternalLink className="h-4 w-4 text-slate-400 ml-auto" />
            </a>
            <a
              href="https://www.first.org/cvss/calculator-information"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 p-3 bg-white border border-slate-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all"
            >
              <BookOpen className="h-5 w-5 text-purple-600" />
              <div className="text-left">
                <div className="font-medium text-slate-900">CVSS Documentation</div>
                <div className="text-xs text-slate-500">General scoring information</div>
              </div>
              <ExternalLink className="h-4 w-4 text-slate-400 ml-auto" />
            </a>
          </div>
        </div>
      </main>

      {/* Footer */}
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
