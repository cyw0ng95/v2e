/**
 * v2e Portal - Desktop State Management Store
 *
 * Architecture Principle: Backend Independence
 * - All desktop functions work without backend
 * - State persisted to localStorage
 * - No blocking API calls during initialization
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type {
  DesktopState,
  WindowConfig,
  DesktopIcon,
  DockConfig,
  ThemeConfig,
  WidgetConfig,
  WindowState,
} from '@/types/desktop';

// ============================================================================
// INITIAL STATE
// ============================================================================

/**
 * Initial desktop state
 * All features work without backend dependency
 */
const initialState: DesktopState = {
  desktopIcons: [],
  windows: {},
  focusedWindowId: null,
  dock: {
    items: [],
    isVisible: true,
    size: 'medium',
    position: 'bottom',
  },
  theme: {
    mode: 'dark',
    wallpaper: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  },
  widgets: [
    {
      id: 'clock-1',
      type: 'clock',
      position: { x: 20, y: 20 },
      isVisible: true,
    },
  ],
};

// ============================================================================
// Z-INDEX MANAGEMENT
// ============================================================================

/**
 * Calculate next z-index for focused window
 * Ensures windows layer correctly without conflicts
 */
function getNextFocusZIndex(currentWindows: Record<string, WindowConfig>): number {
  const windowZIndices = Object.values(currentWindows)
    .map(w => w.zIndex)
    .filter(z => z >= 100 && z <= 999); // Only windows

  if (windowZIndices.length === 0) {
    return 600; // First window base
  }

  const maxZ = Math.max(...windowZIndices);
  return Math.min(maxZ + 100, 999); // Cap at 999 to avoid context menu layer
}

/**
 * Reset window z-index to base value
 * Used when window is minimized or loses focus
 */
function getBaseZIndex(windowOrder: number): number {
  return 100 + windowOrder;
}

// ============================================================================
// STORE ACTIONS
// ============================================================================

interface DesktopActions {
  // Window State Transitions
  transitionWindow: (id: string, fromState: WindowState, toState: WindowState) => void;

  // Desktop Icons
  addDesktopIcon: (icon: Omit<DesktopIcon, 'id'>) => void;
  removeDesktopIcon: (id: string) => void;
  selectDesktopIcon: (id: string) => void;
  updateDesktopIconPosition: (id: string, position: { x: number; y: number }) => void;

  // Windows
  openWindow: (config: Omit<WindowConfig, 'id' | 'zIndex'>) => void;
  closeWindow: (id: string) => void;
  focusWindow: (id: string) => void;
  minimizeWindow: (id: string) => void;
  maximizeWindow: (id: string) => void;
  restoreWindow: (id: string) => void;
  updateWindowPosition: (id: string, position: { x: number; y: number }) => void;
  updateWindowSize: (id: string, size: { width: number; height: number }) => void;

  // Dock
  addDockItem: (item: Omit<DockItem, 'isRunning'>) => void;
  removeDockItem: (appId: string) => void;
  updateDockItemRunning: (appId: string, isRunning: boolean) => void;
  setDockVisibility: (isVisible: boolean) => void;
  setDockSize: (size: 'small' | 'medium' | 'large') => void;

  // Theme
  setThemeMode: (mode: 'light' | 'dark') => void;
  setWallpaper: (wallpaper: string) => void;

  // Widgets
  addWidget: (widget: Omit<WidgetConfig, 'id'>) => void;
  removeWidget: (id: string) => void;
  updateWidgetPosition: (id: string, position: { x: number; y: number }) => void;
  setWidgetVisibility: (id: string, isVisible: boolean) => void;

  // Global
  resetDesktop: () => void;
}
  // Desktop Icons
  addDesktopIcon: (icon: Omit<DesktopIcon, 'id'>) => void;
  removeDesktopIcon: (id: string) => void;
  selectDesktopIcon: (id: string) => void;
  updateDesktopIconPosition: (id: string, position: { x: number; y: number }) => void;

  // Windows
  openWindow: (config: Omit<WindowConfig, 'id' | 'zIndex'>) => void;
  closeWindow: (id: string) => void;
  focusWindow: (id: string) => void;
  minimizeWindow: (id: string) => void;
  maximizeWindow: (id: string) => void;
  restoreWindow: (id: string) => void;
  updateWindowPosition: (id: string, position: { x: number; y: number }) => void;
  updateWindowSize: (id: string, size: { width: number; height: number }) => void;

  // Dock
  addDockItem: (item: Omit<DockItem, 'isRunning'>) => void;
  removeDockItem: (appId: string) => void;
  updateDockItemRunning: (appId: string, isRunning: boolean) => void;
  setDockVisibility: (isVisible: boolean) => void;
  setDockSize: (size: 'small' | 'medium' | 'large') => void;

  // Theme
  setThemeMode: (mode: 'light' | 'dark') => void;
  setWallpaper: (wallpaper: string) => void;

  // Widgets
  addWidget: (widget: Omit<WidgetConfig, 'id'>) => void;
  removeWidget: (id: string) => void;
  updateWidgetPosition: (id: string, position: { x: number; y: number }) => void;
  setWidgetVisibility: (id: string, isVisible: boolean) => void;

  // Global
  resetDesktop: () => void;
}

type DesktopStore = DesktopState & DesktopActions & {
  transitionWindow: (id: string, fromState: WindowState, toState: WindowState) => void;
}

// ============================================================================
// STORE CREATION
// ============================================================================

/**
 * Create desktop store with Zustand
 * Persisted to localStorage, works completely offline
 */
const useDesktopStore = create<DesktopStore>()(
  persist(
    (set, get) => ({
      ...initialState,

      // Desktop Icons Actions
      addDesktopIcon: (icon) =>
        set(state => ({
          desktopIcons: [...state.desktopIcons, { ...icon, id: crypto.randomUUID() }],
        })),

      removeDesktopIcon: (id) =>
        set(state => ({
          desktopIcons: state.desktopIcons.filter(icon => icon.id !== id),
        })),

      selectDesktopIcon: (id) =>
        set(state => ({
          desktopIcons: state.desktopIcons.map(icon =>
            icon.id === id ? { ...icon, isSelected: true } : { ...icon, isSelected: false }
          ),
        })),

      updateDesktopIconPosition: (id, position) =>
        set(state => ({
          desktopIcons: state.desktopIcons.map(icon =>
            icon.id === id ? { ...icon, position } : icon
          ),
        })),

      // Window Actions
      openWindow: (config) =>
        set(state => {
          const id = crypto.randomUUID();
          const windowOrder = Object.keys(state.windows).length;

          return {
            windows: {
              ...state.windows,
              [id]: {
                ...config,
                id,
                state: 'open' as WindowState,
                zIndex: getNextFocusZIndex(state.windows),
              },
            },
            focusedWindowId: id,
            dock: {
              ...state.dock,
              items: state.dock.items.map(item =>
                item.appId === config.appId ? { ...item, isRunning: true, isIndicator: true } : item
              ),
            },
          };
        }),

      closeWindow: (id) =>
        set(state => {
          const { [id]: removedWindow, ...remainingWindows } = state.windows;
          return {
            windows: remainingWindows,
            focusedWindowId: state.focusedWindowId === id ? null : state.focusedWindowId,
            dock: {
              ...state.dock,
              items: state.dock.items.map(item => {
                const appWindows = Object.values(remainingWindows);
                const isStillRunning = appWindows.some(w => w.appId === item.appId);
                return { ...item, isRunning: isStillRunning, isIndicator: isStillRunning };
              }),
            },
          };
        }),

      focusWindow: (id) =>
        set(state => {
          const window = state.windows[id];
          if (!window) return {};

          return {
            windows: {
              ...state.windows,
              [id]: { ...window, state: 'focused' as WindowState, zIndex: getNextFocusZIndex(state.windows) },
            },
            focusedWindowId: id,
          };
        }),

      minimizeWindow: (id) =>
        set(state => {
          const window = state.windows[id];
          if (!window) return {};

          return {
            windows: {
              ...state.windows,
              [id]: { ...window, state: 'minimized' as WindowState, isMinimized: true },
            },
          };
        }),

      maximizeWindow: (id) =>
        set(state => {
          const window = state.windows[id];
          if (!window) return {};

          const isCurrentlyMaximized = window.state === 'maximized';
          const newState: WindowState = isCurrentlyMaximized ? 'restoring' : 'maximizing';
          const finalState: WindowState = isCurrentlyMaximized ? 'focused' : 'maximized';

          return {
            windows: {
              ...state.windows,
              [id]: {
                ...window,
                state: newState,
                isMaximized: !isCurrentlyMaximized,
              },
            },
          };
        }),

      restoreWindow: (id) =>
        set(state => {
          const window = state.windows[id];
          if (!window) return {};

          return {
            windows: {
              ...state.windows,
              [id]: { ...window, state: 'focused' as WindowState, isMinimized: false },
            },
            focusedWindowId: id,
          };
        }),

      updateWindowPosition: (id, position) =>
        set(state => ({
          windows: {
            ...state.windows,
            [id]: state.windows[id] ? { ...state.windows[id], position } : state.windows[id],
          },
        })),

      updateWindowSize: (id, size) =>
        set(state => ({
          windows: {
            ...state.windows,
            [id]: state.windows[id] ? { ...state.windows[id], size } : state.windows[id],
          },
        })),

      // Dock Actions
      addDockItem: (item) =>
        set(state => ({
          dock: {
            ...state.dock,
            items: [...state.dock.items, { ...item, isRunning: false }],
          },
        })),

      removeDockItem: (appId) =>
        set(state => ({
          dock: {
            ...state.dock,
            items: state.dock.items.filter(item => item.appId !== appId),
          },
        })),

      updateDockItemRunning: (appId, isRunning) =>
        set(state => ({
          dock: {
            ...state.dock,
            items: state.dock.items.map(item =>
              item.appId === appId ? { ...item, isRunning, isIndicator: isRunning } : item
            ),
          },
        })),

      setDockVisibility: (isVisible) =>
        set(state => ({
          dock: { ...state.dock, isVisible },
        })),

      setDockSize: (size) =>
        set(state => ({
          dock: { ...state.dock, size },
        })),

      // Theme Actions
      setThemeMode: (mode) =>
        set(state => ({
          theme: { ...state.theme, mode },
        })),

      setWallpaper: (wallpaper) =>
        set(state => ({
          theme: { ...state.theme, wallpaper },
        })),

      // Widget Actions
      addWidget: (widget) =>
        set(state => ({
          widgets: [...state.widgets, { ...widget, id: crypto.randomUUID() }],
        })),

      removeWidget: (id) =>
        set(state => ({
          widgets: state.widgets.filter(widget => widget.id !== id),
        })),

      updateWidgetPosition: (id, position) =>
        set(state => ({
          widgets: state.widgets.map(widget =>
            widget.id === id ? { ...widget, position } : widget
          ),
        })),

      setWidgetVisibility: (id, isVisible) =>
        set(state => ({
          widgets: state.widgets.map(widget =>
            widget.id === id ? { ...widget, isVisible } : widget
          ),
        })),

      // Global Actions
      resetDesktop: () => set(initialState),
    }),
    {
      name: 'v2e-desktop-storage',
      storage: createJSONStorage(() => localStorage),
      // Only persist critical state to avoid quota issues
      partialize: (state) => ({
        desktopIcons: state.desktopIcons,
        windows: state.windows,
        dock: state.dock,
        theme: state.theme,
        widgets: state.widgets,
      }),
    }
  )
);

// ============================================================================
// SELECTORS
// ============================================================================

/**
 * Get all windows sorted by z-index
 * Useful for rendering order
 */
export const selectWindowsSortedByZIndex = (state: DesktopState): WindowConfig[] => {
  return Object.values(state.windows).sort((a, b) => b.zIndex - a.zIndex);
};

/**
 * Get focused window
 */
export const selectFocusedWindow = (state: DesktopState): WindowConfig | undefined => {
  return state.focusedWindowId ? state.windows[state.focusedWindowId] : undefined;
};

/**
 * Get running apps from dock
 */
export const selectRunningDockItems = (state: DesktopState): DockItem[] => {
  return state.dock.items.filter(item => item.isRunning);
};
