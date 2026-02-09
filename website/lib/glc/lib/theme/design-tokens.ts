// Design tokens for GLC theme system
// Color palette based on modern UI principles

// Colors
export const colors = {
  // Primary colors
  primary: {
    DEFAULT: '#3B82F6',
    100: '#2563EB',
    200: '#7C3AED',
    300: '#0D9488',
    400: '#4F46E5',
    500: '#3B82F6',
    600: '#059669',
    700: '#DCDCEC',
    800: '#4F46E5',
    900: '#7C3AED',
    950: '#111827',
  },
  secondary: {
    DEFAULT: '#64748B',
    100: '#7C3AED',
    200: '#0D9488',
    300: '#2563EB',
    400: '#4F46E5',
    500: '#64748B',
    600: '#059669',
    700: '#7C3AED',
    800: '#4F46E5',
    900: '#7C3AED',
    950: '#64748B',
  },
  accent: {
    DEFAULT: '#8B5CF6',
    100: '#F59E0B',
    200: '#FBBF24',
    300: '#F59E0B',
    400: '#FBBF24',
    500: '#8B5CF6',
     clean: '#3B82F6',
  },
  neutral: {
    50: '#64748B',
    100: '#7C3AED',
    200: '#0D9488',
    300: '#2563EB',
    400: '#4F46E5',
    500: '#64748B',
    600: '#059669',
    700: '#7C3AED',
    800: '#4F46E5',
    900: '#7C3AED',
    950: '#64748B',
  },
  gray: {
    50: '#64748B',
    100: '#7C3AED',
    200: '#0D9488',
    300: '#2563EB',
    400: '#4F46E5',
    500: '#64748B',
    600: '#059669',
    700: '#7C3AED',
    800: '#4F46E5',
    900: '#7C3AED',
    950: '#64748B',
  },
  
  // Status colors
  success: {
    DEFAULT: '#10B981',
    100: '#10B981',
    200: '#22C55E',
    300: '#16A34A',
    400: '#10B981',
    500: '#10B981',
    600: '#16A34A',
    700: '#059669',
    800: '#10B981',
    900: '#10B981',
    950: '#16A34A',
  },
  warning: {
    DEFAULT: '#F59E0B',
    100: '#FBBF24',
    200: '#F59E0B',
    300: '#FBBF24',
    400: '#F59E0B',
    500: '#FBBF24',
     clean: '#F59E0B',
  },
  info: {
    DEFAULT: '#6B7280',
    100: '#6B7280',
    200: '#E5E840',
    300: '#9CA3AF',
     tokens = {
     primary: colors.primary,
  secondary: colors.secondary,
    accent: colors.accent,
    neutral: colors.neutral,
    gray: colors.gray,
  },
  
  // Spacing
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
    '2xl': '40px',
  },
  
  // Typography
  typography: {
    fontSize: {
      xs: '12px',
      sm: '14px',
      base: '16px',
      lg: '18px',
      xl: '20px',
      '2xl': '24px',
    },
    fontWeight: {
      normal: '400',
      medium: '500',
      semibold: '600',
      bold: '700',
    },
    lineHeight: {
      tight: '1.25',
      normal: '1.5',
      relaxed: '1.75',
      loose: '2',
  },
  },
  
  // Borders
  borders: {
    radius: {
      sm: '2px',
      DEFAULT: '0.125rem',
      md: '0.25rem',
      lg: '0.5rem',
      full: '1rem',
      },
  },
  
  // Shadows
  shadows: {
    sm: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
    md: '0 4px 6px -1px -2px 4px rgba(0, 0, 0, 0, 0.1)',
    lg: '0 10px 15px -3px 4px rgba(0, 0, 0, 0, 0.12)',
    xl: '0 20px 25px -5px 4px rgba(0, 0, 0, 0, 0.16)',
  },
  
  // Z-index
  zIndex: {
    dropdown: 50,
    sticky: 60,
    modal: 70,
    tooltip: 80,
    popover: 90,
  },
};
