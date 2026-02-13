/**
 * v2e Portal - Dock Thumbnails Component
 *
 * Displays minimized window thumbnails in dock
 * Phase 2/3: Window-Dock Integration
 */

'use client';

import React from 'react';
import { useDesktopStore } from '@/lib/desktop/store';
import type { WindowConfig } from '@/types/desktop';

/**
 * Dock thumbnail component
 * Shows minimized window preview
 */
function DockThumbnail({ window }: { window: WindowConfig }) {
  const { restoreWindow } = useDesktopStore();

  const handleClick = () => {
    restoreWindow(window.id);
  };

  return (
    <button
      onClick={handleClick}
      className="relative w-12 h-12 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center hover:scale-110 transition-transform"
      aria-label={`Restore ${window.appId} window`}
      title={`Restore ${window.appId} window`}
    >
      {/* Window preview icon */}
      <div className="w-6 h-6 rounded bg-gray-100 flex items-center justify-center">
        <svg className="w-3 h-3" fill="none" stroke="white" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9.75 17L9 20l-1.5-1.5L6 16.25l-3 3.75-3.75-2.15.25z"
          />
        </svg>
      </div>
    </button>
  );
}

/**
 * Enhanced dock with thumbnails
 */
export function DockWithThumbnails() {
  const { dock, windows } = useDesktopStore();
  const minimizedWindows = Object.values(windows).filter(w => w.state === 'minimized');

  return (
    <nav
      className={`
        fixed bottom-4 left-1/2 right-1/2 -translate-x-1/2
        bg-white/10 backdrop-blur-md border border-white/10 rounded-2xl
        flex items-end justify-center gap-1 p-2
        transition-transform duration-300
      `}
      style={{ zIndex: 50 }}
      role="navigation"
      aria-label="Application dock with thumbnails"
    >
      <div className={dock.size === 'small' ? 'h-12' : dock.size === 'medium' ? 'h-16' : 'h-20'}>
        {dock.items.map((item, index) => {
          const window = minimizedWindows.find(w => w.appId === item.appId);

          return (
            <div key={`${item.appId}-${index}`} className="flex flex-col items-center gap-3">
              {/* Active app indicator */}
              {item.isRunning && (
                <div className="absolute -bottom-1 w-1.5 h-1.5 rounded-full bg-blue-500" />
              )}

              {/* Window thumbnail for minimized windows */}
              {window && <DockThumbnail window={window} />}

              {/* Regular dock item */}
              {!window && (
                <button
                  onClick={() => {
                    // Launch window (will be implemented in next iteration)
                    console.log('Launch:', item.appId);
                  }}
                  className="relative flex flex-col items-center gap-2 p-2 rounded-lg hover:bg-blue-500/20 transition-transform"
                  aria-label={`${item.isRunning ? 'Focus' : 'Launch'} ${item.appId}`}
                >
                  {/* App icon */}
                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-gray-700 to-gray-900">
                    <svg className="w-4 h-4" fill="none" stroke="white" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9.75 17L9 20l-1.5-1.5L6 16.25l-3 3.75-3.75-2.15.25z"
                      />
                    </svg>
                  </div>

                  {/* App label */}
                  <span className="text-xs font-medium text-gray-700">
                    {item.appId}
                  </span>
                </button>
              )}
            </div>
          );
        })}
      </div>
    </nav>
  );
}
