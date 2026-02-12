/**
 * v2e Portal - Dock Component
 *
 * Bottom dock with glass morphism effect
 * Renders without backend dependency
 */

'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Z_INDEX, WindowState } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';
import { ContextMenu, ContextMenuPresets, useContextMenu } from '@/components/desktop/ContextMenu';
import { getActiveApps, getAppById } from '@/lib/desktop/app-registry';
import type { AppRegistryEntry } from '@/lib/desktop/app-registry';

/**
 * Dock item component
 */
function DockItem({ app, isRunning, isIndicator }: {
  app: AppRegistryEntry;
  isRunning: boolean;
  isIndicator: boolean;
}) {
  const { openWindow, windows, minimizeWindow } = useDesktopStore();
  const contextMenu = useContextMenu();
  const existingWindow = Object.values(windows).find(w => w.appId === app.id);

  const handleClick = () => {
    if (existingWindow) {
      // Window exists - focus or minimize based on state
      if (existingWindow.isFocused) {
        // Focused window - minimize it
        minimizeWindow(existingWindow.id);
      } else {
        // Not focused - bring to front
        // Window already exists, just need to focus it
        const { focusWindow } = useDesktopStore.getState();
        focusWindow(existingWindow.id);
      }
    } else {
      // No window - open new
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
    }
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const preset = isRunning
      ? ContextMenuPresets.dockItemRunning(app.id)
      : ContextMenuPresets.dockItemNotRunning(app.id);
    contextMenu.show(e.clientX, e.clientY, preset);
  };

  return (
    <button
      onClick={handleClick}
      onContextMenu={handleContextMenu}
      className="relative flex flex-col items-center p-2 rounded-lg hover:scale-110 transition-all duration-200"
      aria-label={`${isRunning ? 'Focus' : 'Launch'} ${app.name}`}
      title={`${isRunning ? 'Focus' : 'Launch'} ${app.name}`}
    >
      {/* App icon with app color */}
      <div
        className="w-10 h-10 rounded-full flex items-center justify-center"
        style={{ backgroundColor: app.iconColor || '#3b82f6' }}
      >
        <span className="text-white text-lg font-bold">
          {app.name[0]}
        </span>
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
 * Fixed at bottom with glass morphism and auto-hide
 */
export function Dock() {
  const { dock, windows, setDockVisibility } = useDesktopStore();
  const contextMenu = useContextMenu();
  const [isHovering, setIsHovering] = useState(false);
  const hideTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const dockRef = useRef<HTMLDivElement>(null);

  const sizeClasses = {
    small: 'h-12',
    medium: 'h-20',
    large: 'h-24',
  };

  // Get apps from registry
  const registryApps = getActiveApps();

  // Build dock items with running state
  const dockItems = registryApps.map(app => {
    const isRunning = Object.values(windows).some(w => w.appId === app.id);
    return {
      app,
      isRunning,
      isIndicator: isRunning,
    };
  });

  // Auto-hide logic
  const handleMouseEnter = useCallback(() => {
    setIsHovering(true);
    if (hideTimeoutRef.current) {
      clearTimeout(hideTimeoutRef.current);
      hideTimeoutRef.current = null;
    }
    setDockVisibility(true);
  }, [setDockVisibility]);

  const handleMouseLeave = useCallback(() => {
    setIsHovering(false);
    if (dock.autoHide) {
      hideTimeoutRef.current = setTimeout(() => {
        setDockVisibility(false);
      }, dock.autoHideDelay);
    }
  }, [dock.autoHide, dock.autoHideDelay, setDockVisibility]);

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (hideTimeoutRef.current) {
        clearTimeout(hideTimeoutRef.current);
      }
    };
  }, []);

  // Show dock when mouse moves to bottom edge
  useEffect(() => {
    if (!dock.autoHide) return;

    const handleMouseMove = (e: MouseEvent) => {
      const threshold = 50; // Distance from bottom edge to trigger reveal
      if (window.innerHeight - e.clientY < threshold) {
        setDockVisibility(true);
      }
    };

    document.addEventListener('mousemove', handleMouseMove);
    return () => document.removeEventListener('mousemove', handleMouseMove);
  }, [dock.autoHide, setDockVisibility]);

  // Ensure dock is visible on mount (fixes hidden dock from persisted state)
  useEffect(() => {
    // Only force visible if dock was hidden and auto-hide is disabled
    if (!dock.isVisible && !dock.autoHide) {
      setDockVisibility(true);
    }
  }, []); // Run once on mount

  // Should show dock (either visible or hovering during auto-hide)
  const shouldShow = dock.isVisible || (dock.autoHide && isHovering);

  if (!shouldShow) {
    return null;
  }

  return (
    <>
      <nav
        ref={dockRef}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        className={`
          fixed bottom-4 left-1/2 right-1/2 -translate-x-1/2
          bg-white/10 backdrop-blur-md border border-white/10 rounded-2xl
          flex items-end justify-center gap-1 p-2
          transition-transform duration-300
          ${dock.autoHide ? 'hover:scale-105' : ''}
        `}
        style={{
          zIndex: Z_INDEX.DOCK,
        }}
        role="navigation"
        aria-label="Application dock"
      >
        <div className={sizeClasses[dock.size] + ' flex items-end gap-2'}>
          {dockItems.map((item, index) => (
            <DockItem
              key={`${item.app.id}-${index}`}
              app={item.app}
              isRunning={item.isRunning}
              isIndicator={item.isIndicator}
            />
          ))}
        </div>
      </nav>

      {/* Dock context menu */}
      <ContextMenu
        isVisible={contextMenu.isVisible}
        position={contextMenu.position}
        items={contextMenu.items}
        onClose={contextMenu.hide}
      />
    </>
  );
}
