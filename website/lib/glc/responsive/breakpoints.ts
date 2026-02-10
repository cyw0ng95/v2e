/**
 * GLC Responsive Breakpoints
 *
 * Breakpoint definitions following Tailwind CSS conventions.
 * - xs: 0-639px (mobile)
 * - sm: 640-767px (large mobile)
 * - md: 768-1023px (tablet)
 * - lg: 1024-1279px (desktop)
 * - xl: 1280px+ (large desktop)
 */

export const breakpoints = {
  xs: 0,
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
} as const;

export type Breakpoint = keyof typeof breakpoints;

/**
 * Get the minimum width for a breakpoint
 */
export function getBreakpointMin(breakpoint: Breakpoint): number {
  return breakpoints[breakpoint];
}

/**
 * Get the maximum width for a breakpoint (exclusive)
 */
export function getBreakpointMax(breakpoint: Breakpoint): number | null {
  const keys = Object.keys(breakpoints) as Breakpoint[];
  const index = keys.indexOf(breakpoint);
  if (index === -1 || index === keys.length - 1) {
    return null;
  }
  return breakpoints[keys[index + 1]] - 1;
}

/**
 * Generate a media query string for a breakpoint
 */
export function getMediaQuery(breakpoint: Breakpoint, direction: 'min' | 'max' = 'min'): string {
  if (direction === 'min') {
    return `(min-width: ${breakpoints[breakpoint]}px)`;
  }
  const max = getBreakpointMax(breakpoint);
  if (max === null) {
    return ''; // No max for the largest breakpoint
  }
  return `(max-width: ${max}px)`;
}

/**
 * Generate a media query for a breakpoint range
 */
export function getMediaQueryRange(min: Breakpoint, max: Breakpoint): string {
  const minQuery = getMediaQuery(min, 'min');
  const maxQuery = getMediaQuery(max, 'max');
  if (!maxQuery) return minQuery;
  return `${minQuery} and ${maxQuery}`;
}

/**
 * Check if current width is at or above a breakpoint
 */
export function isAtOrAbove(width: number, breakpoint: Breakpoint): boolean {
  return width >= breakpoints[breakpoint];
}

/**
 * Check if current width is below a breakpoint
 */
export function isBelow(width: number, breakpoint: Breakpoint): boolean {
  return width < breakpoints[breakpoint];
}

/**
 * Get the current breakpoint based on width
 */
export function getCurrentBreakpoint(width: number): Breakpoint {
  if (width >= breakpoints.xl) return 'xl';
  if (width >= breakpoints.lg) return 'lg';
  if (width >= breakpoints.md) return 'md';
  if (width >= breakpoints.sm) return 'sm';
  return 'xs';
}

/**
 * Device type classification
 */
export type DeviceType = 'mobile' | 'tablet' | 'desktop';

/**
 * Get device type based on width
 */
export function getDeviceType(width: number): DeviceType {
  if (width < breakpoints.md) return 'mobile';
  if (width < breakpoints.lg) return 'tablet';
  return 'desktop';
}

/**
 * Touch target size for accessibility (44px minimum)
 */
export const TOUCH_TARGET_SIZE = 44;

/**
 * Spacing constants for responsive layouts
 */
export const responsiveSpacing = {
  xs: { padding: 8, gap: 8 },
  sm: { padding: 12, gap: 12 },
  md: { padding: 16, gap: 16 },
  lg: { padding: 20, gap: 20 },
  xl: { padding: 24, gap: 24 },
} as const;

/**
 * Palette width constants for different breakpoints
 */
export const paletteWidth = {
  collapsed: 48,
  expanded: {
    xs: 0, // Hidden on mobile (uses drawer)
    sm: 240,
    md: 256,
    lg: 280,
    xl: 300,
  },
} as const;

/**
 * Toolbar configuration for different breakpoints
 */
export const toolbarConfig = {
  xs: {
    showLabels: false,
    showSeparators: false,
    buttonSize: 44, // Touch target size
    compact: true,
  },
  sm: {
    showLabels: false,
    showSeparators: true,
    buttonSize: 44,
    compact: true,
  },
  md: {
    showLabels: false,
    showSeparators: true,
    buttonSize: 36,
    compact: false,
  },
  lg: {
    showLabels: true,
    showSeparators: true,
    buttonSize: 32,
    compact: false,
  },
  xl: {
    showLabels: true,
    showSeparators: true,
    buttonSize: 32,
    compact: false,
  },
} as const;
