/**
 * v2e Portal - ChromaGrid Dock Component
 *
 * Chromatic aberration grid dock using React Bits ChromaGrid pattern
 * Integrates with desktop store for window management
 */

'use client';

import React, { useRef, useEffect, useState, useCallback } from 'react';
import { gsap } from 'gsap';
import { Z_INDEX, WindowState } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';
import { ContextMenu, ContextMenuPresets, useContextMenu } from '@/components/desktop/ContextMenu';
import { getActiveApps } from '@/lib/desktop/app-registry';
import type { AppRegistryEntry } from '@/types/desktop';

interface ChromaAppItem {
  id: string;
  name: string;
  gradient: string;
  borderColor: string;
}

interface ChromaGridDockProps {
  className?: string;
  radius?: number;
  damping?: number;
  fadeOut?: number;
  ease?: string;
  autoHide?: boolean;
  autoHideDelay?: number;
}

type SetterFn = (v: number | string) => void;

const iconComponents: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  cve: Star,
  cwe: Bug,
  capec: Crosshair,
  attack: Crosshair,
  cvss: Calculator,
  glc: GitGraph,
  mcards: BookOpen,
  etl: Activity,
  bookmarks: Bookmark,
  folder: Folder,
  sparkles: Sparkles,
  grid: Grid,
  square: Square,
  heart: Heart,
};

const ChromaGridDock: React.FC<ChromaGridDockProps> = ({
  className = '',
  radius = 250,
  damping = 0.35,
  fadeOut = 0.4,
  ease = 'power2.out',
  autoHide = false,
  autoHideDelay = 3000,
}) => {
  const { windows, openWindow, minimizeWindow, restoreWindow, focusWindow, dock, setDockVisibility } = useDesktopStore();
  const contextMenu = useContextMenu();
  const [isHovering, setIsHovering] = useState(false);
  const hideTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const rootRef = useRef<HTMLDivElement>(null);
  const fadeRef = useRef<HTMLDivElement>(null);
  const setX = useRef<SetterFn | null>(null);
  const setY = useRef<SetterFn | null>(null);
  const pos = useRef({ x: 0, y: 0 });

  // Convert apps to chroma items
  const registryApps = getActiveApps();
  const items: ChromaAppItem[] = registryApps.map((app) => {
    const defaultStyle = { gradient: 'linear-gradient(145deg, #3b82f6, #000)', borderColor: '#3b82f6' };
    const style = iconGradients[app.id] || defaultStyle;

    return {
      id: app.id,
      name: app.name,
      gradient: style.gradient,
      borderColor: style.borderColor,
    };
  });

  // Get running app IDs
  const runningAppIds = new Set(Object.values(windows).map(w => w.appId));

  useEffect(() => {
    const el = rootRef.current;
    if (!el) return;
    setX.current = gsap.quickSetter(el, '--x', 'px') as SetterFn;
    setY.current = gsap.quickSetter(el, '--y', 'px') as SetterFn;
    const { width, height } = el.getBoundingClientRect();
    pos.current = { x: width / 2, y: height / 2 };
    setX.current(pos.current.x);
    setY.current(pos.current.y);
  }, []);

  const moveTo = (x: number, y: number) => {
    gsap.to(pos.current, {
      x,
      y,
      duration: damping,
      ease,
      onUpdate: () => {
        setX.current?.(pos.current.x);
        setY.current?.(pos.current.y);
      },
      overwrite: true
    });
  };

  const handleMove = (e: React.PointerEvent) => {
    const r = rootRef.current!.getBoundingClientRect();
    moveTo(e.clientX - r.left, e.clientY - r.top);
    gsap.to(fadeRef.current, { opacity: 0, duration: 0.2, overwrite: true });
  };

  const handleLeave = () => {
    gsap.to(fadeRef.current, {
      opacity: 1,
      duration: fadeOut,
      overwrite: true
    });

    // Auto-hide logic
    if (autoHide) {
      hideTimeoutRef.current = setTimeout(() => {
        setDockVisibility(false);
      }, autoHideDelay);
    }
  };

  const handleMouseEnter = useCallback(() => {
    setIsHovering(true);
    if (hideTimeoutRef.current) {
      clearTimeout(hideTimeoutRef.current);
      hideTimeoutRef.current = null;
    }
    setDockVisibility(true);
  }, [setDockVisibility, autoHide]);

  // Ensure dock is visible on mount
  useEffect(() => {
    if (!dock.isVisible && !autoHide) {
      setDockVisibility(true);
    }
  }, []);

  // Should show dock (either visible or hovering during auto-hide)
  const shouldShow = dock.isVisible || (autoHide && isHovering);

  const handleAppClick = (item: ChromaAppItem) => {
    const existingWindow = Object.values(windows).find(w => w.appId === item.id);

    if (existingWindow) {
      if (existingWindow.isMinimized) {
        restoreWindow(existingWindow.id);
      } else if (existingWindow.isFocused) {
        minimizeWindow(existingWindow.id);
      } else {
        focusWindow(existingWindow.id);
      }
    } else {
      const app = registryApps.find(a => a.id === item.id);
      if (app) {
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
    }
  };

  const handleContextMenu = useCallback((e: React.MouseEvent, appId: string) => {
    e.preventDefault();
    e.stopPropagation();
    const isRunning = runningAppIds.has(appId);
    const preset = isRunning
      ? ContextMenuPresets.dockItemRunning(appId)
      : ContextMenuPresets.dockItemNotRunning(appId);
    contextMenu.show(e.clientX, e.clientY, preset);
  }, [runningAppIds, contextMenu]);

  if (!shouldShow) {
    return null;
  }

  return (
    <>
      <div
        ref={rootRef}
        onPointerMove={handleMove}
        onPointerLeave={handleLeave}
        onMouseEnter={handleMouseEnter}
        className={`fixed bottom-4 left-1/2 transform -translate-x-1/2 ${className}`}
        style={
          {
            '--r': `${radius}px`,
            '--x': '50%',
            '--y': '50%',
            zIndex: Z_INDEX.DOCK,
            maxWidth: '95vw',
          } as React.CSSProperties
        }
      >
        <div className="relative w-full h-full flex flex-wrap justify-center items-start gap-3">
          {items.map((item) => {
            const isRunning = runningAppIds.has(item.id);

            return (
              <article
                key={item.id}
                onClick={() => handleAppClick(item)}
                onContextMenu={(e) => handleContextMenu(e, item.id)}
                className="group relative flex flex-col items-center justify-center px-4 py-3 rounded-[12px] overflow-hidden border-2 border-transparent transition-all duration-300 cursor-pointer"
                style={
                  {
                    '--card-border': item.borderColor,
                    background: item.gradient,
                  } as React.CSSProperties
                }
              >
                {/* Text label - two lines */}
                <div className="relative z-10 text-white text-center">
                  <div className="text-lg font-semibold tracking-wide">
                    {item.name.toUpperCase()}
                  </div>
                  <div className="text-xs opacity-70 tracking-wide">
                    {item.id.toUpperCase()}
                  </div>
                </div>

                {/* Running indicator */}
                {isRunning && (
                  <div className="absolute top-1 right-1 w-1.5 h-1.5 rounded-full bg-white shadow-sm" />
                )}
              </article>
            );
          })}
        </div>

        {/* Chroma overlay - grayscale effect */}
        <div
          className="absolute inset-0 pointer-events-none z-30 rounded-2xl"
          style={{
            backdropFilter: 'grayscale(1) brightness(0.75)',
            WebkitBackdropFilter: 'grayscale(1) brightness(0.75)',
            background: 'rgba(0,0,0,0.001)',
            maskImage:
              'radial-gradient(circle var(--r) at var(--x) var(--y), transparent 0%, transparent 15%, rgba(0,0,0,0.10) 30%, rgba(0,0,0,0.22) 45%, rgba(0,0,0,0.35) 60%, rgba(0,0,0,0.50) 75%, rgba(0,0,0,0.68) 88%, white 100%)',
            WebkitMaskImage:
              'radial-gradient(circle var(--r) at var(--x) var(--y), transparent 0%, transparent 15%, rgba(0,0,0,0.10) 30%, rgba(0,0,0,0.22) 45%, rgba(0,0,0,0.35) 60%, rgba(0,0,0,0.50) 75%, rgba(0,0,0,0.68) 88%, white 100%)'
          }}
        />

        {/* Fade overlay */}
        <div
          ref={fadeRef}
          className="absolute inset-0 pointer-events-none transition-opacity duration-[200ms] z-40 rounded-2xl"
          style={{
            backdropFilter: 'grayscale(1) brightness(0.75)',
            WebkitBackdropFilter: 'grayscale(1) brightness(0.75)',
            background: 'rgba(0,0,0,0.001)',
            maskImage:
              'radial-gradient(circle var(--r) at var(--x) var(--y), white 0%, white 15%, rgba(255,255,255,0.90) 30%, rgba(255,255,255,0.78) 45%, rgba(255,255,255,0.65) 60%, rgba(255,255,255,0.50) 75%, rgba(255,255,255,0.32) 88%, transparent 100%)',
            WebkitMaskImage:
              'radial-gradient(circle var(--r) at var(--x) var(--y), white 0%, white 15%, rgba(255,255,255,0.90) 30%, rgba(255,255,255,0.78) 45%, rgba(255,255,255,0.65) 60%, rgba(255,255,255,0.50) 75%, rgba(255,255,255,0.32) 88%, transparent 100%)',
            opacity: 1
          }}
        />
      </div>

      {/* Dock context menu */}
      <ContextMenu
        isVisible={contextMenu.isVisible}
        position={contextMenu.position}
        items={contextMenu.items}
        onClose={contextMenu.hide}
      />
    </>
  );
};

export default ChromaGridDock;
