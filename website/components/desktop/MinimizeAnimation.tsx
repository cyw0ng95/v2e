/**
 * v2e Portal - Minimize Animation Component
 *
 * Genie effect animation when minimizing window to dock
 * Phase 2/3: Window-Dock Integration
 */

'use client';

import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX } from '@/types/desktop';

interface MinimizeAnimationProps {
  windowId: string;
  windowPosition: { x: number; y: number };
  windowSize: { width: number; height: number };
  onComplete: () => void;
}

/**
 * Minimize animation component
 * Animates window shrinking and moving to dock position
 */
export function MinimizeAnimation({
  windowId,
  windowPosition,
  windowSize,
  onComplete,
}: MinimizeAnimationProps) {
  const { dock } = useDesktopStore();

  // Calculate dock thumbnail position
  // Dock is at bottom, centered horizontally
  const dockCenterX = window.innerWidth / 2;
  const dockY = window.innerHeight - 80; // Dock height

  return (
    <AnimatePresence mode="wait">
      <motion.div
        initial={{ opacity: 0, scale: 1 }}
        animate={{
          opacity: [0, 1, 0],
          scale: [1, 0.3, 0],
          x: [windowPosition.x, dockCenterX - windowSize.width / 2, dockCenterX - windowSize.width / 2],
          y: [windowPosition.y, dockY],
        }}
        transition={{
          duration: 300,
          ease: 'easeInOut',
          times: [0, 0.7, 0.3],
        }}
        onAnimationComplete={() => onComplete()}
        className="fixed pointer-events-none"
        style={{ zIndex: Z_INDEX.QUICK_LAUNCH_MODAL - 1 }}
      >
        {/* Trail effect */}
        <motion.div
          initial={{ opacity: 0.6 }}
          animate={{
            opacity: [0.6, 0],
            scale: [1, 0.8],
          }}
          transition={{ duration: 200, ease: 'easeOut' }}
          className="absolute inset-0 bg-gradient-to-br from-blue-500/30 to-purple-500/30 rounded-lg blur-sm"
          style={{
            left: `${windowPosition.x + windowSize.width / 2}px`,
            top: `${windowPosition.y + windowSize.height / 2}px`,
            width: `${windowSize.width}px`,
            height: `${windowSize.height}px`,
          }}
        />
      </motion.div>
    </AnimatePresence>
  );
}
