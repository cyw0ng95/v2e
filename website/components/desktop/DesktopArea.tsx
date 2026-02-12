/**
 * v2e Portal - Desktop Area Component
 *
 * Main desktop area with icons, wallpaper, and widgets
 * Renders without backend dependency
 */

'use client';

import React from 'react';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX, WindowState } from '@/types/desktop';
import type { DesktopIcon as DesktopIconType } from '@/types/desktop';
import { ContextMenu, ContextMenuPresets, useContextMenu } from '@/components/desktop/ContextMenu';
import { getAppById } from '@/lib/desktop/app-registry';
import { ClockWidget } from './ClockWidget';
import Threads from './Threads';
import type { WidgetConfig } from '@/types/desktop';

/**
 * Desktop icon component
 */
function DesktopIcon({ icon }: { icon: DesktopIconType }) {
  const { selectDesktopIcon, openWindow, isOnline } = useDesktopStore();
  const { updateDesktopIconPosition } = useDesktopStore();
  const contextMenu = useContextMenu();

  // Check if this app requires online access
  const app = getAppById(icon.appId);
  const requiresNetwork = app?.requiresOnline && !isOnline;

  const handleClick = () => {
    selectDesktopIcon(icon.id);
  };

  const handleDoubleClick = () => {
    // Open app window using registry metadata
    if (app) {
      // Check if app requires network and system is offline
      if (app.requiresOnline && !isOnline) {
        alert(`Cannot open ${app.name}: This app requires an internet connection.`);
        return;
      }

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

  const handleDragStart = (e: React.DragEvent) => {
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', icon.id);
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    contextMenu.show(e.clientX, e.clientY, ContextMenuPresets.desktopIcon(icon.id, icon.appId));
  };

  return (
    <div
      onClick={handleClick}
      onDoubleClick={handleDoubleClick}
      onContextMenu={handleContextMenu}
      draggable
      onDragStart={handleDragStart}
      className={`
        absolute flex flex-col items-center gap-2 p-3 rounded-lg cursor-pointer
        transition-all duration-200
        ${icon.isSelected
          ? 'bg-blue-500/20 ring-2 ring-blue-500'
          : 'hover:bg-white/10'}
        ${icon.isSelected ? 'scale-105' : ''}
        hover:scale-105
        ${requiresNetwork ? 'opacity-50 cursor-not-allowed' : ''}
      `}
      style={{
        left: `${icon.position.x}px`,
        top: `${icon.position.y}px`,
        zIndex: Z_INDEX.DESKTOP_ICONS,
      }}
      role="button"
      tabIndex={requiresNetwork ? -1 : 0}
      aria-label={`Desktop icon for ${icon.appId}${requiresNetwork ? ' (requires internet)' : ''}`}
      aria-selected={icon.isSelected}
    >
      {/* Icon container with relative positioning for badge */}
      <div className="relative w-12 h-12">
        {/* Icon placeholder - will be replaced with Lucide icons */}
        <div className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-700 to-gray-900 flex items-center justify-center">
          <span className="text-white text-lg font-bold">
            {icon.appId[0]}
          </span>
        </div>

        {/* Offline badge overlay */}
        {requiresNetwork && (
          <div
            className="absolute -top-1 -right-1 bg-black/60 rounded-full p-0.5"
            title="Requires internet connection"
          >
            <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 6.364l-3.536-3.536M15 3.536v15m0 0l-7.07 7.071 7.071-7.071V15a2 2 0 002-2V7a2 2 0 012 2v10a2 2 0 01-2 2l-7.07 7.071a2 2 0 002 2z" />
            </svg>
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * Desktop area component
 * Renders wallpaper and desktop icons
 */
export function DesktopArea() {
  const { desktopIcons, widgets, theme, isOnline } = useDesktopStore();
  const contextMenu = useContextMenu();

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    contextMenu.show(e.clientX, e.clientY, ContextMenuPresets.desktop());
  };

  return (
    <>
      {/* Threads animation background */}
      <div
        className="absolute inset-0 overflow-hidden"
        style={{
          zIndex: Z_INDEX.DESKTOP_WALLPAPER,
          pointerEvents: 'none',
        }}
        aria-hidden="true"
      >
        <Threads
          color={[0.7450980392156863, 0.741764705882353, 0.7607843137254902]}
          amplitude={1}
          distance={0}
          enableMouseInteraction
          className="w-full h-full"
        />
      </div>

      <main
        className="absolute inset-0 overflow-hidden"
        style={{
          background: theme.wallpaper,
          zIndex: Z_INDEX.DESKTOP_WALLPAPER,
        }}
        onContextMenu={handleContextMenu}
        role="main"
        aria-label="Desktop area"
      >
        {/* Desktop icons grid */}
        <div className="relative w-full h-full">
          {desktopIcons.map(icon => (
            <DesktopIcon key={icon.id} icon={icon} />
          ))}
        </div>

        {/* Desktop widgets */}
        {widgets.filter(w => w.isVisible).map(widget => (
          widget.type === 'clock' ? (
            <ClockWidget key={widget.id} widget={widget} />
          ) : null
        ))}
      </main>

      {/* Global context menu */}
      <ContextMenu
        isVisible={contextMenu.isVisible}
        position={contextMenu.position}
        items={contextMenu.items}
        onClose={contextMenu.hide}
      />
    </>
  );
}
