/**
 * v2e Portal - Menu Bar Component
 *
 * Desktop-top menu bar with glass morphism effect
 * Renders without backend dependency
 */

'use client';

import React from 'react';
import { Z_INDEX } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';

/**
 * Menu bar component
 * Fixed at top, always on top (z-index: 2000)
 */
export function MenuBar() {
  const { theme, setThemeMode, dock, setDockAutoHide } = useDesktopStore();

  return (
    <header
      className="fixed top-0 left-0 right-0 h-7 bg-white/10 backdrop-blur-md border-b border-white/10"
      style={{ zIndex: Z_INDEX.MENU_BAR }}
      role="banner"
    >
      <div className="flex items-center justify-between px-4 h-full">
        {/* Left side: Apple logo placeholder */}
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded-full bg-gradient-to-br from-gray-700 to-gray-900" />
          <span className="text-sm font-medium text-gray-700">v2e</span>
        </div>

        {/* Center: Search placeholder */}
        <div className="flex-1 flex justify-center">
          <div className="w-96 h-5 bg-white/20 rounded-lg flex items-center px-3 text-gray-400 text-sm">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 11 14 0m0 0l-7 7 7 7v0m0 0l7-7-7"
              />
            </svg>
            <span className="ml-2">Search</span>
          </div>
        </div>

        {/* Right side: Controls */}
        <div className="flex items-center gap-3">
          {/* Theme toggle */}
          <button
            onClick={() => setThemeMode(theme.mode === 'dark' ? 'light' : 'dark')}
            className="p-2 rounded-lg hover:bg-white/20 transition-colors"
            aria-label={`Switch to ${theme.mode === 'dark' ? 'light' : 'dark'} mode`}
            title={`Switch to ${theme.mode === 'dark' ? 'light' : 'dark'} mode`}
          >
            {theme.mode === 'dark' ? (
              // Sun icon
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 3v1m0 0v1m0-6h6a6 6 0 016 0v12a6 6 0 01-6 0H6a6 6 0 01-6 0V9a6 6 0 016 0z"
                />
              </svg>
            ) : (
              // Moon icon
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M21 12.79A9 9 0 1111.21 3 7 7 3 0 0-9-9-9 0 0 01-12.65 0 4 4 0 01-7-7 7 0 0 12-12.65 0 13-15a9 9 0 01-9-9-9 0 015-15 0 7 7 0 012 12.65 0 15 0-13-15 0 7-7 7 0zm-3.91-3.91a6 6 0 01-5.29-5.29 0 0-8.48-8.48 0 015.17 0 0-3.86 3.86 0 01-4.13-4.13 0 015.17 0 01-8.48-8.48 0 01-4.13-4.13 0 015.17 0 013.41 13.41 0 7-7 7 0 01-5.29-5.29 0 015.17 0 017.65 0 017.65 0 7-7 7 0z"
                />
              </svg>
            )}
          </button>

          {/* Dock auto-hide toggle */}
          <button
            onClick={() => setDockAutoHide(!dock.autoHide)}
            className={`p-2 rounded-lg transition-colors ${dock.autoHide ? 'bg-blue-500/20 hover:bg-blue-500/30' : 'hover:bg-white/20'}`}
            aria-label={dock.autoHide ? 'Disable dock auto-hide' : 'Enable dock auto-hide'}
            title={dock.autoHide ? 'Dock auto-hide enabled' : 'Dock auto-hide disabled'}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 12v8a2 2 0 002-2V8a2 2 0 00-2-2H6a2 2 0 00-2 2v8a2 2 0 002 2v8a2 2 0 01-2-2h16a2 2 0 002-2v-8a2 2 0 00-2-2H6a2 2 0 00-2 2z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d={dock.autoHide ? "M9 9l6 6m6-6v12m-6-6l6 6" : "M9 15l6-6m-6 6v12m6-6l6-6"}
              />
            </svg>
          </button>

          {/* Settings placeholder */}
          <button
            className="p-2 rounded-lg hover:bg-white/20 transition-colors"
            aria-label="Settings"
            title="Settings"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10.325 4.317c.426-1.601.234-2.417 1.601-2.417-.925-.069a3.001 3.001 0 010 4.828-4.828 13.256 13.256 0 01-4.828 4.828-013.256 13.256c-.925 0-1.601-.234-2.417.828-4.828-.069-.925 0-1.601.234-2.417.925-.069 2.417-1.601-.069.925.234 2.417.069.925 4.828v13.256c0 1.601.234 2.417.069.925 0 4.828 4.828 13.256 13.256 0 01-4.828 4.828 0 013.256-13.256c.925 0 1.601.234 2.417.069 2.417-.925.069a3.001 3.001 0 014.828 14.828c0 .925 1.601.234 2.417.069.925 0 1.601.234 2.417-.069.925 2.417-1.601-.069.925.234 2.417.069.925-4.828-4.828.013.256-13.256 01-4.828 4.828 0-13.256 13.256c.925 0 1.601.234 2.417.069 2.417-.069.925 2.417-1.601-.069z"
              />
            </svg>
          </button>
        </div>
      </div>
    </header>
  );
}
