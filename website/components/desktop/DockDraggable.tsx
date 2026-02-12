/**
 * v2e Portal - Draggable Dock Component
 *
 * Dock with drag-to-reorder using native HTML5 drag-drop
 * Phase 3: Dock Drag-to-Reorder
 * Backend Independence: Works completely offline
 */

'use client';

import React, { useState, useCallback, useRef } from 'react';
import { Z_INDEX } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';
import { ContextMenu, ContextMenuPresets, useContextMenu } from '@/components/desktop/ContextMenu';
import { getActiveApps } from '@/lib/desktop/app-registry';
import type { AppRegistryEntry } from '@/lib/desktop/app-registry';
import { WindowThumbnail } from './MinimizeThumbnails';

/**
 * Dock item with drag support
 */
interface DraggableDockItemProps {
  app: AppRegistryEntry;
  isRunning: boolean;
  isIndicator: boolean;
  index: number;
  isDragging: boolean;
  onDragStart: (index: number) => void;
  onDragOver: (index: number) => void;
  onDragEnd: () => void;
}

function DraggableDockItem({ app, isRunning, isIndicator, index, isDragging, onDragStart, onDragOver, onDragEnd }: DraggableDockItemProps) {
  const { openWindow, windows, minimizeWindow } = useDesktopStore();
  const contextMenu = useContextMenu();
  const window = Object.values(windows).find(w => w.appId === app.id);

  const handleClick = useCallback(() => {
    if (window) {
      if (window.isFocused) {
        minimizeWindow(window.id);
      } else {
        const { focusWindow } = useDesktopStore.getState();
        focusWindow(window.id);
      }
    } else {
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
        innerWidth: window.innerWidth,
        innerHeight: window.innerHeight,
      });
    }
  }, [app, window, openWindow, minimizeWindow]);

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const preset = isRunning
      ? ContextMenuPresets.dockItemRunning(app.id)
      : ContextMenuPresets.dockItemNotRunning(app.id);
    contextMenu.show(e.clientX, e.clientY, preset);
  }, [isRunning, app.id, contextMenu]);

  return (
    <div
      draggable
      onDragStart={(e) => {
        e.dataTransfer.effectAllowed = 'move';
        onDragStart(index);
      }}
      onDragOver={(e) => {
        e.preventDefault();
        onDragOver(index);
      }}
      onDragEnd={onDragEnd}
      className={`
        relative flex flex-col items-center cursor-move
        ${isDragging ? 'opacity-50 scale-105' : 'hover:scale-110'}
        transition-all duration-200
      `}
    >
      <button
        onClick={handleClick}
        onContextMenu={handleContextMenu}
        className="relative flex flex-col items-center p-2 rounded-lg"
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
    </div>
  );
}

/**
 * Dock component with drag-to-reorder
 */
export function DockDraggable() {
  const { dock, windows, setDockVisibility, updateDockItems } = useDesktopStore();
  const [isHovering, setIsHovering] = useState(false);
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  const hideTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const dockRef = useRef<HTMLDivElement>(null);
  const contextMenu = useContextMenu();

  const sizeClasses = {
    small: 'h-12',
    medium: 'h-20',
    large: 'h-24',
  };

  // Get apps from registry
  const registryApps = getActiveApps();

  // Build dock items with running state
  const [dockItems, setDockItemsLocal] = useState(() => {
    const items = registryApps.map(app => {
      const isRunning = Object.values(windows).some(w => w.appId === app.id);
      return {
        app,
        isRunning,
        isIndicator: isRunning,
      };
    });
    return items;
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

  // Show dock when mouse moves to bottom edge
  React.useEffect(() => {
    if (!dock.autoHide) return;

    const handleMouseMove = (e: MouseEvent) => {
      const threshold = 50;
      if (window.innerHeight - e.clientY < threshold) {
        setDockVisibility(true);
      }
    };

    document.addEventListener('mousemove', handleMouseMove);
    return () => document.removeEventListener('mousemove', handleMouseMove);
  }, [dock.autoHide, setDockVisibility]);

  // Drag handlers
  const handleDragStart = useCallback((index: number) => {
    setDraggedIndex(index);
  }, []);

  const handleDragOver = useCallback((index: number) => {
    setDragOverIndex(index);
  }, []);

  const handleDragEnd = useCallback(() => {
    if (draggedIndex !== null && dragOverIndex !== null && draggedIndex !== dragOverIndex) {
      // Reorder items
      const newItems = [...dockItems];
      const [removed] = newItems.splice(draggedIndex, 1);
      newItems.splice(dragOverIndex, 0, removed);
      setDockItemsLocal(newItems);
      updateDockItems(newItems);
    }
    setDraggedIndex(null);
    setDragOverIndex(null);
  }, [draggedIndex, dragOverIndex, dockItems, updateDockItems]);

  // Should show dock
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
            <DraggableDockItem
              key={`${item.app.id}-${index}`}
              app={item.app}
              isRunning={item.isRunning}
              isIndicator={item.isIndicator}
              index={index}
              isDragging={draggedIndex === index}
              isMinimized={minimizedWindows.some(w => w.appId === item.app.id)}
              onDragStart={handleDragStart}
              onDragOver={handleDragOver}
              onDragEnd={handleDragEnd}
            />
          ))}
        </div>

        {/* Minimized window thumbnails */}
        {minimizedWindows.map(win => (
          <WindowThumbnail
            key={win.id}
            window={win}
            onRestore={() => restoreWindow(win.id)}
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
