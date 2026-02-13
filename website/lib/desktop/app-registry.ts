/**
 * v2e Portal - Application Registry
 *
 * Central registry of all available apps
 * Phase 3: App Registry Implementation
 * Backend Independence: Works completely offline
 */

import type { AppCategory } from '@/types/desktop';

/**
 * Application metadata interface
 * Matches APP_REGISTRY_ENTRY structure
 */
export interface AppRegistryEntry {
  id: string; // Unique app identifier
  name: string; // Display name
  path: string; // Route path (for iframe loading)
  icon: string; // Lucide React icon name
  category: AppCategory;

  // Window defaults
  defaultWidth: number;
  defaultHeight: number;
  minWidth: number;
  minHeight: number;
  maxWidth?: number;
  maxHeight?: number;

  // Desktop icon
  iconColor?: string; // Accent color for icon background
  defaultPosition?: { x: number; y: number };

  // Behavior flags
  allowMultipleWindows?: boolean;
  resizable?: boolean;
  maximizable?: boolean;
  minimizable?: boolean;

  // Integration
  contentMode: 'iframe' | 'component';
  status: 'active' | 'planned' | 'deprecated';

  // Network requirements
  requiresOnline?: boolean;
}

/**
 * Active applications (9 apps)
 */
export const ACTIVE_APPS: AppRegistryEntry[] = [
  {
    id: 'cve',
    name: 'CVE Browser',
    path: '/cve',
    icon: 'shield-alert',
    category: 'Database',
    defaultWidth: 1200,
    defaultHeight: 800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'cwe',
    name: 'CWE Database',
    path: '/cwe',
    icon: 'bug',
    category: 'Database',
    defaultWidth: 1200,
    defaultHeight:  800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'capec',
    name: 'CAPEC Encyclopedia',
    path: '/capec',
    icon: 'target',
    category: 'Database',
    defaultWidth: 1200,
    defaultHeight: 800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'attack',
    name: 'ATT&CK Explorer',
    path: '/attack',
    icon: 'crosshair',
    category: 'Database',
    defaultWidth: 1400,
    defaultHeight: 900,
    minWidth: 900,
    minHeight: 600,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'cvss',
    name: 'CVSS Calculator',
    path: '/cvss',
    icon: 'calculator',
    category: 'Tool',
    defaultWidth: 900,
    defaultHeight: 700,
    minWidth: 600,
    minHeight: 500,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'glc',
    name: 'Graphized Learning Canvas',
    path: '/glc',
    icon: 'git-graph',
    category: 'Learning',
    defaultWidth: 1400,
    defaultHeight: 900,
    minWidth: 900,
    minHeight: 600,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'mcards',
    name: 'Mcards',
    path: '/mcards',
    icon: 'library',
    category: 'Learning',
    defaultWidth: 1200,
    defaultHeight: 800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'component',
    status: 'active',
  },
  {
    id: 'etl',
    name: 'ETL Monitor',
    path: '/etl',
    icon: 'activity',
    category: 'System',
    defaultWidth: 1000,
    defaultHeight: 700,
    minWidth: 700,
    minHeight: 500,
    contentMode: 'iframe',
    status: 'active',
  },
  {
    id: 'bookmarks',
    name: 'Bookmarks',
    path: '/bookmarks',
    icon: 'bookmark',
    category: 'Utility',
    defaultWidth: 900,
    defaultHeight: 700,
    minWidth: 600,
    minHeight: 500,
    contentMode: 'iframe',
    status: 'active',
  },
];

/**
 * Planned applications (4 apps) - Phase 4 only
 */
export const PLANNED_APPS: AppRegistryEntry[] = [
  {
    id: 'sysmon',
    name: 'System Monitor',
    path: '/sysmon',
    icon: 'gauge',
    category: 'System',
    defaultWidth: 1000,
    defaultHeight: 700,
    minWidth: 700,
    minHeight: 500,
    contentMode: 'iframe',
    status: 'planned',
  },
  {
    id: 'cce',
    name: 'CCE Database',
    path: '/cce',
    icon: 'file-check',
    category: 'Reference',
    defaultWidth: 1200,
    defaultHeight: 800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'iframe',
    status: 'planned',
  },
  {
    id: 'ssg',
    name: 'SSG Guides',
    path: '/ssg',
    icon: 'scroll-text',
    category: 'Reference',
    defaultWidth: 1200,
    defaultHeight: 800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'iframe',
    status: 'planned',
  },
  {
    id: 'asvs',
    name: 'ASVS',
    path: '/asvs',
    icon: 'check-circle',
    category: 'Reference',
    defaultWidth: 1200,
    defaultHeight: 800,
    minWidth: 800,
    minHeight: 600,
    contentMode: 'iframe',
    status: 'planned',
  },
];

/**
 * Get app by ID
 */
export function getAppById(id: string): AppRegistryEntry | undefined {
  return [...ACTIVE_APPS, ...PLANNED_APPS].find(app => app.id === id);
}

/**
 * Get apps by category
 */
export function getAppsByCategory(category: AppCategory): AppRegistryEntry[] {
  return [...ACTIVE_APPS, ...PLANNED_APPS].filter(app => app.category === category);
}

/**
 * Get active apps only
 */
export function getActiveApps(): AppRegistryEntry[] {
  return ACTIVE_APPS.filter(app => app.status === 'active');
}

/**
 * Get all apps (active + planned)
 */
export function getAllApps(): AppRegistryEntry[] {
  return [...ACTIVE_APPS, ...PLANNED_APPS];
}
