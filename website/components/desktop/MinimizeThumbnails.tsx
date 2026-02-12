/**
 * v2e Portal - Minimize Animation & Thumbnails
 *
 * Enhanced minimize animation with dock thumbnails
 * Phase 4.4: Minimized Window Thumbnails
 * Backend Independence: Works completely offline
 */

'use client';

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { WindowConfig } from '@/types/desktop';
import { Z_INDEX } from '@/types/desktop';
import { useDesktopStore } from '@/lib/desktop/store';

interface WindowThumbnailProps {
  window: WindowConfig;
  onRestore: () => void;
}

/**
 * Window thumbnail component
 * Shows preview of minimized window in dock
 */
function WindowThumbnail({ window, onRestore }: WindowThumbnailProps) {
  const [isHovering, setIsHovering] = useState(false);

  return (
    <div
      className="relative group cursor-pointer"
      onMouseEnter={() => setIsHovering(true)}
      onMouseLeave={() => setIsHovering(false)}
      onClick={onRestore}
      aria-label={`Restore ${window.title}`}
      title={`Click to restore ${window.title}`}
    >
      {/* Thumbnail background */}
      <motion.div
        className={`
          absolute inset-0 rounded-lg overflow-hidden
          ${isHovering ? 'ring-2 ring-blue-500 scale-110' : 'ring-1 ring-gray-300'}
          transition-all duration-200
        `}
        style={{
          width: '180px',
          height: '120px',
          zIndex: Z_INDEX.DOCK_THUMBNAIL,
        }}
        initial={{ opacity: 0, scale: 0.8 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
      >
        {/* Miniature window */}
        <div
          className="absolute top-2 left-2 w-[95%] h-[85%] bg-gradient-to-br from-blue-500 to-purple-600 rounded-md border border-white/20 shadow-xl"
          style={{
            transform: `translateY(25%) translateX(25%)`,
          }}
        >
          {/* Titlebar */}
          <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-100 flex items-center px-2 rounded-t">
            <span className="text-xs font-semibold text-white">
              {window.title}
            </span>
          </div>

          {/* Content area */}
          <div className="flex-1 p-1">
            {/* Simulated app content */}
            <div className="w-full h-[70%] bg-white rounded flex items-center justify-center">
              <div className="text-3xl font-bold text-gray-800">
                {window.title[0]}
              </div>
              <div className="text-xs text-gray-400">
                Content
              </div>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Hover overlay */}
      <motion.div
        className="absolute inset-0 rounded-lg bg-black/50 backdrop-blur-sm"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.15, ease: 'easeOut' }}
      >
        <div className="absolute inset-x-0 bottom-2 right-2 text-white text-xs">
          Hover to preview
        </div>
      </motion.div>
      </div>
    </div>
  );
}

/**
 * Enhanced dock with thumbnails component
 */
interface DockWithThumbnailsProps {
  className?: string;
}

export function DockWithThumbnails({ className = '' }: DockWithThumbnailsProps) {
  const { windows, dock } = useDesktopStore();
  const minimizedWindows = Object.values(windows).filter(w => w.isMinimized);

  return (
    <nav
      className={`
        fixed bottom-4 left-1/2 right-1/2 -translate-x-1/2
        bg-white/10 backdrop-blur-md border border-white/10 rounded-2xl
        flex items-end gap-2 p-3 ${className}
      `}
      style={{
        zIndex: Z_INDEX.DOCK,
      }}
      role="navigation"
      aria-label="Application dock with thumbnails"
    >
      {/* Regular dock items */}
      {dock.items.map(item => {
        const isAppMinimized = minimizedWindows.some(w => w.appId === item.appId);

        return (
          <div key={item.appId} className="relative flex flex-col items-center group">
            {/* Running indicator */}
            <div className="absolute -bottom-1 w-1.5 h-1.5 rounded-full bg-blue-500" />

            {/* Minimized window thumbnail */}
            {minimizedWindows.map(win => {
              if (win.appId === item.appId) {
                return (
                  <WindowThumbnail
                    key={win.id}
                    window={win}
                    onRestore={() => {
                      const { restoreWindow } = useDesktopStore.getState();
                      restoreWindow(win.id);
                    }}
                  />
                );
              }
            })?.[0] // Show first minimized window
          )}
          </div>
        </div>
      )}
    </nav>
  );
}
