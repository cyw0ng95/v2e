/**
 * GLC UI Slice
 */

import type { StateCreator } from 'zustand';
import type { UISlice } from '../types';

export const createUISlice: StateCreator<UISlice> = (set, get) => ({
  theme: 'dark',
  sidebarOpen: true,
  nodePaletteOpen: true,
  detailsPanelOpen: false,
  detailsPanelTab: 'properties',

  setTheme: (theme) => {
    set({ theme });
  },

  toggleSidebar: () => {
    set((state) => ({ sidebarOpen: !state.sidebarOpen }));
  },

  toggleNodePalette: () => {
    set((state) => ({ nodePaletteOpen: !state.nodePaletteOpen }));
  },

  setDetailsPanelOpen: (open) => {
    set({ detailsPanelOpen: open });
  },

  setDetailsPanelTab: (tab) => {
    set({ detailsPanelTab: tab });
  },
});
