/**
 * v2e Portal - Quick Launch Modal Component
 *
 * Spotlight-style app launcher with keyboard navigation
 * Phase 3: Quick Launch System
 * Backend Independence: Works completely offline
 */

'use client';

import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { Search, AppWindow } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useDesktopStore } from '@/lib/desktop/store';
import { getActiveApps } from '@/lib/desktop/app-registry';
import { Z_INDEX, WindowState } from '@/types/desktop';
import type { AppRegistryEntry } from '@/lib/desktop/app-registry';

interface QuickLaunchModalProps {
  isVisible: boolean;
  onClose: () => void;
}

/**
 * Quick launch modal component
 * Cmd+K triggered app launcher with search and keyboard navigation
 */
export function QuickLaunchModal({ isVisible, onClose }: QuickLaunchModalProps) {
  const { openWindow } = useDesktopStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Get all active apps from registry
  const allApps = useMemo(() => getActiveApps(), []);

  // Filter apps based on search query
  const filteredApps = useMemo(() => {
    if (!searchQuery.trim()) {
      return allApps;
    }

    const query = searchQuery.toLowerCase();
    return allApps.filter(app =>
      app.name.toLowerCase().includes(query) ||
      app.id.toLowerCase().includes(query) ||
      app.category.toLowerCase().includes(query)
    );
  }, [searchQuery, allApps]);

  // Reset selection when search results change
  useEffect(() => {
    setSelectedIndex(0);
  }, [filteredApps.length]);

  // Focus input when modal opens
  useEffect(() => {
    if (isVisible && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isVisible]);

  // Keyboard navigation
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'Escape':
        onClose();
        e.preventDefault();
        break;

      case 'ArrowDown':
        setSelectedIndex(prev => (prev + 1) % filteredApps.length);
        e.preventDefault();
        break;

      case 'ArrowUp':
        setSelectedIndex(prev => (prev - 1 + filteredApps.length) % filteredApps.length);
        e.preventDefault();
        break;

      case 'Enter':
        if (filteredApps[selectedIndex]) {
          handleLaunch(filteredApps[selectedIndex]);
        }
        e.preventDefault();
        break;

      case 'Tab':
        e.preventDefault();
        break;
    }
  }, [filteredApps, selectedIndex, onClose]);

  // Launch app
  const handleLaunch = useCallback((app: AppRegistryEntry) => {
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
    setSearchQuery('');
    onClose();
  }, [openWindow, onClose]);

  // Handle click outside
  const handleClickOutside = useCallback((e: MouseEvent) => {
    if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
      onClose();
    }
  }, [onClose]);

  useEffect(() => {
    if (isVisible) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isVisible, handleClickOutside]);

  if (!isVisible) return null;

  return (
    <div
      className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-start justify-center pt-[20vh] z-50"
      style={{ zIndex: Z_INDEX.QUICK_LAUNCH_MODAL }}
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="quick-launch-title"
    >
      <motion.div
        ref={containerRef}
        initial={{ opacity: 0, scale: 0.95, y: -20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: -20 }}
        transition={{ duration: 0.15, ease: 'easeOut' }}
        className="w-full max-w-2xl mx-4 bg-white rounded-xl shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search input */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-200">
          <Search className="w-5 h-5 text-gray-400" />
          <input
            ref={inputRef}
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search apps..."
            className="flex-1 outline-none text-gray-800 placeholder-gray-400 bg-transparent"
            id="quick-launch-search"
            autoComplete="off"
          />
          <span className="text-xs text-gray-400">
            {filteredApps.length} {filteredApps.length === 1 ? 'app' : 'apps'}
          </span>
        </div>

        {/* App list */}
        <div className="max-h-[60vh] overflow-y-auto">
          <AnimatePresence mode="wait">
            {filteredApps.length === 0 ? (
              <motion.div
                key="no-results"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="px-4 py-8 text-center text-gray-500"
              >
                No apps found matching "{searchQuery}"
              </motion.div>
            ) : (
              <motion.div
                key="app-list"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
              >
                {filteredApps.map((app, index) => (
                  <motion.button
                    key={app.id}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.02 }}
                    onClick={() => handleLaunch(app)}
                    className={`
                      w-full flex items-center gap-3 px-4 py-3 text-left
                      hover:bg-blue-50 transition-colors
                      ${index === selectedIndex ? 'bg-blue-50' : ''}
                    `}
                  >
                    {/* App icon */}
                    <div
                      className="w-10 h-10 rounded-lg flex items-center justify-center"
                      style={{ backgroundColor: app.iconColor || '#3b82f6' }}
                    >
                      <AppWindow className="w-5 h-5 text-white" />
                    </div>

                    {/* App info */}
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-gray-800 truncate">
                        {app.name}
                      </div>
                      <div className="text-xs text-gray-500 truncate">
                        {app.category}
                      </div>
                    </div>

                    {/* Keyboard hint */}
                    <div className="text-xs text-gray-400">
                      {index === selectedIndex && (
                        <kbd className="px-1.5 py-0.5 bg-gray-100 rounded border border-gray-200">
                          Enter
                        </kbd>
                      )}
                    </div>
                  </motion.button>
                ))}
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Footer hints */}
        <div className="px-4 py-2 bg-gray-50 border-t border-gray-200 flex items-center gap-4 text-xs text-gray-500">
          <span>
            <kbd className="px-1 py-0.5 bg-white rounded border border-gray-200">↑↓</kbd> Navigate
          </span>
          <span>
            <kbd className="px-1 py-0.5 bg-white rounded border border-gray-200">Enter</kbd> Launch
          </span>
          <span>
            <kbd className="px-1 py-0.5 bg-white rounded border border-gray-200">Esc</kbd> Close
          </span>
        </div>
      </motion.div>
    </div>
  );
}

/**
 * Hook to handle Cmd+K keyboard shortcut
 */
export function useQuickLaunchShortcut() {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd+K or Ctrl+K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsVisible(prev => !prev);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return {
    isVisible,
    show: () => setIsVisible(true),
    hide: () => setIsVisible(false),
    toggle: () => setIsVisible(prev => !prev),
  };
}
