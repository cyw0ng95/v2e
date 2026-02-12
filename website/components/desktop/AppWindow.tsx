/**
 * v2e Portal - App Window Component
 *
 * Complete window management with titlebar, controls, content area
 * Phase 2: Window System
 */

'use client';

import React, { useState, useRef, useEffect } from 'react';
import { X, Minus, Square, Copy } from 'lucide-react';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX, type WindowConfig } from '@/types/desktop';
import { motion, AnimatePresence } from 'framer-motion';

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
          <Copy className="w-3.5 h-3.5 text-gray-600" />
        ) : (
          <Square className="w-3.5 h-3.5 text-gray-600" />
        )}
      </button>
    </div>
  );
}

/**
 * Window titlebar component
 */
function WindowTitlebar({ window }: { window: WindowConfig }) {
  const { focusWindow } = useDesktopStore();
  const titlebarRef = useRef<HTMLDivElement>(null);

  const handleMouseDown = (e: React.MouseEvent) => {
    // Focus window on click
    focusWindow(window.id);
    // TODO: Will initiate drag in Phase 2 (Agent 2)
    console.log('Titlebar mouse down - prepare for drag');
  };

  return (
    <div
      ref={titlebarRef}
      onMouseDown={handleMouseDown}
      className="flex items-center justify-between px-3 py-2 bg-gradient-to-r from-gray-50 to-gray-100 border-b border-gray-200 select-none"
      style={{ cursor: 'move' }}
    >
      {/* App icon and title */}
      <div className="flex items-center gap-2">
        <div className="w-4 h-4 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
          <span className="text-white text-xs font-bold">{window.appId[0]}</span>
        </div>
        <span className="text-sm font-medium text-gray-700">{window.title}</span>
      </div>

      {/* Window controls */}
      <WindowControls window={window} />
    </div>
  );
}

/**
 * Window resize handle component
 */
function WindowResizeHandle({
  position,
  onResizeStart,
}: {
  position: 'top' | 'bottom' | 'left' | 'right' | 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
  onResizeStart: (e: React.MouseEvent) => void;
}) {
  return (
    <div
      onMouseDown={onResizeStart}
      className={`
        absolute
        ${position === 'left' || position === 'right' ? 'w-1 h-full cursor-ew-resize' : ''}
        ${position === 'top' || position === 'bottom' ? 'h-1 w-full cursor-ns-resize' : ''}
        ${position.includes('left') ? 'left-0' : position.includes('right') ? 'right-0' : ''}
        ${position.includes('top') ? 'top-0' : position.includes('bottom') ? 'bottom-0' : ''}
        hover:bg-blue-500/30
      `}
      style={{ zIndex: Z_INDEX.CONTEXT_MENU + 1 }}
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

      {/* Window content - iframe placeholder */}
      <div className="absolute inset-0 top-10 bg-gray-50">
        {window.isMinimized ? (
          <div className="h-full flex items-center justify-center text-gray-400">
            <p>Window minimized</p>
          </div>
        ) : (
          <iframe
            src={`/website${window.appId}`} // Will be updated in Phase 4
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
