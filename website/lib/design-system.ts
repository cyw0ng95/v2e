// Design System Tokens for v2e Website
// Centralized design tokens following 8pt grid system and modern design principles

// Spacing Scale (0-12) following 8pt grid
export const spacing = {
  0: '0',
  1: '0.25rem',    // 4px
  2: '0.5rem',     // 8px
  3: '0.75rem',    // 12px
  4: '1rem',       // 16px
  5: '1.25rem',    // 20px
  6: '1.5rem',     // 24px
  8: '2rem',       // 32px
  10: '2.5rem',    // 40px
  12: '3rem',      // 48px
} as const;

// Typography System
export const typography = {
  fontFamily: {
    sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
    mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'Monaco', 'monospace'],
  },
  fontSize: {
    // Display sizes
    display: {
      xs: { fontSize: '2.5rem', lineHeight: '1.2', fontWeight: 700, letterSpacing: '-0.02em' },
      sm: { fontSize: '3rem', lineHeight: '1.2', fontWeight: 700, letterSpacing: '-0.02em' },
      md: { fontSize: '3.75rem', lineHeight: '1.2', fontWeight: 700, letterSpacing: '-0.02em' },
      lg: { fontSize: '4.5rem', lineHeight: '1.1', fontWeight: 700, letterSpacing: '-0.02em' },
      xl: { fontSize: '6rem', lineHeight: '1', fontWeight: 700, letterSpacing: '-0.03em' },
    },
    // Headings
    heading: {
      1: { fontSize: '2.25rem', lineHeight: '1.25', fontWeight: 700, letterSpacing: '-0.01em' },
      2: { fontSize: '1.875rem', lineHeight: '1.3', fontWeight: 600, letterSpacing: '-0.01em' },
      3: { fontSize: '1.5rem', lineHeight: '1.4', fontWeight: 600, letterSpacing: '-0.005em' },
      4: { fontSize: '1.25rem', lineHeight: '1.5', fontWeight: 600 },
      5: { fontSize: '1.125rem', lineHeight: '1.5', fontWeight: 600 },
      6: { fontSize: '1rem', lineHeight: '1.5', fontWeight: 600 },
    },
    // Body text
    body: {
      xs: { fontSize: '0.75rem', lineHeight: '1.4', fontWeight: 400 },
      sm: { fontSize: '0.875rem', lineHeight: '1.5', fontWeight: 400 },
      base: { fontSize: '1rem', lineHeight: '1.6', fontWeight: 400 },
      lg: { fontSize: '1.125rem', lineHeight: '1.7', fontWeight: 400 },
      xl: { fontSize: '1.25rem', lineHeight: '1.75', fontWeight: 400 },
      '2xl': { fontSize: '1.5rem', lineHeight: '1.8', fontWeight: 400 },
    },
    // Captions
    caption: {
      xs: { fontSize: '0.625rem', lineHeight: '1.3', fontWeight: 500 },
      sm: { fontSize: '0.75rem', lineHeight: '1.4', fontWeight: 500 },
      base: { fontSize: '0.875rem', lineHeight: '1.4', fontWeight: 500 },
    },
  },
  fontWeight: {
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },
  letterSpacing: {
    tighter: '-0.05em',
    tight: '-0.025em',
    normal: '0em',
    wide: '0.025em',
    wider: '0.05em',
    widest: '0.1em',
  },
  lineHeight: {
    none: 1,
    tight: 1.25,
    snug: 1.375,
    normal: 1.5,
    relaxed: 1.625,
    loose: 2,
  },
} as const;

// Border Radius Scale
export const radius = {
  sm: '0.375rem',    // 6px
  md: '0.5rem',      // 8px
  lg: '0.625rem',    // 10px
  xl: '0.75rem',     // 12px
  '2xl': '1rem',     // 16px
  '3xl': '1.5rem',   // 24px
} as const;

// Shadow System
export const shadows = {
  xs: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
  sm: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
  '2xl': '0 25px 50px -12px rgb(0 0 0 / 0.25)',
} as const;

// Transition Durations
export const transitions = {
  fast: '150ms ease-out',
  normal: '300ms ease-out',
  slow: '500ms ease-out',
} as const;

// Z-Index Scale
export const zIndex = {
  auto: 'auto',
  0: '0',
  10: '10',
  20: '20',
  30: '30',
  40: '40',
  50: '50',
  dropdown: '1000',
  sticky: '1020',
  fixed: '1030',
  modalBackdrop: '1040',
  modal: '1050',
  popover: '1060',
  tooltip: '1070',
} as const;

// Utility Functions
export const getSpacing = (key: keyof typeof spacing) => spacing[key];
export const getTypography = (category: keyof typeof typography.fontSize, size: string) => {
  const categoryObj = typography.fontSize[category];
  return categoryObj?.[size as keyof typeof categoryObj] || typography.fontSize.body.base;
};
export const getRadius = (key: keyof typeof radius) => radius[key];
export const getShadow = (key: keyof typeof shadows) => shadows[key];
export const getTransition = (key: keyof typeof transitions) => transitions[key];
export const getZIndex = (key: keyof typeof zIndex) => zIndex[key];

// Color System
export const colors = {
  // Primary - Modern indigo/violet gradient
  primary: {
    50: '#eef2ff',
    100: '#e0e7ff',
    200: '#c7d2fe',
    300: '#a5b4fc',
    400: '#818cf8',
    500: '#6366f1', // Base
    600: '#4f46e5',
    700: '#4338ca',
    800: '#3730a3',
    900: '#312e81',
  },
  // Secondary - Neutral with subtle warmth
  secondary: {
    50: '#fafafa',
    100: '#f4f4f5',
    200: '#e4e4e7',
    300: '#d4d4d8',
    400: '#a1a1aa',
    500: '#71717a', // Base
    600: '#52525b',
    700: '#3f3f46',
    800: '#27272a',
    900: '#18181b',
  },
  // Success - Modern green with blue undertone
  success: {
    50: '#f0fdf4',
    100: '#dcfce7',
    200: '#bbf7d0',
    300: '#86efac',
    400: '#4ade80',
    500: '#22c55e', // Base
    600: '#16a34a',
    700: '#15803d',
    800: '#166534',
    900: '#14532d',
  },
  // Warning - Modern amber/orange
  warning: {
    50: '#fffbeb',
    100: '#fef3c7',
    200: '#fde68a',
    300: '#fcd34d',
    400: '#fbbf24',
    500: '#f59e0b', // Base
    600: '#d97706',
    700: '#b45309',
    800: '#92400e',
    900: '#78350f',
  },
  // Error - Modern red with orange undertone
  error: {
    50: '#fef2f2',
    100: '#fee2e2',
    200: '#fecaca',
    300: '#fca5a5',
    400: '#f87171',
    500: '#ef4444', // Base
    600: '#dc2626',
    700: '#b91c1c',
    800: '#991b1b',
    900: '#7f1d1d',
  },
  // Semantic colors for specific use cases
  semantic: {
    info: '#6366f1',     // Primary blue-violet
    positive: '#10b981', // Emerald
    negative: '#ef4444',  // Red
    warning: '#f59e0b',  // Amber
    neutral: '#6b7280',  // Gray
  },
  // Neutral scales for surfaces
  neutral: {
    0: '#ffffff',
    50: '#fafafa',
    100: '#f4f4f5',
    200: '#e4e4e7',
    300: '#d4d4d8',
    400: '#a1a1aa',
    500: '#71717a',
    600: '#52525b',
    700: '#3f3f46',
    800: '#27272a',
    900: '#18181b',
    950: '#09090b',
  }
} as const;

// Color Utility Functions
export const getColor = (colorName: keyof typeof colors, shade: number = 500) => {
  const colorGroup = colors[colorName];
  if (colorGroup && typeof colorGroup === 'object' && shade in colorGroup) {
    return colorGroup[shade as keyof typeof colorGroup];
  }
  return colorName;
};