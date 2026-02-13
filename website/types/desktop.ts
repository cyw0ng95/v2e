/**
 * v2e Portal - Desktop Type Definitions
 *
 * Architecture Principles:
 * - Backend Independence: Desktop MUST function without backend support
 * - All UI components render without API dependency
 * - State persisted to localStorage (not database)
 */

// ============================================================================
// CORE STATE INTERFACES
// ============================================================================

/**
 * Application categories for organizing apps in the dock
 */
export type AppCategory =
  | 'Security'
  | 'Database'
  | 'Learning'
  | 'System'
  | 'Utility'
  | 'Reference'
  | 'Tool';

/**
 * Desktop icon representation
 * Used for desktop shortcut icons
 */
export interface DesktopIcon {
  id: string;
  appId: string; // Reference to APP_REGISTRY
  position: {
    x: number;
    y: number;
  };
  isSelected: boolean;
}

/**
 * Window state enumeration
 * Tracks lifecycle of each window
 */
export enum WindowState {
  Unopened = 'unopened',
  Opening = 'opening',
  Open = 'open',
  Focusing = 'focusing',
  Focused = 'focused',
  Minimizing = 'minimizing',
  Minimized = 'minimized',
  Maximizing = 'maximizing',
  Maximized = 'maximized',
  Restoring = 'restoring',
  Closing = 'closing',
  Closed = 'closed',
}

/**
 * Window configuration
 * Core window properties
 */
export interface WindowConfig {
  id: string;
  appId: string;
  title: string;
  state: WindowState;

  // Position and size
  position: {
    x: number;
    y: number;
  };
  size: {
    width: number;
    height: number;
  };

  // State flags
  isMinimized: boolean;
  isMaximized: boolean;
  isFocused: boolean;
  zIndex: number;

  // Constraints (from app registry)
  minWidth: number;
  minHeight: number;
  maxWidth?: number;
  maxHeight?: number;
}

/**
 * Dock configuration
 */
export interface DockConfig {
  items: DockItem[];
  isVisible: boolean;
  autoHide: boolean; // Auto-hide dock when mouse leaves
  autoHideDelay: number; // Delay in ms before hiding (default: 200)
  size: 'small' | 'medium' | 'large';
  position: 'bottom' | 'left' | 'right';
}

/**
 * Dock item representation
 */
export interface DockItem {
  appId: string;
  isRunning: boolean;
  isIndicator: boolean;
  app?: {
    id: string;
    name: string;
    icon: string;
    iconColor?: string;
  };
}

/**
 * Desktop theme configuration
 */
export interface ThemeConfig {
  mode: 'light' | 'dark';
  wallpaper: string; // Gradient CSS value
}

/**
 * Complete desktop state
 * Root state object managed by Zustand
 */
export interface DesktopState {
  // Icons
  desktopIcons: DesktopIcon[];

  // Windows
  windows: Record<string, WindowConfig>;
  focusedWindowId: string | null;

  // Dock
  dock: DockConfig;

  // Theme
  theme: ThemeConfig;

  // Widgets
  widgets: WidgetConfig[];

  // Network status
  isOnline: boolean;
}

/**
 * Widget configuration
 */
export interface WidgetConfig {
  id: string;
  type: 'clock' | 'calendar';
  position: {
    x: number;
    y: number;
  };
  isVisible: boolean;
}

// ============================================================================
// Z-INDEX CONSTANTS
// ============================================================================

/**
 * Z-index layering for desktop elements
 * Prevents third-party library conflicts
 */
export const Z_INDEX = {
  MENU_BAR: 2000,
  QUICK_LAUNCH_MODAL: 1500,
  SETTINGS_MODAL: 1400, // Settings modal
  CONTEXT_MENU: 1000,
  FOCUSED_WINDOW_BASE: 600, // Base for focused windows
  WINDOW_MIN: 100, // Minimum for inactive windows
  WINDOW_MAX: 999, // Maximum before context menu layer
  DOCK_THUMBNAIL: 75, // Thumbnails above dock
  DOCK: 50,
  DESKTOP_ICONS: 10,
  DESKTOP_WALLPAPER: 0,
} as const;

// ============================================================================
// TYPE GUARDS
// ============================================================================

/**
 * Type guard for window config
 */
export function isValidWindowConfig(obj: unknown): obj is WindowConfig {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    'id' in obj &&
    'appId' in obj &&
    'position' in obj &&
    'size' in obj &&
    'state' in obj
  );
}

/**
 * Type guard for desktop icon
 */
export function isValidDesktopIcon(obj: unknown): obj is DesktopIcon {
  return (
    typeof obj === 'object' &&
    obj !== null &&
    'id' in obj &&
    'appId' in obj &&
    'position' in obj
  );
}
