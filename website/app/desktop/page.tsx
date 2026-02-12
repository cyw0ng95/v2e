/**
 * v2e Portal - Desktop Page
 *
 * Main desktop entry point
 * Works completely without backend dependency
 */

import { MenuBar } from '@/components/desktop/MenuBar';
import { DesktopArea } from '@/components/desktop/DesktopArea';
import { DockDraggable } from '@/components/desktop/DockDraggable';
import { QuickLaunchModal, useQuickLaunchShortcut } from '@/components/desktop/QuickLaunchModal';
import { WindowManager } from '@/components/desktop/WindowManager';
import { useDesktopStore } from '@/lib/desktop/store';
import { DndContext } from '@dnd-kit/core';

/**
 * Desktop page component
 * Orchestrates all desktop components
 */
export default function DesktopPage() {
  const { desktopIcons } = useDesktopStore();
  const quickLaunch = useQuickLaunchShortcut();

  return (
    <DndContext>
      <div className="h-screen w-screen overflow-hidden bg-gray-100">
        {/* Menu Bar - Always on top */}
        <MenuBar />

        {/* Desktop Area - Main workspace */}
        <DesktopArea />

        {/* Window Manager - Handles all windows */}
        <WindowManager />

        {/* Dock - Bottom navigation with drag-to-reorder */}
        <DockDraggable />

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
