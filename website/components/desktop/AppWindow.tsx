/**
 * v2e Portal - App Window Component
 *
 * Complete window management with titlebar, controls, content area
 * Phase 2: Window System - Enhanced with drag, resize, persistence
 * Iteration 2: Adding full window interaction support
 * Optimized: 60fps drag performance with requestAnimationFrame
 */

'use client';

import React, { useState, useRef, useEffect, useCallback, useMemo, memo } from 'react';
import { X, Minus, Square, Copy, Maximize2 } from 'lucide-react';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX, type WindowConfig } from '@/types/desktop';
import { motion, AnimatePresence } from 'framer-motion';
import { MinimizeAnimation } from './MinimizeAnimation';
import { AppComponent } from './AppComponents';

/**
 * Window controls component
 * Close, minimize, maximize buttons
 */
const WindowControls = memo(function WindowControls({ window }: { window: WindowConfig }) {
  const { closeWindow, minimizeWindow, maximizeWindow, focusWindow } = useDesktopStore();
 
  const handleClose = useCallback(() => closeWindow(window.id), [closeWindow, window.id]);
  const handleMinimize = useCallback(() => minimizeWindow(window.id), [minimizeWindow, window.id]);
  const handleMaximize = useCallback(() => maximizeWindow(window.id), [maximizeWindow, window.id]);
 
  return (
    <div className="flex items-center gap-1">
      {/* Close button */}
      <button
        onClick={handleClose}
        className="p-1.5 hover:bg-red-500/20 rounded transition-colors"
        aria-label="Close window"
        title="Close (⌘W)"
      >
        <X className="w-3.5 h-3.5 text-gray-600" />
      </button>
 
      {/* Minimize button */}
      <button
        onClick={handleMinimize}
        className="p-1.5 hover:bg-gray-500/20 rounded transition-colors"
        aria-label="Minimize window"
        title="Minimize (⌘M)"
      >
        <Minus className="w-3.5 h-3.5 text-gray-600" />
      </button>
 
      {/* Maximize/Restore button */}
      <button
        onClick={handleMaximize}
        className="p-1.5 hover:bg-gray-500/20 rounded transition-colors"
        aria-label={window.isMaximized ? "Restore window" : "Maximize window"}
        title={window.isMaximized ? "Restore (⌘⌥)" : "Maximize (⌘⌃)"}
      >
        {window.isMaximized ? (
          <Maximize2 className="w-3.5 h-3.5 text-gray-600" />
        ) : (
          <Square className="w-3.5 h-3.5 text-gray-600" />
        )}
      </button>
    </div>
  );
});

/**
 * Window titlebar component
 * Handles window dragging with 60fps performance using requestAnimationFrame
 */
function WindowTitlebar({ window }: { window: WindowConfig }) {
  const { focusWindow, updateWindowPosition } = useDesktopStore();
  const titlebarRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);

  // Use refs for drag state to avoid re-renders during drag
  const dragStateRef = useRef({
    isDragging: false,
    startX: 0,
    startY: 0,
    windowStartX: 0,
    windowStartY: 0,
  });

  // Stable refs that don't change
  const windowIdRef = useRef(window.id);
  const windowSizeRef = useRef({ width: window.size.width, height: window.size.height });

  // Update refs when window changes
  useEffect(() => {
    windowIdRef.current = window.id;
    windowSizeRef.current = { width: window.size.width, height: window.size.height };
  }, [window.id, window.size.width, window.size.height]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    // Only left button
    if (e.button !== 0) return;

    // Focus window on click
    focusWindow(window.id);

    // Initialize drag state
    dragStateRef.current = {
      isDragging: true,
      startX: e.clientX,
      startY: e.clientY,
      windowStartX: window.position.x,
      windowStartY: window.position.y,
    };
    setIsDragging(true);
  }, [focusWindow, window.id, window.position.x, window.position.y]);

  useEffect(() => {
    const { isDragging, startX, startY, windowStartX, windowStartY } = dragStateRef.current;

    if (!isDragging) return;

    let rafId: number | null = null;

    const handleMouseMove = (e: MouseEvent) => {
      if (!dragStateRef.current.isDragging) return;

      if (rafId) return; // Skip if frame is pending

      rafId = requestAnimationFrame(() => {
        const state = dragStateRef.current;
        if (!state.isDragging) return;

        const deltaX = e.clientX - state.startX;
        const deltaY = e.clientY - state.startY;

        // Calculate new position with boundary constraints
        let newX = state.windowStartX + deltaX;
        let newY = state.windowStartY + deltaY;

        // Boundary constraints
        const viewportWidth = globalThis.innerWidth || 1024;
        const viewportHeight = globalThis.innerHeight || 768;
        const windowWidth = windowSizeRef.current.width;
        const windowHeight = windowSizeRef.current.height;

        const maxX = viewportWidth - windowWidth - 20;
        const maxY = viewportHeight - windowHeight - 28 - 80 - 20;
        const minY = 28;

        newX = Math.max(0, Math.min(newX, maxX));
        newY = Math.max(minY, Math.min(newY, maxY));

        // Batch update position
        updateWindowPosition(windowIdRef.current, { x: newX, y: newY });
        rafId = null;
      });
    };

    const handleMouseUp = () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
        rafId = null;
      }
      dragStateRef.current.isDragging = false;
      setIsDragging(false);
    };

    document.addEventListener('mousemove', handleMouseMove, { passive: true });
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      if (rafId) cancelAnimationFrame(rafId);
    };
  }, [isDragging, updateWindowPosition]);

  return (
    <div
      ref={titlebarRef}
      onMouseDown={handleMouseDown}
      className={`
        flex items-center justify-between px-3 py-2
        bg-gradient-to-r from-gray-50 to-gray-100
        border-b border-gray-200 select-none
      `}
      style={{ cursor: isDragging ? 'grabbing' : 'grab' }}
    >
      {/* App icon and title */}
      <div className="flex items-center gap-2 select-none">
        <div className="w-4 h-4 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
          <span className="text-white text-xs font-bold">{window.appId[0]}</span>
        </div>
        <span className="text-sm font-medium text-gray-700 select-none">{window.title}</span>
      </div>

      {/* Window controls */}
      <WindowControls window={window} />
    </div>
  );
}

/**
 * Window resize handle component
 * Handles live window resizing with 8 directions
 */
function WindowResizeHandle({
  position,
  onResizeStart,
  window,
}: {
  position: 'top' | 'bottom' | 'left' | 'right' | 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
  onResizeStart: (direction: string, e: React.MouseEvent) => void;
  window?: WindowConfig;
}) {
  const { updateWindowSize } = useDesktopStore();
  const [isResizing, setIsResizing] = useState(false);
  const resizeStateRef = useRef({
    isResizing: false,
    startX: 0,
    startY: 0,
    startWidth: 0,
    startHeight: 0,
  });

  if (!window) {
    return null;
  }

  const handleMouseDown = (e: React.MouseEvent) => {
    e.stopPropagation();
    resizeStateRef.current = {
      isResizing: true,
      startX: e.clientX,
      startY: e.clientY,
      startWidth: window.size.width,
      startHeight: window.size.height,
    };
    setIsResizing(true);
    onResizeStart(position, e);
  };

  useEffect(() => {
    if (!resizeStateRef.current.isResizing) return;

    let rafId: number | null = null;

    const handleMouseMove = (e: MouseEvent) => {
      if (!resizeStateRef.current.isResizing) return;

      if (rafId) return;
      rafId = requestAnimationFrame(() => {
        const state = resizeStateRef.current;
        if (!state.isResizing) return;

        const deltaX = e.clientX - state.startX;
        const deltaY = e.clientY - state.startY;

        let newWidth = state.startWidth;
        let newHeight = state.startHeight;

        // Calculate new size based on resize direction
        if (position.includes('right')) {
          newWidth = Math.max(window.minWidth, state.startWidth + deltaX);
        } else if (position.includes('left')) {
          newWidth = Math.max(window.minWidth, state.startWidth - deltaX);
        }

        if (position.includes('bottom')) {
          newHeight = Math.max(window.minHeight, state.startHeight + deltaY);
        } else if (position.includes('top')) {
          newHeight = Math.max(window.minHeight, state.startHeight - deltaY);
        }

        // Apply max constraints if set
        const maxWidth = window.maxWidth ?? globalThis.innerWidth - 40;
        newWidth = Math.min(newWidth, maxWidth);

        updateWindowSize(window.id, { width: newWidth, height: newHeight });
        rafId = null;
      });
    };

    const handleMouseUp = () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
        rafId = null;
      }
      resizeStateRef.current.isResizing = false;
      setIsResizing(false);
    };

    document.addEventListener('mousemove', handleMouseMove, { passive: true });
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      if (rafId) cancelAnimationFrame(rafId);
    };
  }, [isResizing, position, window.id, window.minWidth, window.minHeight, window.maxWidth, updateWindowSize]);

  const getCursorClass = () => {
    if (position === 'top' || position === 'bottom') return 'cursor-ns-resize';
    if (position === 'left' || position === 'right') return 'cursor-ew-resize';
    return 'cursor-nwse-resize';
  };

  return (
    <div
      onMouseDown={handleMouseDown}
      className={`
        absolute bg-transparent hover:bg-blue-500/30
        ${getCursorClass()}
        transition-colors duration-150
      `}
      style={{ zIndex: Z_INDEX.CONTEXT_MENU + 1 }}
      aria-label={`Resize window ${position}`}
    />
  );
}

/**
 * Main window component
 * Complete window with titlebar, resize handles, content area
 */
export function AppWindow({ window }: { window: WindowConfig }) {
  const handleResizeStart = useCallback((direction: string) => {
    return () => {};
  }, []);

  // Memoize window style to avoid recalculating on every render
  const windowStyle = useMemo<React.CSSProperties>(() => {
    const style: React.CSSProperties = {
      left: `${window.position.x}px`,
      top: `${window.position.y}px`,
      width: `${window.size.width}px`,
      height: `${window.size.height}px`,
      minWidth: `${window.minWidth}px`,
      minHeight: `${window.minHeight}px`,
      zIndex: window.zIndex,
    };

    if (window.isMaximized) {
      style.left = '0';
      style.top = '28px';
      style.width = '100%';
      style.height = 'calc(100vh - 28px - 80px)';
    }

    return style;
  }, [window.position.x, window.position.y, window.size.width, window.size.height,
      window.minWidth, window.minHeight, window.maxWidth, window.zIndex, window.isMaximized]);

  // Memoize class names to avoid recalculating
  const windowClassName = useMemo(() => {
    return [
      'absolute bg-white rounded-lg shadow-2xl overflow-hidden pointer-events-auto',
      window.isFocused ? 'ring-2 ring-blue-500' : '',
      window.isMinimized ? 'opacity-0' : '',
    ].filter(Boolean).join(' ');
  }, [window.isFocused, window.isMinimized]);

  return (
    <motion.div
      layout
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{
        opacity: 1,
        scale: 1,
      }}
      exit={{
        opacity: 0,
        scale: 0.95,
      }}
      transition={{
        duration: window.state === 'closing' ? 0.15 : 0.2,
        ease: 'easeInOut',
      }}
      className={windowClassName}
      style={windowStyle}
      role="dialog"
      aria-labelledby={`window-title-${window.id}`}
      aria-modal={window.isFocused ? 'true' : 'false'}
    >
      {/* Titlebar */}
      <WindowTitlebar window={window} />

      {/* Resize handles - 8 directions */}
      {!window.isMaximized && (
        <>
          <WindowResizeHandle position="top-left" onResizeStart={() => handleResizeStart('top-left')} window={window} />
          <WindowResizeHandle position="top" onResizeStart={() => handleResizeStart('top')} window={window} />
          <WindowResizeHandle position="top-right" onResizeStart={() => handleResizeStart('top-right')} window={window} />
          <WindowResizeHandle position="left" onResizeStart={() => handleResizeStart('left')} window={window} />
          <WindowResizeHandle position="right" onResizeStart={() => handleResizeStart('right')} window={window} />
          <WindowResizeHandle position="bottom-left" onResizeStart={() => handleResizeStart('bottom-left')} window={window} />
          <WindowResizeHandle position="bottom" onResizeStart={() => handleResizeStart('bottom')} window={window} />
          <WindowResizeHandle position="bottom-right" onResizeStart={() => handleResizeStart('bottom-right')} window={window} />
        </>
      )}

      {/* Window content - direct component rendering for SPA architecture */}
      <div className="absolute inset-0 top-10 bg-gray-50">
        {window.isMinimized ? (
          <div className="h-full flex items-center justify-center text-gray-400">
            <p>Window minimized</p>
          </div>
        ) : (
          <AppComponent appId={window.appId} title={window.title} windowId={window.id} />
        )}
      </div>
    </motion.div>
  );
}
