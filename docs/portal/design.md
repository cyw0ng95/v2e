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
┌─────────────────────────────────────────────────────────────────────┐
│  Menu Bar (28px)                                                │
│  ┌────────┐  CVE Database  ▼  Window    Help                     │
│  │ v2e    │  Search... [Cmd+K]      🌙  👤                   │
│  └────────┘                                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Desktop Area (100vh - 28px - 80px = calc(100vh - 108px))    │
│                                                                  │
│  ┌─────┐  ┌─────┐        ┌──────────────────────┐              │
│  │ 🛡️CVE│  │ 🐛CWE│        │  CVSS Calculator    │              │
│  └─────┘  └─────┘        │  ┌──────────────┐   │              │
│                             │  │ Base Score  │   │              │
│  ┌─────┐  ┌─────┐        │  │ 7.5 (High) │   │              │
│  │ 🎯CAPEC│  │ 📍ATT&CK      │  └──────────────┘   │              │
│  └─────┘  └─────┘        │  Vector String       │              │
│                             │  CVSS:3.1/AV:N/... │              │
│  ┌─────┐  ┌─────┐        └──────────────────────┘              │
│  │ 🧮CVSS│  │ 📊GLC│                                              │
│  └─────┘  └─────┘                                               │
│                                                                  │
│  ┌─────┐  ┌─────┐                                                │
│  │ 📚Mcards│  │ 🔖Bookmarks                                        │
│  └─────┘  └─────┘                                                │
│                                                                  │
├─────────────────────────────────────────────────────────────────────┤
│  Dock (80px, auto-hides when inactive)                             │
│  ┌────┐┌────┐┌────┐┌────┐┌────┐┌────┐┌────┐┌────┐    │
│  │ 🛡️ ││ 🐛 ││ 🧮 ││ 📊 ││ 📚 ││ 🔄 ││ ⚙️ ││ 🔖 │    │
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
```

#### Window Styling

```css
/* Window frame */
--window-bg: oklch(1 0 0 / 95%);
--window-bg-dark: oklch(0.15 0.008 264 / 95%);
--window-border: oklch(0 0 0 / 10%);
--window-shadow: 0 10px 40px oklch(0 0 0 / 15%);
--window-glow: 0 0 0 1px oklch(0.52 0.16 265 / 30%);
```

#### Dock Styling

```css
/* Dock container */
--dock-bg: oklch(1 0 0 / 75%);
--dock-bg-dark: oklch(0.15 0.008 264 / 75%);
--dock-shadow: 0 -4px 20px oklch(0 0 0 / 10%);
--dock-radius: 20px;

/* Dock item */
--dock-item-size: 56px;
--dock-indicator: oklch(0.52 0.16 265);
```

### 3.2 Typography

```typescript
const desktopTypography = {
  menuBar: { fontSize: '13px', fontWeight: 500 },
  windowTitle: { fontSize: '13px', fontWeight: 500 },
  iconLabel: { fontSize: '11px', fontWeight: 500 },
  dockTooltip: { fontSize: '12px', fontWeight: 500 },
};
```

### 3.3 Spacing & Layout

```typescript
const desktopLayout = {
  menuBarHeight: 28,
  dockHeight: 80,
  iconSize: 80,
  windowMinWidth: 400,
  windowMinHeight: 300,
  windowDefaultWidth: 1024,
  windowDefaultHeight: 768,
};
```

### 3.4 Effects & Animations

```css
/* Window open animation */
@keyframes window-open {
  0% { opacity: 0; transform: scale(0.95); }
  100% { opacity: 1; transform: scale(1); }
}

/* Window close animation */
@keyframes window-close {
  0% { opacity: 1; transform: scale(1); }
  100% { opacity: 0; transform: scale(0.95); }
}

/* Dock magnification */
.dock-item:hover {
  transform: scale(1.2) translateY(-8px);
}
```

---

## 4. Desktop Architecture

### 4.1 File Structure

```
website/
├── app/desktop/
│   ├── page.tsx                 # Desktop (new landing page)
│   └── layout.tsx
├── components/desktop/
│   ├── menu-bar.tsx
│   ├── desktop-area.tsx
��   ├── desktop-icon.tsx
│   ├── dock.tsx
│   ├── window/
│   │   ├── app-window.tsx
│   │   ├── window-titlebar.tsx
│   │   ├── window-controls.tsx
│   │   └── window-resize.tsx
│   └── context-menus.tsx
├── lib/desktop/
│   ├── desktop-state.ts
│   ├── window-manager.ts
│   └── dock-manager.ts
└── types/desktop.ts
```

### 4.2 State Architecture

```typescript
// Desktop icon position
interface DesktopIcon {
  id: string;
  x: number;
  y: number;
  gridX: number;
  gridY: number;
}

// Window state
interface WindowState {
  id: string;
  appId: string;
  title: string;
  x: number;
  y: number;
  width: number;
  height: number;
  minimized: boolean;
  maximized: boolean;
  focused: boolean;
  zIndex: number;
}

// Global desktop state
interface DesktopState {
  config: DesktopConfig;
  windows: WindowState[];
  activeWindowId: string | null;
  dockConfig: DockConfig;
  activeApps: string[];
}
```

---

## 5. User Interface Design

### 5.1 Desktop Area

Users land here first - a macOS-style desktop with app icons.

**Interactions:**
- **Click**: Select icon
- **Double-click**: Open app window
- **Drag**: Move icon (snap to grid)
- **Right-click**: Context menu

### 5.2 Menu Bar

```
┌─────────────────────────────────────────────────────────────────────┐
│  v2e  |  CVE Database  ▼  |  Window  |  Help                 │
│                                                    🔍 [Cmd+K]  🌙  👤  │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.3 Dock

```
Height: 80px
Background: Glass morphism
Icons: Scale up on hover (1.2x)
```

### 5.4 Window Design

```
┌─────────────────────────────────────────────────────────────────────┐
│  CVE Browser                              ─ □ ✕                     │
├─────────────────────────────────────────────────────────────────────┤
│  (App content - iframe or component)                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 6. Application Registry

See plan.md for implementation details.

---

## 7. Window Management

### 7.1 Window States

| State | Description |
|--------|-------------|
| Normal | Default window state |
| Focused | Active window receives input |
| Unfocused | Inactive window |
| Minimized | Hidden in dock thumbnail |
| Maximized | Fills available space |

### 7.2 Window Layering

```
Z-Index hierarchy:
- Menu bar: 2000
- Quick launch: 1500
- Context menu: 1000
- Focused window: 100+
- Dock: 50
- Desktop icons: 10
```

---

## 8. Dock & App Launcher

### 8.1 Dock Behavior

- **Click**: Launch/focus/minimize
- **Right-click**: Context menu
- **Cmd+Click**: Open new window
- **Hover**: Magnify (1.2x)

---

## 9. Interaction Model

### 9.1 Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Cmd/Ctrl + K | Open quick launch |
| Cmd + W | Close window |
| Cmd + M | Minimize window |
| Cmd + Tab | Cycle windows |

---

## 10. Component Specification

Components defined in architecture section.

---

## Appendices

### Appendix A: CSS Variables

```css
:root {
  --menubar-height: 28px;
  --dock-height: 80px;
  --window-min-width: 400px;
  --window-min-height: 300px;
}
```

### Appendix B: Browser Support

Chrome 120+, Firefox 121+, Safari 17+, Edge 120+

### Appendix C: Accessibility

Full keyboard navigation and screen reader support.

---

**Document Status:** Ready for Implementation
**Last Updated:** 2026-02-12
