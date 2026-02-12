/**
 * v2e Portal - App Window Component
 *
 * Complete window management with titlebar, controls, content area
 * Phase 2: Window System - Enhanced with drag, resize, persistence
 * Iteration 2: Adding full window interaction support
 */

'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { X, Minus, Square, Copy, Maximize2 } from 'lucide-react';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX, type WindowConfig } from '@/types/desktop';
import { motion, AnimatePresence } from 'framer-motion';
import { MinimizeAnimation } from './MinimizeAnimation';

/**
 * Window controls component
 * Close, minimize, maximize buttons
 */
function WindowControls({ window }: { window: WindowConfig }) {
  const { closeWindow, minimizeWindow, maximizeWindow, focusWindow } = useDesktopStore();

  const handleClose = () => closeWindow(window.id);
  const handleMinimize = () => minimizeWindow(window.id);
  const handleMaximize = () => maximizeWindow(window.id);

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
}

/**
 * Window titlebar component
 * Handles window dragging with 60fps performance
 */
function WindowTitlebar({ window }: { window: WindowConfig }) {
  const { focusWindow, updateWindowPosition } = useDesktopStore();
  const titlebarRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const dragStartPosRef = useRef<{ x: number; y: number } | null>(null);
  const windowRef = useRef<HTMLDivElement>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    // Focus window on click
    focusWindow(window.id);

    // Initialize drag
    setIsDragging(true);
    dragStartPosRef.current = {
      x: e.clientX,
      y: e.clientY,
    };
  }, [focusWindow, window.id]);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging || !dragStartPosRef.current) return;

    e.preventDefault();

    const deltaX = e.clientX - dragStartPosRef.current.x;
    const deltaY = e.clientY - dragStartPosRef.current.y;

    // Calculate new position with boundary constraints
    const newPosition = {
      x: window.position.x + deltaX,
      y: window.position.y + deltaY,
    };

    // Constrain to viewport bounds (with menu and dock)
    const maxX = window.innerWidth - window.size.width - 20; // 20px margin
    const maxY = window.innerHeight - window.size.height - 28 - 80 - 20; // Menu + Dock + margin
    const minY = 28; // Below menu bar

    newPosition.x = Math.max(0, Math.min(newPosition.x, maxX));
    newPosition.y = Math.max(minY, Math.min(newPosition.y, maxY));

    // Update position (will trigger re-render)
    updateWindowPosition(window.id, newPosition);

    // Update drag start position for next frame
    dragStartPosRef.current = {
      x: e.clientX,
      y: e.clientY,
    };
  }, [isDragging, window.position.x, window.position.y, window.size.width, window.size.height, window.id, updateWindowPosition]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
    dragStartPosRef.current = null;
  }, []);

  // Add global mouse move/up listeners
  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, handleMouseMove, handleMouseUp]);

  return (
    <div
      ref={titlebarRef}
      onMouseDown={handleMouseDown}
      className={`
        flex items-center justify-between px-3 py-2
        bg-gradient-to-r from-gray-50 to-gray-100
        border-b border-gray-200 select-none
        ${isDragging ? 'cursor-grabbing' : 'cursor-grab'}
      `}
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
  window: WindowConfig;
}) {
  const { updateWindowSize } = useDesktopStore();
  const [isResizing, setIsResizing] = useState(false);
  const resizeStartPosRef = useRef<{ x: number; y: number; width: number; height: number } | null>(null);

  const handleMouseDown = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsResizing(true);
    resizeStartPosRef.current = {
      x: e.clientX,
      y: e.clientY,
      width: window.size.width,
      height: window.size.height,
    };
    onResizeStart(position, e);
  };

  // Handle resize with requestAnimationFrame for 60fps
  useEffect(() => {
    if (!isResizing) return;

    let rafId: number | null = null;

    const handleMouseMove = (e: MouseEvent) => {
      if (!isResizing || !resizeStartPosRef.current) return;

      if (rafId) return;
      rafId = requestAnimationFrame(() => {
        const deltaX = e.clientX - resizeStartPosRef.current.x;
        const deltaY = e.clientY - resizeStartPosRef.current.y;

        let newWidth = resizeStartPosRef.current.width;
        let newHeight = resizeStartPosRef.current.height;

        // Calculate new size based on resize direction
        if (position.includes('right')) {
          newWidth = Math.max(window.minWidth, resizeStartPosRef.current.width + deltaX);
        } else if (position.includes('left')) {
          newWidth = Math.max(window.minWidth, resizeStartPosRef.current.width - deltaX);
        }

        if (position.includes('bottom')) {
          newHeight = Math.max(window.minHeight, resizeStartPosRef.current.height + deltaY);
        } else if (position.includes('top')) {
          newHeight = Math.max(window.minHeight, resizeStartPosRef.current.height - deltaY);
        }

        // Apply max constraints if set
        const maxWidth = window.maxWidth ?? window.innerWidth - 40;
        newWidth = Math.min(newWidth, maxWidth);

        // Update window size
        updateWindowSize(window.id, { width: newWidth, height: newHeight });

        // Update start position for next frame
        resizeStartPosRef.current = {
          x: e.clientX,
          y: e.clientY,
          width: newWidth,
          height: newHeight,
        };
      });
    };

    const handleMouseUp = () => {
      if (rafId) {
        cancelAnimationFrame(rafId);
        rafId = null;
      }
      setIsResizing(false);
      resizeStartPosRef.current = null;
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      if (rafId) cancelAnimationFrame(rafId);
    };
  }, [isResizing, position, window.minWidth, window.minHeight, window.maxWidth, window.id, updateWindowSize]);

  // Cursor styles based on resize direction
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
        ${isResizing ? 'z-50' : 'z-10'}
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
  const { updateWindowSize, updateWindowPosition } = useDesktopStore();
  const [isResizing, setIsResizing] = useState(false);
  const resizeDirectionRef = useRef<string | null>(null);

  const handleResizeStart = (direction: string) => (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsResizing(true);
    resizeDirectionRef.current = direction;
  };

  // Calculate window size based on state
  const windowStyle: React.CSSProperties = {
    left: `${window.position.x}px`,
    top: `${window.position.y}px`,
    width: `${window.size.width}px`,
    height: `${window.size.height}px`,
    minWidth: `${window.minWidth}px`,
    minHeight: `${window.minHeight}px`,
  };

  if (window.isMaximized) {
    // Maximized - fill available space (minus menu and dock)
    windowStyle.left = '0';
    windowStyle.top = '28px'; // Below menu bar
    windowStyle.width = '100%';
    windowStyle.height = 'calc(100vh - 28px - 80px)'; // Minus menu and dock
  }

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
      className={`
        absolute bg-white rounded-lg shadow-2xl overflow-hidden
        ${window.isFocused ? 'ring-2 ring-blue-500' : ''}
        ${window.isMinimized ? 'opacity-0' : ''}
      `}
      style={{
        ...windowStyle,
        zIndex: window.zIndex,
      }}
      role="dialog"
      aria-labelledby={`window-title-${window.id}`}
      aria-modal={window.isFocused ? 'true' : 'false'}
    >
      {/* Titlebar */}
      <WindowTitlebar window={window} />

      {/* Resize handles - 8 directions */}
      {!window.isMaximized && (
        <>
          <WindowResizeHandle position="top-left" onResizeStart={() => handleResizeStart('top-left')} />
          <WindowResizeHandle position="top" onResizeStart={() => handleResizeStart('top')} />
          <WindowResizeHandle position="top-right" onResizeStart={() => handleResizeStart('top-right')} />
          <WindowResizeHandle position="left" onResizeStart={() => handleResizeStart('left')} />
          <WindowResizeHandle position="right" onResizeStart={() => handleResizeStart('right')} />
          <WindowResizeHandle position="bottom-left" onResizeStart={() => handleResizeStart('bottom-left')} />
          <WindowResizeHandle position="bottom" onResizeStart={() => handleResizeStart('bottom')} />
          <WindowResizeHandle position="bottom-right" onResizeStart={() => handleResizeStart('bottom-right')} />
        </>
      )}

      {/* Window content - iframe with APP_REGISTRY integration */}
      <div className="absolute inset-0 top-10 bg-gray-50">
        {window.isMinimized ? (
          <div className="h-full flex items-center justify-center text-gray-400">
            <p>Window minimized</p>
          </div>
        ) : (
          <iframe
            src={window.appId} // Use APP_REGISTRY path directly
            title={window.title}
            className="w-full h-full border-0"
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
            loading="lazy"
          />
        )}
      </div>
    </motion.div>
  );
}
