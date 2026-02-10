'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  breakpoints,
  type Breakpoint,
  type DeviceType,
  getDeviceType,
  getCurrentBreakpoint,
  isAtOrAbove,
  isBelow,
} from './breakpoints';

/**
 * Hook to listen to a media query
 */
export function useMediaQuery(query: string): boolean {
  const getMatches = useCallback((q: string): boolean => {
    // Prevents SSR issues
    if (typeof window === 'undefined') return false;
    return window.matchMedia(q).matches;
  }, []);

  const [matches, setMatches] = useState(() => getMatches(query));

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const mediaQuery = window.matchMedia(query);

    const handler = (event: MediaQueryListEvent) => {
      setMatches(event.matches);
    };

    // Check current value on mount and subscribe
    const currentMatches = mediaQuery.matches;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMatches(currentMatches);

    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, [query]);

  return matches;
}

/**
 * Hook to get current viewport width
 */
export function useViewportWidth(): number {
  const [width, setWidth] = useState(0);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleResize = () => {
      setWidth(window.innerWidth);
    };

    handleResize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return width;
}

/**
 * Hook to get current breakpoint
 */
export function useBreakpoint(): Breakpoint {
  const width = useViewportWidth();
  return getCurrentBreakpoint(width);
}

/**
 * Hook to check if viewport is at or above a breakpoint
 */
export function useIsAtOrAbove(breakpoint: Breakpoint): boolean {
  const width = useViewportWidth();
  return isAtOrAbove(width, breakpoint);
}

/**
 * Hook to check if viewport is below a breakpoint
 */
export function useIsBelow(breakpoint: Breakpoint): boolean {
  const width = useViewportWidth();
  return isBelow(width, breakpoint);
}

/**
 * Hook to get device type (mobile/tablet/desktop)
 */
export function useDeviceType(): DeviceType {
  const width = useViewportWidth();
  return getDeviceType(width);
}

/**
 * Hook to check if device is mobile (< md breakpoint)
 */
export function useIsMobile(): boolean {
  return useIsBelow('md');
}

/**
 * Hook to check if device is tablet (md to lg)
 */
export function useIsTablet(): boolean {
  const width = useViewportWidth();
  return width >= breakpoints.md && width < breakpoints.lg;
}

/**
 * Hook to check if device is desktop (>= lg)
 */
export function useIsDesktop(): boolean {
  return useIsAtOrAbove('lg');
}

/**
 * Hook to check if device supports touch
 */
export function useIsTouchDevice(): boolean {
  const [isTouch, setIsTouch] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const checkTouch = () => {
      setIsTouch(
        'ontouchstart' in window ||
        navigator.maxTouchPoints > 0
      );
    };

    checkTouch();
  }, []);

  return isTouch;
}

/**
 * Comprehensive responsive state hook
 */
export interface ResponsiveState {
  width: number;
  breakpoint: Breakpoint;
  deviceType: DeviceType;
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isTouch: boolean;
  isXs: boolean;
  isSm: boolean;
  isMd: boolean;
  isLg: boolean;
  isXl: boolean;
}

export function useResponsive(): ResponsiveState {
  const width = useViewportWidth();
  const breakpoint = getCurrentBreakpoint(width);
  const deviceType = getDeviceType(width);
  const isTouch = useIsTouchDevice();

  return {
    width,
    breakpoint,
    deviceType,
    isMobile: deviceType === 'mobile',
    isTablet: deviceType === 'tablet',
    isDesktop: deviceType === 'desktop',
    isTouch,
    isXs: breakpoint === 'xs',
    isSm: breakpoint === 'sm',
    isMd: breakpoint === 'md',
    isLg: breakpoint === 'lg',
    isXl: breakpoint === 'xl',
  };
}

/**
 * Hook to get value based on current breakpoint
 */
export function useResponsiveValue<T>(values: Partial<Record<Breakpoint, T>>, defaultValue: T): T {
  const breakpoint = useBreakpoint();

  // Try to get value for current breakpoint, falling back to smaller breakpoints
  const order: Breakpoint[] = ['xs', 'sm', 'md', 'lg', 'xl'];
  const currentIndex = order.indexOf(breakpoint);

  for (let i = currentIndex; i >= 0; i--) {
    if (values[order[i]] !== undefined) {
      return values[order[i]] as T;
    }
  }

  return defaultValue;
}

/**
 * Hook to conditionally render based on breakpoint
 */
export function useBreakpointRender(config: {
  xs?: boolean;
  sm?: boolean;
  md?: boolean;
  lg?: boolean;
  xl?: boolean;
}): boolean {
  const breakpoint = useBreakpoint();
  return config[breakpoint] ?? false;
}

/**
 * Hook for responsive dimension calculations
 */
export function useResponsiveDimensions() {
  const responsive = useResponsive();

  const getTouchTargetSize = useCallback(() => {
    return responsive.isMobile ? 44 : 32;
  }, [responsive.isMobile]);

  const getPaletteWidth = useCallback((isOpen: boolean) => {
    if (!isOpen) return 48;
    if (responsive.isMobile) return 0; // Uses drawer
    if (responsive.isTablet) return 256;
    if (responsive.breakpoint === 'lg') return 280;
    return 300; // xl
  }, [responsive]);

  const getToolbarConfig = useCallback(() => {
    return {
      showLabels: responsive.isDesktop,
      showSeparators: !responsive.isXs,
      buttonSize: responsive.isMobile ? 44 : responsive.isTablet ? 36 : 32,
      compact: responsive.isMobile,
    };
  }, [responsive]);

  const shouldShowMinimap = useCallback(() => {
    return responsive.isDesktop;
  }, [responsive.isDesktop]);

  const shouldUseDrawerPalette = useCallback(() => {
    return responsive.isMobile;
  }, [responsive.isMobile]);

  return {
    responsive,
    getTouchTargetSize,
    getPaletteWidth,
    getToolbarConfig,
    shouldShowMinimap,
    shouldUseDrawerPalette,
  };
}
