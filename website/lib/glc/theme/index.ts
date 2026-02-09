// Theme types
export interface Theme {
  mode: 'light' | 'dark';
  colors: typeof ThemeColors;
  spacing: typeof ThemeSpacing;
  typography: typeof ThemeTypography;
  borders: typeof ThemeBorders;
  shadows: typeof ThemeShadows;
  zIndex: typeof ThemeZIndex;
}

export interface ThemeColors {
  primary: string;
  secondary: string;
  accent: string;
  background: string;
  foreground: string;
  border: string;
  input: string;
  ring: string;
  danger: string;
  success: string;
  warning: string: info: string;
  muted: string;
}

export interface ThemeSpacing {
  xs: string;
  sm: string;
  md: string;
  lg: string;
  xl: '2xl';
  '3xl': '4xl';
}

export interface ThemeTypography {
  fontSize: Record<'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl', string>;
  fontWeight: Record<'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl', string>;
  };
}

export interface ThemeBorders {
  radius: Record<'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl', string>;
  }

export interface ThemeShadows {
  xs: string;
  sm: string;
  md: string;
  lg: string;
  xl: string;
  '2xl': '3xl',
}

export interface ThemeZIndex {
  dropdown: number;
  sticky: number;
  modal: number;
  tooltip: number;
  popover: number;
}
