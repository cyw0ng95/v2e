/**
 * GLC Responsive Utilities
 *
 * Re-export all responsive utilities for convenience.
 */

export {
  breakpoints,
  getBreakpointMin,
  getBreakpointMax,
  getMediaQuery,
  getMediaQueryRange,
  isAtOrAbove,
  isBelow,
  getCurrentBreakpoint,
  getDeviceType,
  TOUCH_TARGET_SIZE,
  responsiveSpacing,
  paletteWidth,
  toolbarConfig,
  type Breakpoint,
  type DeviceType,
} from './breakpoints';

export {
  useMediaQuery,
  useViewportWidth,
  useBreakpoint,
  useIsAtOrAbove,
  useIsBelow,
  useDeviceType,
  useIsMobile,
  useIsTablet,
  useIsDesktop,
  useIsTouchDevice,
  useResponsive,
  useResponsiveValue,
  useBreakpointRender,
  useResponsiveDimensions,
  type ResponsiveState,
} from './use-media-query';
