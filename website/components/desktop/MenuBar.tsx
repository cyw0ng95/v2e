/**
 * v2e Portal - Menu Bar Component
 *
 * Desktop-top menu bar with glassmorphism effect
 * Renders without backend dependency
 */

'use client';

import React, { useEffect, useState } from 'react';
import { Z_INDEX } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';
import { useTheme } from 'next-themes';

/**
 * Menu bar component
 * Fixed at top, always on top (z-index: 2000)
 */
export function MenuBar() {
  const { dock, setDockAutoHide } = useDesktopStore();
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

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

        {/* Right side: Controls */}
        <div className="flex items-center gap-3">
          {/* Theme toggle */}
          {mounted && (
            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 rounded-lg hover:bg-white/20 transition-colors"
              aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
              title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            >
              {theme === 'dark' ? (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 3v1m0 8v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 18 0z"
                  />
                </svg>
              ) : (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M21 12.79A9 9 0 1111.21 3 7 0 0021 12.79z"
                  />
                </svg>
              )}
            </button>
          )}

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
                d="M4 6a2 2 0 012 2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d={dock.autoHide ? "M15 9l6 6m-6 6v12m0-12v-12" : "M9 15l6-6m-6 6v12m6-6l6-6"}
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
                d="M12 15a3 3 0 11-6 0 3 3 0 0 16 0zm5.25-5.25a5.25 5.25 0 01-7.07 0 5.25 5.25 0 007.07 7.07zm-9.9 1.95a9.9 9.9 0 0111.3 0 9.9 9.9 0 01-11.3 0"
              />
            </svg>
          </button>
        </div>
      </div>
    </header>
  );
}
