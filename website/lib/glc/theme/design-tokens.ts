/**
 * GLC Theme System
 *
 * Design tokens and theme utilities
 */

import type { CanvasPresetTheme } from '../types';

// Light theme colors
export const lightTheme: CanvasPresetTheme = {
  background: '#ffffff',
  surface: '#f8fafc',
  text: '#1e293b',
  textMuted: '#64748b',
  primary: '#6366f1',
  accent: '#8b5cf6',
  error: '#ef4444',
  warning: '#f59e0b',
  success: '#22c55e',
  border: '#e2e8f0',
};

// Dark theme colors
export const darkTheme: CanvasPresetTheme = {
  background: '#0f172a',
  surface: '#1e293b',
  text: '#f1f5f9',
  textMuted: '#94a3b8',
  primary: '#818cf8',
  accent: '#a78bfa',
  error: '#f87171',
  warning: '#fbbf24',
  success: '#4ade80',
  border: '#334155',
};

// High contrast theme
export const highContrastTheme: CanvasPresetTheme = {
  background: '#000000',
  surface: '#1a1a1a',
  text: '#ffffff',
  textMuted: '#cccccc',
  primary: '#00ff00',
  accent: '#00ffff',
  error: '#ff0000',
  warning: '#ffff00',
  success: '#00ff00',
  border: '#ffffff',
};

/**
 * Get theme by name
 */
export function getTheme(name: 'light' | 'dark' | 'high-contrast'): CanvasPresetTheme {
  switch (name) {
    case 'light':
      return lightTheme;
    case 'dark':
      return darkTheme;
    case 'high-contrast':
      return highContrastTheme;
    default:
      return darkTheme;
  }
}

/**
 * Detect system color scheme preference
 */
export function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'dark';

  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
}

/**
 * Calculate relative luminance
 */
export function getLuminance(hex: string): number {
  const rgb = hexToRgb(hex);
  if (!rgb) return 0;

  const [r, g, b] = [rgb.r, rgb.g, rgb.b].map((c) => {
    c /= 255;
    return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
  });

  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/**
 * Calculate contrast ratio between two colors
 */
export function getContrastRatio(color1: string, color2: string): number {
  const l1 = getLuminance(color1);
  const l2 = getLuminance(color2);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Check if contrast meets WCAG AA (4.5:1)
 */
export function meetsWCAGAA(foreground: string, background: string): boolean {
  return getContrastRatio(foreground, background) >= 4.5;
}

/**
 * Hex to RGB conversion
 */
function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null;
}

/**
 * Lighten a color
 */
export function lighten(hex: string, percent: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;

  const amount = Math.round(2.55 * percent);
  const r = Math.min(255, rgb.r + amount);
  const g = Math.min(255, rgb.g + amount);
  const b = Math.min(255, rgb.b + amount);

  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
}

/**
 * Darken a color
 */
export function darken(hex: string, percent: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;

  const amount = Math.round(2.55 * percent);
  const r = Math.max(0, rgb.r - amount);
  const g = Math.max(0, rgb.g - amount);
  const b = Math.max(0, rgb.b - amount);

  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
}

/**
 * Add alpha to hex color
 */
export function withAlpha(hex: string, alpha: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;

  return `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${alpha})`;
}
