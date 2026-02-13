/**
 * v2e Portal - Wallpaper Selector Component
 *
 * Wallpaper selection modal with gradient options
 * Phase 4: Wallpaper System
 * Backend Independence: Works completely offline
 */

'use client';

import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX } from '@/types/desktop';

interface WallpaperOption {
  id: string;
  name: string;
  gradient: string;
  preview: string;
}

const wallpaperOptions: WallpaperOption[] = [
  {
    id: 'gradient-purple',
    name: 'Purple Dream',
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    preview: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  },
  {
    id: 'gradient-blue',
    name: 'Ocean Blue',
    gradient: 'linear-gradient(135deg, #3b82f6 0%, #2191ff 100%)',
    preview: 'linear-gradient(135deg, #3b82f6 0%, #2191ff 100%)',
  },
  {
    id: 'gradient-sunset',
    name: 'Sunset Glow',
    gradient: 'linear-gradient(135deg, #f97316 0%, #deadd5d 100%)',
    preview: 'linear-gradient(135deg, #f97316 0%, #deadd5d 100%)',
  },
  {
    id: 'gradient-forest',
    name: 'Forest Calm',
    gradient: 'linear-gradient(135deg, #134e4a3 0%, #119e4a8 100%)',
    preview: 'linear-gradient(135deg, #134e4a3 0%, #119e4a8 100%)',
  },
  {
    id: 'gradient-pink',
    name: 'Cotton Candy',
    gradient: 'linear-gradient(135deg, #fdf2f8 0%, #fde3f8 100%)',
    preview: 'linear-gradient(135deg, #fdf2f8 0%, #fde3f8 100%)',
  },
  {
    id: 'gradient-gray',
    name: 'Silver Mist',
    gradient: 'linear-gradient(135deg, #6b7280 0%, #8c94a8 100%)',
    preview: 'linear-gradient(135deg, #6b7280 0%, #8c94a8 100%)',
  },
  {
    id: 'gradient-dark',
    name: 'Midnight',
    gradient: 'linear-gradient(135deg, #1c1c1e 0%, #2c3e50 100%)',
    preview: 'linear-gradient(135deg, #1c1c1e 0%, #2c3e50 100%)',
  },
];

/**
 * Wallpaper selector modal component
 */
export function WallpaperSelector() {
  const { theme, setWallpaper } = useDesktopStore();
  const [isOpen, setIsOpen] = useState(false);

  const handleSelectWallpaper = (gradient: string) => {
    setWallpaper(gradient);
    setIsOpen(false);
  };

  return (
    <>
      {/* Settings button in menu bar opens this modal */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
            className="fixed inset-0 flex items-center justify-center bg-black/50 backdrop-blur-sm z-50 p-8"
            style={{ zIndex: Z_INDEX.SETTINGS_MODAL }}
            role="dialog"
            aria-modal="true"
            aria-labelledby="wallpaper-modal-title"
          >
            <div className="bg-white rounded-xl shadow-2xl p-6 max-w-4xl w-full">
              <h2
                id="wallpaper-modal-title"
                className="text-xl font-semibold text-gray-900 mb-4"
              >
                Choose Wallpaper
              </h2>

              <div className="grid grid-cols-3 gap-4">
                {wallpaperOptions.map(option => (
                  <motion.button
                    key={option.id}
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.15, ease: 'easeOut' }}
                    whileHover={{ scale: 1.05 }}
                    onClick={() => handleSelectWallpaper(option.gradient)}
                    className="relative group"
                  >
                    {/* Preview */}
                    <div
                      className="w-full aspect-video rounded-lg overflow-hidden mb-2 border-2 border-transparent hover:border-blue-500"
                      style={{
                        background: option.gradient,
                        backgroundImage: `url(${option.preview})`,
                      }}
                    >
                      <div className="absolute inset-0 flex items-center justify-center text-white">
                        <span className="text-2xl font-bold">{option.name}</span>
                      </div>
                    </div>

                    {/* Option label */}
                    <div className="text-center">
                      <span className="text-sm font-medium text-gray-700">{option.name}</span>
                    </div>
                  </motion.button>
                ))}
              </div>

              {/* Close button */}
              <div className="mt-6 flex justify-end">
                <motion.button
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  onClick={() => setIsOpen(false)}
                  className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-lg text-gray-700 transition-colors"
                  whileHover={{ scale: 1.02 }}
                >
                  Cancel
                </motion.button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Trigger button - styled to look like settings */}
      <div className="fixed top-0 right-0 p-4 z-50" aria-label="Wallpaper settings">
        <button
          onClick={() => setIsOpen(true)}
          className="p-2 rounded-lg hover:bg-accent hover:text-accent-foreground transition-all duration-200 cursor-pointer focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
          aria-label="Open wallpaper settings"
          title="Change wallpaper"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 16l4.004 0c4.01-4.016c1.651.651.65c0 1.229.501.501.5-1.52 1.52-.5-5.765a.7 7-5.77-7-.275.275-.275-.275.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.275-.775"
            />
          </svg>
        </button>
      </div>
    </>
  );
}
