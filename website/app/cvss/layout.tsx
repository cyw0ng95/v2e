import React from 'react';

/**
 * CVSS Calculator Layout
 * Shared layout for CVSS calculator pages with header and navigation
 */

import Link from 'next/link';
import { ArrowLeft, BookOpen, Github } from 'lucide-react';
import { cn } from '@/lib/utils';

export default function CVSSLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Top Navigation Bar */}
      <nav
        className="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-slate-200"
        aria-label="Main navigation"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Left: Logo and Back */}
            <div className="flex items-center gap-4">
              <Link
                href="/"
                className="inline-flex items-center gap-2 text-slate-600 hover:text-slate-900 transition-colors"
                aria-label="Go to home page"
              >
                <ArrowLeft className="h-5 w-5" />
                <span className="hidden sm:inline font-medium">Home</span>
              </Link>

              <div className="h-6 w-px bg-slate-300" />

              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                  <span className="text-white font-bold text-sm">CV</span>
                </div>
                <div>
                  <h1 className="font-bold text-slate-900">CVSS Calculator</h1>
                  <p className="text-xs text-slate-500 hidden sm:block">
                    Common Vulnerability Scoring System
                  </p>
                </div>
              </div>
            </div>

            {/* Right: Navigation Links */}
            <div className="flex items-center gap-2">
              {/* Documentation Link */}
              <a
                href="https://www.first.org/cvss/calculator/3.1"
                target="_blank"
                rel="noopener noreferrer"
                className={cn(
                  'hidden md:inline-flex items-center gap-2 px-3 py-2',
                  'rounded-lg text-sm font-medium transition-all duration-200',
                  'bg-slate-100 text-slate-700 hover:bg-slate-200'
                )}
                aria-label="View CVSS specification (opens in new tab)"
              >
                <BookOpen className="h-4 w-4" />
                <span>Specification</span>
              </a>

              {/* GitHub Link */}
              <a
                href="https://github.com/cyw0ng95/v2e"
                target="_blank"
                rel="noopener noreferrer"
                className={cn(
                  'inline-flex items-center gap-2 px-3 py-2',
                  'rounded-lg text-sm font-medium transition-all duration-200',
                  'bg-slate-900 text-white hover:bg-slate-800'
                )}
                aria-label="View source code on Github (opens in new tab)"
              >
                <Github className="h-4 w-4" />
                <span className="hidden sm:inline">Github</span>
              </a>
            </div>
          </div>
        </div>
      </nav>

      {/* Breadcrumb Navigation */}
      <div className="bg-white border-b border-slate-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-2">
          <nav
            className="flex items-center gap-2 text-sm"
            aria-label="Breadcrumb navigation"
          >
            <Link
              href="/"
              className="text-slate-500 hover:text-slate-700 transition-colors"
            >
              Home
            </Link>
            <span className="text-slate-300" aria-hidden="true">/</span>
            <span className="text-slate-900 font-medium">CVSS Calculator</span>
            <span className="ml-auto text-xs text-slate-400">
              Supports CVSS v3.0, v3.1, and v4.0
            </span>
          </nav>
        </div>
      </div>

      {/* Skip Links for Accessibility */}
      <div className="sr-only focus-within:not-sr-only">
        <a
          href="#main-content"
          className="absolute top-20 left-4 px-4 py-2 bg-blue-600 text-white rounded-lg z-50"
        >
          Skip to main content
        </a>
        <a
          href="#cvss-calculator"
          className="absolute top-28 left-4 px-4 py-2 bg-blue-600 text-white rounded-lg z-50"
        >
          Skip to calculator
        </a>
      </div>

      {/* Main Content */}
      <main id="main-content" className="max-w-7xl mx-auto">
        {children}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200 bg-white/80 backdrop-blur-sm mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="grid md:grid-cols-3 gap-8">
            {/* About Section */}
            <div>
              <h3 className="font-semibold text-slate-900 mb-3">About CVSS</h3>
              <p className="text-sm text-slate-600 leading-relaxed">
                The Common Vulnerability Scoring System (CVSS) provides a way to capture
                the principal characteristics of a vulnerability and produce a numerical score
                reflecting its severity.
              </p>
            </div>

            {/* Resources Section */}
            <div>
              <h3 className="font-semibold text-slate-900 mb-3">Resources</h3>
              <ul className="space-y-2">
                <li>
                  <a
                    href="https://www.first.org/cvss/calculator/3.1"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-blue-600 hover:text-blue-700 transition-colors"
                  >
                    CVSS v3.1 Specification
                  </a>
                </li>
                <li>
                  <a
                    href="https://www.first.org/cvss/calculator/4.0"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-blue-600 hover:text-blue-700 transition-colors"
                  >
                    CVSS v4.0 Specification
                  </a>
                </li>
                <li>
                  <a
                    href="https://www.first.org/cvss/support-documentation"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-blue-600 hover:text-blue-700 transition-colors"
                  >
                    CVSS Support Documentation
                  </a>
                </li>
              </ul>
            </div>

            {/* Version Info Section */}
            <div>
              <h3 className="font-semibold text-slate-900 mb-3">Supported Versions</h3>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-green-500" />
                  <span className="text-sm text-slate-600">CVSS v3.0</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-green-500" />
                  <span className="text-sm text-slate-600">CVSS v3.1</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-green-500" />
                  <span className="text-sm text-slate-600">CVSS v4.0</span>
                </div>
              </div>
            </div>
          </div>

          {/* Bottom Bar */}
          <div className="mt-8 pt-8 border-t border-slate-200">
            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
              <p className="text-sm text-slate-500">
                CVSS Calculator following FIRST.org specifications. Scores are calculated client-side.
              </p>
              <p className="text-xs text-slate-400">
                Part of the{' '}
                <a
                  href="https://github.com/cyw0ng95/v2e"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700"
                >
                  v2e project
                </a>
              </p>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
