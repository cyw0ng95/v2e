/**
 * v2e Portal - Dock Component (React Bits Integration)
 *
 * Bottom dock with glass morphism effect using React Bits Dock component
 * Integrates with desktop store for window management
 */

'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import {
  motion,
  MotionValue,
  useMotionValue,
  useSpring,
  useTransform,
  AnimatePresence
} from 'motion/react';
import { Z_INDEX, WindowState } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';
import { ContextMenu, ContextMenuPresets, useContextMenu } from '@/components/desktop/ContextMenu';
import { getActiveApps } from '@/lib/desktop/app-registry';
import type { AppRegistryEntry } from '@/lib/desktop/app-registry';
import {
  Star,
  Bug,
  Crosshair,
  Activity,
  GitGraph,
  BookOpen,
  Zap,
  Bookmark,
  Folder,
  Sparkles,
  Calculator,
  Grid,
  Square,
  Heart,
} from 'lucide-react';

/**
 * Map app icon names from app-registry to icon components
 */
function getAppIcon(appId: string): React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }> {
  const iconMap: Record<string, React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>> = {
    cve: Star,           // Shield alternative
    cwe: Bug,            // Direct match
    capec: Crosshair,     // Target icon
    attack: Crosshair,     // Crosshair icon
    cvss: Calculator,     // Calculator
    glc: GitGraph,        // Git-graph alternative
    mcards: BookOpen,     // Library alternative
    etl: Activity,        // Activity alternative
    bookmarks: Bookmark,   // Bookmark icon
  };

  return iconMap[appId] || Star; // Default to Star if no match
}

/**
 * Dock item component with React Bits magnification animation
 */
interface DockItemProps {
  app: AppRegistryEntry;
  isRunning: boolean;
  isIndicator: boolean;
  mouseX: MotionValue<number>;
  spring: { mass: number; stiffness: number; damping: number };
  distance: number;
  baseItemSize: number;
  magnification: number;
}

function DockItem({
  app,
  isRunning,
  isIndicator,
  mouseX,
  spring,
  distance,
  baseItemSize,
  magnification
}: DockItemProps) {
  const { openWindow, windows, minimizeWindow } = useDesktopStore();
  const contextMenu = useContextMenu();
  const existingWindow = Object.values(windows).find(w => w.appId === app.id);
  const ref = useRef<HTMLDivElement>(null);
  const isHovered = useMotionValue(0);

  // Get icon component for this app
  const AppIcon = getAppIcon(app.id);

  const handleClick = useCallback(() => {
    if (existingWindow) {
      // Window exists - handle based on current state
      if (existingWindow.isMinimized) {
        // Minimized window - restore and focus it
        const { restoreWindow } = useDesktopStore.getState();
        restoreWindow(existingWindow.id);
      } else if (existingWindow.isFocused) {
        // Focused window - minimize it
        minimizeWindow(existingWindow.id);
      } else {
        // Not focused - bring to front
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
  }, [app, existingWindow, openWindow, minimizeWindow]);

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const preset = isRunning
      ? ContextMenuPresets.dockItemRunning(app.id)
      : ContextMenuPresets.dockItemNotRunning(app.id);
    contextMenu.show(e.clientX, e.clientY, preset);
  }, [isRunning, app.id, contextMenu]);

  const mouseDistance = useTransform(mouseX, val => {
    const rect = ref.current?.getBoundingClientRect() ?? {
      x: 0,
      width: baseItemSize
    };
    return val - rect.x - baseItemSize / 2;
  });

  const targetSize = useTransform(mouseDistance, [-distance, 0, distance], [baseItemSize, magnification, baseItemSize]);
  const size = useSpring(targetSize, spring);

  return (
    <motion.div
      ref={ref}
      style={{ width: size, height: size }}
      onHoverStart={() => isHovered.set(1)}
      onHoverEnd={() => isHovered.set(0)}
      onClick={handleClick}
      onContextMenu={handleContextMenu}
      className="relative inline-flex items-center justify-center"
      tabIndex={0}
      role="button"
      aria-label={`${isRunning ? 'Focus' : 'Launch'} ${app.name}`}
      title={`${isRunning ? 'Focus' : 'Launch'} ${app.name}`}
    >
      {/* App icon */}
      <AppIcon
        size={40}
        className="opacity-90 hover:opacity-100 transition-opacity"
        style={{ color: app.iconColor || '#3b82f6' }}
      />

      {/* Active indicator */}
      <AnimatePresence>
        {isIndicator && isRunning && (
          <motion.div
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0 }}
            className="absolute -bottom-1 w-1.5 h-1.5 rounded-full bg-blue-500"
          />
        )}
      </AnimatePresence>
    </motion.div>
  );
}

/**
 * Dock component using React Bits animation system
 * Provides macOS-style magnification effect
 */
export function Dock() {
  const { dock, windows, setDockVisibility } = useDesktopStore();
  const contextMenu = useContextMenu();
  const [isHovering, setIsHovering] = useState(false);
  const hideTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // React Bits Dock spring configuration
  const spring = { mass: 0.1, stiffness: 150, damping: 12 };
  const magnification = 70;
  const distance = 200;
  const panelHeight = 68;

  const mouseX = useMotionValue(Infinity);
  const isHovered = useMotionValue(0);

  // Get apps from registry
  const registryApps = getActiveApps();

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
      <motion.div
        onMouseMove={({ pageX }) => {
          isHovered.set(1);
          mouseX.set(pageX);
        }}
        onMouseLeave={() => {
          isHovered.set(0);
          mouseX.set(Infinity);
        }}
        onMouseEnter={handleMouseEnter}
        className={`
          fixed bottom-4 left-1/2 transform -translate-x-1/2
          backdrop-blur-lg bg-white/80 dark:bg-slate-900/80 border border-border/60 rounded-2xl shadow-lg
          flex items-end justify-center gap-2 p-2
          transition-transform duration-300
          ${dock.autoHide ? 'hover:scale-105' : ''}
        `}
        style={{
          zIndex: Z_INDEX.DOCK,
          height: panelHeight,
        }}
        role="navigation"
        aria-label="Application dock"
      >
        {registryApps.map((app, index) => {
          const isRunning = Object.values(windows).some(w => w.appId === app.id);
          return (
            <DockItem
              key={`${app.id}-${index}`}
              app={app}
              isRunning={isRunning}
              isIndicator={isRunning}
              mouseX={mouseX}
              spring={spring}
              distance={distance}
              magnification={magnification}
              baseItemSize={50}
            />
          );
        })}
      </motion.div>

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
