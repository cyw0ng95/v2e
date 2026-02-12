/**
 * v2e Portal - Context Menu Component
 *
 * Right-click context menu system for desktop elements
 * Phase 3: Context Menu System
 * Backend Independence: Works completely offline
 */

'use client';

import React, { useEffect, useCallback, useState, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useDesktopStore } from '@/lib/desktop/store';
import { Z_INDEX } from '@/types/desktop';

/**
 * Context menu item interface
 */
export interface ContextMenuItem {
  id: string;
  label: string;
  icon?: React.ReactNode;
  shortcut?: string;
  disabled?: boolean;
  separator?: boolean;
  action: () => void;
}

/**
 * Context menu position calculation
 * Ensures menu stays within viewport bounds
 */
function calculateMenuPosition(
  x: number,
  y: number,
  menuWidth: number = 200,
  menuHeight: number = 300
): { x: number; y: number } {
  const maxX = window.innerWidth - menuWidth - 10;
  const maxY = window.innerHeight - menuHeight - 10;

  return {
    x: Math.min(x, maxX),
    y: Math.min(y, maxY),
  };
}

/**
 * Context menu component
 */
interface ContextMenuProps {
  isVisible: boolean;
  position: { x: number; y: number } | null;
  items: ContextMenuItem[];
  onClose: () => void;
}

export function ContextMenu({ isVisible, position, items, onClose }: ContextMenuProps) {
  const [selectedIndex, setSelectedIndex] = useState(-1);

  // Close menu on click outside
  const handleClickOutside = useCallback((e: MouseEvent) => {
    if (isVisible) {
      const target = e.target as HTMLElement;
      if (!target.closest('[data-context-menu]')) {
        onClose();
      }
    }
  }, [isVisible, onClose]);

  // Keyboard navigation
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!isVisible) return;

    switch (e.key) {
      case 'Escape':
        onClose();
        e.preventDefault();
        break;

      case 'ArrowDown':
        setSelectedIndex(prev => {
          const nextIndex = items.findIndex((item, idx) => idx > prev && !item.disabled && !item.separator);
          return nextIndex !== -1 ? nextIndex : prev;
        });
        e.preventDefault();
        break;

      case 'ArrowUp':
        setSelectedIndex(prev => {
          const enabledItems = items
            .map((item, idx) => ({ item, idx }))
            .filter(({ item }) => !item.disabled && !item.separator);

          if (enabledItems.length === 0) return -1;

          const currentIdx = enabledItems.findIndex(({ idx }) => idx === prev);
          const nextIdx = currentIdx <= 0 ? enabledItems.length - 1 : currentIdx - 1;
          return enabledItems[nextIdx]?.idx ?? -1;
        });
        e.preventDefault();
        break;

      case 'Enter':
        if (selectedIndex >= 0 && items[selectedIndex]) {
          items[selectedIndex].action();
          onClose();
        }
        e.preventDefault();
        break;
    }
  }, [isVisible, items, selectedIndex, onClose]);

  useEffect(() => {
    if (isVisible) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
      setSelectedIndex(-1);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isVisible, handleClickOutside, handleKeyDown]);

  // Calculate position to keep menu in viewport
  const calculatedPosition = useMemo(() => {
    if (!position) return { x: 0, y: 0 };
    return calculateMenuPosition(position.x, position.y);
  }, [position]);

  if (!isVisible || !position) return null;

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          data-context-menu="true"
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{ duration: 0.1, ease: 'easeOut' }}
          className="fixed bg-white rounded-lg shadow-xl border border-gray-200 py-1 min-w-[180px] max-w-[240px]"
          style={{
            left: `${calculatedPosition.x}px`,
            top: `${calculatedPosition.y}px`,
            zIndex: Z_INDEX.CONTEXT_MENU,
          }}
          role="menu"
          aria-orientation="vertical"
        >
          {items.map((item, index) => {
            if (item.separator) {
              return (
                <div key={item.id} className="my-1 border-t border-gray-200" />
              );
            }

            const isSelected = index === selectedIndex;

            return (
              <button
                key={item.id}
                onClick={() => {
                  if (!item.disabled) {
                    item.action();
                    onClose();
                  }
                }}
                onMouseEnter={() => setSelectedIndex(index)}
                onMouseLeave={() => setSelectedIndex(-1)}
                disabled={item.disabled}
                className={`
                  w-full flex items-center justify-between gap-2 px-3 py-2 text-sm
                  ${isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'}
                  ${item.disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                  transition-colors duration-75
                `}
                role="menuitem"
                aria-disabled={item.disabled}
              >
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  {item.icon && (
                    <span className="flex-shrink-0 text-gray-500">
                      {item.icon}
                    </span>
                  )}
                  <span className="truncate">{item.label}</span>
                </div>
                {item.shortcut && (
                  <span className="text-xs text-gray-400 flex-shrink-0">
                    {item.shortcut}
                  </span>
                )}
              </button>
            );
          })}
        </motion.div>
      )}
    </AnimatePresence>
  );
}

/**
 * Hook to manage context menu state
 */
export function useContextMenu() {
  const [isVisible, setIsVisible] = useState(false);
  const [position, setPosition] = useState<{ x: number; y: number } | null>(null);
  const [items, setItems] = useState<ContextMenuItem[]>([]);

  const show = useCallback((x: number, y: number, menuItems: ContextMenuItem[]) => {
    setPosition({ x, y });
    setItems(menuItems);
    setIsVisible(true);
  }, []);

  const hide = useCallback(() => {
    setIsVisible(false);
    // Don't clear items/position immediately to allow exit animation
    setTimeout(() => {
      setPosition(null);
      setItems([]);
    }, 100);
  }, []);

  return {
    isVisible,
    position,
    items,
    show,
    hide,
  };
}

/**
 * Preset context menus for different targets
 */
export const ContextMenuPresets = {
  /**
   * Desktop icon context menu
   */
  desktopIcon: (iconId: string, appId: string) => {
    const { openWindow, removeDesktopIcon } = useDesktopStore.getState();

    return [
      {
        id: 'open',
        label: 'Open',
        action: () => {
          // Launch app window (will be implemented with app registry integration)
          console.log('Open app:', appId);
        },
      },
      {
        id: 'new-window',
        label: 'New Window',
        action: () => {
          console.log('Open new window:', appId);
        },
      },
      {
        id: 'separator-1',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'remove',
        label: 'Remove from Desktop',
        action: () => {
          removeDesktopIcon(iconId);
        },
      },
      {
        id: 'separator-2',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'info',
        label: 'Show Info',
        action: () => {
          console.log('Show info for:', appId);
        },
      },
    ] as ContextMenuItem[];
  },

  /**
   * Dock item context menu (running)
   */
  dockItemRunning: (appId: string) => {
    return [
      {
        id: 'show-all',
        label: 'Show All Windows',
        action: () => {
          console.log('Show all windows for:', appId);
        },
      },
      {
        id: 'hide',
        label: 'Hide',
        action: () => {
          console.log('Hide:', appId);
        },
      },
      {
        id: 'separator-1',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'quit',
        label: 'Quit',
        action: () => {
          console.log('Quit:', appId);
        },
      },
      {
        id: 'separator-2',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'remove',
        label: 'Remove from Dock',
        action: () => {
          console.log('Remove from dock:', appId);
        },
      },
    ] as ContextMenuItem[];
  },

  /**
   * Dock item context menu (not running)
   */
  dockItemNotRunning: (appId: string) => {
    return [
      {
        id: 'open',
        label: 'Open',
        action: () => {
          console.log('Open app:', appId);
        },
      },
      {
        id: 'separator-1',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'remove',
        label: 'Remove from Dock',
        action: () => {
          console.log('Remove from dock:', appId);
        },
      },
    ] as ContextMenuItem[];
  },

  /**
   * Window context menu
   */
  window: (windowId: string) => {
    const { closeWindow, minimizeWindow, maximizeWindow } = useDesktopStore.getState();

    return [
      {
        id: 'close',
        label: 'Close',
        shortcut: '⌘W',
        action: () => {
          closeWindow(windowId);
        },
      },
      {
        id: 'minimize',
        label: 'Minimize',
        shortcut: '⌘M',
        action: () => {
          minimizeWindow(windowId);
        },
      },
      {
        id: 'maximize',
        label: 'Maximize',
        action: () => {
          maximizeWindow(windowId);
        },
      },
      {
        id: 'separator-1',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'move-back',
        label: 'Move to Back',
        action: () => {
          console.log('Move to back:', windowId);
        },
      },
      {
        id: 'keep-on-top',
        label: 'Keep on Top',
        action: () => {
          console.log('Keep on top:', windowId);
        },
      },
    ] as ContextMenuItem[];
  },

  /**
   * Desktop empty space context menu
   */
  desktop: () => {
    return [
      {
        id: 'new-folder',
        label: 'New Folder',
        action: () => {
          console.log('Create new folder');
        },
      },
      {
        id: 'separator-1',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'change-wallpaper',
        label: 'Change Wallpaper',
        action: () => {
          console.log('Change wallpaper');
        },
      },
      {
        id: 'separator-2',
        label: '',
        separator: true,
        action: () => {},
      },
      {
        id: 'sort-by',
        label: 'Sort By',
        action: () => {
          console.log('Sort desktop icons');
        },
      },
      {
        id: 'view-options',
        label: 'View Options',
        action: () => {
          console.log('View options');
        },
      },
    ] as ContextMenuItem[];
  },
};
