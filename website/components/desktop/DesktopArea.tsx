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
  const { selectDesktopIcon, openWindow } = useDesktopStore();
  const { updateDesktopIconPosition } = useDesktopStore();
  const contextMenu = useContextMenu();

  const handleClick = () => {
    selectDesktopIcon(icon.id);
  };

  const handleDoubleClick = () => {
    // Open app window using registry metadata
    const app = getAppById(icon.appId);
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
      `}
      style={{
        left: `${icon.position.x}px`,
        top: `${icon.position.y}px`,
        zIndex: Z_INDEX.DESKTOP_ICONS,
      }}
      role="button"
      tabIndex={0}
      aria-label={`Desktop icon for ${icon.appId}`}
      aria-selected={icon.isSelected}
    >
      {/* Icon placeholder - will be replaced with Lucide icons */}
      <div className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-700 to-gray-900 flex items-center justify-center">
        <span className="text-white text-lg font-bold">
          {icon.appId[0]}
        </span>
      </div>
    </div>
  );
}

/**
 * Desktop area component
 * Renders wallpaper and desktop icons
 */
export function DesktopArea() {
  const { desktopIcons, widgets, theme } = useDesktopStore();
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
          color={[0.6941176470588235, 0.6666666666666666, 0.6666666666666666]}
          amplitude={1}
          distance={0.2}
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
