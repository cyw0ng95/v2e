/**
 * v2e Portal - Window Manager
 *
 * Orchestrates window rendering with animations
 * Phase 2: Window System - Animation Layer
 */

'use client';

import { AnimatePresence } from 'framer-motion';
import { useDesktopStore } from '@/lib/desktop/store';
import { AppWindow } from './AppWindow';
import { Z_INDEX } from '@/types/desktop';

/**
 * Window manager component
 * Handles animated window entry/exit
 */
export function WindowManager() {
  const { windows, focusedWindowId } = useDesktopStore();
  // Only render non-minimized windows
  const windowList = Object.values(windows).filter(w => !w.isMinimized);

  return (
    <div
      className="fixed inset-0 pointer-events-none"
      style={{ zIndex: Z_INDEX.DESKTOP_ICONS }}
      role="region"
      aria-label="Window container"
    >
      <AnimatePresence mode="popLayout">
        {windowList.map((window) => (
          <AppWindow
            key={window.id}
            window={window}
          />
        ))}
      </AnimatePresence>
    </div>
  );
}
