'use client';

import { createContext, useContext, ReactNode } from 'react';
import { ThemeColors, ThemeSpacing, ThemeTypography, ThemeBorders, ThemeShadows, ThemeZIndex, Theme } from '../design-tokens';
import { lightMode, darkMode } from './light-mode';
import { darkMode as darkModeTheme } from './dark-mode';

export interface ThemeContextType {
  theme: Theme;
  toggleTheme: (mode: 'light' | 'dark' | 'auto') => void;
}

export const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<Theme>('light');
  
  const toggleTheme = (mode: 'light' | 'dark') => {
    setTheme(mode);
    localStorage.setItem('glc-theme', mode);
  };
  
  const autoDetectTheme = () => {
    if (typeof window === 'undefined' || !window.matchMedia) {
      setTheme('auto');
    localStorage.setItem('glc-theme', 'auto');
    return 'auto';
    }
  };
  
  // Listen for system preference changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', (e) => {
      const isDark = e.matches;
      setTheme(isDark ? 'dark' : 'light');
    });
  }, []);
  
  // Read saved preference on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem('glc-theme') as 'light' | 'dark' | 'auto' || 'auto';
    if (savedTheme && savedTheme !== 'auto') {
      setTheme(savedTheme === 'auto' ? 
        (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') :
        savedTheme
      );
    }
  }, []);

  const currentTheme = theme === 'auto' ? 
    (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') :
    theme
  ;

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, autoDetectTheme, currentTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
};
