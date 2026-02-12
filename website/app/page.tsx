/**
 * v2e Portal - Single Page Application Entry
 *
 * Main SPA entry point that renders the desktop UI
 * Supports deep linking via URL query parameters (e.g., ?app=cvss)
 */

'use client';

import { useEffect, useRef } from 'react';
import { useSearchParams } from 'next/navigation';
import { MenuBar } from '@/components/desktop/MenuBar';
import { DesktopArea } from '@/components/desktop/DesktopArea';
import ChromaGridDock from '@/components/desktop/ChromaGridDock';
import { QuickLaunchModal, useQuickLaunchShortcut } from '@/components/desktop/QuickLaunchModal';
import { WindowManager } from '@/components/desktop/WindowManager';
import { useDesktopStore } from '@/lib/desktop/store';
import { useNetworkStatus } from '@/lib/hooks/useNetworkStatus';
import { DndContext } from '@dnd-kit/core';
import { getAppById } from '@/lib/desktop/app-registry';
import { WindowState } from '@/types/desktop';
import Threads from '@/components/backgrounds/Threads';
import { useTheme } from 'next-themes';

/**
 * SPA Root Page - Desktop Application
 * Orchestrates all desktop components and handles deep linking
 */
export default function HomePage() {
  const searchParams = useSearchParams();
  const { desktopIcons, openWindow } = useDesktopStore();
  const quickLaunch = useQuickLaunchShortcut();
  const hasOpenedAppRef = useRef(false);
  const { theme } = useTheme();

  // Initialize network status detection
  useNetworkStatus();

  // Determine background class based on theme
  const bgClass = theme === 'light' ? 'bg-black' : 'bg-white';

  // Handle deep linking via URL query parameters
  useEffect(() => {
    const appParam = searchParams.get('app');

    // Only attempt to open app once
    if (appParam && !hasOpenedAppRef.current) {
      const app = getAppById(appParam);
      if (app) {
        // Mark that we've attempted to open
        hasOpenedAppRef.current = true;

        // Small delay to ensure store is hydrated from localStorage
        setTimeout(() => {
          openWindow({
            appId: app.id,
            title: app.name,
            position: {
              x: Math.max(0, (window.innerWidth - app.defaultWidth) / 2),
              y: Math.max(28, (window.innerHeight - app.defaultHeight) / 2),
            },
            size: {
              width: app.defaultWidth,
              height: app.defaultHeight,
            },
            minWidth: app.minWidth,
            minHeight: app.minHeight,
            maxWidth: app.maxWidth,
            maxHeight: app.maxHeight,
            isFocused: true,
            isMinimized: false,
            isMaximized: false,
            state: WindowState.Open,
          });
        }, 100);
      }
    }
  }, [searchParams, openWindow]);

  return (
    <DndContext>
      <div className={`h-screen w-screen overflow-hidden relative transition-colors duration-300 ${bgClass}`}>
        {/* Threads Background - animated threads */}
        <Threads />
        {/* Menu Bar - Always on top */}
        <MenuBar />

        {/* Desktop Area - Main workspace */}
        <DesktopArea />

        {/* Window Manager - Handles all windows */}
        <WindowManager />

        {/* Dock - Bottom navigation with ChromaGrid effect */}
        <ChromaGridDock />

        {/* Quick Launch Modal - Cmd+K triggered */}
        <QuickLaunchModal
          isVisible={quickLaunch.isVisible}
          onClose={quickLaunch.hide}
        />

        {/* Initial state notice - shown when no icons exist */}
        {desktopIcons.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div className={`backdrop-blur-sm rounded-lg p-8 text-center max-w-md ${theme === 'light' ? 'bg-white/10 text-white' : 'bg-black/10 text-gray-900'}`}>
              <h2 className="text-xl font-semibold mb-2">
                Welcome to v2e Portal
              </h2>
              <p className={theme === 'light' ? 'text-white/80' : 'text-gray-600 mb-4'}>
                Desktop is ready. Add apps from dock or right-click to customize.
              </p>
              <div className={`text-sm ${theme === 'light' ? 'text-white/70' : 'text-gray-500'}`}>
                <p>Backend Status: <span className="font-medium">Not Required</span></p>
                <p className="mt-1">All features work offline</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </DndContext>
  );
}
