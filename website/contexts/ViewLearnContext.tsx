'use client';

import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react';

export type ViewLearnMode = 'view' | 'learn';

interface ViewLearnContextType {
  mode: ViewLearnMode;
  setMode: (mode: ViewLearnMode) => void;
}

const ViewLearnContext = createContext<ViewLearnContextType | undefined>(undefined);

interface ViewLearnProviderProps {
  children: ReactNode;
  defaultMode?: ViewLearnMode;
}

export function ViewLearnProvider({ children, defaultMode = 'view' }: ViewLearnProviderProps) {
  const [mode, setMode] = useState<ViewLearnMode>(defaultMode);

  // Persist mode to localStorage for better UX
  useEffect(() => {
    const saved = localStorage.getItem('v2e-view-learn-mode') as ViewLearnMode;
    if (saved && (saved === 'view' || saved === 'learn')) {
      setMode(saved);
    }
  }, []);

  // Save mode changes to localStorage
  useEffect(() => {
    localStorage.setItem('v2e-view-learn-mode', mode);
  }, [mode]);

  return (
    <ViewLearnContext.Provider value={{ mode, setMode }}>
      <div data-learn-mode={mode === 'learn'}>
        {children}
      </div>
    </ViewLearnContext.Provider>
  );
}

export function useViewLearnMode() {
  const context = useContext(ViewLearnContext);
  if (!context) {
    throw new Error('useViewLearnMode must be used within ViewLearnProvider');
  }
  return context;
}
