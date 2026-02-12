/**
 * v2e Portal - Dock Component
 *
 * Bottom dock with glass morphism effect
 * Renders without backend dependency
 */

'use client';

import React from 'react';
import { Z_INDEX } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';

/**
 * Dock item component
 */
function DockItem({ appId, isRunning, isIndicator }: {
  appId: string;
  isRunning: boolean;
  isIndicator: boolean;
}) {
  const { openWindow, windows } = useDesktopStore();
  const window = windows[appId];

  const handleClick = () => {
    if (window) {
      // Window exists - focus or minimize based on state
      if (window.isFocused) {
        // Focused window - minimize it
        // TODO: Will be implemented in Phase 2
        console.log('Minimize window:', appId);
      } else {
        // Not focused - bring to front
        openWindow({
          appId,
          title: appId.charAt(0).toUpperCase() + appId.slice(1),
          position: { x: 100, y: 100 },
          size: { width: 1200, height: 800 },
          minWidth: 800,
          minHeight: 600,
        });
      }
    } else {
      // No window - open new
      openWindow({
        appId,
        title: appId.charAt(0).toUpperCase() + appId.slice(1),
        position: { x: 100, y: 100 },
        size: { width: 1200, height: 800 },
        minWidth: 800,
        minHeight: 600,
      });
    }
  };

  return (
    <button
      onClick={handleClick}
      className="relative flex flex-col items-center p-2 rounded-lg hover:scale-110 transition-all duration-200"
      aria-label={`${isRunning ? 'Focus' : 'Launch'} ${appId}`}
      title={`${isRunning ? 'Focus' : 'Launch'} ${appId}`}
    >
      {/* App icon placeholder */}
      <div className="w-10 h-10 rounded-full bg-gradient-to-br from-gray-700 to-gray-900 flex items-center justify-center">
        <svg className="w-6 h-6" fill="none" stroke="white" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9.75 17L9 20l-1.5-1.5L6 16.25l-3 3.75-3.75-2.15.25z"
          />
        </svg>
      </div>

      {/* Active indicator */}
      {isIndicator && isRunning && (
        <div className="absolute -bottom-1 w-1.5 h-1.5 rounded-full bg-blue-500" />
      )}
    </button>
  );
}

/**
 * Dock component
 * Fixed at bottom with glass morphism
 */
export function Dock() {
  const { dock } = useDesktopStore();
  const sizeClasses = {
    small: 'h-12',
    medium: 'h-20',
    large: 'h-24',
  };

  // Default dock items - will be replaced with APP_REGISTRY in Phase 4
  const defaultItems = [
    { appId: 'cve', isRunning: false, isIndicator: true },
    { appId: 'cwe', isRunning: false, isIndicator: true },
    { appId: 'capec', isRunning: false, isIndicator: true },
    { appId: 'attack', isRunning: false, isIndicator: true },
  ];

  if (!dock.isVisible) {
    return null; // Auto-hide - will be implemented in Phase 3
  }

  return (
    <nav
      className={`
        fixed bottom-4 left-1/2 right-1/2 -translate-x-1/2
        bg-white/10 backdrop-blur-md border border-white/10 rounded-2xl
        flex items-end justify-center gap-1 p-2
        transition-transform duration-300
      `}
      style={{
        zIndex: Z_INDEX.DOCK,
      }}
      role="navigation"
      aria-label="Application dock"
    >
      <div className={sizeClasses[dock.size] + ' flex items-end gap-2'}>
        {defaultItems.map((item, index) => (
          <DockItem
            key={`${item.appId}-${index}`}
            appId={item.appId}
            isRunning={item.isRunning}
            isIndicator={item.isIndicator}
          />
        ))}
      </div>
    </nav>
  );
}
