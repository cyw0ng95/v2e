/**
 * v2e Portal - Desktop Area Component
 *
 * Main desktop area with icons and wallpaper
 * Renders without backend dependency
 */

'use client';

import React from 'react';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX } from '@/types/desktop';
import type { DesktopIcon as DesktopIconType } from '@/types/desktop';

/**
 * Desktop icon component
 */
function DesktopIcon({ icon }: { icon: DesktopIconType }) {
  const { selectDesktopIcon } = useDesktopStore();
  const { updateDesktopIconPosition } = useDesktopStore();

  const handleClick = () => {
    selectDesktopIcon(icon.id);
  };

  const handleDoubleClick = () => {
    // TODO: Will be implemented in Phase 2 with window system
    console.log('Launch app:', icon.appId);
  };

  const handleDragStart = (e: React.DragEvent) => {
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', icon.id);
  };

  return (
    <div
      onClick={handleClick}
      onDoubleClick={handleDoubleClick}
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
        <svg className="w-6 h-6" fill="none" stroke="white" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9.75 17L9 20l-1.5-1.5L6 16.25l-3 3.75-3.75-2.15.25z"
          />
        </svg>
      </div>

      {/* Icon label */}
      <div className="text-center">
        <span className="text-xs font-medium text-white drop-shadow-lg">
          {icon.appId}
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
  const { desktopIcons, theme } = useDesktopStore();

  return (
    <main
      className="absolute inset-0 overflow-hidden"
      style={{
        background: theme.wallpaper,
        zIndex: Z_INDEX.DESKTOP_WALLPAPER,
      }}
      role="main"
      aria-label="Desktop area"
    >
      {/* Desktop icons grid */}
      <div className="relative w-full h-full">
        {desktopIcons.map(icon => (
          <DesktopIcon key={icon.id} icon={icon} />
        ))}
      </div>
    </main>
  );
}
