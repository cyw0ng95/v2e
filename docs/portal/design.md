# v2e Portal Design Specification
## macOS Desktop-Inspired Application Portal

**Version:** 2.0.0
**Status:** Design Proposal
**Last Updated:** 2026-02-12
**Platform:** Desktop Only (1024px+)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Design Philosophy](#2-design-philosophy)
3. [Visual Design System](#3-visual-design-system)
4. [Desktop Architecture](#4-desktop-architecture)
5. [User Interface Design](#5-user-interface-design)
6. [Application Registry](#6-application-registry)
7. [Window Management](#7-window-management)
8. [Dock & App Launcher](#8-dock--app-launcher)
9. [Interaction Model](#9-interaction-model)
10. [Component Specification](#10-component-specification)
11. [Appendices](#appendices)

---

## 1. Overview

### 1.1 Purpose

The v2e Portal is a **desktop-native web experience** that provides a macOS-like desktop environment for accessing all security applications. Users land directly on a desktop interface where applications can be launched, managed, and interacted with in a familiar windowed environment.

**Key Characteristics:**
- **Desktop-First**: Designed exclusively for desktop screens (1024px minimum)
- **Desktop Metaphor**: Users land on a macOS-style desktop with app icons
- **Window Management**: Apps open in movable, resizable windows
- **Persistent State**: Desktop layout and window positions persist
- **Seamless Integration**: Feels like a native OS, not a website

### 1.2 Current Applications

| Application | Path | Purpose | Status |
|-------------|------|---------|--------|
| CVE Database | `/cve` | CVE vulnerability browsing and analysis | Active |
| CWE Database | `/cwe` | Common Weakness Enumeration reference | Active |
| CAPEC Database | `/capec` | Attack pattern encyclopedia | Active |
| ATT&CK | `/attack` | MITRE ATT&CK framework explorer | Active |
| CVSS Calculator | `/cvss` | CVSS v3.0/v3.1/v4.0 calculator | Active |
| GLC | `/glc` | Graphized Learning Canvas | Active |
| Mcards | `/mcards` | Flashcard-based learning system | Active |
| ETL Monitor | `/etl` | ETL engine monitoring dashboard | Active |
| System Monitor | `/sysmon` | System performance metrics | Planned |
| Bookmarks | `/bookmarks` | Personal bookmark management | Active |
| CCE Database | `/cce` | Common Configuration Enumeration | Planned |
| SSG Guides | `/ssg` | Security Scanning Guide references | Planned |
| ASVS | `/asvs` | Application Security Verification Standard | Planned |

### 1.3 Design Goals

1. **Native Desktop Feel**: Users feel like they're using a native OS, not a web app
2. **Direct Access**: Double-click any desktop icon to launch in a window
3. **Spatial Organization**: Arrange apps anywhere on the desktop
4. **Window Management**: Move, resize, minimize, maximize, close windows
5. **Persistent Layout**: Desktop and window positions saved between sessions
6. **Keyboard Native**: Full keyboard support like a real desktop OS

---

## 2. Design Philosophy

### 2.1 macOS Desktop Design Principles

| Principle | Description | Application |
|-----------|-------------|--------------|
| **Desktop Metaphor** | Apps live on a desktop surface | Icons arranged by user on desktop |
| **Window Hierarchy** | Windows layer with clear focus state | Active window on top, dimmed others |
| **Dock Access** | Frequently used apps in bottom dock | Quick access to favorites/recent apps |
| **Menu Bar** | System status and global controls | Top bar for search, theme, user |
| **Spatial Persistence** | Where you put things, they stay | Icon positions saved per user |
| **Smooth Animations** | Natural window open/close motion | 200ms scale/fade transitions |
| **Context Menus** | Right-click for options | Desktop and window context menus |

### 2.2 Desktop Layout Structure

```
?─────────────────────────────────────────────────────────────────────?
│  Menu Bar (28px)                                                │
│  ?────────┐  CVE Database  ▼  Window    Help                     │
│  │ v2e    │  Search... [Cmd+K]      🌙  👤                   │
│  └────────┘                                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Desktop Area (100vh - 28px - 80px = calc(100vh - 108px))    │
│                                                                  │
│  ?─────?  ?─────?        ?──────────────────────?              │
│  │ 🛡️CVE│  │ 🐛CWE│        │  CVSS Calculator    │              │
│  └─────┘  └─────┘        │  ?──────────────?   │              │
│                             │  │ Base Score  │   │              │
│  ?─────?  ?─────?        │  │ 7.5 (High) │   │              │
│  │ 🎯CAPEC│  │ 📍ATT&CK      │  └──────────────┘   │              │
│  └─────┘  └─────┘        │  Vector String       │              │
│                             │  CVSS:3.1/AV:N/... │              │
│  ?─────?  ?─────?        └──────────────────────┘              │
│  │ 🧮CVSS│  │ 📊GLC?│                                              │
│  └─────┘  └─────┘                                               │
│                                                                  │
│  ?─────?  ?─────?                                                │
│  │ 📚Mcards│  │ 🔖Bookmarks                                        │
│  └─────┘  └─────┘                                                │
│                                                                  │
├─────────────────────────────────────────────────────────────────────┤
│  Dock (80px, auto-hides when inactive)                             │
│  ?────┐?────┐?────┐?────┐?────┐?────┐?────┐?────┐    │
│  │ 🛡️ ?│ 🐛 ││ 🧮 ││ 📊 ││ 📚 ││ 🔄 ││ ⚙️ ││ 🔖 │    │
│  └────┘└────┘└────┘└────┘└────┘└────┘└────┘└────┘    │
│     CVE  CWE  CVSS  GLC  Cards  ETL  Sys   Marks              │
│                                                                  │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.3 Anti-Patterns to Avoid

- **Mobile Responsive**: No mobile design - this is desktop-only
- **Web Navigation Patterns**: No back button, no browser-like navigation
- **Page Transitions**: Apps open as windows, not page navigations
- **Emoji Icons**: Use SVG icons (Lucide React) instead of emojis
- **Heavy Gradients**: Subtle tints only, no aggressive gradient overlays
- **Abrupt Transitions**: All window operations use smooth animations
- **Low Contrast**: Ensure 4.5:1 minimum contrast ratio in all modes
- **Excessive Hover Effects**: Subtle glows preferred over scale transforms on icons

---

## 3. Visual Design System

### 3.1 Color Palette

#### Desktop Background

```css
/* Desktop wallpaper - subtle gradient */
--desktop-bg-light: linear-gradient(135deg,
  oklch(0.95 0.005 265) 0%,
  oklch(0.92 0.008 240) 50%,
  oklch(0.94 0.006 220) 100%
);

--desktop-bg-dark: linear-gradient(135deg,
  oklch(0.18 0.01 265) 0%,
  oklch(0.15 0.015 240) 50%,
  oklch(0.17 0.012 220) 100%
);

/* Desktop overlay for depth */
--desktop-overlay-light: oklch(1 0 0 / 20%);
--desktop-overlay-dark: oklch(0 0 0 / 20%);
```

#### Window Styling

```css
/* Window frame */
--window-bg: oklch(1 0 0 / 95%);
--window-bg-dark: oklch(0.15 0.008 264 / 95%);
--window-border: oklch(0 0 0 / 10%);
--window-border-dark: oklch(1 0 0 / 10%);
--window-shadow: 0 10px 40px oklch(0 0 0 / 15%),
                 0 4px 12px oklch(0 0 0 / 8%);
--window-shadow-dark: 0 10px 40px oklch(0 0 0 / 40%),
                    0 4px 12px oklch(0 0 0 / 20%);

/* Title bar */
--titlebar-bg: oklch(0.97 0.004 264 / 80%);
--titlebar-bg-dark: oklch(0.20 0.01 264 / 80%);
--titlebar-border: oklch(0 0 0 / 6%);
--titlebar-border-dark: oklch(1 0 0 / 6%);

/* Active window glow */
--window-glow: 0 0 0 1px oklch(0.52 0.16 265 / 30%);
--window-glow-dark: 0 0 0 1px oklch(0.68 0.20 265 / 30%);
```

#### Desktop Icons

```css
/* Icon container */
--icon-size: 80px;
--icon-bg: oklch(1 0 0 / 60%);
--icon-bg-dark: oklch(0 0 0 / 40%);
--icon-shadow: 0 2px 8px oklch(0 0 0 / 15%);
--icon-shadow-dark: 0 2px 8px oklch(0 0 0 / 30%);
--icon-radius: 16px;

/* Icon selection */
--icon-selected-bg: oklch(0.52 0.16 265 / 20%);
--icon-selected-border: oklch(0.52 0.16 265 / 50%);
--icon-selected-bg-dark: oklch(0.68 0.20 265 / 20%);
--icon-selected-border-dark: oklch(0.68 0.20 265 / 50%);

/* Icon label */
--icon-label-color: oklch(1 0 0);
--icon-label-shadow: 0 1px 3px oklch(0 0 0 / 50%);
--icon-label-color-dark: oklch(0.98 0.002 264);
--icon-label-shadow-dark: 0 1px 3px oklch(0 0 0 / 70%);
```

#### Dock Styling

```css
/* Dock container */
--dock-bg: oklch(1 0 0 / 75%);
--dock-bg-dark: oklch(0.15 0.008 264 / 75%);
--dock-border: oklch(0 0 0 / 10%);
--dock-border-dark: oklch(1 0 0 / 10%);
--dock-shadow: 0 -4px 20px oklch(0 0 0 / 10%);
--dock-shadow-dark: 0 -4px 20px oklch(0 0 0 / 30%);
--dock-radius: 20px;

/* Dock item */
--dock-item-size: 56px;
--dock-item-hover: 1.2; /* scale on hover */
--dock-item-bg: oklch(0.95 0.004 264);
--dock-item-bg-dark: oklch(0.22 0.01 264);
--dock-item-radius: 12px;

/* Active indicator */
--dock-indicator: oklch(0.52 0.16 265);
--dock-indicator-dark: oklch(0.68 0.20 265);
--dock-indicator-size: 4px;
```

### 3.2 Typography

```typescript
// Desktop typography scale
const desktopTypography = {
  // Menu bar
  menuBar: {
    fontSize: '13px',
    fontWeight: 500,
    letterSpacing: '0em',
  },
  // Window title
  windowTitle: {
    fontSize: '13px',
    fontWeight: 500,
    letterSpacing: '-0.01em',
  },
  // Icon label
  iconLabel: {
    fontSize: '11px',
    fontWeight: 500,
    letterSpacing: '0.01em',
    textShadow: '0 1px 3px oklch(0 0 0 / 50%)',
  },
  // Dock tooltip
  dockTooltip: {
    fontSize: '12px',
    fontWeight: 500,
    letterSpacing: '0em',
  },
};
```

### 3.3 Spacing & Layout

```typescript
// Desktop layout constants
const desktopLayout = {
  // Dimensions
  menuBarHeight: 28,
  dockHeight: 80,
  dockCollapsedHeight: 4,
  iconSize: 80,
  iconLabelHeight: 20,
  iconSpacing: 20,

  // Window
  windowMinWidth: 400,
  windowMinHeight: 300,
  windowDefaultWidth: 1024,
  windowDefaultHeight: 768,
  titlebarHeight: 36,
  windowBorderRadius: 10,
  windowPadding: 0,

  // Dock
  dockItemSize: 56,
  dockItemSpacing: 8,
  dockPadding: 8,
  dockMaxWidth: 800,

  // Grid
  desktopGrid: {
    horizontal: 20, // spacing between icons
    vertical: 20,
  },
};
```

### 3.4 Effects & Animations

```css
/* Window open animation */
@keyframes window-open {
  0% {
    opacity: 0;
    transform: scale(0.95);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.window-opening {
  animation: window-open 200ms ease-out;
}

/* Window close animation */
@keyframes window-close {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 0;
    transform: scale(0.95);
  }
}

.window-closing {
  animation: window-close 150ms ease-in forwards;
}

/* Minimize animation (genie effect) */
@keyframes minimize-genie {
  0% {
    clip-path: inset(0 0 0 0);
    transform: scale(1) translateY(0);
  }
  100% {
    clip-path: inset(0 0 100% 0);
    transform: scale(0.1) translateY(100vh);
  }
}

.window-minimizing {
  animation: minimize-genie 300ms ease-in forwards;
}

/* Dock item hover (magnification) */
.dock-item {
  transition: transform 150ms ease-out;
}

.dock-item:hover {
  transform: scale(1.2) translateY(-8px);
}

/* Icon selection pulse */
@keyframes icon-select-pulse {
  0% {
    background-color: oklch(0.52 0.16 265 / 20%);
  }
  50% {
    background-color: oklch(0.52 0.16 265 / 30%);
  }
  100% {
    background-color: oklch(0.52 0.16 265 / 20%);
  }
}

.icon-selected {
  animation: icon-select-pulse 2s ease-in-out infinite;
}

/* Window focus transition */
.window {
  transition: box-shadow 200ms ease-out;
}

.window-focused {
  box-shadow: var(--window-glow),
              var(--window-shadow);
}

.window-unfocused {
  opacity: 0.95;
  filter: brightness(0.98);
}

/* Desktop icon drag */
.icon-dragging {
  opacity: 0.7;
  transform: scale(1.1);
  cursor: grabbing;
}

/* Window resize handle */
.resize-handle {
  transition: background-color 150ms ease-out;
}

.resize-handle:hover {
  background-color: oklch(0.52 0.16 265 / 50%);
}
```

---

## 4. Desktop Architecture

### 4.1 File Structure

```
website/
├── app/
│   ├── desktop/
│   │   ├── page.tsx                 # Desktop (new landing page)
│   │   └── layout.tsx               # Desktop layout (no header/footer)
│   ├── portal/
│   │   ├── page.tsx                # Portal view (app launcher grid)
│   │   └── layout.tsx
│   └── (existing apps remain unchanged, can be opened in windows)
├── components/
│   ├── desktop/
│   │   ├── menu-bar.tsx            # Top menu bar
│   │   ├── desktop-area.tsx         # Desktop surface
│   │   ├── desktop-icon.tsx         # Desktop app icon
│   │   ├── dock.tsx                # Bottom dock
│   │   ├── dock-item.tsx            # Dock app icon
│   │   ├── window/
│   │   │   ├── app-window.tsx       # Main window component
│   │   │   ├── window-titlebar.tsx   # Window title bar
│   │   │   ├── window-controls.tsx   # Close/min/max buttons
│   │   │   ├── window-resize.tsx    # Resize handles
│   │   │   └── window-content.tsx    # App iframe/container
│   │   ├── context-menus.tsx        # Desktop/window context menus
│   │   ├── desktop-widgets.tsx      # Clock, weather widgets
│   │   └── wallpaper-selector.tsx   # Background picker
│   └── (existing components remain unchanged)
├── lib/
│   ├── desktop/
│   │   ├── desktop-state.ts         # Desktop state management
│   │   ├── window-manager.ts        # Window management logic
│   │   ├── dock-manager.ts         # Dock state
│   │   ├── icon-layout.ts          # Icon grid layout engine
│   │   └── keyboard-handler.ts     # Global keyboard shortcuts
│   ├── portal/
│   │   ├── app-registry.ts        # Application registry
│   │   └── app-categories.ts      # Category definitions
│   └── (existing libs remain unchanged)
└── types/
    └── desktop.ts                  # Desktop TypeScript types
```

### 4.2 State Architecture

```typescript
// types/desktop.ts

// Desktop icon position
interface DesktopIcon {
  id: string;              // App ID
  x: number;              // X position in pixels
  y: number;              // Y position in pixels
  gridX: number;           // Grid X position
  gridY: number;           // Grid Y position
}

// Window state
interface WindowState {
  id: string;              // Unique window ID
  appId: string;           // App ID (from registry)
  title: string;           // Window title
  x: number;              // X position
  y: number;              // Y position
  width: number;           // Window width
  height: number;          // Window height
  minimized: boolean;       // Is minimized to dock
  maximized: boolean;      // Is maximized
  focused: boolean;        // Has focus
  zIndex: number;          // Z-index for layering
  url?: string;            // Iframe URL or content path
}

// Dock configuration
interface DockConfig {
  items: string[];         // App IDs in dock (order preserved)
  autoHide: boolean;       // Auto-hide when inactive
  size: 'small' | 'medium' | 'large';
  position: 'bottom' | 'left' | 'right'; // Future expansion
}

// Desktop configuration
interface DesktopConfig {
  icons: DesktopIcon[];     // All desktop icons
  wallpaper: string;         // Wallpaper URL or gradient
  showIconLabels: boolean;   // Show labels under icons
  gridSize: number;         // Icon grid size (default: 100px)
  snapToGrid: boolean;     // Snap icons to grid
}

// Global desktop state
interface DesktopState {
  // Desktop
  config: DesktopConfig;

  // Windows
  windows: WindowState[];
  activeWindowId: string | null;
  topWindowId: string | null;

  // Dock
  dockConfig: DockConfig;
  activeApps: string[];     // Apps with open windows

  // Selection
  selectedIcons: string[];  // Selected icon IDs

  // Modals
  contextMenu: {
    visible: boolean;
    x: number;
    y: number;
    items: ContextMenuItem[];
  } | null;
}

// lib/desktop/desktop-state.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export const useDesktopStore = create<DesktopState>()(
  persist(
    (set, get) => ({
      // Desktop config
      config: {
        icons: getDefaultIconLayout(),
        wallpaper: 'default',
        showIconLabels: true,
        gridSize: 100,
        snapToGrid: true,
      },

      // Windows
      windows: [],
      activeWindowId: null,
      topWindowId: null,

      // Dock
      dockConfig: {
        items: ['cve', 'cvss', 'glc', 'mcards', 'etl'],
        autoHide: false,
        size: 'medium',
        position: 'bottom',
      },
      activeApps: [],

      // Selection
      selectedIcons: [],

      // Context menu
      contextMenu: null,

      // Actions
      openWindow: (appId: string) => {
        const app = APP_REGISTRY.find(a => a.id === appId);
        if (!app) return;

        const newWindow: WindowState = {
          id: `window-${Date.now()}`,
          appId,
          title: app.name,
          x: 100 + get().windows.length * 30,
          y: 50 + get().windows.length * 30,
          width: 1024,
          height: 768,
          minimized: false,
          maximized: false,
          focused: true,
          zIndex: 100 + get().windows.length,
        };

        set(state => ({
          windows: [...state.windows, newWindow],
          activeWindowId: newWindow.id,
          topWindowId: newWindow.id,
          activeApps: [...state.activeApps, appId],
        }));
      },

      closeWindow: (windowId: string) => {
        set(state => {
          const window = state.windows.find(w => w.id === windowId);
          if (!window) return state;

          const remainingWindows = state.windows.filter(w => w.id !== windowId);
          const stillHasAppWindow = remainingWindows.some(w => w.appId === window.appId);

          return {
            windows: remainingWindows,
            activeWindowId: state.activeWindowId === windowId
              ? remainingWindows[remainingWindows.length - 1]?.id || null
              : state.activeWindowId,
            topWindowId: remainingWindows[remainingWindows.length - 1]?.id || null,
            activeApps: stillHasAppWindow
              ? state.activeApps
              : state.activeApps.filter(id => id !== window.appId),
          };
        });
      },

      focusWindow: (windowId: string) => {
        const maxZ = Math.max(...get().windows.map(w => w.zIndex));

        set(state => ({
          windows: state.windows.map(w => ({
            ...w,
            focused: w.id === windowId,
            zIndex: w.id === windowId ? maxZ + 1 : w.zIndex,
          })),
          activeWindowId: windowId,
        }));
      },

      minimizeWindow: (windowId: string) => {
        set(state => ({
          windows: state.windows.map(w =>
            w.id === windowId ? { ...w, minimized: true, focused: false } : w
          ),
          activeWindowId: state.activeWindowId === windowId ? null : state.activeWindowId,
        }));
      },

      maximizeWindow: (windowId: string) => {
        set(state => ({
          windows: state.windows.map(w =>
            w.id === windowId
              ? { ...w, maximized: !w.maximized }
              : w
          ),
        }));
      },

      moveWindow: (windowId: string, x: number, y: number) => {
        set(state => ({
          windows: state.windows.map(w =>
            w.id === windowId ? { ...w, x, y } : w
          ),
        }));
      },

      resizeWindow: (windowId: string, width: number, height: number) => {
        set(state => ({
          windows: state.windows.map(w =>
            w.id === windowId ? { ...w, width, height } : w
          ),
        }));
      },

      updateIconPosition: (iconId: string, x: number, y: number) => {
        set(state => ({
          config: {
            ...state.config,
            icons: state.config.icons.map(icon =>
              icon.id === iconId ? { ...icon, x, y } : icon
            ),
          },
        }));
      },

      toggleDockItem: (appId: string) => {
        set(state => ({
          dockConfig: {
            ...state.dockConfig,
            items: state.dockConfig.items.includes(appId)
              ? state.dockConfig.items.filter(id => id !== appId)
              : [...state.dockConfig.items, appId],
          },
        }));
      },

      selectIcon: (iconId: string, multiSelect: boolean) => {
        set(state => ({
          selectedIcons: multiSelect
            ? [...state.selectedIcons, iconId]
            : [iconId],
        }));
      },

      clearSelection: () => set({ selectedIcons: [] }),

      showContextMenu: (x: number, y: number, items: ContextMenuItem[]) => {
        set({ contextMenu: { visible: true, x, y, items } });
      },

      hideContextMenu: () => set({ contextMenu: null }),
    }),
    {
      name: 'v2e-desktop-storage',
      partialize: (state) => ({
        config: state.config,
        dockConfig: state.dockConfig,
      }),
    }
  )
);

// Default icon layout (grid based)
function getDefaultIconLayout(): DesktopIcon[] {
  const positions = [
    { x: 20, y: 20 },
    { x: 120, y: 20 },
    { x: 220, y: 20 },
    { x: 320, y: 20 },
    { x: 20, y: 140 },
    { x: 120, y: 140 },
    { x: 220, y: 140 },
    { x: 320, y: 140 },
    { x: 20, y: 260 },
    { x: 120, y: 260 },
    { x: 220, y: 260 },
    { x: 320, y: 260 },
  ];

  return APP_REGISTRY.slice(0, 12).map((app, index) => ({
    id: app.id,
    x: positions[index]?.x || 20,
    y: positions[index]?.y || 20,
    gridX: index % 4,
    gridY: Math.floor(index / 4),
  }));
}
```

---

## 5. User Interface Design

### 5.1 Desktop Area (Main View)

**Users land here first** - a macOS-style desktop with app icons arranged in a grid.

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                  │
│  Desktop Area (clickable/draggable surface)                     │
│                                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                      │
│  │         │  │         │  │         │                      │
│  │   🛡️   │  │   🐛    │  │   🎯    │                      │
│  │         │  │         │  │         │                      │
│  │  CVE    │  │  CWE    │  │  CAPEC  │                      │
│  └─────────┘  └─────────┘  └─────────┘                      │
│                                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                      │
│  │   📍    │  │   🧮    │  │   📊    │                      │
│  │ ATT&CK  │  │  CVSS   │  │   GLC   │                      │
│  └─────────┘  └─────────┘  └─────────┘                      │
│                                                                  │
│  ┌─────────┐  ┌─────────┐                                      │
│  │   📚    │  │   🔖    │                                      │
│  │  Mcards  │  │Bookmarks│                                      │
│  └─────────┘  └─────────┘                                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────────┘
```

**Interactions:**
- **Click**: Select icon (shows selection highlight)
- **Double-click**: Open app in window
- **Drag**: Move icon to new position (snaps to grid)
- **Right-click**: Context menu (Open, Show Info, Remove from Desktop)

### 5.2 Menu Bar (Top)

```
┌─────────────────────────────────────────────────────────────────────┐
│  v2e  |  CVE Database  ▼  |  Window  |  Help                 │
│                                                    🔍 [Cmd+K]  🌙  👤  │
└─────────────────────────────────────────────────────────────────────┘

Left side:
- Apple-like menu with v2e logo
- Active app menu (changes based on focused window)
- Window menu (global window controls)

Right side:
- Search button (opens Spotlight-like quick launcher)
- Theme toggle
- User profile

Width: 28px
Background: Glass morphism (blur + semi-transparent)
```

### 5.3 Dock (Bottom)

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                  │
│  ┌────┐ ┌────┐ ┌────┐ ┌─��──┐ ┌────┐ ┌────┐ ┌────┐     │
│  │ 🛡️│ │ 🐛 │ │ 🧮 │ │ 📊 │ │ 📚 │ │ 🔄 │ │ ⚙️ │     │
│  └────┘ └────┘ └────┘ └────┘ └────┘ └────┘ └────┘     │
│     CVE    CWE   CVSS  GLC  Cards  ETL   Sys              │
│  ┌─────────────────────────────────────────────────────┐     │
│  │                                             │     │
│  └─────────────────────────────────────────────────────┘     │
│   Minimized windows (thumbnails)                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────────┘

Height: 80px (expanded), 4px (collapsed)
Background: Glass morphism with rounded corners
Icons: Scale up on hover (1.2x)
Separators: Between app groups (optional)
```

**Dock Features:**
- **Click**: Launch app or focus window
- **Right-click**: Context menu (Quit, Show All Windows, Options)
- **Cmd+Click**: Open new window
- **Minimized windows**: Show as thumbnails in dock
- **Drag to dock**: Add app to dock
- **Drag from dock**: Remove app from dock

### 5.4 Window Design

```
┌─────────────────────────────────────────────────────────────────────┐
│  CVE Browser                              ─ □ ✕                     │ 36px titlebar
├─────────────────────────────────────────────────────────────────────┤
│                                                                  │
│  (App content - iframe or direct component render)              │
│                                                                  │
│  Search...                                                      │
│  ┌────────────────────────────────────────┐                    │
│  │  CVE-2024-1234     High  9.8     │                    │
│  │  Remote code execution...             │                    │
│  └────────────────────────────────────────┘                    │
│                                                                  │
│  [Load More]                                                    │
│                                                                  │
└───────────────────────────────────────────────────────────────────────────┘

Titlebar:
- Left: App icon + title
- Right: Close (red), Minimize (yellow), Maximize (green)
- Drag area: Move window
- Double-click: Maximize/restore

Window controls:
- Close button: Red circle with ✕
- Minimize: Yellow circle with -
- Maximize: Green circle with □
- Hover shows symbol, normal is colored circle

Resize:
- 8px resize handles on edges and corners
- Cursor changes on hover
- Maintain minimum size (400x300)
```

### 5.5 Quick Launch (Spotlight-style)

```
┌─────────────────────────────────────────────────────────────────────┐
│           Search Apps and Files...                         ×      │
│                                                                  │
│  Applications                                                   │
│  ┌────────────────────────────────────────────────┐              │
│  │  🧮  CVSS Calculator                              Enter ││
│  │      Calculate CVSS scores                         │              │
│  └────────────────────────────────────────────────┘              │
│  ┌────────────────────────────────────────────────┐              │
│  │  🛡️  CVE Browser                                Enter ││
│  │      Browse CVE database                            │              │
│  └────────────────────────────────────────────────┘              │
│                                                                  │
│  ↓ Enter to open • Esc to close • ↑↓ to navigate                  │
└─────────────────────────────────────────────────────────────────────┘

Trigger: Cmd/Ctrl + K
Position: Centered on screen
Background: Glass morphism
Width: 600px
Max results: 8
```

---

## 6. Application Registry

```typescript
// lib/portal/app-registry.ts

export interface AppRegistryEntry {
  id: string;
  name: string;
  shortName: string;
  description: string;
  path: string;
  icon: {
    name: string;              // Lucide icon name
    component?: React.FC;     // Custom icon component
    color: string;           // Icon accent color
    gradient: string;        // Background gradient
  };
  category: AppCategory;
  status: AppStatus;
  window: {
    defaultWidth: number;
    defaultHeight: number;
    minWidth: number;
    minHeight: number;
    resizable: boolean;
    maximizable: boolean;
  };
  features: string[];
  tags: string[];
  metadata: {
    priority: number;
    isNew: boolean;
    isFeatured: boolean;
    inDockByDefault: boolean;
    onDesktopByDefault: boolean;
  };
}

export type AppCategory =
  | 'database'
  | 'tools'
  | 'learning'
  | 'analysis'
  | 'system'
  | 'utilities';

export type AppStatus =
  | 'active'
  | 'beta'
  | 'planned'
  | 'deprecated';

export const APP_REGISTRY: AppRegistryEntry[] = [
  // DATABASE APPS
  {
    id: 'cve',
    name: 'CVE Browser',
    shortName: 'CVE',
    description: 'Browse and analyze CVE vulnerability records.',
    path: '/cve',
    icon: {
      name: 'shield-alert',
      color: 'oklch(0.52 0.16 265)',
      gradient: 'from-blue-500 to-indigo-500',
    },
    category: 'database',
    status: 'active',
    window: { defaultWidth: 1200, defaultHeight: 800, minWidth: 400, minHeight: 300, resizable: true, maximizable: true },
    features: ['Search & Filter', 'Severity Analysis', 'Related CWEs', 'Timeline'],
    tags: ['vulnerability', 'cve', 'security'],
    metadata: { priority: 95, inDockByDefault: true, onDesktopByDefault: true },
  },
  {
    id: 'cwe',
    name: 'CWE Explorer',
    shortName: 'CWE',
    description: 'Common Weakness Enumeration reference.',
    path: '/cwe',
    icon: {
      name: 'bug',
      color: 'oklch(0.55 0.14 220)',
      gradient: 'from-blue-500 to-cyan-500',
    },
    category: 'database',
    status: 'active',
    window: { defaultWidth: 1200, defaultHeight: 800, minWidth: 400, minHeight: 300, resizable: true, maximizable: true },
    features: ['Weakness Catalog', 'Relationships', 'Mitigations'],
    tags: ['weakness', 'cwe', 'taxonomy'],
    metadata: { priority: 85, inDockByDefault: true, onDesktopByDefault: true },
  },
  {
    id: 'capec',
    name: 'CAPEC Directory',
    shortName: 'CAPEC',
    description: 'Attack pattern encyclopedia.',
    path: '/capec',
    icon: {
      name: 'target',
      color: 'oklch(0.58 0.18 290)',
      gradient: 'from-blue-500 to-purple-500',
    },
    category: 'database',
    status: 'active',
    window: { defaultWidth: 1200, defaultHeight: 800, minWidth: 400, minHeight: 300, resizable: true, maximizable: true },
    features: ['Attack Patterns', 'Execution Flow'],
    tags: ['attack', 'pattern', 'capec'],
    metadata: { priority: 75, onDesktopByDefault: true },
  },
  {
    id: 'attack',
    name: 'ATT&CK Navigator',
    shortName: 'ATT&CK',
    description: 'MITRE ATT&CK framework explorer.',
    path: '/attack',
    icon: {
      name: 'crosshair',
      color: 'oklch(0.55 0.18 25)',
      gradient: 'from-blue-500 to-red-500',
    },
    category: 'database',
    status: 'active',
    window: { defaultWidth: 1400, defaultHeight: 900, minWidth: 500, minHeight: 400, resizable: true, maximizable: true },
    features: ['Tactics & Techniques', 'Matrix View'],
    tags: ['attack', 'mitre', 'tactics'],
    metadata: { priority: 80, onDesktopByDefault: true },
  },

  // TOOLS
  {
    id: 'cvss',
    name: 'CVSS Calculator',
    shortName: 'CVSS',
    description: 'Calculate CVSS scores.',
    path: '/cvss',
    icon: {
      name: 'calculator',
      color: 'oklch(0.55 0.18 300)',
      gradient: 'from-violet-500 to-purple-500',
    },
    category: 'tools',
    status: 'active',
    window: { defaultWidth: 900, defaultHeight: 700, minWidth: 350, minHeight: 500, resizable: true, maximizable: true },
    features: ['CVSS v4.0', 'CVSS v3.1', 'Vector Export'],
    tags: ['cvss', 'calculator', 'scoring'],
    metadata: { priority: 98, isFeatured: true, isNew: true, inDockByDefault: true, onDesktopByDefault: true },
  },

  // LEARNING
  {
    id: 'glc',
    name: 'Graphized Learning Canvas',
    shortName: 'GLC',
    description: 'Interactive graph-based learning.',
    path: '/glc',
    icon: {
      name: 'gitgraph',
      color: 'oklch(0.58 0.16 148)',
      gradient: 'from-emerald-500 to-teal-500',
    },
    category: 'learning',
    status: 'active',
    window: { defaultWidth: 1400, defaultHeight: 900, minWidth: 600, minHeight: 400, resizable: true, maximizable: true },
    features: ['Graph Canvas', 'Presets', 'D3FEND'],
    tags: ['learning', 'graph', 'canvas'],
    metadata: { priority: 90, isFeatured: true, inDockByDefault: true, onDesktopByDefault: true },
  },
  {
    id: 'mcards',
    name: 'Memory Cards',
    shortName: 'Mcards',
    description: 'Flashcard-based learning system.',
    path: '/mcards',
    icon: {
      name: 'layers',
      color: 'oklch(0.60 0.14 160)',
      gradient: 'from-emerald-500 to-green-500',
    },
    category: 'learning',
    status: 'active',
    window: { defaultWidth: 1000, defaultHeight: 700, minWidth: 400, minHeight: 300, resizable: true, maximizable: true },
    features: ['Flashcards', 'Spaced Repetition'],
    tags: ['flashcard', 'learning', 'study'],
    metadata: { priority: 70, onDesktopByDefault: true },
  },

  // SYSTEM
  {
    id: 'etl',
    name: 'ETL Monitor',
    shortName: 'ETL',
    description: 'ETL engine monitoring dashboard.',
    path: '/etl',
    icon: {
      name: 'activity',
      color: 'oklch(0.68 0.14 75)',
      gradient: 'from-amber-500 to-orange-500',
    },
    category: 'system',
    status: 'active',
    window: { defaultWidth: 1200, defaultHeight: 800, minWidth: 500, minHeight: 400, resizable: true, maximizable: true },
    features: ['Provider Status', 'Performance Metrics'],
    tags: ['etl', 'monitoring', 'metrics'],
    metadata: { priority: 65, inDockByDefault: true },
  },
  {
    id: 'sysmon',
    name: 'System Monitor',
    shortName: 'Sysmon',
    description: 'System performance monitoring.',
    path: '/sysmon',
    icon: {
      name: 'cpu',
      color: 'oklch(0.72 0.14 85)',
      gradient: 'from-amber-500 to-yellow-500',
    },
    category: 'system',
    status: 'planned',
    window: { defaultWidth: 1000, defaultHeight: 700, minWidth: 400, minHeight: 300, resizable: true, maximizable: true },
    features: ['CPU/Memory', 'Disk I/O'],
    tags: ['monitoring', 'performance'],
    metadata: { priority: 40 },
  },

  // UTILITIES
  {
    id: 'bookmarks',
    name: 'Bookmarks',
    shortName: 'Bookmarks',
    description: 'Personal bookmark management.',
    path: '/bookmarks',
    icon: {
      name: 'bookmark',
      color: 'oklch(0.62 0.20 345)',
      gradient: 'from-rose-500 to-pink-500',
    },
    category: 'utilities',
    status: 'active',
    window: { defaultWidth: 900, defaultHeight: 700, minWidth: 350, minHeight: 300, resizable: true, maximizable: true },
    features: ['Save Items', 'Collections'],
    tags: ['bookmark', 'save'],
    metadata: { priority: 60 },
  },
];
```

---

## 7. Window Management

### 7.1 Window States

| State | Description | Visual Indicator |
|--------|-------------|------------------|
| **Normal** | Default window state | Standard titlebar, full shadow |
| **Focused** | Active window receives input | Glow effect, brighter titlebar |
| **Unfocused** | Inactive window | Dimmed opacity (0.95), reduced shadow |
| **Minimized** | Hidden in dock thumbnail | Not visible on desktop |
| **Maximized** | Fills available space | Window at max size, no resize handles |
| **Fullscreen** | Covers menu bar and dock (optional) | Window at 100% viewport |

### 7.2 Window Layering

```
Z-Index hierarchy:
- Menu bar: 2000
- Quick launch: 1500
- Context menu: 1000
- Focused window: 100+
- Each subsequent window: +1
- Dock: 50
- Desktop icons: 10
- Desktop background: 0
```

### 7.3 Window Operations

```typescript
// lib/desktop/window-manager.ts

export class WindowManager {
  // Open a new window for an app
  openWindow(appId: string): string {
    const app = APP_REGISTRY.find(a => a.id === appId);
    if (!app) throw new Error(`App not found: ${appId}`);

    // Calculate position (cascade effect)
    const existingWindows = get().windows;
    const baseX = 100;
    const baseY = 50;
    const offset = existingWindows.length * 30;

    const window: WindowState = {
      id: generateId(),
      appId,
      title: app.name,
      x: baseX + offset,
      y: baseY + offset,
      width: app.window.defaultWidth,
      height: app.window.defaultHeight,
      minimized: false,
      maximized: false,
      focused: true,
      zIndex: 100 + existingWindows.length,
    };

    // Bring to front
    this.focusWindow(window.id);
    return window.id;
  }

  // Close a window
  closeWindow(windowId: string): void {
    // Check if other windows of same app exist
    const window = get().windows.find(w => w.id === windowId);
    if (!window) return;

    const hasOtherWindows = get().windows.some(
      w => w.appId === window.appId && w.id !== windowId
    );

    // Remove from active apps if last window
    if (!hasOtherWindows) {
      get().activeApps = get().activeApps.filter(id => id !== window.appId);
    }
  }

  // Focus a window (bring to front)
  focusWindow(windowId: string): void {
    const maxZ = Math.max(...get().windows.map(w => w.zIndex));
    const newZ = maxZ + 1;

    get().windows = get().windows.map(w => ({
      ...w,
      focused: w.id === windowId,
      zIndex: w.id === windowId ? newZ : w.zIndex,
    }));
  }

  // Minimize window (genie effect to dock)
  minimizeWindow(windowId: string): void {
    const window = get().windows.find(w => w.id === windowId);
    if (!window) return;

    // Create dock thumbnail
    // Play genie animation
    // Set minimized = true
  }

  // Maximize window
  maximizeWindow(windowId: string): void {
    const window = get().windows.find(w => w.id === windowId);
    if (!window || !window.maximizable) return;

    if (window.maximized) {
      // Restore to previous size
      window.x = window.prevX!;
      window.y = window.prevY!;
      window.width = window.prevWidth!;
      window.height = window.prevHeight!;
      window.maximized = false;
    } else {
      // Store current, then maximize
      window.prevX = window.x;
      window.prevY = window.y;
      window.prevWidth = window.width;
      window.prevHeight = window.height;

      const menuBarHeight = 28;
      const dockHeight = 80;
      const margin = 10;

      window.x = margin;
      window.y = menuBarHeight + margin;
      window.width = window.innerWidth - margin * 2;
      window.height = window.innerHeight - menuBarHeight - dockHeight - margin * 2;
      window.maximized = true;
    }
  }

  // Move window
  moveWindow(windowId: string, deltaX: number, deltaY: number): void {
    const window = get().windows.find(w => w.id === windowId);
    if (!window || window.maximized) return;

    window.x += deltaX;
    window.y += deltaY;

    // Snap to edges (optional)
    // Snap to other windows (optional)
  }

  // Resize window
  resizeWindow(
    windowId: string,
    edge: 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw',
    deltaX: number,
    deltaY: number
  ): void {
    const window = get().windows.find(w => w.id === windowId);
    if (!window || !window.resizable || window.maximized) return;

    // Apply resize based on edge
    // Enforce min/max dimensions
  }
}
```

---

## 8. Dock & App Launcher

### 8.1 Dock Behavior

```typescript
// Dock item states
interface DockItem {
  appId: string;
  running: boolean;          // App has open windows
  windowCount: number;        // Number of open windows
  minimizedWindows: string[]; // Minimized window IDs
}

// Dock interactions
interface DockInteractions {
  // Click
  onClick: (appId: string) => void;
  // - If not running: Launch app
  // - If running with no focused windows: Focus most recent
  // - If running with focused windows: Minimize all
  // - If running with minimized windows: Restore window

  // Right-click context menu
  onContextMenu: (appId: string) => ContextMenuItem[];
  // Items: Open, New Window, Show All Windows, Options, Quit

  // Cmd+Click
  onCmdClick: (appId: string) => void;
  // - Open new window even if running

  // Drag
  onDragStart: (appId: string) => void;
  onDragEnd: (appId: string, newPosition: number) => void;
  // - Reorder dock items

  // Hover
  onHover: (appId: string) => void;
  // - Show tooltip
  // - Magnify icon (1.2x scale)
  // - Show window thumbnails (if multiple windows)
}
```

### 8.2 Dock Visual States

```
Normal:    ┌────┐           Hover:     ┌────┐
           │ 🛡️│  56px                 │ 🛡️│  67px (1.2x)
           └────┘                        └────┘

Running:   ┌────┐           Multiple:
           │ 🛡️│  56px        ┌────┐
           └────┐                │ 🛡️│
              ● indicator         └────┐
              4px dot              ●●●  3 indicators
```

---

## 9. Interaction Model

### 9.1 Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + K` | Open quick launch |
| `Cmd/Ctrl + Space` | Open quick launch (alternative) |
| `Cmd + ,` | Open settings |
| `Cmd + H` | Hide current window |
| `Cmd + Option + H` | Hide all other windows |
| `Cmd + M` | Minimize current window |
| `Cmd + W` | Close current window |
| `Cmd + Q` | Quit application |
| `Cmd + Tab` | Cycle through open windows |
| `Cmd + Shift + Tab` | Reverse cycle through windows |
| `F11` | Toggle fullscreen |
| `Esc` | Close modal / exit fullscreen |
| `Cmd + +` | Zoom in |
| `Cmd + -` | Zoom out |
| `Cmd + 0` | Reset zoom |

### 9.2 Mouse Interactions

| Target | Action | Result |
|---------|---------|---------|
| Desktop icon | Click | Select icon |
| Desktop icon | Double-click | Open app window |
| Desktop icon | Drag | Move icon (snap to grid) |
| Desktop icon | Right-click | Context menu |
| Window titlebar | Drag | Move window |
| Window titlebar | Double-click | Maximize/restore |
| Window edge | Drag | Resize window |
| Window content | Click | Focus window |
| Dock item | Click | Launch/focus/minimize |
| Dock item | Right-click | Dock context menu |
| Dock item | Drag out | Remove from dock |
| Desktop | Right-click | Desktop context menu |
| Desktop | Double-click (empty) | Create folder (future) |

### 9.3 Context Menus

**Desktop Icon Context Menu:**
```
┌─────────────────────┐
│  Open              │
│  Show Info         │
│  ────────────────  │
│  Remove from Desktop│
│  Add to Dock       │
└─────────────────────┘
```

**Dock Item Context Menu:**
```
┌─────────────────────┐
│  Open              │
│  New Window        │
│  ────────────────  │
│  Show All Windows  │
│  Options >         │
│  ────────────────  │
│  Remove from Dock   │
│  Quit              │
└─────────────────────┘
```

**Window Context Menu:**
```
┌─────────────────────┐
│  Close             │
│  Minimize          │
│  Maximize          │
│  ────────────────  │
│  Move             │
│  Resize            │
│  ────────────────  │
│  Send to Back      │
│  Bring to Front    │
└─────────────────────┘
```

---

## 10. Component Specification

### 10.1 MenuBar

```tsx
interface MenuBarProps {
  activeApp?: string;
  onOpenQuickLaunch: () => void;
}

// Features:
// - v2e logo (left)
// - Active app name menu
// - Window menu
// - Help menu
// - Search trigger (right, Cmd+K indicator)
// - Theme toggle
// - User profile
// - Glass morphism background
// - Height: 28px
```

### 10.2 DesktopArea

```tsx
interface DesktopAreaProps {
  icons: DesktopIcon[];
  onIconClick: (id: string) => void;
  onIconDoubleClick: (id: string) => void;
  onIconDrag: (id: string, x: number, y: number) => void;
  onContextMenu: (x: number, y: number) => void;
  children: React.ReactNode; // Windows render here
}

// Features:
// - Full viewport area (minus menu bar and dock)
// - Click deselects all icons
// - Double-click on empty space (future: new folder)
// - Right-click context menu
// - Background wallpaper
// - Grid snap overlay (when dragging)
```

### 10.3 DesktopIcon

```tsx
interface DesktopIconProps {
  app: AppRegistryEntry;
  x: number;
  y: number;
  selected: boolean;
  onDrag: (x: number, y: number) => void;
  onDoubleClick: () => void;
  onContextMenu: () => void;
}

// Features:
// - App icon with gradient background
// - App label below icon
// - Selection highlight (rounded rect)
// - Drag with snap to grid
// - Hover effect (subtle glow)
// - Double-click to open
```

### 10.4 AppWindow

```tsx
interface AppWindowProps {
  window: WindowState;
  app: AppRegistryEntry;
  focused: boolean;
  onFocus: () => void;
  onClose: () => void;
  onMinimize: () => void;
  onMaximize: () => void;
  onMove: (x: number, y: number) => void;
  onResize: (width: number, height: number) => void;
}

// Features:
// - Title bar with controls
// - Close (red), Minimize (yellow), Maximize (green)
// - Drag to move
// - Resize handles (8 directions)
// - Focus glow when active
// - Dimmed when inactive
// - Content iframe or direct render
```

### 10.5 Dock

```tsx
interface DockProps {
  items: string[];
  activeApps: string[];
  minimizedWindows: Record<string, string[]>;
  onItemClick: (appId: string) => void;
  onItemContextMenu: (appId: string) => void;
  onItemReorder: (from: number, to: number) => void;
}

// Features:
// - Glass morphism background
// - Rounded corners
// - Icon magnification on hover (1.2x)
// - Active indicator dots
// - Minimized window thumbnails
// - Drag to reorder
// - Auto-hide (optional)
// - Position: bottom
```

### 10.6 QuickLaunch

```tsx
interface QuickLaunchProps {
  isOpen: boolean;
  onClose: () => void;
  onAppSelect: (appId: string) => void;
}

// Features:
// - Centered modal
// - Search input with icon
// - Filtered app list
// - Keyboard navigation
// - Arrow key selection
// - Enter to open
// - Esc to close
// - Recent apps section
```

---

## Appendices

### Appendix A: CSS Variables Reference

```css
:root {
  /* Layout Dimensions */
  --menubar-height: 28px;
  --dock-height: 80px;
  --dock-height-collapsed: 4px;
  --icon-size: 80px;
  --dock-item-size: 56px;

  /* Window */
  --window-min-width: 400px;
  --window-min-height: 300px;
  --titlebar-height: 36px;
  --window-border-radius: 10px;
  --window-border-radius-large: 12px;

  /* Desktop Grid */
  --icon-grid-size: 100px;
  --icon-spacing: 20px;

  /* Animation */
  --transition-window-open: 200ms ease-out;
  --transition-window-close: 150ms ease-in;
  --transition-window-minimize: 300ms ease-in;
  --transition-dock-hover: 150ms ease-out;
  --transition-icon-select: 150ms ease-out;

  /* Z-Index */
  --z-desktop: 0;
  --z-desktop-icons: 10;
  --z-dock: 50;
  --z-window-base: 100;
  --z-context-menu: 1000;
  --z-quick-launch: 1500;
  --z-menubar: 2000;
}
```

### Appendix B: Browser Support

| Browser | Minimum Version | Notes |
|---------|----------------|--------|
| Chrome | 120+ | Full support |
| Firefox | 121+ | Full support |
| Safari | 17+ | Full support |
| Edge | 120+ | Full support |

**Required Features:**
- CSS Grid
- CSS Custom Properties
- Backdrop Filter
- CSS OKLCH colors
- ES2022 JavaScript
- Intersection Observer
- Resize Observer
- Drag and Drop API

### Appendix C: Accessibility

### Desktop Navigation

| Element | Keyboard Access | Screen Reader |
|---------|----------------|---------------|
| Desktop icons | Tab + arrows | "CVE icon, double-click to open" |
| Windows | Tab + Cmd+Tab | Announces window title and app |
| Dock items | Opt+Tab + arrows | "Dock, CVE Browser, running" |
| Menu bar | Ctrl+F2 | Menu items announced |
| Quick launch | Cmd+K, arrows, Enter | Search results announced |

### Focus Management

- Desktop icons: Tab through grid, arrows navigate
- Windows: Cmd+Tab cycles through windows
- Dock: Opt+Tab cycles through dock items
- Menu bar: Ctrl+F2 enters menu bar
- Quick launch: Esc closes, returns to desktop

---

**Document Status:** Ready for Implementation
**Platform:** Desktop Only (1024px minimum)
**Last Updated:** 2026-02-12
