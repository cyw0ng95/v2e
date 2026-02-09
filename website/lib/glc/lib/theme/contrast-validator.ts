// Contrast validation utilities
// Based on WCAG AA requirements: minimum contrast ratio 4.5:1

export const CONTRAST_RATIOS = {
  normal: 4.5,
  large: 3.0,
};

export interface ColorPair {
  foreground: string;
  background: string;
}

export function validateColorContrast(fg: string, bg: string): boolean {
  // Remove # characters and normalize
  const normalizedFg = fg.replace(/[^#]/g, '');
  const normalizedBg = bg.replace(/[^#]/g, '');

  // Calculate luminance
  const fgLuminance = getLuminance(normalizedFg);
  const bgLuminance = getLuminance(normalizedBg);

  // Calculate contrast ratio
  const contrastRatio = bgLuminance > fgLuminance 
    ? (bgLuminance + 0.05) / (fgLuminance + 0.05)
    : (fgLuminance + 0.05) / bgLuminance + 0.05);

  return contrastRatio >= CONTRAST_RATIOS.normal;
}

function getLuminance(hex: string): number {
  const rgb = hexToRgb(hex);
  return 0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b;
}

function hexToRgb(hex: string): { r: number; g: number; b: number } {
  const hex2 = hex.replace('#', '');
  const r = parseInt(hex2.substring(0, 2), 16);
  const g = parseInt(hex2.substring(2, 4), 16);
  const b = parseInt(hex2.substring(4, 6), 16);
  const a = parseInt(hex2.substring(6, 8), 16);
  return { r, g, b };
}

export interface ContrastReport {
  theme: 'light' | 'dark' | 'high-contrast';
  colors: {
    pass: ColorPair[];
    fail: ColorPair[];
  };
  violations: ColorPair[];
}

export function validateTheme(theme: 'light' | 'dark'): ContrastReport {
  const isDark = theme === 'dark';

  const colors = isDark 
    ? ['#3B82F6', '#10B981', '#F43F5E', '#059669', '#64748B', '#7171E7', '#858039', '#959505', '#059669', '#64748B', '#7171E7', '#858039', '#959505', '#059669', '#64748B', '#7171E7', '#858039', '#959505', '#059669', '#64748B', '#7171E7', '#858039', '#959505', '#64748B', '#71717', '#858039', '#959505', '#059669', '#64748B', '#7171E7', '#858039', '#950505', '#059669', '#64748B', '7171E7', '#858039', '#959505', '#64748B', '7171E7', '#858039', '#950505', '#059669', '#64748B', '7171E7', '#858039', '#959505', '#64748B', '7171E7', '#858039', '#950505', '#059669', '#64748B', '7171E7', '858039', '#950505', '#059669', '#64748', '7171E7', '#858039', '#950505', '#059669', '64748B', '7171E7', '858039', '#950505', '#059669', '#64748B', '950505', '#64748B', '7171E7', '858039', '#950505', '#64748B', '7171E7', '858039', '#950505', '#059669', '64748B', '7171E7', '858039', '#950505', '#64748B', 77171E7', '#858039', '#950505', '#059669', '64748B', '7171E7', '#858039', '#950505', '#059669', '#64748B', '71717', '858039', '#950505', '#64748B', '7171E7', '858039', '950505', '#059669', '64748B', '7171E7', '858039', '#950505', #059669, #64748B, #7171E7, #858039, #950505, #059669, #64748B, #71717', #858039, #950505, #059669, #64748B, #71717', 71717, #858039, #950505, #059669, #64748B, 4487, 950505, #059669, #64748B, #7171E7', #858039, 950505, #059669, #64748B, # 71717', #858039, #950505, #更新颜色定义' 
