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
import { Dock } from '@/components/desktop/Dock';
import { QuickLaunchModal, useQuickLaunchShortcut } from '@/components/desktop/QuickLaunchModal';
import { WindowManager } from '@/components/desktop/WindowManager';
import { useDesktopStore } from '@/lib/desktop/store';
import { useNetworkStatus } from '@/lib/hooks/useNetworkStatus';
import { DndContext } from '@dnd-kit/core';
import { getAppById } from '@/lib/desktop/app-registry';
import { WindowState } from '@/types/desktop';

/**
 * SPA Root Page - Desktop Application
 * Orchestrates all desktop components and handles deep linking
 */
export default function HomePage() {
  const searchParams = useSearchParams();
  const { desktopIcons, openWindow } = useDesktopStore();
  const quickLaunch = useQuickLaunchShortcut();
  const hasOpenedAppRef = useRef(false);

  // Initialize network status detection
  useNetworkStatus();

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
      <div className="h-screen w-screen overflow-hidden bg-gray-100">
        {/* Menu Bar - Always on top */}
        <MenuBar />

        {/* Desktop Area - Main workspace */}
        <DesktopArea />

        {/* Window Manager - Handles all windows */}
        <WindowManager />

        {/* Dock - Bottom navigation with React Bits animation */}
        <Dock />

        {/* Quick Launch Modal - Cmd+K triggered */}
        <QuickLaunchModal
          isVisible={quickLaunch.isVisible}
          onClose={quickLaunch.hide}
        />

        {/* Initial state notice - shown when no icons exist */}
        {desktopIcons.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div className="bg-white/10 backdrop-blur-sm rounded-lg p-8 text-center max-w-md">
              <h2 className="text-xl font-semibold text-gray-900 mb-2">
                Welcome to v2e Portal
              </h2>
              <p className="text-gray-600 mb-4">
                Desktop is ready. Add apps from dock or right-click to customize.
              </p>
              <div className="text-sm text-gray-500">
                <p>Backend Status: <span className="text-green-600 font-medium">Not Required</span></p>
                <p className="mt-1">All features work offline</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </DndContext>
  );
}
