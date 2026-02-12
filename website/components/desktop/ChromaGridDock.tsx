/**
 * v2e Portal - ChromaGrid Dock Component (Windows Taskbar Style)
 *
 * Windows-style taskbar with chromatic aberration effect
 * Running apps show green bar indicator
 */
'use client';

import React, { useRef, useEffect, useState, useCallback } from 'react';
import { gsap } from 'gsap';
import { Z_INDEX, WindowState } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';
import { getActiveApps } from '@/lib/desktop/app-registry';

interface ChromaAppItem {
  id: string;
  name: string;
  gradient: string;
  borderColor: string;
}

interface ChromaGridDockProps {
  className?: string;
  autoHide?: boolean;
  autoHideDelay?: number;
}

type SetterFn = (v: number | string) => void;

const iconGradients: Record<string, { gradient: string; borderColor: string }> = {
  cve: { gradient: 'linear-gradient(145deg, #4F46E5, #000)', borderColor: '#4F46E5' },
  cwe: { gradient: 'linear-gradient(210deg, #10B981, #000)', borderColor: '#10B981' },
  capec: { gradient: 'linear-gradient(165deg, #F59E0B, #000)', borderColor: '#F59E0B' },
  attack: { gradient: 'linear-gradient(195deg, #EF4444, #000)', borderColor: '#EF4444' },
  cvss: { gradient: 'linear-gradient(225deg, #8B5CF6, #000)', borderColor: '#8B5CF6' },
  glc: { gradient: 'linear-gradient(135deg, #06B6D4, #000)', borderColor: '#06B6D4' },
  mcards: { gradient: 'linear-gradient(145deg, #EC4899, #000)', borderColor: '#EC4899' },
  etl: { gradient: 'linear-gradient(210deg, #14B8A6, #000)', borderColor: '#14B8A6' },
  bookmarks: { gradient: 'linear-gradient(165deg, #F97316, #000)', borderColor: '#F97316' },
};

const ChromaGridDock: React.FC<ChromaGridDockProps> = ({
  className = '',
  autoHide = false,
  autoHideDelay = 3000,
}) => {
  const rootRef = useRef<HTMLDivElement>(null);
  const fadeRef = useRef<HTMLDivElement>(null);
  const setX = useRef<SetterFn | null>(null);
  const setY = useRef<SetterFn | null>(null);
  const pos = useRef({ x: 0, y: 0 });
  const hideTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [isHovering, setIsHovering] = useState(false);
  const { windows, dock, setDockVisibility, setDockAutoHide, addDockItem, openWindow, focusWindow, minimizeWindow, restoreWindow, closeWindow, selectDesktopIcon } = useDesktopStore();

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
      duration: 0.35,
      ease: 'power2.out',
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
      duration: 0.4,
      overwrite: true
    });
  };

  // Auto-hide logic
  const autoHideTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleMouseEnter = useCallback(() => {
    setIsHovering(true);
    if (hideTimeoutRef.current) {
      clearTimeout(hideTimeoutRef.current);
      hideTimeoutRef.current = null;
    }
    setDockVisibility(true);
  }, [setDockVisibility, autoHide, setIsHovering]);

  const handleMouseLeave = useCallback(() => {
    setIsHovering(false);
    if (autoHide) {
      hideTimeoutRef.current = setTimeout(() => {
        setDockVisibility(false);
      }, autoHideDelay);
    }
  }, [setIsHovering, autoHide]);

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
        className={`fixed bottom-0 left-0 right-0 h-14 bg-white/10 backdrop-blur-md border-t border-white/20 rounded-xl ${className}`}
        style={{
          '--x': '50%',
          '--y': '50%',
          zIndex: Z_INDEX.DOCK
        }}
      >
        {/* Taskbar-style horizontal bar */}
        <div className="relative w-full h-full flex items-center px-2 gap-1">
          {items.map((item) => {
            const isRunning = runningAppIds.has(item.id);

            return (
              <div
                key={item.id}
                onClick={() => handleAppClick(item)}
                className="relative flex-1 items-center gap-2 px-3 py-2 rounded-t-lg border-2 border-transparent transition-all duration-300 cursor-pointer group"
                style={{
                  background: item.gradient,
                  borderColor: item.borderColor,
                } as React.CSSProperties}
              >
                {/* App icon/color indicator */}
                <div
                  className="relative w-8 h-8 rounded-lg flex items-center justify-center"
                  style={{ backgroundColor: isRunning ? 'rgba(34, 197, 94, 0.3)' : 'transparent' }}
                >
                  <div className="w-3 h-3 rounded-full bg-green-500" style={{ opacity: isRunning ? 1 : 0 }} />
                </div>

                {/* App name */}
                <div className="flex-1 text-white text-xs font-medium tracking-wide">
                  {item.name}
                </div>

                {/* Running indicator bar */}
                {isRunning && (
                  <div className="absolute bottom-0 left-0 right-0 h-1 bg-green-500" />
                )}
              </div>
            );
          })}
        </div>

        {/* Chroma overlay - grayscale effect */}
        <div
          className="absolute inset-0 pointer-events-none z-30 rounded-xl"
          style={{
            backdropFilter: 'grayscale(1) brightness(0.75)',
            WebkitBackdropFilter: 'grayscale(1) brightness(0.75)',
            background: 'rgba(0, 0, 0, 0.001)',
            maskImage:
              'radial-gradient(circle at var(--r) at var(--x) var(--y), transparent 0%, transparent 15%, rgba(0, 0, 0.10) 30%, rgba(0, 0, 0.22) 45%, rgba(0, 0, 0.35) 60%, rgba(0, 0, 0.50) 75%, rgba(0, 0, 0.68) 88%, white 100%)',
            WebkitMaskImage:
              'radial-gradient(circle at var(--r) at var(--x) var(--y), transparent 0%, transparent 15%, rgba(0, 0, 0.10) 30%, rgba(0, 0, 0.22) 45%, rgba(0, 0, 0.35) 60%, rgba(0, 0, 0.50) 75%, rgba(0, 0, 0.68) 88%, white 100%)'
          }}
        />

        {/* Fade overlay */}
        <div
          ref={fadeRef}
          className="absolute inset-0 pointer-events-none transition-opacity duration-[200ms] z-40 rounded-xl"
          style={{
            backdropFilter: 'grayscale(1) brightness(0.75)',
            WebkitBackdropFilter: 'grayscale(1) brightness(0.75)',
            background: 'rgba(0, 0, 0, 0.001)',
            maskImage:
              'radial-gradient(circle at var(--r) at var(--x) var(--y), transparent 0%, transparent 15%, rgba(0, 0, 0.10) 30%, rgba(0, 0, 0.22) 45%, rgba(0, 0, 0.35) 60%, rgba(0, 0, 0.50) 75%, rgba(0, 0, 0.68) 88%, white 100%)',
            opacity: 1
          }}
        />
      </div>
    </>
  );
};

export default ChromaGridDock;
